// What magpie changed of the wrapper: hooks only in the run's own
// settings.local.json, run by the absolute Bun; the JSON Schema bounds;
// private, whole writes.
import { afterAll, describe, expect, test } from "bun:test";
import { spawnSync } from "node:child_process";
import { mkdtempSync, readFileSync, statSync, writeFileSync, existsSync, mkdirSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
import { buildRuntimeHookCommand, installJsonSchemaStopHook, installRuntimeHooks } from "../src/hook-install.ts";
import { JSON_LIMITS, extractAndValidateStructuredOutput, jsonShapeError, loadJsonSchemaFromText, validateJsonAgainstSchema } from "../src/json-schema.ts";
import { writePrivateAtomic } from "../src/fs-atomic.ts";
import { canBypassPermissions, preAcceptProject } from "../src/preaccept.ts";
import { parseCliArgs } from "../src/args.ts";

test("native reply-on-resume is forwarded without inventing a prompt", () => {
  const result = parseCliArgs(["-p", "--resume", "saved-session", "--reply-on-resume"]);
  expect(result.ok).toBe(true);
  if (!result.ok) throw new Error(result.error.message);
  expect(result.parsed.initialPrompt).toBeUndefined();
  expect(result.parsed.claudeArgs).toEqual(["--resume", "saved-session", "--reply-on-resume"]);
});

const testRoot = mkdtempSync(join(tmpdir(), "magpie-wrapper-test-"));
afterAll(() => rmSync(testRoot, { recursive: true, force: true }));
const dir = () => mkdtempSync(join(testRoot, "case-"));

const permissionCases: { name: string; env: NodeJS.ProcessEnv; rootBypass: boolean }[] = [
  { name: "unset sandbox", env: {}, rootBypass: false },
  { name: "empty sandbox", env: { IS_SANDBOX: "" }, rootBypass: false },
  { name: "disabled sandbox", env: { IS_SANDBOX: "0" }, rootBypass: false },
  { name: "boolean sandbox", env: { IS_SANDBOX: "true" }, rootBypass: false },
  { name: "explicit sandbox", env: { IS_SANDBOX: "1" }, rootBypass: true },
  { name: "empty bubblewrap", env: { CLAUDE_CODE_BUBBLEWRAP: "" }, rootBypass: false },
  { name: "bubblewrap", env: { CLAUDE_CODE_BUBBLEWRAP: "/sandbox/bwrap" }, rootBypass: true },
  { name: "bubblewrap with disabled sandbox", env: { IS_SANDBOX: "0", CLAUDE_CODE_BUBBLEWRAP: "1" }, rootBypass: true },
];

describe("permission bypass", () => {
  for (const uid of [0, 1000, undefined]) {
    for (const c of permissionCases) {
      test(`uid ${uid}, ${c.name}`, () => {
        expect(canBypassPermissions(c.env, uid)).toBe(uid !== 0 || c.rootBypass);
      });
    }
  }

  for (const uid of [0, 1000, undefined]) {
    for (const c of permissionCases) {
      test(`CLI settings and launch: uid ${uid}, ${c.name}`, () => {
        const d = dir();
        const work = join(d, "work");
        const profile = join(d, "profile");
        mkdirSync(join(work, ".claude"), { recursive: true });
        const settingsPath = join(work, ".claude", "settings.local.json");
        writeFileSync(settingsPath, JSON.stringify({
          defaultMode: "bypassPermissions",
          skipDangerousModePermissionPrompt: true,
          env: { KEEP_TEST_SETTING: "1" },
        }));
        const tmux = join(d, "tmux");
        const launchPath = join(d, "launch.json");
        writeFileSync(tmux, `#!${process.execPath}\nrequire("node:fs").writeFileSync(process.env.TEST_LAUNCH_PATH, JSON.stringify(process.argv.slice(2))); process.exit(73);\n`, { mode: 0o700 });
        const entry = join(d, "entry.ts");
        const cli = fileURLToPath(new URL("../src/cli.ts", import.meta.url));
        writeFileSync(entry, `process.getuid = ${uid === undefined ? "undefined" : `() => ${uid}`}; await import(${JSON.stringify(cli)});\n`);
        const env: NodeJS.ProcessEnv = {
          ...process.env,
          HOME: d,
          USERPROFILE: d,
          XDG_CONFIG_HOME: join(d, "config"),
          XDG_CACHE_HOME: join(d, "cache"),
          XDG_DATA_HOME: join(d, "data"),
          XDG_STATE_HOME: join(d, "state"),
          APPDATA: join(d, "appdata"),
          LOCALAPPDATA: join(d, "localappdata"),
          CLAUDE_CONFIG_DIR: profile,
          CLAUDE_BRIDGE_TMUX_PATH: tmux,
          CLAUDE_BRIDGE_CLAUDE_PATH: join(d, "unused-claude"),
          TEST_LAUNCH_PATH: launchPath,
        };
        for (const key of ["IS_SANDBOX", "CLAUDE_CODE_BUBBLEWRAP", "CLAUDE_BRIDGE_TMUX_SOCKET", "CLAUDE_BRIDGE_USER_CONTEXT_PATH", "MAGPIE_CLAUDE_TUI_DEBUG_API", "MAGPIE_CLAUDE_NATIVE_TMUX_GUARD_PATH"]) delete env[key];
        Object.assign(env, c.env);
        const result = spawnSync(process.execPath, [entry, "-p", "test prompt", "--desplega-cwd", work], { env, encoding: "utf8", timeout: 10_000 });
        expect(result.error).toBeUndefined();
        expect(result.stderr).toContain("tmux new-session exited 73");
        expect(result.status).toBe(1);
        const settings = JSON.parse(readFileSync(settingsPath, "utf8"));
        const bypass = uid !== 0 || c.rootBypass;
        expect(settings.defaultMode).toBe(bypass ? "bypassPermissions" : undefined);
        expect(settings.skipDangerousModePermissionPrompt).toBe(bypass ? true : undefined);
        expect(settings.env.KEEP_TEST_SETTING).toBe("1");
        const launch: string[] = JSON.parse(readFileSync(launchPath, "utf8"));
        expect(launch[0]).toBe("new-session");
        expect(launch.at(-1)?.includes("--dangerously-skip-permissions")).toBe(bypass);
      });
    }
  }
});

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
  test("magpie refuses a missing account config before accepting project trust", () => {
    const d = dir();
    const cli = fileURLToPath(new URL("../src/cli.ts", import.meta.url));
    const env = {
      ...process.env,
      HOME: d,
      USERPROFILE: d,
      CLAUDE_CONFIG_DIR: "",
      CLAUDE_BRIDGE_TMUX_SOCKET: join(d, "unused.sock"),
      CLAUDE_BRIDGE_USER_CONTEXT_PATH: "",
    };
    const result = spawnSync(process.execPath, [cli, "-p", "test prompt", "--desplega-cwd", d], { env, encoding: "utf8", timeout: 10_000 });
    expect(result.error).toBeUndefined();
    expect(result.status).toBe(2);
    expect(result.stderr).toContain("account's own config folder (CLAUDE_CONFIG_DIR)");
    expect(existsSync(join(d, ".claude.json"))).toBe(false);
  });

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
