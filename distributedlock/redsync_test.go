package distributedlock

import (
	"context"
	"fmt"
	goredislib "github.com/go-redis/redis/v7"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v7"
	"log"
	"sync"
	"testing"
	"time"
)

func TestTemp(t *testing.T) {
	// Create a pool with go-redis (or redigo) which is the pool redisync will
	// use while communicating with Redis. This can also be any pool that
	// implements the `redis.Pool` interface.
	client := goredislib.NewClient(&goredislib.Options{
		Addr: "localhost:6379",
	})
	client.AddHook(&hook{})
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)

	// Create an instance of redisync to be used to obtain a mutual exclusion
	// lock.
	rs := redsync.New(pool)

	// Obtain a new mutex by using the same name for all instances wanting the
	// same lock.
	mutexname := "my-global-mutex"
	mutex := rs.NewMutex(mutexname)

	if err := mutex.Lock(); err != nil {
		fmt.Println("failed to get first lock.", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Second)
		ok, err := mutex.Unlock()
		log.Printf("unlock:%v, %v", ok, err)
	}()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		m := rs.NewMutex(mutexname)
		err := m.LockContext(ctx)
		fmt.Println("result of second mutext:", err)
	}()

	wg.Wait()
}
