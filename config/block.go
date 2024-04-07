package config

import (
	"container/list"

	"golang.org/x/exp/slices"

	"github.com/r2dtools/gonginx/internal/rawparser"
)

type Block struct {
	config     *Config
	rawBlock   *rawparser.BlockDirective
	Name       string
	Parameters []string
	Comments   []Comment
}

func (b *Block) FindDirectives(directiveName string) []Directive {
	prevEntries := list.New()

	return b.config.findDirectivesRecursivelyInLoop(directiveName, b.rawBlock.GetEntries(), prevEntries)
}

func (b *Block) FindBlocks(blockName string) []Block {
	var blocks []Block

	prevEntries := list.New()

	for _, entry := range b.rawBlock.GetEntries() {
		blocks = append(blocks, b.config.findBlocksRecursively(blockName, entry, prevEntries)...)
	}

	return blocks
}

func (b *Block) AddDirective(name string, values []string) {
	entries := b.rawBlock.GetEntries()
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

	entries = append(entries, entry)
	b.rawBlock.SetEntries(entries)
}

func (b *Block) DeleteDirective(directive Directive) {
	deleteDirective(b.rawBlock, func(rawDirective *rawparser.Directive) bool {
		return rawDirective.Identifier == directive.Name && slices.Equal(rawDirective.GetExpressions(), directive.Values)
	})
}

func (b *Block) DeleteDirectiveByName(directiveName string) {
	deleteDirective(b.rawBlock, func(rawDirective *rawparser.Directive) bool {
		return rawDirective.Identifier == directiveName
	})
}

func (b *Block) setEntries(entries []*rawparser.Entry) {
	entriesCount := len(entries)

	if entriesCount > 0 {
		if len(entries[0].StartNewLines) == 0 {
			entries[0].StartNewLines = []string{"\n"}
		}

		if len(entries[entriesCount-1].EndNewLines) == 0 {
			entries[entriesCount-1].EndNewLines = []string{"\n"}
		}
	}

	b.rawBlock.SetEntries(entries)
}
