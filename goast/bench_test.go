package goast_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/younggglcy/tsgo-ast/goast"
)

func buildCommentHeavySource(repeat int) string {
	var builder strings.Builder
	builder.Grow(repeat * 80)
	for i := 0; i < repeat; i++ {
		builder.WriteString("// leading note ")
		builder.WriteString(strconv.Itoa(i))
		builder.WriteByte('\n')
		builder.WriteString("const value")
		builder.WriteString(strconv.Itoa(i))
		builder.WriteString(" = ")
		builder.WriteString(strconv.Itoa(i))
		builder.WriteString("; /* trailing note ")
		builder.WriteString(strconv.Itoa(i))
		builder.WriteString(" */\n")
	}
	return builder.String()
}

func buildTSXFixture(repeat int) string {
	var builder strings.Builder
	builder.Grow(repeat * 120)
	builder.WriteString("type Item = { id: number; label: string }\n")
	builder.WriteString("export function App(props: { items: Item[] }) {\n")
	builder.WriteString("  return <section>\n")
	for i := 0; i < repeat; i++ {
		builder.WriteString("    <article key={props.items[")
		builder.WriteString(strconv.Itoa(i))
		builder.WriteString("]?.id ?? ")
		builder.WriteString(strconv.Itoa(i))
		builder.WriteString("}>{props.items[")
		builder.WriteString(strconv.Itoa(i))
		builder.WriteString("]?.label ?? \"item\"}</article>\n")
	}
	builder.WriteString("  </section>\n")
	builder.WriteString("}\n")
	return builder.String()
}

func parseSourceWithLang(code, lang string) *ast.SourceFile {
	fileName := "/bench.ts"
	scriptKind := core.ScriptKindTS
	switch lang {
	case "tsx":
		fileName = "/bench.tsx"
		scriptKind = core.ScriptKindTSX
	case "jsx":
		fileName = "/bench.jsx"
		scriptKind = core.ScriptKindJSX
	case "js":
		fileName = "/bench.js"
		scriptKind = core.ScriptKindJS
	}

	return parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: fileName,
		Path:     tspath.Path(fileName),
	}, code, scriptKind)
}

func BenchmarkParseAndSerialize(b *testing.B) {
	cases := []struct {
		name string
		lang string
		code string
	}{
		{name: "small-ts", lang: "ts", code: "const value: number = 1\nexport const doubled = value * 2\n"},
		{name: "medium-tsx", lang: "tsx", code: buildTSXFixture(40)},
		{name: "unicode-comments", lang: "ts", code: buildCommentHeavySource(120) + "const café = '🎉'\n"},
		{name: "large-tsx", lang: "tsx", code: buildTSXFixture(220)},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tc.code)))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				result := goast.Parse(tc.code, tc.lang)
				if result.AST == nil {
					b.Fatal("expected AST")
				}
			}
		})
	}
}

func BenchmarkSerializeOnly(b *testing.B) {
	cases := []struct {
		name string
		lang string
		code string
	}{
		{name: "small-ts", lang: "ts", code: "const value: number = 1\nexport const doubled = value * 2\n"},
		{name: "unicode-comments", lang: "ts", code: buildCommentHeavySource(120) + "const café = '🎉'\n"},
		{name: "large-tsx", lang: "tsx", code: buildTSXFixture(220)},
	}

	for _, tc := range cases {
		sourceFile := parseSourceWithLang(tc.code, tc.lang)
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tc.code)))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				serializer := goast.NewSerializer(sourceFile)
				if got := serializer.SerializeNode(sourceFile.AsNode()); got == nil {
					b.Fatal("expected serialized source file")
				}
			}
		})
	}
}
