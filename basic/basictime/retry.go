package basictime

import (
	"context"
	"fmt"
	"log"
	"time"
)

type RetryStrategy interface {
	NextBackoff() time.Duration
}

type linearBackoff time.Duration

func (l linearBackoff) NextBackoff() time.Duration {
	return time.Duration(l)
}

func NoRetry() RetryStrategy {
	return linearBackoff(0)
}

func LinearBackoff(backoff time.Duration) RetryStrategy {
	return linearBackoff(backoff)
}

type TaskResult struct {
	TryCount int
	Canceled bool
}

func (r TaskResult) String() string {
	return fmt.Sprintf("TaskResult{TryCount:%d, Canceled:%v}", r.TryCount, r.Canceled)
}

func DoWork(ctx context.Context, ttl time.Duration, retry RetryStrategy) *TaskResult {
	log.Println("DoWork().. ttl:", ttl, ", retryStrategy:", retry)
	taskResult := &TaskResult{}
	var (
		timer *time.Timer
	)
	for deadline := time.Now().Add(ttl); time.Now().Before(deadline); {
		taskResult.TryCount++
		// Try to do something such connect to server in here.
		log.Printf("[Try %d] Deadline:%v\n", taskResult.TryCount, deadline)

		if retry.NextBackoff() < 1 {
			log.Println("No retry")
			break
		}

		if timer == nil {
			timer = time.NewTimer(retry.NextBackoff())
		} else {
			timer.Reset(retry.NextBackoff())
		}

		select {
		case <-ctx.Done():
			log.Println("timeout..")
			taskResult.Canceled = true
			return taskResult
		case <-timer.C:
		}
	}
	return taskResult
}
