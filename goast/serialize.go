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
func SerializeNode(node *ast.Node) map[string]any {
	if node == nil {
		return nil
	}

	result := map[string]any{
		"type":  KindName(node.Kind),
		"start": node.Pos(),
		"end":   node.End(),
	}

	// Get concrete value via As*() methods
	concrete := GetConcreteValue(node)
	if concrete != nil {
		val := reflect.ValueOf(concrete)
		if val.Kind() == reflect.Ptr && !val.IsNil() {
			val = val.Elem()
		}
		if val.Kind() == reflect.Struct {
			serializeStructFields(val, result)
		}
	}

	// Add Name if accessible and not already present
	if _, hasName := result["Name"]; !hasName {
		if name := node.Name(); name != nil {
			result["Name"] = SerializeNode(name)
		}
	}

	// Add Modifiers if accessible and not already present
	if _, hasMods := result["Modifiers"]; !hasMods {
		if mods := node.Modifiers(); mods != nil && len(mods.Nodes) > 0 {
			result["Modifiers"] = SerializeNodeSlice(mods.Nodes)
		}
	}

	// Add text for identifiers
	if node.Kind == ast.KindIdentifier {
		if id := node.AsIdentifier(); id != nil {
			result["text"] = id.Text
		}
	}

	// Add text for literals via LiteralLikeData
	if data := node.LiteralLikeData(); data != nil {
		result["text"] = data.Text
	}

	return result
}

func serializeStructFields(val reflect.Value, result map[string]any) {
	t := val.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		if baseTypeNames[field.Name] || skipFieldNames[field.Name] {
			continue
		}
		// Skip base types that are embedded structs
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			// Recurse into exported base types to find useful fields
			if !baseTypeNames[field.Type.Name()] {
				serializeStructFields(val.Field(i), result)
			}
			continue
		}

		fieldVal := val.Field(i)
		serialized := serializeFieldValue(field.Name, fieldVal)
		if serialized != nil {
			result[field.Name] = serialized
		}
	}
}

func serializeFieldValue(_ string, fieldVal reflect.Value) any {
	if !fieldVal.IsValid() {
		return nil
	}

	fieldType := fieldVal.Type()

	// Check pointer types first
	if fieldType.Kind() == reflect.Ptr {
		if fieldVal.IsNil() {
			return nil
		}
		switch fieldType {
		case nodePtrType:
			node := fieldVal.Interface().(*ast.Node)
			return SerializeNode(node)
		case nodeListPtrType:
			nl := fieldVal.Interface().(*ast.NodeList)
			return serializeNodeList(nl)
		case modifierListPtrType:
			ml := fieldVal.Interface().(*ast.ModifierList)
			if ml == nil || len(ml.Nodes) == 0 {
				return nil
			}
			return SerializeNodeSlice(ml.Nodes)
		default:
			// Unknown pointer type (FlowNode, Symbol, etc.) — skip
			return nil
		}
	}

	switch fieldType.Kind() {
	case reflect.String:
		s := fieldVal.String()
		if s == "" {
			return nil
		}
		return s
	case reflect.Bool:
		b := fieldVal.Bool()
		if !b {
			return nil
		}
		return b
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v := fieldVal.Int()
		if v == 0 {
			return nil
		}
		return v
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v := fieldVal.Uint()
		if v == 0 {
			return nil
		}
		return v
	case reflect.Slice:
		// Handle []*Node slices
		if fieldType.Elem() == nodePtrType {
			nodes := make([]*ast.Node, fieldVal.Len())
			for j := range fieldVal.Len() {
				nodes[j] = fieldVal.Index(j).Interface().(*ast.Node)
			}
			return SerializeNodeSlice(nodes)
		}
		return nil
	default:
		return nil
	}
}

func serializeNodeList(nl *ast.NodeList) any {
	if nl == nil || len(nl.Nodes) == 0 {
		return nil
	}
	return SerializeNodeSlice(nl.Nodes)
}

// SerializeNodeSlice converts a slice of AST nodes to serialized maps.
// Returns []any (not []map[string]any) for js.ValueOf() compatibility.
func SerializeNodeSlice(nodes []*ast.Node) []any {
	result := make([]any, 0, len(nodes))
	for _, child := range nodes {
		if s := SerializeNode(child); s != nil {
			result = append(result, s)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
