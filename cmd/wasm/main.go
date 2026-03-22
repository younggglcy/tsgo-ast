//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/younggglcy/tsgo-ast/goast"
)

func langToScriptKind(lang string) core.ScriptKind {
	switch lang {
	case "tsx":
		return core.ScriptKindTSX
	case "ts":
		return core.ScriptKindTS
	case "jsx":
		return core.ScriptKindJSX
	case "js":
		return core.ScriptKindJS
	default:
		return core.ScriptKindTS
	}
}

func langToFileName(lang string) string {
	switch lang {
	case "tsx":
		return "/input.tsx"
	case "ts":
		return "/input.ts"
	case "jsx":
		return "/input.jsx"
	case "js":
		return "/input.js"
	default:
		return "/input.ts"
	}
}

func makeErrorResult(msg string) js.Value {
	result := map[string]any{
		"ast":            nil,
		"errors":         []string{msg},
		"sourceFileInfo": nil,
	}
	jsonBytes, _ := json.Marshal(result)
	return js.Global().Get("JSON").Call("parse", string(jsonBytes))
}

func serializeFileRefs(refs []*ast.FileReference) []any {
	if len(refs) == 0 {
		return nil
	}
	result := make([]any, 0, len(refs))
	for _, ref := range refs {
		result = append(result, map[string]any{
			"fileName": ref.FileName,
			"start":    ref.Pos(),
			"end":      ref.End(),
		})
	}
	return result
}

func extractPragmas(sf *ast.SourceFile) []string {
	if len(sf.Pragmas) == 0 {
		return nil
	}
	names := make([]string, 0, len(sf.Pragmas))
	for _, p := range sf.Pragmas {
		names = append(names, p.Name)
	}
	return names
}

func buildSourceFileInfo(sf *ast.SourceFile) map[string]any {
	return map[string]any{
		"isDeclarationFile":       sf.IsDeclarationFile,
		"pragmas":                 extractPragmas(sf),
		"referencedFiles":         serializeFileRefs(sf.ReferencedFiles),
		"typeReferenceDirectives": serializeFileRefs(sf.TypeReferenceDirectives),
	}
}

func parseAST(_ js.Value, args []js.Value) (ret any) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("goParseAST panic:", r)
			ret = makeErrorResult(fmt.Sprintf("panic: %v", r))
		}
	}()

	code := args[0].String()
	lang := args[1].String()

	scriptKind := langToScriptKind(lang)
	fileName := langToFileName(lang)

	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: fileName,
		Path:     tspath.Path(fileName),
	}, code, scriptKind)

	// Use enriched Serializer
	serializer := goast.NewSerializer(sourceFile)
	astMap := serializer.SerializeNode(sourceFile.AsNode())

	var errors []string
	for _, diag := range sourceFile.Diagnostics() {
		errors = append(errors, diag.String())
	}

	result := map[string]any{
		"ast":            astMap,
		"errors":         errors,
		"sourceFileInfo": buildSourceFileInfo(sourceFile),
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return makeErrorResult(fmt.Sprintf("json.Marshal failed: %s", err.Error()))
	}

	return js.Global().Get("JSON").Call("parse", string(jsonBytes))
}

func main() {
	js.Global().Set("goParseAST", js.FuncOf(parseAST))
	select {}
}
