package database

import (
	"context"
	"fmt"
	"github.com/go-redis/cache/v8"
	"go-workspace/poc/cachedb/user/model"
	"time"
)

var (
	_ UserCacheEvictor = &userCache{}
	_ Compositor       = &userCache{}
)

type Users []*model.User

type userCache struct {
	delegate UserDB
	cache    *cache.Cache // TODO: implement cache instead of go-redis/cache
	ttl      time.Duration
}

func NewUserCache(c *cache.Cache, delegate UserDB) UserDB {
	return &userCache{
		delegate: delegate,
		cache:    c,
		ttl:      10 * time.Second,
	}
}

func (uc *userCache) Save(ctx context.Context, u *model.User) error {
	if err := uc.delegate.Save(ctx, u); err != nil {
		return err
	}

	_ = uc.putUser(ctx, u)
	_ = uc.evictUsers(ctx)
	return nil
}

func (uc *userCache) Update(ctx context.Context, u *model.User) error {
	if err := uc.delegate.Update(ctx, u); err != nil {
		return err
	}

	_ = uc.putUser(ctx, u)
	_ = uc.evictUsers(ctx)
	return nil
}

func (uc *userCache) FindByID(ctx context.Context, userID uint) (*model.User, error) {
	var (
		find model.User
		item = cache.Item{
			Ctx:   ctx,
			Key:   uc.generateUserKey(userID),
			Value: &find,
			TTL:   uc.ttl,
			Do: func(item *cache.Item) (interface{}, error) {
				return uc.delegate.FindByID(ctx, userID)
			},
		}
	)
	if err := uc.cache.Once(&item); err != nil {
		return nil, err
	}
	return &find, nil
}

func (uc *userCache) FindAll(ctx context.Context) ([]*model.User, error) {
	var (
		users []*model.User
		item  = cache.Item{
			Ctx:   ctx,
			Key:   uc.generateUsersKey(),
			Value: &users,
			TTL:   uc.ttl,
			Do: func(item *cache.Item) (interface{}, error) {
				users, err := uc.delegate.FindAll(ctx)
				if err != nil {
					return nil, err
				}
				return &users, nil
			},
		}
	)
	err := uc.cache.Once(&item)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (uc *userCache) DeleteByID(ctx context.Context, userID uint) error {
	if err := uc.DeleteByID(ctx, userID); err != nil {
		return err
	}

	_ = uc.evictUser(ctx, userID)
	_ = uc.evictUsers(ctx)
	return nil
}

func (uc *userCache) EvictUserCache(ctx context.Context, userID uint) error {
	return uc.evictUser(ctx, userID)
}

func (uc *userCache) EvictUsers(ctx context.Context) error {
	return uc.evictUsers(ctx)
}

func (uc *userCache) GetOriginDB() UserDB {
	return uc.delegate
}

func (uc *userCache) putUsers(ctx context.Context, users []*model.User) error {
	return uc.cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   uc.generateUsersKey(),
		Value: users,
		TTL:   uc.ttl,
	})
}

func (uc *userCache) evictUsers(ctx context.Context) error {
	return uc.cache.Delete(ctx, uc.generateUsersKey())
}

func (uc *userCache) putUser(ctx context.Context, user *model.User) error {
	return uc.cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   uc.generateUserKey(user.ID),
		Value: user,
		TTL:   uc.ttl,
	})
}

func (uc *userCache) evictUser(ctx context.Context, userID uint) error {
	return uc.cache.Delete(ctx, uc.generateUserKey(userID))
}

func (uc *userCache) generateUsersKey() string {
	return "users"
}

func (uc *userCache) generateUserKey(userID uint) string {
	return fmt.Sprintf("users:%d", userID)
}
