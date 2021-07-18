package redsyncexample

import (
	"context"
	"github.com/go-redis/redis"
	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/stretchr/testify/assert"
	"go-workspace/redis/redishooks"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLock(t *testing.T) {
	redis.SetLogger(log.New(os.Stdout, "redis: ", log.LstdFlags|log.Lshortfile))
	client, rs := fixtures()
	client.AddHook(redishooks.NewLoggingHook(redishooks.LoggingHookParams{
		AfterProcess:         true,
		AfterProcessPipeline: true,
	}))
	mutexName := "mylock"

	for i := 0; i < 2; i++ {
		if i != 0 {
			log.Println("Try to acquire a lock from index 2")
		}
		mutex := rs.NewMutex(mutexName, redsync.WithExpiry(time.Minute*5))
		if err := mutex.Lock(); err != nil {
			log.Println("")
		} else {
			go func() {
				time.Sleep(time.Minute * 5)
				mutex.Unlock()
			}()
		}
	}
}

func TestLockConcurrency(t *testing.T) {
	_, rs := fixtures()
	wg := sync.WaitGroup{}
	mutexName := "mylock.conccurency"
	acquireLock := int32(0)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			mutex := rs.NewMutex(mutexName, redsync.WithExpiry(time.Second*10))
			var deadline time.Duration
			if i == 4 {
				deadline = time.Second * 6
			} else {
				deadline = time.Second * 3
			}
			ctx, _ := context.WithTimeout(context.Background(), deadline)
			log.Printf("[Worker-%d] try to acquire a lock.", i)
			if err := mutex.LockContext(ctx); err != nil {
				log.Printf("[Worker-%d] failed to acquire a lock. err: %v", i, err)
				return
			}
			atomic.AddInt32(&acquireLock, 1)
			log.Printf("[Worker-%d] success to acquire a lock", i)
			time.Sleep(time.Second * 5)
			mutex.Unlock()
		}()
	}
	wg.Wait()

	assert.EqualValues(t, 2, acquireLock)
}

func fixtures() (*goredislib.ClusterClient, *redsync.Redsync) {
	client := goredislib.NewClusterClient(&goredislib.ClusterOptions{
		Addrs: []string{
			"localhost:7000", "localhost:7001", "localhost:7002", "localhost:7003", "localhost:7004", "localhost:7005",
		},
		ReadTimeout:   3 * time.Second,
		WriteTimeout:  5 * time.Second,
		DialTimeout:   5 * time.Second,
		ReadOnly:      true,
		RouteRandomly: true,
	})
	pool := goredis.NewPool(client)
	return client, redsync.New(pool)
}
