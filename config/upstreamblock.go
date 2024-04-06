package config

type UpstreamBlock struct {
	Block
}

func (b *UpstreamBlock) GetUpstreamName() string {
	if len(b.Parameters) > 0 {
		return b.Parameters[0]
	}

	return ""
}

func (b *UpstreamBlock) GetServers() []UpstreamServer {
	var servers []UpstreamServer

	serverDirectives := b.FindDirectives("server")

	for _, serverDirective := range serverDirectives {
		flags := []string{}
		address := serverDirective.GetFirstValue()

		values := serverDirective.Values

		if len(values) > 1 {
			flags = values[1:]
		}

		servers = append(servers, UpstreamServer{
			Directive: serverDirective,
			Address:   address,
			Flags:     flags,
		})
	}

	return servers
}
