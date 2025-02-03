package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/r2dtools/gonginxconf/internal/rawdumper"
	"github.com/r2dtools/gonginxconf/internal/rawparser"
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

		for _, entry := range tree.GetEntries() {
			directives = append(
				directives,
				c.findDirectivesRecursively(directiveName, tree, entry, false)...,
			)
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

		for _, entry := range tree.Entries {
			blocks = append(blocks, c.findBlocksRecursively(blockName, key, tree, entry, false)...)
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

func (c *Config) AddConfigFile(filePath string) (*ConfigFile, error) {
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		configFile := ConfigFile{
			FilePath:   filePath,
			configFile: &rawparser.Config{},
			config:     c,
		}

		return &configFile, nil
	}

	return nil, fmt.Errorf("file %s already exists", filePath)
}

func (c *Config) ParseFile(filePath string) error {
	return c.parseRecursively(filePath)
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
	var files []string

	stat, err := os.Stat(filePath)

	if err == nil && stat.Mode().IsRegular() {
		files = []string{filePath}
	} else {
		files, err = filepath.Glob(filePath)

		if err != nil {
			return nil, err
		}
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
	container entryContainer,
	entry *rawparser.Entry,
	withInclude bool,
) []Directive {
	var directives []Directive
	directive := entry.Directive
	blockDirective := entry.BlockDirective

	if directive != nil {
		identifier := directive.Identifier

		if withInclude && identifier == "include" {
			include := c.getAbsPath(directive.GetFirstValueStr())
			includeFiles, err := filepath.Glob(include)

			if err != nil {
				return directives
			}

			for _, includePath := range includeFiles {
				includeConfig, ok := c.parsedFiles[includePath]

				if !ok {
					continue
				}

				for _, entry := range includeConfig.GetEntries() {
					directives = append(
						directives,
						c.findDirectivesRecursively(directiveName, includeConfig, entry, withInclude)...,
					)
				}
			}
		}

		if identifier == directiveName {
			directives = append(directives, Directive{
				rawDirective: directive,
				container:    container,
			})

			return directives
		}
	}

	if blockDirective != nil {
		for _, bEntry := range blockDirective.GetEntries() {
			directives = append(
				directives,
				c.findDirectivesRecursively(directiveName, blockDirective, bEntry, withInclude)...,
			)
		}

		return directives
	}

	return directives
}

func (c *Config) findBlocksRecursively(
	blockName string,
	path string,
	container entryContainer,
	entry *rawparser.Entry,
	withInclude bool,
) []Block {
	var blocks []Block
	directive := entry.Directive
	blockDirective := entry.BlockDirective

	if withInclude && directive != nil && directive.Identifier == "include" {
		include := c.getAbsPath(directive.GetFirstValueStr())
		includeFiles, err := filepath.Glob(include)

		if err != nil {
			return blocks
		}

		for _, includePath := range includeFiles {
			includeConfig, ok := c.parsedFiles[includePath]

			if !ok {
				continue
			}

			for _, entry := range includeConfig.Entries {
				blocks = append(
					blocks,
					c.findBlocksRecursively(blockName, includePath, includeConfig, entry, withInclude)...,
				)
			}
		}

		return blocks
	}

	if blockDirective != nil {
		identifier := blockDirective.Identifier

		if identifier == blockName {
			blocks = append(blocks, Block{
				FilePath:  path,
				config:    c,
				container: container,
				rawBlock:  blockDirective,
				rawDumper: &rawdumper.RawDumper{},
			})
		} else {
			// blocks can be nested
			for _, httpBlockEntry := range blockDirective.GetEntries() {
				blocks = append(
					blocks,
					c.findBlocksRecursively(blockName, path, blockDirective, httpBlockEntry, withInclude)...,
				)
			}
		}

		return blocks
	}

	return blocks
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
