//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/younggglcy/tsgo-ast/goast"
)

func makeErrorResult(msg string) js.Value {
	result := map[string]any{
		"offsetEncoding": "utf-16",
		"ast":            nil,
		"errors":         []string{msg},
		"sourceFileInfo": nil,
	}
	jsonBytes, _ := json.Marshal(result)
	return js.Global().Get("JSON").Call("parse", string(jsonBytes))
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
	result := goast.Parse(code, lang)

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
