package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindDirectivesInServerBlock(t *testing.T) {
	config := parseConfig(t)

	serverBlocks := config.FindServerBlocksByServerName("example2.com")
	assert.Len(t, serverBlocks, 1)

	serverBlock := serverBlocks[0]
	directives := serverBlock.FindDirectives("ssl_certificate_key")
	assert.Len(t, directives, 1)

	directive := directives[0]
	assert.Equal(t, "ssl_certificate_key", directive.GetName())
	assert.ElementsMatch(t, directive.GetValues(), []string{"/opt/webmng/test/certificate/example.com.key"})

	directives = serverBlock.FindDirectives("listen")
	assert.Len(t, directives, 4)
	assert.ElementsMatch(t, directives[2].GetValues(), []string{"[::]:443", "ssl", "ipv6only=on"})

	locations := serverBlock.FindLocationBlocks()
	assert.Len(t, locations, 1)
}

func TestServerBlock(t *testing.T) {
	config := parseConfig(t)
	serverBlocks := config.FindServerBlocksByServerName("example2.com")

	assert.Len(t, serverBlocks, 1)

	block := serverBlocks[0]
	assert.Equal(t, "server", block.GetName())
	assert.ElementsMatch(t, block.GetServerNames(), []string{"example2.com", "www.example2.com"})
	assert.Equal(t, true, block.HasSSL())
	assert.Equal(t, "/var/www/html", block.GetDocumentRoot())

	listens := block.GetListens()
	assert.Len(t, listens, 4)
}

func TestAddDirectiveInServerBlock(t *testing.T) {
	testWithConfigFileRollback(t, example2ConfigFilePath, func(t *testing.T) {
		config, serverBlock := getServerBlock(t, "example2.com")
		directive := NewDirective("test", []string{"test_value"})
		serverBlock.AddDirective(directive, true)
		err := config.Dump()
		assert.Nil(t, err)

		config, directive = getServerBlockDirective(t, "example2.com", "test")
		assert.Equal(t, "test_value", directive.GetFirstValue())
	})
}

func TestDeleteDirectiveByNameInServerBlock(t *testing.T) {
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

func TestDeleteDirectiveInServerBlock(t *testing.T) {
	testWithConfigFileRollback(t, example2ConfigFilePath, func(t *testing.T) {
		config, serverBlock := getServerBlock(t, "example2.com")
		directives := serverBlock.FindDirectives("listen")
		assert.Len(t, directives, 4)

		serverBlock.DeleteDirective(directives[2])
		err := config.Dump()
		assert.Nil(t, err)

		_, serverBlock = getServerBlock(t, "example2.com")
		directives = serverBlock.FindDirectives("listen")
		assert.Len(t, directives, 3)
		assert.Equal(t, []string{"443", "ssl"}, directives[2].GetValues())
	})
}

func TestFindServerBlockComments(t *testing.T) {
	configFile := getConfigFile(t, example2ConfigFileName)
	serverBlocks := configFile.FindServerBlocksByServerName("example2.com")
	assert.Len(t, serverBlocks, 1)

	serverBlock := serverBlocks[0]

	comments := serverBlock.FindComments()
	assert.Len(t, comments, 20)

	inlineComment := comments[len(comments)-1]
	assert.Equal(t, "inline comment", inlineComment.Content)
	assert.Equal(t, "updated by the nginx packaging team.", comments[9].Content)
}

func TestSetServerBlockComments(t *testing.T) {
	testWithConfigFileRollback(t, example2ConfigFilePath, func(t *testing.T) {
		configFile := getConfigFile(t, example2ConfigFileName)
		serverBlocks := configFile.FindServerBlocksByServerName("example2.com")
		assert.Len(t, serverBlocks, 1)

		serverBlock := serverBlocks[0]
		serverBlock.SetComments([]string{"test comment1", "test comment2", "test comment3"})
		err := configFile.Dump()
		assert.Nil(t, err)

		comments := serverBlock.FindComments()
		assert.Len(t, comments, 4)

		assert.Equal(t, "test comment1", comments[0].Content)
	})
}

func getServerBlockDirective(t *testing.T, serverName, directiveName string) (*Config, Directive) {
	config, serverBlock := getServerBlock(t, serverName)
	directives := serverBlock.FindDirectives(directiveName)
	assert.Len(t, directives, 1)

	return config, directives[0]
}

func getServerBlock(t *testing.T, serverName string) (*Config, ServerBlock) {
	config := parseConfig(t)

	serverBlocks := config.FindServerBlocksByServerName(serverName)
	assert.Len(t, serverBlocks, 1)

	serverBlock := serverBlocks[0]

	return config, serverBlock
}
