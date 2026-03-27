package goast_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/younggglcy/tsgo-ast/goast"
)

func parseSource(code string) *ast.SourceFile {
	return parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.tsx",
		Path:     tspath.Path("/test.tsx"),
	}, code, core.ScriptKindTSX)
}

func mustNodeMap(t *testing.T, value any, message string) map[string]any {
	t.Helper()
	node, ok := value.(map[string]any)
	if !ok {
		t.Fatal(message)
	}
	return node
}

func mustNodeSlice(t *testing.T, value any, message string) []any {
	t.Helper()
	nodes, ok := value.([]any)
	if !ok {
		t.Fatal(message)
	}
	return nodes
}

func buildLongSingleLineSource(repeat int) string {
	var builder strings.Builder
	builder.Grow(repeat * 24)
	for i := range repeat {
		builder.WriteString("const café")
		builder.WriteString(strconv.Itoa(i))
		builder.WriteString(` = "🎉";`)
	}
	return builder.String()
}

func TestSerializerLocField(t *testing.T) {
	sf := parseSource("const x = 1")
	s := goast.NewSerializer(sf)
	result := s.SerializeNode(sf.AsNode())

	loc, ok := result["loc"].(map[string]any)
	if !ok {
		t.Fatal("expected loc field on root node")
	}
	if loc["startLine"] == nil || loc["startColumn"] == nil {
		t.Error("expected startLine and startColumn in loc")
	}
	if loc["endLine"] == nil || loc["endColumn"] == nil {
		t.Error("expected endLine and endColumn in loc")
	}
}

func TestSerializerFlagsField(t *testing.T) {
	sf := parseSource("const x = 1")
	s := goast.NewSerializer(sf)
	result := s.SerializeNode(sf.AsNode())

	stmts, ok := result["Statements"].([]any)
	if !ok || len(stmts) == 0 {
		t.Fatal("expected Statements")
	}
	varStmt := mustNodeMap(t, stmts[0], "expected map for first statement")
	declList := mustNodeMap(t, varStmt["DeclarationList"], "expected DeclarationList")
	flags, ok := declList["flags"].([]string)
	if !ok {
		t.Fatal("expected flags on VariableDeclarationList for const")
	}
	found := false
	for _, f := range flags {
		if f == "const" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'const' in flags, got %v", flags)
	}
}

func TestSerializerComments(t *testing.T) {
	sf := parseSource("// hello\nconst x = 1")
	s := goast.NewSerializer(sf)
	result := s.SerializeNode(sf.AsNode())

	stmts, ok := result["Statements"].([]any)
	if !ok || len(stmts) == 0 {
		t.Fatal("expected Statements")
	}
	varStmt := mustNodeMap(t, stmts[0], "expected map for first statement")
	leading, ok := varStmt["leadingComments"].([]any)
	if !ok || len(leading) == 0 {
		t.Fatal("expected leadingComments on statement after comment")
	}
	comment := mustNodeMap(t, leading[0], "expected comment map")
	if comment["type"] != "line" {
		t.Errorf("expected comment type 'line', got %v", comment["type"])
	}
}

func TestSerializerIncludesFunctionLikeBaseFields(t *testing.T) {
	sf := parseSource(`const fn = (value: number) => value + 1`)
	s := goast.NewSerializer(sf)
	result := s.SerializeNode(sf.AsNode())

	statements := mustNodeSlice(t, result["Statements"], "expected Statements")
	varStmt := mustNodeMap(t, statements[0], "expected variable statement")
	declList := mustNodeMap(t, varStmt["DeclarationList"], "expected DeclarationList")
	declarations := mustNodeSlice(t, declList["Declarations"], "expected Declarations")
	decl := mustNodeMap(t, declarations[0], "expected declaration")
	arrowFn := mustNodeMap(t, decl["Initializer"], "expected arrow function initializer")

	if arrowFn["Kind"] != "KindArrowFunction" {
		t.Fatalf("expected KindArrowFunction, got %v", arrowFn["Kind"])
	}
	parameters := mustNodeSlice(t, arrowFn["Parameters"], "expected Parameters on arrow function")
	if len(parameters) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(parameters))
	}
	parameter := mustNodeMap(t, parameters[0], "expected parameter node")
	if _, ok := arrowFn["Body"].(map[string]any); !ok {
		t.Fatalf("expected Body on arrow function, got %T", arrowFn["Body"])
	}
	if _, ok := parameter["Type"].(map[string]any); !ok {
		t.Fatalf("expected parameter Type, got %T", parameter["Type"])
	}
}

