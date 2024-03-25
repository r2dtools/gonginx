package parser

import (
	"strings"
)

type ServerBlock struct {
	Block
}

type Listen struct {
	HostPort string
	Ssl      bool
	Ipv6only bool
}

func (s *ServerBlock) GetServerNames() []string {
	serverNames := []string{}

	directives := s.FindDirectives("server_name")

	if len(directives) == 0 {
		return serverNames
	}

	for _, value := range directives[0].Values {
		serverNames = append(serverNames, strings.TrimSpace(value))
	}

	return serverNames
}

func (s *ServerBlock) GetDocumentRoot() string {
	directives := s.FindDirectives("root")

	if len(directives) == 0 {
		return ""
	}

	return directives[0].GetFirstValue()
}

func (s *ServerBlock) GetListens() []Listen {
	listens := []Listen{}

	listenDirectives := s.FindDirectives("listen")
	sslDirectives := s.FindDirectives("ssl")
	serverSsl := false
	ipv6only := false

	// check first server block directive: ssl "on"
	for _, sslDirective := range sslDirectives {
		if sslDirective.GetFirstValue() == "on" {
			serverSsl = true
			break
		}
	}

	for _, listenDirective := range listenDirectives {
		hostPort := listenDirective.GetFirstValue()
		ssl := serverSsl

		for _, value := range listenDirective.Values {
			// check listen directive for "ssl" value
			// listen 443 ssl http2;
			if !ssl && value == "ssl" {
				ssl = true
			}

			if value == "ipv6only=on" {
				ipv6only = true
			}
		}

		listen := Listen{
			HostPort: hostPort,
			Ssl:      ssl,
			Ipv6only: ipv6only,
		}
		listens = append(listens, listen)
	}

	return listens
}
