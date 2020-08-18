package fasthttpexample

import (
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"testing"
	"time"
)

func TestHttpClient(t *testing.T) {
	e := NewArticleServer()
	go func() {
		if err := e.Run(":3000"); err != nil {
			panic(err)
		}
	}()

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI("http://localhost:3000/articles")
	req.Header.SetMethod("GET")
	req.Header.SetContentType("application/json")

	// https://gist.github.com/erikdubbelboer/fe4095419fca55e2c92b3d0432ccd7fc
	client := &fasthttp.Client{
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		MaxConnsPerHost:               10000,
		ReadBufferSize:                4096, // Make sure to set this big enough that your whole request can be read at once.
		WriteBufferSize:               4096, // Same but for your response.
		ReadTimeout:                   time.Second,
		WriteTimeout:                  time.Second,
		MaxIdleConnDuration:           time.Minute,
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this.
	}
	if err := client.Do(req, resp); err != nil {
		panic(err)
	}
	type Article struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Author  struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	}
	var result []Article
	err := json.Unmarshal(resp.Body(), &result)
	if err != nil {
		panic(err)
	}
	for i, article := range result {
		fmt.Printf("%d >> %v\n", i, article)
	}
}

func TestHttpClientTimeout(t *testing.T) {
	e := NewArticleServer()
	go func() {
		if err := e.Run(":3000"); err != nil {
			panic(err)
		}
	}()

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	// https://gist.github.com/erikdubbelboer/fe4095419fca55e2c92b3d0432ccd7fc
	client := &fasthttp.Client{
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		MaxConnsPerHost:               10000,
		ReadBufferSize:                4096, // Make sure to set this big enough that your whole request can be read at once.
		WriteBufferSize:               4096, // Same but for your response.
		ReadTimeout:                   time.Second,
		WriteTimeout:                  time.Second,
		MaxIdleConnDuration:           time.Minute,
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this.
	}

	req.SetRequestURI("http://localhost:3000/sleep/5")
	req.Header.SetMethod("GET")
	req.Header.SetContentType("application/json")

	if err := client.DoTimeout(req, resp, 3*time.Second); err != nil {
		fmt.Println("Error :", err)
		if err == fasthttp.ErrTimeout {
			fmt.Println("error")
		}
		return
	}
	type Response struct {
		Status string `json:"status"`
	}
	var result Response
	err := json.Unmarshal(resp.Body(), &result)
	if err != nil {
		panic(err)
	}
	fmt.Println("Status ::", result.Status)
}
