import { initGoAst, parseAST } from "../npm/index.js";

function buildCommentHeavySource(repeat) {
  const parts = [];
  for (let i = 0; i < repeat; i++) {
    parts.push(`// leading note ${i}`);
    parts.push(`const value${i} = ${i}; /* trailing note ${i} */`);
  }
  parts.push(`const cafe = "caf\u00e9";`);
  parts.push(`const party = "\ud83c\udf89";`);
  return parts.join("\n");
}

function buildTSXFixture(repeat) {
  const parts = [
    "type Item = { id: number; label: string };",
    "export function App(props: { items: Item[] }) {",
    "  return <section>",
  ];

  for (let i = 0; i < repeat; i++) {
    parts.push(
      `    <article key={props.items[${i}]?.id ?? ${i}}>{props.items[${i}]?.label ?? "item"}</article>`,
    );
  }

  parts.push("  </section>;");
  parts.push("}");
  return parts.join("\n");
}

function formatNumber(value) {
  return value.toFixed(3);
}

function measureFixture({ name, lang, code, iterations }) {
  const first = parseAST(code, lang);
  if (first.ast == null) {
    throw new Error(`Fixture ${name} did not return an AST`);
  }
  if (first.errors?.length) {
    throw new Error(`Fixture ${name} produced parse errors: ${first.errors.join("; ")}`);
  }

  const start = performance.now();
  for (let i = 0; i < iterations; i++) {
    parseAST(code, lang);
  }
  const totalMs = performance.now() - start;
  const avgMs = totalMs / iterations;
  const opsPerSecond = (iterations * 1000) / totalMs;

  return {
    name,
    lang,
    bytes: Buffer.byteLength(code),
    iterations,
    totalMs,
    avgMs,
    opsPerSecond,
  };
}

const fixtures = [
  {
    name: "small-ts",
    lang: "ts",
    iterations: 400,
    code: [
      "const value: number = 1;",
      "export const doubled = value * 2;",
      "export function square(input: number) {",
      "  return input * input;",
      "}",
    ].join("\n"),
  },
  {
    name: "medium-tsx",
    lang: "tsx",
    iterations: 160,
    code: buildTSXFixture(40),
  },
  {
    name: "unicode-comments",
    lang: "ts",
    iterations: 160,
    code: buildCommentHeavySource(120),
  },
  {
    name: "large-tsx",
    lang: "tsx",
    iterations: 24,
    code: buildTSXFixture(220),
  },
];

const wasmUrl = new URL("../npm/tsgo-ast.wasm", import.meta.url);
await initGoAst(wasmUrl);

console.log("Steady-state parseAST() benchmark");
console.log("WASM runtime initialized once before timing");

for (const fixture of fixtures) {
  const result = measureFixture(fixture);
  console.log(
    [
      result.name.padEnd(18),
      `lang=${result.lang}`,
      `bytes=${result.bytes}`,
      `iters=${result.iterations}`,
      `totalMs=${formatNumber(result.totalMs)}`,
      `avgMs=${formatNumber(result.avgMs)}`,
      `ops/s=${formatNumber(result.opsPerSecond)}`,
    ].join("  "),
  );
}
