package config

import (
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

	for _, block := range finder.FindBlocks("server") {
		serverBlocks = append(serverBlocks, ServerBlock{
			Block: block,
		})
	}

	return serverBlocks
}

func findLocationBlocks(finder blockFinder) []LocationBlock {
	var locationBlocks []LocationBlock

	for _, block := range finder.FindBlocks("location") {
		locationBlocks = append(locationBlocks, LocationBlock{
			Block: block,
		})
	}

	return locationBlocks
}

func findHttpBlocks(finder blockFinder) []HttpBlock {
	var httpBlocks []HttpBlock

	for _, block := range finder.FindBlocks("http") {
		httpBlocks = append(httpBlocks, HttpBlock{
			Block: block,
		})
	}

	return httpBlocks
}

func findUpstreamBlocks(finder blockFinder) []UpstreamBlock {
	var upstreamBlocks []UpstreamBlock

	for _, block := range finder.FindBlocks("upstream") {
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

func newBlock(container entryContainer, config *Config, name string, parameters []string) Block {
	rawBlock := &rawparser.BlockDirective{
		Identifier: name,
		Content:    &rawparser.BlockContent{},
	}
	rawBlock.SetParameters(parameters)

	block := Block{
		config:   config,
		rawBlock: rawBlock,
	}

	entries := container.GetEntries()
	entries = append(entries, &rawparser.Entry{
		StartNewLines:  []string{"\n"},
		BlockDirective: rawBlock,
		EndNewLines:    []string{"\n"},
	})
	container.SetEntries(entries)

	return block
}
