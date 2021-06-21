package bindings

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Validator struct {
	delegate *validator.Validate
}

func (v *Validator) Validate(i interface{}) error {
	return v.delegate.Struct(i)
}

type Handler struct {
}

func (h *Handler) Route(e *echo.Echo) {
	e.POST("/article", h.HandlePostArticle)
}

func (h *Handler) HandlePostArticle(ctx echo.Context) error {
	req := &ArticleCreateRequest{}
	if err := req.bind(ctx); err != nil {
		return wrapBindError(err)
	}
	return ctx.JSON(http.StatusOK, echo.Map{
		"title":       req.Article.Title,
		"description": req.Article.Description,
		"body":        req.Article.Description,
		"tags":        req.Article.Description,
	})
}

func httpErrorHandler(err error, ctx echo.Context) {
	code := http.StatusInternalServerError
	switch v := err.(type) {
	case *echo.HTTPError:
		code = v.Code
	case *HttpError:
		code = v.Code
	}
	if !ctx.Response().Committed {
		if ctx.Request().Method == http.MethodHead {
			ctx.NoContent(code)
		} else {
			ctx.JSON(code, err)
		}
	}
}
