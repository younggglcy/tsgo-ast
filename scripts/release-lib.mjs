const VERSION_RE = /^\d+\.\d+\.\d+$/;
const CHANGELOG_HEADER = "# Changelog";

export function validateVersion(version) {
  if (!VERSION_RE.test(version)) {
    throw new Error(`Invalid version: ${version}`);
  }

  return version;
}

export function isReleaseCommitSubject(subject, tag) {
  const escapedTag = tag.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const pattern = new RegExp(`^release: ${escapedTag}( \\(#\\d+\\))?$`);
  return pattern.test(subject);
}

export function renderChangelogEntry({ version, date, commits }) {
  if (!Array.isArray(commits) || commits.length === 0) {
    throw new Error("Release changelog requires at least one commit");
  }

  return [
    `## ${version} - ${date}`,
    "",
    ...commits.map((commit) => `- ${commit}`),
    "",
  ].join("\n");
}

export function prependChangelogEntry(current, entry) {
  const body = current.trim();
  const renderedEntry = renderChangelogEntry(entry);

  if (!body) {
    return `${CHANGELOG_HEADER}\n\n${renderedEntry}`;
  }

  if (!body.startsWith(CHANGELOG_HEADER)) {
    throw new Error("CHANGELOG.md must start with '# Changelog'");
  }

  const rest = body.slice(CHANGELOG_HEADER.length).trimStart();
  return `${CHANGELOG_HEADER}\n\n${renderedEntry}${rest ? `\n${rest}\n` : ""}`;
}

export function extractReleaseNotes(changelog, version) {
  const lines = changelog.split("\n");
  const heading = `## ${version} - `;
  const start = lines.findIndex((line) => line.startsWith(heading));

  if (start === -1) {
    throw new Error(`Release notes not found for version ${version}`);
  }

  let end = lines.length;
  for (let i = start + 1; i < lines.length; i++) {
    if (lines[i].startsWith("## ")) {
      end = i;
      break;
    }
  }

  return lines
    .slice(start + 1, end)
    .join("\n")
    .trim();
}
