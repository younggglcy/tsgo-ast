package goast

import "github.com/microsoft/typescript-go/shim/ast"

// nodeFlagEntry maps a single NodeFlags bit to its display name.
type nodeFlagEntry struct {
	flag ast.NodeFlags
	name string
}

// flagsTable lists flags meaningful for AST visualization.
// Internal parser context flags are excluded.
var flagsTable = []nodeFlagEntry{
	{ast.NodeFlagsLet, "let"},
	{ast.NodeFlagsConst, "const"},
	{ast.NodeFlagsUsing, "using"},
	{ast.NodeFlagsOptionalChain, "optionalChain"},
	{ast.NodeFlagsHasJSDoc, "hasJSDoc"},
	{ast.NodeFlagsAmbient, "ambient"},
	{ast.NodeFlagsAwaitContext, "awaitContext"},
	{ast.NodeFlagsYieldContext, "yieldContext"},
	{ast.NodeFlagsDecoratorContext, "decoratorContext"},
	{ast.NodeFlagsThisNodeHasError, "hasError"},
}

// DecodeNodeFlags converts a NodeFlags bitmask to a slice of human-readable strings.
// Returns nil when no meaningful flags are set (callers should omit the field).
func DecodeNodeFlags(flags ast.NodeFlags) []string {
	if flags == 0 {
		return nil
	}
	var result []string
	for _, entry := range flagsTable {
		if flags&entry.flag != 0 {
			result = append(result, entry.name)
		}
	}
	return result
}
