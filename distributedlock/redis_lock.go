// redis lock is working..
package distributedlock

import (
	"context"
	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v7"
	"sync"
	"time"
)

// compile time check
var _ LockRegistry = (*RedisLockRegistry)(nil)

type RedisLockRegistry struct {
	locker  *redislock.Client
	mutex   sync.Mutex
	lockMap map[string]*lockHolder
}

func NewRedisLockRegistry(client *redis.Client) *RedisLockRegistry {
	locker := redislock.New(client)
	// TODO : lock manage loop
	return &RedisLockRegistry{
		locker: locker,
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
		holder = &lockHolder{locker: l.locker, taskId: taskId}
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
}

func (l *lockHolder) tryLock(ctx context.Context) bool {
	lock, err := l.locker.Obtain(l.taskId, time.Second, &redislock.Options{
		Context: ctx,
	})
	if err != nil {
		return false
	}
	l.lock = lock
	return true
}

func (l *lockHolder) unlock() {
	ttl, err := l.lock.TTL()
	if err != nil {
		return
	}

	if ttl > 0 {
		l.lock.Release()
	}
}
