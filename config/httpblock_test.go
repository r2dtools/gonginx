package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddUpstreamBlockInHttpBlock(t *testing.T) {
	testWithConfigFileRollback(t, nginxConfigFilePath, func(t *testing.T) {
		config, httpBlock := getHttpBlock(t)
		upstreamBlock := httpBlock.AddUpstreamBlock("backend", false)
		upstreamServer := NewUpstreamServer("backend.example.com", []string{"backup"})
		upstreamBlock.AddServer(upstreamServer)

		err := config.Dump()
		assert.Nil(t, err)
	})
}

func getHttpBlock(t *testing.T) (*Config, HttpBlock) {
	config := parseConfig(t)

	httpBlocks := config.FindHttpBlocks()
	assert.Len(t, httpBlocks, 1)

	httpBlock := httpBlocks[0]

	return config, httpBlock
}
