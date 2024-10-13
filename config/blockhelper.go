package config

import (
	"github.com/r2dtools/gonginx/internal/rawdumper"
	"github.com/r2dtools/gonginx/internal/rawparser"
	"golang.org/x/exp/slices"
)

type blockFinder interface {
	FindBlocks(blockName string) []Block
}

type serverBlockFinder interface {
	FindServerBlocks() []ServerBlock
}

type upstreamBlockFinder interface {
	FindUpstreamBlocks() []UpstreamBlock
}

func findServerBlocks(finder blockFinder) []ServerBlock {
	var serverBlocks []ServerBlock

	for _, block := range finder.FindBlocks(serverBlockName) {
		serverBlocks = append(serverBlocks, ServerBlock{
			Block: block,
		})
	}

	return serverBlocks
}

func findLocationBlocks(finder blockFinder) []LocationBlock {
	var locationBlocks []LocationBlock

	for _, block := range finder.FindBlocks(locationBlockName) {
		locationBlocks = append(locationBlocks, LocationBlock{
			Block: block,
		})
	}

	return locationBlocks
}

func findHttpBlocks(finder blockFinder) []HttpBlock {
	var httpBlocks []HttpBlock

	for _, block := range finder.FindBlocks(httpBlockName) {
		httpBlocks = append(httpBlocks, HttpBlock{
			Block: block,
		})
	}

	return httpBlocks
}

func findUpstreamBlocks(finder blockFinder) []UpstreamBlock {
	var upstreamBlocks []UpstreamBlock

	for _, block := range finder.FindBlocks(upstreamBlockName) {
		upstreamBlocks = append(upstreamBlocks, UpstreamBlock{
			Block: block,
		})
	}

	return upstreamBlocks
}

func findServerBlocksByServerName(finder serverBlockFinder, serverName string) []ServerBlock {
	var serverBlocks []ServerBlock

	for _, serverBlock := range finder.FindServerBlocks() {
		serverNames := serverBlock.GetServerNames()

		if slices.Contains(serverNames, serverName) {
			serverBlocks = append(serverBlocks, serverBlock)
		}
	}

	return serverBlocks
}

func findUpstreamBlocksByName(finder upstreamBlockFinder, upstreamName string) []UpstreamBlock {
	var upstreamBlocks []UpstreamBlock

	for _, upstreamBlock := range finder.FindUpstreamBlocks() {
		if upstreamBlock.GetUpstreamName() == upstreamName {
			upstreamBlocks = append(upstreamBlocks, upstreamBlock)
		}
	}

	return upstreamBlocks
}

func newBlock(container entryContainer, config *Config, name string, parameters []string, begining bool) Block {
	rawBlock := &rawparser.BlockDirective{
		Identifier: name,
		Content:    &rawparser.BlockContent{},
	}
	rawBlock.SetParameters(parameters)

	block := Block{
		config:    config,
		container: container,
		rawBlock:  rawBlock,
		rawDumper: &rawdumper.RawDumper{},
	}

	entries := container.GetEntries()

	indexToInsert := -1
	similarBlocksIndexes := []int{}

	for index, entry := range entries {
		if entry.BlockDirective != nil && entry.BlockDirective.Identifier == name {
			similarBlocksIndexes = append(similarBlocksIndexes, index)
		}
	}

	if len(similarBlocksIndexes) != 0 {
		if begining {
			indexToInsert = similarBlocksIndexes[0]

			// skip block comments befor insert
			for i := similarBlocksIndexes[0] - 1; i >= 0; i-- {
				if entries[i].Comment == nil {
					break
				}

				indexToInsert = i
			}
		} else {
			indexToInsert = similarBlocksIndexes[len(similarBlocksIndexes)-1]

			if indexToInsert == len(entries)-1 {
				indexToInsert = -1
			} else {
				indexToInsert += 1
			}
		}
	}

	entry := &rawparser.Entry{
		BlockDirective: rawBlock,
		EndNewLines:    []string{"\n\n"},
	}

	if indexToInsert == -1 {
		entries = append(entries, entry)
	} else {
		if indexToInsert == 0 {
			entry.StartNewLines = []string{"\n"}
		}
		entries = slices.Insert(entries, indexToInsert, entry)
	}

	setEntries(container, entries)

	return block
}

func deleteBlockByName(c entryContainer, name string) {
	deleteBlockEntityContainer(c, func(block *rawparser.BlockDirective) bool {
		return block.Identifier == name
	})
}

func deleteBlock(c entryContainer, block Block) {
	deleteBlockEntityContainer(c, func(rawBlock *rawparser.BlockDirective) bool {
		return block.rawBlock == rawBlock
	})
}

func deleteBlockEntityContainer(c entryContainer, callback func(block *rawparser.BlockDirective) bool) {
	entries := c.GetEntries()
	dEntries := []*rawparser.Entry{}
	indexesToDelete := []int{}

	for index, entry := range entries {
		if entry.BlockDirective == nil {
			continue
		}

		if callback(entry.BlockDirective) {
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

func addLocationBlock(b *Block, modifier, match string, begining bool) LocationBlock {
	parameters := []string{}

	if modifier != "" {
		parameters = append(parameters, modifier)
	}

	if match != "" {
		parameters = append(parameters, match)
	}

	block := b.addBlock(locationBlockName, parameters, begining)

	return LocationBlock{
		Block: block,
	}
}
