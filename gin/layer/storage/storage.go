package storage

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go-workspace/gin/layer/model"
)

type ArticleStorage interface {
	SaveArticle(ctx *gin.Context, article *model.Article) (*model.Article, error)
	GetArticles(ctx *gin.Context) ([]*model.Article, error)
	GetArticlesByTitle(ctx *gin.Context, title string) ([]*model.Article, error)
}

type articleStorage struct {
	db *gorm.DB
}

func NewArticleStorage() (ArticleStorage, error) {
	db, err := gorm.Open("mysql", "root:password@tcp(127.0.0.1:13306)/my_database?charset=utf8&parseTime=True")
	if err != nil {
		return nil, err
	}

	db.DropTableIfExists(&model.Article{})
	db.CreateTable(&model.Article{})
	db.LogMode(true)
	return &articleStorage{db: db}, nil
}

func (s *articleStorage) SaveArticle(ctx *gin.Context, article *model.Article) (*model.Article, error) {
	db, ok := getDatabaseFromContext(ctx)
	if !ok {
		db = s.db
	}
	if err := db.Save(article).Error; err != nil {
		return nil, err
	}
	return article, nil
}

func (s *articleStorage) GetArticles(ctx *gin.Context) ([]*model.Article, error) {
	db, ok := getDatabaseFromContext(ctx)
	if !ok {
		db = s.db
	}
	var articles []*model.Article
	if err := db.Find(&articles).Error; err != nil {
		return []*model.Article{}, err
	}
	return articles, nil
}

func (s *articleStorage) GetArticlesByTitle(ctx *gin.Context, title string) ([]*model.Article, error) {
	db, ok := getDatabaseFromContext(ctx)
	if !ok {
		db = s.db
	}
	var articles []*model.Article
	if err := db.Where("title = ?", title).Find(&articles).Error; err != nil {
		return []*model.Article{}, err
	}
	return articles, nil
}

func getDatabaseFromContext(ctx *gin.Context) (*gorm.DB, bool) {
	if ctx == nil {
		return nil, false
	}
	db, ok := ctx.Get("db")
	if !ok {
		return nil, ok
	}
	d, ok := db.(*gorm.DB)
	return d, ok
}
