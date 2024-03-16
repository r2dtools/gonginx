package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfigName(t *testing.T) {
	type HostsData struct {
		filePath, configName string
	}

	items := []HostsData{
		{"/etc/webserver/sites-available/example.com.conf", "example.com.conf"},
	}

	for _, item := range items {
		host := Host{
			FilePath: item.filePath,
		}
		configName := host.GetConfigName()
		assert.Equal(t, item.configName, configName)
	}
}
