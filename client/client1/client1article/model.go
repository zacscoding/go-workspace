package client1article

type StatusCode int

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
	Code    int    `json:"code"`
	Message string `json:"message"`
}
