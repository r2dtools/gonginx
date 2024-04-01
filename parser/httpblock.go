package parser

type HttpBlock struct {
	Block
}

func (b *HttpBlock) FindBlocks(blockName string) []Block {
	return b.parser.FindBlocks(blockName)
}

func (b *HttpBlock) FindServerBlocks() []ServerBlock {
	return b.parser.FindServerBlocks()
}

func (b *HttpBlock) FindServerBlocksByServerName(serverName string) []ServerBlock {
	return b.parser.FindServerBlocksByServerName(serverName)
}

func (b *HttpBlock) FindUpstreamBlocks() []UpstreamBlock {
	return b.parser.FindUpstreamBlocks()
}

func (b *HttpBlock) FindUpstreamBlocksByName(upstreamName string) []UpstreamBlock {
	return b.parser.FindUpstreamBlocksByName(upstreamName)
}

func (b *HttpBlock) FindLocationBlocks() []LocationBlock {
	return b.parser.FindLocationBlocks()
}
