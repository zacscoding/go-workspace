package dlock

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

const (
	prefix = "dlock"
)

type Lock struct {
	cli     redis.UniversalClient
	lockKey string
}

func NewLock() *Lock {
	return &Lock{}
}

// Lock try to acquire a lock with lockKey in Lock and returns nil if success, otherwise returns an error.
func (l *Lock) Lock(ctx context.Context) error {
	panic("")
}

// Refresh resets current lock's expiry and returns nil if success to refresh, otherwise an error.
func (l *Lock) Refresh(ctx context.Context, ttl time.Duration) error {
	panic("")
}

// Unlock releases current lock and returns nil if success to release, otherwise an error.
func (l *Lock) Unlock(ctx context.Context) error {
	panic("")
}
