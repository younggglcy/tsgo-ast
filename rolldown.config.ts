import { defineConfig } from "rolldown";
import { isolatedDeclarationPlugin } from "rolldown/experimental";

export default defineConfig({
  input: "src/index.ts",
  output: {
    dir: "npm",
    format: "esm",
    entryFileNames: "index.js",
  },
  plugins: [isolatedDeclarationPlugin()],
  external: ["./wasm_exec.js"],
});
