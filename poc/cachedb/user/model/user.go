package model

import (
	"encoding/json"
	"gorm.io/plugin/soft_delete"
	"time"
)

const (
	TableNameUser = "users"
)

// User represents database model for users.
type User struct {
	ID        uint                  `gorm:"column:user_id" json:"-"`
	Email     string                `gorm:"column:email" json:"email"`
	Name      string                `gorm:"column:name" json:"username"`
	Password  string                `gorm:"column:password" json:"-"`
	Bio       string                `gorm:"column:bio" json:"bio"`
	Image     string                `gorm:"column:image" json:"image"`
	CreatedAt time.Time             `gorm:"column:created_at" json:"-"`
	UpdatedAt time.Time             `gorm:"column:updated_at" json:"-"`
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;uniqueIndex:idx_name_deleted_at_unix" json:"-"`
}

func (u User) TableName() string {
	return TableNameUser
}

func (u User) String() string {
	b, _ := json.Marshal(&u)
	return string(b)
}
