package main

import (
	"github.com/asaskevich/EventBus"
	"github.com/fatih/color"
	"go-workspace/basic/event"
	"go-workspace/basic/event/eventbus"
	"log"
	"strings"
	"time"
)

func main() {
	// 1) bus
	bus := EventBus.New()

	// 2) producer
	producer := eventbus.NewProducer(bus)

	// 3. consumer
	memberConsumer1, err := eventbus.NewConsumer(bus, func(e *event.Event) {
		prefix := "[member-consumer-upper]"
		c := color.New(color.FgRed)
		if e.Type != event.TypeMember {
			c.Printf("[%s] skip to consume event: %v\n", prefix, e)
			return
		}

		payload := e.Payload.(*event.MemberPayload)
		c.Printf("[%s] Name: %s\n", prefix, strings.ToUpper(payload.Name))
	})
	if err != nil {
		log.Panic(err)
	}
	defer memberConsumer1.Close()

	memberConsumer2, err := eventbus.NewConsumer(bus, func(e *event.Event) {
		prefix := "[member-consumer-lower]"
		c := color.New(color.FgGreen)
		if e.Type != event.TypeMember {
			c.Printf("[%s] skip to consume event: %v\n", prefix, e)
		}

		payload := e.Payload.(*event.MemberPayload)
		c.Printf("[%s] Name: %s\n", prefix, strings.ToLower(payload.Name))
	})
	defer memberConsumer2.Close()

	articleConsumer, err := eventbus.NewConsumer(bus, func(e *event.Event) {
		prefix := "[article-consumer]"
		c := color.New(color.FgBlue)
		if e.Type != event.TypeArticle {
			c.Printf("[%s] skip to consume event: %v\n", prefix, e)
		}

		payload := e.Payload.(*event.ArticlePayload)
		c.Printf("[%s] Title: %s, Content: %s\n", prefix, payload.Title, payload.Content)
	})
	defer articleConsumer.Close()

	//producer.Publish(&event.Event{
	//	Type: event.TypeMember,
	//	Payload: &event.MemberPayload{
	//		Name: "ZacCoding",
	//		Age:  10,
	//	},
	//})
	//memberConsumer1.Close()

	producer.Publish(&event.Event{
		Type: event.TypeMember,
		Payload: &event.MemberPayload{
			Name: "HelloGoWorld",
			Age:  10,
		},
	})

	//producer.Publish(&event.Event{
	//	Type: event.TypeArticle,
	//	Payload: &event.ArticlePayload{
	//		Title:   "Article1",
	//		Content: "Content",
	//	},
	//})

	time.Sleep(5 * time.Second)
}
