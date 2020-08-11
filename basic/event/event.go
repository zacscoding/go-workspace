package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
)

type Type string

const (
	// type of events
	TypeMember  = Type("member")
	TypeArticle = Type("article")
)

var (
	// errors
	ErrInvalidType    = errors.New("invalid event type")
	ErrInvalidPayload = errors.New("invalid payload")
)

type Event struct {
	Type    Type        `json:"type"`
	Payload interface{} `json:"payload"`
}

type MemberPayload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (p MemberPayload) String() string {
	return fmt.Sprintf("MemberPayload{Name:%s, Age:%d}", p.Name, p.Age)
}

type ArticlePayload struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (p ArticlePayload) String() string {
	return fmt.Sprintf("ArticlePayload{Title:%s, Content:%s}", p.Title, p.Content)
}

func (e *Event) UnmarshalJSON(data []byte) error {
	typeResult := gjson.GetBytes(data, "type")
	if !typeResult.Exists() {
		return ErrInvalidType
	}
	typeValue, ok := typeResult.Value().(string)
	if !ok {
		return ErrInvalidType
	}

	switch Type(typeValue) {
	case TypeMember:
		e.Type = TypeMember
		e.Payload = &MemberPayload{}
	case TypeArticle:
		e.Type = TypeArticle
		e.Payload = &ArticlePayload{}
	default:
		return ErrInvalidType
	}

	payloadResult := gjson.GetBytes(data, "payload")
	if !payloadResult.Exists() {
		return nil
	}
	if err := json.Unmarshal([]byte(payloadResult.Raw), e.Payload); err != nil {
		return ErrInvalidPayload
	}
	return nil
}
