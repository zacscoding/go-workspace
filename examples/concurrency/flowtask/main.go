// TODO : temporary code... will test from go-workspace/basic/workerpool
package main

import (
	"fmt"
	"go-workspace/basic/workerpool"
	"time"
)

func main() {
	resultCh := make(chan interface{})
	wp := workerpool.NewWorkerPool(2, resultCh)
	wp.Submit(func() interface{} {
		fmt.Println("Start task...", 1)
		defer fmt.Println("Complete task...", 1)
		time.Sleep(1 * time.Second)
		return "Work1"
	})
	wp.Submit(func() interface{} {
		fmt.Println("Start task...", 2)
		defer fmt.Println("Complete task...", 2)
		time.Sleep(1 * time.Second)
		return "Work2"
	})
	wp.Submit(func() interface{} {
		fmt.Println("Start task...", 3)
		defer fmt.Println("Complete task...", 3)
		time.Sleep(1 * time.Second)
		return "Work3"
	})
	wp.Submit(func() interface{} {
		fmt.Println("Start task...", 4)
		defer fmt.Println("Complete task...", 4)
		time.Sleep(1 * time.Second)
		return "Work4"
	})
	//for i := 0; i < 10; i++ {
	//	wp.Submit(func() interface{} {
	//		fmt.Println("Start task...", i)
	//		defer fmt.Println("Complete task...", i)
	//		time.Sleep(1 * time.Second)
	//		if i == 5 {
	//			return errors.New("force error")
	//		}
	//		return "Work " + strconv.Itoa(i)
	//	})
	//}
	//
	//waitJobs(wp, resultCh)
}

func waitJobs(wp *workerpool.WorkerPool, resultCh chan interface{}) {
	timer := time.NewTimer(5 * time.Second)
	defer wp.Close()
	for {
		select {
		case res := <-resultCh:
			switch res.(type) {
			case error:
				fmt.Println("Error occur!", res.(error).Error())
			default:
				fmt.Println("Receive result..", res)
			}
		case <-timer.C:
			fmt.Println("Timeout!!!")
		}
	}
}

//import (
//	"errors"
//	"fmt"
//	"math/rand"
//	"time"
//)
//
//type doFunc func() interface{}
//
//type WorkerPool struct {
//	jobCh    chan doFunc
//	resultCh chan<- interface{}
//	quitCh   chan struct{}
//}
//
//func (w *WorkerPool) loopWorker() {
//	for {
//		select {
//		case do := <-w.jobCh:
//			go func() {
//				w.resultCh <- do()
//			}()
//		case <-w.quitCh:
//			return
//		}
//	}
//}
//
//func (w *WorkerPool) Submit(doFunc func() interface{}) {
//	w.jobCh <- doFunc
//}
//
//func (w *WorkerPool) Close() {
//	w.quitCh <- struct{}{}
//	close(w.jobCh)
//}
//
//func NewWorkerPool(size int, resultCh chan<- interface{}) *WorkerPool {
//	worker := &WorkerPool{
//		jobCh:    make(chan doFunc),
//		resultCh: resultCh,
//		quitCh:   make(chan struct{}, 1),
//	}
//	go worker.loopWorker()
//	return worker
//}
//
//func main() {
//	timer := time.NewTicker(5 * time.Second)
//	resultCh := make(chan interface{})
//	worker := NewWorkerPool(3, resultCh)
//	taskCount := 10
//	remain := taskCount
//
//	for i := 0; i < taskCount; i++ {
//		worker.Submit(func() interface{} {
//			fmt.Println("Start worker...")
//			time.Sleep(time.Second)
//			fmt.Println(">> Complete")
//			if rand.Intn(10) == 1 {
//				return errors.New("force error")
//			}
//			return "Complete"
//		})
//	}
//
//	for {
//		select {
//		case res := <-resultCh:
//			remain--
//			switch res.(type) {
//			case error:
//				fmt.Println("Error occur")
//				worker.Close()
//				break
//			default:
//				fmt.Println("## Receive ::", res)
//			}
//			if remain == 0 {
//				break
//			}
//		case <-timer.C:
//			fmt.Println("Timeout!!")
//			break
//		}
//	}
//}
