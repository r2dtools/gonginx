package config

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

func (b *HttpBlock) AddUpstreamBlock(upstreamName string) UpstreamBlock {
	block := b.addBlock("upstream", nil)

	return UpstreamBlock{
		Block: block,
	}
}

func (b *HttpBlock) AddServerBlock() ServerBlock {
	block := b.addBlock("server", nil)

	return ServerBlock{
		Block: block,
	}
}
