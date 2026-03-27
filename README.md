# tsgo-ast

`tsgo-ast` is a repository for exposing the Go-based `typescript-go` parser to JavaScript and TypeScript through WebAssembly.

This repository contains two layers:

- the published npm package, built under `npm/`
- the Go + TypeScript source code that produces that package

If you are looking for package installation and usage docs, see `npm/README.md`.
If you are trying to understand or contribute to the repository itself, this file is the right starting point.

## What This Repository Contains

At a high level, the repository turns `typescript-go` parser output into a JS-friendly JSON AST:

1. `src/index.ts` exposes a small JavaScript API
2. `cmd/wasm/main.go` bridges JavaScript and Go in a WASM runtime
3. `goast/` serializes `typescript-go` AST nodes into JSON-friendly objects
4. `scripts/build.sh` and `rolldown.config.ts` produce the final npm package under `npm/`

The result is a package that lets JS/TS consumers call `initGoAst()` once and then parse source text synchronously with `parseAST(code, lang)`.

## Repository Layout

```text
tsgo-ast/
├── cmd/wasm/            # Go WASM entry point
├── goast/               # AST serialization and enrichments
├── src/                 # public JS/TS API source
├── scripts/             # build and release helpers
├── npm/                 # publish root for the npm package
├── tsgolint/            # required git submodule with typescript-go + shims
├── README.md            # repository overview
└── ARCHITECTURE.md      # maintainer-focused design document
```

## How It Works

The runtime path is intentionally narrow:

- `initGoAst()` loads `wasm_exec.js`, fetches `tsgo-ast.wasm`, and starts the Go runtime
- `parseAST(code, lang)` calls the global `goParseAST` function registered by the WASM entry point
- `cmd/wasm/main.go` parses source text with `typescript-go`
- `goast.NewSerializer(sourceFile)` turns the Go AST into a JSON-friendly structure with enrichments such as:
  - UTF-16 offsets
  - line / column locations
  - node flags
  - leading / trailing comments
  - literal and identifier text

For deeper implementation details, see `ARCHITECTURE.md`.

## Development

### Requirements

- Go `1.26`
- Bun
- initialized git submodules

### Setup

```bash
git submodule update --init --recursive
bun install
```

### Common Commands

```bash
bun run build
go test ./...
```

What those commands do:

- `bun run build`
  - builds `npm/tsgo-ast.wasm`
  - copies `wasm_exec.js` into `npm/`
  - bundles `src/index.ts` into `npm/index.js`
  - emits declaration output for the published package
- `go test ./...`
  - runs serializer and Go-side behavior tests

## Publish Model

The published package lives under `npm/`, which is the package root used for npm release artifacts.

Important maintenance rules:

- treat `npm/` as the publish root
- assume most files under `npm/` are generated artifacts
- never hand-edit `npm/wasm_exec.js`
- keep Go result shapes and TypeScript types in sync

For package-facing documentation, keep `npm/README.md` aligned with the public API.

## Release Flow

The project uses a release PR flow instead of publishing directly from arbitrary local state:

1. start from a clean local `main`
2. run `bun run release:pr <version>`
3. merge the generated release PR into `main`
4. let `.github/workflows/ci.yml` pass on `main`
5. let `.github/workflows/release.yml` build, publish, tag, and create the GitHub release

## Where To Read Next

- `npm/README.md` for package installation and usage
- `ARCHITECTURE.md` for the runtime boundary, serialization pipeline, and maintenance invariants
