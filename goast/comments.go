package goast

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// commentKindToString maps ast.Kind to "line" or "block".
func commentKindToString(kind ast.Kind) string {
	if kind == ast.KindSingleLineCommentTrivia {
		return "line"
	}
	return "block"
}

// CollectLeadingComments returns serialized leading comments at the given position.
// Returns nil if there are no comments.
func CollectLeadingComments(factory *ast.NodeFactory, text string, pos int) []any {
	var result []any
	for cr := range scanner.GetLeadingCommentRanges(factory, text, pos) {
		result = append(result, map[string]any{
			"type":  commentKindToString(cr.Kind),
			"text":  text[cr.Pos():cr.End()],
			"start": cr.Pos(),
			"end":   cr.End(),
		})
	}
	return result
}

// CollectTrailingComments returns serialized trailing comments at the given position.
// Returns nil if there are no comments.
func CollectTrailingComments(factory *ast.NodeFactory, text string, pos int) []any {
	var result []any
	for cr := range scanner.GetTrailingCommentRanges(factory, text, pos) {
		result = append(result, map[string]any{
			"type":  commentKindToString(cr.Kind),
			"text":  text[cr.Pos():cr.End()],
			"start": cr.Pos(),
			"end":   cr.End(),
		})
	}
	return result
}
