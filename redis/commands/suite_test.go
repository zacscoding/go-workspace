package commands

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type RedisSuite struct {
	suite.Suite
	cli redis.UniversalClient
}

func TestRunRedisSuite(t *testing.T) {
	suite.Run(t, new(RedisSuite))
}

func (s *RedisSuite) SetupSuite() {
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
	s.NoError(cli.Ping(context.Background()).Err())
	s.cli = cli
}
