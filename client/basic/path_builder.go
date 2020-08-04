package basic

import (
	"net/url"
)

// What to do
// [x] query builder
// [ ] uri builder
// e.g) builder := NewUriBuilder(http://localhost:8080/:address).WithUri("address", "123").WithQuery("q1","v1")
//     => builder.String() == http://localhost:8080/123?q1=v1

type UriBuilder struct {
	baseUrl *url.URL
	values  url.Values
}

func NewUriBuilderFrom(rawUrl string) (*UriBuilder, error) {
	baseUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	return &UriBuilder{
		baseUrl: baseUrl,
		values:  baseUrl.Query(),
	}, nil
}

func NewUriBuilder(scheme, host, path string) *UriBuilder {
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
