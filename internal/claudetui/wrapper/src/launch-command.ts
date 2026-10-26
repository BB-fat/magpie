export type ClaudeLaunchCommandOptions = {
  claudePath: string;
  claudeArgs: string[];
  unsetEnvArgs: string[];
  envArgs: string[];
  exitStatusPath: string;
  holdMs: number;
  /** When set, launch Claude Code via node -r preload (avoids MAX_ARG_STRLEN on exec). */
  extraBodyPreloadPath?: string;
  claudeNodeLaunch?: { interpreter: string; entryScript: string };
};

export function interactiveMcpBackgroundEnvArgs(): string[] {
  // tmux servers can predate this wrapper and retain their original environment.
  return [
    "CLAUDE_CODE_MCP_AUTO_BACKGROUND_MS=0",
    // This private MCP server is the only source of client tools. Join its
    // alwaysLoad listing before the first prompt instead of racing startup.
    "CLAUDE_CODE_MCP_ALWAYS_LOAD_JOIN_BEFORE_INPUT=1",
    "CLAUDE_CODE_MCP_STARTUP_WAIT_MS=30000",
  ];
}

/**
 * The Go Runtime decides which side owns automatic compaction. A tmux server
 * can outlive that Runtime and retain a different (or missing) environment.
 * Pass this decision explicitly to the *native Claude* child, not just to the
 * Bun wrapper started by Go.
 */
export function nativeAutoCompactionLaunchEnv(env: NodeJS.ProcessEnv): {
  unsetEnvArgs: string[];
  envArgs: string[];
} {
  if (env.DISABLE_AUTO_COMPACT === "1") {
    return { unsetEnvArgs: [], envArgs: ["DISABLE_AUTO_COMPACT=1"] };
  }
  // Standalone one-shot calls must not inherit a stale tmux server's flag.
  return { unsetEnvArgs: ["-u", "DISABLE_AUTO_COMPACT"], envArgs: [] };
}

export function shellQuote(value: string): string {
  if (value.length === 0) return "''";
  return `'${value.replace(/'/g, `'\\''`)}'`;
}

function claudeLaunchArgv(opts: ClaudeLaunchCommandOptions): string[] {
  if (opts.extraBodyPreloadPath && opts.claudeNodeLaunch) {
    const { interpreter, entryScript } = opts.claudeNodeLaunch;
    return [interpreter, "-r", opts.extraBodyPreloadPath, entryScript, ...opts.claudeArgs];
  }
  return [opts.claudePath, ...opts.claudeArgs];
}

export function buildClaudeLaunchCommand(opts: ClaudeLaunchCommandOptions): string {
  const envCommand = ["env", ...opts.unsetEnvArgs, ...opts.envArgs, ...claudeLaunchArgv(opts)]
    .map(shellQuote)
    .join(" ");
  const holdSeconds = Math.max(1, Math.ceil(opts.holdMs / 1000));
  const script = [
    envCommand,
    "status=$?",
    `printf '%s\\n' "$status" > ${shellQuote(opts.exitStatusPath)}`,
    `printf '\\n[claude-bridge] claude exited with status %s\\n' "$status" >&2`,
    `sleep ${holdSeconds}`,
    `exit "$status"`,
  ].join("\n");

  return `/bin/sh -lc ${shellQuote(script)}`;
}
