package cacheredis

import (
	"encoding/json"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v7"
	"time"
)

type RedisClient struct {
	cli *redis.Client
}

func (c *RedisClient) HSet(key, field string, v interface{}, ttl time.Duration) error {
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
	pipe.HSet(key, field, valueString)
	pipe.Expire(key, ttl)
	_, err := pipe.Exec()
	return err
}

func (c *RedisClient) HGet(key, field string, v interface{}) error {
	ret := c.cli.HGet(key, field)
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
	if err := rdb.Ping().Err(); err != nil {
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
