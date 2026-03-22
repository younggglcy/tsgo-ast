package goast

import (
	"reflect"

	"github.com/microsoft/typescript-go/shim/ast"
)

var (
	nodePtrType         = reflect.TypeOf((*ast.Node)(nil))
	nodeListPtrType     = reflect.TypeOf((*ast.NodeList)(nil))
	modifierListPtrType = reflect.TypeOf((*ast.ModifierList)(nil))
)

// Base type names to skip when reflecting over struct fields.
var baseTypeNames = map[string]bool{
	"NodeBase":                 true,
	"NodeDefault":              true,
	"StatementBase":            true,
	"ExpressionBase":           true,
	"DeclarationBase":          true,
	"ExportableBase":           true,
	"ModifiersBase":            true,
	"LocalsContainerBase":      true,
	"FunctionLikeBase":         true,
	"FunctionLikeWithBodyBase": true,
	"BodyBase":                 true,
	"FlowNodeBase":             true,
	"ClassElementBase":         true,
	"ClassLikeBase":            true,
	"AccessorDeclarationBase":  true,
	"NamedMemberBase":          true,
	"TypeElementBase":          true,
	"LiteralLikeBase":          true,
	"TemplateLiteralLikeBase":  true,
}

// Internal metadata fields to skip (not useful for AST visualization).
var skipFieldNames = map[string]bool{
	"EndOfFileToken":          true,
	"ExternalModuleIndicator": true,
	"IdentifierCount":         true,
	"LanguageVariant":         true,
	"ModuleAugmentations":     true,
	"NestedCJSExports":        true,
	"NodeCount":               true,
	"ReparsedClones":          true,
	"ScriptKind":              true,
	"TextCount":               true,
}

// SerializeNode converts an AST node to a map with Go type names.
// This is a backward-compatible wrapper; use Serializer for enriched output.
func SerializeNode(node *ast.Node) map[string]any {
	s := NewSerializer(nil)
	return s.SerializeNode(node)
}

// SerializeNodeSlice converts a slice of AST nodes to serialized maps.
func SerializeNodeSlice(nodes []*ast.Node) []any {
	s := NewSerializer(nil)
	return s.SerializeNodeSlice(nodes)
}
