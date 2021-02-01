package basic

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/ratelimit"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type RateLimitHttpClient struct {
	cli         *http.Client
	rateLimiter ratelimit.Limiter
}

func (c *RateLimitHttpClient) Do(req *http.Request) (*http.Response, error) {
	c.rateLimiter.Take()
	return c.cli.Do(req)
}

func TestRateLimit(t *testing.T) {
	cli := RateLimitHttpClient{
		cli:         http.DefaultClient,
		rateLimiter: ratelimit.New(10),
	}
	calls := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&calls, 1)
		rw.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	running := int32(1)
	timer := time.NewTimer(time.Second * 10)
	go func() {
		url := server.URL
		for {
			if atomic.LoadInt32(&running) != 1 {
				return
			}
			req, _ := http.NewRequest("GET", url, nil)
			cli.Do(req)
		}
	}()
	select {
	case <-timer.C:
		atomic.CompareAndSwapInt32(&running, 1, 0)
	}
	assert.Greater(t, int32(105), calls)
	assert.Less(t, int32(95), calls)
}
