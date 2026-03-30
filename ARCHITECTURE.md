---
last_updated: 2026-03-29
---

# Architecture

`tsgo-ast` exposes the Go-based `typescript-go` parser to JavaScript and TypeScript through WebAssembly. The public API is intentionally small: callers first `await initGoAst()` to initialize the Go WASM runtime, then use the synchronous `parseAST(code, lang)` API to obtain an enriched JSON AST.

The current implementation is optimized around three goals:

- Keep the JavaScript-facing API thin so Go / WASM details do not leak into consumers
- Preserve as much structure from `typescript-go` as possible while adding locations, comments, and flags that JS tooling expects
- Keep all publishable artifacts under `npm/` so local builds, CI releases, and package contents stay aligned

## High-Level Data Flow

```text
source code string
  │
  ▼
`src/index.ts`
  ├─ `initGoAst(wasmUrl?)`
  │    ├─ dynamically loads `wasm_exec.js`
  │    ├─ creates `new Go()`
  │    ├─ fetches and instantiates `tsgo-ast.wasm`
  │    └─ starts the Go runtime with `go.run(instance)`
  │
  └─ `parseAST(code, lang)`
       └─ calls global `goParseAST(code, lang)`
                │
                ▼
        `cmd/wasm/main.go`
          ├─ extracts JS arguments
          ├─ delegates to `goast.Parse(code, lang)`
          └─ returns `json.Marshal(...)` + `JSON.parse(...)`
                │
                ▼
          `goast/parse.go`
            ├─ maps `lang` to parser config
            ├─ calls `parser.ParseSourceFile(...)`
            ├─ creates `goast.NewSerializer(sourceFile)`
            ├─ runs `serializer.SerializeNode(sourceFile.AsNode())`
            ├─ collects diagnostics
            └─ builds `sourceFileInfo`
                │
                ▼
      `ParseResult { offsetEncoding, ast, errors, sourceFileInfo }`
```

## Directory Structure

```text
tsgo-ast/
├── cmd/wasm/
│   └── main.go              # JS ↔ Go bridge, registers goParseAST
├── goast/
│   ├── concrete.go          # ast.Kind → As*() dispatch
│   ├── parse.go             # parse orchestration and typed result envelope
│   ├── serializer.go        # SourceFile-aware enriched serializer
│   ├── bench_test.go        # Go-side benchmarks
│   └── serializer_test.go   # serializer behavior tests
├── bench/
│   └── parse.bench.mjs      # end-to-end JS benchmark
├── src/
│   ├── index.ts             # public JS/TS API
│   └── wasm_exec.js.d.ts    # type declarations for Go WASM glue
├── scripts/
│   ├── build.sh             # builds wasm and copies wasm_exec.js into npm/
│   ├── release-lib.mjs      # release helper logic
│   └── release-pr.mjs       # release PR generator
├── npm/                     # publish root; final artifacts live here
│   └── package.json
├── tsgolint/                # required submodule that vendors typescript-go and shim packages
├── ARCHITECTURE.md
├── README.md
├── go.mod
├── package.json
└── rolldown.config.ts
```

## Runtime Boundary

### TypeScript side — `src/index.ts`

`src/index.ts` compresses the runtime model into three exports:

- `initGoAst(wasmUrl?)`
  - Uses lazy singleton initialization backed by `initPromise`
  - Dynamically imports `./wasm_exec.js`
  - Prefers `WebAssembly.instantiateStreaming()`
  - Falls back to `arrayBuffer()` + `WebAssembly.instantiate()`
- `parseAST(code, lang)`
  - Performs light argument forwarding only
  - Requires the caller to finish `await initGoAst()` first
  - Delegates actual parsing to the global `goParseAST`
- `isInitialized()`
  - Only reports whether initialization has started, not whether the runtime is fully ready

Two design choices matter here:

1. **Parsing stays synchronous.** Initialization is async, but once the runtime is ready, parsing uses a synchronous API to keep consumer code simple.
2. **Runtime failures do not leak as unhandled rejections.** `go.run(...)` is not awaited; instead it is wrapped in `.catch(...)` because Go `main()` blocks forever in `select {}`.

### Go side — `cmd/wasm/main.go`

`cmd/wasm/main.go` is the single WASM entry point, with a deliberately narrow responsibility set:

1. Extract `code` and `lang` from JS
2. Delegate to `goast.Parse(code, lang)`
3. `json.Marshal` the typed result and convert it back with JS `JSON.parse`

The project uses `json.Marshal` → `JSON.parse` instead of constructing a deep `js.Value` tree manually for two reasons:

- The serialized shape is large and deeply nested, so manual JS bridge code is harder to maintain correctly
- The JSON path is more stable and makes it easier to keep Go and TypeScript result shapes aligned

The whole `parseAST` path is also wrapped with `recover()`, so a panic is downgraded into an `{ errors: [...] }` style result instead of crashing the WASM runtime for the caller.

## Serialization Pipeline

`goast/` is the main domain layer of the repository. It now owns both parse orchestration and serialization, while keeping the WASM entrypoint intentionally thin.

### `goast/parse.go`

`Parse(code, lang)` is the internal pipeline used by the WASM bridge. It is responsible for:

- mapping language strings to `core.ScriptKind` and virtual file names
- calling `parser.ParseSourceFile(...)`
- creating `goast.NewSerializer(sourceFile)`
- serializing the root node
- collecting diagnostics
- building the typed `sourceFileInfo` envelope

### `goast/serializer.go`

`Serializer` is bound to a single `*ast.SourceFile`, which gives it the context needed for richer output and hot-path caching:

