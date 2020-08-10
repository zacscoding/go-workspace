package eventbus

import "go-workspace/basic/event"

type MemberEventCallback func(payload *event.MemberPayload)
type ArticleEventCallback func(payload *event.ArticlePayload)

type Consumer interface {
	GetTopic() string

	Callback() interface{}

	UnSubscribe() error
}

type MemberConsumer struct {
	p        *Producer
	callback interface{}
}

func (m *MemberConsumer) GetTopic() string {
	return TopicMember
}

func (m *MemberConsumer) Callback() interface{} {
	return m.callback
}

func (m *MemberConsumer) UnSubscribe() error {
	return m.p.UnRegister(m)
}

func NewMemberConsumer(p *Producer, callback MemberEventCallback) (*MemberConsumer, error) {
	consumer := MemberConsumer{
		p: p,
		callback: func(v interface{}) {
			e, ok := v.(*event.Event)
			if !ok {
				return
			}
			p, ok := e.Payload.(*event.MemberPayload)
			if !ok {
				return
			}
			callback(p)
		},
	}
	if err := p.Register(&consumer); err != nil {
		return nil, err
	}
	return &consumer, nil
}
