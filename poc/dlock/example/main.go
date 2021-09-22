package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-zookeeper/zk"
	"go-workspace/poc/dlock"
	"go-workspace/poc/dlock/zklock"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"
)

var (
	resourceUsingMills = 2000
	lockTimeoutMills   = 3000
)

type BeforeCloseFn func(key string)

func main() {
	// Setup
	var (
		// key           = "/" + uuid.NewString()[:4] + "/task1"
		key           = "/task1"
		registry      dlock.LockRegistry
		beforeCloseFn BeforeCloseFn
		err           error
		resource      = &stubResouce{sleepMills: resourceUsingMills}
		workers       []*worker
		wg            = sync.WaitGroup{}
	)
	registry, beforeCloseFn, err = newZKRegistry()
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		w := worker{
			id:       fmt.Sprintf("Worker-%d", i+1),
			registry: registry,
			key:      key,
			resource: resource,
			group:    &wg,
		}
		workers = append(workers, &w)
		go w.doWork(rand.Intn(5000))
	}

	wg.Wait()

	var (
		successes, failures []*worker
	)
	for _, w := range workers {
		if w.result.Acquired {
			successes = append(successes, w)
		} else {
			failures = append(failures, w)
		}
	}

	sort.Slice(successes, func(i, j int) bool {
		lt := time.Time(*successes[i].result.AcquireAt)
		rt := time.Time(*successes[j].result.AcquireAt)
		return lt.UnixNano()-rt.UnixNano() < 0
	})
	sort.Slice(failures, func(i, j int) bool {
		lt := time.Time(*failures[i].result.FailureAt)
		rt := time.Time(*failures[j].result.FailureAt)
		return lt.UnixNano()-rt.UnixNano() < 0
	})

	log.Println("## Check success workers")
	for i := 0; i < len(successes); i++ {
		if i != 0 {
			prevLeaseAt := time.Time(*successes[i-1].result.LeaseAt)
			acquireAt := time.Time(*successes[i].result.AcquireAt)
			if acquireAt.Before(prevLeaseAt) {
				log.Println("## Find invalid acquiredAt time")
			}
		}
		b, _ := json.Marshal(&successes[i].result)
		log.Println(">", string(b))
	}

	log.Println("## Check failure workers")
	successLastLeaseAt := time.Time(*successes[len(successes)-1].result.LeaseAt)
	for _, failure := range failures {
		attemptAt := time.Time(*failure.result.AttemptAt)
		failureAt := time.Time(*failure.result.FailureAt)
		if failureAt.After(successLastLeaseAt) {
			log.Println("## Find invalid failureAt")
		}
		timeoutMills := int(failureAt.Sub(attemptAt).Milliseconds())
		if lockTimeoutMills-100 >= timeoutMills || timeoutMills >= lockTimeoutMills+100 {
			log.Println("## Find invalid lock timeout")
		}
		b, _ := json.Marshal(&failure.result)
		log.Println(">", string(b))
	}
	log.Printf("Success Workers: %d, Failure Workers: %d", len(successes), len(failures))

	if beforeCloseFn != nil {
		beforeCloseFn(key)
	}
}

func newZKRegistry() (dlock.LockRegistry, BeforeCloseFn, error) {
	var (
		zkServers = []string{"localhost:2181"}
	)

	conn, eventCH, err := zk.Connect(zkServers, time.Minute, zk.WithLogger(&zkLogger{}))
	if err != nil {
		return nil, nil, err
	}
	for e := range eventCH {
		if e.Type == zk.EventSession && e.State == zk.StateConnected {
			break
		}
	}

	beforeCloseFn := func(key string) {
		defer conn.Close()
		log.Println("Check children:", key)
		children, _, err := conn.Children(key)
		if err != nil {
			if err != zk.ErrNoNode {
				log.Println("failed to get children. err:", err)
			}
			log.Println("empty children")
			return
		}
		log.Printf("> children: #%d", len(children))
		for c := range children {
			log.Println("> ", c)
		}
	}
	return zklock.NewZKLockRegistry(conn, zklock.WithLogger(&dlock.StdLogger{})), beforeCloseFn, nil
}

type zkLogger struct {
}

func (l *zkLogger) Printf(foramt string, v ...interface{}) {
	log.Printf("[ZK] "+foramt, v...)
}
