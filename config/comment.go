package config

import "github.com/r2dtools/gonginx/internal/rawparser"

type CommentPosition int

const (
	Inline CommentPosition = iota
	Before
)

type Comment struct {
	rawCommet *rawparser.Comment
	Content   string
	Position  CommentPosition
}
