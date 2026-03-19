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

This project uses [@changesets/cli](https://github.com/changesets/changesets) for versioning and changelog generation, with OIDC token-less publishing to npm.

### Developer workflow

1. Make changes and add a changeset: `bunx changeset`
2. Commit the changeset file with your code
3. Push and merge PR to `main`

### What happens on CI

1. `changesets/action` detects pending changesets and opens a "Version Packages" PR
2. The PR bumps `npm/package.json` version and updates `CHANGELOG.md`
3. When the Version PR is merged, CI automatically:
   - Builds WASM + JS (`bun run build`)
   - Publishes to npm with OIDC provenance (`npm publish --provenance --access public`)
   - Creates a GitHub Release with the changelog as body

### Manual steps (one-time setup, already done)

- npm trusted publishing configured for `younggglcy/tsgo-ast` + `release.yml`
- No `NPM_TOKEN` secret needed — OIDC handles authentication

## Common Tasks

### Add support for a new AST node kind

1. Add the `case ast.KindXxx:` → `return node.AsXxx()` in `goast/concrete.go`
2. If the node has fields that should be hidden, update `skipFieldNames` in `goast/serialize.go`
3. Rebuild: `bun run build`

### Update typescript-go version

1. `cd tsgolint && git pull origin main && cd ..`
2. `go mod tidy`
3. Rebuild and verify: `bun run build`
