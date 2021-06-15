package cancelexample

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"time"
)

type Worker struct {
	ctx     context.Context
	cancel  context.CancelFunc
	running int32
}

func (w *Worker) Start() error {
	if !atomic.CompareAndSwapInt32(&w.running, 0, 1) {
		return errors.New("already running worker")
	}

	w.ctx, w.cancel = context.WithCancel(context.Background())
	go w.loopWork1()
	go w.loopWork2()
	return nil
}

func (w *Worker) Stop() {
	if !atomic.CompareAndSwapInt32(&w.running, 1, 0) {
		return
	}
	w.cancel()
}

func (w *Worker) Running() bool {
	return atomic.LoadInt32(&w.running) == 1
}

func (w *Worker) loopWork1() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	proceed := 0
	for {
		select {
		case <-ticker.C:
			log.Printf("Do Work1()... %d", proceed)
			proceed++
		case <-w.ctx.Done():
			log.Println("Terminate Work1()...")
			return
		}
	}
}

func (w *Worker) loopWork2() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	proceed := 0
	for {
		select {
		case <-ticker.C:
			log.Printf("Do Work2()... %d", proceed)
			proceed++
		case <-w.ctx.Done():
			log.Println("Terminate Work2()...")
			return
		}
	}
}
