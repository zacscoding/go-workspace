package httputil

import (
	"context"
	"errors"
	"fmt"
	"github.com/sethvargo/go-retry"
	"github.com/valyala/fasthttp"
	"go.uber.org/ratelimit"
	"net/http"
	"time"
)

var (
	ErrNoMethod = errors.New("require method in request")
	ErrNoUrl    = errors.New("require url in request")
)

var defaultRetryableCodes = map[int]struct{}{
	http.StatusTooManyRequests:    {},
	http.StatusBadGateway:         {},
	http.StatusServiceUnavailable: {},
	http.StatusGatewayTimeout:     {},
}

var defaultAcceptableCodes = map[int]struct{}{
	http.StatusOK:                   {},
	http.StatusCreated:              {},
	http.StatusAccepted:             {},
	http.StatusNonAuthoritativeInfo: {},
	http.StatusNoContent:            {},
	http.StatusResetContent:         {},
	http.StatusPartialContent:       {},
	http.StatusMultiStatus:          {},
	http.StatusAlreadyReported:      {},
	http.StatusIMUsed:               {},
}

type Hook interface {
	BeforeRequest(req *fasthttp.Request, attempts int)
}

type RequestOpt struct {
	Method          string
	Url             string
	Body            []byte
	Header          http.Header
	Hooks           []Hook
	AcceptableCodes map[int]struct{}
	ExtraCodes      map[int]struct{}
}

type HttpClient interface {
	Get(ctx context.Context, url string, header http.Header) (int, []byte, error)
	GetWithOpt(ctx context.Context, opt *RequestOpt) (int, []byte, error)
	Post(ctx context.Context, url string, body []byte, header http.Header) (int, []byte, error)
	PostWithOpt(ctx context.Context, opt *RequestOpt) (int, []byte, error)
	Put(ctx context.Context, url string, body []byte, header http.Header) (int, []byte, error)
	PutWithOpt(ctx context.Context, opt *RequestOpt) (int, []byte, error)
	Delete(ctx context.Context, url string, body []byte, header http.Header) (int, []byte, error)
	DeleteWithOpt(ctx context.Context, opt *RequestOpt) (int, []byte, error)
	Do(ctx context.Context, opt *RequestOpt) (int, []byte, error)
}

type httpClient struct {
	cli     *fasthttp.Client
	hooks   []Hook
	limiter ratelimit.Limiter
	backoff func() retry.Backoff
}

func (c *httpClient) Get(ctx context.Context, url string, header http.Header) (int, []byte, error) {
	return c.do(ctx, c.newDefaultRequestOpt("GET", url, nil, header))
}

func (c *httpClient) GetWithOpt(ctx context.Context, opt *RequestOpt) (int, []byte, error) {
	opt.Method = "GET"
	return c.do(ctx, opt)
}

func (c *httpClient) Post(ctx context.Context, url string, body []byte, header http.Header) (int, []byte, error) {
	return c.do(ctx, c.newDefaultRequestOpt("POST", url, body, header))
}

func (c *httpClient) PostWithOpt(ctx context.Context, opt *RequestOpt) (int, []byte, error) {
	opt.Method = "POST"
	return c.do(ctx, opt)
}

func (c *httpClient) Put(ctx context.Context, url string, body []byte, header http.Header) (int, []byte, error) {
	return c.do(ctx, c.newDefaultRequestOpt("PUT", url, body, header))
}

func (c *httpClient) PutWithOpt(ctx context.Context, opt *RequestOpt) (int, []byte, error) {
	opt.Method = "PUT"
	return c.do(ctx, opt)
}

func (c *httpClient) Delete(ctx context.Context, url string, body []byte, header http.Header) (int, []byte, error) {
	return c.do(ctx, c.newDefaultRequestOpt("DELETE", url, body, header))
}

func (c *httpClient) DeleteWithOpt(ctx context.Context, opt *RequestOpt) (int, []byte, error) {
	opt.Method = "DELETE"
	return c.do(ctx, opt)
}

func (c *httpClient) Do(ctx context.Context, opt *RequestOpt) (int, []byte, error) {
	return c.do(ctx, opt)
}

func (c *httpClient) newDefaultRequestOpt(method, url string, body []byte, header http.Header) *RequestOpt {
	return &RequestOpt{
		Method:          method,
		Url:             url,
		Body:            body,
		Header:          header,
		Hooks:           c.hooks,
		AcceptableCodes: defaultAcceptableCodes,
		ExtraCodes:      defaultRetryableCodes,
	}
}

