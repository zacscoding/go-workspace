package main

import (
	"log"
	"sync"
	"time"
)

func main() {
	w := &worker{
		quitChannel: make(chan struct{}, 1),
	}
	go w.doWork()
	time.Sleep(2 * time.Second)
	log.Println("Terminate!")
	w.Close()
}

type worker struct {
	quitChannel chan struct{}
	running     sync.WaitGroup
}

func (w *worker) doWork() {
	w.running.Add(1)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	defer w.running.Done()
	for {
		select {
		case <-ticker.C:
			log.Println("Start work...")
			time.Sleep(time.Millisecond)
			log.Println("Completed work..")
		case <-w.quitChannel:
			log.Println("Terminate work..!")
			return
		}
	}
}

func (w *worker) Close() {
	close(w.quitChannel)
	w.running.Wait()
}
