package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sync"
)

func main() {
	var wait sync.WaitGroup
	wait.Add(2)
	// start server1 (local env)
	go func() {
		e := gin.Default()
		e.GET("/article", func(ctx *gin.Context) {
			log.Println("## Local server")
			getArticleHandle(ctx)
		})
		e.POST("/article", func(ctx *gin.Context) {
			log.Println("## Local server")
			postArticleHandle(ctx)
		})
		if err := e.Run(":3000"); err != nil {
			log.Fatal(err)
		}
		wait.Done()
	}()
	// start server2 (dev env)
	go func() {
		e := gin.Default()
		e.GET("/article", func(ctx *gin.Context) {
			log.Println("## Dev server")
			getArticleHandle(ctx)
		})
		e.POST("/article", func(ctx *gin.Context) {
			log.Println("## Dev server")
			postArticleHandle(ctx)
		})
		if err := e.Run(":3100"); err != nil {
			log.Fatal(err)
		}
		wait.Done()
	}()
	wait.Wait()
}

func getArticleHandle(ctx *gin.Context) {
	log.Println("## Request /article")
	displayHeaders(ctx)
	ctx.JSON(http.StatusOK, gin.H{
		"title":   "My Article!",
		"content": "This is content!!",
		"author": gin.H{
			"name": "Zaccoding",
			"bio":  "It is not working!!!",
		},
	})
}

func postArticleHandle(ctx *gin.Context) {
	type Article struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content"`
		Author  struct {
			Name string `json:"name"`
			Bio  string `json:"bio"`
		} `json:"author"`
	}
	log.Println("## Request /article")
	displayHeaders(ctx)

	var article Article
	if err := ctx.ShouldBind(&article); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	b, _ := json.Marshal(article)
	log.Println("article: ", string(b))
	ctx.JSON(http.StatusOK, gin.H{
		"status": "created",
	})
}

func displayHeaders(ctx *gin.Context) {
	for key, values := range ctx.Request.Header {
		log.Printf("## Header key: %s, values: %v\n", key, values)
	}
}
