package main

import (
	"encoding/json"
	"github.com/asaskevich/EventBus"
	"github.com/fatih/color"
	"go-workspace/basic/event"
	"go-workspace/basic/event/eventbus"
	"strings"
	"time"
)

func main() {
	// runBasic()
	// runProducerByPoc1()
	runProducerByPoc2()
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

	// 3) consumer1 - upper case
	consumer1, err := eventbus.NewMemberConsumer(producer, func(payload *event.MemberPayload) {
		prefix := "[member-consumer-upper]"
		c := color.New(color.FgRed)
		c.Printf("[%s] Name: %s\n", prefix, strings.ToUpper(payload.Name))
	})
	if err != nil {
		panic(err)
	}

	// 4) consumer2 - lower case
	_, err = eventbus.NewMemberConsumer(producer, func(payload *event.MemberPayload) {
		prefix := "[member-consumer-lower]"
		c := color.New(color.FgRed)
		c.Printf("[%s] Name: %s\n", prefix, strings.ToLower(payload.Name))
	})
	if err != nil {
		panic(err)
	}

	// 5) publish event
	producer.Publish(&event.Event{
		Type: event.TypeMember,
		Payload: &event.MemberPayload{
			Name: "ZacCoding",
			Age:  10,
		},
	})

	_ = consumer1.UnSubscribe()

	// 5) publish event
	producer.Publish(&event.Event{
		Type: event.TypeMember,
		Payload: &event.MemberPayload{
			Name: "ZacCoding",
			Age:  10,
		},
	})

	producer.WaitAsync()
}
