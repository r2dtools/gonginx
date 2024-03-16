package config

import (
	"github.com/r2dtools/gonginx/parser"
)

type Config interface {
	GetHosts() ([]Host, error)
}

type config struct {
	parser *parser.Parser
}

func (c *config) GetHosts() ([]Host, error) {
	var hosts []Host
	serverBlocks := c.parser.GetServerBlocks()

	for _, serverBlock := range serverBlocks {
		serverNames := serverBlock.GetServerNames()
		serverName := ""
		aliases := []string{}

		if len(serverNames) > 0 {
			serverName = serverNames[0]
			aliases = serverNames[1:]
		}

		listens := serverBlock.GetListens()
		addresses := make(map[string]HostAddress)
		ssl := false

		for _, listen := range listens {
			address := CreateHostAddressFromString(listen.HostPort)
			addresses[address.GetHash()] = address

			if listen.Ssl {
				ssl = true
			}

		}

		host := Host{
			FilePath:   serverBlock.FilePath,
			ServerName: serverName,
			DocRoot:    serverBlock.GetDocumentRoot(),
			Aliases:    aliases,
			Addresses:  addresses,
			Ssl:        ssl,
		}
		hosts = append(hosts, host)
	}

	return hosts, nil
}

func GetConfig(serverRoot string) (Config, error) {
	parser, err := parser.GetParser(serverRoot)

	if err != nil {
		return nil, err
	}

	return &config{
		parser: parser,
	}, nil
}
