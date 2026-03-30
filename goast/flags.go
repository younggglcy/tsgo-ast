package goast

import "github.com/microsoft/typescript-go/shim/ast"

// DecodeNodeFlags converts a NodeFlags bitmask to a slice of human-readable strings.
// Returns nil when no meaningful flags are set.
func DecodeNodeFlags(flags ast.NodeFlags) []string {
	return decodeNodeFlags(flags)
}
