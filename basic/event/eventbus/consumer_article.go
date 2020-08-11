package eventbus

import "go-workspace/basic/event"

type ArticleConsumer struct {
	p        *Producer
	callback interface{}
}

func (m *ArticleConsumer) GetTopic() string {
	return TopicArticle
}

func (m *ArticleConsumer) Callback() interface{} {
	return m.callback
}

func (m *ArticleConsumer) UnSubscribe() error {
	return m.p.UnRegister(m)
}

func NewArticleConsumer(p *Producer, callback ArticleEventCallback) (*ArticleConsumer, error) {
	consumer := ArticleConsumer{
		p: p,
		callback: func(v interface{}) {
			e, ok := v.(*event.Event)
			if !ok {
				return
			}
			p, ok := e.Payload.(*event.ArticlePayload)
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
