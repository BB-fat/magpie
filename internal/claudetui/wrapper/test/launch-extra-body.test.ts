import { afterAll, describe, expect, test } from "bun:test";
import { mkdtempSync, readFileSync, rmSync, writeFileSync, mkdirSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";
import { buildClaudeLaunchCommand } from "../src/launch-command.ts";
import { resolveClaudeNodeLaunch } from "../src/claude-node-entry.ts";
import {
  NATIVE_INLINE_EXTRA_BODY_MAX,
  nativeExtraBodyTooLargeMessage,
  planExtraBodyLaunch,
} from "../src/extra-body-launch.ts";

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

  test("native claude gets inline EXTRA_BODY when small", () => {
    const d = mkdtempSync(join(root, "native-inline-"));
    const bodyFile = join(d, "extra-body.json");
    const body = JSON.stringify({ safeguards: [{ classifier_context: "tiny" }] });
    writeFileSync(bodyFile, body);
    mkdirSync(join(d, "bin"), { recursive: true });
    const claude = join(d, "bin", "claude");
    writeFileSync(claude, Buffer.from([0x7f, 0x45, 0x4c, 0x46, 0]), { mode: 0o755 });
    writeFileSync(
      join(d, "probe.sh"),
      `#!/bin/sh
printf '%s' "$CLAUDE_CODE_EXTRA_BODY" > "$PROBE_OUT"
`,
      { mode: 0o755 }
    );
    const planned = planExtraBodyLaunch({
      extraBodyFile: bodyFile,
      claudePath: claude,
      runDir: d,
      preloadSourcePath: new URL("../src/extra-body-preload.cjs", import.meta.url).pathname,
    });
    expect(planned).toEqual({ kind: "native-inline", inlineExtraBody: body });
    if (!("kind" in planned) || planned.kind !== "native-inline") throw new Error("unreachable");
    const cmd = buildClaudeLaunchCommand({
      claudePath: join(d, "probe.sh"),
      claudeArgs: [],
      unsetEnvArgs: ["-u", "CLAUDE_CODE_EXTRA_BODY_FILE"],
      envArgs: [`CLAUDE_CODE_EXTRA_BODY=${planned.inlineExtraBody}`, `PROBE_OUT=${join(d, "probe-out.txt")}`],
      exitStatusPath: join(d, "exit-status"),
      holdMs: 1000,
    });
    const res = spawnSync("/bin/sh", ["-lc", cmd], {
      encoding: "utf8",
      timeout: 30_000,
      env: { ...process.env, CLAUDE_CODE_EXTRA_BODY_FILE: bodyFile },
    });
    expect(res.error).toBeUndefined();
    expect(res.status).toBe(0);
    expect(readFileSync(join(d, "probe-out.txt"), "utf8")).toBe(body);
  });

  test("native claude rejects safeguards above inline limit", () => {
    const d = mkdtempSync(join(root, "native-large-"));
    const bodyFile = join(d, "extra-body.json");
    writeFileSync(bodyFile, JSON.stringify({ safeguards: ["x".repeat(NATIVE_INLINE_EXTRA_BODY_MAX + 1)] }));
    const bodyLen = readFileSync(bodyFile, "utf8").length;
    const claude = join(d, "claude");
    writeFileSync(claude, Buffer.from([0x7f, 0x45, 0x4c, 0x46, 0]), { mode: 0o755 });
    const planned = planExtraBodyLaunch({
      extraBodyFile: bodyFile,
      claudePath: claude,
      runDir: d,
      preloadSourcePath: new URL("../src/extra-body-preload.cjs", import.meta.url).pathname,
    });
    expect("error" in planned).toBe(true);
    if (!("error" in planned)) throw new Error("unreachable");
    expect(planned.error).toBe(nativeExtraBodyTooLargeMessage(bodyLen));
  });
});
