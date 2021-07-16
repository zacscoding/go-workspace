package simple

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go-workspace/redis/socialgraph/graph"
	"go.uber.org/zap"
)

const (
	prefix = "simple"
)

// SocialGraph conforms to the socialgraph.SocialGraph interface for providing social graphs.
type SocialGraph struct {
	logger *zap.SugaredLogger
	cli    redis.UniversalClient
}

// NewSimpleSocialGraph creates a new simple socialgraph.SocialGraph with given redis.
func NewSimpleSocialGraph(logger *zap.SugaredLogger, cli redis.UniversalClient) graph.SocialGraph {
	return &SocialGraph{
		logger: logger,
		cli:    cli,
	}
}

func (s *SocialGraph) Follow(ctx context.Context, userID, followingID string) error {
	var (
		followerKey  = s.followerKey(followingID)
		followingKey = s.followingKey(userID)
	)
	s.logger.Infow("try to save a following relation", "userID", userID, "followingID", followingKey)

	pipe := s.cli.Pipeline()
	pipe.SAdd(ctx, followerKey, userID)
	pipe.SAdd(ctx, followingKey, followingID)
	results, err := pipe.Exec(ctx)

	if err != nil {
		s.logger.Errorw("failed to save the following releation", "err", err)
		return err
	}
	if results[0].(*redis.IntCmd).Val() != 1 {
		s.logger.Errorw("already follow user", "userID", userID, "followingID", followingID)
		return graph.ErrAlreadyUnFollow
	}
	return nil
}

func (s *SocialGraph) UnFollow(ctx context.Context, userID, followingID string) error {
	var (
		followerKey  = s.followerKey(followingID)
		followingKey = s.followingKey(userID)
	)
	s.logger.Infow("try to remove following relation", "userID", userID, "followingID", followingKey)

	pipe := s.cli.Pipeline()
	pipe.SRem(ctx, followerKey, userID)
	pipe.SRem(ctx, followingKey, followingID)
	results, err := pipe.Exec(ctx)

	if err != nil {
		s.logger.Errorw("failed to remove the following releation", "err", err)
		return err
	}
	if results[0].(*redis.IntCmd).Val() != 1 {
		s.logger.Errorw("already unfollow user", "userID", userID, "followingID", followingID)
		return graph.ErrAlreadyUnFollow
	}
	return nil
}

func (s *SocialGraph) IsFollow(ctx context.Context, userID, followingID string) (bool, error) {
	var (
		followingKey = s.followingKey(userID)
	)
	s.logger.Infow("try to check following or not", "userID", userID, "followingID", followingID)

	follow, err := s.cli.SIsMember(ctx, followingKey, followingID).Result()
	if err != nil {
		s.logger.Errorw("failed to check following or not", "err", err)
		return false, err
	}
	return follow, nil
}

func (s *SocialGraph) Followers(ctx context.Context, userID string) ([]string, error) {
	var (
		followerKey = s.followerKey(userID)
	)
	s.logger.Infow("try to get followers", "userID", userID)

	followers, err := s.cli.SMembers(ctx, followerKey).Result()
	if err != nil {
		s.logger.Errorw("failed to get followers", "err", err)
		return nil, err
	}
	return followers, nil
}

func (s *SocialGraph) Followings(ctx context.Context, userID string) ([]string, error) {
	var (
		followingKey = s.followingKey(userID)
	)
	s.logger.Infow("try to get followings", "userID", userID)

	followings, err := s.cli.SMembers(ctx, followingKey).Result()
	if err != nil {
		s.logger.Errorw("failed to get followings", "err", err)
		return nil, err
	}
	return followings, nil
}

func (s *SocialGraph) followerKey(userID string) string {
	return fmt.Sprintf("%s.user.follower:%s", prefix, userID)
}

func (s *SocialGraph) followingKey(userID string) string {
	return fmt.Sprintf("%s.user.following:%s", prefix, userID)
}
