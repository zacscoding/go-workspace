package cacheredis

import (
	"fmt"
	"github.com/go-redis/redis/v7"
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

// Redis cmd
// HSet articles t-articles1 article1
// HSet articles t-articles2 article2
// HGet articles t-articles1
// HGet articles t-articles2
func TestHSet(t *testing.T) {
	server, err := NewInmemoryRedisServer()
	assert.NoError(t, err)
	client, err := NewRedisClient(&redis.Options{
		Addr:     server.Addr(),
		Password: "",
	})

	err = client.HSet("articles", "t-"+article1.Title, article1, 1*time.Second)
	assert.NoError(t, err)

	err = client.HSet("articles", "t-"+article2.Title, article2, 1*time.Second)
	assert.NoError(t, err)

	var (
		find1, find2 Article
	)
	err = client.HGet("articles", "t-"+article1.Title, &find1)
	assert.NoError(t, err)
	assert.Equal(t, article1, &find1)

	err = client.HGet("articles", "t-"+article2.Title, &find2)
	assert.NoError(t, err)
	assert.Equal(t, article2, &find2)
}

// Redis cmd
// HSet articles t-articles1 article1
// HKeys articles
func TestHSetFields(t *testing.T) {
	server, err := NewInmemoryRedisServer()
	assert.NoError(t, err)
	client, err := NewRedisClient(&redis.Options{
		Addr:     server.Addr(),
		Password: "",
	})

	err = client.HSet("articles", "t-"+article1.Title, article1, 1*time.Second)
	assert.NoError(t, err)

	keysResult := client.cli.HKeys("articles")
	keys, err := keysResult.Val(), keysResult.Err()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(keys))
	assert.Equal(t, "t-"+article1.Title, keys[0])
}

func TestHGetHDelete(t *testing.T) {
	server, err := NewInmemoryRedisServer()
	assert.NoError(t, err)
	client, err := NewRedisClient(&redis.Options{
		Addr:     server.Addr(),
		Password: "",
	})

	getResult := client.cli.HGet("articles", "t-"+article1.Title)
	value, err := getResult.Val(), getResult.Err()
	assert.Error(t, err)
	assert.Empty(t, value)

	delResult := client.cli.HDel("articles", "t-"+article1.Title)
	rows, err := delResult.Val(), delResult.Err()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), rows)
}

func TestKeysPattern(t *testing.T) {
	server, err := NewInmemoryRedisServer()
	assert.NoError(t, err)
	client, err := NewRedisClient(&redis.Options{
		Addr:     server.Addr(),
		Password: "",
	})
	client.HSet("key1:addr:addr1", "field:field1", "key1:addr:addr1+field:field1", time.Second)
	client.HSet("key1:addr:addr2", "field:field1", "key1:addr:addr2+field:field1", time.Second)
	client.HSet("key1:addr:addr3", "field:field2", "key1:addr:addr3+field:field2", time.Second)
	client.HSet("key1:noaddr:addr3", "field:field2", "key1:noaddr:addr3+field:field2", time.Second)

	keys, err := client.cli.Keys("key1:addr:*").Result()
	assert.NoError(t, err)
	for _, key := range keys {
		fmt.Println("## Key:", key)
	}
}
