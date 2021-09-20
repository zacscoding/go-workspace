package main

import (
	"errors"
	"math/rand"
	"sync/atomic"
	"time"
)

type stubResouce struct {
	running    int32
	sleepMills int
}

func (r *stubResouce) Use() error {
	rand.Seed(time.Now().UnixNano())
	if !atomic.CompareAndSwapInt32(&r.running, 0, 1) {
		return errors.New("atomic collision")
	}

	sleep := time.Duration(rand.Intn(r.sleepMills))
	time.Sleep(sleep * time.Millisecond)

	atomic.CompareAndSwapInt32(&r.running, 1, 0)
	return nil
}
