package config

import (
	"strings"

	"github.com/r2dtools/gonginx/internal/rawparser"
	"golang.org/x/exp/slices"
)

type Directive struct {
	rawDirective *rawparser.Directive
	container    entryContainer
}

func (d *Directive) GetName() string {
	return d.rawDirective.Identifier
}

func (d *Directive) GetValues() []string {
	return d.rawDirective.GetExpressions()
}

func (d *Directive) GetFirstValue() string {
	values := d.GetValues()

	if len(values) == 0 {
		return ""
	}

	return values[0]
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

func (d *Directive) FindComments() []Comment {
	entries := d.container.GetEntries()
	comments := []Comment{}

	var (
		index                    int
		dEntry, entry, nextEntry *rawparser.Entry
	)

	for index, entry = range entries {
		if entry.Directive == nil {
			continue
		}

		if entry.Directive == d.rawDirective {
			dEntry = entry
			break
		}
	}

	if dEntry == nil {
		return comments
	}

	if index < len(entries)-1 {
		nextEntry = entries[index+1]
	}

	comment := d.findInlineComment(dEntry, nextEntry)

	if comment != nil {
		comments = append(comments, *comment)
	}

	for prevIndex := index - 1; prevIndex >= 0; prevIndex-- {
		entry := entries[prevIndex]

		if entry.Comment == nil {
			break
		}

		comment := d.createComment(entry.Comment, CommentPosition(Before))
		comments = append([]Comment{comment}, comments...)

	}

	return comments
}

func (d *Directive) SetComments(comments []string) {
	var (
		index         int
		dEntry, entry *rawparser.Entry
	)
	entries := d.container.GetEntries()

	for index, entry = range entries {
		if entry.Directive == nil {
			continue
		}

		if entry.Directive == d.rawDirective {
			dEntry = entry
			break
		}
	}

	if dEntry == nil {
		return
	}

	pEntries := []*rawparser.Entry{}

	for _, content := range comments {
		commentEntry := rawparser.Entry{
			Comment:     newComment(content).rawComment,
			EndNewLines: []string{"\n"},
		}
		pEntries = append(pEntries, &commentEntry)
	}

	nIndex := index

	for rIndex := index - 1; rIndex >= 0; rIndex-- {
		rEntry := entries[rIndex]

		if rEntry.Comment == nil {
			break
		}

		nIndex = rIndex
	}

	if nIndex != index {
		entries = slices.Delete(entries, nIndex, index)
	}

	entries = slices.Insert(entries, nIndex, pEntries...)

	setEntries(d.container, entries)
}

func (d *Directive) findInlineComment(entry, nextEntry *rawparser.Entry) *Comment {
	if nextEntry == nil || nextEntry.Comment == nil {
		return nil
	}

	if len(entry.EndNewLines) != 0 || len(nextEntry.StartNewLines) != 0 {
		return nil
	}

	comment := d.createComment(nextEntry.Comment, CommentPosition(Inline))

	return &comment
}

func (d *Directive) createComment(rawComment *rawparser.Comment, position CommentPosition) Comment {
	return Comment{
		rawComment: rawComment,
		Content:    strings.Trim(rawComment.Value, "\n# "),
		Position:   position,
	}
}

func (d *Directive) setContainer(container entryContainer) {
	d.container = container
}

func NewDirective(name string, values []string) Directive {
	directiveValues := []*rawparser.Value{}

	for _, value := range values {
		directiveValues = append(directiveValues, &rawparser.Value{Expression: value})
	}

	rawDirective := &rawparser.Directive{
		Identifier: name,
		Values:     directiveValues,
	}

	return Directive{
		rawDirective: rawDirective,
	}
}
