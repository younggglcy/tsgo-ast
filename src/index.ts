declare global {
  class Go {
    importObject: WebAssembly.Imports;
    run(instance: WebAssembly.Instance): Promise<void>;
  }
  function goParseAST(code: string, lang: string): ParseResult;
}

export interface GoAstNode {
  type: string;
  start: number;
  end: number;
  [key: string]: unknown;
}

export interface ParseResult {
  ast: GoAstNode;
  errors: string[];
}

let initialized = false;

/**
 * Initialize the Go WASM runtime. Must be called before parseAST().
 * @param wasmUrl - Custom URL to the WASM file. Defaults to the bundled tsgo-ast.wasm.
 */
export async function initGoAst(wasmUrl?: string | URL): Promise<void> {
  if (initialized) return;
  await import("./wasm_exec.js");
  const go = new Go();
  const url = wasmUrl ?? new URL("./tsgo-ast.wasm", import.meta.url);
  const result = await WebAssembly.instantiateStreaming(
    fetch(url),
    go.importObject,
  );
  // Don't await — Go main() blocks forever with select{}
  go.run(result.instance);
  initialized = true;
}

/**
 * Parse TypeScript/JavaScript code and return the AST with Go type names.
 * @param code - Source code to parse
 * @param lang - Source language (default: 'tsx')
 */
export function parseAST(
  code: string,
  lang: "ts" | "tsx" | "js" | "jsx" = "tsx",
): ParseResult {
  if (!initialized) throw new Error("Call initGoAst() first");
  return goParseAST(code, lang);
}

/**
 * Check whether the Go WASM runtime has been initialized.
 */
export function isInitialized(): boolean {
  return initialized;
}
