package httputil

import (
	"github.com/google/go-querystring/query"
	"net/url"
)

type UriBuilder struct {
	rawURL       string
	parsed       bool
	parsedUrl    string
	queryStructs []interface{}
	queryValues  url.Values
}

func NewUriBuilder(rawURL string) *UriBuilder {
	return &UriBuilder{
		rawURL:      rawURL,
		queryValues: make(url.Values),
	}
}

func (b *UriBuilder) Path(path string) *UriBuilder {
	parsedRawUrl, rawUrlErr := url.Parse(b.rawURL)
	pathUrl, pathUrlErr := url.Parse(path)
	if rawUrlErr == nil && pathUrlErr == nil {
		b.rawURL = parsedRawUrl.ResolveReference(pathUrl).String()
		b.parsed = false
	}
	return b
}

func (b *UriBuilder) QueryStruct(s interface{}) *UriBuilder {
	if s != nil {
		b.queryStructs = append(b.queryStructs, s)
		b.parsed = false
	}
	return b
}

func (b *UriBuilder) Query(key, value string, ommitEmpty bool) *UriBuilder {
	if value != "" || (value == "" && !ommitEmpty) {
		b.queryValues.Add(key, value)
	}
	return b
}

func (b *UriBuilder) ToString() (string, error) {
	if b.parsed {
		return b.parsedUrl, nil
	}
	// add query values
	urlValues := make(url.Values)
	for key, values := range b.queryValues {
		for _, v := range values {
			urlValues.Add(key, v)
		}
	}
	for _, queryStruct := range b.queryStructs {
		queryValues, err := query.Values(queryStruct)
		if err != nil {
			return "", err
		}
		for key, values := range queryValues {
			for _, v := range values {
				urlValues.Add(key, v)
			}
		}
	}
	rawUrl, err := url.Parse(b.rawURL)
	if err != nil {
		return "", err
	}
	rawUrl.RawQuery = urlValues.Encode()
	b.parsed = true
	return rawUrl.String(), nil
}
