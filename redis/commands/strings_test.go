package commands

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

func (s *RedisSuite) TestGetSet() {
	ctx := context.Background()
	k, v := "testgetset:key1", "my value"

	if exists, _ := s.cli.Exists(ctx, k).Result(); exists != 0 {
		s.NoError(s.cli.Del(ctx, k).Err())
	}
	// ========================
	// Set without TTL.

	res, err := s.cli.Set(ctx, k, v, 0).Result()
	s.NoError(err)
	s.T().Logf("set result: %s", res)

	// ========================
	// Get value.
	find, err := s.cli.Get(ctx, k).Result()
	s.NoError(err)
	s.Equal(v, find)

	ttl, err := s.cli.PTTL(ctx, k).Result()
	s.NoError(err)
	s.EqualValues(time.Duration(-1), ttl)

	// ========================
	// Delete a key
	deleted, err := s.cli.Del(ctx, k).Result()
	s.NoError(err)
	s.EqualValues(1, deleted)

	// ========================
	// Set with TTL(ex)
	res, err = s.cli.Set(ctx, k, v, time.Second*5).Result()
	s.NoError(err)

	// ========================
	// TTL
	ttl, err = s.cli.PTTL(ctx, k).Result()
	s.NoError(err)
	s.LessOrEqual(ttl, time.Second*5)

	time.Sleep(time.Second * 5)

	// ========================
	// GET
	v, err = s.cli.Get(ctx, k).Result()
	log.Println(v)
	s.Equal(redis.Nil, err)
}
