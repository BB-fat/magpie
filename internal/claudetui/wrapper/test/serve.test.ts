// A run kept for more than one turn (--desplega-serve): its control lines,
// a Stop that names its turn, and a Claude Code that does nothing of its
// own while it waits.
import { afterAll, describe, expect, test } from "bun:test";
import { spawnSync } from "node:child_process";
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
import { parseServeControl, stopEventIsTurns } from "../src/serve-protocol.ts";
import { recordRuntimeHook, SERVE_TURN_FILE } from "../src/stop-hook.ts";
import { IDLE_QUIET_SETTINGS, writeWorkdirSettings } from "../src/preaccept.ts";

const testRoot = mkdtempSync(join(tmpdir(), "magpie-wrapper-serve-"));
afterAll(() => rmSync(testRoot, { recursive: true, force: true }));
const dir = () => mkdtempSync(join(testRoot, "case-"));

describe("control lines", () => {
  test("a submit names its turn and carries its prompt", () => {
    expect(parseServeControl(JSON.stringify({ op: "submit", segment: "seg-1", input: "hi\nthere" })))
      .toEqual({ op: "submit", segment: "seg-1", input: "hi\nthere" });
    expect(parseServeControl('{"op":"shutdown"}')).toEqual({ op: "shutdown" });
  });

  test("a turn name that could leave the run folder, an empty prompt or another op is refused", () => {
    for (const line of [
      JSON.stringify({ op: "submit", segment: "../stop-event", input: "x" }),
      JSON.stringify({ op: "submit", segment: "a/b", input: "x" }),
      JSON.stringify({ op: "submit", segment: "x".repeat(65), input: "x" }),
      JSON.stringify({ op: "submit", segment: "seg", input: "" }),
      JSON.stringify({ op: "submit", segment: 7, input: "x" }),
      JSON.stringify({ op: "cancel", segment: "seg" }),
      "[1]",
      "not json",
    ]) {
      expect("error" in parseServeControl(line)).toBe(true);
    }
  });
});

describe("a Stop names its turn", () => {
  const stop = (runDir: string, turn?: string) => {
    if (turn !== undefined) writeFileSync(join(runDir, SERVE_TURN_FILE), turn + "\n");
    recordRuntimeHook(JSON.stringify({ hook_event_name: "Stop", transcript_path: "/t.jsonl", last_assistant_message: "PONG" }), {
      CLAUDE_BRIDGE_RUNTIME_HOOK: "1",
      CLAUDE_BRIDGE_RUN_DIR: runDir,
    });
    return JSON.parse(readFileSync(join(runDir, "stop-event.json"), "utf8"));
  };

  test("the turn in flight is written into the event, and only it completes that turn", () => {
    const d = dir();
    const first = stop(d, "seg-1");
    expect(first.bridge_turn).toBe("seg-1");
    expect(first.last_assistant_message).toBe("PONG");
    expect(stopEventIsTurns(first, "seg-1")).toBe(true);
    // the last turn's event, still on disk as the next turn starts
    writeFileSync(join(d, SERVE_TURN_FILE), "seg-2\n");
    expect(stopEventIsTurns(first, "seg-2")).toBe(false);
    expect(stopEventIsTurns(stop(d), "seg-2")).toBe(true);
  });

  test("a one-turn run writes the event as it came, which completes no serve turn", () => {
    const d = dir();
    const event = stop(d);
    expect(event.bridge_turn).toBeUndefined();
    expect(stopEventIsTurns(event, "seg-1")).toBe(false);
  });
});

describe("waiting quietly", () => {
  test("suggestions, the away recap and idle compaction are off, beside the user's own settings", () => {
    const d = dir();
    const settings = join(d, ".claude", "settings.local.json");
    writeWorkdirSettings(d, { bypassPermissions: false });
    writeFileSync(settings, JSON.stringify({ ...JSON.parse(readFileSync(settings, "utf8")), model: "kept", promptSuggestionEnabled: true }));
    writeWorkdirSettings(d, { bypassPermissions: false });
    const written = JSON.parse(readFileSync(settings, "utf8"));
    expect(written).toMatchObject(IDLE_QUIET_SETTINGS);
    expect(written.model).toBe("kept");
  });
});

test("a serve run takes no prompt argument", () => {
  const d = dir();
  const cli = fileURLToPath(new URL("../src/cli.ts", import.meta.url));
  const env = { ...process.env, HOME: d, CLAUDE_CONFIG_DIR: d };
  const result = spawnSync(process.execPath, [cli, "-p", "a prompt", "--desplega-serve", "--desplega-cwd", d], { env, encoding: "utf8", timeout: 10_000 });
  expect(result.status).toBe(2);
  expect(result.stderr).toContain("control lines on stdin");
});
