import { afterAll, describe, expect, test } from "bun:test";
import { mkdtempSync, readFileSync, rmSync, writeFileSync, mkdirSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";
import { buildClaudeLaunchCommand } from "../src/launch-command.ts";
import { resolveClaudeNodeLaunch } from "../src/claude-node-entry.ts";

const root = mkdtempSync(join(tmpdir(), "launch-extra-body-"));
afterAll(() => rmSync(root, { recursive: true, force: true }));

describe("large EXTRA_BODY launch", () => {
  test("resolveClaudeNodeLaunch finds node entry", () => {
    const d = mkdtempSync(join(root, "bin-"));
    writeFileSync(
      join(d, "claude"),
      "#!/usr/bin/env node\nimport './entry.js';\n",
      { mode: 0o755 }
    );
    writeFileSync(join(d, "entry.js"), "export {};\n");
    const launch = resolveClaudeNodeLaunch(join(d, "claude"));
    expect(launch?.interpreter).toBe("node");
    expect(launch?.entryScript).toEndWith("entry.js");
  });

  test("node -r preload loads >128KiB without execve E2BIG", () => {
    const d = mkdtempSync(join(root, "exec-"));
    const bodyFile = join(d, "extra-body.json");
    const body = JSON.stringify({ safeguards: ["x".repeat(140 * 1024)] });
    writeFileSync(bodyFile, body);
    const preload = join(d, "extra-body-preload.cjs");
    writeFileSync(
      preload,
      readFileSync(new URL("../src/extra-body-preload.cjs", import.meta.url), "utf8")
    );
    mkdirSync(join(d, "bin"), { recursive: true });
    writeFileSync(
      join(d, "bin", "claude"),
      "#!/usr/bin/env node\nrequire('../probe.js');\n",
      { mode: 0o755 }
    );
    writeFileSync(
      join(d, "probe.js"),
      `const v = process.env.CLAUDE_CODE_EXTRA_BODY || "";
require("fs").writeFileSync(process.env.PROBE_OUT, String(v.length));
`
    );
    const launch = resolveClaudeNodeLaunch(join(d, "bin", "claude"));
    expect(launch).not.toBeNull();
    const cmd = buildClaudeLaunchCommand({
      claudePath: join(d, "bin", "claude"),
      claudeArgs: [],
      unsetEnvArgs: ["-u", "CLAUDE_CODE_EXTRA_BODY"],
      envArgs: [`CLAUDE_CODE_EXTRA_BODY_FILE=${bodyFile}`, `PROBE_OUT=${join(d, "probe-out.txt")}`],
      exitStatusPath: join(d, "exit-status"),
      holdMs: 1000,
      extraBodyPreloadPath: preload,
      claudeNodeLaunch: launch!,
    });
    const prev = process.env.CLAUDE_CODE_EXTRA_BODY_FILE;
    process.env.CLAUDE_CODE_EXTRA_BODY_FILE = bodyFile;
    const res = spawnSync("/bin/sh", ["-lc", cmd], { encoding: "utf8", timeout: 30_000 });
    if (prev === undefined) delete process.env.CLAUDE_CODE_EXTRA_BODY_FILE;
    else process.env.CLAUDE_CODE_EXTRA_BODY_FILE = prev;
    expect(res.error).toBeUndefined();
    expect(res.status).toBe(0);
    const n = Number(readFileSync(join(d, "probe-out.txt"), "utf8"));
    expect(n).toBe(body.length);
    expect(n).toBeGreaterThan(128 * 1024);
  });
});
