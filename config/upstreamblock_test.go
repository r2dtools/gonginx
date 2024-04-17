package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpstreamBlocks(t *testing.T) {
	config := parseConfig(t)

	blocks := config.FindUpstreamBlocksByName("dynamic")
	assert.Len(t, blocks, 1)

	block := blocks[0]
	assert.Equal(t, "upstream", block.GetName())
	assert.Equal(t, "dynamic", block.GetUpstreamName())

	servers := block.GetServers()
	assert.Len(t, servers, 7)

	server := servers[1]
	assert.Equal(t, "server", server.GetName())
	assert.Equal(t, "backend2.example.com:8080", server.GetAddress())
	assert.ElementsMatch(t, server.GetFlags(), []string{"fail_timeout=5s", "slow_start=30s"})
}

func TestUpstreamBlockAddServer(t *testing.T) {
	testWithConfigFileRollback(t, nginxConfigFilePath, func(t *testing.T) {
		config := parseConfig(t)
		upstreamBlocks := config.FindUpstreamBlocksByName("dynamic")
		assert.Len(t, upstreamBlocks, 1)
		upstreamBlock := upstreamBlocks[0]
		upstreamServer := NewUpstreamServer("127.0.0.1", nil)
		upstreamBlock.SetServers([]UpstreamServer{upstreamServer})

		err := config.Dump()
		assert.Nil(t, err)
		config = parseConfig(t)
		upstreamBlocks = config.FindUpstreamBlocksByName("dynamic")
		assert.Len(t, upstreamBlocks, 1)
		upstreamBlock = upstreamBlocks[0]
		servers := upstreamBlock.GetServers()
		assert.Len(t, servers, 1)

		assert.Equal(t, "127.0.0.1", servers[0].GetAddress())
	})
}