func TestSerializerIncludesClassLikeBaseFields(t *testing.T) {
	sf := parseSource(`class Foo extends Bar { x = 1 }`)
	s := goast.NewSerializer(sf)
	result := s.SerializeNode(sf.AsNode())

	statements := mustNodeSlice(t, result["Statements"], "expected Statements")
	classDecl := mustNodeMap(t, statements[0], "expected class declaration")

	if classDecl["Kind"] != "KindClassDeclaration" {
		t.Fatalf("expected KindClassDeclaration, got %v", classDecl["Kind"])
	}
	heritage := mustNodeSlice(t, classDecl["HeritageClauses"], "expected HeritageClauses")
	if len(heritage) != 1 {
		t.Fatalf("expected 1 heritage clause, got %d", len(heritage))
	}
	members := mustNodeSlice(t, classDecl["Members"], "expected Members")
	if len(members) != 1 {
		t.Fatalf("expected 1 class member, got %d", len(members))
	}
}

func TestSerializerUsesUTF16Offsets(t *testing.T) {
	code := `const skull = "💀"; const next = 1`
	sf := parseSource(code)
	s := goast.NewSerializer(sf)
	result := s.SerializeNode(sf.AsNode())

	statements := mustNodeSlice(t, result["Statements"], "expected Statements")
	if len(statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(statements))
	}

	secondStmt := mustNodeMap(t, statements[1], "expected second statement")
	declList := mustNodeMap(t, secondStmt["DeclarationList"], "expected DeclarationList")
	declarations := mustNodeSlice(t, declList["Declarations"], "expected Declarations")
	decl := mustNodeMap(t, declarations[0], "expected declaration")
	name := mustNodeMap(t, decl["Name"], "expected Name node")
	loc := mustNodeMap(t, name["loc"], "expected Name.loc")
	nameNode := sf.Statements.Nodes[1].
		AsVariableStatement().
		DeclarationList.
		AsVariableDeclarationList().
		Declarations.Nodes[0].
		AsVariableDeclaration().
		Name()

	byteStart := nameNode.Pos()
	byteEnd := nameNode.End()
	expectedLine, expectedStartColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, byteStart)
	expectedEndLine, expectedEndColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, byteEnd)
	_, byteStartColumn := scanner.GetECMALineAndByteOffsetOfPosition(sf, byteStart)

	expectedStart := s.EncodeOffset(byteStart)
	expectedEnd := s.EncodeOffset(byteEnd)

	if name["start"] != expectedStart {
		t.Fatalf("expected UTF-16 start %d, got %v", expectedStart, name["start"])
	}
	if name["end"] != expectedEnd {
		t.Fatalf("expected UTF-16 end %d, got %v", expectedEnd, name["end"])
	}
	if int(expectedStartColumn) == byteStartColumn {
		t.Fatal("expected UTF-16 and byte columns to differ for unicode source")
	}
	if loc["startLine"] != expectedLine {
		t.Fatalf("expected startLine %d, got %v", expectedLine, loc["startLine"])
	}
	if loc["startColumn"] != int(expectedStartColumn) {
		t.Fatalf("expected UTF-16 startColumn %d, got %v", expectedStartColumn, loc["startColumn"])
	}
	if loc["endLine"] != expectedEndLine {
		t.Fatalf("expected endLine %d, got %v", expectedEndLine, loc["endLine"])
	}
	if loc["endColumn"] != int(expectedEndColumn) {
		t.Fatalf("expected UTF-16 endColumn %d, got %v", expectedEndColumn, loc["endColumn"])
	}
}

