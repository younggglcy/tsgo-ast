# Changelog

## 0.2.0 - 2026-03-30

- refactor: simplify parse pipeline and add benchmarks (#13)
- ci: upgrade setup-node to v6 for release publish (#12)
- fix: add directory field to repository in package.json
- fix(release): handle squash-merged release commits (#11)

## 0.1.1 - 2026-03-28

- fix(release): handle inherited stdio in release helper
- ci: add required test workflow
- docs: split repo and package docs (#9)
- ci: defer submodule checkout in release (#8)
- fix: align AST offsets and restore shared node fields (#7)
- ci: switch to local release PR workflow (#6)
- docs: add AGENTS.md repo guide (#5)
- build: use rolldown for DTS emission instead of tsc (#4)
- test(goast): add tests for enriched serialization
- feat: update TypeScript types for enriched AST output
- feat(wasm): use enriched Serializer and add sourceFileInfo
- refactor(goast): extract Serializer struct with loc/flags/comments enrichment
- feat(goast): add comment extraction via scanner shim
- feat(goast): add NodeFlags decoder
- chore: add shim/scanner dependency for comment and position APIs
