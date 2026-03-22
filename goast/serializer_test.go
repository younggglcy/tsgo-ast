package goast_test

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/younggglcy/tsgo-ast/goast"
)

func parseSource(code string) *ast.SourceFile {
	return parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.tsx",
		Path:     tspath.Path("/test.tsx"),
	}, code, core.ScriptKindTSX)
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
	varStmt, ok := stmts[0].(map[string]any)
	if !ok {
		t.Fatal("expected map for first statement")
	}
	declList, ok := varStmt["DeclarationList"].(map[string]any)
	if !ok {
		t.Fatal("expected DeclarationList")
	}
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
	varStmt, ok := stmts[0].(map[string]any)
	if !ok {
		t.Fatal("expected map for first statement")
	}
	leading, ok := varStmt["leadingComments"].([]any)
	if !ok || len(leading) == 0 {
		t.Fatal("expected leadingComments on statement after comment")
	}
	comment, ok := leading[0].(map[string]any)
	if !ok {
		t.Fatal("expected comment map")
	}
	if comment["type"] != "line" {
		t.Errorf("expected comment type 'line', got %v", comment["type"])
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
