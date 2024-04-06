package config

import (
	"container/list"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/r2dtools/gonginx/internal/rawparser"
	"github.com/unknwon/com"
	"golang.org/x/exp/maps"
)

var repeatableDirectives = []string{"server_name", "listen", "include", "rewrite", "add_header"}

var ErrInvalidDirective = errors.New("entry is not a directive")

type Config struct {
	rawConfig   *rawparser.RawParser
	parsedFiles map[string]*rawparser.Config
	serverRoot  string
	configRoot  string
	quiteMode   bool
}

func (c *Config) FindHttpBlocks() []HttpBlock {
	return findHttpBlocks(c)
}

func (c *Config) FindServerBlocks() []ServerBlock {
	return findServerBlocks(c)
}

func (c *Config) FindServerBlocksByServerName(serverName string) []ServerBlock {
	return findServerBlocksByServerName(c, serverName)
}

func (c *Config) FindUpstreamBlocks() []UpstreamBlock {
	return findUpstreamBlocks(c)
}

func (c *Config) FindUpstreamBlocksByName(upstreamName string) []UpstreamBlock {
	return findUpstreamBlocksByName(c, upstreamName)
}

func (c *Config) FindDirectives(directiveName string) []Directive {
	var directives []Directive

	keys := maps.Keys(c.parsedFiles)
	sort.Strings(keys)

	for _, key := range keys {
		tree, ok := c.parsedFiles[key]

		if !ok {
			continue
		}

		entryList := list.New()

		for _, entry := range tree.Entries {
			directives = append(directives, c.findDirectivesRecursively(directiveName, entry, entryList)...)
		}
	}

	return directives
}

func (c *Config) FindBlocks(blockName string) []Block {
	var blocks []Block

	keys := maps.Keys(c.parsedFiles)
	sort.Strings(keys)

	for _, key := range keys {
		tree, ok := c.parsedFiles[key]

		if !ok {
			continue
		}

		entryList := list.New()

		for _, entry := range tree.Entries {
			blocks = append(blocks, c.findBlocksRecursively(blockName, entry, entryList)...)
		}
	}

	return blocks
}

func (c *Config) FindLocationBlocks() []LocationBlock {
	return findLocationBlocks(c)
}

func (c *Config) parse() error {
	c.parsedFiles = make(map[string]*rawparser.Config)

	return c.parseRecursively(c.configRoot)
}

