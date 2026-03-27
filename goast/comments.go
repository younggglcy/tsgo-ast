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

func commentOffset(pos int, encodeOffset func(int) int) int {
	if encodeOffset == nil {
		return pos
	}
	return encodeOffset(pos)
}

// CollectLeadingComments returns serialized leading comments at the given position.
// Returns nil if there are no comments.
func CollectLeadingComments(factory *ast.NodeFactory, text string, pos int, encodeOffset func(int) int) []any {
	var result []any
	for cr := range scanner.GetLeadingCommentRanges(factory, text, pos) {
		result = append(result, map[string]any{
			"type":  commentKindToString(cr.Kind),
			"text":  text[cr.Pos():cr.End()],
			"start": commentOffset(cr.Pos(), encodeOffset),
			"end":   commentOffset(cr.End(), encodeOffset),
		})
	}
	return result
}

// CollectTrailingComments returns serialized trailing comments at the given position.
// Returns nil if there are no comments.
func CollectTrailingComments(factory *ast.NodeFactory, text string, pos int, encodeOffset func(int) int) []any {
	var result []any
	for cr := range scanner.GetTrailingCommentRanges(factory, text, pos) {
		result = append(result, map[string]any{
			"type":  commentKindToString(cr.Kind),
			"text":  text[cr.Pos():cr.End()],
			"start": commentOffset(cr.Pos(), encodeOffset),
			"end":   commentOffset(cr.End(), encodeOffset),
		})
	}
	return result
}
