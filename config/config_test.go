package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindHttpBlocks(t *testing.T) {
	config := parseConfig(t)
	httpBlocks := config.FindHttpBlocks()

	assert.Len(t, httpBlocks, 1)
	httpBlock := httpBlocks[0]

	assert.Equal(t, "http", httpBlock.Name)
	assert.Empty(t, httpBlock.Parameters)
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

func TestFindServerBlocksDirectives(t *testing.T) {
	config := parseConfig(t)

	serverBlocks := config.FindServerBlocksByServerName("example2.com")
	assert.Len(t, serverBlocks, 1)

	serverBlock := serverBlocks[0]
	directives := serverBlock.FindDirectives("ssl_certificate_key")
	assert.Len(t, directives, 1)

	directive := directives[0]
	assert.Equal(t, "ssl_certificate_key", directive.Name)
	assert.ElementsMatch(t, directive.Values, []string{"/opt/webmng/test/certificate/example.com.key"})

	directives = serverBlock.FindDirectives("listen")
	assert.Len(t, directives, 4)
	assert.ElementsMatch(t, directives[2].Values, []string{"[::]:443", "ssl", "ipv6only=on"})

	locations := serverBlock.FindLocationBlocks()
	assert.Len(t, locations, 1)

	commments := serverBlock.Comments
	assert.Len(t, commments, 19)
}

func TestServerBlock(t *testing.T) {
	config := parseConfig(t)
	serverBlocks := config.FindServerBlocksByServerName("example2.com")

	assert.Len(t, serverBlocks, 1)

	block := serverBlocks[0]
	assert.Equal(t, "server", block.Name)
	assert.ElementsMatch(t, block.GetServerNames(), []string{"example2.com", "www.example2.com"})
	assert.Equal(t, true, block.HasSSL())
	assert.Equal(t, "/var/www/html", block.GetDocumentRoot())

	listens := block.GetListens()
	assert.Len(t, listens, 4)
}

func TestFindDirectives(t *testing.T) {
	config := parseConfig(t)

	directives := config.FindDirectives("server_name")
	assert.Len(t, directives, 8)
	assert.Equal(t, directives[0].Values, []string{"example.com"})
}

func TestDirective(t *testing.T) {
	config := parseConfig(t)

	directives := config.FindDirectives("ssl_certificate")
	assert.Len(t, directives, 5)

	directive := directives[3]
	assert.Equal(t, "ssl_certificate", directive.Name)
	assert.Equal(t, "/opt/webmng/test/certificate/example.com.crt", directive.GetFirstValue())

	comments := directive.Comments
	assert.Len(t, comments, 3)
	assert.Equal(t, "SSL", comments[0].Content)
	assert.Equal(t, "Some comment", comments[1].Content)
	assert.Equal(t, "inline comment", comments[2].Content)
}

func TestDump(t *testing.T) {
	config := parseConfig(t)

	err := config.Dump()
	assert.Nil(t, err)
}

func TestDirectiveSetValue(t *testing.T) {
	certPath := "/path/to/certificate"

	config, directive := getServerBlockDirective(t, "example2.com", "ssl_certificate_key")
	directive.SetValue(certPath)
	err := config.Dump()
	assert.Nil(t, err)

	config, directive = getServerBlockDirective(t, "example2.com", "ssl_certificate_key")
	assert.Equal(t, certPath, directive.GetFirstValue())
}

func TestAddDirective(t *testing.T) {
	config, serverBlock := getServerBlock(t, "example2.com")
	serverBlock.AddDirective("test", []string{"test_value"})
	err := config.Dump()
	assert.Nil(t, err)

	config, directive := getServerBlockDirective(t, "example2.com", "test")
	assert.Equal(t, "test_value", directive.GetFirstValue())
}

func TestDeleteDirectiveByName(t *testing.T) {
	config, serverBlock := getServerBlock(t, "example2.com")
	serverBlock.DeleteDirectiveByName("listen")
	err := config.Dump()
	assert.Nil(t, err)

	_, serverBlock = getServerBlock(t, "example2.com")
	directives := serverBlock.FindDirectives("listen")
	assert.Empty(t, directives)
}

func parseConfig(t *testing.T) *Config {
	config, err := GetConfig("../test/nginx", "", false)
	assert.Nilf(t, err, "could not create config: %v", err)

	return config
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
