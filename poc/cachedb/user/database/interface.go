package database

import (
	"context"
	"go-workspace/poc/cachedb/user/model"
)

type UserDB interface {
	Save(ctx context.Context, u *model.User) error

	Update(ctx context.Context, u *model.User) error

	FindByID(ctx context.Context, userID uint) (*model.User, error)

	FindAll(ctx context.Context) ([]*model.User, error)

	DeleteByID(ctx context.Context, userID uint) error
}

type Compositor interface {
	GetOriginDB() UserDB
}

type UserCacheEvictor interface {
	EvictUserCache(ctx context.Context, userID uint) error
	EvictUsers(ctx context.Context) error
}
