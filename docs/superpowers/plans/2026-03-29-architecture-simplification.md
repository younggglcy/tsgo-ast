# Architecture Simplification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Simplify the Go/WASM architecture, keep the current AST output contract, and add benchmarks with end-to-end `parseAST()` as the primary performance metric.

**Architecture:** Move parse orchestration and result assembly into `goast`, reduce `cmd/wasm/main.go` to a thin JS bridge, and consolidate serializer behavior into a single implementation with caches for position maps, reflection metadata, and comment lookups. Preserve the rich output shape and validate the refactor with Go tests, Go microbenchmarks, and a JS end-to-end benchmark.

**Tech Stack:** Go 1.26, Bun, WebAssembly, rolldown, TypeScript, `typescript-go`

---

### File Map

**Create:**
- `goast/parse.go` - typed parse pipeline, language mapping, diagnostics extraction, source-file info assembly
- `goast/parse_test.go` - parse pipeline regression tests
- `goast/bench_test.go` - Go benchmarks for parse+serialize and serializer hot paths
- `bench/parse.bench.mjs` - end-to-end JS benchmark after `initGoAst()`

**Modify:**
- `cmd/wasm/main.go` - thin WASM adapter only
- `goast/serializer.go` - consolidated serializer implementation and caches
- `goast/serializer_test.go` - regression coverage for refactored serializer behavior
- `package.json` - benchmark command
- `README.md` - benchmark usage and updated architecture summary
- `ARCHITECTURE.md` - updated module responsibilities

**Delete:**
- `goast/serialize.go` - fold into `goast/serializer.go`
- `goast/comments.go` - fold into `goast/serializer.go`
- `goast/flags.go` - fold into `goast/serializer.go`

### Task 1: Lock Behavior With Tests

**Files:**
- Create: `goast/parse_test.go`
- Modify: `goast/serializer_test.go`

- [ ] **Step 1: Write failing parse pipeline tests**

Write tests for:
- default language fallback behavior
- `offsetEncoding` always being `"utf-16"`
- `sourceFileInfo` fields for pragmas and file references
- parse errors surfacing as strings without crashing

- [ ] **Step 2: Run the targeted Go tests to verify they fail**

Run: `go test ./goast -run 'TestParse|TestSerializer'`
Expected: FAIL because `goast.Parse(...)` and related typed result helpers do not exist yet

- [ ] **Step 3: Extend serializer regression coverage for current enrichments**

Add or adjust tests covering:
- UTF-16 offsets
- locations
- flags
- leading and trailing comments
- shared base fields like `Parameters`, `Body`, `Members`

- [ ] **Step 4: Re-run the targeted tests**

Run: `go test ./goast -run 'TestParse|TestSerializer'`
Expected: still FAIL only for unimplemented pipeline changes, with serializer expectations compiling cleanly

- [ ] **Step 5: Commit**

```bash
git add goast/parse_test.go goast/serializer_test.go
git commit -m "test: lock parse pipeline behavior"
```

### Task 2: Extract Typed Parse Pipeline

**Files:**
- Create: `goast/parse.go`
- Modify: `cmd/wasm/main.go`
- Test: `goast/parse_test.go`

- [ ] **Step 1: Implement typed parse result structs and language mapping**

Define:
- `ParseResult`
- `SourceFileInfo`
- `FileReference`
- one shared `lang` lookup table returning script kind and virtual filename

- [ ] **Step 2: Implement `goast.Parse(code, lang)`**

Include:
- source parsing
- serializer construction
- diagnostics extraction
- source file info assembly
- default language fallback behavior matching today

- [ ] **Step 3: Thin `cmd/wasm/main.go` down to the bridge**

Keep:
- JS argument extraction
- panic recovery
- JSON marshaling and `JSON.parse(...)`

Move out:
- language mapping
- result assembly
- source file info assembly

- [ ] **Step 4: Run focused Go tests**

