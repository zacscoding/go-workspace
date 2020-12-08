package distributedlock

import (
	"context"
	"github.com/go-redis/redis/v7"
	"log"
	"testing"
	"time"
)

type hook struct {
}

func (h *hook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	log.Printf("BeforeProcess... name:%s, args:%v", cmd.Name(), cmd.Args())
	return ctx, nil
}

func (h *hook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	log.Printf("AfterProcess... name:%s, args:%v", cmd.Name(), cmd.Args())
	return nil
}

func (h *hook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	log.Printf("BeforeProcessPipeline..")
	for i, cmd := range cmds {
		log.Printf("[CMD-%d] name:%s, args:%v", i, cmd.Name(), cmd.Args())
	}
	return ctx, nil
}

func (h *hook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	log.Printf("AfterProcessPipeline..")
	for i, cmd := range cmds {
		log.Printf("[CMD-%d] name:%s, args:%v", i, cmd.Name(), cmd.Args())
	}
	return nil
}

func TestRedis01(t *testing.T) {
	taskId := "task01"
	redisCli := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	redisCli.AddHook(&hook{})

	locker := NewRedisLockRegistry(redisCli, 1*time.Minute)

	log.Println("Try to acquire a lock")
	acquire := locker.TryLockWithTimeout(taskId, 1*time.Second)
	log.Println("Result:", acquire)

	go func() {
		log.Println("Try to acquire lock2")
		acquire2 := locker.TryLockWithTimeout(taskId, 30*time.Second)
		log.Println("acquire2:", acquire2)
	}()

	time.Sleep(time.Second * 10)
	locker.Unlock(taskId)
	time.Sleep(time.Second * 30)
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
