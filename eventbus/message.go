package eventbus

type Type string

const (
	TypeUnknown = Type("unknown")
	TypeArticle = Type("article.v1")
	TypeComment = Type("comment.v1")
)

func (t Type) GetTopic() string {
	return string(t)
}

type ArticleMessage struct {
	ID      int64
	Title   string
	Content string
}

type CommentMessage struct {
	ID      int64
	Comment string
}
