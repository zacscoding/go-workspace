package timeexample

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	ch := make(chan struct{})
	task := TimeoutTest{
		Duration: time.Millisecond * 200,
		Close:    ch,
	}
	go task.loopTimer()

	time.Sleep(time.Second)
	close(ch)
	fmt.Println("## Total executed :", task.Executed)
	// Output
	// ## Total executed : 1
}

func TestAfter(t *testing.T) {
	ch := make(chan struct{})
	task := TimeoutTest{
		Duration: time.Millisecond * 200,
		Close:    ch,
	}
	go task.loopAfter()

	time.Sleep(time.Second)
	close(ch)
	fmt.Println("## Total executed :", task.Executed)
	// Output
	// ## Total executed : 1
}

func TestT(t *testing.T) {
	ctx := context.Background()
	ch := time.After(time.Second)

	select {
	case <-ctx.Done():
		fmt.Println("ctx is done!")
	case <-ch:
		fmt.Println("after is done!")
	}
}

type TimeoutTest struct {
	Duration time.Duration
	Executed int
	Close    chan struct{}
}

func (t *TimeoutTest) loopTimer() {
	timer := time.NewTimer(t.Duration)
	for {
		select {
		case <-t.Close:
			return
		case <-timer.C:
			t.Executed++
		}
	}
}

func (t *TimeoutTest) loopAfter() {
	ch := time.After(t.Duration) // equals to time.NewTimer(duration).C
	for {
		select {
		case <-t.Close:
			return
		case <-ch:
			t.Executed++
		}
	}
}
