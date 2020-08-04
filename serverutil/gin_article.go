package serverutil

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type Server struct {
	*gin.Engine
	articles []*Article
}

type Article struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content"`
	Author  struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required"`
	} `json:"author"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewGinArticleServer() *Server {
	gin.SetMode(gin.TestMode)
	s := &Server{
		Engine: gin.Default(),
	}
	for i := 1; i <= 5; i++ {
		title := "title" + strconv.Itoa(i)
		s.articles = append(s.articles, &Article{
			Title:   title,
			Content: "content" + strconv.Itoa(i),
			Author: struct {
				Name  string `json:"name" binding:"required"`
				Email string `json:"email" binding:"required"`
			}{
				Name:  "Name" + strconv.Itoa(i),
				Email: "email" + strconv.Itoa(i) + "@email.com",
			},
		})
	}
	s.sortArticles()

	s.GET("/articles", s.getArticles)
	s.POST("/articles", s.saveArticle)
	s.GET("/articles/:title", s.getArticle)
	s.PUT("/articles/:title", s.updateArticle)
	s.DELETE("/articles/:title", s.deleteArticle)
	return s
}

// getArticles handle "GET /articles?limit=3&offset=3"
func (s *Server) getArticles(ctx *gin.Context) {
	type PageRequest struct {
		Limit  uint `form:"limit,default=5" binding:"omitempty"`
		Offset uint `form:"offset" binding:"omitempty"`
	}
	var pageRequest PageRequest
	if err := ctx.ShouldBind(&pageRequest); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	skip := int(pageRequest.Offset)
	var res []Article
	for _, article := range s.articles {
		if skip != 0 {
			skip--
			continue
		}
		res = append(res, *article)
		if len(res) >= int(pageRequest.Limit) {
			break
		}
	}
	ctx.JSON(http.StatusOK, res)
}

// saveArticle handle "POST /articles" with article body
func (s *Server) saveArticle(ctx *gin.Context) {
	var article Article
	if err := ctx.ShouldBindJSON(&article); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}
	if idx, _ := s.articleIndexOf(article.Title); idx != -1 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "duplicate article title: " + article.Title,
		})
		return
	}
	s.articles = append(s.articles, &article)
	s.sortArticles()
	ctx.JSON(http.StatusOK, map[string]string{
		"status": "created",
	})
}

// getArticle handle "GET /articles/:title"
func (s *Server) getArticle(ctx *gin.Context) {
	type PathVariable struct {
		Title string `uri:"title"`
	}
	var path PathVariable
	if err := ctx.ShouldBindUri(&path); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	idx, article := s.articleIndexOf(path.Title)
	if idx == -1 {
		ctx.AbortWithStatusJSON(http.StatusNotFound, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "no article with title: " + path.Title,
		})
		return
	}
	ctx.JSON(http.StatusOK, article)
}

func (s *Server) updateArticle(ctx *gin.Context) {
	type PathVariable struct {
		Title string `uri:"title"`
	}
	var path PathVariable
	if err := ctx.ShouldBindUri(&path); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	idx, article := s.articleIndexOf(path.Title)
	if idx == -1 {
		ctx.AbortWithStatusJSON(http.StatusNotFound, ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "no article with title: " + path.Title,
		})
		return
	}

	type UpdateArticle struct {
		Content string `json:"content"`
	}
	var update UpdateArticle
	if err := ctx.ShouldBindJSON(&update); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	article.Content = update.Content
	ctx.JSON(http.StatusOK, article)
}

func (s *Server) deleteArticle(ctx *gin.Context) {
	type PathVariable struct {
		Title string `uri:"title"`
	}
	var path PathVariable
	if err := ctx.ShouldBindUri(&path); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	idx, _ := s.articleIndexOf(path.Title)
	if idx == -1 {
		ctx.AbortWithStatusJSON(http.StatusNotFound, ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "no article with title: " + path.Title,
		})
		return
	}

	s.articles = append(s.articles[:idx], s.articles[idx+1:]...)
	ctx.JSON(http.StatusOK, map[string]string{
		"status": "deleted",
	})
}

func (s *Server) articleIndexOf(title string) (int, *Article) {
	for i, article := range s.articles {
		if article.Title == title {
			return i, article
		}
	}
	return -1, nil
}

func (s *Server) sortArticles() {
	sort.Slice(s.articles, func(i, j int) bool {
		return strings.Compare(s.articles[i].Title, s.articles[j].Title) < 0
	})
}
