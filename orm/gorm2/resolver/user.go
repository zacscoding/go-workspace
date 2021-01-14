package resolver

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name string `gorm:"name;unique"`
}

type UserDB struct {
}
