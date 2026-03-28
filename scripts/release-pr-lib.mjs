export function normalizeExecOutput(output) {
  if (typeof output !== "string") {
    return "";
  }

  return output.trim();
}
