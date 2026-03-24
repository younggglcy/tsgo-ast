# Design: Local Release PR Flow

## Goal

Replace the current changesets-driven release automation with a simpler flow:

1. A local command creates a dedicated release PR.
2. That command generates `CHANGELOG.md` from commits since the previous release tag.
3. Merging the release PR to `main` creates the git tag, GitHub Release, and publishes the npm package.

## Why Change

The current setup uses changesets to either open a version PR or publish directly on pushes to `main`. That is more automation than needed for this repository and requires maintaining changeset files in normal feature work.

The desired workflow is intentionally simpler:

- release intent is explicit
- the release PR is prepared locally in one command
- changelog generation is automatic but based only on git history
- merge-to-release stays automated on GitHub

## Release Flow

### Local release preparation

The developer runs:

```bash
bun run release:pr 0.2.0
```

That command:

1. Verifies the working tree is clean.
2. Verifies the current branch is `main`.
3. Fetches remote refs.
4. Verifies the target version argument is a valid semver-like `x.y.z` string.
5. Finds the most recent `v*` tag. If none exists, it uses the first commit as the lower bound.
6. Collects commit subjects between the last release tag and `HEAD`, excluding merge commits.
7. Prepends a new entry to `CHANGELOG.md`.
8. Updates `npm/package.json` to the requested version.
9. Creates and checks out `release/v<version>`.
10. Commits the release changes, pushes the branch, and opens a PR with `gh pr create`.

### Merge-time publishing

After the release PR is merged into `main`, GitHub Actions:

1. Verifies that the head commit is a release commit.
2. Reads the version from `npm/package.json`.
3. Exits early if `v<version>` already exists.
4. Builds the package and publishes to npm.
5. Creates the annotated git tag `v<version>`.
6. Creates a GitHub Release whose notes come from the matching section in `CHANGELOG.md`.

## Changelog Format

The changelog stays intentionally simple:

```md
# Changelog

## 0.2.0 - 2026-03-25

- feat: add sourceFileInfo to enriched AST output
- test(goast): add tests for enriched serialization
- build: use rolldown for DTS emission instead of tsc (#4)
```

Rules:

- newest release goes first
- one bullet per commit subject
- merge commits are omitted
- if no commits exist since the last tag, release creation fails

## Files and Responsibilities

- `scripts/release-pr.mjs`: orchestrates local release PR preparation
- `scripts/release-lib.mjs`: shared helpers for version parsing, changelog generation, and release note extraction
- `scripts/release-lib.test.mjs`: regression tests for the pure release helper logic
- `.github/workflows/release.yml`: merge-time release workflow
- `package.json`: release command entrypoints and dependency cleanup
- `CLAUDE.md`: repository release-process documentation

## Key Decisions

### Commit-based changelog generation

Use commit subjects instead of PR metadata. This keeps the implementation independent of PR merge conventions and avoids GitHub API lookups during local release preparation.

### Manual version selection

The release command requires the version as an argument. Version bump policy remains human-controlled and explicit.

### Local PR creation

The command uses `gh pr create` so the human only needs one local invocation to prepare the release PR.

### Release workflow no longer creates PRs

GitHub Actions should publish only after a release PR is merged. It should no longer generate version PRs or maintain release metadata in the branch automatically.

## Risks

- Commit subjects may occasionally be noisy. This is acceptable for now because the repository history is already fairly clean.
- Local release creation depends on `gh` being installed and authenticated.
- The release workflow must remain idempotent so reruns do not duplicate tags or GitHub Releases.
