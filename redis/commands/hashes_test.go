package commands

import (
	"context"
	"time"
)

func (s *RedisSuite) TestHSetHGet() {
	ctx := context.Background()
	k := "hsetgetkey"
	field1, value1 := "name", "zacscoding"
	field2, value2 := "email", "zacscoding@gmail.com"
	field3, value3 := "age", "20"
	if exists, _ := s.cli.Exists(ctx, k).Result(); exists != 0 {
		s.cli.Del(ctx, k)
	}

	// ========================
	// HSet without TTL.
	affected, err := s.cli.HSet(ctx, k, field1, value1, field2, value2).Result()
	s.NoError(err)
	s.EqualValues(2, affected)

	// ========================
	// HGet to find the specified field value
	v, err := s.cli.HGet(ctx, k, field1).Result()
	s.NoError(err)
	s.Equal(value1, v)

	// ========================
	// HGetAll to find all field values.
	res, err := s.cli.HGetAll(ctx, k).Result()
	s.NoError(err)
	s.Len(res, 2)
	s.Equal(value1, res[field1])
	s.Equal(value2, res[field2])

	// ========================
	// Delete a key
	affected, err = s.cli.Del(ctx, k).Result()
	s.NoError(err)
	s.EqualValues(1, affected)

	// ========================
	// HSet with TTL.
	pipe := s.cli.Pipeline()
	pipe.HSet(ctx, k, field1, value1, field2, value2)
	pipe.Expire(ctx, k, time.Second*5)
	cmds, err := pipe.Exec(ctx)
	s.NoError(err)
	for _, cmd := range cmds {
		s.T().Logf("cmd: %v", cmd)
	}

	// ========================
	// TTL
	ttl, err := s.cli.PTTL(ctx, k).Result()
	s.NoError(err)
	s.LessOrEqual(ttl, time.Second*5)

	// ========================
	// HSet
	time.Sleep(time.Second)
	err = s.cli.HSet(ctx, k, field3, value3).Err()
	s.NoError(err)

	newTTL, err := s.cli.PTTL(ctx, k).Result()
	s.NoError(err)
	s.Less(newTTL, ttl)
}
