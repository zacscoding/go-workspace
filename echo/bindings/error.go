package bindings

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"net/http"
)

type HttpError struct {
	Code     int                      `json:"-"`
	Errors   []map[string]interface{} `json:"errors"`
	Internal error                    `json:"-"`
}

func (he *HttpError) Error() string {
	return he.Internal.Error()
}

func wrapBindError(err error) error {
	he := &HttpError{
		Internal: err,
	}

	switch v := err.(type) {
	case *echo.HTTPError:
		he.Code = http.StatusBadRequest
		he.Errors = append(he.Errors, map[string]interface{}{
			"message": v.Message,
		})
	case validator.ValidationErrors:
		he.Code = http.StatusBadRequest
		for _, ve := range v {
			he.Errors = append(he.Errors, map[string]interface{}{
				"message": fmt.Sprintf("invalid %s field. reason: %s", ve.Field(), ve.Tag()),
			})
		}
	default:
		he.Code = http.StatusInternalServerError
		he.Errors = append(he.Errors, map[string]interface{}{
			"message": err.Error(),
		})
	}
	return he
}
