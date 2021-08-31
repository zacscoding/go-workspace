package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Client struct {
}

func (c *Client) DoWork(prefix string, sleep time.Duration) {
	log.Printf("%s try to work.. with %s", prefix, sleep)
	time.Sleep(sleep)
	log.Printf("%s complete to work", prefix)
}

func main() {
	ch := make(chan *Client, 5)
	for i := 0; i < 5; i++ {
		ch <- &Client{}
	}

	wg := sync.WaitGroup{}
	start := time.Now()
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			c := <-ch
			ch <- c

			c.DoWork(id, time.Second)

		}(fmt.Sprintf("Worker-%d", i+1))
	}
	wg.Wait()
	elapsed := time.Now().Sub(start)
	log.Printf("Complete: %v", elapsed)
}
