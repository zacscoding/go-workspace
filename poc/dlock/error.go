package dlock

import "errors"

var (
	// ErrDeadlock is returned by Lock when trying to lock twice without unlocking first
	ErrDeadlock = errors.New("trying to acquire a lock twice")

	// ErrTimeoutAcquireLock is returned by Lock when occur timeout.
	ErrTimeoutAcquireLock = errors.New("timeout to acquire a lock")

	// ErrNotLocked is returned by Unlock when trying to release a lock that has not first be acquired.
	ErrNotLocked = errors.New("not locked")
)
