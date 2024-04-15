package config

type LocationBlock struct {
	Block
}

func (l *LocationBlock) GetModifier() string {
	parameters := l.GetParameters()

	if len(parameters) > 1 {
		return parameters[0]
	}

	return ""
}

func (l *LocationBlock) SetModifier(modifier string) {
	parameters := l.GetParameters()

	if len(parameters) == 0 {
		parameters = []string{modifier}
	} else {
		parameters[0] = modifier
	}

	l.SetParameters(parameters)
}

func (l *LocationBlock) GetLocationMatch() string {
	parameters := l.GetParameters()

	if len(parameters) > 1 {
		return parameters[1]
	}

	if len(parameters) == 1 {
		return parameters[0]
	}

	return ""
}

func (l *LocationBlock) SetLocationMatch(match string) {
	parameters := l.GetParameters()

	if len(parameters) > 1 {
		parameters[1] = match
	} else {
		parameters[0] = match
	}

	l.SetParameters(parameters)
}

func (l *LocationBlock) AddLocationBlock(modifier, match string) LocationBlock {
	return addLocationBlock(&l.Block, modifier, match)
}

func (l *LocationBlock) DeleteLocationBlock(locationBlock LocationBlock) {
	deleteBlock(l.rawBlock, locationBlock.Block)
}
