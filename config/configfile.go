package config

import (
	"container/list"
	"os"

	"github.com/r2dtools/gonginx/internal/rawparser"
)

type ConfigFile struct {
	FilePath   string
	configFile *rawparser.Config
	config     *Config
}

func (c *ConfigFile) FindDirectives(directiveName string) []Directive {
	prevEntries := list.New()

	return c.config.findDirectivesRecursivelyInLoop(directiveName, c.configFile.GetEntries(), prevEntries)
}

func (c *ConfigFile) FindBlocks(blockName string) []Block {
	var blocks []Block
	prevEntries := list.New()

	for _, entry := range c.configFile.GetEntries() {
		blocks = append(blocks, c.config.findBlocksRecursively(blockName, entry, prevEntries)...)
	}

	return blocks
}

func (c *ConfigFile) FindHttpBlocks() []HttpBlock {
	return findHttpBlocks(c)
}

func (c *ConfigFile) FindServerBlocks() []ServerBlock {
	return findServerBlocks(c)
}

func (c *ConfigFile) FindServerBlocksByServerName(serverName string) []ServerBlock {
	return findServerBlocksByServerName(c, serverName)
}

func (c *ConfigFile) FindUpstreamBlocks() []UpstreamBlock {
	return findUpstreamBlocks(c)
}

func (c *ConfigFile) FindUpstreamBlocksByName(upstreamName string) []UpstreamBlock {
	return findUpstreamBlocksByName(c, upstreamName)
}

func (c *ConfigFile) DeleteDirective(directive Directive) {
	deleteDirective(c.configFile, directive)
}

func (c *ConfigFile) DeleteDirectiveByName(directiveName string) {
	deleteDirectiveByName(c.configFile, directiveName)
}

func (c *ConfigFile) AddDirective(directive Directive, begining bool) {
	addDirective(c.configFile, directive, begining)
}

func (c *ConfigFile) AddHttpBlock() HttpBlock {
	block := c.addBlock("http", nil)

	return HttpBlock{
		Block: block,
	}
}

func (c *ConfigFile) addBlock(name string, parameters []string) Block {
	return newBlock(c.configFile, c.config, name, parameters)
}

func (c *ConfigFile) Dump() error {
	content, err := c.config.rawDumper.Dump(c.configFile)

	if err != nil {
		return err
	}

	file, err := os.OpenFile(c.FilePath, os.O_WRONLY|os.O_TRUNC, 0666)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(content)

	if err != nil {
		return err
	}

	return nil
}
