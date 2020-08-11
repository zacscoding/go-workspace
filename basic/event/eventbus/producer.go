package eventbus

import (
	"github.com/asaskevich/EventBus"
	"go-workspace/basic/event"
)

type UnsubscribeFunc func() error

type Producer struct {
	bus EventBus.Bus
}

func (p *Producer) Publish(e *event.Event) {
	topic := "event:"
	switch e.Type {
	case event.TypeMember:
		topic += "member"
	case event.TypeArticle:
		topic += "article"
	default:
		return
	}
	p.bus.Publish(topic, e)
}

// POC 1)

func (p *Producer) Register(consumer Consumer) error {
	return p.bus.Subscribe(consumer.GetTopic(), consumer.Callback())
}

func (p *Producer) RegisterAsync(consumer Consumer, transactional bool) error {
	return p.bus.SubscribeAsync(consumer.GetTopic(), consumer.Callback(), transactional)
}

func (p *Producer) UnRegister(consumer Consumer) error {
	return p.bus.Unsubscribe(consumer.GetTopic(), consumer.Callback())
}

// POC 2)
func (p *Producer) RegisterMemberEvent(callback MemberEventCallback) (UnsubscribeFunc, error) {
	wrappedCallback := func(v interface{}) {
		e, ok := v.(*event.Event)
		if !ok {
			return
		}
		p, ok := e.Payload.(*event.MemberPayload)
		if !ok {
			return
		}
		callback(p)
	}
	err := p.bus.Subscribe(TopicMember, wrappedCallback)
	if err != nil {
		return nil, err
	}
	return func() error {
		return p.bus.Unsubscribe(TopicMember, wrappedCallback)
	}, nil
}

func (p *Producer) WaitAsync() {
	p.bus.WaitAsync()
}

func NewProducer(bus EventBus.Bus) *Producer {
	return &Producer{bus: bus}
}
