package config

import (
	"container/list"

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
	var directives []Directive
	entryList := list.New()

	for _, entry := range b.rawBlock.GetEntries() {
		directives = append(directives, b.config.findDirectivesRecursively(directiveName, entry, entryList)...)
	}

	return directives
}

func (b *Block) FindBlocks(blockName string) []Block {
	var blocks []Block

	entryList := list.New()

	for _, entry := range b.rawBlock.GetEntries() {
		blocks = append(blocks, b.config.findBlocksRecursively(blockName, entry, entryList)...)
	}

	return blocks
}
