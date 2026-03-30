package goast

import (
	"reflect"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
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

type serializableField struct {
	index            int
	name             string
	flattenAnonymous bool
}

var serializableFieldCache sync.Map

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

// Serializer converts AST nodes to maps with enriched metadata.
type Serializer struct {
	sf                   *ast.SourceFile
	text                 string
	factory              *ast.NodeFactory
	lineStarts           []core.TextPos
	positionMap          *ast.PositionMap
	lineStartUTF16Offset []int
	leadingCommentCache  map[int][]any
	trailingCommentCache map[int][]any
}

// NewSerializer creates a Serializer for the given SourceFile.
// If sf is nil, enrichments (loc, comments) that require it are skipped.
func NewSerializer(sf *ast.SourceFile) *Serializer {
	s := &Serializer{sf: sf}
	if sf != nil {
		s.text = sf.Text()
		s.factory = ast.NewNodeFactory(ast.NodeFactoryHooks{})
		s.lineStarts = scanner.GetECMALineStarts(sf)
		// SourceFile.GetPositionMap() can incorrectly treat unicode input as ASCII-only
		// in this parser path, so compute a fresh map for serialization correctness.
		s.positionMap = ast.ComputePositionMap(s.text)
		s.lineStartUTF16Offset = make([]int, len(s.lineStarts))
		for i, start := range s.lineStarts {
			s.lineStartUTF16Offset[i] = s.positionMap.UTF8ToUTF16(int(start))
		}
		s.leadingCommentCache = make(map[int][]any)
		s.trailingCommentCache = make(map[int][]any)
	}
	return s
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

// EncodeOffset converts a byte offset from typescript-go into a UTF-16 code unit offset.
func (s *Serializer) EncodeOffset(pos int) int {
	if s.positionMap == nil {
		return pos
	}
	return s.positionMap.UTF8ToUTF16(pos)
}

type encodedPosition struct {
	offset int
	line   int
	column int
}

func (s *Serializer) encodePosition(pos int) encodedPosition {
	offset := s.EncodeOffset(pos)
	if len(s.lineStarts) == 0 {
		return encodedPosition{offset: offset}
	}

	line := scanner.ComputeLineOfPosition(s.lineStarts, pos)
	return encodedPosition{
		offset: offset,
		line:   line,
		column: offset - s.lineStartUTF16Offset[line],
	}
}

// SerializeNode converts an AST node to a map with Go type names and enrichments.
func (s *Serializer) SerializeNode(node *ast.Node) map[string]any {
	if node == nil {
		return nil
	}

	start := s.encodePosition(node.Pos())
	end := s.encodePosition(node.End())
	result := map[string]any{
		"type":  KindName(node.Kind),
		"Kind":  node.KindString(),
		"start": start.offset,
		"end":   end.offset,
	}

	// Enrichment: line/column positions
	if s.sf != nil {
		result["loc"] = map[string]any{
			"startLine":   start.line,
			"startColumn": start.column,
			"endLine":     end.line,
			"endColumn":   end.column,
		}
	}

	// Enrichment: node flags
	if flags := decodeNodeFlags(node.Flags); flags != nil {
		result["flags"] = flags
	}

	// Enrichment: comments
	if s.factory != nil {
		if leading := s.leadingComments(node.Pos()); leading != nil {
			result["leadingComments"] = leading
		}
		if trailing := s.trailingComments(node.End()); trailing != nil {
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
	for _, field := range serializableFields(val.Type()) {
		fieldVal := val.Field(field.index)
		if field.flattenAnonymous {
			s.serializeStructFields(fieldVal, result)
			continue
		}

		serialized := s.serializeFieldValue(fieldVal)
		if serialized != nil {
			result[field.name] = serialized
		}
	}
}

func (s *Serializer) serializeFieldValue(fieldVal reflect.Value) any {
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
			for j := 0; j < fieldVal.Len(); j++ {
				nodes[j] = fieldVal.Index(j).Interface().(*ast.Node)
			}
			return s.SerializeNodeSlice(nodes)
		}
		return nil
	default:
		return nil
	}
}

func serializableFields(t reflect.Type) []serializableField {
	if cached, ok := serializableFieldCache.Load(t); ok {
		return cached.([]serializableField)
	}

	fields := make([]serializableField, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		if baseTypeNames[field.Name] || skipFieldNames[field.Name] {
			continue
		}
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			if baseTypeNames[field.Type.Name()] {
				continue
			}
			fields = append(fields, serializableField{
				index:            i,
				flattenAnonymous: true,
			})
			continue
		}
		fields = append(fields, serializableField{
			index: i,
			name:  field.Name,
		})
	}

	actual, _ := serializableFieldCache.LoadOrStore(t, fields)
	return actual.([]serializableField)
}

func commentKindToString(kind ast.Kind) string {
	if kind == ast.KindSingleLineCommentTrivia {
		return "line"
	}
	return "block"
}

func (s *Serializer) leadingComments(pos int) []any {
	if comments, ok := s.leadingCommentCache[pos]; ok {
		return comments
	}

	var result []any
	for cr := range scanner.GetLeadingCommentRanges(s.factory, s.text, pos) {
		result = append(result, map[string]any{
			"type":  commentKindToString(cr.Kind),
			"text":  s.text[cr.Pos():cr.End()],
			"start": s.EncodeOffset(cr.Pos()),
			"end":   s.EncodeOffset(cr.End()),
		})
	}

	s.leadingCommentCache[pos] = result
	return result
}

func (s *Serializer) trailingComments(pos int) []any {
	if comments, ok := s.trailingCommentCache[pos]; ok {
		return comments
	}

	var result []any
	for cr := range scanner.GetTrailingCommentRanges(s.factory, s.text, pos) {
		result = append(result, map[string]any{
			"type":  commentKindToString(cr.Kind),
			"text":  s.text[cr.Pos():cr.End()],
			"start": s.EncodeOffset(cr.Pos()),
			"end":   s.EncodeOffset(cr.End()),
		})
	}

	s.trailingCommentCache[pos] = result
	return result
}

func decodeNodeFlags(flags ast.NodeFlags) []string {
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
