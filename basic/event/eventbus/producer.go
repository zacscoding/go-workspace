package eventbus

import (
	"github.com/asaskevich/EventBus"
	"go-workspace/basic/event"
)

type Producer struct {
	bus EventBus.Bus
}

func (p *Producer) Publish(e *event.Event) {
	topic := string(e.Type)
	p.bus.Publish(topic, e)
}

func NewProducer(bus EventBus.Bus) *Producer {
	return &Producer{bus: bus}
}
