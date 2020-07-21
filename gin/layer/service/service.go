package service

import (
	"github.com/gin-gonic/gin"
	"go-workspace/gin/layer/cache"
	"go-workspace/gin/layer/model"
	"go-workspace/gin/layer/storage"
	"strconv"
)

type ArticleService interface {
	// GetArticles return article list given optional title with error
	GetArticles(ctx *gin.Context, title string) ([]*model.Article, error)
}

type articleService struct {
	articleStorage storage.ArticleStorage
	cache          cache.Cache
}

func NewArticleService(articleStorage storage.ArticleStorage) (ArticleService, error) {
	a := &articleService{articleStorage: articleStorage}

	// init articles
	for i := 1; i <= 5; i++ {
		for j := 1; j <= 5; j++ {
			article := &model.Article{
				Slug:        "slug-" + strconv.Itoa(i),
				Title:       "title-" + strconv.Itoa(i),
				Description: "description-" + strconv.Itoa(j),
			}
			a.articleStorage.SaveArticle(nil, article)
		}
	}

	return a, nil
}

func (s *articleService) GetArticles(ctx *gin.Context, title string) ([]*model.Article, error) {
	// no cache
	if title == "" {
		return s.articleStorage.GetArticles(ctx)
	}
	key := "getarticles"
	field := "field"

	var results []*model.Article

	err := s.cache.HGet(key, field, &key)
	if err == nil {
		return results, nil
	}

	results, err = s.articleStorage.GetArticles(ctx)
	if err != nil {
		return nil, err
	}
	_ = s.cache.HSet(key, field, results) // ignore to set cache
	return results, nil
}
