package model

import (
	"github.com/jinzhu/gorm"
)

type Article struct {
	gorm.Model
	Slug        string `gorm:"unique_index; not null"`
	Title       string
	Description string `gorm:"size:2048"`
}
