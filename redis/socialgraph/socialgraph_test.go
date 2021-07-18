package socialgraph

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
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
	user4 = "user4"
	user5 = "user5"
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

func (s *TestSuite) TestFollow_Fail() {
	ctx := context.Background()
	s.NoError(s.sgraph.Follow(ctx, user1, user2))
	s.NoError(s.sgraph.Follow(ctx, user2, user3))
	s.NoError(s.sgraph.Follow(ctx, user3, user2))
	cases := []struct {
		name        string
		userID      string
		followingID string
		msg         string
	}{
		{
			name:        "already follow",
			userID:      user1,
			followingID: user2,
			msg:         graph.ErrAlreadyFollow.Error(),
		},
	}

	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			err := s.sgraph.Follow(ctx, tc.userID, tc.followingID)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.msg)
		})
	}
}

func (s *TestSuite) TestUnFollow() {
	ctx := context.Background()
	err := s.sgraph.Follow(ctx, user1, user2)
	s.NoError(err)
	follow, err := s.sgraph.IsFollow(ctx, user1, user2)
	s.NoError(err)
	s.True(follow)

	err = s.sgraph.UnFollow(ctx, user1, user2)

	s.NoError(err)
	follow, err = s.sgraph.IsFollow(ctx, user1, user2)
	s.NoError(err)
	s.False(follow)
}

func (s *TestSuite) TestUnFollow_Fail() {
	ctx := context.Background()
	s.NoError(s.sgraph.Follow(ctx, user1, user2))
	s.NoError(s.sgraph.Follow(ctx, user2, user3))
	s.NoError(s.sgraph.Follow(ctx, user3, user2))
	cases := []struct {
		name        string
		userID      string
		followingID string
		msg         string
	}{
		{
			name:        "already unfollow",
			userID:      user2,
			followingID: user1,
			msg:         graph.ErrAlreadyUnFollow.Error(),
		},
	}

	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			err := s.sgraph.UnFollow(ctx, tc.userID, tc.followingID)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.msg)
		})
	}
}

func (s *TestSuite) TestIsFollowBulk() {
	ctx := context.Background()
	err := s.sgraph.Follow(ctx, user1, user2)
	s.NoError(err)
	err = s.sgraph.Follow(ctx, user1, user3)
	s.NoError(err)
	err = s.sgraph.Follow(ctx, user2, user1)
	s.NoError(err)

	res, err := s.sgraph.IsFollowBulk(ctx, user1, []string{user2, user3, user4, user5})

	s.NoError(err)
	s.Len(res, 4)
	s.True(res[user2])
	s.True(res[user3])
	s.False(res[user4])
	s.False(res[user5])
}

func (s *TestSuite) TestFollowers() {
	ctx := context.Background()
	err := s.sgraph.Follow(ctx, user1, user2)
	s.NoError(err)
	err = s.sgraph.Follow(ctx, user3, user2)
	s.NoError(err)
	err = s.sgraph.Follow(ctx, user2, user3)
	s.NoError(err)

	followers, err := s.sgraph.Followers(ctx, user2)

	s.NoError(err)
	s.Len(followers, 2)
	s.Contains(followers, user1)
	s.Contains(followers, user3)
}

func (s *TestSuite) TestFollowings() {
	ctx := context.Background()
	err := s.sgraph.Follow(ctx, user1, user2)
	s.NoError(err)
	err = s.sgraph.Follow(ctx, user1, user3)
	s.NoError(err)
	err = s.sgraph.Follow(ctx, user2, user3)
	s.NoError(err)

	followings, err := s.sgraph.Followings(ctx, user1)

	s.NoError(err)
	s.Len(followings, 2)
	s.Contains(followings, user2)
	s.Contains(followings, user3)
}

func (s *TestSuite) TestRecommandUsers() {
	ctx := context.Background()
	s.NoError(s.sgraph.Follow(ctx, user1, user2))
	s.NoError(s.sgraph.Follow(ctx, user1, user3))
	s.NoError(s.sgraph.Follow(ctx, user4, user1))

	users, err := s.sgraph.RecommandUsers(ctx, user5, user1)

	s.NoError(err)
	s.Len(users, 3)
	s.Contains(users, user2)
	s.Contains(users, user3)
	s.Contains(users, user4)
}
