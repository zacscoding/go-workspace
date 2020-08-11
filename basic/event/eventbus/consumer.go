package eventbus

import "go-workspace/basic/event"

type MemberEventCallback func(payload *event.MemberPayload)
type ArticleEventCallback func(payload *event.ArticlePayload)

type Consumer interface {
	GetTopic() string

	Callback() interface{}

	UnSubscribe() error
}
