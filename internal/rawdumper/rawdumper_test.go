package rawdumper

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/r2dtools/gonginxconf/internal/rawparser"
	"github.com/stretchr/testify/assert"
)

func TestDump(t *testing.T) {
	config := &rawparser.Config{}
	configData, err := os.ReadFile("../../test/nginx.conf.json")
	assert.Nilf(t, err, "could not read file with parsed nginx config: %v", err)

	err = json.Unmarshal(configData, config)
	assert.Nilf(t, err, "could not decode parsed nginx config: %v", err)

	dumper := &RawDumper{}
	result, err := dumper.Dump(config)

	fmt.Println(result)
	assert.Nilf(t, err, "could not dump parsed nginx config: %v", err)

	dumpedConfig, err := os.ReadFile("../../test/nginx.dumped.conf")
	assert.Nilf(t, err, "could not read file with dumped nginx config: %v", err)
	assert.Equal(t, string(dumpedConfig), result, "dumped config is invalid")
}