Run: `go test ./goast -run 'TestParse|TestSerializer'`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add goast/parse.go cmd/wasm/main.go goast/parse_test.go
git commit -m "refactor: extract go parse pipeline"
```

### Task 3: Consolidate The Serializer

**Files:**
- Modify: `goast/serializer.go`
- Delete: `goast/serialize.go`
- Delete: `goast/comments.go`
- Delete: `goast/flags.go`
- Test: `goast/serializer_test.go`

- [ ] **Step 1: Fold helper behavior into `goast/serializer.go`**

Move in:
- wrapper behavior from `serialize.go`
- comment extraction helpers
- node flag decoding helpers

- [ ] **Step 2: Keep public serializer entrypoints stable**

Preserve:
- `NewSerializer`
- `(*Serializer).SerializeNode`
- `(*Serializer).SerializeNodeSlice`
- package-level `SerializeNode` and `SerializeNodeSlice` wrappers if still needed by tests or callers

- [ ] **Step 3: Delete obsolete helper files**

Remove:
- `goast/serialize.go`
- `goast/comments.go`
- `goast/flags.go`

- [ ] **Step 4: Run serializer tests**

Run: `go test ./goast -run TestSerializer`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add goast/serializer.go goast/serializer_test.go goast/parse.go cmd/wasm/main.go
git rm goast/serialize.go goast/comments.go goast/flags.go
git commit -m "refactor: consolidate serializer implementation"
```

### Task 4: Add Hot-Path Caches

**Files:**
- Modify: `goast/serializer.go`
- Test: `goast/serializer_test.go`

- [ ] **Step 1: Reuse `SourceFile.GetPositionMap()`**

Change serializer setup to use the source file's cached position map instead of recomputing it.

- [ ] **Step 2: Cache reflection metadata per concrete type**

Add a package-level cache keyed by `reflect.Type` that stores the exported fields that should be serialized.

- [ ] **Step 3: Cache comment lookups by position**

Memoize leading and trailing comment extraction using byte position keys.

- [ ] **Step 4: Re-run serializer and parse tests**

Run: `go test ./goast -run 'TestParse|TestSerializer'`
Expected: PASS with no output-shape regressions

- [ ] **Step 5: Commit**

```bash
git add goast/serializer.go goast/serializer_test.go
git commit -m "perf: cache serializer hot path data"
```

### Task 5: Add Benchmarks And Commands

**Files:**
- Create: `goast/bench_test.go`
- Create: `bench/parse.bench.mjs`
- Modify: `package.json`

- [ ] **Step 1: Add Go benchmarks**

Cover:
- parse + serialize
- serialize-only
- unicode-heavy input
- comment-heavy input

- [ ] **Step 2: Add JS end-to-end benchmark**

Benchmark:
- `await initGoAst()` once
- repeated `parseAST()` calls
- small, medium, unicode/comment-heavy, and larger fixture inputs

- [ ] **Step 3: Add package script**

Add:
- `bench`

- [ ] **Step 4: Run benchmarks**

Run:
- `go test -bench . ./goast`
- `bun run bench`

Expected:
- both commands complete successfully and print comparable timing output

- [ ] **Step 5: Commit**

```bash
git add goast/bench_test.go bench/parse.bench.mjs package.json
git commit -m "bench: add parse benchmarks"
```

### Task 6: Update Docs And Verify End-To-End

**Files:**
- Modify: `README.md`
- Modify: `ARCHITECTURE.md`

- [ ] **Step 1: Update docs to match the new architecture**

Document:
- thinner WASM entrypoint
- `goast` parse pipeline ownership
- benchmark commands

- [ ] **Step 2: Run full verification**

Run:
- `go test ./...`
- `bun run build`
- `bun run bench`

Expected:
- tests pass
- build succeeds
- benchmark harness runs successfully

- [ ] **Step 3: Review the diff for OSS quality**

Check for:
- dead code
- inconsistent naming
- undocumented behavior changes
- benchmark noise from cold-start initialization

- [ ] **Step 4: Commit final polish**

```bash
git add README.md ARCHITECTURE.md package.json goast cmd/wasm bench
git commit -m "docs: finalize architecture simplification"
```

- [ ] **Step 5: Prepare PR**

Run:
- `git status --short`
- `git log --oneline --decorate -n 10`

Expected:
- clean worktree except for intentional submodule state if unchanged by this work
- a coherent commit sequence ready to push and open as a PR
