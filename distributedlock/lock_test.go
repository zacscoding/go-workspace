package distributedlock

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

var l LockRegistry

func init() {
	l = NewStandaloneLockRegistry()
}

func TestTryLockWithTimeout(t *testing.T) {
	taskId := "task01"

	result := l.TryLockWithTimeout(taskId, 1*time.Second)
	if !result {
		t.Errorf("acquire lock: expected true, but false")
	}

	result = l.TryLockWithTimeout(taskId, 1*time.Second)
	if result {
		t.Errorf("try lock after acquired: expected false, but but true")
	}

	l.Unlock(taskId)

	result = l.TryLockWithTimeout(taskId, 1*time.Second)
	if !result {
		t.Errorf("try lock after unlock: expected true, but but false")
	}
}

func TestTryLockWithContext(t *testing.T) {
	taskId := "task01"

	// acquire lock
	result := l.TryLockWithTimeout(taskId, 1*time.Second)
	if !result {
		t.Errorf("acquire lock: expected true, but false")
	}

	wait := sync.WaitGroup{}

	cancelCtx, cancel := context.WithCancel(context.Background())
	timeoutCtx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)

	wait.Add(2)
	// try lock with cancel ctx
	go func() {
		defer wait.Done()
		result = l.TryLockWithContext(taskId, cancelCtx)
		if result {
			t.Errorf("try lock with cancel: expected false but true")
		}
	}()
	// try lock with timeout ctx
	go func() {
		defer wait.Done()
		result = l.TryLockWithContext(taskId, timeoutCtx)
		if result {
			t.Errorf("try lock with timeout: expected false but true")
		}
	}()
	cancel()

	wait.Wait()
}

func TestTryLockWithRace(t *testing.T) {
	taskId := "task01"

	go func() {
		doTask("task01", taskId, l)
	}()
	go func() {
		doTask("task02", taskId, l)
	}()
	go func() {
		doTask("task03", taskId, l)
	}()

	time.Sleep(1 * time.Minute)
}

func doTask(taskName, taskId string, l LockRegistry) {
	for {
		result := l.TryLockWithTimeout(taskId, 1*time.Second)
		if !result {
			fmt.Printf("[%s] failed to acquire lock\n", taskName)
			time.Sleep(1 * time.Second)
		}

		sleep := rand.Intn(3) + 1
		fmt.Printf("[%s] success to acquire lock with sleep: %d\n", taskName, sleep)
		time.Sleep(time.Duration(sleep) * time.Second)
		fmt.Printf("[%s] release lock\n", taskName)
		l.Unlock(taskId)
	}
}