- `sf.Text()`
- `scanner.GetECMALineStarts(sf)`
- `ast.ComputePositionMap(text)`
- A `NodeFactory` used for comment extraction
- cached reflection metadata for concrete structs
- memoized leading / trailing comment lookups by position

That context allows it to add more than just common node fields:

- `loc`
- `flags`
- `leadingComments`
- `trailingComments`
- `text` for `Identifier` and literal-like nodes
- additional data from `SourceFile` and shared base-node helpers

One important invariant is that **all exported offsets use UTF-16**. `typescript-go` uses byte offsets internally, while JS/TS tooling typically works with UTF-16 code units, so `Serializer.EncodeOffset()` converts positions into UTF-16 offsets.

### `goast/concrete.go`

`GetConcreteValue()` dispatches on `ast.Kind` and calls the matching `As*()` method so the reflection layer can inspect the correct concrete struct. The dispatch table is long, but the responsibility is intentionally narrow: **make sure each Kind resolves to the right concrete type**.

`KindName()` converts the underlying `Kind` name into the JS-facing `type` field.

The serializer intentionally keeps this dispatch table separate from the parse pipeline because the current JSON shape still depends on reflecting the correct concrete struct for each node kind.

## Result Shape

The stable JS-facing result shape is:

```ts
interface ParseResult {
  offsetEncoding: "utf-16";
  ast: GoAstNode;
  errors: string[] | null;
  sourceFileInfo: {
    isDeclarationFile: boolean;
    pragmas: string[] | null;
    referencedFiles: FileReference[] | null;
    typeReferenceDirectives: FileReference[] | null;
  };
}
```

Compared with the earlier `{ ast, errors }` model, the current contract adds two important constraints:

- `offsetEncoding` is part of the public protocol, not just an implementation detail
- `sourceFileInfo` is a stable export of `SourceFile` metadata, so Go output and TS types must stay in sync

That means any change to the result envelope, exported field names, or TypeScript declarations should be reviewed across all three layers together:

- `src/index.ts`
- `cmd/wasm/main.go`
- `goast/*`

## Build and Packaging

### Local build

The root `package.json` defines three core commands:

```bash
bun run build:wasm
bun run build:js
bun run build
bun run bench
```

The actual flow is:

1. `scripts/build.sh`
   - Copies runtime glue from `$(go env GOROOT)/lib/wasm/wasm_exec.js` to `npm/wasm_exec.js`
   - Builds `npm/tsgo-ast.wasm` with `GOOS=js GOARCH=wasm`
2. `rolldown.config.ts`
   - Bundles `src/index.ts` into `npm/index.js`
   - Uses `isolatedDeclarationPlugin()` to emit declaration output
   - Marks `./wasm_exec.js` as external so the bundler does not inline it

### Benchmarks

The repository tracks performance at two levels:

- `bun run bench`
  - rebuilds the publishable package
  - initializes the WASM runtime once
  - measures steady-state end-to-end `parseAST()` throughput from JS
- `go test -bench . ./goast`
  - measures Go-side parse + serialize and serializer-only hot paths

The publish root is `npm/`, not the repository root. That affects three maintenance rules:

- Build artifacts must land in `npm/`
- `npm/package.json` `files` and `exports` must stay aligned with the build output
- Any manual edit under `npm/` should be treated with suspicion until you confirm whether the file is generated

`npm/wasm_exec.js` in particular must never be edited by hand; it is always copied from the local Go installation.

## Release Flow

The release process is not “push a tag and publish.” It is a two-stage flow:

1. On a clean local `main`, run `bun run release:pr <version>`
   - Generate `CHANGELOG.md`
   - Bump `npm/package.json`
   - Push the `release/v<version>` branch
   - Open the release PR
2. After that release PR is merged into `main`, `.github/workflows/release.yml` is triggered by `push main`
   - The workflow detects whether the merge is actually a release merge
   - If yes, it runs `bun run build`
   - Publishes the npm package
   - Creates the git tag
   - Creates the GitHub Release from `CHANGELOG.md`

So the current responsibility of `release.yml` is **final release execution after the merge lands on `main`**, not a simple tag-driven publish job.

## External Dependency Layout

This repository depends on the `tsgolint/` submodule for its `typescript-go` integration. The `go.mod` file uses `replace` directives to redirect these imports to local submodule paths:

- `github.com/microsoft/typescript-go`
- `github.com/microsoft/typescript-go/shim/ast`
- `github.com/microsoft/typescript-go/shim/core`
- `github.com/microsoft/typescript-go/shim/parser`
- `github.com/microsoft/typescript-go/shim/tspath`
- `github.com/microsoft/typescript-go/shim/scanner`

That implies two practical constraints:

- If submodules are not initialized, Go builds and tests will fail
- When debugging parser behavior, prefer reading the local `tsgolint/` checkout before assuming remote docs match the pinned source

## Maintenance Invariants

These are the easiest invariants to break over time:

1. **Go result shapes and TypeScript types must stay synchronized.**
2. **`npm/` is the publish root, but most files there are generated artifacts.**
3. **`npm/wasm_exec.js` may only be replaced by the build script.**
4. **Enriched output must be produced through `goast.NewSerializer(sourceFile)`.**
5. **All exported positions use UTF-16 encoding.**

If the protocol keeps evolving, this is the safest review order:

1. Does `src/index.ts` need new or updated TypeScript types?
2. Does `cmd/wasm/main.go` need a changed result envelope?
3. Does `goast/serializer.go` already have the context required to compute the new field?
4. Does `goast/serializer_test.go` need new assertions?

Following that order helps avoid a common maintenance failure mode: Go returns a new field, but TypeScript types and documentation still describe the old contract.
