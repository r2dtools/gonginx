package config

type UpstreamServer struct {
	Directive
}

func (s *UpstreamServer) GetAddress() string {
	values := s.rawDirective.GetExpressions()

	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func (s *UpstreamServer) GetFlags() []string {
	values := s.rawDirective.GetExpressions()

	return values[1:]
}

func (s *UpstreamServer) SetAddress(address string) {
	values := s.rawDirective.GetExpressions()
	values[0] = address

	s.rawDirective.SetValues(values)
}

func (s *UpstreamServer) SetFlags(flags []string) {
	values := s.rawDirective.GetExpressions()

	if len(values) == 0 {
		s.rawDirective.SetValues(flags)
	} else {
		values = []string{values[0]}
		values = append(values, flags...)
		s.rawDirective.SetValues(values)
	}
}

func NewUpstreamServer(address string, flags []string) UpstreamServer {
	values := []string{address}
	values = append(values, flags...)
	directive := NewDirective("server", values)

	return UpstreamServer{
		Directive: directive,
	}
}
