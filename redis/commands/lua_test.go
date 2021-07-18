package commands

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"log"
	"strconv"
	"time"
)

var (
	ErrLockNotHeld = errors.New("lock not held")
)

var (
	refreshScript = redis.NewScript(`
	if redis.call("get", KEYS[1]) == ARGV[1] then 
		return redis.call("pexpire", KEYS[1], ARGV[2]) 
	else 
		return 0 end
`)
	releaseScript = redis.NewScript(`
	if redis.call("get", KEYS[1]) == ARGV[1] then 
		return redis.call("del", KEYS[1]) 
	else 
		return 0 end
`)
)

func (s *RedisSuite) TestSimple() {
	key := "mylock"
	value := "myvalue"
	// (1) acquire lock
	ok, err := acquire(context.Background(), s.cli, key, value, time.Minute)
	s.NoError(err)
	s.True(ok)

	stopWatchCH := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				d := s.cli.PTTL(context.Background(), key).Val()
				log.Printf("[TTL Check] TTL: %v", d)
			case <-stopWatchCH:
				return
			}
		}
	}()
	log.Println("Success to acquire a lock")
	time.Sleep(time.Second * 5)
	log.Println("Try to refresh lock")
	err = refresh(context.Background(), s.cli, key, value, time.Second*10)
	log.Println("> Err:", err)

	time.Sleep(time.Second * 5)
	log.Println("Try to refresh lock with different value")
	err = refresh(context.Background(), s.cli, key, value+"unknown", time.Second*10)
	log.Println("> Err:", err)

	log.Println("Try to release a lock")
	err = release(context.Background(), s.cli, key, value)
	log.Println("> Err:", err)

	log.Println("Try to release a lock with different value")
	err = release(context.Background(), s.cli, key, value+"unknown")
	log.Println("> Err:", err)
}

func acquire(ctx context.Context, cli redis.UniversalClient, key, value string, ttl time.Duration) (bool, error) {
	return cli.SetNX(ctx, key, value, ttl).Result()
}

func refresh(ctx context.Context, cli redis.UniversalClient, key, value string, ttl time.Duration) error {
	ttlVal := strconv.FormatInt(int64(ttl/time.Millisecond), 10)
	status, err := refreshScript.Run(ctx, cli, []string{key}, value, ttlVal).Result()
	if err != nil {
		return err
	} else if status == int64(1) {
		return nil
	}
	return ErrLockNotHeld
}

func release(ctx context.Context, cli redis.UniversalClient, key, value string) error {
	res, err := releaseScript.Run(ctx, cli, []string{key}, value).Result()
	if err == redis.Nil {
		return ErrLockNotHeld
	} else if err != nil {
		return err
	}
	if v, ok := res.(int64); !ok || v != 1 {
		return nil
	}
	return ErrLockNotHeld
}
