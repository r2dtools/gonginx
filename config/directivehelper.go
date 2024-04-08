package config

import (
	"github.com/r2dtools/gonginx/internal/rawparser"
	"golang.org/x/exp/slices"
)

type entryContainer interface {
	GetEntries() []*rawparser.Entry
	SetEntries(entries []*rawparser.Entry)
}

func deleteDirectiveByName(c entryContainer, directiveName string) {
	deleteDirectiveInEntityContainer(c, func(rawDirective *rawparser.Directive) bool {
		return rawDirective.Identifier == directiveName
	})
}

func deleteDirective(c entryContainer, directive Directive) {
	deleteDirectiveInEntityContainer(c, func(rawDirective *rawparser.Directive) bool {
		return rawDirective.Identifier == directive.Name && slices.Equal(rawDirective.GetExpressions(), directive.Values)
	})
}

func deleteDirectiveInEntityContainer(c entryContainer, callback func(directive *rawparser.Directive) bool) {
	entries := c.GetEntries()
	dEntries := []*rawparser.Entry{}
	indexesToDelete := []int{}

	for index, entry := range entries {
		if entry.Directive == nil {
			continue
		}

		if callback(entry.Directive) {
			indexesToDelete = append(indexesToDelete, index)
			indexesToDelete = append(indexesToDelete, findDirectiveCommentIndexesToDelete(entries, index)...)
		}
	}

	for index, entry := range entries {
		if !slices.Contains(indexesToDelete, index) {
			dEntries = append(dEntries, entry)
		}
	}

	setEntries(c, dEntries)
}

func addDirective(c entryContainer, name string, values []string, begining bool) {
	entries := c.GetEntries()
	directiveValues := []*rawparser.Value{}

	for _, value := range values {
		directiveValues = append(directiveValues, &rawparser.Value{Expression: value})
	}

	directive := &rawparser.Directive{
		Identifier: name,
		Values:     directiveValues,
	}
	entry := &rawparser.Entry{
		Directive:   directive,
		EndNewLines: []string{"\n"},
	}

	var prevEntry *rawparser.Entry

	if len(entries) > 0 {
		prevEntry = entries[len(entries)-1]
	}

	if prevEntry == nil || len(prevEntry.EndNewLines) == 0 {
		entry.StartNewLines = []string{"\n"}
	}

	if begining {
		entries = append([]*rawparser.Entry{entry}, entries...)
	} else {
		entries = append(entries, entry)
	}

	c.SetEntries(entries)
}

func findDirectiveCommentIndexesToDelete(entries []*rawparser.Entry, index int) []int {
	indexesToDelete := []int{}

	for i := index - 1; i >= 0; i-- {
		if entries[i].Comment == nil {
			break
		}

		indexesToDelete = append(indexesToDelete, i)
	}

	if index >= len(entries) {
		return indexesToDelete
	}

	inlineCommentEntry := entries[index+1]

	if inlineCommentEntry.Comment == nil {
		return indexesToDelete
	}

	if len(inlineCommentEntry.StartNewLines) == 0 && len(entries[index].EndNewLines) == 0 {
		indexesToDelete = append(indexesToDelete, index+1)
	}

	return indexesToDelete
}

func setEntries(c entryContainer, entries []*rawparser.Entry) {
	entriesCount := len(entries)

	if entriesCount > 0 {
		if len(entries[0].StartNewLines) == 0 {
			entries[0].StartNewLines = []string{"\n"}
		}

		if len(entries[entriesCount-1].EndNewLines) == 0 {
			entries[entriesCount-1].EndNewLines = []string{"\n"}
		}
	}

	c.SetEntries(entries)
}
