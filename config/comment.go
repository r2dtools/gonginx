package config

type CommentPosition int

const (
	Inline CommentPosition = iota
	Before
	After
)

type Comment struct {
	Content  string
	Position CommentPosition
}
