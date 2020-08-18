package fasthttpexample

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func NewArticleServer() *gin.Engine {
	gin.SetMode(gin.TestMode)
	e := gin.Default()
	e.GET("/articles", func(context *gin.Context) {
		type Article struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			Author  struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"author"`
		}
		articles := []Article{
			{
				Title:   "title1",
				Content: "content1",
				Author: struct {
					Name  string `json:"name"`
					Email string `json:"email"`
				}{
					Name:  "zac",
					Email: "zac@gmail.com",
				},
			}, {
				Title:   "title2",
				Content: "content2",
				Author: struct {
					Name  string `json:"name"`
					Email string `json:"email"`
				}{
					Name:  "evan",
					Email: "evan@gmail.com",
				},
			},
		}
		context.JSON(http.StatusOK, articles)
	})

	e.GET("/sleep/:sec", func(ctx *gin.Context) {
		type PathVariable struct {
			Sleep int `uri:"sec" binding="required"`
		}
		var path PathVariable
		if err := ctx.ShouldBindUri(&path); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}
		time.Sleep(time.Duration(path.Sleep) * time.Second)
		ctx.AbortWithStatusJSON(http.StatusOK, map[string]interface{}{
			"status": "ok",
		})
	})
	return e
}
