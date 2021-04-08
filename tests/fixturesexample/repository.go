package fixturesexample

import (
	"gorm.io/gorm"
	"time"
)

type Article struct {
	ID        uint      `gorm:"column:id;primaryKey"`
	Title     string    `gorm:"column:title;size:256"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	Author    string    `gorm:"column:author;size:256"`
}

type Repository struct {
	db *gorm.DB
}

func (r *Repository) FindArticlesByAuthor(author string) ([]*Article, error) {
	var articles []*Article
	return articles, r.db.Debug().Model(new(Article)).Order("id DESC").Find(&articles, "author = ?", author).Error
}
