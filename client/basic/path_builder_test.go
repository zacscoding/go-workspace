package basic

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewUriBuilder(t *testing.T) {
	tests := []struct {
		RawUrl  string
		Queries map[string][]string
		Expect  string
	}{
		{
			RawUrl: "http://localhost:8080/path1/path2",
			Queries: map[string][]string{
				"q1": {
					"q1v1", "q1v2",
				},
				"q2": {
					"q2v1",
				},
			},
			Expect: "http://localhost:8080/path1/path2?q1=q1v1&q1=q1v2&q2=q2v1",
		}, {
			RawUrl: "http://localhost:8080",
			Expect: "http://localhost:8080",
		},
	}

	for i, test := range tests {
		builder, err := NewUriBuilderFrom(test.RawUrl)
		if err != nil {
			t.Errorf("%d expected nil error, got %v", i, err)
			t.Fail()
		}

		for key, values := range test.Queries {
			for _, value := range values {
				builder.WithQuery(key, value)
			}
		}

		result := builder.String()
		if result != test.Expect {
			t.Errorf("[#%d] expected: %s, got: %s", i, test.Expect, result)
			t.Fail()
			continue
		}
	}
}

func TestNewUriBuilderFrom(t *testing.T) {
	tests := []struct {
		RawUrl string
		Msg    string
	}{
		{
			RawUrl: "it is not raw url",
			Msg:    "asdfasdf",
		},
	}

	for _, test := range tests {
		_, err := NewUriBuilderFrom(test.RawUrl)
		if err == nil {
			t.Errorf("expected error with %s, got nil", test.Msg)
			t.Fail()
			continue
		}
		if !strings.Contains(err.Error(), test.Msg) {
			t.Errorf("expected: %s, got :%s", test.Msg, err.Error())
			t.Fail()
		}
	}
}

func TestBuilder(t *testing.T) {
	b := NewUriBuilder("http", "localhost:8080", "path1/path2")
	uri := b.WithQuery("q1", "v1").
		WithQuery("q2", "v2").
		WithQuery("q3", "v3").
		WithQuery("", "v3").String()
	fmt.Println(uri)
}
