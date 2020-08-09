package distributedlock

import (
	"github.com/go-redis/redis/v7"
	"log"
	"testing"
	"time"
)

func TestRedis01(t *testing.T) {
	taskId := "task01"
	redisCli := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	locker := NewRedisLockRegistry(redisCli, 3*time.Second)
	defer locker.Unlock(taskId)

	log.Println("Try to acquire a lock")
	acquire := locker.TryLockWithTimeout(taskId, 1*time.Second)
	log.Println("Result:", acquire)

	time.Sleep(1 * time.Minute)
}

func TestRedis02(t *testing.T) {
	taskId := "task01"
	redisCli := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	locker := NewRedisLockRegistry(redisCli, 3*time.Second)
	defer locker.Unlock(taskId)

	// First : success
	log.Println("Try to acquire a lock")
	acquire := locker.TryLockWithTimeout(taskId, 1*time.Second)
	log.Println("Result:", acquire)
	if !acquire {
		t.Errorf("expected true, got: false")
		t.Fail()
	}

	// Second : false
	log.Println("Try to acquire a lock")
	acquire = locker.TryLockWithTimeout(taskId, 1*time.Second)
	log.Println("Result:", acquire)
	if acquire {
		t.Errorf("expected false, got: true")
		t.Fail()
	}

	// Release
	log.Println("Try to release a lock")
	locker.Unlock(taskId)

	// Third : false
	log.Println("Try to acquire a lock")
	acquire = locker.TryLockWithTimeout(taskId, 1*time.Second)
	log.Println("Result:", acquire)
	if !acquire {
		t.Errorf("expected true, got: false")
		t.Fail()
	}
}
