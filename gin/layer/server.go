package layer

import (
	"github.com/gin-gonic/gin"
	"go-workspace/gin/layer/service"
)

type SampleServer struct {
	engine         *gin.Engine
	articleService service.ArticleService
}

func NewSampleServer() *SampleServer {
	s := &SampleServer{
		engine: gin.Default(),
	}
	s.engine.GET("/articles", s.getArticles)

	return s
}

func (s *SampleServer) Run() (err error) {
	return s.engine.Run(":3000")
}

func (s *SampleServer) getArticles(c *gin.Context) {

}
