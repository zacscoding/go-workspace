package workerpool

import "encoding/json"

type BatchArticle struct {
	Id      int
	Title   string
	Content string
	Author  struct {
		Name string
		Age  string
	}
}

func (a BatchArticle) String() string {
	b, _ := json.Marshal(a)
	return string(b)
}
