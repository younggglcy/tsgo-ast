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
  loc?: {
    startLine: number;
    startColumn: number;
    endLine: number;
    endColumn: number;
  };
  flags?: string[];
  leadingComments?: AstComment[];
  trailingComments?: AstComment[];
  [key: string]: unknown;
}

export interface AstComment {
  type: "line" | "block";
  text: string;
  start: number;
  end: number;
}

export interface SourceFileInfo {
  isDeclarationFile: boolean;
  pragmas: string[] | null;
  referencedFiles: FileReference[] | null;
  typeReferenceDirectives: FileReference[] | null;
}

export interface FileReference {
  fileName: string;
  start: number;
  end: number;
}

export interface ParseResult {
  ast: GoAstNode;
  errors: string[] | null;
  sourceFileInfo: SourceFileInfo;
}

let initPromise: Promise<void> | null = null;

/**
 * Initialize the Go WASM runtime. Must be called before parseAST().
 * Safe to call concurrently — only the first call triggers initialization.
 * @param wasmUrl - Custom URL to the WASM file. Defaults to the bundled tsgo-ast.wasm.
 */
export function initGoAst(wasmUrl?: string | URL): Promise<void> {
  if (!initPromise) {
    initPromise = doInit(wasmUrl);
  }
  return initPromise;
}

async function doInit(wasmUrl?: string | URL): Promise<void> {
  await import("./wasm_exec.js");
  const go = new Go();
  const url = wasmUrl ?? new URL("./tsgo-ast.wasm", import.meta.url);
  const response = fetch(url);
  let result: WebAssembly.WebAssemblyInstantiatedSource;
  try {
    result = await WebAssembly.instantiateStreaming(
      response,
      go.importObject,
    );
  } catch {
    // Fallback for environments where instantiateStreaming fails
    // (e.g. missing application/wasm MIME type, file:// protocol)
    const bytes = await (await response).arrayBuffer();
    result = await WebAssembly.instantiate(bytes, go.importObject);
  }
  // Don't await — Go main() blocks forever with select{}
  // Attach .catch() to prevent unhandled rejection if runtime fails
  go.run(result.instance).catch((err) => {
    console.error("Go WASM runtime error:", err);
  });
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
  if (!initPromise) throw new Error("Call initGoAst() first");
  return goParseAST(code, lang);
}

/**
 * Check whether the Go WASM runtime has been initialized.
 */
export function isInitialized(): boolean {
  return initPromise !== null;
}
