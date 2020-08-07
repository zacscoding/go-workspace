package jsonexample

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
)

const (
	TypeEvent1 = "event1"
	TypeEvent2 = "event2"
)

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Event1Payload struct {
	Name string `json:"name"`
}

type Event2Payload struct {
	Timestamp int `json:"timestamp"`
}

func (e *Event) UnmarshalJSON(data []byte) error {
	return unmarshal1(e, data)
}

// use gjson
func unmarshal1(e *Event, data []byte) error {
	typeResult := gjson.GetBytes(data, "type")
	if !typeResult.Exists() {
		return nil
	}
	typeValue, ok := typeResult.Value().(string)
	if !ok {
		return errors.New("type must be string")
	}
	e.Type = typeValue

	switch e.Type {
	case TypeEvent1:
		e.Payload = &Event1Payload{}
	case TypeEvent2:
		e.Payload = &Event2Payload{}
	default:
		return errors.New("unknown type:" + e.Type)
	}

	payloadResult := gjson.GetBytes(data, "payload")
	if !payloadResult.Exists() {
		return nil
	}
	return json.Unmarshal([]byte(payloadResult.Raw), e.Payload)
}

func (e Event1Payload) String() string {
	return fmt.Sprintf("Event1Payload{name:%s}", e.Name)
}

func (e Event2Payload) String() string {
	return fmt.Sprintf("Event2Payload{timestamp:%d}", e.Timestamp)
}
