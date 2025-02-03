# GoNginxConf

GoNginxConf is a library that parses Nginx configurations, enabling you to manipulate and regenerate your Nginx config files within your Go applications. GoNginxConf uses [Participle v2](https://github.com/alecthomas/participle) library under the hood.

## Install
```
go get github.com/r2dtools/gonginxconf
```

## Examples
### Parse the entire nginx configuration
```go
package main

import (
	"fmt"

	nginxConfig "github.com/r2dtools/gonginxconf/config"
)

func main() {
	config, err := nginxConfig.GetConfig("/etc/nginx", "", false)

	if err != nil {
		panic(err)
	}

	serverBlocks := config.FindServerBlocksByServerName("example.com")
	serverBlock := serverBlocks[0]

	directives := serverBlock.FindDirectives("ssl_certificate")
	directive := directives[0]

	fmt.Println(directive.GetValues())
}
```

### Add directive to a server block
```go
package main

import (
	nginxConfig "github.com/r2dtools/gonginxconf/config"
)

func main() {
	config, err := nginxConfig.GetConfig("/etc/nginx", "", false)

	if err != nil {
		panic(err)
	}

	serverBlocks := config.FindServerBlocksByServerName("example.com")
	serverBlock := serverBlocks[0]
	directive := nginxConfig.NewDirective("ssl_certificate", []string{"/path/to/certificate"})
	serverBlock.AddDirective(directive, false, true)

	err = config.Dump()

	if err != nil {
		panic(err)
	}
}
```
### Add upstream block
```go
package main

import (
	nginxConfig "github.com/r2dtools/gonginxconf/config"
)

func main() {
	config, err := nginxConfig.GetConfig("/etc/nginx", "", false)

	if err != nil {
		panic(err)
	}

	httpBlocks := config.FindHttpBlocks()
	httpBlock := httpBlocks[0]

	upstreamBlock := httpBlock.AddUpstreamBlock("my_upstream", false)
	upstreamBlock.AddServer(nginxConfig.NewUpstreamServer("127.0.0.1", []string{"weight=5"}))
	err = config.Dump()

	if err != nil {
		panic(err)
	}
}
```

<p>For more examples check tests for config package.</p>
