package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirective(t *testing.T) {
	config := parseConfig(t)

	directives := config.FindDirectives("ssl_certificate")
	assert.Len(t, directives, 5)

	directive := directives[3]
	assert.Equal(t, "ssl_certificate", directive.Name)
	assert.Equal(t, "/opt/webmng/test/certificate/example.com.crt", directive.GetFirstValue())

	comments := directive.Comments
	assert.Len(t, comments, 3)
	assert.Equal(t, "SSL", comments[0].Content)
	assert.Equal(t, "Some comment", comments[1].Content)
	assert.Equal(t, "inline comment", comments[2].Content)
}

func TestDirectiveSetValue(t *testing.T) {
	testWithConfigFileRollback(t, example2ConfigFilePath, func(t *testing.T) {
		certPath := "/path/to/certificate"

		config, directive := getServerBlockDirective(t, "example2.com", "ssl_certificate_key")

		directive.SetValue(certPath)
		err := config.Dump()
		assert.Nil(t, err)

		config, directive = getServerBlockDirective(t, "example2.com", "ssl_certificate_key")
		assert.Equal(t, certPath, directive.GetFirstValue())
	})
}
