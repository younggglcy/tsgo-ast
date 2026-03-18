import { defineConfig } from "rolldown";

export default defineConfig({
  input: "src/index.ts",
  output: {
    dir: "npm",
    format: "esm",
    entryFileNames: "index.js",
  },
  external: ["./wasm_exec.js"],
});
