package basictime

import (
	"context"
	"testing"
	"time"
)

// NoRetry and TTL with 3 secs
func Test01(t *testing.T) {
	ctx := context.Background()
	retry := NoRetry()
	complete := make(chan *TaskResult)

	go func(c chan *TaskResult) {
		c <- DoWork(ctx, 3*time.Second, retry)
	}(complete)

	select {
	case r := <-complete:
		if r.TryCount != 1 {
			t.Errorf("expected try count:%d, got: %d", 1, r.TryCount)
			t.Fail()
		}
		if r.Canceled {
			t.Errorf("expected canceled:%v, got: %v", false, r.Canceled)
			t.Fail()
		}
	}
}

// Retry with 1 secs and TTL with 3.5 secs
func Test02(t *testing.T) {
	ctx := context.Background()
	retry := LinearBackoff(1 * time.Second)
	complete := make(chan *TaskResult)

	go func(c chan *TaskResult) {
		c <- DoWork(ctx, 3500*time.Millisecond, retry)
	}(complete)

	select {
	case r := <-complete:
		if r.TryCount < 3 {
			t.Errorf("expected try count is greater than 3, got:%d", r.TryCount)
			t.Fail()
		}
		if r.Canceled {
			t.Errorf("expected canceled:%v, got: %v", false, r.Canceled)
			t.Fail()
		}
	}
}

// Retry with 1 secs and Ctx with timeout 3 secs
func Test03(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	retry := LinearBackoff(1 * time.Second)
	complete := make(chan *TaskResult)

	go func(c chan *TaskResult) {
		c <- DoWork(ctx, 10*time.Second, retry)
	}(complete)

	select {
	case r := <-complete:
		if r.TryCount < 3 {
			t.Errorf("expected try count is greater than 3, got:%d", r.TryCount)
			t.Fail()
		}
		if !r.Canceled {
			t.Errorf("expected canceled:%v, got: %v", true, r.Canceled)
			t.Fail()
		}
	}
}
