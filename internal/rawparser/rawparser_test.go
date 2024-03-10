package rawparser

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	parser, err := GetRawParser()
	assert.Nilf(t, err, "could not create parser: %v", err)

	content, err := os.ReadFile("../../test/unit/nginx.conf")
	assert.Nilf(t, err, "could not read config file")

	parsedConfig, err := parser.Parse(string(content))
	assert.Nilf(t, err, "could not parse config: %v", err)

	expectedData := &Config{}
	data, err := os.ReadFile("../../test/unit/nginx.conf.json")
	assert.Nilf(t, err, "could not read file with expected data: %v", err)

	err = json.Unmarshal(data, expectedData)
	assert.Nilf(t, err, "could not decode expected data: %v", err)

	assert.Equal(t, expectedData, parsedConfig, "parsed data is invalid")
}
