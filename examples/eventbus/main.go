package main

import (
	"encoding/json"
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/fatih/color"
	"go-workspace/basic/event"
	"go-workspace/basic/event/eventbus"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	// runBasic()
	// runProducerByPoc1()
	// runProducerByPoc2()
	runProducerByPoc2_1()
}

func runBasic() {
	// 1) bus
	bus := EventBus.New()

	callback1 := func(e *event.Event) {
		prefix := "[consumer1]"
		c := color.New(color.FgRed)
		b, _ := json.Marshal(e)
		c.Printf("[%s]: %s\n", prefix, string(b))
	}
	err := bus.Subscribe("topic1", callback1)
	if err != nil {
		panic(err)
	}

	callback2 := func(e *event.Event) {
		prefix := "[consumer2]"
		c := color.New(color.FgGreen)
		b, _ := json.Marshal(e)
		c.Printf("[%s]: %s\n", prefix, string(b))
	}
	err = bus.Subscribe("topic1", callback2)
	if err != nil {
		panic(err)
	}

	bus.Publish("topic1", &event.Event{
		Type: event.TypeMember,
		Payload: event.MemberPayload{
			Name: "Zaccoding",
			Age:  10,
		},
	})

	bus.Publish("topic1", &event.Event{
		Type: event.TypeArticle,
		Payload: event.ArticlePayload{
			Title:   "Article1",
			Content: "Content1",
		},
	})

	bus.Publish("topic1", "invalid event") // error occur
}

func runProducerByPoc1() {
	// 1) bus
	bus := EventBus.New()

	// 2) producer
	producer := eventbus.NewProducer(bus)

	// 3) register consumer1
	unsubscribe1, err := producer.RegisterMemberEvent(func(payload *event.MemberPayload) {
		prefix := "[member-consumer-upper]"
		c := color.New(color.FgRed)
		c.Printf("%s Name: %s\n", prefix, strings.ToUpper(payload.Name))
	})
	if err != nil {
		panic(err)
	}

	// 4) register consumer2
	_, err = producer.RegisterMemberEvent(func(payload *event.MemberPayload) {
		prefix := "[member-consumer-lower]"
		c := color.New(color.FgRed)
		c.Printf("%s Name: %s\n", prefix, strings.ToLower(payload.Name))
	})
	if err != nil {
		panic(err)
	}

	producer.Publish(&event.Event{
		Type: event.TypeMember,
		Payload: &event.MemberPayload{
			Name: "ZacCoding",
			Age:  10,
		},
	})
	if err := unsubscribe1(); err != nil {
		panic(err)
	}

	producer.Publish(&event.Event{
		Type: event.TypeMember,
		Payload: &event.MemberPayload{
			Name: "HelloGoWorld",
			Age:  10,
		},
	})

	time.Sleep(5 * time.Second)
}

func runProducerByPoc2() {
	// 1) bus
	bus := EventBus.New()

	// 2) Producer
	producer := eventbus.NewProducer(bus)

	// 3) member consumer1 - upper case
	consumer1, err := eventbus.NewMemberConsumer(producer, func(payload *event.MemberPayload) {
		prefix := "member-consumer-upper"
		c := color.New(color.FgRed)
		c.Printf("[%s] Name: %s\n", prefix, strings.ToUpper(payload.Name))
	})
	if err != nil {
		panic(err)
	}

	// 4) member consumer2 - lower case
	_, err = eventbus.NewMemberConsumer(producer, func(payload *event.MemberPayload) {
		prefix := "member-consumer-lower"
		c := color.New(color.FgRed)
		c.Printf("[%s] Name: %s\n", prefix, strings.ToLower(payload.Name))
	})
	if err != nil {
		panic(err)
	}

	// 5) article consumer1
	_, err = eventbus.NewArticleConsumer(producer, func(payload *event.ArticlePayload) {
		prefix := "article-consumer"
		c := color.New(color.FgRed)
		c.Printf("[%s] Title: %s, Content:%s\n", prefix, payload.Title, payload.Content)
	})

	// publish member event
	producer.Publish(&event.Event{
		Type: event.TypeMember,
		Payload: &event.MemberPayload{
			Name: "ZacCoding",
			Age:  10,
		},
	})

	_ = consumer1.UnSubscribe()

	// publish member event
	producer.Publish(&event.Event{
		Type: event.TypeMember,
		Payload: &event.MemberPayload{
			Name: "ZacCoding",
			Age:  10,
		},
	})

	// publish article event
	producer.Publish(&event.Event{
		Type: event.TypeArticle,
		Payload: &event.ArticlePayload{
			Title:   "ArticleTitle",
			Content: "ArticleContent",
		},
	})

	producer.WaitAsync()
}

func runProducerByPoc2_1() {
	// 1) bus
	bus := EventBus.New()

	// 2) Producer
	producer := eventbus.NewProducer(bus)

	// 3) slow consumer
	eventbus.NewMemberConsumer(producer, func(payload *event.MemberPayload) {
		prefix := "member-consumer-lower"
		c := color.New(color.FgRed)

		sleep := rand.Intn(5) + 1
		c.Printf("[%s] receive event. sleep : %d secs\n", prefix, sleep)
		time.Sleep(time.Duration(sleep) * time.Second)
		c.Printf("[%s] event: %s\n", prefix, payload)
	})

	var wait sync.WaitGroup
	for i := 0; i < 50; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			name := "Name-" + strconv.Itoa(rand.Intn(100))
			age := rand.Intn(100)

			started := time.Now()
			producer.Publish(&event.Event{
				Type: event.TypeMember,
				Payload: &event.MemberPayload{
					Name: name,
					Age:  age,
				},
			})
			fmt.Printf("Success to produce: %v\n", time.Since(started))
		}()
	}
	wait.Wait()
}
