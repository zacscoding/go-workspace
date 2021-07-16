package socialgraph

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/suite"
	"go-workspace/redis/socialgraph/graph"
	"go-workspace/redis/socialgraph/simple"
	"go.uber.org/zap"
	"testing"
	"time"
)

var (
	user1 = "user1"
	user2 = "user2"
	user3 = "user3"
)

type TestSuite struct {
	suite.Suite
	sgraph graph.SocialGraph
	cli    redis.UniversalClient
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupSuite() {
	cli := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			":7000", ":7001", ":7002", ":7003", ":7004", ":7005",
		},
		ReadTimeout:   3 * time.Second,
		WriteTimeout:  5 * time.Second,
		DialTimeout:   5 * time.Second,
		ReadOnly:      true,
		RouteRandomly: true,
	})
	s.NoError(cli.Ping(context.Background()).Err())
	s.cli = cli
	s.sgraph = simple.NewSimpleSocialGraph(zap.NewNop().Sugar(), cli)

	// remove all keys
	ctx := context.Background()
	err := cli.ForEachMaster(ctx, func(ctx context.Context, client *redis.Client) error {
		keys, err := client.Keys(ctx, "*").Result()
		if err != nil {
			return err
		}
		for _, key := range keys {
			s.cli.Del(ctx, key)
		}
		return nil
	})
	s.NoError(err)
}

func (s *TestSuite) TestFollow() {
	ctx := context.Background()
	follow, err := s.sgraph.IsFollow(ctx, user1, user2)
	s.NoError(err)
	s.False(follow)

	err = s.sgraph.Follow(ctx, user1, user2)
	s.NoError(err)

	follow, err = s.sgraph.IsFollow(ctx, user1, user2)
	s.NoError(err)
	s.True(follow)
}
