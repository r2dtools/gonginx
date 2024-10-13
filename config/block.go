package config

import (
	"strings"

	"github.com/r2dtools/gonginx/internal/rawdumper"
	"github.com/r2dtools/gonginx/internal/rawparser"
	"golang.org/x/exp/slices"
)

const (
	upstreamBlockName = "upstream"
	locationBlockName = "location"
	httpBlockName     = "http"
	serverBlockName   = "server"
)

type Block struct {
	FilePath  string
	config    *Config
	container entryContainer
	rawBlock  *rawparser.BlockDirective
	rawDumper *rawdumper.RawDumper
}

func (b *Block) GetName() string {
	return b.rawBlock.Identifier
}

func (b *Block) GetParameters() []string {
	return b.rawBlock.GetParametersExpressions()
}

func (b *Block) SetParameters(parameters []string) {
	b.rawBlock.SetParameters(parameters)
}

func (b *Block) FindDirectives(directiveName string) []Directive {
	var directives []Directive

	for _, entry := range b.rawBlock.GetEntries() {
		directives = append(directives, b.config.findDirectivesRecursively(directiveName, b.rawBlock, entry, true)...)
	}

	return directives
}

func (b *Block) FindBlocks(blockName string) []Block {
	var blocks []Block

	for _, entry := range b.rawBlock.GetEntries() {
		blocks = append(blocks, b.config.findBlocksRecursively(blockName, b.FilePath, b.rawBlock, entry, true)...)
	}

	return blocks
}

func (b *Block) AddDirective(directive Directive, begining bool, endWithNewLine bool) {
	addDirective(b.rawBlock, directive, begining, endWithNewLine)
}

func (b *Block) DeleteDirective(directive Directive) {
	deleteDirective(b.rawBlock, directive)
}

func (b *Block) DeleteDirectiveByName(directiveName string) {
	deleteDirectiveByName(b.rawBlock, directiveName)
}

func (b *Block) FindComments() []Comment {
	entries := b.container.GetEntries()
	comments := []Comment{}

	var (
		index         int
		bEntry, entry *rawparser.Entry
	)

	for index, entry = range entries {
		if entry.BlockDirective == nil {
			continue
		}

		if entry.BlockDirective == b.rawBlock {
			bEntry = entry
			break
		}
	}

	if bEntry == nil {
		return comments
	}

	comment := b.findInlineComment(bEntry.BlockDirective)

	if comment != nil {
		comments = append(comments, *comment)
	}

	for prevIndex := index - 1; prevIndex >= 0; prevIndex-- {
		entry := entries[prevIndex]

		if entry.Comment == nil {
			break
		}

		comment := b.createComment(entry.Comment, CommentPosition(Before))
		comments = append([]Comment{comment}, comments...)

	}

	return comments
}

func (b *Block) SetComments(comments []string) {
	var (
		index         int
		bEntry, entry *rawparser.Entry
	)
	entries := b.container.GetEntries()

	for index, entry = range entries {
		if entry.BlockDirective == nil {
			continue
		}

		if entry.BlockDirective == b.rawBlock {
			bEntry = entry
			break
		}
	}

	if bEntry == nil {
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

	setEntries(b.container, entries)
}

func (b *Block) Dump() string {
	entry := rawparser.Entry{
		BlockDirective: b.rawBlock,
	}

	return b.rawDumper.DumpEntry(&entry)
}

func (b *Block) findInlineComment(blockDirective *rawparser.BlockDirective) *Comment {
	if blockDirective == nil {
		return nil
	}

	content := blockDirective.Content

	if content == nil || len(content.Entries) == 0 {
		return nil
	}

	firstEntry := content.Entries[0]

	if firstEntry.Comment != nil {
		comment := b.createComment(firstEntry.Comment, CommentPosition(Inline))

		return &comment
	}

	return nil
}

func (b *Block) createComment(rawComment *rawparser.Comment, position CommentPosition) Comment {
	return Comment{
		rawComment: rawComment,
		Content:    strings.Trim(rawComment.Value, "\n# "),
		Position:   position,
	}
}

func (b *Block) addBlock(name string, parameters []string, begining bool) Block {
	return newBlock(b.rawBlock, b.config, name, parameters, begining)
}

func (b *Block) setContainer(container entryContainer) {
	b.container = container
}
