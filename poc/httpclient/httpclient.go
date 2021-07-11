package httpclient

import (
	"context"
	"net/http"
)

type Option func(c *httpclient)

type HttpClient interface {
	New(ctx context.Context) *RequestContext
}

func New() (HttpClient, error) {
	return &httpclient{}, nil
}

func WithRatelimit(rps int) Option {
	return func(c *httpclient) {
	}
}

type httpclient struct {
}

func (h *httpclient) New(ctx context.Context) *RequestContext {
	return &RequestContext{
		context: ctx,
		method:  "",
		baseURL: "",
		header:  make(http.Header),
	}
}
