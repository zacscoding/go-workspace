package httpclient

import (
	"context"
	"github.com/valyala/fasthttp"
	"net/http"
	"testing"
)

func TestUsage(t *testing.T) {
	cli, _ := New()
	cli.New(context.Background()).Get("http://localhost:8900/api/v1/articles")
	h1 := make(http.Header)
	h2 := fasthttp.Request{}
	var res http.Response
	fasthttp.AcquireResponse()
}
