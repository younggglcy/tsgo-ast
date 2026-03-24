# Local Release PR Flow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the changesets-based publish flow with a local command that creates release PRs from commits since the previous tag, then publish automatically when that PR is merged.

**Architecture:** Add a small release helper library plus a local CLI script. The local script owns changelog generation, version bumping, branch creation, push, and PR creation; GitHub Actions only handles merge-time publish, tag creation, and GitHub Release creation. Tests focus on the pure helper logic so changelog formatting and release-note extraction are stable.

**Tech Stack:** Bun, Node.js ESM scripts, GitHub Actions, npm publish with provenance

---

### Task 1: Add failing tests for release helper logic

**Files:**
- Create: `scripts/release-lib.test.mjs`
- Create: `scripts/release-lib.mjs`

- [ ] **Step 1: Write the failing tests**

Add tests for:
- prepending a new changelog entry into an empty file
- prepending a new changelog entry above an existing release
- extracting release notes for a specific version from `CHANGELOG.md`
- rejecting invalid version strings

- [ ] **Step 2: Run test to verify it fails**

Run: `bun test scripts/release-lib.test.mjs`
Expected: FAIL because `scripts/release-lib.mjs` does not yet provide the required helpers.

- [ ] **Step 3: Write minimal implementation**

Implement pure helpers in `scripts/release-lib.mjs` for:
- version validation
- changelog section rendering
- changelog prepend behavior
- release-note extraction

- [ ] **Step 4: Run test to verify it passes**

Run: `bun test scripts/release-lib.test.mjs`
Expected: PASS

### Task 2: Build the local release PR command

**Files:**
- Create: `scripts/release-pr.mjs`
- Modify: `package.json`
- Modify: `npm/package.json`

- [ ] **Step 1: Add the failing command behavior**

Create the command entrypoint and verify it fails cleanly when invoked without a version:

Run: `bun run release:pr`
Expected: FAIL with a usage error.

- [ ] **Step 2: Implement the command**

Implement:
- clean worktree check
- current branch check
- fetch from origin
- last-tag discovery
- commit collection since last tag
- changelog update using `scripts/release-lib.mjs`
- `npm/package.json` version update
- `release/v<version>` branch creation
- commit, push, and `gh pr create`

- [ ] **Step 3: Verify helper-only behavior locally**

Run a dry validation command:

`node scripts/release-pr.mjs 9.9.9 --dry-run`

Expected: prints planned release details without creating a branch or editing files permanently.

### Task 3: Replace the GitHub release workflow

**Files:**
- Modify: `.github/workflows/release.yml`

- [ ] **Step 1: Write the failing workflow assumptions**

Document the new assumptions inline in the workflow comments:
- it runs only on pushes to `main`
- it exits unless the head commit is a release commit
- it must not recreate an existing tag

- [ ] **Step 2: Implement the workflow changes**

Replace `changesets/action` with steps that:
- checkout full history
- setup Go, Node, and Bun
- install dependencies
- read the version from `npm/package.json`
- detect existing tags
- build and publish to npm
- create the git tag
- extract the release notes for that version
- create the GitHub Release

- [ ] **Step 3: Verify workflow syntax**

Run: `bunx prettier --check .github/workflows/release.yml`
Expected: PASS or no formatting changes required

### Task 4: Remove obsolete changesets release plumbing

**Files:**
- Delete: `.changeset/config.json`
- Delete: `.changeset/README.md`
- Delete: `scripts/sync-changelog.js`
- Modify: `package.json`

- [ ] **Step 1: Remove changesets dependencies and scripts**

Delete changesets scripts/dependencies from `package.json` and add the new `release:pr` script plus any test script needed for the release helpers.

- [ ] **Step 2: Remove old release support files**

Delete changesets config and the changelog sync script.

- [ ] **Step 3: Verify install state**

Run: `bun install`
Expected: lockfile updates cleanly with changesets packages removed.

### Task 5: Update repository docs

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Replace the release process documentation**

Document:
- `bun run release:pr <version>`
- changelog generation from commits since the previous release
- merge-to-main publish behavior
- requirement for `gh` authentication locally

- [ ] **Step 2: Sanity-check the documented command**

Run: `bun run release:pr --help`
Expected: shows usage information

### Task 6: Final verification

**Files:**
- Test: `scripts/release-lib.test.mjs`
- Test: `.github/workflows/release.yml`

- [ ] **Step 1: Run targeted automated checks**

Run:
- `bun test scripts/release-lib.test.mjs`
- `go test ./...`

Expected: PASS

- [ ] **Step 2: Run release command dry-run**

Run: `bun run release:pr 9.9.9 --dry-run`
Expected: prints the branch name, previous tag, collected commits, and intended PR title without mutating git state.

- [ ] **Step 3: Inspect git diff**

Run: `git status --short`
Expected: only intended release workflow, script, lockfile, and docs changes remain.
