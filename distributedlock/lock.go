package distributedlock

import (
	"context"
	"time"
)

type LockRegistry interface {
	// TryLockWithTimeout try to acquire a lock with given taskId and timeout duration.
	// returns a true if success to acquire a lock, otherwise false
	TryLockWithTimeout(taskId string, duration time.Duration) bool

	// TryLockWithContext try to acquire a lock with given taskId and context to cancel.
	// returns a true if success to acquire a lock, otherwise false
	TryLockWithContext(taskId string, ctx context.Context) bool

	// Unlock release a lock with given task id.
	Unlock(taskId string)
}
