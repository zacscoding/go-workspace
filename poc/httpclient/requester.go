package httpclient

import (
	"context"
	"net/http"
)

type RequestContext struct {
	context context.Context
	method  string
	baseURL string
	header  http.Header
}

func (rc *RequestContext) Get(baseURL string) *RequestContext {
	rc.method = http.MethodGet
	rc.baseURL = baseURL
	return rc
}

func (rc *RequestContext) Set(key, value string) *RequestContext {
	rc.header.Set(key, value)
	return rc
}
