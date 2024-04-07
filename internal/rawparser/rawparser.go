package rawparser

import (
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type Config struct {
	Entries []*Entry `@@*`
}

type Entry struct {
	StartNewLines  []string        `@NewLine*`
	Comment        *Comment        `( @@`
	Directive      *Directive      `| @@`
	BlockDirective *BlockDirective `| @@ )`
	EndNewLines    []string        `@NewLine*`
}

type Comment struct {
	Value string `@Comment`
}

type Directive struct {
	Identifier string   `@Ident`
	Values     []*Value `@@*";"`
}

type BlockDirective struct {
	Identifier string        `@Ident`
	Parameters []*Value      `@@*`
	Content    *BlockContent `"{" @@ "}"`
}

type Value struct {
	Expression string `@Expression | @StringDoubleQuoted | @StringSingleQuoted`
}

type BlockContent struct {
	Entries []*Entry `@@*`
}

func (c *Config) GetEntries() []*Entry {
	entries := make([]*Entry, 0)

	if c.Entries == nil {
		return entries
	}

	return c.Entries
}

func (d *Directive) GetFirstValueStr() string {
	if len(d.Values) == 0 {
		return ""
	}

	return d.Values[0].Expression
}

func (d *Directive) GetExpressions() []string {
	return getExpressions(d.Values)
}

func (d *Directive) GetValues() []*Value {
	values := []*Value{}

	for _, value := range d.Values {
		if value != nil {
			values = append(values, value)
		}
	}

	return values
}

func (d *Directive) SetValues(expressions []string) {
	values := []*Value{}

	for _, expression := range expressions {
		values = append(values, &Value{Expression: expression})
	}

	d.Values = values
}

func (b *BlockDirective) GetEntries() []*Entry {
	entries := make([]*Entry, 0)

	if b.Content == nil {
		return entries
	}

	return b.Content.Entries
}

func (b *BlockDirective) FindEntriesWithIdentifier(identifier string) []*Entry {
	entries := []*Entry{}

	for _, entry := range b.GetEntries() {
		if entry != nil && entry.GetIdentifier() == identifier {
			entries = append(entries, entry)
		}
	}

	return entries
}

func (b *BlockDirective) GetParametersExpressions() []string {
	return getExpressions(b.Parameters)
}

func (b *BlockDirective) SetEntries(entries []*Entry) {
	if b.Content == nil {
		b.Content = &BlockContent{}
	}

	b.Content.Entries = entries
}

func (b *BlockDirective) GetEntriesByIdentifier(identifier string) []*Entry {
	entries := []*Entry{}

	for _, entry := range b.GetEntries() {
		if entry == nil {
			continue
		}

		if strings.ToLower(entry.GetIdentifier()) == identifier {
			entries = append(entries, entry)
		}
	}

	return entries
}

func (e *Entry) GetIdentifier() string {
	if e.Directive != nil {
		return e.Directive.Identifier
	}

	if e.BlockDirective != nil {
		return e.BlockDirective.Identifier
	}

	return ""
}

type RawParser struct {
	participleParser *participle.Parser[Config]
}

func (p *RawParser) Parse(content string) (*Config, error) {
	return p.participleParser.ParseString("", content)
}

func GetRawParser() (*RawParser, error) {
	def := lexer.MustStateful(lexer.Rules{
		"Root": {
			{Name: `NewLine`, Pattern: `[\r\n]+`, Action: nil},
			{Name: `whitespace`, Pattern: `[^\S\r\n]+`, Action: nil},
			{Name: `Comment`, Pattern: `(?:#)[^\n]*\n?`, Action: nil},
			{Name: "BlockEnd", Pattern: `}`, Action: nil},
			{Name: `Ident`, Pattern: `[\w\-.\/]+`, Action: lexer.Push("IdentParse")},
		},
		"IdentParse": {
			{Name: `NewLine`, Pattern: `[\r\n]+`, Action: nil},
			{Name: `whitespace`, Pattern: `[^\S\r\n]+`, Action: nil},
			{Name: `StringDoubleQuoted`, Pattern: `"[^"]*"`, Action: nil},
			{Name: `StringSingleQuoted`, Pattern: `'[^']*'`, Action: nil},
			{Name: "Semicolon", Pattern: `;`, Action: lexer.Pop()},
			{Name: "BlockStart", Pattern: `{`, Action: lexer.Pop()},
			{Name: "BlockEnd", Pattern: `}`, Action: lexer.Pop()},
			{Name: "Expression", Pattern: `[^;{}#\s]+`, Action: nil},
			{Name: `Comment`, Pattern: `(?:#)[^\n]*\n?`, Action: nil},
		},
	})

	participleParser, err := participle.Build[Config](
		participle.Lexer(def),
		participle.UseLookahead(50),
	)

	if err != nil {
		return nil, err
	}

	parser := RawParser{
		participleParser: participleParser,
	}

	return &parser, nil
}

func getExpressions(values []*Value) []string {
	expressions := []string{}

	for _, value := range values {
		if value != nil {
			expressions = append(expressions, value.Expression)
		}
	}

	return expressions
}
