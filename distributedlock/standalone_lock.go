package distributedlock

import (
	"context"
	"sync"
	"time"
)

// compile time check
var _ LockRegistry = (*StandaloneLockRegistry)(nil)

type StandaloneLockRegistry struct {
	lock        sync.Mutex
	lockChanMap map[string]chan struct{}
}

func NewStandaloneLockRegistry() *StandaloneLockRegistry {
	l := &StandaloneLockRegistry{}
	l.lockChanMap = make(map[string]chan struct{})
	return l
}

func (l *StandaloneLockRegistry) TryLockWithTimeout(taskId string, duration time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	return l.TryLockWithContext(taskId, ctx)
}

func (l *StandaloneLockRegistry) TryLockWithContext(taskId string, ctx context.Context) bool {
	l.lock.Lock()
	lockChan, ok := l.lockChanMap[taskId]
	if !ok {
		lockChan = make(chan struct{}, 1)
		l.lockChanMap[taskId] = lockChan
	}
	l.lock.Unlock()

	select {
	case lockChan <- struct{}{}:
		return true
	case <-ctx.Done():
		// timeout or cancel
		return false
	}
}

func (l *StandaloneLockRegistry) Unlock(taskId string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	lockChan, ok := l.lockChanMap[taskId]
	if !ok {
		return
	}
	<-lockChan
}
