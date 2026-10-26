import { closeSync, fsyncSync, openSync, renameSync, unlinkSync, writeSync } from "node:fs";
import { dirname, join } from "node:path";
import { randomUUID } from "node:crypto";

// writePrivateAtomic replaces path with data, readable by this user alone,
// through a file beside it flushed to the disk first: a reader (Claude
// Code itself, at startup) sees the old whole file or the new one, never
// one half written.
export function writePrivateAtomic(path: string, data: string): void {
  const tmp = join(dirname(path), `.${randomUUID()}.tmp`);
  let fd: number | undefined;
  try {
    fd = openSync(tmp, "wx", 0o600);
    const buf = Buffer.from(data, "utf8");
    let off = 0;
    while (off < buf.length) off += writeSync(fd, buf, off);
    fsyncSync(fd);
    closeSync(fd);
    fd = undefined;
    renameSync(tmp, path);
  } catch (error) {
    if (fd !== undefined) closeSync(fd);
    try { unlinkSync(tmp); } catch {}
    throw error;
  }
}
