package parser

import (
	"testing"

	"github.com/r2dtools/gonginx/internal/rawparser"
	"github.com/stretchr/testify/assert"
)

func TestGetBaseDirectives(t *testing.T) {
	type testData struct {
		blockDirective      *rawparser.BlockDirective
		expectedServerNames []string
	}

	var serverName = "example.com"
	var serverAlias = "alias.example.com"
	var docRoot = "/var/www/html"

	items := []testData{
		{
			blockDirective:      nil,
			expectedServerNames: []string{},
		},
		{
			blockDirective:      &rawparser.BlockDirective{Content: nil},
			expectedServerNames: []string{},
		},
		{
			blockDirective:      &rawparser.BlockDirective{Content: &rawparser.BlockContent{}},
			expectedServerNames: []string{},
		},
		{
			blockDirective: &rawparser.BlockDirective{
				Content: &rawparser.BlockContent{
					Entries: []*rawparser.Entry{
						nil,
						{
							Directive: &rawparser.Directive{
								Identifier: "server_name",
								Values: []*rawparser.Value{
									{Expression: serverName},
									{Expression: serverAlias},
								},
							},
						},
						{
							Directive: &rawparser.Directive{
								Identifier: "fake",
								Values:     nil,
							},
						},
					},
				},
			},
			expectedServerNames: []string{serverName, serverAlias},
		},
	}

	for _, item := range items {
		serverBlock := ServerBlock{block: item.blockDirective}
		assert.ElementsMatch(t, item.expectedServerNames, serverBlock.GetServerNames(), "invalid server names received")
	}

	docRootBlock := &rawparser.BlockDirective{
		Content: &rawparser.BlockContent{
			Entries: []*rawparser.Entry{
				nil,
				{
					Directive: &rawparser.Directive{
						Identifier: "root",
						Values: []*rawparser.Value{
							{Expression: docRoot},
						},
					},
				},
			},
		},
	}
	serverBlock := ServerBlock{block: docRootBlock}
	assert.Equal(t, docRoot, serverBlock.GetDocumentRoot())
}

func TestGetListens(t *testing.T) {
	type testData struct {
		block    *rawparser.BlockDirective
		expected []Listen
	}

	items := []testData{
		{
			block: &rawparser.BlockDirective{
				Content: &rawparser.BlockContent{
					Entries: []*rawparser.Entry{
						{
							Directive: &rawparser.Directive{
								Identifier: "ssl",
								Values: []*rawparser.Value{
									{Expression: "on"},
								},
							},
						},
						{
							Directive: &rawparser.Directive{
								Identifier: "listen",
								Values: []*rawparser.Value{
									{Expression: "8443"},
								},
							},
						},
						{
							Directive: &rawparser.Directive{
								Identifier: "listen",
								Values: []*rawparser.Value{
									{Expression: "[::]:8443"},
								},
							},
						},
					},
				},
			},
			expected: []Listen{
				{
					HostPort: "8443",
					Ssl:      true,
				},
				{
					HostPort: "[::]:8443",
					Ssl:      true,
				},
			},
		},
		{
			block: &rawparser.BlockDirective{
				Content: &rawparser.BlockContent{
					Entries: []*rawparser.Entry{
						{
							Directive: &rawparser.Directive{
								Identifier: "listen",
								Values: []*rawparser.Value{
									{Expression: "443"},
									{Expression: "ssl"},
									{Expression: "http2"},
								},
							},
						},
						{
							Directive: &rawparser.Directive{
								Identifier: "listen",
								Values: []*rawparser.Value{
									{Expression: "[::]:443"},
									{Expression: "ssl"},
									{Expression: "http2"},
								},
							},
						},
					},
				},
			},
			expected: []Listen{
				{
					HostPort: "443",
					Ssl:      true,
				},
				{
					HostPort: "[::]:443",
					Ssl:      true,
				},
			},
		},
		{
			block: &rawparser.BlockDirective{
				Content: &rawparser.BlockContent{
					Entries: []*rawparser.Entry{
						{
							Directive: &rawparser.Directive{
								Identifier: "listen",
								Values: []*rawparser.Value{
									{Expression: "80"},
								},
							},
						},
						{
							Directive: &rawparser.Directive{
								Identifier: "listen",
								Values: []*rawparser.Value{
									{Expression: "[::]:80"},
								},
							},
						},
					},
				},
			},
			expected: []Listen{
				{
					HostPort: "80",
					Ssl:      false,
				},
				{
					HostPort: "[::]:80",
					Ssl:      false,
				},
			},
		},
	}

	for _, item := range items {
		serverBlock := ServerBlock{block: item.block}
		listens := serverBlock.GetListens()

		assert.Equal(t, item.expected, listens)
	}
}
