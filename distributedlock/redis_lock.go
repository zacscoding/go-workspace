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
		locker:  redislock.New(client),
		lockMap: make(map[string]*lockHolder),
		ttl:     ttl,
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
	l.mutex.Unlock()
	return holder.tryLock(ctx)
}

func (l *RedisLockRegistry) Unlock(taskId string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
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
		Context:       ctx,
		RetryStrategy: redislock.LinearBackoff(1 * time.Second),
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

func (h *lockHolder) loopRefresh() {
	prefix := ">>>>>> loopRefresh:"
	log.Printf("%s start\n", prefix)
	for {
		ttl, err := h.lock.TTL()
		if err != nil {
			log.Printf("%s could not fetch ttl: %v\n", prefix, err)
			break
		}
		if ttl == 0 {
			log.Printf("%s: zero ttl\n", prefix)
			return
		}

		ttlMills := ttl.Milliseconds() - 500
		if ttlMills < 0 {
			ttlMills = 1
		}
		log.Printf("%s wait %d ms\n", prefix, ttlMills)
		timer := time.NewTimer(time.Millisecond * time.Duration(ttlMills))

		select {
		case <-h.cancel:
			log.Printf("%s canceled\n", prefix)
			return
		case <-timer.C:
			log.Printf("%s try to refresh\n", prefix)
			err := h.lock.Refresh(h.ttl, nil)
			if err != nil {
				log.Printf("%s failed to refresh:%v\n", prefix, err.Error())
			}
		}
	}
}