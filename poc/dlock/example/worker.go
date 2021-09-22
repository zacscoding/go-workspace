package main

import (
	"context"
	"fmt"
	"go-workspace/poc/dlock"
	"sync"
	"time"
)

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("15:04:05.000"))
	return []byte(stamp), nil
}

type worker struct {
	id       string
	registry dlock.LockRegistry
	key      string
	resource *stubResouce
	group    *sync.WaitGroup
	result   *result
}

type result struct {
	WorkerID string `json:"workerId"`
	Acquired bool   `json:"acquired"`

	AttemptAt *JSONTime `json:"attemptAt,omitempty"`
	AcquireAt *JSONTime `json:"acquireAt,omitempty"`
	FailureAt *JSONTime `json:"failureAt,omitempty"`
	LeaseAt   *JSONTime `json:"leaseAt,omitempty"`

	ErrLock        string `json:"errLock,omitempty"`
	ErrUseResource string `json:"errUseResource,omitempty"`
	ErrUnlock      string `json:"errUnlock,omitempty"`
}

func (w *worker) doWork(startSleepMills int) {
	if startSleepMills > 0 {
		time.Sleep(time.Duration(startSleepMills) * time.Millisecond)
	}
	defer w.group.Done()
	w.result = &result{
		WorkerID: w.id,
	}

	lock := w.registry.NewLock(w.key)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(lockTimeoutMills)*time.Millisecond)
	defer cancel()

	w.result.AttemptAt = now()
	err := lock.Lock(ctx, 0)
	if err != nil {
		w.result.FailureAt = now()
		w.result.ErrLock = err.Error()
		return
	}
	defer func() {
		w.result.LeaseAt = now()
		if err := lock.Unlock(); err != nil {
			w.result.ErrUnlock = err.Error()
		}
	}()

	w.result.Acquired = true
	w.result.AcquireAt = now()

	if err := w.resource.Use(); err != nil {
		w.result.ErrUseResource = err.Error()
	}
}

func now() *JSONTime {
	n := JSONTime(time.Now())
	return &n
}
