package config

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

type ServerAddress struct {
	IsIpv6 bool
	Host   string
	Port   string
}

type Listen struct {
	HostPort string
	Ssl      bool
	Ipv6only bool
}

// CreateServerAddressFromString parses address string and returns Address structure
func CreateServerAddressFromString(addrStr string) ServerAddress {
	var host, port string
	// ipv6 addresses starts with
	if strings.HasPrefix(addrStr, "[") {
		lastIndex := strings.LastIndex(addrStr, "]")
		host = addrStr[:lastIndex+1]

		if len(addrStr) > lastIndex+2 && string(addrStr[lastIndex+1]) == ":" {
			port = addrStr[lastIndex+2:]
		}

		return ServerAddress{
			Host:   host,
			Port:   port,
			IsIpv6: true,
		}
	}

	parts := strings.Split(addrStr, ":")

	if len(parts) == 0 {
		return ServerAddress{}
	}

	if _, err := strconv.Atoi(parts[0]); err == nil {
		port = parts[0]
	} else {
		host = parts[0]

		if len(parts) > 1 {
			port = parts[1]
		}
	}

	return ServerAddress{
		Host: host,
		Port: port,
	}
}

func (a ServerAddress) IsWildcardPort() bool {
	return a.Port == "*" || a.Port == ""
}

// GetHash returns addr hash based on host an port
func (a ServerAddress) GetHash() string {
	addr := fmt.Sprintf("%s:%s", a.Host, a.Port)

	return base64.StdEncoding.EncodeToString([]byte(addr))
}

func (a ServerAddress) ToString() string {
	if a.Port != "" {
		return fmt.Sprintf("%s:%s", a.Host, a.Port)
	}

	return a.Host
}

// GetAddressWithNewPort returns new a ServerAddress instance with changed port
func (a ServerAddress) GetAddressWithNewPort(port string) ServerAddress {
	return ServerAddress{
		Host:   a.Host,
		Port:   port,
		IsIpv6: a.IsIpv6,
	}
}

// GetNormalizedHost returns normalized host.
// Normalization occurres only for ipv6 address. Ipv4 returns as is.
// For example: [fd00:dead:beaf::1] -> fd00:dead:beaf:0:0:0:0:1
func (a ServerAddress) GetNormalizedHost() string {
	if a.IsIpv6 {
		return a.GetNormalizedIpv6()
	}

	return a.Host
}

// GetNormalizedIpv6 returns normalized IPv6
// For example: [fd00:dead:beaf::1] -> fd00:dead:beaf:0:0:0:0:1
func (a ServerAddress) GetNormalizedIpv6() string {
	if !a.IsIpv6 {
		return ""
	}

	return strings.Join(a.normalizeIpv6(a.Host), ":")
}

func (a ServerAddress) IsEqual(b ServerAddress) bool {
	if a.Port != b.Port {
		return false
	}

	return a.GetNormalizedHost() == b.GetNormalizedHost()
}

func (a ServerAddress) normalizeIpv6(addr string) []string {
	addr = strings.Trim(addr, "[]")

	return a.explodeIpv6(addr)
}

func (a ServerAddress) explodeIpv6(addr string) []string {
	result := []string{"0", "0", "0", "0", "0", "0", "0", "0"}
	addrParts := strings.Split(addr, ":")
	var appendToEnd bool

	if len(addrParts) > len(result) {
		addrParts = addrParts[:len(result)]
	}

	for i, block := range addrParts {
		if block == "" {
			appendToEnd = true
			continue
		}

		if len(block) > 1 {
			block = strings.TrimLeft(block, "0")
		}

		if !appendToEnd {
			result[i] = block
		} else {
			result[len(result)-len(addrParts)+i] = block
		}
	}

	return result
}
