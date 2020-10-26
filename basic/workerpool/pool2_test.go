package workerpool

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var workerCount int32

type worker struct {
}

func (w *worker) DoSomthing() {
	secs := rand.Intn(3)
	fmt.Println("Sleep :", secs)
	time.Sleep(time.Duration(secs) * time.Second)
}

func TestWorkers(t *testing.T) {
	pool := sync.Pool{
		New: func() interface{} {
			atomic.AddInt32(&workerCount, 1)
			return &worker{}
		},
	}

	wg := sync.WaitGroup{}
	workers := 100
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			w := pool.Get().(*worker)
			w.DoSomthing()
			pool.Put(w)
		}()
	}
	wg.Wait()
	fmt.Println("Total workers:", workerCount)
}

type workerPool struct {
	limited    uint
	workerChan chan struct{}
}

func NewWorkerPool(limited uint) {
	
}
