package custom

import (
	"context"
	"encoding/json"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisClient struct {
	cli *redis.Client
}

func (c *RedisClient) HSet(ctx context.Context, key, field string, v interface{}, ttl time.Duration) error {
	var valueString string
	if _, ok := v.(string); ok {
		valueString = v.(string)
	} else {
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		valueString = string(b)
	}

	pipe := c.cli.Pipeline()
	pipe.HSet(ctx, key, field, valueString)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *RedisClient) HGet(ctx context.Context, key, field string, v interface{}) error {
	ret := c.cli.HGet(ctx, key, field)
	value, err := ret.Val(), ret.Err()
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(value), v)
	if err != nil {
		return err
	}
	return nil
}

func NewRedisClient(opt *redis.Options) (*RedisClient, error) {
	rdb := redis.NewClient(opt)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &RedisClient{cli: rdb}, nil
}

func NewInmemoryRedisServer() (*miniredis.Miniredis, error) {
	s, err := miniredis.Run()
	if err != nil {
		return nil, err
	}
	return s, nil
}
