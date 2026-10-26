// What magpie changed of the wrapper: hooks only in the run's own
// settings.local.json, run by the absolute Bun; the JSON Schema bounds;
// private, whole writes.
import { describe, expect, test } from "bun:test";
import { mkdtempSync, readFileSync, statSync, writeFileSync, existsSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { buildRuntimeHookCommand, installJsonSchemaStopHook, installRuntimeHooks } from "../src/hook-install.ts";
import { JSON_LIMITS, extractAndValidateStructuredOutput, jsonShapeError, loadJsonSchemaFromText, validateJsonAgainstSchema } from "../src/json-schema.ts";
import { writePrivateAtomic } from "../src/fs-atomic.ts";
import { preAcceptProject } from "../src/preaccept.ts";

const dir = () => mkdtempSync(join(tmpdir(), "magpie-wrapper-test-"));

describe("hooks", () => {
  test("run by the absolute Bun, whatever PATH holds", () => {
    const cmd = buildRuntimeHookCommand(["/x/magpie-bun", "/s/claude-bridge.js"], "/opt/magpie-bun");
    expect(cmd).toBe("'/opt/magpie-bun' '/s/claude-bridge.js' '--desplega-internal-runtime-hook'");
    expect(cmd).not.toContain("env bun");
  });

  test("written to the settings file named, private, never another", () => {
    const d = dir();
    const path = join(d, ".claude", "settings.local.json");
    installRuntimeHooks({ settingsPath: path, command: "'/b' '/w.js' '--desplega-internal-runtime-hook'" });
    installJsonSchemaStopHook({ settingsPath: path, command: "'/b' '/w.js' '--desplega-internal-json-schema-stop-hook'" });
    const s = JSON.parse(readFileSync(path, "utf8"));
    expect(Object.keys(s.hooks).sort()).toEqual(["MessageDisplay", "PreToolUse", "SessionStart", "Stop", "StopFailure", "UserPromptSubmit"]);
    expect(statSync(path).mode & 0o777).toBe(0o600);
  });

  test("a settings file that doesn't parse is not written over", () => {
    const d = dir();
    const path = join(d, "settings.local.json");
    writeFileSync(path, "{not json");
    expect(() => installRuntimeHooks({ settingsPath: path, command: "c --desplega-internal-runtime-hook" })).toThrow();
    expect(readFileSync(path, "utf8")).toBe("{not json");
  });
});

describe("JSON Schema bounds", () => {
  test("a schema over 256 KiB is refused", () => {
    const big = JSON.stringify({ type: "object", description: "x".repeat(JSON_LIMITS.schemaBytes) });
    expect(() => loadJsonSchemaFromText(big, 1e9)).toThrow(/256 KiB/);
  });

  test("a schema at the depth bound is taken, one deeper is refused", () => {
    // n arrays one in another, a number in the last: n values deep
    const nest = (n: number): unknown => (n <= 1 ? 0 : [nest(n - 1)]);
    expect(jsonShapeError(nest(JSON_LIMITS.depth), "s")).toBeNull();
    expect(jsonShapeError(nest(JSON_LIMITS.depth + 1), "s")).toMatch(/deeper than 64/);
  });

  test("a list past 10,000 items and a value past 100,000 nodes are refused", () => {
    expect(jsonShapeError(new Array(JSON_LIMITS.items).fill(0), "v")).toBeNull();
    expect(jsonShapeError(new Array(JSON_LIMITS.items + 1).fill(0), "v")).toMatch(/10000 items/);
    const wide = Array.from({ length: 11 }, () => new Array(JSON_LIMITS.items).fill(0));
    expect(jsonShapeError(wide, "v")).toMatch(/100000 values/);
  });

  test("a reply over 4 MiB is not checked", () => {
    const r = extractAndValidateStructuredOutput(`{"a":"${"x".repeat(JSON_LIMITS.outputBytes)}"}`, { type: "object" });
    expect(r.ok).toBe(false);
    if (!r.ok) expect(r.message).toMatch(/4 MiB/);
  });

  test("at most 8 issues are said, and how many more", () => {
    const props: Record<string, unknown> = {};
    for (let i = 0; i < 12; i++) props[`k${i}`] = { type: "string" };
    const v = validateJsonAgainstSchema({}, { type: "object", properties: props, required: Object.keys(props) } as never);
    expect(v.ok).toBe(false);
    if (!v.ok) {
      expect(v.issues.length).toBe(9);
      expect(v.issues[8]).toBe("and 4 more");
    }
  });
});

describe("profile writes", () => {
  test("whole and private", () => {
    const d = dir();
    const path = join(d, "f.json");
    writePrivateAtomic(path, "{}");
    writePrivateAtomic(path, '{"a":1}');
    expect(readFileSync(path, "utf8")).toBe('{"a":1}');
    expect(statSync(path).mode & 0o777).toBe(0o600);
  });

  test("trust goes to the profile CLAUDE_CONFIG_DIR names", () => {
    const d = dir();
    const old = process.env.CLAUDE_CONFIG_DIR;
    process.env.CLAUDE_CONFIG_DIR = d;
    try {
      preAcceptProject({ workdir: "/tmp/w", mcpServerNames: [] });
    } finally {
      if (old === undefined) delete process.env.CLAUDE_CONFIG_DIR;
      else process.env.CLAUDE_CONFIG_DIR = old;
    }
    const j = JSON.parse(readFileSync(join(d, ".claude.json"), "utf8"));
    expect(j.projects["/tmp/w"].hasTrustDialogAccepted).toBe(true);
    expect(existsSync(join(d, ".claude.json"))).toBe(true);
  });
});
