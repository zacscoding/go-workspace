package archive

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"time"
)

type EmbeddedUser struct {
	gorm.Model
	Name string `json:"name" gorm:"unique"`
}

func (u *EmbeddedUser) String() string {
	return fmt.Sprintf("EmbeddedUser{ID:%d, CreatedAt:%v, UpdatedAt:%v, DeletedAt:%v, Name:%s}",
		u.ID, u.CreatedAt, u.UpdatedAt, u.DeletedAt, u.Name)
}

type Sol1User struct {
	ID            uint      `json:"id" gorm:"primarykey"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Name          string    `gorm:"size:256;uniqueIndex:idx_name_deleted_at_unix"`
	DeletedAtUnix int64     `gorm:"uniqueIndex:idx_name_deleted_at_unix"`
}

func (u *Sol1User) String() string {
	return fmt.Sprintf("Sol1User{ID:%d, CreatedAt:%v, UpdatedAt:%v, DeletedAtUnix:%d, Name:%s}",
		u.ID, u.CreatedAt, u.UpdatedAt, u.DeletedAtUnix, u.Name)
}

type Sol2User struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Name      string    `gorm:"size:256;uniqueIndex:idx_name_active"`
	Active    *bool     `gorm:"uniqueIndex:idx_name_active;default:1"`
}

func (u *Sol2User) String() string {
	return fmt.Sprintf("Sol2User{ID:%d, CreatedAt:%v, UpdatedAt:%v, Active:%v, Name:%s}",
		u.ID, u.CreatedAt, u.UpdatedAt, u.Active, u.Name)
}

func (u *Sol2User) IsActive() bool {
	return u.Active != nil && *u.Active
}

type Sol3User struct {
	ID        uint                  `json:"id" gorm:"primarykey"`
	CreatedAt time.Time             `json:"createdAt"`
	UpdatedAt time.Time             `json:"updatedAt"`
	Name      string                `gorm:"size:256;uniqueIndex:idx_name_deleted_at_unix"`
	DeletedAt soft_delete.DeletedAt `gorm:"uniqueIndex:idx_name_deleted_at_unix"`
}

func (u *Sol3User) String() string {
	return fmt.Sprintf("Sol3User{ID:%d, CreatedAt:%v, UpdatedAt:%v, Name:%s, DeletedAt: %v}",
		u.ID, u.CreatedAt, u.UpdatedAt, u.Name, u.DeletedAt)
}
