package jsondecode

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnmarshalJSON(t *testing.T) {
	// given
	data := `
	{
	  "items": [
		{
		  "type": "article",
		  "title": "this is article title"
		},
		{
		  "type": "comment",
		  "content": "this is comment's content"
		}
	  ],
	  "_link": {
		"prevPage": "?page=1&size=5",
		"nextPage": "?page=3&size=5"
	  }
	}`
	// when
	var items IterableItems
	err := json.Unmarshal([]byte(data), &items)

	// then
	assert.NoError(t, err)
	assert.Len(t, items.Items, 2)
	article, ok := items.Items[0].(*Article)
	assert.True(t, ok)
	assert.Equal(t, "article", article.GetType())
	assert.Equal(t, "this is article title", article.Title)
	comment, ok := items.Items[1].(*Comment)
	assert.True(t, ok)
	assert.Equal(t, "comment", comment.GetType())
	assert.Equal(t, "this is comment's content", comment.Content)
	assert.Equal(t, "?page=1&size=5", items.Link.PrevPage)
	assert.Equal(t, "?page=3&size=5", items.Link.NextPage)
}

func TestEmptyItems(t *testing.T) {
	cases := []struct {
		Desc string
		Data string
	}{
		{
			Desc: "Items Key Not Found",
			Data: `{
			  "_link": {
				"prevPage": "?page=1&size=5",
				"nextPage": "?page=3&size=5"
			  }
			}`,
		},
		{
			Desc: "Items Empty",
			Data: `{
			  "items": [],
			  "_link": {
				"prevPage": "?page=1&size=5",
				"nextPage": "?page=3&size=5"
			  }
			}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Desc, func(t *testing.T) {
			var items IterableItems
			// when
			err := json.Unmarshal([]byte(tc.Data), &items)
			// then
			assert.NoError(t, err)
			assert.NotNil(t, items.Items)
			assert.Empty(t, items.Items)
		})
	}

}
