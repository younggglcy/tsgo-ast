package goast

import (
	"reflect"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// Serializer converts AST nodes to maps with enriched metadata.
type Serializer struct {
	sf                   *ast.SourceFile
	text                 string
	factory              *ast.NodeFactory
	lineStartUTF16Offset []int
}

// NewSerializer creates a Serializer for the given SourceFile.
// If sf is nil, enrichments (loc, comments) that require it are skipped.
func NewSerializer(sf *ast.SourceFile) *Serializer {
	s := &Serializer{sf: sf}
	if sf != nil {
		s.text = sf.Text()
		s.factory = ast.NewNodeFactory(ast.NodeFactoryHooks{})
		lineStarts := scanner.GetECMALineStarts(sf)
		s.lineStartUTF16Offset = make([]int, len(lineStarts))
		prevBytePos := 0
		prevUTF16Offset := 0
		for i, start := range lineStarts {
			bytePos := int(start)
			if i == 0 {
				s.lineStartUTF16Offset[i] = 0
				prevBytePos = bytePos
				continue
			}
			prevUTF16Offset += int(core.UTF16Len(s.text[prevBytePos:bytePos]))
			s.lineStartUTF16Offset[i] = prevUTF16Offset
			prevBytePos = bytePos
		}
	}
	return s
}

// EncodeOffset converts a byte offset from typescript-go into a UTF-16 code unit offset.
func (s *Serializer) EncodeOffset(pos int) int {
	if s.sf == nil {
		return pos
	}
	line, column := scanner.GetECMALineAndUTF16CharacterOfPosition(s.sf, pos)
	return s.lineStartUTF16Offset[line] + int(column)
}

// SerializeNode converts an AST node to a map with Go type names and enrichments.
func (s *Serializer) SerializeNode(node *ast.Node) map[string]any {
	if node == nil {
		return nil
	}

	result := map[string]any{
		"type":  KindName(node.Kind),
		"Kind":  node.KindString(),
		"start": s.EncodeOffset(node.Pos()),
		"end":   s.EncodeOffset(node.End()),
	}

	// Enrichment: line/column positions
	if s.sf != nil {
		startLine, startCol := scanner.GetECMALineAndUTF16CharacterOfPosition(s.sf, node.Pos())
		endLine, endCol := scanner.GetECMALineAndUTF16CharacterOfPosition(s.sf, node.End())
		result["loc"] = map[string]any{
			"startLine":   startLine,
			"startColumn": int(startCol),
			"endLine":     endLine,
			"endColumn":   int(endCol),
		}
	}

	// Enrichment: node flags
	if flags := DecodeNodeFlags(node.Flags); flags != nil {
		result["flags"] = flags
	}

	// Enrichment: comments
	if s.factory != nil {
		if leading := CollectLeadingComments(s.factory, s.text, node.Pos(), s.EncodeOffset); leading != nil {
			result["leadingComments"] = leading
		}
		if trailing := CollectTrailingComments(s.factory, s.text, node.End(), s.EncodeOffset); trailing != nil {
			result["trailingComments"] = trailing
		}
	}

	// Get concrete value via As*() methods (same as before)
	concrete := GetConcreteValue(node)
	if concrete != nil {
		val := reflect.ValueOf(concrete)
		if val.Kind() == reflect.Ptr && !val.IsNil() {
			val = val.Elem()
		}
		if val.Kind() == reflect.Struct {
			s.serializeStructFields(val, result)
		}
	}

	s.enrichSharedBaseFields(node, result)

	// Add Name if accessible and not already present
	if _, hasName := result["Name"]; !hasName {
		if name := node.Name(); name != nil {
			result["Name"] = s.SerializeNode(name)
		}
	}

	// Add Modifiers if accessible and not already present
	if _, hasMods := result["Modifiers"]; !hasMods {
		if mods := node.Modifiers(); mods != nil && len(mods.Nodes) > 0 {
			result["Modifiers"] = s.SerializeNodeSlice(mods.Nodes)
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

func (s *Serializer) enrichSharedBaseFields(node *ast.Node, result map[string]any) {
	if classLike := node.ClassLikeData(); classLike != nil {
		if _, hasTypeParameters := result["TypeParameters"]; !hasTypeParameters {
			result["TypeParameters"] = s.serializeNodeList(classLike.TypeParameters)
		}
		if _, hasHeritageClauses := result["HeritageClauses"]; !hasHeritageClauses {
			result["HeritageClauses"] = s.serializeNodeList(classLike.HeritageClauses)
		}
		if _, hasMembers := result["Members"]; !hasMembers {
			result["Members"] = s.serializeNodeList(classLike.Members)
		}
	}

	if functionLike := node.FunctionLikeData(); functionLike != nil {
		if _, hasTypeParameters := result["TypeParameters"]; !hasTypeParameters {
			result["TypeParameters"] = s.serializeNodeList(functionLike.TypeParameters)
		}
		if _, hasParameters := result["Parameters"]; !hasParameters {
			result["Parameters"] = s.serializeNodeList(functionLike.Parameters)
		}
		if _, hasType := result["Type"]; !hasType && functionLike.Type != nil {
			result["Type"] = s.SerializeNode(functionLike.Type)
		}
		if _, hasFullSignature := result["FullSignature"]; !hasFullSignature && functionLike.FullSignature != nil {
			result["FullSignature"] = s.SerializeNode(functionLike.FullSignature)
		}
	}

	if body := node.BodyData(); body != nil {
		if _, hasAsteriskToken := result["AsteriskToken"]; !hasAsteriskToken && body.AsteriskToken != nil {
			result["AsteriskToken"] = s.SerializeNode(body.AsteriskToken)
		}
		if _, hasBody := result["Body"]; !hasBody && body.Body != nil {
			result["Body"] = s.SerializeNode(body.Body)
		}
	}
}

func (s *Serializer) serializeStructFields(val reflect.Value, result map[string]any) {
	t := val.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		if baseTypeNames[field.Name] || skipFieldNames[field.Name] {
			continue
		}
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			if !baseTypeNames[field.Type.Name()] {
				s.serializeStructFields(val.Field(i), result)
			}
			continue
		}

		fieldVal := val.Field(i)
		serialized := s.serializeFieldValue(field.Name, fieldVal)
		if serialized != nil {
			result[field.Name] = serialized
		}
	}
}

func (s *Serializer) serializeFieldValue(_ string, fieldVal reflect.Value) any {
	if !fieldVal.IsValid() {
		return nil
	}

	fieldType := fieldVal.Type()

	if fieldType.Kind() == reflect.Ptr {
		if fieldVal.IsNil() {
			return nil
		}
		switch fieldType {
		case nodePtrType:
			node := fieldVal.Interface().(*ast.Node)
			return s.SerializeNode(node)
		case nodeListPtrType:
			nl := fieldVal.Interface().(*ast.NodeList)
			return s.serializeNodeList(nl)
		case modifierListPtrType:
			ml := fieldVal.Interface().(*ast.ModifierList)
			if ml == nil || len(ml.Nodes) == 0 {
				return nil
			}
			return s.SerializeNodeSlice(ml.Nodes)
		default:
			return nil
		}
	}

	switch fieldType.Kind() {
	case reflect.String:
		str := fieldVal.String()
		if str == "" {
			return nil
		}
		return str
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
		if fieldType.Elem() == nodePtrType {
			nodes := make([]*ast.Node, fieldVal.Len())
			for j := range fieldVal.Len() {
				nodes[j] = fieldVal.Index(j).Interface().(*ast.Node)
			}
			return s.SerializeNodeSlice(nodes)
		}
		return nil
	default:
		return nil
	}
}

func (s *Serializer) serializeNodeList(nl *ast.NodeList) any {
	if nl == nil || len(nl.Nodes) == 0 {
		return nil
	}
	return s.SerializeNodeSlice(nl.Nodes)
}

// SerializeNodeSlice converts a slice of AST nodes to serialized maps.
func (s *Serializer) SerializeNodeSlice(nodes []*ast.Node) []any {
	result := make([]any, 0, len(nodes))
	for _, child := range nodes {
		if m := s.SerializeNode(child); m != nil {
			result = append(result, m)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
