import { closeSync, fsyncSync, openSync, renameSync, unlinkSync, writeSync } from "node:fs";
import { dirname, isAbsolute, join } from "node:path";
import { createHash, randomUUID } from "node:crypto";

// A probe using a different tmux socket namespace could mistakenly report
// "no server" while the original detached native Claude is still alive.
// Persist no private path, only a digest of the effective socket settings.
function socketEnvironmentDigest(): string {
  return createHash("sha256")
    .update(`${process.env.TMUX_TMPDIR ?? ""}\0${process.env.TMUX ?? ""}`)
    .digest("hex");
}

// The Go Runtime holds the per-outer lease before launching this wrapper.
// Persist the exact detached tmux identity BEFORE tmux new-session so that
// an abruptly killed wrapper cannot leave an unidentified live Claude pane.
// The marker is private and is never removed by normal shutdown: the next
// Runtime must query tmux to prove the old pane has actually stopped.
export function writeNativeTmuxGuard(path: string, session: string, socket = ""): void {
  if (!isAbsolute(path) || !/^claude-bridge-[a-fA-F0-9]{8}$/.test(session)) {
    throw new Error("invalid native tmux ownership marker path/session");
  }
  const folder = dirname(path);
  const tmp = join(folder, `.native-tmux-${randomUUID()}.tmp`);
  let fd: number | undefined;
  try {
    fd = openSync(tmp, "wx", 0o600);
    const payload = Buffer.from(JSON.stringify({ version: 2, session,
      socket_env_sha256: socketEnvironmentDigest(), ...(socket ? { socket } : {}) }) + "\n", "utf8");
    const n = writeSync(fd, payload);
    if (n !== payload.length) throw new Error("incomplete native tmux marker write");
    fsyncSync(fd);
    closeSync(fd);
    fd = undefined;
    renameSync(tmp, path);
    // On Unix, fsync the containing directory so that a sudden Runtime
    // crash cannot roll back the rename while leaving the tmux pane alive.
    const directory = openSync(folder, "r");
    try { fsyncSync(directory); } finally { closeSync(directory); }
  } catch (error) {
    if (fd !== undefined) closeSync(fd);
    try { unlinkSync(tmp); } catch {}
    throw error;
  }
}
