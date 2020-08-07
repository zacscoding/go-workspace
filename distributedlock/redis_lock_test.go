package distributedlock

import (
	"context"
	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v7"
	"log"
	"testing"
	"time"
)

func Test01(t *testing.T) {
	redisCli := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	locker := redislock.New(redisCli)
	defer redisCli.Close()

	ctx := context.Background()
	taskId := "task01"
	lock, err := locker.Obtain(taskId, time.Second, &redislock.Options{
		Context: ctx,
	})
	if err != nil {
		panic(err)
	}
	log.Println("Success to obtain lock")
	log.Println("key:", lock.Key(), ", metadata:", lock.Metadata(), ", token:", lock.Token())

	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				ttl, err := lock.TTL()
				if err != nil {
					log.Println("TTL() error:", err.Error())
				} else {
					log.Println("TTL() :", ttl)
					err2 := lock.Refresh(1*time.Second, &redislock.Options{
						Context: ctx,
					})
					if err2 != nil {
						log.Println("Refresh() error:", err2.Error())
					}
				}
				//newLock, err := locker.Obtain(taskId, time.Second, &redislock.Options{
				//	Context: context.Background(),
				//})
				//if err != nil {
				//	log.Println("failed to obtain")
				//} else {
				//	log.Println("Success to obtain lock")
				//	log.Println("key:", lock.Key(), ", metadata:", lock.Metadata(), ", token:", lock.Token())
				//	lock = newLock
				//}
			}
		}
	}()

	time.Sleep(1 * time.Minute)
}
