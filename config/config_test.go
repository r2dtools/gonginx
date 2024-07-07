package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var nginxConfigFilePath = "../test/nginx/nginx.conf"
var example2ConfigFilePath = "../test/nginx/sites-enabled/example2.com.conf"
var exampleConfigFilePath = "../test/nginx/sites-enabled/example.com.conf"
var example2ConfigFileName = "example2.com.conf"
var exampleConfigFileName = "example.com.conf"

func TestFindHttpBlocks(t *testing.T) {
	config := parseConfig(t)
	httpBlocks := config.FindHttpBlocks()

	assert.Len(t, httpBlocks, 1)
	httpBlock := httpBlocks[0]

	assert.Equal(t, "http", httpBlock.GetName())
	assert.Empty(t, httpBlock.GetParameters())
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

func TestAddConfigFile(t *testing.T) {
	configFilePath := "../test/nginx/sites-enabled/example3.com.conf"

	config := parseConfig(t)
	configFile, err := config.AddConfigFile(configFilePath)
	assert.Nil(t, err)

	directive := NewDirective("directive", []string{"test"})
	configFile.AddDirective(directive, true)

	httpBlock := configFile.AddHttpBlock()
	directive = NewDirective("http_directive", []string{"http", "directive"})
	httpBlock.AddDirective(directive, false)

	err = configFile.Dump()
	assert.Nil(t, err)
	defer os.Remove(configFilePath)

	config = parseConfig(t)
	configFile = config.GetConfigFile("example3.com.conf")
	assert.NotNil(t, configFile)
	directives := configFile.FindDirectives("directive")
	assert.Len(t, directives, 1)
	assert.Equal(t, "directive", directives[0].GetName())
	assert.Equal(t, []string{"test"}, directives[0].GetValues())

	httpBlocks := configFile.FindHttpBlocks()
	assert.Len(t, httpBlocks, 1)
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
