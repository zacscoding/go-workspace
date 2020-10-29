package workerpool

import (
	"fmt"
	"github.com/gammazero/workerpool"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBatch2(t *testing.T) {
	var articles []*BatchArticle
	for i := 0; i < 10; i++ {
		articles = append(articles, &BatchArticle{
			Id:      i,
			Title:   "title-" + strconv.Itoa(i),
			Content: "content-" + strconv.Itoa(i),
		})
	}

	wp := workerpool.New(2)
	mutex := sync.Mutex{}
	var errs []error
	hasErr := int32(0)
	for _, article := range articles {
		article := article
		wp.Submit(func() {
			if !atomic.CompareAndSwapInt32(&hasErr, 0, 0) {
				log.Println("skip to fetch article:", article.String())
				return
			}
			if err := fetchArticle(article); err != nil {
				mutex.Lock()
				errs = append(errs, err)
				mutex.Unlock()
				log.Println("Find error..")
				atomic.CompareAndSwapInt32(&hasErr, 0, 1)
			}
		})
	}
	log.Println("complete to submit")
	wp.StopWait()
	if len(errs) == 0 {
		log.Println("complete to stop wait.")
	} else {
		log.Println("complete to stop wait with error")
	}
	time.Sleep(10 * time.Second)
}

func fetchArticle(a *BatchArticle) error {
	sleepSec := rand.Intn(3) + 1
	log.Printf("Start to get article %s's author. slee: %d secs\n", a, sleepSec)

	time.Sleep(time.Duration(sleepSec) * time.Second)
	if rand.Intn(10) == 1 {
		log.Println("Return force error")
		//return errors.New("force err")
		return fmt.Errorf("force error")
	}
	return nil
}
