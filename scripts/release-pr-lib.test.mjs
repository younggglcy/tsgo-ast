import { describe, expect, test } from "bun:test";
import { normalizeExecOutput } from "./release-pr-lib.mjs";

describe("normalizeExecOutput", () => {
  test("returns an empty string for null exec output", () => {
    expect(normalizeExecOutput(null)).toBe("");
  });

  test("trims string exec output", () => {
    expect(normalizeExecOutput("  hello world \n")).toBe("hello world");
  });
});
