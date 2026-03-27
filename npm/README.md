# tsgo-ast

`tsgo-ast` exposes the Go-based `typescript-go` parser to JavaScript and TypeScript through WebAssembly.

Initialize the WASM runtime once with `initGoAst()`, then parse source text synchronously with `parseAST(code, lang)`.

## Features

- Built on top of `typescript-go`
- Works in browsers, Node.js, Bun, and Deno
- Returns UTF-16 offsets for JS-friendly position handling
- Includes locations, flags, comments, and literal text in serialized nodes
- Returns diagnostics through `errors` while still providing an AST in many cases

## Installation

```bash
npm install tsgo-ast
```

## Quick Start

```ts
import { initGoAst, parseAST } from "tsgo-ast";

await initGoAst();

const result = parseAST(
  `export const answer: number = 42;`,
  "ts",
);

console.log(result.offsetEncoding);
console.log(result.errors);
console.log(result.ast.type);
console.log(result.ast.Statements);
```

If you want to host the `.wasm` file yourself, pass a custom URL:

```ts
await initGoAst(new URL("./assets/tsgo-ast.wasm", import.meta.url));
```

## API

### `initGoAst(wasmUrl?)`

Initializes the Go WASM runtime.

- Uses lazy singleton initialization
- Loads the bundled `./tsgo-ast.wasm` by default
- Falls back to `arrayBuffer()` when `WebAssembly.instantiateStreaming()` cannot be used

### `parseAST(code, lang = "tsx")`

Synchronously parses source text and returns:

```ts
type ParseResult = {
  offsetEncoding: "utf-16";
  ast: GoAstNode;
  errors: string[] | null;
  sourceFileInfo: {
    isDeclarationFile: boolean;
    pragmas: string[] | null;
    referencedFiles: FileReference[] | null;
    typeReferenceDirectives: FileReference[] | null;
  };
};
```

Constraints and behavior:

- You must finish `await initGoAst()` before calling it
- `lang` supports `"ts" | "tsx" | "js" | "jsx"`
- Position fields use UTF-16 code unit offsets
- Syntax diagnostics are returned in `errors`

### `isInitialized()`

Returns whether runtime initialization has started.

## AST Shape

Every serialized node includes at least these common fields:

```ts
type GoAstNode = {
  type: string;
  Kind?: string;
  start: number;
  end: number;
  loc?: {
    startLine: number;
    startColumn: number;
    endLine: number;
    endColumn: number;
  };
  flags?: string[];
  leadingComments?: AstComment[];
  trailingComments?: AstComment[];
  [key: string]: unknown;
};
```

Depending on the node kind, the serializer may also include:

- reflected exported struct fields
- shared base-node data such as `Name` and `Modifiers`
- `text` for identifier and literal-like nodes
- `SourceFile` metadata such as pragmas and file references

## Runtime Notes

- Initialization is asynchronous, parsing is synchronous
- The package expects `wasm_exec.js` and `tsgo-ast.wasm` to be available from the published package output
- Offsets are normalized to UTF-16 even though `typescript-go` works with byte offsets internally

## Repository

If you want repository-level architecture and contributor docs, see the project root `README.md` and `ARCHITECTURE.md`.
