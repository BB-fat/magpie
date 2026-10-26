import { existsSync, readFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";

export type ClaudeNodeLaunch = {
  interpreter: string;
  entryScript: string;
};

/** Resolve a Claude Code `claude` bin script to node -r … <entry.js>. */
export function resolveClaudeNodeLaunch(claudePath: string): ClaudeNodeLaunch | null {
  const abs = resolve(claudePath);
  if (!existsSync(abs)) return null;
  let text: string;
  try {
    text = readFileSync(abs, "utf8").slice(0, 16_384);
  } catch {
    return null;
  }
  if (!text.startsWith("#!")) return null;

  const interpreter = shebangInterpreter(text);
  if (!interpreter.includes("node")) return null;

  const entry = findClaudeEntryScript(text, dirname(abs));
  if (!entry) return null;
  return { interpreter, entryScript: entry };
}

function shebangInterpreter(firstLines: string): string {
  const line = firstLines.split("\n", 1)[0] ?? "";
  const match = line.match(/^#!\s*(?:\/usr\/bin\/env\s+(\S+)|(\S+))/);
  if (!match) return "node";
  return match[1] ?? match[2] ?? "node";
}

function findClaudeEntryScript(text: string, binDir: string): string | null {
  const patterns = [
    /require\s*\(\s*['"]([^'"]+\.(?:cjs|js|mjs))['"]\s*\)/,
    /import\s+['"]([^'"]+\.(?:cjs|js|mjs))['"]/,
    /import\s*\(\s*['"]([^'"]+\.(?:cjs|js|mjs))['"]\s*\)/,
  ];
  for (const re of patterns) {
    const hit = text.match(re);
    if (!hit?.[1]) continue;
    const candidate = resolve(binDir, hit[1]);
    if (existsSync(candidate)) return candidate;
  }
  return null;
}
