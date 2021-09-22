package dlock

import (
	"context"
	"time"
)

const (
	DefaultLockTTL = time.Second * 5
)

type LockRegistry interface {
	NewLock(key string) Lock
}

type Lock interface {
	Lock(ctx context.Context, ttl time.Duration) error
	LockWithData(ctx context.Context, ttl time.Duration, data []byte) error
	Extend() (bool, error)
	Unlock() error
}
