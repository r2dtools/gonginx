package rawdumper

import (
	"errors"
	"strings"

	"github.com/r2dtools/gonginx/internal/rawparser"
)

const (
	tab   = "    "
	space = " "
)

type RawDumper struct {
	nestingLevel int
}

func (d *RawDumper) Dump(config *rawparser.Config) (string, error) {
	if config == nil {
		return "", errors.New("config is empty")
	}

	result := d.dumpEntries(config.Entries)

	return result, nil
}

func (d *RawDumper) dumpEntries(entries []*rawparser.Entry) string {
	var result string

	for _, entry := range entries {
		if entry != nil {
			result += strings.Join(entry.StartNewLines, "")
			result += d.dumpEntry(entry)
			result += strings.Join(entry.EndNewLines, "")
		}
	}

	return result
}

func (d *RawDumper) dumpEntry(entry *rawparser.Entry) string {
	result := ""

	if entry.BlockDirective != nil {
		result += d.dumpBlockDirective(entry)
	} else if entry.Directive != nil {
		result += d.dumpDirective(entry)
	} else if entry.Comment != nil {
		result += d.dumpComment(entry)
	}

	return result
}

func (d *RawDumper) dumpBlockDirective(entry *rawparser.Entry) string {
	result := d.getCurrentIdent() + entry.GetIdentifier()
	blockDirective := entry.BlockDirective
	parameters := strings.Join(blockDirective.GetParametersExpressions(), space)

	if parameters != "" {
		result += space + parameters
	}

	result += space + "{"

	if blockDirective.Content != nil {
		d.increaseNestingLevel()
		result += d.dumpEntries(blockDirective.GetEntries())
		d.decreaseNestingLevel()
	}

	result += d.getCurrentIdent() + "}"

	return result
}

func (d *RawDumper) dumpDirective(entry *rawparser.Entry) string {
	expression := strings.Join(entry.Directive.GetExpressions(), space)

	return d.getCurrentIdent() + entry.GetIdentifier() + space + expression + ";"
}

func (d *RawDumper) dumpComment(entry *rawparser.Entry) string {
	return d.getCurrentIdent() + entry.Comment.Value
}

func (d *RawDumper) getCurrentIdent() string {
	return strings.Repeat(tab, d.nestingLevel)
}

func (d *RawDumper) increaseNestingLevel() {
	d.nestingLevel++
}

func (d *RawDumper) decreaseNestingLevel() {
	if d.nestingLevel > 0 {
		d.nestingLevel--
	}
}
