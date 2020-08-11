package fastarticle

import (
	"encoding/json"
	"go-workspace/client/fastclient"
	"strconv"
)

type ArticleClient interface {
	GetArticles(endpoint string, limit, offset int) ([]*ArticleResult, error)

	GetArticle(endpoint string, title string) (*ArticleResult, error)

	CreateArticle(endpoint string, title, content, authorName, authorEmail string) (*ArticleResult, error)

	UpdateArticle(endpoint string, title, content string) (*ArticleResult, error)

	DeleteArticle(endpoint string, title string) (*StatusResult, error)
}

func NewArticleClient(endpoint string) ArticleClient {
	return &articleClient{
		endpoint: endpoint,
		cli:      fastclient.NewClient(),
		// command:  &fastclient.NoopCommand{},
		command: fastclient.NewCircuitBreakerCommand("article"),
	}
}

// TODO : with config
type articleClient struct {
	endpoint string
	cli      *fastclient.Client
	command  fastclient.Command
}

func (c *articleClient) GetArticles(endpoint string, limit, offset int) ([]*ArticleResult, error) {
	var (
		articles      []*ArticleResult
		errorResponse *ErrorResponse
	)

	err := c.command.Execute(func() error {
		// setup
		// to test fail, setting endpoint
		if endpoint == "" {
			endpoint = c.endpoint
		}
		builder, _ := fastclient.NewUriBuilderFrom(endpoint, "articles")
		if limit > 0 {
			builder.WithQuery("limit", strconv.Itoa(limit))
		}
		if offset > 0 {
			builder.WithQuery("offset", strconv.Itoa(limit))
		}

		// request
		code, body, err := c.cli.FastGet(builder.String(), nil)

		// handle response
		if err != nil {
			return err
		}

		if code.Is2xxSuccessful() {
			if err := json.Unmarshal(body, &articles); err != nil {
				return err
			}
			return nil
		}

		if err := json.Unmarshal(body, &errorResponse); err != nil {
			return err
		}
		errorResponse.StatusCode = code
		return nil
	})

	if err != nil {
		return nil, err
	}
	if errorResponse != nil {
		return nil, errorResponse
	}
	return articles, nil
}

func (c *articleClient) GetArticle(endpoint string, title string) (*ArticleResult, error) {
	panic("")
}

func (c *articleClient) CreateArticle(endpoint string, title, content, authorName, authorEmail string) (*ArticleResult, error) {
	panic("")
}

func (c *articleClient) UpdateArticle(endpoint string, title, content string) (*ArticleResult, error) {
	panic("")
}

func (c *articleClient) DeleteArticle(endpoint string, title string) (*StatusResult, error) {
	panic("")
}
