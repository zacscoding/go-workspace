package main

import (
	"fmt"
	"github.com/fatih/color"
	"go-workspace/distributedlock"

	"math/rand"
	"strconv"
	"time"
)

func main() {
	// runSingleTask()
	runMultipleTask()
}

type Worker struct {
	lockRegistry distributedlock.LockRegistry
	TaskId       string
	Name         string
	SleepSec     int
	Logger       *color.Color
}

func runSingleTask() {
	taskId := "task01"
	lock := distributedlock.NewStandaloneLockRegistry()
	for i := 1; i <= 5; i++ {
		t := &Worker{
			lockRegistry: lock,
			TaskId:       taskId,
			Name:         "Task" + strconv.Itoa(i),
			SleepSec:     3,
		}
		t.DoTask()
	}

	time.Sleep(5 * time.Minute)
}

func runMultipleTask() {
	lock := distributedlock.NewStandaloneLockRegistry()
	taskCount := 3
	workerCount := 3
	for i := 1; i <= taskCount; i++ {
		taskId := "task-" + strconv.Itoa(i)
		for j := 1; j <= workerCount; j++ {
			workerName := "worker-" + strconv.Itoa(i) + "-" + strconv.Itoa(j)
			t := &Worker{
				lockRegistry: lock,
				TaskId:       taskId,
				Name:         workerName,
				SleepSec:     3,
				Logger:       newColor(i),
			}
			t.DoTask()
		}
	}

	time.Sleep(5 * time.Minute)
}

func (w Worker) DoTask() {
	go func() {
		ticker := time.NewTicker(time.Duration(w.SleepSec) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fmt.Printf("[%s] Try to get [%s] lock\n", w.Name, w.TaskId)
				if w.lockRegistry.TryLockWithTimeout(w.TaskId, 1*time.Second) {
					w.Logger.Printf("[%s] >>>>>>>>>> [%s] Acquire lock <<<<<<<<<<\n", w.TaskId, w.Name)
					time.Sleep(time.Duration(rand.Intn(5)+1) * time.Second)
					w.Logger.Printf("[%s] <<<<<<<<<< [%s] Release lock >>>>>>>>>>\n", w.TaskId, w.Name)
					w.lockRegistry.Unlock(w.TaskId)
					continue
				}
				fmt.Printf("[%s] Failed to get [%s] lock\n", w.Name, w.TaskId)
			}
		}
	}()
}

func newColor(order int) *color.Color {
	return color.New(color.Attribute(int(color.FgBlack) + order))
}
