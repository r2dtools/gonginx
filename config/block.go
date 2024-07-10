package config

import (
	"strings"

	"github.com/r2dtools/gonginx/internal/rawparser"
)

type Block struct {
	config    *Config
	container entryContainer
	rawBlock  *rawparser.BlockDirective
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
	return b.config.findDirectivesRecursivelyInLoop(directiveName, b.rawBlock)
}

func (b *Block) FindBlocks(blockName string) []Block {
	var blocks []Block

	for _, entry := range b.rawBlock.GetEntries() {
		blocks = append(blocks, b.config.findBlocksRecursively(blockName, b.rawBlock, entry)...)
	}

	return blocks
}

func (b *Block) AddDirective(directive Directive, begining bool) {
	addDirective(b.rawBlock, directive, begining)
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

func (b *Block) addBlock(name string, parameters []string) Block {
	return newBlock(b.rawBlock, b.config, name, parameters)
}

func (b *Block) setContainer(container entryContainer) {
	b.container = container
}
