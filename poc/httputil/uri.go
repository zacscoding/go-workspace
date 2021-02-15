package httputil

import "net/url"

type UriBuilder struct {
	baseUrl *url.URL
}

func NewUriBuilder(rawURL string) (*UriBuilder, error) {
	baseUrl, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return &UriBuilder{
		baseUrl: baseUrl,
	}, nil
}
