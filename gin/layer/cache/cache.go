package cache

import (
	"encoding/json"
	"github.com/go-redis/redis/v7"
	"time"
)

type Cache interface {
	HGet(key, field string, value interface{}) error
	HSet(key, field string, value interface{}) error
}

type Client struct {
	c *redis.Client
}

func NewClient() (*Client, error) {
	c := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	if err := c.Ping().Err(); err != nil {
		return nil, err
	}
	return &Client{c: c}, nil
}

func (c *Client) HGet(key, field string, value interface{}) error {
	result := c.c.HGet(key, field)
	if result.Err() != nil {
		return result.Err()
	}
	err := json.Unmarshal([]byte(result.Val()), value)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) HSet(key, field string, value interface{}) error {
	var v string
	switch value.(type) {
	case string:
		v = value.(string)
	default:
		b, err := json.Marshal(value)
		if err != nil {
			return err
		}
		v = string(b)
	}

	pipe := c.c.Pipeline()
	pipe.HSet(key, field, v)
	pipe.Expire(key, 10*time.Second)
	_, err := pipe.Exec()
	return err
}
