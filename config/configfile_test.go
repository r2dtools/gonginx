package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindServerBlocksByNameInConfigFile(t *testing.T) {
	configFile := getConfigFile(t, example2ConfigFileName)
	serverBlocks := configFile.FindServerBlocksByServerName("example2.com")
	assert.Len(t, serverBlocks, 1)
}

func TestDeleteDirectiveByNameInConfigFile(t *testing.T) {
	testWithConfigFileRollback(t, example2ConfigFilePath, func(t *testing.T) {
		config, serverBlock := getServerBlock(t, "example2.com")
		serverBlock.DeleteDirectiveByName("listen")
		err := config.Dump()
		assert.Nil(t, err)

		_, serverBlock = getServerBlock(t, "example2.com")
		directives := serverBlock.FindDirectives("listen")
		assert.Empty(t, directives)
	})

}

func TestAddDirectiveInConfigFile(t *testing.T) {
	testWithConfigFileRollback(t, example2ConfigFilePath, func(t *testing.T) {
		configFile := getConfigFile(t, example2ConfigFileName)
		configFile.AddDirective("test", []string{"test_value"}, true)
		err := configFile.Dump()
		assert.Nil(t, err)

		configFile = getConfigFile(t, example2ConfigFileName)
		directives := configFile.FindDirectives("test")
		assert.Len(t, directives, 1)
		assert.Equal(t, "test_value", directives[0].GetFirstValue())
	})
}

func getConfigFile(t *testing.T, name string) *ConfigFile {
	config := parseConfig(t)

	configFile := config.GetConfigFile(name)
	assert.NotNil(t, configFile)

	return configFile
}
