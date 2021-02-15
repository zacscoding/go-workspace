package httputil

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"net/http"
)

// StatusError
type StatusError struct {
	Method       string
	URL          string
	StatusCode   int
	Header       http.Header
	ResponseDump string
}

func (s *StatusError) Error() string {
	return fmt.Sprintf("StatusError{Method:%s, URL:%s, StatusCode:%d}", s.Method, s.URL, s.StatusCode)
}

func NewStatusError(method, url string, resp *fasthttp.Response) *StatusError {
	statusErr := StatusError{
		Method:       method,
		URL:          url,
		StatusCode:   resp.StatusCode(),
		Header:       make(http.Header),
		ResponseDump: string(resp.Body()),
	}
	resp.Header.VisitAll(func(key, value []byte) {
		statusErr.Header.Add(string(key), string(value))
	})
	return &statusErr
}
