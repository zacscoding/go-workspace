package database

import (
	"context"
	"go-workspace/poc/cachedb/user/model"
	"gorm.io/gorm"
	"log"
	"time"
)

type userDB struct {
	db *gorm.DB
}

func NewUserDB(db *gorm.DB) UserDB {
	return &userDB{
		db: db,
	}
}

func (db *userDB) Save(ctx context.Context, u *model.User) error {
	log.Printf("userDB:Update try to save an user: %s", u)
	if err := db.db.WithContext(ctx).Create(u).Error; err != nil {
		log.Printf("userDB:Save failed to save an user. err: %v", err)
		return err

	}
	return nil
}

func (db *userDB) Update(ctx context.Context, u *model.User) error {
	log.Printf("userDB:Update try to update an user: %s", u)
	result := db.db.WithContext(ctx).
		Model(new(model.User)).
		Where("user_id = ?", u.ID).
		Updates(model.User{
			Email:     u.Email,
			Name:      u.Name,
			Password:  u.Password,
			Bio:       u.Bio,
			Image:     u.Image,
			UpdatedAt: time.Now(),
		})
	if result.Error != nil {
		log.Printf("userDB:Update failed to update an user. err: %v", result.Error)
		return result.Error
	}
	if result.RowsAffected < 1 {
		log.Printf("userDB:Update failed to update an user. err: %v", gorm.ErrRecordNotFound)
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (db *userDB) FindByID(ctx context.Context, userID uint) (*model.User, error) {
	log.Printf("userDB:FindByID try to find an user: %d", userID)

	var u model.User
	if err := db.db.WithContext(ctx).First(&u, "user_id = ?", userID).Error; err != nil {
		log.Printf("userDB:FindByID failed to find an user. err: %v", err)
		return nil, err
	}
	return &u, nil
}

func (db *userDB) FindAll(ctx context.Context) ([]*model.User, error) {
	log.Printf("userDB:FindByID try to find all users")

	var users []*model.User
	if err := db.db.WithContext(ctx).Find(&users).Error; err != nil {
		log.Printf("userDB:FindByID failed to find all users. err: %v", err)
		return nil, err
	}
	return users, nil
}

func (db *userDB) DeleteByID(ctx context.Context, userID uint) error {
	log.Printf("userDB:DeleteByID try to delete an user: %d", userID)

	if err := db.db.WithContext(ctx).Delete(&model.User{ID: userID}).Error; err != nil {
		log.Printf("userDB:DeleteByID failed to delete an user. err: %v", err)
		return err
	}
	return nil
}
