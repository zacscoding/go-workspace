package app

import (
	"context"
	"fmt"
	"gorm.io/gorm"
)

type UserStore struct {
	db *gorm.DB
}

func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{db: db}
}

func (us *UserStore) FindAllByLastName(_ context.Context, lastName string, pageable Pageable) (
	[]*User, int64, error) {

	var users []*User
	if err := us.db.Where("last_name = ?", lastName).
		Scopes(AddPageable(pageable, func(s Sort) bool {
			switch s.Property {
			case "email", "first_name":
				return true
			default:
				return false
			}
		})).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	var total int64
	if err := us.db.Model(new(User)).
		Where("last_name = ?", lastName).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func AddPageable(pageable Pageable, sortFilter func(s Sort) bool) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for _, sort := range pageable.Sorts {
			if sortFilter != nil && sortFilter(sort) {
				db = db.Order(fmt.Sprintf("%s %s", sort.Property, sort.Direction))
			}
		}
		return db.Offset(pageable.GetOffset()).Limit(pageable.GetLimit())
	}
}
