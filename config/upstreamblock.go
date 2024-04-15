package config

type UpstreamBlock struct {
	Block
}

func (b *UpstreamBlock) GetUpstreamName() string {
	parameters := b.GetParameters()

	if len(parameters) > 0 {
		return parameters[0]
	}

	return ""
}

func (b *UpstreamBlock) SetUpstreamName(name string) {
	parameters := b.GetParameters()

	if len(parameters) == 0 {
		parameters = []string{name}
	} else {
		parameters[0] = name
	}

	b.rawBlock.SetParameters(parameters)
}

func (b *UpstreamBlock) GetServers() []UpstreamServer {
	var servers []UpstreamServer
	serverDirectives := b.FindDirectives("server")

	for _, serverDirective := range serverDirectives {
		servers = append(servers, UpstreamServer{
			Directive: serverDirective,
		})
	}

	return servers
}

func (b *UpstreamBlock) AddServer(upstreamServer UpstreamServer) {
	b.AddDirective(upstreamServer.Directive, false)
}

func (b *UpstreamBlock) SetServer(upstreamServers []UpstreamServer) {
	b.DeleteDirectiveByName("server")

	for _, upstreamServer := range upstreamServers {
		b.AddServer(upstreamServer)
	}
}

func (b *UpstreamBlock) DeleteServer(upstreamServer UpstreamServer) {
	b.DeleteDirective(upstreamServer.Directive)
}
