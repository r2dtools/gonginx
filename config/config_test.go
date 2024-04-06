package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindHttpBlocks(t *testing.T) {
	parser := createParser(t)
	httpBlocks := parser.FindHttpBlocks()

	assert.Len(t, httpBlocks, 1)
	httpBlock := httpBlocks[0]

	assert.Equal(t, "http", httpBlock.Name)
	assert.Empty(t, httpBlock.Parameters)
	assert.Len(t, httpBlock.Comments, 2)

	commentBefore := httpBlock.Comments[0]
	assert.Equal(t, "# http block", commentBefore.Content)

	inlineComment := httpBlock.Comments[1]
	assert.Equal(t, "# http block inline comment", inlineComment.Content)
}

func TestFindServerBlocks(t *testing.T) {
	parser := createParser(t)
	serverBlocks := parser.FindServerBlocks()

	assert.Len(t, serverBlocks, 8)
}

func TestFindLocationBlocks(t *testing.T) {
	parser := createParser(t)
	locationBlocks := parser.FindLocationBlocks()

	assert.Len(t, locationBlocks, 16)
}

func TestUpstreamBlocks(t *testing.T) {
	parser := createParser(t)

	blocks := parser.FindUpstreamBlocksByName("dynamic")
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
	parser := createParser(t)

	serverBlocks := parser.FindServerBlocksByServerName("example2.com")
	assert.Len(t, serverBlocks, 1)

	serverBlocks = parser.FindServerBlocksByServerName("example.com")
	assert.Len(t, serverBlocks, 2)

	serverBlocks = parser.FindServerBlocksByServerName("*.example.com")
	assert.Len(t, serverBlocks, 1)

	serverBlocks = parser.FindServerBlocksByServerName("www.example.com")
	assert.Len(t, serverBlocks, 1)
}

func TestFindServerBlocksDirectives(t *testing.T) {
	parser := createParser(t)

	serverBlocks := parser.FindServerBlocksByServerName("example2.com")
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
	parser := createParser(t)
	serverBlocks := parser.FindServerBlocksByServerName("example2.com")

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
	parser := createParser(t)

	directives := parser.FindDirectives("server_name")
	assert.Len(t, directives, 8)
	assert.Equal(t, directives[0].Values, []string{"example.com"})
}

func TestDirective(t *testing.T) {
	parser := createParser(t)

	directives := parser.FindDirectives("ssl_certificate")
	assert.Len(t, directives, 5)

	directive := directives[3]
	assert.Equal(t, "ssl_certificate", directive.Name)
	assert.Equal(t, "/opt/webmng/test/certificate/example.com.crt", directive.GetFirstValue())

	comments := directive.Comments
	assert.Len(t, comments, 3)
	assert.Equal(t, "# SSL", comments[0].Content)
	assert.Equal(t, "# Some comment", comments[1].Content)
	assert.Equal(t, "# inline comment", comments[2].Content)
}

func createParser(t *testing.T) *Config {
	parser, err := GetConfig("../test/nginx", "", false)
	assert.Nilf(t, err, "could not create parser: %v", err)

	return parser
}
