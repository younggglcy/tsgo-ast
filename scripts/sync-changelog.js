import { existsSync, copyFileSync } from "node:fs";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const rootDir = resolve(__dirname, "..");
const src = resolve(rootDir, "npm/CHANGELOG.md");
const dest = resolve(rootDir, "CHANGELOG.md");

if (existsSync(src)) {
  copyFileSync(src, dest);
  console.log("Copied npm/CHANGELOG.md → CHANGELOG.md");
} else {
  console.log("No npm/CHANGELOG.md found, skipping sync");
}
