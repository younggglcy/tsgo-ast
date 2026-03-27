#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# Copy Go WASM glue file with a writable mode so repeated builds can overwrite it.
install -m 0644 "$(go env GOROOT)/lib/wasm/wasm_exec.js" "${ROOT_DIR}/npm/wasm_exec.js"

# Build Go WASM
cd "${ROOT_DIR}"
GOOS=js GOARCH=wasm go build -o "${ROOT_DIR}/npm/tsgo-ast.wasm" ./cmd/wasm

echo "Built tsgo-ast.wasm ($(du -h "${ROOT_DIR}/npm/tsgo-ast.wasm" | cut -f1))"
