package commands

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

func (s *RedisSuite) TestPublish() {
	//s.cli.AddHook(redishooks.NewLoggingHook(redishooks.LoggingHookParams{
	//	AfterProcess:         true,
	//	AfterProcessPipeline: true,
	//}))
	var (
		ctx           = context.Background()
		ch            = "mytopic.key1"
		totalMessages = 3
		received      = int32(0)
		wg            = sync.WaitGroup{}
	)
	// ========================
	// Publish message without subscribers.
	subscribes, err := s.cli.Publish(ctx, ch, "message-first").Result()
	log.Printf("[Publisher] publish message without subscribers. #subscribers: %d, err: %v", subscribes, err)

	// ========================
	// Subscribe given channel.
	pubsub := s.cli.Subscribe(ctx, ch)
	wg.Add(1)
	ready := make(chan struct{}, 1)
	go func() {
		defer wg.Done()
		ch := pubsub.Channel()
		ready <- struct{}{}
		for {
			select {
			case m := <-ch:
				log.Println("[Subscriber] receive message.", m)
				if atomic.AddInt32(&received, 1) == int32(totalMessages) {
					log.Println("[Subscriber] terminate subscriber")
					return
				}
			}
		}
	}()
	<-ready

	// ========================
	// Publish messages with 3 times.
	for i := 0; i < totalMessages; i++ {
		if i != 0 {
			time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)
		}
		subscribes, err := s.cli.Publish(ctx, ch, fmt.Sprintf("message-%d", i)).Result()
		if err != nil {
			log.Println("[Publisher] failed to subscribe message. err:", err)
			continue
		}
		log.Println("[Publisher] success to publish message. #subscribes:", subscribes)
	}

	// ========================
	// Shutting down subscriber.
	wg.Wait()
	log.Println("#received messages:", received)

	// ========================
	// Publish one message without unsubscribe.
	subscribes, err = s.cli.Publish(ctx, ch, "message-without-unsubscribe").Result()
	log.Printf("[Publisher] publish message after terminating goroutines but not unsubscribe #subscribers: %d, err: %v",
		subscribes, err)

	// ========================
	// Publish one message after unsubscription
	err = pubsub.Unsubscribe(ctx, ch)
	log.Println("unsubscribe. err:", err)
	subscribes, err = s.cli.Publish(ctx, ch, "message-after-unsubscribe").Result()
	log.Printf("[Publisher] publish message after unsubscribe #subscribers: %d, err: %v",
		subscribes, err)

	//2021/07/18 16:15:14 [Publisher] publish message without subscribers. #subscribers: 0, err: <nil>
	//2021/07/18 16:15:14 [Publisher] success to publish message. #subscribes: 1
	//2021/07/18 16:15:14 [Subscriber] receive message. Message<mytopic.key1: message-0>
	//2021/07/18 16:15:16 [Publisher] success to publish message. #subscribes: 1
	//2021/07/18 16:15:16 [Subscriber] receive message. Message<mytopic.key1: message-1>
	//2021/07/18 16:15:18 [Subscriber] receive message. Message<mytopic.key1: message-2>
	//2021/07/18 16:15:18 [Subscriber] terminate subscriber
	//2021/07/18 16:15:18 [Publisher] success to publish message. #subscribes: 1
	//2021/07/18 16:15:18 #received messages: 3
	//2021/07/18 16:15:18 [Publisher] publish message after terminating goroutines but not unsubscribe #subscribers: 1, err: <nil>
	//2021/07/18 16:15:18 unsubscribe. err: <nil>
	//2021/07/18 16:15:18 [Publisher] publish message after unsubscribe #subscribers: 0, err: <nil>
}
