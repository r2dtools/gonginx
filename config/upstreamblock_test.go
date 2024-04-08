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
	assert.Equal(t, "upstream", block.Name)
	assert.Equal(t, "dynamic", block.GetUpstreamName())

	servers := block.GetServers()
	assert.Len(t, servers, 7)

	server := servers[1]
	assert.Equal(t, "server", server.Name)
	assert.Equal(t, "backend2.example.com:8080", server.Address)
	assert.ElementsMatch(t, server.Flags, []string{"fail_timeout=5s", "slow_start=30s"})
}
