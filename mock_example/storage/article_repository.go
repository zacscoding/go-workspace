package storage

type Article struct {
}

// $ go generate ./...
// //go:generate mockgen -destination=mock_article_repository.go -package=storage -source=article_repository.go
// //go:generate mockery -name ArticleRepository

//go:generate mockery -name ArticleRepository -output ../mocks/mock_article_repository.go -outpkg mock_storage
type ArticleRepository interface {
	GetArticles(limit, offset int) ([]*Article, int, error)

	GetArticle(title string) (*Article, error)

	SaveArticle(article Article) (*Article, error)
}
