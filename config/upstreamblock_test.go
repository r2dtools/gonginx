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
