# Architecture Simplification Design

## Context

`tsgo-ast` exposes the Go-based `typescript-go` parser to JS/TS through WebAssembly. The current runtime path is intentionally small on the outside, but the internal implementation has accumulated architectural weight in three places:

- the WASM entrypoint in `cmd/wasm/main.go` owns both JS bridge concerns and parse/result orchestration
- the serializer logic is split across multiple files with overlapping responsibilities
- the hot path does repeated per-node work for reflection, comments, and position encoding

The goal of this design is to reduce code size and improve steady-state `parseAST()` performance without making the public output leaner.

## Goals

- Keep `parseAST(code, lang)` as the primary public API
- Keep the current rich output shape as the default behavior
- Reduce internal architectural sprawl and file-level indirection
- Improve end-to-end `parseAST()` performance after initialization
- Add benchmarks that track both user-visible and Go-side performance

## Non-Goals

- No attempt to make the AST output smaller by default
- No major redesign of the JS public API
- No migration to an upstream binary AST protocol for normal parsing
- No unrelated repo or release-process cleanup beyond what directly supports the architecture change

## Current Problems

### Split responsibilities across the runtime boundary

`cmd/wasm/main.go` currently does more than WASM registration and JS bridging. It also owns language mapping, parse orchestration, diagnostics collection, and result-envelope assembly. That makes the runtime boundary harder to reason about and harder to benchmark in isolation.

### Serializer logic is fragmented

The serializer is functionally centered in `goast/serializer.go`, but behavior is spread across:

- `goast/serialize.go`
- `goast/comments.go`
- `goast/flags.go`
- `goast/concrete.go`

This split makes the path from `SourceFile` to JSON AST harder to follow than it needs to be. Some of the split is useful, but some is only historical layering.

### Hot-path repeated work

The current serializer performs repeated work that is avoidable:

- it computes its own position map instead of reusing the `SourceFile` cache
- it discovers reflection field metadata repeatedly per node
- it asks the scanner for leading and trailing comments repeatedly during traversal
- it duplicates language-to-parser metadata logic in more than one switch

### Some complexity is unavoidable

The rich output shape preserves named properties such as `Statements`, `Initializer`, `Parameters`, and `Body`. Because `typescript-go`'s `ForEachChild` traversal does not expose property names, the current output contract still requires access to concrete node structs. This means the custom `Kind` to `As*()` dispatch cannot be removed entirely without changing the output format.

## Proposed Architecture

### Target module split

The runtime should be reduced to three layers:

1. `src/index.ts`
   - owns JS API, initialization, and runtime boot
2. `cmd/wasm/main.go`
   - owns only WASM registration, JS argument extraction, panic recovery, and final JS bridge conversion
3. `goast/`
   - owns parse orchestration, result assembly, and AST serialization

### WASM boundary rule

`cmd/wasm/main.go` should become a thin adapter. It should not own parser configuration tables, source-file metadata assembly, or serializer setup beyond delegating to a single `goast.Parse(...)` style entrypoint.

### Go pipeline rule

`goast/` should expose one internal pipeline that turns source text plus language into a typed parse result:

- map language to script kind and virtual file name
- parse source text
- serialize the root node
- collect diagnostics
- build `sourceFileInfo`
- return a typed result object ready for JSON marshaling

This consolidates parse-related decisions into one place and removes orchestration logic from the WASM shim.

### Serializer rule

`Serializer` should become the single owner of:

- position encoding
- line and column location computation
- node flag decoding
- comment extraction
- reflection-based field serialization
- shared node enrichment

To support this, fold the helper behavior from `goast/serialize.go`, `goast/comments.go`, and `goast/flags.go` into `goast/serializer.go`. Keep `goast/concrete.go` as the isolated concrete-node dispatch layer because the output contract still depends on it.

## Proposed File-Level Changes

### `cmd/wasm/main.go`

Reduce this file to:

- `goParseAST` registration
- JS argument extraction
- `recover()` handling
- conversion of a Go parse result to `JSON.parse(...)`

Move out:

- `langToScriptKind`
- `langToFileName`
- source file info assembly
- diagnostics/result-envelope building

### `goast/parse.go` or `goast/api.go`

Add a focused orchestration file responsible for:

- language mapping table
- source-file parsing
- top-level typed `ParseResult` definition
- `SourceFileInfo` assembly
- diagnostics extraction

