package distributedlock

import (
	"context"
	"time"
)

type LockRegistry interface {
	TryLockWithTimeout(taskId string, duration time.Duration) bool
	TryLockWithContext(taskId string, ctx context.Context) bool
	Unlock(taskId string)
}
