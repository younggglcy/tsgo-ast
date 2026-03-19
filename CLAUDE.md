---
based_on: fa82667c40cf68131a6becbd02038a2957b44a80  # main: Initial commit
last_updated: 2026-03-19
---

# CLAUDE.md

## Project Overview

`tsgo-ast` — Exposes Microsoft's typescript-go (TypeScript 7) parser to JavaScript via WebAssembly. Ships as an npm package (`tsgo-ast`) containing a WASM binary + thin TypeScript API.

## Language

All commit messages, code comments, and documentation in **English**.

## Package Manager

Use **bun** (not npm or pnpm).

## Tech Stack

- **Go 1.26** — WASM compilation target (`GOOS=js GOARCH=wasm`)
- **TypeScript 5.9** — API wrapper + type declarations
- **Rolldown** — ESM bundler
- **Git submodules** — `tsgolint` submodule provides typescript-go

## Build Commands

```bash
bun run build          # full build: WASM + JS
bun run build:wasm     # Go → WASM only (scripts/build.sh)
bun run build:js       # TypeScript → JS + .d.ts (rolldown + tsc)
```

## Project Structure

```
cmd/wasm/main.go       # WASM entry point — registers goParseAST() on globalThis
goast/serialize.go     # Recursive AST node → map[string]any serialization
goast/concrete.go      # ast.Kind → As*() dispatch (400+ node types)
src/index.ts           # Public TypeScript API: initGoAst, parseAST, isInitialized
npm/                   # Published package output (build artifacts, do NOT edit manually)
scripts/build.sh       # WASM build script
tsgolint/              # Git submodule (oxc-project/tsgolint → typescript-go)
```

## Key Architecture Decisions

- **`npm/` is the publish root** — all build artifacts land here; `npm/package.json` is the published manifest
- **Root `package.json` is dev-only** (`"private": true`) — build scripts and devDependencies only
- **`wasm_exec.js` is copied from GOROOT** at build time — never edit it manually
- **Go `select{}` keeps WASM alive** — `go.run()` is intentionally not awaited on the JS side
- **Serialization uses `reflect`** — struct fields are extracted dynamically, filtered by base type and skip lists
- **JSON round-trip for interop** — Go marshals to JSON string, JS side `JSON.parse`s it (more reliable than `js.ValueOf` for nested structures)

## Conventions

### Go Code

- Build tag: `//go:build js && wasm` for WASM entry point
- New AST node types: add case to `GetConcreteValue()` in `goast/concrete.go`
- Fields to hide from output: add to `baseTypeNames` or `skipFieldNames` in `goast/serialize.go`

### TypeScript Code

- ESM only (`"type": "module"`)
- Strict mode enabled
- `wasm_exec.js` is external in Rolldown config — it's loaded via dynamic `import()` at runtime

### Git

- `tsgolint/` is a submodule — clone with `--recurse-submodules` or run `git submodule update --init --recursive`
- Build artifacts in `npm/` are gitignored (except `npm/package.json`)

## Release Process

1. Update version in `npm/package.json`
2. Push a `v*` tag (e.g., `git tag v0.2.0 && git push --tags`)
3. GitHub Actions builds and publishes to npm automatically

## Common Tasks

### Add support for a new AST node kind

1. Add the `case ast.KindXxx:` → `return node.AsXxx()` in `goast/concrete.go`
2. If the node has fields that should be hidden, update `skipFieldNames` in `goast/serialize.go`
3. Rebuild: `bun run build`

### Update typescript-go version

1. `cd tsgolint && git pull origin main && cd ..`
2. `go mod tidy`
3. Rebuild and verify: `bun run build`
