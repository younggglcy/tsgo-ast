# AGENTS.md

## Repo

`tsgo-ast` exposes `typescript-go` to JS/TS through WebAssembly.
Flow: `src/index.ts` → `cmd/wasm/main.go` → `goast/*` → JSON AST.

## Rules

- `tsgolint/` is a required submodule.

## Edit Map

- JS API / WASM boot: `src/index.ts`
- JS ↔ Go bridge / parse envelope: `cmd/wasm/main.go`
- Enriched serialization: `goast/serializer.go`
- Skip lists / compatibility wrappers: `goast/serialize.go`
- New AST kind dispatch: `goast/concrete.go`
- Serializer tests: `goast/serializer_test.go`
- WASM build: `scripts/build.sh`
- JS bundle output: `rolldown.config.ts`
- Published manifest: `npm/package.json`

## Invariants

- `npm/` is the publish root; most files there are generated.
- Never hand-edit `npm/wasm_exec.js`.
- Keep Go result shape and TS types in sync.
- Use `goast.NewSerializer(sourceFile)` for enriched output.

## Verify

- `go test ./...`
- `bun run build`

## Release

- Start from a clean local `main`.
- Run `bun run release:pr <version>`.
- The command generates `CHANGELOG.md` from commits since the previous `v*` tag, bumps `npm/package.json`, pushes `release/v<version>`, and opens a PR.
- Merging that release PR to `main` triggers `.github/workflows/release.yml`, which builds, publishes to npm, tags `v<version>`, and creates the GitHub Release from `CHANGELOG.md`.
- Local release creation requires `gh` to be installed and authenticated.

Read `ARCHITECTURE.md` for deeper context.
