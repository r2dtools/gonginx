package config

import (
	"github.com/r2dtools/gonginx/internal/rawparser"
	"golang.org/x/exp/slices"
)

func deleteDirectiveByName(c entryContainer, directiveName string) {
	deleteDirectiveInEntityContainer(c, func(rawDirective *rawparser.Directive) bool {
		return rawDirective.Identifier == directiveName
	})
}

func deleteDirective(c entryContainer, directive Directive) {
	deleteDirectiveInEntityContainer(c, func(rawDirective *rawparser.Directive) bool {
		return directive.rawDirective == rawDirective
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
			indexesToDelete = append(indexesToDelete, findEntryCommentIndexesToDelete(entries, index)...)
		}
	}

	for index, entry := range entries {
		if !slices.Contains(indexesToDelete, index) {
			dEntries = append(dEntries, entry)
		}
	}

	setEntries(c, dEntries)
}

func addDirective(c entryContainer, directive Directive, begining bool) {
	entries := c.GetEntries()
	entry := &rawparser.Entry{
		Directive:   directive.rawDirective,
		EndNewLines: []string{"\n"},
	}

	var prevEntry *rawparser.Entry

	if len(entries) > 0 && !begining {
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

	setEntries(c, entries)
}
