package closechannel

import (
	"log"
	"testing"
	"time"
)

func TestCloseChannel(t *testing.T) {
	u := unit{done: make(chan struct{})}
	go func() {
		u.Run(time.Second * 5)
	}()
	go func() {
		time.Sleep(2 * time.Second)
		u.Cancel()
	}()
	u.Wait()
	log.Println("Complete..")
	time.Sleep(5 * time.Second)
}

type unit struct {
	done   chan struct{}
	closed bool
}

func (u *unit) Run(sleep time.Duration) {
	log.Printf("start to run.. sleep:%v", sleep)
	time.Sleep(sleep)
	if !u.closed {
		u.done <- struct{}{}
	}
	log.Println("complete to run..")
}

func (u *unit) Cancel() {
	log.Println("cancel")
	close(u.done)
	u.closed = true
}

func (u *unit) Wait() {
	<-u.done
}
