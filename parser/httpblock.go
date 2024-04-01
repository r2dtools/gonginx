package parser

type HttpBlock struct {
	Block
}

func (b *HttpBlock) FindBlocks(blockName string) []Block {
	return b.Block.FindBlocks(blockName)
}

func (b *HttpBlock) FindServerBlocks() []ServerBlock {
	return findServerBlocks(&b.Block)
}

func (b *HttpBlock) FindServerBlocksByServerName(serverName string) []ServerBlock {
	return findServerBlocksByServerName(b, serverName)
}

func (b *HttpBlock) FindUpstreamBlocks() []UpstreamBlock {
	return findUpstreamBlocks(&b.Block)
}

func (b *HttpBlock) FindUpstreamBlocksByName(upstreamName string) []UpstreamBlock {
	return findUpstreamBlocksByName(b, upstreamName)
}

func (b *HttpBlock) FindLocationBlocks() []LocationBlock {
	return findLocationBlocks(&b.Block)
}
