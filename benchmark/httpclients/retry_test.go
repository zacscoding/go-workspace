package httpclients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gojektech/heimdall/v6"
	"github.com/gojektech/heimdall/v6/httpclient"
	"github.com/sethvargo/go-retry"
	"github.com/valyala/fasthttp"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

// go test -bench=. -benchtime 500x

var ErrInternalServer = errors.New("internal server error")

const (
	backoffInterval = 500 * time.Millisecond
	jitter          = 10 * time.Millisecond
	maxRetries      = 5
)

// go test -bench=. -benchtime 200x -benchmem
// Benchmark/fasthttp-12                200         168256867 ns/op            3784 B/op         41 allocs/op
// Benchmark/heimdall-12                200         170782323 ns/op           23746 B/op        174 allocs/op
func Benchmark(b *testing.B) {
	b.Run("fasthttp", func(b *testing.B) {
		s := NewTestServer()
		client := &fasthttp.Client{
			NoDefaultUserAgentHeader:      true,
			MaxConnsPerHost:               10000,
			ReadBufferSize:                4096,
			WriteBufferSize:               4096,
			ReadTimeout:                   time.Second,
			WriteTimeout:                  time.Second,
			MaxIdleConnDuration:           time.Minute,
			DisableHeaderNamesNormalizing: true,
		}
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			backoff, _ := retry.NewConstant(backoffInterval)
			backoff = retry.WithMaxRetries(maxRetries, backoff)
			backoff = retry.WithJitter(jitter, backoff)

			req := fasthttp.AcquireRequest()
			resp := fasthttp.AcquireResponse()
			url := fmt.Sprintf("%s?count=%d", s.URL, i)
			req.SetRequestURI(url)
			req.Header.SetMethod("GET")
			retryCnt := -1
			err := retry.Do(ctx, backoff, func(ctx context.Context) error {
				retryCnt++
				resp.Reset()
				err := client.Do(req, resp)
				if err != nil {
					return retry.RetryableError(err)
				}
				if resp.StatusCode() != http.StatusOK {
					return retry.RetryableError(ErrInternalServer)
				}
				return nil
			})

			if err != nil {
				fasthttp.ReleaseRequest(req)
				fasthttp.ReleaseResponse(resp)
				b.Fail()
			} else if resp.StatusCode() != http.StatusOK {
				fasthttp.ReleaseRequest(req)
				fasthttp.ReleaseResponse(resp)
				b.Fail()
			}
			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
		}
	})

	b.Run("heimdall", func(b *testing.B) {
		s := NewTestServer()
		backoff := heimdall.NewConstantBackoff(backoffInterval, jitter)
		retrier := heimdall.NewRetrier(backoff)
		client := httpclient.NewClient(
			httpclient.WithHTTPTimeout(time.Second),
			httpclient.WithRetrier(retrier),
			httpclient.WithRetryCount(maxRetries),
		)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			url := fmt.Sprintf("%s?count=%d", s.URL, i)
			resp, err := client.Get(url, nil)
			if err != nil {
				b.Fail()
			}
			if resp == nil || resp.StatusCode != http.StatusOK {
				b.Fail()
			}
		}
	})
}

func NewTestServer() *httptest.Server {
	callCount := int64(0)
	countMap := make(map[int]int)
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		defer func() {
			atomic.AddInt64(&callCount, 1)
		}()
		count := req.URL.Query().Get("count")
		if count == "" {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		c, err := strconv.Atoi(count)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		called, ok := countMap[c]
		if !ok {
			called = 0
		}
		called++
		countMap[c] = called
		if c%3 == 0 {
			if called == 1 {
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		rw.WriteHeader(http.StatusOK)
		b, _ := json.Marshal(map[string]interface{}{
			"status": "ok",
		})
		rw.Write(b)
	}))
}
