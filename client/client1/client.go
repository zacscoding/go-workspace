package client1

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

// TODO : opts
func NewClient() *Client {
	c := &Client{
		cli: &fasthttp.Client{
			NoDefaultUserAgentHeader:      true,
			MaxConnsPerHost:               10000,
			ReadBufferSize:                4096,
			WriteBufferSize:               4096,
			ReadTimeout:                   time.Second,
			WriteTimeout:                  time.Second,
			MaxIdleConnDuration:           time.Minute,
			DisableHeaderNamesNormalizing: true,
		},
	}
	return c
}

func (c *Client) FastGet(uri string) (StatusCode, []byte, error) {
	return c.do("GET", uri, []byte{})
}

func (c *Client) FastPost(uri string, body []byte) (StatusCode, []byte, error) {
	return c.do("POST", uri, body)
}

func (c *Client) FastPut(uri string, body []byte) (StatusCode, []byte, error) {
	return c.do("PUT", uri, body)
}

func (c *Client) FastDelete(uri string, body []byte) (StatusCode, []byte, error) {
	return c.do("DELETE", uri, body)
}

func (c *Client) do(method, uri string, body []byte) (StatusCode, []byte, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(uri)
	req.Header.SetMethod(method)
	req.Header.Add("Accept", AcceptJson)
	if len(body) != 0 {
		req.SetBody(body)
	}

	if c.BeforeExecute != nil {
		c.BeforeExecute(req)
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
