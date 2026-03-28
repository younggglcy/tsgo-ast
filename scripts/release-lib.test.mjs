import { describe, expect, test } from "bun:test";
import {
  extractReleaseNotes,
  isReleaseCommitSubject,
  prependChangelogEntry,
  validateVersion,
} from "./release-lib.mjs";

describe("validateVersion", () => {
  test("accepts x.y.z versions", () => {
    expect(validateVersion("1.2.3")).toBe("1.2.3");
  });

  test("rejects invalid version strings", () => {
    expect(() => validateVersion("1.2")).toThrow("Invalid version");
    expect(() => validateVersion("v1.2.3")).toThrow("Invalid version");
  });
});

describe("prependChangelogEntry", () => {
  test("creates a changelog with a header when empty", () => {
    const next = prependChangelogEntry("", {
      version: "0.2.0",
      date: "2026-03-25",
      commits: ["feat: add release pipeline"],
    });

    expect(next).toBe(`# Changelog

## 0.2.0 - 2026-03-25

- feat: add release pipeline
`);
  });

  test("prepends a new entry above an existing release", () => {
    const current = `# Changelog

## 0.1.0 - 2026-03-20

- feat: initial release
`;

    const next = prependChangelogEntry(current, {
      version: "0.2.0",
      date: "2026-03-25",
      commits: ["feat: add release pipeline", "fix: tighten workflow gating"],
    });

    expect(next).toBe(`# Changelog

## 0.2.0 - 2026-03-25

- feat: add release pipeline
- fix: tighten workflow gating

## 0.1.0 - 2026-03-20

- feat: initial release
`);
  });
});

describe("extractReleaseNotes", () => {
  test("returns only the requested version section", () => {
    const changelog = `# Changelog

## 0.2.0 - 2026-03-25

- feat: add release pipeline
- fix: tighten workflow gating

## 0.1.0 - 2026-03-20

- feat: initial release
`;

    expect(extractReleaseNotes(changelog, "0.2.0")).toBe(`- feat: add release pipeline
- fix: tighten workflow gating`);
  });
});

describe("isReleaseCommitSubject", () => {
  test("matches an exact release commit subject", () => {
    expect(isReleaseCommitSubject("release: v0.2.0", "v0.2.0")).toBe(true);
  });

  test("matches a squash-merge release subject with PR suffix", () => {
    expect(isReleaseCommitSubject("release: v0.2.0 (#10)", "v0.2.0")).toBe(true);
  });

  test("rejects unrelated subjects", () => {
    expect(isReleaseCommitSubject("fix: release helper", "v0.2.0")).toBe(false);
  });
});
