package jsonexample

import (
	"encoding/json"
	"fmt"
	"testing"
)

var (
	event1 = `
{
	"type": "event1",
	"payload": {
		"name": "event1"
	}
}`
	event2 = `
{
	"type": "event2",
	"payload": {
		"timestamp": 123
	}
}`
	event3 = `
{
	"type": "event3",
	"payload": {
		"error": "occur"
	}
}`
)

func TestMarshal(t *testing.T) {
	e1 := Event{
		Type: "event1",
		Payload: Event1Payload{
			Name: "zaccoding",
		},
	}

	e1Bytes, _ := json.Marshal(e1)
	fmt.Println("Event1 :", string(e1Bytes))

	e2 := Event{
		Type: "event2",
		Payload: Event2Payload{
			Timestamp: 123,
		},
	}

	e2Bytes, _ := json.Marshal(e2)
	fmt.Println("Event1 :", string(e2Bytes))
}

func TestUnmarshal(t *testing.T) {
	var (
		e1 Event
		e2 Event
		e3 Event
	)

	err := json.Unmarshal([]byte(event1), &e1)
	if err != nil {
		t.Fail()
	}
	if payload, ok := e1.Payload.(*Event1Payload); !ok {
		t.Fail()
	} else if payload.Name != "event1" {
		t.Fail()
	}

	err = json.Unmarshal([]byte(event2), &e2)
	if err != nil {
		t.Fail()
	}
	if payload, ok := e2.Payload.(*Event2Payload); !ok {
		t.Fail()
	} else if payload.Timestamp != 123 {
		t.Fail()
	}

	err = json.Unmarshal([]byte(event3), &e3)
	if err == nil {
		t.Fail()
	}

	fmt.Println(e1)
	fmt.Println(e2)
	fmt.Println(err.Error())
}
