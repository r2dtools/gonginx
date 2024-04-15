package config

import (
	"strings"
)

type ServerBlock struct {
	Block
}

func (s *ServerBlock) GetServerNames() []string {
	serverNames := []string{}

	directives := s.FindDirectives("server_name")

	if len(directives) == 0 {
		return serverNames
	}

	for _, value := range directives[0].GetValues() {
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

		for _, value := range listenDirective.GetValues() {
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

func (s *ServerBlock) IsIpv6Enabled() bool {
	addresses := s.GetAddresses()

	for _, address := range addresses {
		if address.IsIpv6 {
			return true
		}
	}

	return false
}

func (s *ServerBlock) IsIpv4Enabled() bool {
	addresses := s.GetAddresses()

	if len(addresses) == 0 {
		return true
	}

	for _, address := range addresses {
		if !address.IsIpv6 {
			return true
		}
	}

	return false
}

func (s *ServerBlock) GetAddresses() []ServerAddress {
	listens := s.GetListens()
	addresses := []ServerAddress{}

	for _, listen := range listens {
		address := CreateServerAddressFromString(listen.HostPort)
		addresses = append(addresses, address)
	}

	return addresses
}

func (s *ServerBlock) HasSSL() bool {
	for _, listen := range s.GetListens() {
		if listen.Ssl {
			return true
		}
	}

	return false
}

func (s *ServerBlock) FindLocationBlocks() []LocationBlock {
	return findLocationBlocks(&s.Block)
}
