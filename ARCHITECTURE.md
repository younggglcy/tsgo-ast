---
based_on: fa82667c40cf68131a6becbd02038a2957b44a80  # main: Initial commit
last_updated: 2026-03-19
---

# Architecture

`tsgo-ast` exposes Microsoft's [typescript-go](https://github.com/nicolo-ribaudo/tc39-proposal-structs) (TypeScript 7, Project Corsa) parser to JavaScript/TypeScript via WebAssembly. Consumers call `parseAST(code, lang)` and receive a JSON tree whose node types mirror the Go-side `ast.Kind` names.

## High-Level Data Flow

```
┌──────────────────────────────────────────────────────────────────┐
│  JavaScript Runtime (Browser / Node.js / Bun / Deno)             │
│                                                                  │
│  import { initGoAst, parseAST } from "tsgo-ast"                 │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  npm package (tsgo-ast)                                    │  │
│  │                                                            │  │
│  │  index.js          ← TypeScript API wrapper (ESM)          │  │
│  │  wasm_exec.js      ← Go WASM runtime glue (from GOROOT)   │  │
│  │  tsgo-ast.wasm     ← compiled Go → WASM binary            │  │
│  └────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
                              │
                    WebAssembly boundary
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│  WASM Module (Go)                                                │
│                                                                  │
│  cmd/wasm/main.go                                                │
│    ├─ registers global `goParseAST(code, lang)` via syscall/js   │
│    ├─ maps lang → ScriptKind / fileName                          │
│    └─ blocks forever with `select{}`                             │
│                                                                  │
│  goast/serialize.go                                              │
│    ├─ SerializeNode()  → recursive AST node → map[string]any    │
│    ├─ reflect-based struct field extraction                      │
│    └─ filters base types & internal metadata                     │
│                                                                  │
│  goast/concrete.go                                               │
│    ├─ GetConcreteValue()  → Kind → As*() dispatch (400+ cases)  │
│    └─ KindName()  → strips "Kind" prefix for display            │
│                                                                  │
│  typescript-go (via tsgolint submodule)                          │
│    ├─ parser.ParseSourceFile()                                   │
│    ├─ ast.Node / ast.Kind / As*() methods                        │
│    └─ shim packages: ast, core, parser, tspath                   │
└──────────────────────────────────────────────────────────────────┘
```

## Directory Structure

```
tsgo-ast/
├── cmd/wasm/              # WASM entry point (Go)
│   └── main.go            #   registers goParseAST, handles JS ↔ Go interop
├── goast/                 # AST serialization layer (Go)
│   ├── concrete.go        #   Kind → concrete struct dispatch
│   └── serialize.go       #   recursive node → JSON-friendly map
├── src/                   # TypeScript API source
│   ├── index.ts           #   public API: initGoAst, parseAST, isInitialized
│   └── wasm_exec.js.d.ts  #   type declarations for Go WASM glue
├── npm/                   # published package output (build artifacts)
│   └── package.json       #   npm metadata (name, version, exports)
├── scripts/
│   └── build.sh           #   Go → WASM compilation script
├── tsgolint/              # git submodule → oxc-project/tsgolint
│   └── (contains typescript-go + shim packages)
├── go.mod                 # Go module with `replace` directives to submodule
├── package.json           # dev-only: build scripts, devDependencies
├── rolldown.config.ts     # Rolldown bundler config (src → npm)
├── tsconfig.json          # TypeScript config (declaration emit)
└── .github/workflows/
    └── release.yml        # tag-triggered CI: build + npm publish
```

## Key Components

### 1. WASM Entry Point — `cmd/wasm/main.go`

Build-tagged `//go:build js && wasm`. Registers a single global function `goParseAST(code, lang)` on the JS `globalThis`. The function:

1. Converts `lang` string (`ts`/`tsx`/`js`/`jsx`) to `core.ScriptKind` and a virtual filename
2. Calls `parser.ParseSourceFile()` from typescript-go
3. Serializes the resulting AST via `goast.SerializeNode()`
4. Collects diagnostics from `sourceFile.Diagnostics()`
5. Returns `{ ast, errors }` as a JS object (via `JSON.parse` on the JS side)

Panic recovery wraps the entire parse path to prevent WASM crashes from surfacing as unhandled errors.

### 2. Serialization Layer — `goast/`

**`serialize.go`** — Recursive conversion of `*ast.Node` → `map[string]any`:

