package parser

import (
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

var includeDirective = "include"
var repeatableDirectives = []string{"server_name", "listen", includeDirective, "rewrite", "add_header"}

var ErrInvalidDirective = errors.New("entry is not a directive")

type Parser struct {
	rawParser   *rawparser.RawParser
	parsedFiles map[string]*rawparser.Config
	serverRoot  string
	configRoot  string
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

func (p *Parser) GetServerBlocks() []ServerBlock {
	var blocks []ServerBlock
	keys := maps.Keys(p.parsedFiles)
	sort.Strings(keys)

	for _, key := range keys {
		tree, ok := p.parsedFiles[key]

		if !ok {
			continue
		}

		for _, entry := range tree.Entries {
			blocks = append(blocks, p.getServerBlocksRecursively(key, entry)...)
		}
	}

	return blocks
}

func (p *Parser) getServerBlocksRecursively(filePath string, entry *rawparser.Entry) []ServerBlock {
	var blocks []ServerBlock
	block := entry.BlockDirective

	if block == nil {
		return blocks
	}

	serverBlock := ServerBlock{
		FilePath: filePath,
		block:    block,
	}

	if strings.ToLower(entry.GetIdentifier()) == "server" {
		blocks = append(blocks, serverBlock)

		return blocks // server blocks could not be nested
	}

	for _, entry := range block.GetEntries() {
		blocks = append(blocks, p.getServerBlocksRecursively(filePath, entry)...)
	}

	return blocks
}

func GetParser(serverRoot string) (*Parser, error) {
	serverRoot, err := filepath.Abs(serverRoot)

	if err != nil {
		return nil, err
	}

	configFiles := []string{"nginx.conf"}
	var configRoot string

	for _, file := range configFiles {
		path := path.Join(serverRoot, file)

		if com.IsFile(path) {
			configRoot = path
			break
		}
	}

	if configRoot == "" {
		return nil, fmt.Errorf(
			"could not find any of the config files \"%s\" in the directory \"%s\"",
			strings.Join(configFiles, ", "),
			serverRoot,
		)
	}

	rawParser, err := rawparser.GetRawParser()

	if err != nil {
		return nil, err
	}

	parser := Parser{
		rawParser:  rawParser,
		serverRoot: serverRoot,
		configRoot: configRoot,
	}

	if err := parser.parse(); err != nil {
		return nil, err
	}

	return &parser, nil
}
