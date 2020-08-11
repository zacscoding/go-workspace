package fastarticle

import (
	"fmt"
	"go-workspace/client/fastclient"
)

type ArticleResult struct {
	Title   string       `json:"title"`
	Content string       `json:"content"`
	Author  AuthorResult `json:"author"`
}

type AuthorResult struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type StatusResult struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	StatusCode fastclient.StatusCode `json:"-"`
	Code       int                   `json:"code"`
	Message    string                `json:"message"`
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("statusCode:%d, code:%d, message:%s", e.StatusCode, e.Code, e.Message)
}