func (c *httpClient) do(ctx context.Context, opt *RequestOpt) (int, []byte, error) {
	if opt.Method == "" {
		return 0, nil, ErrNoMethod
	}
	if opt.Url == "" {
		return 0, nil, ErrNoUrl
	}
	if opt.Header == nil {
		opt.Header = make(http.Header)
	}
	var (
		statusCode    int
		respBodyBytes []byte
		attempts      = 1
	)
	err := retry.Do(ctx, c.backoff(), func(ctx context.Context) error {
		// setup requests
		var (
			req  = fasthttp.AcquireRequest()
			resp = fasthttp.AcquireResponse()
		)
		defer fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(resp)
		buildRequest(req, opt)
		for _, hook := range opt.Hooks {
			hook.BeforeRequest(req, attempts)
		}
		attempts++
		// do request
		c.limiter.Take()
		var err error
		deadline, ok := ctx.Deadline()
		if ok {
			err = c.cli.DoDeadline(req, resp, deadline)
		} else {
			err = c.cli.Do(req, resp)
		}
		if err != nil {
			return retry.RetryableError(err) // mark as a retryable
		}
		code := resp.StatusCode()
		// success
		if isAcceptable(opt.AcceptableCodes, code) {
			statusCode = code
			respBody := resp.Body()
			copied := make([]byte, len(respBody))
			copy(copied, respBody)
			respBodyBytes = respBody
			return nil
		}
		statusErr := NewStatusError(opt.Method, opt.Url, resp)
		if isRetryable(opt.ExtraCodes, code) {
			return retry.RetryableError(statusErr)
		}
		return statusErr
	})
	if err != nil {
		if unwrap := errors.Unwrap(err); unwrap != nil {
			err = unwrap
		}
		return 0, []byte{}, err
	}
	return statusCode, respBodyBytes, nil
}

func buildRequest(req *fasthttp.Request, opt *RequestOpt) {
	if req == nil {
		return
	}
	if accept := opt.Header.Get("Accept"); accept == "" {
		opt.Header.Set("Accept", "application/json")
	}
	req.SetRequestURI(opt.Url)
	req.Header.SetMethod(opt.Method)
	for key, values := range opt.Header {
		for _, v := range values {
			req.Header.Add(key, v)
		}
	}
	if len(opt.Body) != 0 {
		req.SetBody(opt.Body)
	}
}

func isAcceptable(acceptable map[int]struct{}, statusCode int) bool {
	_, ok := acceptable[statusCode]
	return ok
}

func isRetryable(extraCodes map[int]struct{}, statusCode int) bool {
	// check extraCodes
	if _, ok := extraCodes[statusCode]; ok {
		return true
	}
	// check default retryable codes
	_, ok := defaultRetryableCodes[statusCode]
	return ok
}

// ClientOption sets parameters for the HttpClient
type ClientOption func(cli *httpClient)

// Cli sets HttpClient's http doer i.e fasthttp.Client
func Cli(client *fasthttp.Client) ClientOption {
	return func(cli *httpClient) {
		cli.cli = client
	}
}

func ConstantBackoff(maxRetries uint64, duration time.Duration) ClientOption {
	if duration <= 0 {
		panic(fmt.Errorf("duration must be greater than 0"))
	}
	return func(cli *httpClient) {
		cli.backoff = func() retry.Backoff {
			b, _ := retry.NewConstant(duration)
			return retry.WithMaxRetries(maxRetries, b)
		}
	}
}

func ExponentialBackoff(maxRetries uint64, base time.Duration) ClientOption {
	if base <= 0 {
		panic(fmt.Errorf("base must be greater than 0"))
	}
	return func(cli *httpClient) {
		cli.backoff = func() retry.Backoff {
			b, _ := retry.NewExponential(base)
			return retry.WithMaxRetries(maxRetries, b)
		}
	}
}

func FibonacciBackoff(maxRetries uint64, base time.Duration) ClientOption {
	if base <= 0 {
		panic(fmt.Errorf("base must be greater than 0"))
	}
	return func(cli *httpClient) {
		cli.backoff = func() retry.Backoff {
			b, err := retry.NewFibonacci(base)
			if err != nil {
				panic(err)
			}
			return retry.WithMaxRetries(maxRetries, b)
		}
	}
}

func BackOff(backoff retry.Backoff) ClientOption {
	return func(cli *httpClient) {
		cli.backoff = func() retry.Backoff {
			return backoff
		}
	}
}

func Limiter(rate int) ClientOption {
	return func(cli *httpClient) {
		cli.limiter = ratelimit.New(rate)
	}
}

func NewHttpClient(opts ...ClientOption) (HttpClient, error) {
	cli := httpClient{
		cli:     &fasthttp.Client{},
		limiter: ratelimit.NewUnlimited(),
		backoff: func() retry.Backoff {
			return &noopBackoff{}
		},
	}
	for _, opt := range opts {
		opt(&cli)
	}
	return &cli, nil
}

type noopBackoff struct {
}

func (n *noopBackoff) Next() (next time.Duration, stop bool) {
	return 0, true
}
