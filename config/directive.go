package config

import "github.com/r2dtools/gonginx/internal/rawparser"

type Directive struct {
	rawDirective *rawparser.Directive
	Name         string
	Values       []string
	Comments     []Comment
}

func (d *Directive) GetFirstValue() string {
	if len(d.Values) == 0 {
		return ""
	}

	return d.Values[0]
}

func (d *Directive) AddValue(expression string) {
	expressions := d.rawDirective.GetExpressions()
	expressions = append(expressions, expression)

	d.rawDirective.SetValues(expressions)
}

func (d *Directive) SetValues(expressions []string) {
	d.rawDirective.SetValues(expressions)
}

func (d *Directive) SetValue(expression string) {
	d.SetValues([]string{expression})
}
