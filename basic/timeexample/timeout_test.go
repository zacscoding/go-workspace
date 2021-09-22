package timeexample

import (
	"context"
	"fmt"
	"log"
	"sync"
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

func TestTimerStop(t *testing.T) {
	timer := time.NewTimer(time.Second * 5)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-timer.C:
			log.Println("Received time from timer.C")
		}
	}()
	time.Sleep(time.Second * 3)
	if !timer.Stop() {
		<-timer.C
	}
	wg.Wait()
}

func TestTimerReset(t *testing.T) {
	var (
		timerDuration = time.Second * 5
		resetAfterDuration = time.Second * 4
		// resetAfterDuration = time.Second * 6
		start              = time.Now()
		timer              = time.NewTimer(timerDuration)
		elapsed            time.Duration
		wg                 = sync.WaitGroup{}
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		select {
		case <-timer.C:
			elapsed = time.Now().Sub(start)
			log.Println("Receive from timer.C")
		}
	}()
	go func() {
		defer wg.Done()
		time.Sleep(resetAfterDuration)
		ok := timer.Reset(time.Second * 10)
		log.Println("timer.Reset >", ok)
		//if timer.Stop() {
		//	log.Println("Success to stop timer")
		//	ok := timer.Reset(time.Second * 10)
		//	log.Println("timer.Reset >", ok)
		//} else {
		//	log.Println("Failed to stop timer")
		//}
	}()
	wg.Wait()

	log.Println("Elapsed:", elapsed)
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
