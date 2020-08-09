package event

import (
	"encoding/json"
	"github.com/go-playground/assert/v2"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		Val      string
		Expected Event
	}{
		{
			Val: `
			{
			  "type": "member",
			  "payload": {
				"Name": "zaccoding",
				"Age": 15
			  }
			}`,
			Expected: Event{
				Type: TypeMember,
				Payload: &MemberPayload{
					Name: "zaccoding",
					Age:  15,
				},
			},
		}, {
			Val: `
			{
			  "type": "article",
			  "payload": {
				"title": "article1",
				"content": "content1"
			  }
			}`,
			Expected: Event{
				Type: TypeMember,
				Payload: &ArticlePayload{
					Title:   "article1",
					Content: "content1",
				},
			},
		},
	}

	for _, test := range tests {
		var event Event
		err := json.Unmarshal([]byte(test.Val), &event)
		if err != nil {
			t.Errorf("expected nil error, got: %v", err)
			t.Fail()
		}
		if event.Type != event.Type {
			t.Errorf("expected event type:%s, got: %s", test.Expected.Type, event.Type)
			t.Fail()
		}
		assert.Equal(t, test.Expected.Payload, event.Payload)
	}
}
