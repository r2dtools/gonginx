package config

import "github.com/r2dtools/gonginxconf/internal/rawparser"

type CommentPosition int

const (
	Inline CommentPosition = iota
	Before
)

type Comment struct {
	rawComment *rawparser.Comment
	Content    string
	Position   CommentPosition
}

func newComment(content string) Comment {
	rawComment := &rawparser.Comment{
		Value: "# " + content,
	}

	return Comment{
		rawComment: rawComment,
		Content:    content,
		Position:   CommentPosition(Before),
	}
}
