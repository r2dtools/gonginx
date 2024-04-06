package config

type LocationBlock struct {
	Block
}

func (l *LocationBlock) GetModifier() string {
	parameters := l.Parameters

	if len(parameters) > 1 {
		return parameters[0]
	}

	return ""
}

func (l *LocationBlock) GetLocationMatch() string {
	parameters := l.Parameters

	if len(parameters) > 1 {
		return parameters[1]
	}

	if len(parameters) == 1 {
		return parameters[0]
	}

	return ""
}
