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
	assert.Equal(t, "ssl_certificate", directive.GetName())
	assert.Equal(t, "/opt/webmng/test/certificate/example.com.crt", directive.GetFirstValue())
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

func TestFindDirectiveComments(t *testing.T) {
	configFile := getConfigFile(t, exampleConfigFileName)
	directives := configFile.FindDirectives("ssl_certificate")
	assert.Len(t, directives, 2)

	directive := directives[1]

	comments := directive.FindComments()
	assert.Len(t, comments, 3)
	assert.Equal(t, "SSL", comments[0].Content)
	assert.Equal(t, "Some comment", comments[1].Content)
	assert.Equal(t, "inline comment", comments[2].Content)
}

func TestSetDirectiveComments(t *testing.T) {
	testWithConfigFileRollback(t, exampleConfigFilePath, func(t *testing.T) {
		configFile := getConfigFile(t, exampleConfigFileName)
		directives := configFile.FindDirectives("ssl_certificate")
		assert.Len(t, directives, 2)

		directive := directives[0]
		comments := directive.FindComments()
		assert.Len(t, comments, 3)

		directive.SetComments([]string{"test comment1", "test comment2", "test comment3"})
		err := configFile.Dump()
		assert.Nil(t, err)

		comments = directive.FindComments()
		assert.Len(t, comments, 4)

		assert.Equal(t, "test comment1", comments[0].Content)
	})
}
