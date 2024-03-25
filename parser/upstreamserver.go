package parser

type UpstreamServer struct {
	Directive

	Address string
	Flags   []string
}
