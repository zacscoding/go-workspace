package eventbus

import (
	"github.com/asaskevich/EventBus"
	"go-workspace/basic/event"
)

type Consumer struct {
	bus     EventBus.Bus
	onEvent func(e *event.Event)
}

func NewConsumer(bus EventBus.Bus, onEvent func(e *event.Event)) (*Consumer, error) {
	c := &Consumer{
		bus:     bus,
		onEvent: onEvent,
	}
	err := c.bus.Subscribe(string(event.TypeMember), c.callback)
	if err != nil {
		return nil, err
	}
	err = c.bus.Subscribe(string(event.TypeArticle), c.callback)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Consumer) Close() {
	_ = c.bus.Unsubscribe(string(event.TypeMember), c.callback)
	_ = c.bus.Unsubscribe(string(event.TypeArticle), c.callback)
}

func (c *Consumer) callback(v interface{}) {
	event, ok := v.(*event.Event)
	if !ok {
		return
	}
	c.onEvent(event)
}
