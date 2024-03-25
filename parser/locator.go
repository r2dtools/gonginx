package parser

import (
	"container/list"

	"github.com/r2dtools/gonginx/internal/rawparser"
)

func findDirectivesRecursively(
	directiveName string,
	configMap map[string]*rawparser.Config,
	entry *rawparser.Entry,
	entryList *list.List,
) []Directive {
	var directives []Directive
	directive := entry.Directive
	blockDirective := entry.BlockDirective

	if directive != nil {
		identifier := directive.Identifier

		if identifier == "include" {
			include := directive.GetFirstValueStr()
			includeConfig, ok := configMap[include]

			if ok {
				for _, entry := range includeConfig.Entries {
					directives = append(
						directives,
						findDirectivesRecursively(directiveName, configMap, entry, entryList)...,
					)
				}
			}
		}

		if identifier == directiveName {
			directives = append(directives, Directive{
				rawDirective: directive,
				Name:         directive.Identifier,
				Values:       directive.GetExpressions(),
				Comments:     findNearesComments(entryList),
			})

			return directives
		}
	}

	if blockDirective != nil && blockDirective.Content != nil {
		for _, entry := range blockDirective.Content.Entries {
			if entry == nil {
				continue
			}

			directives = append(directives, findDirectivesRecursively(directiveName, configMap, entry, entryList)...)
		}

		return directives
	}

	entryList.PushBack(entry)

	return directives
}

func findNearesComments(entryList *list.List) []Comment {
	var commets []Comment

	for element := entryList.Back(); element != nil; element = element.Prev() {
		entry := element.Value.(*rawparser.Entry)

		if entry.Comment == nil {
			break
		}

		if len(entry.StartNewLines) != 0 {
			comment := Comment{
				rawCommet: entry.Comment,
				Content:   entry.Comment.Value,
				Position:  CommentPosition(Before),
			}
			commets = append(commets, comment)
		}

	}

	return commets
}

func findBlocksRecursively(
	blockName string,
	configMap map[string]*rawparser.Config,
	entry *rawparser.Entry,
	entryList *list.List,
) []Block {
	var blocks []Block
	directive := entry.Directive
	blockDirective := entry.BlockDirective

	if directive != nil && directive.Identifier == "include" {
		include := directive.GetFirstValueStr()
		includeConfig, ok := configMap[include]

		if ok {
			for _, entry := range includeConfig.Entries {
				blocks = append(
					blocks,
					findBlocksRecursively(blockName, configMap, entry, entryList)...,
				)
			}
		}

		return blocks
	}

	if blockDirective != nil {
		identifier := blockDirective.Identifier

		if identifier == blockName {
			blocks = append(blocks, Block{
				configMap:  configMap,
				rawBlock:   blockDirective,
				Name:       blockDirective.Identifier,
				Parameters: blockDirective.GetParametersExpressions(),
				Comments:   findNearesComments(entryList),
			})
		}

		return blocks
	}

	entryList.PushBack(entry)

	return blocks
}
