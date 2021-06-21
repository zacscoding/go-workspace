package bindings

import "github.com/labstack/echo/v4"

type ArticleCreateRequest struct {
	Article struct {
		Title       string   `json:"title" validate:"required,lte=20"`
		Description string   `json:"description" validate:"required"`
		Body        string   `json:"body" validate:"required"`
		Tags        []string `json:"tags" validate:"omitempty"`
	} `json:"article"`
}

func (r *ArticleCreateRequest) bind(ctx echo.Context) error {
	if err := ctx.Bind(r); err != nil {
		return err
	}
	if err := ctx.Validate(r); err != nil {
		return err
	}
	return nil
}