- Each node gets `{ type, start, end }` plus all exported struct fields
- Uses Go `reflect` to walk struct fields dynamically
- Filters out:
  - Base type embeddings (`NodeBase`, `StatementBase`, `ExpressionBase`, etc.)
  - Internal metadata fields (`EndOfFileToken`, `NodeCount`, `ScriptKind`, etc.)
- Special handling for `Name()`, `Modifiers()`, identifier text, and literal text
- Handles `*ast.Node`, `*ast.NodeList`, `*ast.ModifierList`, and `[]*ast.Node` field types

**`concrete.go`** — Type dispatch via `GetConcreteValue()`:

- Maps `ast.Kind` → the appropriate `As*()` method (400+ cases)
- Covers all TypeScript AST node categories: identifiers, literals, expressions, statements, declarations, type nodes, JSX, JSDoc, clauses, class/type members, binding patterns
- `KindName()` strips the `Kind` prefix for human-readable type names

### 3. TypeScript API — `src/index.ts`

Public surface:

| Export | Description |
|--------|-------------|
| `initGoAst(wasmUrl?)` | Load WASM runtime (lazy singleton, safe to call concurrently) |
| `parseAST(code, lang)` | Synchronous parse after init; returns `{ ast: GoAstNode, errors: string[] }` |
| `isInitialized()` | Check if WASM runtime is ready |

Initialization loads `wasm_exec.js` (Go's JS glue), fetches the `.wasm` binary, instantiates it with streaming (fallback to `arrayBuffer` for `file://` protocol), and starts the Go runtime without awaiting it (since `main()` blocks forever).

### 4. typescript-go Dependency

The actual parser comes from Microsoft's typescript-go, accessed through the `tsgolint` git submodule (from `oxc-project/tsgolint`). The `go.mod` uses `replace` directives to point all `github.com/microsoft/typescript-go` imports to the local submodule paths:

- `shim/ast` — AST node types, `Kind` enum, `As*()` methods
- `shim/core` — `ScriptKind`, diagnostics
- `shim/parser` — `ParseSourceFile()`
- `shim/tspath` — path handling utilities

## Build Pipeline

### Local Build

```bash
bun run build          # full build (wasm + js)
bun run build:wasm     # Go → WASM only
bun run build:js       # TypeScript → JS + declarations only
```

**WASM build** (`scripts/build.sh`):
1. Copies `wasm_exec.js` from `$GOROOT/lib/wasm/` into `npm/`
2. Compiles with `GOOS=js GOARCH=wasm go build -o npm/tsgo-ast.wasm ./cmd/wasm`

**JS build** (`bun run build:js`):
1. Rolldown bundles `src/index.ts` → `npm/index.js` (ESM, `wasm_exec.js` marked external)
2. `tsc --emitDeclarationOnly` generates `npm/index.d.ts`

### CI/CD — `.github/workflows/release.yml`

Triggered by `v*` tags. Steps:
1. Checkout with recursive submodules
2. Setup Go 1.26, Node.js 22, Bun
3. `bun install` → `bash scripts/build.sh` → `bun run build:js`
4. `cd npm && npm publish` (authenticated via `NPM_TOKEN` secret)

## npm Package

Published as `tsgo-ast` on npm. The `npm/` directory is the package root:

```
npm/
├── package.json       # name, version, exports, sideEffects
├── index.js           # bundled ESM entry
├── index.d.ts         # TypeScript declarations
├── wasm_exec.js       # Go WASM runtime (marked as sideEffect)
└── tsgo-ast.wasm      # compiled WASM binary
```

- ESM-only (`"type": "module"`)
- Single export entry with types
- `wasm_exec.js` declared as `sideEffects` to prevent tree-shaking (it mutates `globalThis`)

## Tech Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Parser | [typescript-go](https://github.com/nicolo-ribaudo/tc39-proposal-structs) (Go 1.26) | TypeScript/JavaScript AST parsing |
| WASM interop | `syscall/js` | Go ↔ JavaScript bridge |
| Serialization | `encoding/json` + `reflect` | AST → JSON conversion |
| Bundler | [Rolldown](https://rolldown.rs) | TypeScript → ESM bundle |
| Type generation | TypeScript 5.9 | `.d.ts` declaration emit |
| Package manager | [Bun](https://bun.sh) | Dependency management + script runner |
| CI/CD | GitHub Actions | Tag-triggered npm publish |
| Dependency | Git submodules | typescript-go via tsgolint |
