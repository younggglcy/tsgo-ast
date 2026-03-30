package goast_test

import (
	"strings"
	"testing"

	"github.com/younggglcy/tsgo-ast/goast"
)

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestParseDefaultsUnknownLanguageToTypeScript(t *testing.T) {
	result := goast.Parse("const value: number = 1", "unknown")

	if result.OffsetEncoding != "utf-16" {
		t.Fatalf("expected utf-16 offset encoding, got %q", result.OffsetEncoding)
	}
	if result.AST == nil {
		t.Fatal("expected AST for unknown language fallback")
	}
	if len(result.Errors) != 0 {
		t.Fatalf("expected no parse errors for TypeScript fallback, got %v", result.Errors)
	}
}

func TestParseIncludesSourceFileInfo(t *testing.T) {
	code := strings.Join([]string{
		"/// <reference path=\"./foo.d.ts\" />",
		"/// <reference types=\"node\" />",
		"/* @jsxRuntime automatic */",
		"const element = <div />",
	}, "\n")

	result := goast.Parse(code, "tsx")

	if result.SourceFileInfo.ReferencedFiles == nil || len(result.SourceFileInfo.ReferencedFiles) != 1 {
		t.Fatalf("expected one referenced file, got %#v", result.SourceFileInfo.ReferencedFiles)
	}
	if result.SourceFileInfo.ReferencedFiles[0].FileName != "./foo.d.ts" {
		t.Fatalf("expected reference to ./foo.d.ts, got %#v", result.SourceFileInfo.ReferencedFiles[0])
	}
	if result.SourceFileInfo.TypeReferenceDirectives == nil || len(result.SourceFileInfo.TypeReferenceDirectives) != 1 {
		t.Fatalf("expected one type reference directive, got %#v", result.SourceFileInfo.TypeReferenceDirectives)
	}
	if result.SourceFileInfo.TypeReferenceDirectives[0].FileName != "node" {
		t.Fatalf("expected type reference to node, got %#v", result.SourceFileInfo.TypeReferenceDirectives[0])
	}
	if result.SourceFileInfo.Pragmas == nil {
		t.Fatal("expected pragmas to be collected")
	}
	if !containsString(result.SourceFileInfo.Pragmas, "jsxruntime") {
		t.Fatalf("expected jsxruntime pragma, got %#v", result.SourceFileInfo.Pragmas)
	}
}

func TestParseReturnsDiagnosticsForInvalidSource(t *testing.T) {
	result := goast.Parse("const =", "ts")

	if len(result.Errors) == 0 {
		t.Fatal("expected parse diagnostics for invalid source")
	}
	if result.AST == nil {
		t.Fatal("expected parser to still return an AST")
	}
}
