package workerpool

import (
	"fmt"
	"github.com/gammazero/workerpool"
	"math/rand"
	"strconv"
	"sync"
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

	var (
		doneChan   = make(chan struct{})
		cancelChan = make(chan struct{})
	)
	wp := workerpool.New(3)

	wg := sync.WaitGroup{}
	wg.Add(len(articles))

	for _, article := range articles {
		article := article
		wp.Submit(func() {
			defer wg.Wait()
			if fetchArticle(article) != nil {
				cancelChan <- struct{}{}
				wp.Stop()
			}
		})
	}

	go func() {
		wg.Wait()
		doneChan <- struct{}{}
	}()

	select {
	case <-doneChan:
		fmt.Println("Complete to work!")
	case <-cancelChan:
		fmt.Println("Canceled work!")
	}
	time.Sleep(10 * time.Second)
}

func fetchArticle(a *BatchArticle) error {
	sleepSec := rand.Intn(3) + 1
	fmt.Printf("Start to get article %s's author. slee: %d secs\n", a, sleepSec)

	time.Sleep(time.Duration(sleepSec) * time.Second)
	if rand.Intn(10) == 1 {
		// fmt.Println("Return force error")
		// return errors.New("force err")
	}
	return nil
}
