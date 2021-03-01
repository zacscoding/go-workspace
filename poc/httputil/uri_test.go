package httputil

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type SearchOpt struct {
	Filter  string `url:"filter"`
	Keyword string `url:"keyword"`
}

type PageOpt struct {
	Page int `url:"page"`
	Size int `url:"size,omitempty"`
}

type Sling struct {
	rawURL string
}

func Test1(t *testing.T) {
	builder := NewUriBuilder("http://localhost:8800")
	rawURL, err := builder.Path("/v1/api/search").
		QueryStruct(&SearchOpt{
			Filter:  "articles",
			Keyword: "",
		}).
		QueryStruct(&PageOpt{
			Page: 10,
			Size: 0,
		}).
		Query("id", "32", false).
		Query("age", "", true).
		Query("author", "", false).
		ToString()
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8800/v1/api/search?author=&filter=articles&id=32&keyword=&page=10", rawURL)
}
