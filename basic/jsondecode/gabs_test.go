package jsondecode

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test1(t *testing.T) {
	data := `
	{
	  "items": [
		{
		  "type": "article",
		  "title": "this is aricle title"
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
	items, err := decodeIterableItems([]byte(data))
	assert.NoError(t, err)
	for i, item := range items.Items {
		fmt.Printf("[#%d] %s\n", i, item.String())
	}
	fmt.Println("Link:", items.Link)
}

func decodeIterableItems(data []byte) (*IterableItems, error) {
	parsed, err := gabs.ParseJSON(data)
	if err != nil {
		return nil, err
	}
	var ret IterableItems
	items := parsed.S("items").Children()
	if len(items) != 0 {
		for _, item := range items {
			dtype, ok := item.Path("type").Data().(string)
			if !ok {
				return nil, errors.New("require type in items")
			}
			var i Item
			switch dtype {
			case "article":
				i = &Article{}
			case "comment":
				i = &Comment{}
			default:
				return nil, errors.New("unknown type:" + dtype)
			}
			if err := json.Unmarshal(item.Bytes(), i); err != nil {
				return nil, err
			}
			ret.Items = append(ret.Items, i)
		}
	}
	if parsed.ExistsP("_link") {
		err := json.Unmarshal(parsed.Path("_link").Bytes(), &ret.Link)
		if err != nil {
			return nil, err
		}
	}
	return &ret, nil
}

type IterableItems struct {
	Items []Item `json:"items"`
	Link  Link   `json:"_link"`
}

type Link struct {
	PrevPage string `json:"prevPage"`
	NextPage string `json:"nextPage"`
}

type Item interface {
	GetType() string
	String() string
}

type Article struct {
	Type  string `json:"type"`
	Title string `json:"title"`
}

func (a *Article) GetType() string {
	return "article"
}

func (a *Article) String() string {
	return fmt.Sprintf("Article{Type:%s, Title:%s}", a.Type, a.Title)
}

type Comment struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

func (a *Comment) GetType() string {
	return "comment"
}

func (a *Comment) String() string {
	return fmt.Sprintf("Comment{Type:%s, Content:%s}", a.Type, a.Content)
}
