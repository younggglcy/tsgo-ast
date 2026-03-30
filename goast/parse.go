package goast

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
)

type ParseResult struct {
	OffsetEncoding string         `json:"offsetEncoding"`
	AST            map[string]any `json:"ast"`
	Errors         []string       `json:"errors"`
	SourceFileInfo SourceFileInfo `json:"sourceFileInfo"`
}

type SourceFileInfo struct {
	IsDeclarationFile       bool            `json:"isDeclarationFile"`
	Pragmas                 []string        `json:"pragmas"`
	ReferencedFiles         []FileReference `json:"referencedFiles"`
	TypeReferenceDirectives []FileReference `json:"typeReferenceDirectives"`
}

type FileReference struct {
	FileName string `json:"fileName"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
}

type langConfig struct {
	scriptKind core.ScriptKind
	fileName   string
}

var defaultLangConfig = langConfig{
	scriptKind: core.ScriptKindTS,
	fileName:   "/input.ts",
}

var langConfigs = map[string]langConfig{
	"ts": {
		scriptKind: core.ScriptKindTS,
		fileName:   "/input.ts",
	},
	"tsx": {
		scriptKind: core.ScriptKindTSX,
		fileName:   "/input.tsx",
	},
	"js": {
		scriptKind: core.ScriptKindJS,
		fileName:   "/input.js",
	},
	"jsx": {
		scriptKind: core.ScriptKindJSX,
		fileName:   "/input.jsx",
	},
}

func Parse(code, lang string) ParseResult {
	config, ok := langConfigs[lang]
	if !ok {
		config = defaultLangConfig
	}

	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: config.fileName,
		Path:     tspath.Path(config.fileName),
	}, code, config.scriptKind)

	serializer := NewSerializer(sourceFile)

	return ParseResult{
		OffsetEncoding: "utf-16",
		AST:            serializer.SerializeNode(sourceFile.AsNode()),
		Errors:         diagnosticsToStrings(sourceFile.Diagnostics()),
		SourceFileInfo: buildSourceFileInfo(sourceFile, serializer),
	}
}

func diagnosticsToStrings(diags []*ast.Diagnostic) []string {
	if len(diags) == 0 {
		return nil
	}
	errors := make([]string, 0, len(diags))
	for _, diag := range diags {
		errors = append(errors, diag.String())
	}
	return errors
}

func buildSourceFileInfo(sf *ast.SourceFile, serializer *Serializer) SourceFileInfo {
	return SourceFileInfo{
		IsDeclarationFile:       sf.IsDeclarationFile,
		Pragmas:                 extractPragmas(sf),
		ReferencedFiles:         serializeFileRefs(sf.ReferencedFiles, serializer),
		TypeReferenceDirectives: serializeFileRefs(sf.TypeReferenceDirectives, serializer),
	}
}

func extractPragmas(sf *ast.SourceFile) []string {
	if len(sf.Pragmas) == 0 {
		return nil
	}
	names := make([]string, 0, len(sf.Pragmas))
	for _, pragma := range sf.Pragmas {
		names = append(names, pragma.Name)
	}
	return names
}

func serializeFileRefs(refs []*ast.FileReference, serializer *Serializer) []FileReference {
	if len(refs) == 0 {
		return nil
	}
	result := make([]FileReference, 0, len(refs))
	for _, ref := range refs {
		result = append(result, FileReference{
			FileName: ref.FileName,
			Start:    serializer.EncodeOffset(ref.Pos()),
			End:      serializer.EncodeOffset(ref.End()),
		})
	}
	return result
}
