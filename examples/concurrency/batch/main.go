package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/pool.v3"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"time"
)

type BatchArticle struct {
	Id      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Author  struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	} `json:"author"`
}

type Cancelled struct {
	cancelled bool
}

func (a BatchArticle) String() string {
	b, _ := json.Marshal(a)
	return string(b)
}

func main() {
	gin.SetMode(gin.DebugMode)
	e := gin.Default()
	pprof.Register(e)

	e.GET("/articles", func(ctx *gin.Context) {
		type QueryParam struct {
			Size int  `form:"size" binding:"required"`
			Fail bool `form:"fail" binding:"omitempty"`
		}
		var query QueryParam
		if err := ctx.ShouldBind(&query); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}

		articles, err := fetchArticles(query.Size, query.Fail)
		fmt.Println("FetchArticles. Articles:", len(articles), "Err:", err)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, map[string]interface{}{
				"err": err.Error(),
			})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusOK, articles)
	})

	if err := e.Run(":3000"); err != nil {
		log.Fatal(err)
	}
}

func fetchArticles(size int, fail bool) ([]*BatchArticle, error) {
	var articles []*BatchArticle
	for i := 0; i < size; i++ {
		articles = append(articles, &BatchArticle{
			Id:      i,
			Title:   "title-" + strconv.Itoa(i),
			Content: "content-" + strconv.Itoa(i),
		})
	}

	p := pool.NewLimited(3)
	defer p.Close()
	batch := p.Batch()
	cancelled := &Cancelled{}

	go func() {
		for _, article := range articles {
			batch.Queue(fetchArticleAuthor(article, fail, cancelled))
		}
		// DO NOT FORGET THIS OR GOROUTINES WILL DEADLOCK
		// if calling Cancel() it calles QueueComplete() internally
		batch.QueueComplete()
	}()

	success := true
	for article := range batch.Results() {
		if err := article.Error(); err != nil {
			fmt.Println("## Find error " + err.Error())
			success = false
			break
		}
		a := article.Value().(*BatchArticle)
		fmt.Println("Complete >> ", a.String())
	}
	if success {
		return articles, nil
	}
	return nil, errors.New("failed to fetch")
}

func fetchArticleAuthor(article *BatchArticle, fail bool, cancelled *Cancelled) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		sleepSec := rand.Intn(3) + 1
		fmt.Printf("Start to get article %s's author. slee: %d secs\n", article.Title, sleepSec)

		time.Sleep(time.Duration(sleepSec) * time.Second)
		if cancelled.cancelled {
			fmt.Printf("%s worker is cancelled\n", article.Title)
			return nil, nil
		}
		if wu.IsCancelled() {
			fmt.Printf("%s worker is cancelled\n", article.Title)
			// return values not used
			return nil, nil
		}

		if fail && rand.Intn(10) == 1 {
			return nil, errors.New("force err")
		}

		article.Author = struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}{
			Name: "author-" + strconv.Itoa(article.Id),
			Age:  article.Id,
		}

		fmt.Printf("will return %s -> %v\n", article.Title, article.Author)

		return article, nil
	}
}
