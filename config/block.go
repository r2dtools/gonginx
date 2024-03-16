package config

type Block struct {
	Name       string
	Parameters []string
	Directives []Directive
	Blocks     []Block
	Comments   []Comment
}
