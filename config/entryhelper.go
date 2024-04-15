package config

import "github.com/r2dtools/gonginx/internal/rawparser"

type entryContainer interface {
	GetEntries() []*rawparser.Entry
	SetEntries(entries []*rawparser.Entry)
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

func findEntryCommentIndexesToDelete(entries []*rawparser.Entry, index int) []int {
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
