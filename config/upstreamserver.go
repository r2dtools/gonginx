package config

type UpstreamServer struct {
	Directive

	Address string
	Flags   []string
}
