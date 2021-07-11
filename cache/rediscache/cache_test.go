package rediscache

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type Article struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Author  struct {
		Name string `json:"name"`
	} `json:"author"`
}

var (
	article1 = &Article{
		Title:   "article1",
		Content: "content1",
		Author: struct {
			Name string `json:"name"`
		}{
			Name: "user1",
		},
	}
	article2 = &Article{
		Title:   "article2",
		Content: "content2",
		Author: struct {
			Name string `json:"name"`
		}{
			Name: "user2",
		},
	}
)

func TestBasic(t *testing.T) {
	c := newCache(t)
	err := c.Set(&cache.Item{
		Ctx:   context.Background(),
		Key:   fmt.Sprintf("article:%s", article1.Title),
		Value: article1,
		TTL:   time.Minute,
	})
	assert.NoError(t, err)

	var find Article
	err = c.Get(context.Background(), fmt.Sprintf("article:%s", article1.Title), &find)
	assert.NoError(t, err)
	assert.Equal(t, article1, &find)

	var notfound Article
	err = c.Get(context.Background(), "article:unknown", &notfound)
	assert.Error(t, err)
	assert.Equal(t, cache.ErrCacheMiss, err)
}

func TestOnce(t *testing.T) {
	c := newCache(t)
	callCounts := make(map[string]int)
	articleFunc := func(item *cache.Item) (interface{}, error) {
		callCounts[item.Key] = callCounts[item.Key] + 1
		if item.Key == fmt.Sprintf("article:%s", article1.Title) {
			return article1, nil
		}
		if item.Key == fmt.Sprintf("article:%s", article2.Title) {
			return article2, nil
		}
		return nil, errors.New("not found")
	}

	var find1 Article
	err := c.Once(&cache.Item{
		Key:   fmt.Sprintf("article:%s", article1.Title),
		Value: &find1,
		TTL:   time.Minute,
		Do:    articleFunc,
	})
	assert.NoError(t, err)
	assert.Equal(t, article1, &find1)

}

func newCache(t *testing.T) *cache.Cache {
	cli := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			"localhost:7000", "localhost:7001", "localhost:7002", "localhost:7003", "localhost:7004", "localhost:7005",
		},
		ReadTimeout:   3 * time.Second,
		WriteTimeout:  5 * time.Second,
		DialTimeout:   5 * time.Second,
		ReadOnly:      true,
		RouteRandomly: true,
	})
	assert.NoError(t, cli.Ping(context.Background()).Err())

	return cache.New(&cache.Options{
		Redis: cli,
	})
}
