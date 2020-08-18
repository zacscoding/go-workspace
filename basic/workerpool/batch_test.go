package workerpool

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/go-playground/pool.v3"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

type BatchArticle struct {
	Id      int
	Title   string
	Content string
	Author  struct {
		Name string
		Age  string
	}
}

func (a BatchArticle) String() string {
	b, _ := json.Marshal(a)
	return string(b)
}

func TestBatch(t *testing.T) {
	var articles []*BatchArticle
	for i := 0; i < 10; i++ {
		articles = append(articles, &BatchArticle{
			Id:      i,
			Title:   "title-" + strconv.Itoa(i),
			Content: "content-" + strconv.Itoa(i),
		})
	}

	p := pool.NewLimited(3)
	defer p.Close()
	batch := p.Batch()

	go func() {
		for _, article := range articles {
			batch.Queue(fetchArticleAuthor(article))
		}
		// DO NOT FORGET THIS OR GOROUTINES WILL DEADLOCK
		// if calling Cancel() it calles QueueComplete() internally
		batch.QueueComplete()
	}()

	for article := range batch.Results() {
		if err := article.Error(); err != nil {
			fmt.Println("## Find error " + err.Error())
			batch.Cancel()
			break
		}
		a := article.Value().(*BatchArticle)
		fmt.Println("Complete >> ", a.String())
	}
	fmt.Println("Complete to task!!")
}

func fetchArticleAuthor(article *BatchArticle) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		sleepSec := rand.Intn(3) + 1
		fmt.Printf("Start to get article %s's author. slee: %d secs\n", article, sleepSec)

		time.Sleep(time.Duration(sleepSec) * time.Second)
		if wu.IsCancelled() {
			fmt.Printf("%s worker is cancelled\n", article)
			// return values not used
			return nil, nil
		}

		if rand.Intn(10) == 1 {
			return nil, errors.New("force err")
		}

		return article, nil
	}
}
