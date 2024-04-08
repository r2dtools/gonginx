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
	addDirective(b.rawBlock, name, values)
}

func (b *Block) DeleteDirective(directive Directive) {
	deleteDirective(b.rawBlock, directive)
}

func (b *Block) DeleteDirectiveByName(directiveName string) {
	deleteDirectiveByName(b.rawBlock, directiveName)
}
