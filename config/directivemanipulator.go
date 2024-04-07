package config

import (
	"github.com/r2dtools/gonginx/internal/rawparser"
	"golang.org/x/exp/slices"
)

type entryContainer interface {
	GetEntries() []*rawparser.Entry
	SetEntries(entries []*rawparser.Entry)
}

func deleteDirective(c entryContainer, callback func(directive *rawparser.Directive) bool) {
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
