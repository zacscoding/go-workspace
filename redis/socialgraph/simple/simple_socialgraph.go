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
		return graph.ErrAlreadyFollow
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

func (s *SocialGraph) IsFollowBulk(ctx context.Context, userID string, followingIDs []string) (map[string]bool, error) {
	var (
		followingKey = s.followingKey(userID)
	)
	s.logger.Infow("try to check following or not", "userID", userID, "followingIDs", followingIDs)

	if len(followingIDs) == 0 {
		return make(map[string]bool, 0), nil
	}

	// SMISMEMBER is not supported current version, so use SIsMember with pipeline.
	pipe := s.cli.Pipeline()
	for _, id := range followingIDs {
		pipe.SIsMember(ctx, followingKey, id)
	}
	results, err := pipe.Exec(ctx)
	if err != nil {
		s.logger.Errorw("failed to check following bulk", "err", err)
		return nil, err
	}

	m := make(map[string]bool, len(followingIDs))
	for i, result := range results {
		var (
			follow      = false
			followingID = followingIDs[i]
		)
		if result.Err() == nil && result.(*redis.BoolCmd).Val() {
			follow = true
		}
		m[followingID] = follow
	}
	return m, nil
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

func (s *SocialGraph) RecommandUsers(ctx context.Context, userID string, targetUserID string) ([]string, error) {
	var (
		followerKey  = s.followerKey(targetUserID)
		followingKey = s.followingKey(targetUserID)
	)
	s.logger.Infof("try to find recommand users from %s to %s", targetUserID, userID)

	pipe := s.cli.Pipeline()
	pipe.SMembers(ctx, followerKey)
	pipe.SMembers(ctx, followingKey)
	results, err := pipe.Exec(ctx)
	if err != nil {
		s.logger.Errorw("failed to get recommanded users",
			"followerKey", followerKey,
			"followingKey", followingKey,
			"err", err)
		return nil, err
	}

	var users []string
	for _, r := range results {
		us, err := r.(*redis.StringSliceCmd).Result()
		if err != nil {
			s.logger.Errorw("failed to execute SMembers", "err", err)
			continue
		}
		users = append(users, us...)
	}
	return users, nil
}

func (s *SocialGraph) followerKey(userID string) string {
	return fmt.Sprintf("%s.user.follower:%s", prefix, userID)
}

func (s *SocialGraph) followingKey(userID string) string {
	return fmt.Sprintf("%s.user.following:%s", prefix, userID)
}
