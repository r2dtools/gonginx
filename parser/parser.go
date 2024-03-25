package parser

import (
	"container/list"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/r2dtools/gonginx/internal/rawparser"
	"github.com/unknwon/com"
	"golang.org/x/exp/maps"
)

var repeatableDirectives = []string{"server_name", "listen", "include", "rewrite", "add_header"}

var ErrInvalidDirective = errors.New("entry is not a directive")

type Parser struct {
	rawParser   *rawparser.RawParser
	parsedFiles map[string]*rawparser.Config
	serverRoot  string
	configRoot  string
	quiteMode   bool
}

func (p *Parser) FindServerBlocks() []ServerBlock {
	var serverBlocks []ServerBlock

	blocks := p.FindBlocks("server")

	if len(blocks) == 0 {
		return serverBlocks
	}

	for _, block := range blocks {
		serverBlocks = append(serverBlocks, ServerBlock{
			Block: block,
		})
	}

	return serverBlocks
}

func (p *Parser) FindServerBlocksByServerName(serverName string) []ServerBlock {
	var serverBlocks []ServerBlock

	for _, serverBlock := range p.FindServerBlocks() {
		serverNames := serverBlock.GetServerNames()

		if slices.Contains(serverNames, serverName) {
			serverBlocks = append(serverBlocks, serverBlock)
		}
	}

	return serverBlocks
}

func (p *Parser) FindUpstreamBlocks() []UpstreamBlock {
	var upstreamBlocks []UpstreamBlock

	blocks := p.FindBlocks("upstream")

	if len(blocks) == 0 {
		return upstreamBlocks
	}

	for _, block := range blocks {
		upstreamBlocks = append(upstreamBlocks, UpstreamBlock{
			Block: block,
		})
	}

	return upstreamBlocks
}

func (p *Parser) FindUpstreamBlocksByName(upstreamName string) []UpstreamBlock {
	var upstreamBlocks []UpstreamBlock

	for _, upstreamBlock := range p.FindUpstreamBlocks() {
		if upstreamBlock.Name == upstreamName {
			upstreamBlocks = append(upstreamBlocks, upstreamBlock)
		}
	}

	return upstreamBlocks
}

func (p *Parser) FindDirectives(directiveName string) []Directive {
	var directives []Directive

	keys := maps.Keys(p.parsedFiles)
	sort.Strings(keys)

	for _, key := range keys {
		tree, ok := p.parsedFiles[key]

		if !ok {
			continue
		}

		entryList := list.New()

		for _, entry := range tree.Entries {
			directives = append(directives, findDirectivesRecursively(directiveName, p.parsedFiles, entry, entryList)...)
		}
	}

	return directives
}

func (p *Parser) FindBlocks(blockName string) []Block {
	var blocks []Block

	keys := maps.Keys(p.parsedFiles)
	sort.Strings(keys)

	for _, key := range keys {
		tree, ok := p.parsedFiles[key]

		if !ok {
			continue
		}

		entryList := list.New()

		for _, entry := range tree.Entries {
			blocks = append(blocks, findBlocksRecursively(blockName, p.parsedFiles, entry, entryList)...)
		}
	}

	return blocks
}

func (p *Parser) parse() error {
	p.parsedFiles = make(map[string]*rawparser.Config)

	return p.parseRecursively(p.configRoot)
}

func (p *Parser) parseRecursively(configFilePath string) error {
	configFilePathAbs := p.getAbsPath(configFilePath)
	trees, err := p.parseFilesByPath(configFilePathAbs, false)

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
					p.parseRecursively(includeFile)
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
							p.parseRecursively(includeFile)
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
									p.parseRecursively(includeFile)
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

func (p *Parser) parseFilesByPath(filePath string, override bool) ([]*rawparser.Config, error) {
	files, err := filepath.Glob(filePath)

	if err != nil {
		return nil, err
	}

	var trees []*rawparser.Config

	for _, file := range files {
		if _, ok := p.parsedFiles[file]; ok && !override {
			continue
		}

		content, err := os.ReadFile(file)

		if err != nil {
			if p.quiteMode {
				continue
			}

			return nil, err
		}

		config, err := p.rawParser.Parse(string(content))

		if err != nil {
			return nil, err
		}

		p.parsedFiles[file] = config
		trees = append(trees, config)
	}

	return trees, nil
}

func (p *Parser) getAbsPath(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}

	return filepath.Clean(filepath.Join(p.serverRoot, path))
}

func GetParser(serverRootPath, configFilePath string, quiteMode bool) (*Parser, error) {
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

	parser := Parser{
		rawParser:  rawParser,
		serverRoot: serverRootPath,
		configRoot: configFilePath,
		quiteMode:  quiteMode,
	}

	if err := parser.parse(); err != nil {
		return nil, err
	}

	return &parser, nil
}
