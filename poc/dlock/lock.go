package dlock

import "context"

// TODO: add TTL / extend lease

type LockRegistry interface {
	NewLock(key string) Lock
}

type Lock interface {
	Lock(ctx context.Context) error
	LockWithData(ctx context.Context, data []byte) error
	Unlock() error
}