func (c *Config) parseRecursively(configFilePath string) error {
	configFilePathAbs := c.getAbsPath(configFilePath)
	trees, err := c.parseFilesByPath(configFilePathAbs, false)

	if err != nil {
		return err
	}

	for _, tree := range trees {
		for _, entry := range tree.Entries {
			identifier := strings.ToLower(entry.GetIdentifier())
			// Parse the top-level included file
			if identifier == "include" {
				if entry.Directive == nil {
					return ErrInvalidDirective
				}

				includeFile := entry.Directive.GetFirstValueStr()
				if includeFile != "" {
					c.parseRecursively(includeFile)
				}
				continue
			}

			// Look for includes in the top-level 'http'/'server' context
			if identifier == "http" || identifier == "server" {
				if entry.BlockDirective == nil {
					continue
				}

				for _, subEntry := range entry.BlockDirective.GetEntries() {
					subIdentifier := strings.ToLower(subEntry.GetIdentifier())
					if subIdentifier == "include" {
						if subEntry.Directive == nil {
							return ErrInvalidDirective
						}

						includeFile := subEntry.Directive.GetFirstValueStr()
						if includeFile != "" {
							c.parseRecursively(includeFile)
						}
						continue
					}

					// Look for includes in a 'server' context within an 'http' context
					if identifier == "http" && subIdentifier == "server" {
						if subEntry.BlockDirective == nil {
							continue
						}

						for _, serverEntry := range subEntry.BlockDirective.GetEntries() {
							if strings.ToLower(serverEntry.GetIdentifier()) == "include" {
								if serverEntry.Directive == nil {
									return ErrInvalidDirective
								}

								includeFile := serverEntry.Directive.GetFirstValueStr()
								if includeFile != "" {
									c.parseRecursively(includeFile)
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (c *Config) parseFilesByPath(filePath string, override bool) ([]*rawparser.Config, error) {
	files, err := filepath.Glob(filePath)

	if err != nil {
		return nil, err
	}

	var trees []*rawparser.Config

	for _, file := range files {
		if _, ok := c.parsedFiles[file]; ok && !override {
			continue
		}

		content, err := os.ReadFile(file)

		if err != nil {
			if c.quiteMode {
				continue
			}

			return nil, err
		}

		config, err := c.rawConfig.Parse(string(content))

		if err != nil {
			return nil, err
		}

		c.parsedFiles[file] = config
		trees = append(trees, config)
	}

	return trees, nil
}

func (c *Config) getAbsPath(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}

	return filepath.Clean(filepath.Join(c.serverRoot, path))
}

func (c *Config) findDirectivesRecursively(
	directiveName string,
	entry *rawparser.Entry,
	entryList *list.List,
) []Directive {
	var directives []Directive
	directive := entry.Directive
	blockDirective := entry.BlockDirective

	if entryList == nil {
		entryList = list.New()
	}

	if directive != nil {
		identifier := directive.Identifier

		if identifier == "include" {
			include := directive.GetFirstValueStr()
			includeFiles, err := filepath.Glob(include)

			if err != nil {
				return directives
			}

			for _, includePath := range includeFiles {
				includeConfig, ok := c.parsedFiles[includePath]

				if ok {
					for _, entry := range includeConfig.Entries {
						directives = append(
							directives,
							c.findDirectivesRecursively(directiveName, entry, nil)...,
						)
					}
				}
			}
		}

		if identifier == directiveName {
			directives = append(directives, Directive{
				rawDirective: directive,
				Name:         directive.Identifier,
				Values:       directive.GetExpressions(),
				Comments:     c.findNearesComments(entryList),
			})

			return directives
		}
	}

	if blockDirective != nil {
		for _, entry := range blockDirective.GetEntries() {
			directives = append(directives, c.findDirectivesRecursively(directiveName, entry, entryList)...)
		}

		return directives
	}

	entryList.PushBack(entry)

	return directives
}

func (c *Config) findBlocksRecursively(
	blockName string,
	entry *rawparser.Entry,
	entryList *list.List,
) []Block {
	var blocks []Block
	directive := entry.Directive
	blockDirective := entry.BlockDirective

	if entryList == nil {
		entryList = list.New()
	}

	if directive != nil && directive.Identifier == "include" {
		include := directive.GetFirstValueStr()
		includeFiles, err := filepath.Glob(include)

		if err != nil {
			return blocks
		}

		for _, includePath := range includeFiles {
			includeConfig, ok := c.parsedFiles[includePath]

			if ok {
				for _, entry := range includeConfig.Entries {
					blocks = append(
						blocks,
						c.findBlocksRecursively(blockName, entry, nil)...,
					)
				}
			}
		}

		return blocks
	}

	if blockDirective != nil {
		identifier := blockDirective.Identifier

		if identifier == blockName {
			comments := c.findNearesComments(entryList)
			inlineComment := c.findBlockInlineComment(blockDirective.Content)

			if inlineComment != nil {
				comments = append(comments, *inlineComment)
			}

			blocks = append(blocks, Block{
				config:     c,
				rawBlock:   blockDirective,
				Name:       blockDirective.Identifier,
				Parameters: blockDirective.GetParametersExpressions(),
				Comments:   comments,
			})
		} else {
			// blocks can be nested
			for _, httpBlockEntry := range blockDirective.GetEntries() {
				blocks = append(
					blocks,
					c.findBlocksRecursively(blockName, httpBlockEntry, entryList)...,
				)
			}

			entryList.PushBack(entry)
		}

		return blocks
	}

	entryList.PushBack(entry)

	return blocks
}

func (c *Config) findBlockInlineComment(content *rawparser.BlockContent) *Comment {
	if content == nil || len(content.Entries) == 0 {
		return nil
	}

	firstEntry := content.Entries[0]

	if firstEntry.Comment != nil {
		return &Comment{
			rawCommet: firstEntry.Comment,
			Content:   strings.Trim(firstEntry.Comment.Value, "\n"),
			Position:  CommentPosition(Inline),
		}
	}

	return nil
}

func (c *Config) findNearesComments(entryList *list.List) []Comment {
	var commets []Comment

	for element := entryList.Back(); element != nil; element = element.Prev() {
		entry := element.Value.(*rawparser.Entry)

		if entry.Comment == nil {
			break
		}

		comment := Comment{
			rawCommet: entry.Comment,
			Content:   strings.Trim(entry.Comment.Value, "\n"),
			Position:  CommentPosition(Before),
		}
		commets = append([]Comment{comment}, commets...)

	}

	return commets
}

func GetConfig(serverRootPath, configFilePath string, quiteMode bool) (*Config, error) {
	var err error

	if serverRootPath != "" {
		serverRootPath, err = filepath.Abs(serverRootPath)

		if err != nil {
			return nil, err
		}
	}

	if configFilePath == "" {
		configFilePath = path.Join(serverRootPath, "nginx.conf")
	}

	if !filepath.IsAbs(configFilePath) {
		configFilePath = filepath.Clean(filepath.Join(serverRootPath, configFilePath))
	}

	if !com.IsFile(configFilePath) {
		return nil, fmt.Errorf("could not find '%s' config file", configFilePath)
	}

	rawConfig, err := rawparser.GetRawParser()

	if err != nil {
		return nil, err
	}

	parser := Config{
		rawConfig:  rawConfig,
		serverRoot: serverRootPath,
		configRoot: configFilePath,
		quiteMode:  quiteMode,
	}

	if err := parser.parse(); err != nil {
		return nil, err
	}

	return &parser, nil
}
