package client1

import (
	"fmt"
	"go-workspace/serverutil"
	"testing"
)

const (
	scheme = "http"
	host   = "localhost:3000"
)

func Test(t *testing.T) {
	s := serverutil.NewGinArticleServer()
	go func() {
		if err := s.Run(":3000"); err != nil {
			panic(err)
		}
	}()

	client := NewClient()

	// get articles
	fmt.Println("------------------------------------------------------------------------------------")
	getArticles(client)
	fmt.Println("------------------------------------------------------------------------------------")
	getArticlesWithQuery(client)
	fmt.Println("------------------------------------------------------------------------------------")
	getArticle(client)
	fmt.Println("------------------------------------------------------------------------------------")
	getArticleNotFound(client)
}

func getArticles(client *Client) {
	uri := NewUriBuilder(scheme, host, "articles").String()
	fmt.Println("Get articles:", uri)
	status, body, err := client.FastGet(uri)
	if err != nil {
		panic(err)
	}
	if status%200 >= 100 {
		fmt.Println("get articles. error >> status:", status, ", body:", string(body))
	} else {
		fmt.Println("get articles >> ", string(body))
	}
}

func getArticlesWithQuery(client *Client) {
	uri := NewUriBuilder(scheme, host, "articles").WithQuery("limit", "5").WithQuery("offset", "3").String()
	fmt.Println("Get articles with query:", uri)
	status, body, err := client.FastGet(uri)
	if err != nil {
		panic(err)
	}
	if status%200 >= 100 {
		fmt.Println("get articles. error >> status:", status, ", body:", string(body))
	} else {
		fmt.Println("get articles >> ", string(body))
	}
}

func getArticle(client *Client) {
	title := "title1"
	uri := NewUriBuilder(scheme, host, "articles", title).String()
	fmt.Println("Get article:", uri)

	status, body, err := client.FastGet(uri)
	if err != nil {
		panic(err)
	}
	if status%200 >= 100 {
		fmt.Println("get article. error >> status:", status, ", body:", string(body))
	} else {
		fmt.Println("get article >> ", string(body))
	}
}

func getArticleNotFound(client *Client) {
	title := "title0"
	uri := NewUriBuilder(scheme, host, "articles", title).String()
	fmt.Println("Get article for notfound:", uri)

	status, body, err := client.FastGet(uri)
	if err != nil {
		panic(err)
	}
	if status%200 >= 100 {
		fmt.Println("get article. error >> status:", status, ", body:", string(body))
	} else {
		fmt.Println("get article >> ", string(body))
	}
}
