package config

import (
	"path/filepath"
	"strings"
)

type Host struct {
	FilePath,
	ServerName,
	DocRoot string
	Addresses map[string]HostAddress
	Aliases   []string
	Ssl       bool
}

// GetConfigName returns config name of a host
func (h Host) GetConfigName() string {
	return filepath.Base(h.FilePath)
}

// GetAddressesString return address as a string: "172.10.52.2:80 172.10.52.3:8080"
func (h Host) GetAddressesString(hostsOnly bool) string {
	var addresses []string
	for _, address := range h.Addresses {
		if hostsOnly {
			addresses = append(addresses, address.Host)
		} else {
			addresses = append(addresses, address.ToString())
		}
	}

	return strings.Join(addresses, " ")
}

func (h Host) IsIpv6Enabled() bool {
	for _, address := range h.Addresses {
		if address.IsIpv6 {
			return true
		}
	}

	return false
}

func (h Host) IsIpv4Enabled() bool {
	if len(h.Addresses) == 0 {
		return true
	}

	for _, address := range h.Addresses {
		if !address.IsIpv6 {
			return true
		}
	}

	return false
}
