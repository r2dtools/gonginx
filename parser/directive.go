package parser

type Directive struct {
	Name          string
	Values        []string
	NewLineBefore bool
	NewLineAfter  bool
}

func (d *Directive) AddValues(values ...string) {
	d.Values = append(d.Values, values...)
}
