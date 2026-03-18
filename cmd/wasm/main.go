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

func parseAST(_ js.Value, args []js.Value) any {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("goParseAST panic:", r)
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

	astMap := goast.SerializeNode(sourceFile.AsNode())

	var errors []string
	for _, diag := range sourceFile.Diagnostics() {
		errors = append(errors, diag.String())
	}

	result := map[string]any{
		"ast":    astMap,
		"errors": errors,
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return js.ValueOf(fmt.Sprintf(`{"error":%q}`, err.Error()))
	}

	// Use JSON.parse on the JS side for reliable conversion
	return js.Global().Get("JSON").Call("parse", string(jsonBytes))
}

func main() {
	js.Global().Set("goParseAST", js.FuncOf(parseAST))
	select {}
}
