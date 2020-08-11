package fastclient

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewUriBuilderFrom(t *testing.T) {
	tests := []struct {
		RawUrl string
		Paths  []string
		// expected
		EncodedUri string
		ErrMsg     string
	}{
		{
			RawUrl:     "http://localhost:3000",
			Paths:      []string{"path1/path2/path3"},
			EncodedUri: "http://localhost:3000/path1/path2/path3",
		}, {
			RawUrl:     "http://localhost:3000?q1=v1&q2=v2",
			Paths:      []string{"path1", "/path2", "path3//"},
			EncodedUri: "http://localhost:3000/path1/path2/path3?q1=v1&q2=v2",
		},
	}

	for _, test := range tests {
		builder, err := NewUriBuilderFrom(test.RawUrl, test.Paths...)
		if test.ErrMsg != "" {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), test.ErrMsg)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, test.EncodedUri, builder.String())
	}
}

func TestNewUriBuilder(t *testing.T) {
	tests := []struct {
		Scheme string
		Host   string
		Paths  []string
		// expected
		EncodedUri string
	}{
		{
			Scheme:     "https",
			Host:       "th-api.com",
			Paths:      []string{"v2", "contract", "ft"},
			EncodedUri: "https://th-api.com/v2/contract/ft",
		}, {
			Scheme:     "https",
			Host:       "th-api.com",
			Paths:      []string{"v2", "", "ft"},
			EncodedUri: "https://th-api.com/v2//ft",
		}, {
			Scheme:     "https",
			Host:       "th-api.com",
			Paths:      []string{"v2", "/contract", "//ft"},
			EncodedUri: "https://th-api.com/v2/contract/ft",
		}, {
			Scheme:     "https",
			Host:       "th-api.com",
			Paths:      []string{"v2", "contract/", "ft//"},
			EncodedUri: "https://th-api.com/v2/contract/ft",
		},
	}

	for _, test := range tests {
		builder := NewUriBuilder(test.Scheme, test.Host, test.Paths...)
		assert.Equal(t, test.EncodedUri, builder.String())
	}
}

func TestWithQuery(t *testing.T) {
	rawUrl := "https://th-api.com"
	path := "/path1/path2"
	tests := []struct {
		QueryValues map[string]string
		EncodedUri  string
	}{
		{
			QueryValues: map[string]string{
				"q1": "v1",
				"q2": "v2",
			},
			EncodedUri: rawUrl + path + "?q1=v1&q2=v2",
		}, {
			QueryValues: map[string]string{
				"":   "v1",
				"q2": "v2",
			},
			EncodedUri: rawUrl + path + "?=v1&q2=v2",
		},
	}

	for _, test := range tests {
		builder, err := NewUriBuilderFrom(rawUrl, path)
		for k, v := range test.QueryValues {
			builder.WithQuery(k, v)
		}

		assert.NoError(t, err)
		assert.Equal(t, test.EncodedUri, builder.String())
	}
}
