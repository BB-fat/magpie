import { chmodSync, readFileSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { resolveClaudeNodeLaunch, type ClaudeNodeLaunch } from "./claude-node-entry.ts";

/** Stay under Linux MAX_ARG_STRLEN (~128 KiB) with shell/env overhead. */
export const NATIVE_INLINE_EXTRA_BODY_MAX = 120 * 1024;

export type ExtraBodyLaunchPlan =
  | { kind: "node-preload"; extraBodyPreloadPath: string; claudeNodeLaunch: ClaudeNodeLaunch }
  | { kind: "native-inline"; inlineExtraBody: string };

export function nativeExtraBodyTooLargeMessage(byteLength: number): string {
  return (
    `[claude-bridge] Claude Code is a native binary (not a Node launcher); ` +
    `safeguards are ${byteLength} bytes but inline CLAUDE_CODE_EXTRA_BODY is limited to ` +
    `${NATIVE_INLINE_EXTRA_BODY_MAX} bytes on this install. Shorten classifier context or use a Node-based Claude install.\n`
  );
}

/** How first-turn safeguards reach Claude Code inside tmux. */
export function planExtraBodyLaunch(opts: {
  extraBodyFile: string;
  claudePath: string;
  runDir: string;
  preloadSourcePath: string;
}): ExtraBodyLaunchPlan | { error: string } {
  const claudeNodeLaunch = resolveClaudeNodeLaunch(opts.claudePath);
  if (claudeNodeLaunch) {
    const extraBodyPreloadPath = join(opts.runDir, "extra-body-preload.cjs");
    writeFileSync(extraBodyPreloadPath, readFileSync(opts.preloadSourcePath));
    chmodSync(extraBodyPreloadPath, 0o600);
    return { kind: "node-preload", extraBodyPreloadPath, claudeNodeLaunch };
  }

  let inlineExtraBody: string;
  try {
    inlineExtraBody = readFileSync(opts.extraBodyFile, "utf8");
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return { error: `[claude-bridge] CLAUDE_CODE_EXTRA_BODY_FILE read failed: ${message}\n` };
  }
  if (inlineExtraBody.length > NATIVE_INLINE_EXTRA_BODY_MAX) {
    return { error: nativeExtraBodyTooLargeMessage(inlineExtraBody.length) };
  }
  return { kind: "native-inline", inlineExtraBody };
}
