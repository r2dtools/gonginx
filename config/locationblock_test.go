package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocationBlock(t *testing.T) {
	config := parseConfig(t)

	serverBlocks := config.FindServerBlocksByServerName("example.com")
	assert.Len(t, serverBlocks, 2)

	serverBlock := serverBlocks[0]
	locationBlocks := serverBlock.FindLocationBlocks()
	assert.Len(t, locationBlocks, 3)

	locationBlock := locationBlocks[2]
	assert.Equal(t, "~", locationBlock.GetModifier())
	assert.Equal(t, "\\.php$", locationBlock.GetLocationMatch())
	assert.Equal(t, "location", locationBlock.GetName())
}

func TestLocationBlockSetLocationMattch(t *testing.T) {
	testWithConfigFileRollback(t, example2ConfigFilePath, func(t *testing.T) {
		configFile := getConfigFile(t, example2ConfigFileName)
		serverBlocks := configFile.FindServerBlocksByServerName("example2.com")

		assert.Len(t, serverBlocks, 1)

		block := serverBlocks[0]
		locationBlocks := block.FindLocationBlocks()
		assert.Len(t, locationBlocks, 1)

		locationBlock := locationBlocks[0]
		assert.Equal(t, "/", locationBlock.GetLocationMatch())

		locationBlock.SetLocationMatch("\\.php$")
		locationBlock.SetModifier("~")

		err := configFile.Dump()
		assert.Nil(t, err)

		configFile = getConfigFile(t, example2ConfigFileName)
		serverBlocks = configFile.FindServerBlocksByServerName("example2.com")

		assert.Len(t, serverBlocks, 1)

		block = serverBlocks[0]
		locationBlocks = block.FindLocationBlocks()
		assert.Len(t, locationBlocks, 1)

		locationBlock = locationBlocks[0]
		assert.Equal(t, "\\.php$", locationBlock.GetLocationMatch())
		assert.Equal(t, "~", locationBlock.GetModifier())
	})
}
