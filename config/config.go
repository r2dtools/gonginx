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

	"github.com/r2dtools/gonginx/internal/rawdumper"
	"github.com/r2dtools/gonginx/internal/rawparser"
	"github.com/unknwon/com"
	"golang.org/x/exp/maps"
)

var repeatableDirectives = []string{"server_name", "listen", "include", "rewrite", "add_header"}

var ErrInvalidDirective = errors.New("entry is not a directive")

type Config struct {
	rawParser   *rawparser.RawParser
	rawDumper   *rawdumper.RawDumper
	parsedFiles map[string]*rawparser.Config
	serverRoot  string
	configRoot  string
	quiteMode   bool
}

func (c *Config) GetConfigFile(configFileName string) *ConfigFile {
	for configFilePath, config := range c.parsedFiles {
		pConfigFileName := filepath.Base(configFilePath)

		if configFileName == pConfigFileName {
			return &ConfigFile{
				FilePath:   configFilePath,
				configFile: config,
				config:     c,
			}
		}
	}

	return nil
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

		prevEntries := list.New()
		directives = append(
			directives,
			c.findDirectivesRecursivelyInLoop(directiveName, tree.Entries, prevEntries)...,
		)
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

func (c *Config) Dump() error {
	for filePath, config := range c.parsedFiles {
		content, err := c.rawDumper.Dump(config)

		if err != nil {
			return err
		}

		file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0666)

		if err != nil {
			return err
		}

		defer file.Close()

		_, err = file.WriteString(content)

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) findDirectivesRecursivelyInLoop(
	directiveName string,
	entries []*rawparser.Entry,
	prevEntries *list.List,
) []Directive {
	var directives []Directive
	entriesCount := len(entries)

	for i := 0; i < entriesCount; i++ {
		var nextEntry *rawparser.Entry
		entry := entries[i]

		if i < entriesCount-1 {
			nextEntry = entries[i+1]
		}

		directives = append(directives, c.findDirectivesRecursively(directiveName, entry, nextEntry, prevEntries)...)
	}

	return directives
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

		config, err := c.rawParser.Parse(string(content))

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
	nextEntry *rawparser.Entry,
	prevEnties *list.List,
) []Directive {
	var directives []Directive
	directive := entry.Directive
	blockDirective := entry.BlockDirective

	if prevEnties == nil {
		prevEnties = list.New()
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
					directives = append(
						directives,
						c.findDirectivesRecursivelyInLoop(directiveName, includeConfig.Entries, prevEnties)...,
					)
				}
			}
		}

		if identifier == directiveName {
			comments := c.findNearesComments(prevEnties)
			inlineComment := c.findDirectiveInlineComment(entry, nextEntry)

			if inlineComment != nil {
				comments = append(comments, *inlineComment)
			}

			directives = append(directives, Directive{
				rawDirective: directive,
				Comments:     comments,
			})

			return directives
		}
	}

	if blockDirective != nil {
		directives = append(
			directives,
			c.findDirectivesRecursivelyInLoop(directiveName, blockDirective.GetEntries(), prevEnties)...,
		)

		return directives
	}

	prevEnties.PushBack(entry)

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
				config:   c,
				rawBlock: blockDirective,
				Comments: comments,
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
		comment := c.createComment(firstEntry.Comment, CommentPosition(Inline))

		return &comment
	}

	return nil
}

func (c *Config) findDirectiveInlineComment(entry, nextEntry *rawparser.Entry) *Comment {
	if nextEntry == nil || nextEntry.Comment == nil {
		return nil
	}

	if len(entry.EndNewLines) != 0 || len(nextEntry.StartNewLines) != 0 {
		return nil
	}

	comment := c.createComment(nextEntry.Comment, CommentPosition(Inline))

	return &comment
}

func (c *Config) findNearesComments(entryList *list.List) []Comment {
	var commets []Comment

	for element := entryList.Back(); element != nil; element = element.Prev() {
		entry := element.Value.(*rawparser.Entry)

		if entry.Comment == nil {
			break
		}

		comment := c.createComment(entry.Comment, CommentPosition(Before))
		commets = append([]Comment{comment}, commets...)

	}

	return commets
}

func (c *Config) createComment(rawComment *rawparser.Comment, position CommentPosition) Comment {
	return Comment{
		rawCommet: rawComment,
		Content:   strings.Trim(rawComment.Value, "\n# "),
		Position:  position,
	}
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

	rawParser, err := rawparser.GetRawParser()

	if err != nil {
		return nil, err
	}

	parser := Config{
		rawParser:  rawParser,
		rawDumper:  &rawdumper.RawDumper{},
		serverRoot: serverRootPath,
		configRoot: configFilePath,
		quiteMode:  quiteMode,
	}

	if err := parser.parse(); err != nil {
		return nil, err
	}

	return &parser, nil
}
