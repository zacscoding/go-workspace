package bindings

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test3(t *testing.T) {
	h := Handler{}

	e := echo.New()
	e.Validator = &Validator{validator.New()}
	e.HTTPErrorHandler = httpErrorHandler
	e.POST("/articles", h.HandlePostArticle)
	if err := e.Start(":8900"); err != nil {
		e.Logger.Fatal(err)
	}
}

func TestTemp(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name: "valid request",
			body: `
			{
			  "article": {
				"title": "my title",
				"description": "my description",
				"body": "contents",
				"tags": [
				  "tag1",
				  "tag2"
				]
			  }
			}`,
			expected: `{"body":"my description","description":"my description","tags":"my description","title":"my title"}`,
		},
		{
			name: "empty title",
			body: `
			{
			  "article": {
				"title": "",
				"description": "my description",
				"body": "contents",
				"tags": []
			  }
			}`,
			expected: `
			{
			  "errors": [
				{
				  "message": "invalid Title field. reason: required"
				}
			  ]
			}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			defer e.Close()
			e.Validator = &Validator{validator.New()}
			e.HTTPErrorHandler = httpErrorHandler

			req := httptest.NewRequest(http.MethodPost, "/article", strings.NewReader(tc.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			h := Handler{}
			h.Route(e)

			e.ServeHTTP(rec, req)

			fmt.Println("Resp > ", rec.Body.String())
			assert.JSONEq(t, tc.expected, rec.Body.String())
		})
	}
}
