import { execFileSync } from "node:child_process";
import { existsSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import {
  prependChangelogEntry,
  validateVersion,
} from "./release-lib.mjs";

const __dirname = dirname(fileURLToPath(import.meta.url));
const rootDir = resolve(__dirname, "..");
const changelogPath = resolve(rootDir, "CHANGELOG.md");
const npmPackagePath = resolve(rootDir, "npm/package.json");

function printUsage() {
  console.log(`Usage: bun run release:pr <version> [--dry-run]

Create a release PR from commits since the previous v* tag.

Options:
  --dry-run   Print planned changes without mutating git state
  --help      Show this help message`);
}

function fail(message) {
  console.error(message);
  process.exit(1);
}

function run(command, args, options = {}) {
  return execFileSync(command, args, {
    cwd: rootDir,
    encoding: "utf8",
    stdio: options.stdio ?? ["ignore", "pipe", "pipe"],
  }).trim();
}

function runAllowFailure(command, args, options = {}) {
  try {
    return {
      ok: true,
      stdout: run(command, args, options),
    };
  } catch (error) {
    return {
      ok: false,
      stdout: error.stdout?.toString().trim() ?? "",
      stderr: error.stderr?.toString().trim() ?? "",
      status: error.status ?? 1,
    };
  }
}

function ensureCleanWorktree() {
  const status = run("git", ["status", "--porcelain"]);
  if (status) {
    fail("Working tree must be clean before creating a release PR.");
  }
}

function ensureOnMain() {
  const branch = run("git", ["branch", "--show-current"]);
  if (branch !== "main") {
    fail(`Release PR must be created from main, current branch is '${branch}'.`);
  }
}

function ensureBranchDoesNotExist(branchName) {
  const local = runAllowFailure("git", [
    "show-ref",
    "--verify",
    "--quiet",
    `refs/heads/${branchName}`,
  ]);
  if (local.ok) {
    fail(`Branch already exists locally: ${branchName}`);
  }

  const remote = runAllowFailure("git", [
    "ls-remote",
    "--exit-code",
    "--heads",
    "origin",
    branchName,
  ]);
  if (remote.ok) {
    fail(`Branch already exists on origin: ${branchName}`);
  }
}

function ensureTagDoesNotExist(tagName) {
  const local = runAllowFailure("git", [
    "show-ref",
    "--verify",
    "--quiet",
    `refs/tags/${tagName}`,
  ]);
  if (local.ok) {
    fail(`Tag already exists locally: ${tagName}`);
  }

  const remote = runAllowFailure("git", [
    "ls-remote",
    "--exit-code",
    "--tags",
    "origin",
    tagName,
  ]);
  if (remote.ok) {
    fail(`Tag already exists on origin: ${tagName}`);
  }
}

function findLastTag() {
  const result = runAllowFailure("git", [
    "describe",
    "--tags",
    "--abbrev=0",
    "--match",
    "v[0-9]*",
  ]);

  return result.ok ? result.stdout : null;
}

function collectCommitsSince(lastTag) {
  const args = ["log", "--pretty=format:%s", "--no-merges"];
  if (lastTag) {
    args.splice(1, 0, `${lastTag}..HEAD`);
  }

  const output = run("git", args);
  const commits = output
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean);

  if (commits.length === 0) {
    fail("No commits found since the previous release.");
  }

  return commits;
}

function loadPackageJson() {
  return JSON.parse(readFileSync(npmPackagePath, "utf8"));
}

function updatePackageVersion(version) {
  const packageJson = loadPackageJson();
  packageJson.version = version;
  return `${JSON.stringify(packageJson, null, 2)}\n`;
}

function getDateStamp() {
  const date = new Date();
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function createPrBody(version, lastTag, commits) {
  const compareTarget = lastTag ?? "initial commit";
  const bullets = commits.map((commit) => `- ${commit}`).join("\n");
  return `## Release Summary

- Version: \`v${version}\`
- Previous release: \`${compareTarget}\`
- Commits included: ${commits.length}

## Included Changes

${bullets}`;
}

function createReleasePlan(version) {
  const branchName = `release/v${version}`;
  const tagName = `v${version}`;
  const lastTag = findLastTag();
  const commits = collectCommitsSince(lastTag);
  const changelogCurrent = existsSync(changelogPath)
    ? readFileSync(changelogPath, "utf8")
    : "";
  const changelogNext = prependChangelogEntry(changelogCurrent, {
    version,
    date: getDateStamp(),
    commits,
  });
  const packageJsonNext = updatePackageVersion(version);

  return {
    branchName,
    tagName,
    lastTag,
    commits,
    changelogNext,
    packageJsonNext,
    prTitle: `release: ${tagName}`,
    prBody: createPrBody(version, lastTag, commits),
  };
}

function main() {
  const args = process.argv.slice(2);
  const dryRun = args.includes("--dry-run");
  const help = args.includes("--help");
  const versionArg = args.find((arg) => !arg.startsWith("--"));

  if (help) {
    printUsage();
    return;
  }

  if (!versionArg) {
    printUsage();
    process.exit(1);
  }

  const version = validateVersion(versionArg);
  const branchName = `release/v${version}`;
  const tagName = `v${version}`;

  if (!dryRun) {
    ensureCleanWorktree();
    ensureOnMain();

    console.log("Fetching origin...");
    run("git", ["fetch", "--tags", "origin"], { stdio: "inherit" });

    ensureBranchDoesNotExist(branchName);
    ensureTagDoesNotExist(tagName);
  }

  const plan = createReleasePlan(version);

  if (dryRun) {
    console.log(`Dry run for ${plan.prTitle}`);
    console.log(`Branch: ${plan.branchName}`);
    console.log(`Previous tag: ${plan.lastTag ?? "(none)"}`);
    console.log("Commits:");
    for (const commit of plan.commits) {
      console.log(`- ${commit}`);
    }
    console.log("");
    console.log("CHANGELOG preview:");
    console.log(plan.changelogNext);
    return;
  }

  writeFileSync(changelogPath, plan.changelogNext);
  writeFileSync(npmPackagePath, plan.packageJsonNext);

  run("git", ["checkout", "-b", plan.branchName], { stdio: "inherit" });
  run("git", ["add", "CHANGELOG.md", "npm/package.json"], { stdio: "inherit" });
  run("git", ["commit", "-m", plan.prTitle], { stdio: "inherit" });
  run("git", ["push", "-u", "origin", plan.branchName], { stdio: "inherit" });
  run(
    "gh",
    [
      "pr",
      "create",
      "--base",
      "main",
      "--head",
      plan.branchName,
      "--title",
      plan.prTitle,
      "--body",
      plan.prBody,
    ],
    { stdio: "inherit" },
  );
}

main();