func TestSerializerCommentOffsetsUseUTF16(t *testing.T) {
	code := `const skull = "💀"; // note`
	sf := parseSource(code)
	s := goast.NewSerializer(sf)
	result := s.SerializeNode(sf.AsNode())

	statements := mustNodeSlice(t, result["Statements"], "expected Statements")
	statement := mustNodeMap(t, statements[0], "expected statement")
	trailing := mustNodeSlice(t, statement["trailingComments"], "expected trailingComments")
	comment := mustNodeMap(t, trailing[0], "expected comment map")
	var rawStart, rawEnd int
	found := false
	for cr := range scanner.GetTrailingCommentRanges(ast.NewNodeFactory(ast.NodeFactoryHooks{}), code, sf.Statements.Nodes[0].End()) {
		rawStart = cr.Pos()
		rawEnd = cr.End()
		found = true
		break
	}
	if !found {
		t.Fatal("expected trailing comment range")
	}
	expectedStart := s.EncodeOffset(rawStart)
	expectedEnd := s.EncodeOffset(rawEnd)
	_, byteStartColumn := scanner.GetECMALineAndByteOffsetOfPosition(sf, rawStart)
	_, utf16StartColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, rawStart)

	if comment["start"] != expectedStart {
		t.Fatalf("expected UTF-16 comment start %d, got %v", expectedStart, comment["start"])
	}
	if comment["end"] != expectedEnd {
		t.Fatalf("expected UTF-16 comment end %d, got %v", expectedEnd, comment["end"])
	}
	if byteStartColumn == int(utf16StartColumn) {
		t.Fatal("expected comment UTF-16 and byte columns to differ for unicode source")
	}
}

func TestSerializerUsesUTF16OffsetsAcrossLines(t *testing.T) {
	code := "const café = 1\nconst skull = \"💀\"\nconst next = 1"
	sf := parseSource(code)
	s := goast.NewSerializer(sf)
	result := s.SerializeNode(sf.AsNode())

	statements := mustNodeSlice(t, result["Statements"], "expected Statements")
	if len(statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(statements))
	}

	thirdStmt := mustNodeMap(t, statements[2], "expected third statement")
	declList := mustNodeMap(t, thirdStmt["DeclarationList"], "expected DeclarationList")
	declarations := mustNodeSlice(t, declList["Declarations"], "expected Declarations")
	decl := mustNodeMap(t, declarations[0], "expected declaration")
	name := mustNodeMap(t, decl["Name"], "expected Name node")
	loc := mustNodeMap(t, name["loc"], "expected Name.loc")
	nameNode := sf.Statements.Nodes[2].
		AsVariableStatement().
		DeclarationList.
		AsVariableDeclarationList().
		Declarations.Nodes[0].
		AsVariableDeclaration().
		Name()

	byteStart := nameNode.Pos()
	byteEnd := nameNode.End()
	expectedLine, expectedStartColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, byteStart)
	expectedEndLine, expectedEndColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, byteEnd)

	if name["start"] != s.EncodeOffset(byteStart) {
		t.Fatalf("expected UTF-16 start %d, got %v", s.EncodeOffset(byteStart), name["start"])
	}
	if name["end"] != s.EncodeOffset(byteEnd) {
		t.Fatalf("expected UTF-16 end %d, got %v", s.EncodeOffset(byteEnd), name["end"])
	}
	if loc["startLine"] != expectedLine {
		t.Fatalf("expected startLine %d, got %v", expectedLine, loc["startLine"])
	}
	if loc["startColumn"] != int(expectedStartColumn) {
		t.Fatalf("expected startColumn %d, got %v", expectedStartColumn, loc["startColumn"])
	}
	if loc["endLine"] != expectedEndLine {
		t.Fatalf("expected endLine %d, got %v", expectedEndLine, loc["endLine"])
	}
	if loc["endColumn"] != int(expectedEndColumn) {
		t.Fatalf("expected endColumn %d, got %v", expectedEndColumn, loc["endColumn"])
	}
}

func TestSerializerBackwardCompat(t *testing.T) {
	sf := parseSource("const x = 1")
	result := goast.SerializeNode(sf.AsNode())
	if result == nil {
		t.Fatal("expected non-nil result from package-level SerializeNode")
	}
	if result["type"] != "SourceFile" {
		t.Errorf("expected type SourceFile, got %v", result["type"])
	}
	if _, hasLoc := result["loc"]; hasLoc {
		t.Error("package-level SerializeNode should not produce loc without sf")
	}
}

func BenchmarkSerializerLongSingleLineUTF16Offsets(b *testing.B) {
	code := buildLongSingleLineSource(4000)
	sf := parseSource(code)

	b.ReportAllocs()
	b.SetBytes(int64(len(code)))
	b.ResetTimer()

	for range b.N {
		serializer := goast.NewSerializer(sf)
		if got := serializer.SerializeNode(sf.AsNode()); got == nil {
			b.Fatal("expected serialized source file")
		}
	}
}
