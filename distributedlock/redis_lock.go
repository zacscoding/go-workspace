// redis lock is working..
package distributedlock

import (
	"context"
	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v7"
	"log"
	"sync"
	"time"
)

// compile time check
var _ LockRegistry = (*RedisLockRegistry)(nil)

type RedisLockRegistry struct {
	locker  *redislock.Client
	mutex   sync.Mutex
	lockMap map[string]*lockHolder
	ttl     time.Duration
}

func NewRedisLockRegistry(client *redis.Client, ttl time.Duration) *RedisLockRegistry {
	return &RedisLockRegistry{
		locker: redislock.New(client),
		ttl:    ttl,
	}
}

func (l *RedisLockRegistry) TryLockWithTimeout(taskId string, duration time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	return l.TryLockWithContext(taskId, ctx)
}

func (l *RedisLockRegistry) TryLockWithContext(taskId string, ctx context.Context) bool {
	l.mutex.Lock()
	holder, ok := l.lockMap[taskId]
	if !ok {
		holder = &lockHolder{
			locker: l.locker,
			taskId: taskId,
			cancel: make(chan bool),
			ttl:    l.ttl,
		}
		l.lockMap[taskId] = holder
	}
	return holder.tryLock(ctx)
}

func (l *RedisLockRegistry) Unlock(taskId string) {
	l.mutex.Lock()
	lock, ok := l.lockMap[taskId]
	if !ok {
		return
	}
	lock.unlock()
}

type lockHolder struct {
	locker *redislock.Client
	taskId string
	lock   *redislock.Lock
	ttl    time.Duration
	cancel chan bool
}

func (h *lockHolder) tryLock(ctx context.Context) bool {
	lock, err := h.locker.Obtain(h.taskId, h.ttl, &redislock.Options{
		Context: ctx,
	})
	if err != nil {
		return false
	}
	// this lock is held during 1 seconds because of above Obtain argument. So have to refresh like lock.Refresh()
	h.lock = lock
	go h.loopRefresh()
	return true
}

func (h *lockHolder) unlock() {
	ttl, err := h.lock.TTL()
	if err != nil {
		return
	}
	h.cancel <- true
	if ttl > 0 {
		h.lock.Release()
	}
}

// TODO : impl
func (h *lockHolder) loopRefresh() {
	for {
		ttl, err := h.lock.TTL()
		if err != nil {
			log.Printf("loopRefresh: could not fetch ttl: %v\n", err)
			break
		}

		ttlMills := ttl.Milliseconds() - 500
		if ttlMills < 0 {
			ttlMills = 1
		}
		timer := time.NewTimer(time.Millisecond * time.Duration(ttlMills))
		select {
		case <-h.cancel:
			log.Println("loopRefresh: canceled")
			return
		case <-timer.C:
			err := h.lock.Refresh(h.ttl, nil)
			if err != nil {
				log.Println("loopRefresh: failed to refresh")
			}
		}
	}
}
