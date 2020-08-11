package fastclient

import (
	"github.com/valyala/fasthttp"
	"time"
)

type StatusCode int

const (
	AcceptJson = "application/json"
)

type Client struct {
	cli           *fasthttp.Client
	BeforeExecute func(req *fasthttp.Request)
}

type ClientConfig struct {
	ReadBufferSize     int
	WriteBufferSize    int
	ReadTimeout        int
	WriteTimeout       int
	MaxIdleConnSeconds int
}

func NewClient() *Client {
	return NewClientWith(ClientConfig{
		ReadBufferSize:     4096,
		WriteBufferSize:    4096,
		ReadTimeout:        1,
		WriteTimeout:       1,
		MaxIdleConnSeconds: 60,
	})
}

func NewClientWith(config ClientConfig) *Client {
	c := &Client{
		cli: &fasthttp.Client{
			NoDefaultUserAgentHeader:      true,
			ReadBufferSize:                config.ReadBufferSize,
			WriteBufferSize:               config.WriteBufferSize,
			ReadTimeout:                   time.Duration(config.ReadTimeout) * time.Second,
			WriteTimeout:                  time.Duration(config.WriteTimeout) * time.Second,
			MaxIdleConnDuration:           time.Duration(config.MaxIdleConnSeconds) * time.Second,
			DisableHeaderNamesNormalizing: true,
		},
	}
	return c
}

func (c *Client) FastGet(uri string, headers map[string]string) (StatusCode, []byte, error) {
	return c.do("GET", uri, []byte{}, headers)
}

func (c *Client) FastPost(uri string, body []byte, headers map[string]string) (StatusCode, []byte, error) {
	return c.do("POST", uri, body, headers)
}

func (c *Client) FastPut(uri string, body []byte, headers map[string]string) (StatusCode, []byte, error) {
	return c.do("PUT", uri, body, headers)
}

func (c *Client) FastDelete(uri string, body []byte, headers map[string]string) (StatusCode, []byte, error) {
	return c.do("DELETE", uri, body, headers)
}

func (c *Client) do(method, uri string, body []byte, headers map[string]string) (StatusCode, []byte, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(uri)

	req.Header.SetMethod(method)
	req.Header.Add("Accept", AcceptJson)
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	if len(body) != 0 {
		req.SetBody(body)
	}

	err := c.cli.Do(req, resp)
	if err != nil {
		return 0, []byte{}, err
	}
	return StatusCode(resp.StatusCode()), resp.Body(), nil
}

func (c StatusCode) Is2xxSuccessful() bool {
	return c >= 200 && c < 300
}
