package fastarticle

import (
	"github.com/stretchr/testify/suite"
	"go-workspace/serverutil"
	"testing"
)

type ClientSuite struct {
	suite.Suite
	endpoint string
	client   ArticleClient
}

func (s *ClientSuite) SetupTest() {
	server := serverutil.NewGinArticleServer()
	go func() {
		err := server.Run(":3000")
		s.NoError(err)
	}()
	s.endpoint = "http://localhost:3000"
	s.client = NewArticleClient(s.endpoint)
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}

func (s *ClientSuite) TestGetArticles() {
	articles, err := s.client.GetArticles(0, 0)
	s.NoError(err)
	s.Equal(5, len(articles))
}
