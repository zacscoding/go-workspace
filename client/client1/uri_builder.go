package client1

import (
	"net/url"
	"strings"
)

type UriBuilder struct {
	baseUrl *url.URL
	values  url.Values
}

func NewUriBuilderFrom(rawUrl string, paths ...string) (*UriBuilder, error) {
	baseUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	baseUrl.Path = strings.TrimLeft(baseUrl.Path, "/")
	for i, p := range paths {
		if i != 0 {
			baseUrl.Path += "/"
		}
		baseUrl.Path += strings.Trim(p, "/")
	}
	return &UriBuilder{
		baseUrl: baseUrl,
		values:  baseUrl.Query(),
	}, nil
}

func NewUriBuilder(scheme, host string, paths ...string) *UriBuilder {
	path := ""
	for i, p := range paths {
		if i != 0 {
			path += "/"
		}
		path += strings.Trim(p, "/")
	}

	return &UriBuilder{
		baseUrl: &url.URL{
			Scheme: scheme,
			Host:   host,
			Path:   path,
		},
		values: url.Values{},
	}
}

func (b *UriBuilder) WithQuery(key, value string) *UriBuilder {
	b.values.Add(key, value)
	return b
}

func (b *UriBuilder) String() string {
	b.baseUrl.RawQuery = b.values.Encode()
	return b.baseUrl.String()
}
