# Design: OIDC + Changesets Publish System

## Problem

The current release process has several gaps:

- Uses `NPM_TOKEN` secret for npm authentication (less secure than OIDC)
- No changelog generation — users have no visibility into what changed between versions
- No GitHub Releases — tags exist but without release notes
- Manual version bumping and tag pushing — error-prone
- No provenance attestation on published packages

## Goals

1. Replace `NPM_TOKEN` with OIDC token-less publishing (npm trusted publishers)
2. Integrate `@changesets/cli` for versioning and changelog generation
3. Auto-create GitHub Releases with changelog content as body
4. Keep a single `CHANGELOG.md` at project root covering Go + TS changes
5. Fully automated: merge to main → Version PR → merge → publish

## Architecture

### Release Flow

```
Developer workflow:
  1. Make changes
  2. Run `bunx changeset` to create a changeset file
  3. Commit changeset + code → push → merge PR to main

CI automation (on push to main):
  1. changesets/action detects pending changesets
  2. Creates/updates "Version Packages" PR:
     - Bumps version in npm/package.json
     - Generates CHANGELOG.md at project root
  3. When Version PR is merged:
     - Builds WASM + JS
     - Publishes to npm via OIDC (no token)
     - Creates GitHub Release with changelog as body
```

### OIDC Configuration

**npm side (one-time manual setup):**

- Go to npmjs.com → `tsgo-ast` package settings → Trusted publishing
- Add: owner=`younggglcy`, repo=`tsgo-ast`, workflow=`release.yml`

**GitHub Actions permissions:**

```yaml
permissions:
  contents: write       # Create tags + GitHub Releases
  pull-requests: write  # Create Version PR
  id-token: write       # OIDC token-less publishing
```

**Publish command:**

```bash
npm publish --provenance --access public
```

- `--provenance`: SLSA signed attestation (build origin traceable)
- `--access public`: Required for public unscoped packages
- No `NODE_AUTH_TOKEN` or `NPM_TOKEN` needed

**Prerequisites:**

- Package must have been published at least once with a token (v0.1.0 already exists ✅)
- GitHub repository must be public
- Must use `npm` CLI (not bun) for provenance/OIDC features

### Changesets Integration

**Package discovery:** Add `"workspaces": ["npm"]` to root `package.json` so changesets' `@manypkg/get-packages` finds `npm/package.json` as the publishable package.

**CHANGELOG.md location:** Changesets generates CHANGELOG.md in the package directory (`npm/`). A post-version script moves it to the project root.

**New files:**

```
.changeset/
  config.json          # Changesets configuration
  README.md            # Auto-generated instructions
CHANGELOG.md           # Project root, all version history
```

**`.changeset/config.json`:**

```json
{
  "$schema": "https://unpkg.com/@changesets/config@3/schema.json",
  "changelog": [
    "@changesets/changelog-github",
    { "repo": "younggglcy/tsgo-ast" }
  ],
  "commit": false,
  "fixed": [],
  "linked": [],
  "access": "public",
  "baseBranch": "main",
  "updateInternalDependencies": "patch",
  "ignore": []
}
```

**New devDependencies:**

- `@changesets/cli`
- `@changesets/changelog-github`

**New scripts in root `package.json`:**

```json
{
  "scripts": {
    "changeset": "changeset",
    "version": "changeset version && node scripts/sync-changelog.js",
    "release": "bun run build && cd npm && npm publish --provenance --access public"
  }
}
```

### GitHub Actions Workflow (`release.yml`)

Replaces the current tag-triggered workflow with a main-branch-triggered one:

```yaml
name: Release
on:
  push:
    branches: [main]

permissions:
  contents: write
  pull-requests: write
  id-token: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"

      - uses: actions/setup-node@v4
        with:
          node-version: "22"
          registry-url: "https://registry.npmjs.org"

      - uses: oven-sh/setup-bun@v2

      - run: bun install

      - name: Create Release Pull Request or Publish
        uses: changesets/action@v1
        with:
          version: bun run version
          publish: bun run release
          createGithubReleases: true
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### CHANGELOG.md Sync Script

`scripts/sync-changelog.js` — moves `npm/CHANGELOG.md` to project root after `changeset version`:

```js
import { existsSync, renameSync } from "node:fs";
const src = "npm/CHANGELOG.md";
const dest = "CHANGELOG.md";
if (existsSync(src)) {
  renameSync(src, dest);
}
```

### Go Codebase and CHANGELOG

- Go code is not independently published — it's WASM compilation source only
- A single CHANGELOG.md covers the entire project (Go + TS + WASM)
- Changeset descriptions are user-facing and don't distinguish Go/TS internals
- If Go module publishing is needed in the future, the same CHANGELOG and git tags are compatible with Go module versioning conventions

## Migration Steps

1. Configure npm trusted publishing on npmjs.com
2. Install changesets devDependencies
3. Initialize changesets (`bunx changeset init`)
4. Configure `.changeset/config.json`
5. Add workspaces to root `package.json`
6. Add version/release scripts
7. Create `scripts/sync-changelog.js`
8. Replace `release.yml` workflow
9. Update CLAUDE.md release process documentation
10. Test with a real release
11. Delete `NPM_TOKEN` secret from GitHub repo settings

## Risks and Mitigations

| Risk | Mitigation |
|------|-----------|
| OIDC misconfiguration blocks publish | Keep `NPM_TOKEN` as fallback until first OIDC publish succeeds |
| Changesets doesn't find `npm/` package | Workspaces declaration ensures discovery |
| CHANGELOG.md in wrong location | `sync-changelog.js` script handles relocation |
| No tests before publish | Out of scope for this design, but noted as future improvement |
| `changesets/action` version incompatibility | Pin to `@v1`, monitor for breaking changes |
