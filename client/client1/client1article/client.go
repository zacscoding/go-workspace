package client1article

import (
	"encoding/json"
	"go-workspace/client/client1"
	"strconv"
)

type ArticleClient interface {
	GetArticles(limit, offset int) ([]*ArticleResult, error)

	GetArticle(title string) (*ArticleResult, error)

	CreateArticle(title, content, authorName, authorEmail string) (*ArticleResult, error)

	UpdateArticle(title, content string) (*ArticleResult, error)

	DeleteArticle(title string) (*StatusResult, error)
}

func NewArticleClient(endpoint string) ArticleClient {
	return &articleClient{
		endpoint: endpoint,
		cli:      client1.NewClient(),
		executor: &client1.NoopExecutor{},
	}
}

// TODO : with config
type articleClient struct {
	endpoint string
	cli      *client1.Client
	executor client1.Executor
}

func (c *articleClient) GetArticles(limit, offset int) ([]*ArticleResult, error) {
	articleResults, err := c.executor.Execute(func() (interface{}, error) {
		builder, _ := client1.NewUriBuilderFrom(c.endpoint, "articles")
		if limit > 0 {
			builder.WithQuery("limit", strconv.Itoa(limit))
		}
		if offset > 0 {
			builder.WithQuery("offset", strconv.Itoa(limit))
		}

		code, body, err := c.cli.FastGet(builder.String())

		if err != nil {
			return nil, err
		}
		if code.Is2xxSuccessful() {
			var res []*ArticleResult
			if err := json.Unmarshal(body, &res); err != nil {
				return nil, err
			}
			return res, nil
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return articleResults.([]*ArticleResult), nil
}

func (c *articleClient) GetArticle(title string) (*ArticleResult, error) {
	articleResult, err := c.executor.Execute(func() (interface{}, error) {
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return articleResult.(*ArticleResult), nil
}

func (c *articleClient) CreateArticle(title, content, authorName, authorEmail string) (*ArticleResult, error) {
	articleResult, err := c.executor.Execute(func() (interface{}, error) {
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return articleResult.(*ArticleResult), nil
}

func (c *articleClient) UpdateArticle(title, content string) (*ArticleResult, error) {
	articleResult, err := c.executor.Execute(func() (interface{}, error) {
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return articleResult.(*ArticleResult), nil
}

func (c *articleClient) DeleteArticle(title string) (*StatusResult, error) {
	panic("implement me")
}
