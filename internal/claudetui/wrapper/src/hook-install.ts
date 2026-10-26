import { copyFileSync, existsSync, mkdirSync, readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { writePrivateAtomic } from "./fs-atomic.ts";

export const JSON_SCHEMA_STOP_HOOK_ARG = "--desplega-internal-json-schema-stop-hook";
export const RUNTIME_HOOK_ARG = "--desplega-internal-runtime-hook";

type HookHandler = {
  type?: string;
  command?: string;
  [key: string]: unknown;
};

type HookGroup = {
  hooks?: HookHandler[];
  [key: string]: unknown;
};

type ClaudeSettings = {
  hooks?: Record<string, HookGroup[]>;
  [key: string]: unknown;
};

export function buildJsonSchemaStopHookCommand(argv = process.argv, bun = process.execPath): string {
  return buildInternalHookCommand(JSON_SCHEMA_STOP_HOOK_ARG, argv, bun);
}

export function buildRuntimeHookCommand(argv = process.argv, bun = process.execPath): string {
  return buildInternalHookCommand(RUNTIME_HOOK_ARG, argv, bun);
}

// A hook runs the very Bun this wrapper runs on (process.execPath), by its
// absolute path: neither PATH nor a `#!/usr/bin/env bun` need find one.
function buildInternalHookCommand(arg: string, argv = process.argv, bun = process.execPath): string {
  const entry = argv[1] ? resolve(process.cwd(), argv[1]) : resolve("src/cli.ts");
  return [bun, entry, arg].map(shellQuote).join(" ");
}

export function installRuntimeHooks(opts: {
  settingsPath: string;
  command: string;
}): { settingsPath: string; changed: boolean } {
  const settingsPath = opts.settingsPath;
  const settings = readSettings(settingsPath);
  settings.hooks ??= {};

  const before = JSON.stringify(settings.hooks);
  settings.hooks.Stop = upsertBridgeHook(settings.hooks.Stop ?? [], opts.command, RUNTIME_HOOK_ARG);
  settings.hooks.StopFailure = upsertBridgeHook(
    settings.hooks.StopFailure ?? [],
    opts.command,
    RUNTIME_HOOK_ARG
  );
  settings.hooks.MessageDisplay = upsertBridgeHook(
    settings.hooks.MessageDisplay ?? [],
    opts.command,
    RUNTIME_HOOK_ARG
  );
  settings.hooks.UserPromptSubmit = upsertBridgeHook(
    settings.hooks.UserPromptSubmit ?? [],
    opts.command,
    RUNTIME_HOOK_ARG
  );
  settings.hooks.SessionStart = upsertBridgeHook(
    settings.hooks.SessionStart ?? [],
    opts.command,
    RUNTIME_HOOK_ARG
  );
  settings.hooks.PreToolUse = upsertBridgeHook(
    settings.hooks.PreToolUse ?? [],
    opts.command,
    RUNTIME_HOOK_ARG
  );

  const changed = before !== JSON.stringify(settings.hooks);
  if (changed) writeSettings(settingsPath, settings);
  return { settingsPath, changed };
}

export function installUserContextHooks(opts: { settingsPath: string; command: string }): void {
  const settings = readSettings(opts.settingsPath);
  settings.hooks ??= {};
  for (const event of ["PostToolUse", "PostToolUseFailure"]) {
    settings.hooks[event] = upsertBridgeHook(settings.hooks[event] ?? [], opts.command, RUNTIME_HOOK_ARG);
  }
  writeSettings(opts.settingsPath, settings);
}

export function installJsonSchemaStopHook(opts: {
  settingsPath: string;
  command: string;
}): { settingsPath: string; changed: boolean } {
  const settingsPath = opts.settingsPath;
  const settings = readSettings(settingsPath);
  settings.hooks ??= {};
  const stopGroups = settings.hooks.Stop ?? [];
  const current = removeStaleJsonSchemaStopHookGroups(stopGroups, opts.command);
  const alreadyInstalled = current.some(group =>
    (group.hooks ?? []).some(handler => handler.command === opts.command)
  );

  settings.hooks.Stop = alreadyInstalled
    ? current
    : [
        ...current,
        {
          matcher: "*",
          hooks: [
            {
              type: "command",
              command: opts.command,
              timeout: 30,
            },
          ],
        },
      ];

  const changed =
    stopGroups.length !== settings.hooks.Stop.length ||
    !alreadyInstalled ||
    JSON.stringify(stopGroups) !== JSON.stringify(settings.hooks.Stop);
  if (changed) writeSettings(settingsPath, settings);
  return { settingsPath, changed };
}

function readSettings(settingsPath: string): ClaudeSettings {
  if (!existsSync(settingsPath)) return {};
  const raw = readFileSync(settingsPath, "utf8").trim();
  if (!raw) return {};
  return JSON.parse(raw) as ClaudeSettings;
}

function writeSettings(settingsPath: string, settings: ClaudeSettings): void {
  mkdirSync(dirname(settingsPath), { recursive: true });
  const backup = settingsPath + ".claude-bridge-backup";
  if (existsSync(settingsPath) && !existsSync(backup)) copyFileSync(settingsPath, backup);
  writePrivateAtomic(settingsPath, JSON.stringify(settings, null, 2) + "\n");
}

function removeStaleJsonSchemaStopHookGroups(groups: HookGroup[], keepCommand?: string): HookGroup[] {
  return groups
    .map(group => ({
      ...group,
      hooks: (group.hooks ?? []).filter(
        handler =>
          !handler.command?.includes(JSON_SCHEMA_STOP_HOOK_ARG) ||
          (keepCommand !== undefined && handler.command === keepCommand && group.matcher === "*")
      ),
    }))
    .filter(group => (group.hooks ?? []).length > 0);
}

function upsertBridgeHook(groups: HookGroup[], command: string, marker: string): HookGroup[] {
  const current = groups
    .map(group => ({
      ...group,
      hooks: (group.hooks ?? []).filter(
        handler =>
          !handler.command?.includes(marker) ||
          (handler.command === command && group.matcher === "*")
      ),
    }))
    .filter(group => (group.hooks ?? []).length > 0);
  const alreadyInstalled = current.some(group =>
    (group.hooks ?? []).some(handler => handler.command === command)
  );
  if (alreadyInstalled) return current;
  return [
    ...current,
    {
      matcher: "*",
      hooks: [
        {
          type: "command",
          command,
          timeout: 30,
        },
      ],
    },
  ];
}

function shellQuote(value: string): string {
  return `'${value.replace(/'/g, `'\\''`)}'`;
}
