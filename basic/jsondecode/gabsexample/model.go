package gabsexample

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Jeffail/gabs/v2"
)

type IterableItems struct {
	Items []Item `json:"items"`
	Link  Link   `json:"_link"`
}

func (i *IterableItems) UnmarshalJSON(data []byte) error {
	parsed, err := gabs.ParseJSON(data)
	if err != nil {
		return err
	}
	err = json.Unmarshal(parsed.Path("_link").Bytes(), &i.Link)
	if err != nil {
		return err
	}
	items := parsed.S("items").Children()
	if len(items) == 0 {
		i.Items = make([]Item, 0)
		return nil
	}
	for _, item := range items {
		dtype, ok := item.Path("type").Data().(string)
		if !ok {
			return errors.New("require type in items")
		}
		var item0 Item
		switch dtype {
		case "article":
			item0 = &Article{}
		case "comment":
			item0 = &Comment{}
		default:
			return errors.New("unknown type:" + dtype)
		}
		err := json.Unmarshal(item.Bytes(), &item0)
		if err != nil {
			return err
		}
		i.Items = append(i.Items, item0)
	}
	return nil
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