This becomes the internal API consumed by the WASM entrypoint.

### `goast/serializer.go`

Expand this file into the single serializer implementation.

Responsibilities:

- serializer context setup from `*ast.SourceFile`
- cached offset conversion helpers
- cached comment lookup
- cached reflection field metadata
- field serialization
- shared node enrichment

### `goast/concrete.go`

Keep this file, but treat it as a narrow compatibility layer:

- `KindName`
- `GetConcreteValue`

It remains necessary because `ForEachChild` cannot recreate the current named-property output by itself.

### Delete or fold helper files

Delete these files by folding their behavior into `goast/serializer.go`:

- `goast/serialize.go`
- `goast/comments.go`
- `goast/flags.go`

This is a code-size reduction and a readability improvement. It keeps the serializer in one place instead of forcing maintainers to chase small helpers across files.

## Performance Changes

### Reuse `SourceFile` position map

Use `sourceFile.GetPositionMap()` rather than calling `ast.ComputePositionMap(s.text)` in the serializer. `typescript-go` already computes and caches this lazily.

Expected effect:

- less repeated setup work
- lower allocation pressure for unicode-heavy files

### Cache reflection metadata per concrete type

Precompute the exported serializable field list for each reflected struct type once and reuse it across nodes.

Expected effect:

- lower per-node reflection overhead
- smaller constant factors in large AST traversals

### Cache comment ranges by position

Memoize comment extraction results using the queried byte position as the cache key.

Expected effect:

- avoid repeated scanner traversals for identical positions
- lower cost on comment-heavy sources

### Collapse language mapping into one lookup table

Replace duplicate switches with one static mapping that yields both:

- `core.ScriptKind`
- virtual file name

Expected effect:

- slightly smaller code
- simpler control flow

### Use typed top-level result structs

Build the parse envelope from typed Go structs and leave dynamic maps only for AST nodes.

Expected effect:

- less ad hoc map construction
- clearer contract between orchestration and serializer

## Benchmarks

Benchmarks are part of the architectural work and should land in the same effort.

### Primary benchmark

Add an end-to-end JS benchmark for steady-state `parseAST()` after `await initGoAst()`.

Suggested file:

- `bench/parse.bench.mjs`

Suggested scenarios:

- small ASCII TypeScript input
- medium TSX input with JSX and type annotations
- unicode plus comments
- larger representative fixture

This benchmark is the main success metric because it tracks what consumers actually pay for.

### Secondary benchmarks

Add Go benchmarks to explain where wins or regressions come from.

Suggested file:

- `goast/bench_test.go`

Suggested cases:

- parse + serialize
- serialize only from an already parsed `SourceFile`
- unicode-heavy source
- comment-heavy source

### Commands

Add benchmark commands at the root:

- `bun run bench`
- `go test -bench . ./goast`

The JS benchmark should be positioned as authoritative for performance claims. The Go benchmarks should be diagnostic support.

## Migration Plan

1. Add end-to-end and Go-side benchmarks before structural changes so the current baseline can be measured.
2. Introduce the new `goast.Parse(...)` orchestration path while keeping behavior unchanged.
3. Simplify `cmd/wasm/main.go` to call the new pipeline.
4. Fold serializer helpers into one implementation file.
5. Add serializer caches for position map reuse, reflection metadata, and comments.
6. Remove obsolete helper files and compatibility wrappers.
7. Re-run tests and benchmarks to compare before and after.

## Risks And Constraints

### Output compatibility risk

Because the output shape must stay rich, serializer refactoring must be validated against the existing tests and new benchmarks. Any simplification that drops named fields or enrichments is out of scope.

### Dispatch-table maintenance risk

`goast/concrete.go` remains large. This design accepts that cost because replacing it would either fail to preserve the current shape or introduce a more complex upstream coupling than the current repo needs.

### Benchmark noise risk

WASM startup and one-time initialization should be excluded from the main benchmark. The benchmark should measure steady-state `parseAST()` after initialization so it reflects the hot path rather than the cold-start path.

## Testing And Verification

Required verification for implementation:

- `go test ./...`
- `bun run build`
- benchmark runs before and after the refactor

Success criteria:

- fewer serializer-related implementation files
- a thinner `cmd/wasm/main.go`
- no intentional public output reductions
- measurable improvement or at least no regression in end-to-end `parseAST()` benchmarks
