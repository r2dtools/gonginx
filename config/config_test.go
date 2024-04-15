package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var example2ConfigFilePath = "../test/nginx/sites-enabled/example2.com.conf"
var example2ConfigFileName = "example2.com.conf"

func TestFindHttpBlocks(t *testing.T) {
	config := parseConfig(t)
	httpBlocks := config.FindHttpBlocks()

	assert.Len(t, httpBlocks, 1)
	httpBlock := httpBlocks[0]

	assert.Equal(t, "http", httpBlock.GetName())
	assert.Empty(t, httpBlock.GetParameters())
	assert.Len(t, httpBlock.Comments, 2)

	commentBefore := httpBlock.Comments[0]
	assert.Equal(t, "http block", commentBefore.Content)

	inlineComment := httpBlock.Comments[1]
	assert.Equal(t, "http block inline comment", inlineComment.Content)
}

func TestFindServerBlocks(t *testing.T) {
	config := parseConfig(t)
	serverBlocks := config.FindServerBlocks()

	assert.Len(t, serverBlocks, 8)
}

func TestFindLocationBlocks(t *testing.T) {
	config := parseConfig(t)
	locationBlocks := config.FindLocationBlocks()

	assert.Len(t, locationBlocks, 16)
}

func TestFindServerBlocksByName(t *testing.T) {
	config := parseConfig(t)

	serverBlocks := config.FindServerBlocksByServerName("example2.com")
	assert.Len(t, serverBlocks, 1)

	serverBlocks = config.FindServerBlocksByServerName("example.com")
	assert.Len(t, serverBlocks, 2)

	serverBlocks = config.FindServerBlocksByServerName("*.example.com")
	assert.Len(t, serverBlocks, 1)

	serverBlocks = config.FindServerBlocksByServerName("www.example.com")
	assert.Len(t, serverBlocks, 1)
}

func TestFindDirectives(t *testing.T) {
	config := parseConfig(t)

	directives := config.FindDirectives("server_name")
	assert.Len(t, directives, 8)
	assert.Equal(t, directives[0].GetValues(), []string{"example.com"})
}

func TestDump(t *testing.T) {
	config := parseConfig(t)

	err := config.Dump()
	assert.Nil(t, err)
}

func parseConfig(t *testing.T) *Config {
	config, err := GetConfig("../test/nginx", "", false)
	assert.Nilf(t, err, "could not create config: %v", err)

	return config
}

func testWithConfigFileRollback(t *testing.T, configFilePath string, testFunc func(t *testing.T)) {
	configFileContent, err := os.ReadFile(configFilePath)
	assert.Nil(t, err)

	testFunc(t)

	err = os.WriteFile(configFilePath, configFileContent, 0666)
	assert.Nil(t, err)
}
