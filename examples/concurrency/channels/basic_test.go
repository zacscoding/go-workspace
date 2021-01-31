package channels

import (
	"fmt"
	log "log"
	"math/rand"
	"testing"
	"time"
)

func Test1(t *testing.T) {
	var (
		ch             = make(chan string)
		doneProducer   = make(chan struct{})
		doneSubscriber = make(chan struct{})
	)

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case message, more := <-ch:
				if !more {
					log.Println("no more message!")
					doneSubscriber <- struct{}{}
					return
				}
				log.Println("Receive :", message)
			case <-ticker.C:
				log.Println("Do batch..")
			}
		}
	}()

	go func() {
		for i := 0; i < 10; i++ {
			message := fmt.Sprintf("Message-%d", i)
			sleep := rand.Intn(5)
			log.Printf("%s will sleep %d secs\n", message, sleep)
			time.Sleep(time.Duration(sleep) * time.Second)
			ch <- message
		}
		doneProducer <- struct{}{}
	}()
	<-doneProducer
	log.Println("Complete to produce")
	time.Sleep(time.Second * 30)
	close(ch)
	<-doneSubscriber
}
