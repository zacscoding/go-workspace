package graph

import (
	"context"
	"errors"
)

var (
	ErrAlreadyFollow   = errors.New("already follow user")
	ErrAlreadyUnFollow = errors.New("already unfollow user")
)

type SocialGraph interface {
	// Follow adds the following relation i.e userID following follinwgID.
	Follow(ctx context.Context, userID, followingID string) error

	// UnFollow removes the following releation i.e userID unfollowing followingID.
	UnFollow(ctx context.Context, userID, followingID string) error

	// IsFollow returns a true if given userID follows the user of followingID, otherwise false.
	IsFollow(ctx context.Context, userID, followingID string) (bool, error)

	// IsFollowBulk returns a result map of given userID follow users.
	IsFollowBulk(ctx context.Context, userID string, followingIDs []string) (map[string]bool, error)

	// Followers returns the followers of given userID.
	Followers(ctx context.Context, userID string) ([]string, error)

	// Followings returns the following list of given userID.
	Followings(ctx context.Context, userID string) ([]string, error)

	// RecommandUsers returns recommanded users for userID.
	// this functions is a simple implement which just collect targetUserID's followers and followings.
	RecommandUsers(ctx context.Context, userID string, targetUserID string) ([]string, error)
}
