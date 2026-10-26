"use strict";
// Loaded with node -r before Claude Code's entry module. Exec env keeps only
// CLAUDE_CODE_EXTRA_BODY_FILE (short path); this preload reads the file into
// process.env in memory without passing the JSON through execve (MAX_ARG_STRLEN).
const fs = require("fs");
const path = process.env.CLAUDE_CODE_EXTRA_BODY_FILE;
if (path) {
  try {
    process.env.CLAUDE_CODE_EXTRA_BODY = fs.readFileSync(path, "utf8");
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    process.stderr.write(`[claude-bridge] CLAUDE_CODE_EXTRA_BODY_FILE read failed: ${message}\n`);
    process.exit(1);
  }
}
