#!/usr/bin/env bun
// @bun
var __defProp = Object.defineProperty;
var __returnValue = (v) => v;
function __exportSetter(name, newValue) {
  this[name] = __returnValue.bind(null, newValue);
}
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, {
      get: all[name],
      enumerable: true,
      configurable: true,
      set: __exportSetter.bind(all, name)
    });
};

// node_modules/zod/v4/core/regexes.js
var exports_regexes = {};
__export(exports_regexes, {
  anyString: () => anyString,
  base64: () => base64,
  base64url: () => base64url,
  bigint: () => bigint,
  boolean: () => boolean,
  browserEmail: () => browserEmail,
  cidrv4: () => cidrv4,
  cidrv6: () => cidrv6,
  creditCard: () => creditCard,
  cuid: () => cuid,
  cuid2: () => cuid2,
  currencyCode: () => currencyCode,
  date: () => date,
  datetime: () => datetime,
  domain: () => domain,
  duration: () => duration,
  e164: () => e164,
  email: () => email,
  emoji: () => emoji,
  extendedDuration: () => extendedDuration,
  guid: () => guid,
  hex: () => hex,
  hostname: () => hostname,
  html5Email: () => html5Email,
  httpProtocol: () => httpProtocol,
  iban: () => iban,
  idnEmail: () => idnEmail,
  integer: () => integer,
  ipv4: () => ipv4,
  ipv6: () => ipv6,
  ksuid: () => ksuid,
  lowercase: () => lowercase,
  mac: () => mac,
  md5_base64: () => md5_base64,
  md5_base64url: () => md5_base64url,
  md5_hex: () => md5_hex,
  nanoid: () => nanoid,
  nanoidOfLength: () => nanoidOfLength,
  null: () => _null,
  number: () => number,
  rfc5322Email: () => rfc5322Email,
  sha1_base64: () => sha1_base64,
  sha1_base64url: () => sha1_base64url,
  sha1_hex: () => sha1_hex,
  sha256_base64: () => sha256_base64,
  sha256_base64url: () => sha256_base64url,
  sha256_hex: () => sha256_hex,
  sha384_base64: () => sha384_base64,
  sha384_base64url: () => sha384_base64url,
  sha384_hex: () => sha384_hex,
  sha512_base64: () => sha512_base64,
  sha512_base64url: () => sha512_base64url,
  sha512_hex: () => sha512_hex,
  string: () => string,
  time: () => time2,
  ulid: () => ulid,
  undefined: () => _undefined,
  unicodeEmail: () => unicodeEmail,
  uppercase: () => uppercase,
  uuid: () => uuid,
  uuid4: () => uuid4,
  uuid6: () => uuid6,
  uuid7: () => uuid7,
  xid: () => xid
});

// node_modules/zod/v4/classic/schemas.js
var exports_schemas2 = {};
__export(exports_schemas2, {
  ZodAny: () => ZodAny,
  ZodArray: () => ZodArray,
  ZodBase64: () => ZodBase64,
  ZodBase64URL: () => ZodBase64URL,
  ZodBigInt: () => ZodBigInt,
  ZodBigIntFormat: () => ZodBigIntFormat,
  ZodBoolean: () => ZodBoolean,
  ZodCIDRv4: () => ZodCIDRv4,
  ZodCIDRv6: () => ZodCIDRv6,
  ZodCUID: () => ZodCUID,
  ZodCUID2: () => ZodCUID2,
  ZodCatch: () => ZodCatch,
  ZodCodec: () => ZodCodec,
  ZodCreditCard: () => ZodCreditCard,
  ZodCustom: () => ZodCustom,
  ZodCustomStringFormat: () => ZodCustomStringFormat,
  ZodDate: () => ZodDate,
  ZodDefault: () => ZodDefault,
  ZodDiscriminatedUnion: () => ZodDiscriminatedUnion,
  ZodE164: () => ZodE164,
  ZodEmail: () => ZodEmail,
  ZodEmoji: () => ZodEmoji,
  ZodEnum: () => ZodEnum,
  ZodExactOptional: () => ZodExactOptional,
  ZodFile: () => ZodFile,
  ZodFunction: () => ZodFunction,
  ZodGUID: () => ZodGUID,
  ZodIBAN: () => ZodIBAN,
  ZodIPv4: () => ZodIPv4,
  ZodIPv6: () => ZodIPv6,
  ZodISODate: () => ZodISODate,
  ZodISODateTime: () => ZodISODateTime,
  ZodISODuration: () => ZodISODuration,
  ZodISOTime: () => ZodISOTime,
  ZodInstanceOf: () => ZodInstanceOf,
  ZodIntersection: () => ZodIntersection,
  ZodJWT: () => ZodJWT,
  ZodKSUID: () => ZodKSUID,
  ZodLazy: () => ZodLazy,
  ZodLiteral: () => ZodLiteral,
  ZodMAC: () => ZodMAC,
  ZodMap: () => ZodMap,
  ZodNaN: () => ZodNaN,
  ZodNanoID: () => ZodNanoID,
  ZodNever: () => ZodNever,
  ZodNonOptional: () => ZodNonOptional,
  ZodNull: () => ZodNull,
  ZodNullable: () => ZodNullable,
  ZodNumber: () => ZodNumber,
  ZodNumberFormat: () => ZodNumberFormat,
  ZodObject: () => ZodObject,
  ZodOptional: () => ZodOptional,
  ZodPipe: () => ZodPipe,
  ZodPrefault: () => ZodPrefault,
  ZodPreprocess: () => ZodPreprocess,
  ZodPromise: () => ZodPromise,
  ZodReadonly: () => ZodReadonly,
  ZodRecord: () => ZodRecord,
  ZodSet: () => ZodSet,
  ZodString: () => ZodString,
  ZodStringFormat: () => ZodStringFormat,
  ZodSuccess: () => ZodSuccess,
  ZodSymbol: () => ZodSymbol,
  ZodTemplateLiteral: () => ZodTemplateLiteral,
  ZodTransform: () => ZodTransform,
  ZodTuple: () => ZodTuple,
  ZodType: () => ZodType,
  ZodULID: () => ZodULID,
  ZodURL: () => ZodURL,
  ZodUUID: () => ZodUUID,
  ZodUndefined: () => ZodUndefined,
  ZodUnion: () => ZodUnion,
  ZodUnknown: () => ZodUnknown,
  ZodVoid: () => ZodVoid,
  ZodXID: () => ZodXID,
  ZodXor: () => ZodXor,
  _ZodString: () => _ZodString,
  _default: () => _default,
  _function: () => _function,
  any: () => any,
  array: () => array,
  base64: () => base642,
  base64url: () => base64url2,
  bigint: () => bigint2,
  boolean: () => boolean2,
  catch: () => _catch,
  check: () => check,
  cidrv4: () => cidrv42,
  cidrv6: () => cidrv62,
  codec: () => codec,
  creditCard: () => creditCard2,
  cuid: () => cuid3,
  cuid2: () => cuid22,
  currencyCode: () => currencyCode2,
  custom: () => custom,
  date: () => date2,
  describe: () => describe2,
  discriminatedUnion: () => discriminatedUnion,
  e164: () => e1642,
  email: () => email2,
  emoji: () => emoji2,
  enum: () => _enum,
  exactOptional: () => exactOptional,
  file: () => file,
  float32: () => float32,
  float64: () => float64,
  function: () => _function,
  guid: () => guid2,
  hash: () => hash,
  hex: () => hex2,
  hostname: () => hostname2,
  httpUrl: () => httpUrl,
  iban: () => iban2,
  instanceof: () => _instanceof,
  int: () => int,
  int32: () => int32,
  int64: () => int64,
  intersection: () => intersection,
  invertCodec: () => invertCodec,
  ipv4: () => ipv42,
  ipv6: () => ipv62,
  json: () => json,
  jwt: () => jwt,
  keyof: () => keyof,
  ksuid: () => ksuid2,
  lazy: () => lazy,
  literal: () => literal,
  looseObject: () => looseObject,
  looseRecord: () => looseRecord,
  mac: () => mac2,
  map: () => map,
  meta: () => meta2,
  nan: () => nan,
  nanoid: () => nanoid2,
  nativeEnum: () => nativeEnum,
  never: () => never,
  nonoptional: () => nonoptional,
  null: () => _null3,
  nullable: () => nullable,
  nullish: () => nullish2,
  number: () => number2,
  object: () => object,
  optional: () => optional,
  partialRecord: () => partialRecord,
  pipe: () => pipe,
  prefault: () => prefault,
  preprocess: () => preprocess,
  promise: () => promise,
  readonly: () => readonly,
  record: () => record,
  refine: () => refine,
  set: () => set,
  strictObject: () => strictObject,
  string: () => string2,
  stringFormat: () => stringFormat,
  stringbool: () => stringbool,
  success: () => success,
  superRefine: () => superRefine,
  symbol: () => symbol,
  templateLiteral: () => templateLiteral,
  transform: () => transform,
  tuple: () => tuple,
  uint32: () => uint32,
  uint64: () => uint64,
  ulid: () => ulid2,
  undefined: () => _undefined3,
  union: () => union,
  unknown: () => unknown,
  url: () => url,
  uuid: () => uuid2,
  uuidv4: () => uuidv4,
  uuidv6: () => uuidv6,
  uuidv7: () => uuidv7,
  void: () => _void2,
  xid: () => xid2,
  xor: () => xor
});

// node_modules/zod/v4/classic/checks.js
var exports_checks2 = {};
__export(exports_checks2, {
  endsWith: () => _endsWith,
  gt: () => _gt,
  gte: () => _gte,
  includes: () => _includes,
  length: () => _length,
  lowercase: () => _lowercase,
  lt: () => _lt,
  lte: () => _lte,
  maxLength: () => _maxLength,
  maxSize: () => _maxSize,
  mime: () => _mime,
  minLength: () => _minLength,
  minSize: () => _minSize,
  multipleOf: () => _multipleOf,
  negative: () => _negative,
  nonnegative: () => _nonnegative,
  nonpositive: () => _nonpositive,
  normalize: () => _normalize,
  overwrite: () => _overwrite,
  positive: () => _positive,
  properties: () => _properties,
  property: () => _property,
  regex: () => _regex,
  size: () => _size,
  slugify: () => _slugify,
  startsWith: () => _startsWith,
  toLowerCase: () => _toLowerCase,
  toUpperCase: () => _toUpperCase,
  trim: () => _trim,
  uppercase: () => _uppercase
});

// node_modules/zod/v4/classic/iso.js
var exports_iso = {};
__export(exports_iso, {
  ZodISODate: () => ZodISODate,
  ZodISODateTime: () => ZodISODateTime,
  ZodISODuration: () => ZodISODuration,
  ZodISOTime: () => ZodISOTime,
  date: () => date3,
  datetime: () => datetime2,
  duration: () => duration2,
  time: () => time3
});

// node_modules/zod/v4/classic/coerce.js
var exports_coerce = {};
__export(exports_coerce, {
  bigint: () => bigint3,
  boolean: () => boolean3,
  date: () => date4,
  number: () => number3,
  string: () => string3
});

// src/cli.ts
import { spawn, spawnSync } from "child_process";
import { createInterface } from "readline";
import { chmodSync, existsSync as existsSync5, mkdirSync as mkdirSync4, readFileSync as readFileSync7, statSync as statSync2, writeFileSync as writeFileSync3 } from "fs";
import { join as join6, resolve as resolve5, dirname as dirname6 } from "path";
import { fileURLToPath } from "url";
import { randomUUID as randomUUID3 } from "crypto";

// src/preaccept.ts
import { readFileSync, existsSync, mkdirSync, copyFileSync } from "fs";
import { homedir } from "os";
import { dirname as dirname2, join as join2, resolve } from "path";

// src/fs-atomic.ts
import { closeSync, fsyncSync, openSync, renameSync, unlinkSync, writeSync } from "fs";
import { dirname, join } from "path";
import { randomUUID } from "crypto";
function writePrivateAtomic(path, data) {
  const tmp = join(dirname(path), `.${randomUUID()}.tmp`);
  let fd;
  try {
    fd = openSync(tmp, "wx", 384);
    const buf = Buffer.from(data, "utf8");
    let off = 0;
    while (off < buf.length)
      off += writeSync(fd, buf, off);
    fsyncSync(fd);
    closeSync(fd);
    fd = undefined;
    renameSync(tmp, path);
  } catch (error) {
    if (fd !== undefined)
      closeSync(fd);
    try {
      unlinkSync(tmp);
    } catch {}
    throw error;
  }
}

// src/preaccept.ts
function trustPromptKeys(pane) {
  if (/\u276F\s*No,\s*exit/i.test(pane))
    return ["Down", "Enter"];
  if (/\u276F\s*Yes,\s*I trust/i.test(pane))
    return ["Enter"];
  return [];
}
function claudePaneReady(pane) {
  const trust = /(Quick safety check|Is this a project you created)/i.test(pane) && /^\s*\u276F\s*(?:No,\s*exit|Yes,\s*I trust this folder)\s*$/im.test(pane);
  const selector = /^\s*\u276F\s*\S/m.test(pane) && /(Choose the text style that looks best with your terminal|Do you want to use this API key\?)/i.test(pane);
  if (trust || selector)
    return false;
  return /(^|\n)\s*(?:\u276F|>)\s*(?!\d+\.\s)[^\n]*(?:\n|$)/.test(pane) || /-- INSERT --/.test(pane);
}
function claudeGlobalConfigPath() {
  const configDir = process.env.CLAUDE_CONFIG_DIR;
  if (!configDir)
    return join2(homedir(), ".claude.json");
  return join2(resolve(configDir), ".claude.json");
}
function preAcceptProject(opts) {
  const claudeJsonPath = claudeGlobalConfigPath();
  mkdirSync(dirname2(claudeJsonPath), { recursive: true });
  if (!existsSync(claudeJsonPath)) {
    writePrivateAtomic(claudeJsonPath, JSON.stringify({ projects: {} }, null, 2));
  }
  const backup = claudeJsonPath + ".claude-bridge-backup";
  if (!existsSync(backup))
    copyFileSync(claudeJsonPath, backup);
  const raw = readFileSync(claudeJsonPath, "utf8");
  const obj = JSON.parse(raw);
  obj.hasCompletedOnboarding = true;
  obj.lastOnboardingVersion ??= "2.1.0";
  obj.projects ??= {};
  const entry = obj.projects[opts.workdir] ??= {};
  entry.hasTrustDialogAccepted = true;
  entry.hasCompletedProjectOnboarding = true;
  const approved = new Set([
    ...entry.approvedMcprcServers ?? [],
    ...opts.mcpServerNames
  ]);
  entry.approvedMcprcServers = [...approved];
  writePrivateAtomic(claudeJsonPath, JSON.stringify(obj, null, 2));
}
function writeWorkdirSettings(workdir, opts = { bypassPermissions: true }) {
  const dir = join2(workdir, ".claude");
  mkdirSync(dir, { recursive: true });
  const file = join2(dir, "settings.local.json");
  const existing = existsSync(file) ? JSON.parse(readFileSync(file, "utf8")) : {};
  const settings = { ...existing };
  if (opts.bypassPermissions) {
    settings.skipDangerousModePermissionPrompt = true;
    settings.defaultMode = "bypassPermissions";
  } else {
    if (settings.defaultMode === "bypassPermissions")
      delete settings.defaultMode;
    if (settings.skipDangerousModePermissionPrompt === true) {
      delete settings.skipDangerousModePermissionPrompt;
    }
  }
  writePrivateAtomic(file, JSON.stringify(settings, null, 2) + `
`);
}

// src/auth-env.ts
var DEFAULT_AUTH_ENV_NAMES = [
  "HOME",
  "CLAUDE_CONFIG_DIR",
  "CLAUDE_CODE_OAUTH_TOKEN"
];
var LOCAL_AUTH_ENV_NAMES = [
  "ANTHROPIC_API_KEY",
  "ANTHROPIC_AUTH_TOKEN",
  "ANTHROPIC_BASE_URL",
  "ANTHROPIC_CUSTOM_HEADERS",
  "ANTHROPIC_MODEL"
];
var TRANSCRIPT_BREAKING_ENV_NAMES = ["CLAUDE_CODE_SKIP_PROMPT_HISTORY"];
function claudeAuthEnvArgs(env, options) {
  const names = options.localAuth ? [...DEFAULT_AUTH_ENV_NAMES, ...LOCAL_AUTH_ENV_NAMES] : DEFAULT_AUTH_ENV_NAMES;
  const args = [];
  for (const name of names) {
    const value = env[name];
    if (value)
      args.push(`${name}=${value}`);
  }
  return args;
}
function claudeUnsetEnvArgs() {
  return [...LOCAL_AUTH_ENV_NAMES, ...TRANSCRIPT_BREAKING_ENV_NAMES].flatMap((name) => [
    "-u",
    name
  ]);
}

// src/transcript.ts
var {Glob } = globalThis.Bun;
import { join as join3, basename, dirname as dirname3, resolve as resolve2 } from "path";
import { homedir as homedir2 } from "os";
import { readFileSync as readFileSync2, statSync } from "fs";
import { readdir } from "fs/promises";
import { StringDecoder } from "string_decoder";
var POLL_MS = 100;
var START_TIMEOUT_MS = 60000;
function claudeProjectsRoot(home) {
  if (home === undefined) {
    const configDir = process.env.CLAUDE_CONFIG_DIR;
    if (configDir)
      return join3(resolve2(configDir), "projects");
    home = homedir2();
  }
  return join3(home, ".claude", "projects");
}
async function listTranscriptPaths(projectFolder) {
  const glob = new Glob("*.jsonl");
  const paths = new Set;
  try {
    for await (const name of glob.scan(projectFolder))
      paths.add(join3(projectFolder, name));
  } catch {
    return paths;
  }
  return paths;
}
async function listAllTranscriptPaths(root = claudeProjectsRoot()) {
  const paths = new Set;
  let projectFolders = [];
  try {
    projectFolders = (await readdir(root)).map((name) => join3(root, name));
  } catch {
    return paths;
  }
  for (const folder of projectFolders) {
    for (const path of await listTranscriptPaths(folder))
      paths.add(path);
  }
  return paths;
}
async function findTranscriptBySessionId(sessionId, root = claudeProjectsRoot()) {
  const name = `${sessionId}.jsonl`;
  for (const path of await listAllTranscriptPaths(root)) {
    if (basename(path) === name)
      return path;
  }
  return null;
}
var sleep = (ms) => new Promise((r) => setTimeout(r, ms));
async function waitForFreshTranscriptForCwd(cwd, before, signal, root = claudeProjectsRoot()) {
  const targetCwd = resolve2(cwd);
  const startedAt = Date.now();
  while (Date.now() - startedAt < START_TIMEOUT_MS) {
    if (signal?.aborted)
      throw new Error("aborted");
    const paths = await listAllTranscriptPaths(root);
    const fresh = [...paths].filter((p) => !before.has(p)).sort();
    for (const path of fresh) {
      const rows = await readTranscript(path);
      if (rows.some((row) => typeof row.cwd === "string" && resolve2(row.cwd) === targetCwd)) {
        return path;
      }
      if (rows.length === 0)
        continue;
    }
    await sleep(POLL_MS);
  }
  throw new Error(`Timed out waiting for a fresh transcript for cwd ${targetCwd}`);
}
async function readTranscript(transcriptPath) {
  return (await readTranscriptEntries(transcriptPath)).map((entry) => entry.row);
}
function transcriptUserText(row) {
  if (row.type !== "user")
    return;
  const message = row.message;
  if (!message || typeof message !== "object")
    return;
  const content = message.content;
  if (typeof content === "string")
    return content;
  if (!Array.isArray(content))
    return;
  const text = content.filter((part) => !!part && typeof part === "object").filter((part) => part.type === "text" && typeof part.text === "string").map((part) => part.text).join("");
  return text || undefined;
}
function transcriptRowsAfterBaseline(rows, baselineRows) {
  if (!Number.isInteger(baselineRows) || baselineRows < 0 || baselineRows > rows.length) {
    return null;
  }
  return rows.slice(baselineRows);
}
function transcriptRowsFromLastUserPrompt(rows, prompt) {
  let start = -1;
  for (let i = 0;i < rows.length; i++) {
    const text = transcriptUserText(rows[i]);
    if (text === prompt || text === "Human: " + prompt)
      start = i;
  }
  return start >= 0 ? rows.slice(start) : null;
}
async function readTranscriptEntries(transcriptPath) {
  const file = Bun.file(transcriptPath);
  if (!await file.exists())
    return [];
  const text = await file.text();
  return text.split(`
`).filter(Boolean).map((line) => {
    try {
      return { row: JSON.parse(line), line };
    } catch {
      return { row: { type: "bridge_parse_error", line }, line };
    }
  });
}
async function tailTranscript(transcriptPath, onRow, signal) {
  let emitted = 0;
  while (!signal.aborted) {
    const entries = await readTranscriptEntries(transcriptPath);
    for (let i = emitted;i < entries.length; i++) {
      const entry = entries[i];
      onRow(entry.row, i, entry.line);
      if (signal.aborted)
        break;
    }
    if (entries.length > emitted)
      emitted = entries.length;
    await sleep(POLL_MS);
  }
}
async function tailTranscriptLines(transcriptPath, onLine, signal, fromEnd = false) {
  let offset = 0;
  if (fromEnd) {
    try {
      offset = statSync(transcriptPath).size;
    } catch {
      offset = 0;
    }
  }
  let buffer = "";
  let emitted = 0;
  let decoder = new StringDecoder("utf8");
  while (!signal.aborted) {
    try {
      const size = statSync(transcriptPath).size;
      if (size < offset) {
        offset = 0;
        buffer = "";
        emitted = 0;
        decoder = new StringDecoder("utf8");
      }
      if (size > offset) {
        const chunk = decoder.write(readFileSync2(transcriptPath).subarray(offset, size));
        offset = size;
        buffer += chunk;
        const newline = buffer.lastIndexOf(`
`);
        if (newline >= 0) {
          const complete = buffer.slice(0, newline).split(`
`);
          buffer = buffer.slice(newline + 1);
          for (const line of complete) {
            if (!line)
              continue;
            onLine(line, emitted++);
            if (signal.aborted)
              break;
          }
        }
      }
    } catch {}
    await sleep(POLL_MS);
  }
}
function isTurnDurationRow(row) {
  return row.type === "system" && row.subtype === "turn_duration";
}
async function waitForTranscriptTurnEnd(transcriptPath, timeoutMs, signal, currentRows = (rows) => rows) {
  const startedAt = Date.now();
  while (!signal?.aborted && Date.now() - startedAt < timeoutMs) {
    const rows = currentRows(await readTranscript(transcriptPath));
    if (rows?.some(isTurnDurationRow))
      return true;
    await sleep(POLL_MS);
  }
  return false;
}
function sessionIdFromPath(transcriptPath) {
  return basename(transcriptPath).replace(/\.jsonl$/, "");
}
function projectFolderFromTranscript(transcriptPath) {
  return dirname3(transcriptPath);
}

// src/view.ts
var C = {
  reset: "\x1B[0m",
  dim: "\x1B[2m",
  bold: "\x1B[1m",
  red: "\x1B[31m",
  green: "\x1B[32m",
  yellow: "\x1B[33m",
  blue: "\x1B[34m",
  magenta: "\x1B[35m",
  cyan: "\x1B[36m",
  grey: "\x1B[90m"
};
var useColor = process.stdout.isTTY && !process.env.NO_COLOR;
var c = (code, s) => useColor ? code + s + C.reset : s;
var time = (iso) => {
  const d = iso ? new Date(iso) : new Date;
  const hh = String(d.getHours()).padStart(2, "0");
  const mm = String(d.getMinutes()).padStart(2, "0");
  const ss = String(d.getSeconds()).padStart(2, "0");
  return c(C.grey, `${hh}:${mm}:${ss}`);
};
var truncate = (s, n = 400) => s.length > n ? s.slice(0, n) + c(C.dim, `\u2026 (+${s.length - n} chars)`) : s;
var textFromContent = (content) => {
  if (typeof content === "string")
    return content;
  if (Array.isArray(content)) {
    return content.map((p) => {
      if (!p || typeof p !== "object")
        return "";
      const part = p;
      if (part.type === "text")
        return String(part.text ?? "");
      if (part.type === "tool_use")
        return `[tool_use ${part.name} ${JSON.stringify(part.input ?? {})}]`;
      if (part.type === "tool_result") {
        const out = part.content;
        if (typeof out === "string")
          return `[tool_result] ${out}`;
        if (Array.isArray(out))
          return `[tool_result] ${out.map((o) => o.text ?? "").join("")}`;
        return "[tool_result]";
      }
      if (part.type === "thinking")
        return c(C.dim, `(thinking) ${String(part.thinking ?? "")}`);
      return "";
    }).filter(Boolean).join(`
`);
  }
  return "";
};
var formatTranscriptRow = (row) => {
  const ts = time(typeof row.timestamp === "string" ? row.timestamp : undefined);
  const type = String(row.type ?? "?");
  if (type === "user") {
    const msg = row.message ?? {};
    const text = textFromContent(msg.content);
    if (!text.trim())
      return null;
    return `${ts} ${c(C.cyan, "user      ")} ${truncate(text)}`;
  }
  if (type === "assistant") {
    const msg = row.message ?? {};
    const text = textFromContent(msg.content);
    if (!text.trim())
      return null;
    return `${ts} ${c(C.green, "assistant ")} ${truncate(text)}`;
  }
  if (type === "system") {
    const sub = String(row.subtype ?? "");
    if (sub === "init") {
      const m = row.message ?? {};
      return `${ts} ${c(C.yellow, "system    ")} init session=${row.session_id ?? m.session_id ?? "?"}`;
    }
    if (sub === "turn_duration") {
      return `${ts} ${c(C.yellow, "system    ")} turn_duration=${row.durationMs}ms`;
    }
    return null;
  }
  if (type === "bridge_parse_error") {
    return `${ts} ${c(C.red, "parse_err ")} ${String(row.line ?? "").slice(0, 200)}`;
  }
  return null;
};
function formatEnvelope(env, opts = {}) {
  const ts = time();
  const type = String(env.type ?? "?");
  switch (type) {
    case "push":
      return `${ts} ${c(C.bold + C.magenta, "\u2192 push    ")} ${c(C.dim, `id=${env.id}`)} ${String(env.content ?? "")}`;
    case "reply":
      return `${ts} ${c(C.bold + C.blue, "\u2190 reply   ")} ${c(C.dim, `id=${env.chat_id}`)} ${String(env.text ?? "")}`;
    case "channel_hello":
      return `${ts} ${c(C.dim, "channel mcp connected pid=" + env.pid)}`;
    case "transcript_folder":
      return `${ts} ${c(C.dim, "transcript folder " + env.path)}`;
    case "transcript_open":
      return `${ts} ${c(C.dim, "transcript " + env.path)}`;
    case "transcript": {
      const row = env.row ?? {};
      const friendly = formatTranscriptRow(row);
      const rawLine = `${ts} ${c(C.dim, "raw       ")} ${truncate(JSON.stringify(row), 600)}`;
      if (opts.showRaw)
        return friendly ? `${friendly}
${c(C.dim, rawLine)}` : rawLine;
      return friendly;
    }
    case "bridge_warning":
      return `${ts} ${c(C.yellow, "warning   ")} ${String(env.message ?? "")}`;
    case "bridge_error":
      return `${ts} ${c(C.red, "error     ")} ${String(env.where ?? "")}: ${String(env.message ?? "")}`;
    case "desplega_debug": {
      const data = env.data ? ` ${truncate(JSON.stringify(env.data), 500)}` : "";
      return `${ts} ${c(C.dim, "debug     ")} ${String(env.message ?? "")}${data}`;
    }
    default:
      return `${ts} ${c(C.grey, type.padEnd(10).slice(0, 10))} ${JSON.stringify(env)}`;
  }
}
if (false) {}

// src/args.ts
var BLOCKED_OPTIONS = new Map([
  [
    "--init-only",
    {
      label: "--init-only",
      reason: "it sets Claude Code isInteractive=false and violates the bridge billing invariant."
    }
  ],
  [
    "--sdk-url",
    {
      label: "--sdk-url",
      reason: "Agent SDK mode sets Claude Code isInteractive=false and violates the bridge billing invariant."
    }
  ],
  [
    "--tmux",
    {
      label: "--tmux",
      reason: "the wrapper owns tmux session creation."
    }
  ],
  [
    "-w",
    {
      label: "-w, --worktree",
      reason: "the wrapper owns the per-run workdir."
    }
  ],
  [
    "--worktree",
    {
      label: "-w, --worktree",
      reason: "the wrapper owns the per-run workdir."
    }
  ],
  [
    "--replay-user-messages",
    {
      label: "--replay-user-messages",
      reason: "replay mode is not supported through this wrapper."
    }
  ]
]);
var CLAUDE_COMMANDS = new Map([
  ["agents", "agents"],
  ["auth", "auth"],
  ["auto-mode", "auto-mode"],
  ["doctor", "doctor"],
  ["install", "install"],
  ["mcp", "mcp"],
  ["plugin", "plugin"],
  ["plugins", "plugin"],
  ["project", "project"],
  ["setup-token", "setup-token"],
  ["ultrareview", "ultrareview"],
  ["update", "update"],
  ["upgrade", "update"]
]);
var OPTION_ARITY = new Map([
  ["--add-dir", "variadic"],
  ["--agent", "required"],
  ["--agents", "required"],
  ["--allow-dangerously-skip-permissions", "none"],
  ["--allowedTools", "variadic"],
  ["--allowed-tools", "variadic"],
  ["--append-system-prompt", "required"],
  ["--append-system-prompt-file", "required"],
  ["--bare", "none"],
  ["--betas", "variadic"],
  ["--brief", "none"],
  ["--chrome", "none"],
  ["-c", "none"],
  ["--continue", "none"],
  ["--dangerously-skip-permissions", "none"],
  ["-d", "optional"],
  ["--debug", "optional"],
  ["--debug-file", "required"],
  ["--disable-slash-commands", "none"],
  ["--disallowedTools", "variadic"],
  ["--disallowed-tools", "variadic"],
  ["--effort", "required"],
  ["--exclude-dynamic-system-prompt-sections", "none"],
  ["--fallback-model", "required"],
  ["--file", "variadic"],
  ["--fork-session", "none"],
  ["--from-pr", "required"],
  ["-h", "none"],
  ["--help", "none"],
  ["--ide", "none"],
  ["--include-hook-events", "none"],
  ["--include-partial-messages", "none"],
  ["--input-format", "required"],
  ["--json-schema", "required"],
  ["--max-budget-usd", "required"],
  ["--mcp-config", "variadic"],
  ["--mcp-debug", "none"],
  ["--model", "required"],
  ["-n", "required"],
  ["--name", "required"],
  ["--no-chrome", "none"],
  ["--no-session-persistence", "none"],
  ["--output-format", "required"],
  ["--permission-mode", "required"],
  ["--plugin-dir", "variadic"],
  ["--plugin-url", "required"],
  ["-p", "none"],
  ["--print", "none"],
  ["--prompt-suggestions", "none"],
  ["--remote-control", "none"],
  ["--remote-control-session-name-prefix", "required"],
  ["--replay-user-messages", "variadic"],
  ["-r", "optional"],
  ["--resume", "optional"],
  ["--session-id", "required"],
  ["--setting-sources", "variadic"],
  ["--settings", "required"],
  ["--strict-mcp-config", "none"],
  ["--system-prompt", "required"],
  ["--tmux", "optional"],
  ["--tools", "variadic"],
  ["--verbose", "none"],
  ["-w", "optional"],
  ["--worktree", "optional"]
]);
function parseCliArgs(argv) {
  const rawClaudeArgs = [];
  const desplegaArgs = [];
  let claudeHelp = false;
  for (let i = 0;i < argv.length; i++) {
    const arg = argv[i];
    if (arg === "--claude-help") {
      claudeHelp = true;
      continue;
    }
    if (arg === "--desplega") {
      return {
        ok: false,
        error: {
          arg,
          message: "`--desplega` is not a valid wrapper option.",
          hint: "Use `--desplega-<name>` or `--desplega-<name>=<value>`."
        }
      };
    }
    if (arg.startsWith("--desplega-")) {
      const parsed = parseDesplegaArg(arg);
      if (!parsed) {
        return {
          ok: false,
          error: {
            arg,
            message: "`--desplega-` requires an option name.",
            hint: "Use `--desplega-verbose` or another `--desplega-<name>` option."
          }
        };
      }
      if (parsed.value === true && ["cwd", "json-schema-max-tokens", "schema-max-tokens"].includes(parsed.name) && i + 1 < argv.length && !looksLikeOption(argv[i + 1])) {
        parsed.value = argv[++i];
      }
      desplegaArgs.push(parsed);
      continue;
    }
    rawClaudeArgs.push(arg);
  }
  if (containsHelp(rawClaudeArgs)) {
    return {
      ok: true,
      parsed: {
        help: true,
        claudeHelp,
        version: false,
        print: false,
        outputFormat: "text",
        jsonSchema: undefined,
        claudeArgs: [],
        desplegaArgs,
        desplegaVerbose: hasDesplegaVerbose(desplegaArgs)
      }
    };
  }
  const extracted = extractClaudeArgs(rawClaudeArgs);
  if (!extracted.ok)
    return extracted;
  return {
    ok: true,
    parsed: {
      help: false,
      claudeHelp,
      version: extracted.parsed.version,
      print: extracted.parsed.print,
      outputFormat: extracted.parsed.outputFormat,
      jsonSchema: extracted.parsed.jsonSchema,
      claudeArgs: extracted.parsed.claudeArgs,
      initialPrompt: extracted.parsed.initialPrompt,
      desplegaArgs,
      desplegaVerbose: hasDesplegaVerbose(desplegaArgs)
    }
  };
}
function formatWrapperHelp() {
  const lines = [
    "Usage: claude-bridge [wrapper options] [claude options] [prompt]",
    "",
    "Starts Claude Code inside the tmux/transcript orchestrator. Normal Claude",
    "interactive options are forwarded as-is.",
    "",
    "Wrapper options:",
    "  -p, --print                     Non-interactive mode; exit after the reply",
    "  --output-format <format>        Output: text, json, or stream-json",
    "  --json-schema <schema|file>     Extract and validate final JSON in print mode",
    "  --desplega-format               Use bridge JSON envelopes in json modes",
    "  --desplega-verbose              Emit extra wrapper debug output",
    "  --desplega-cwd <path>           Run Claude in this cwd (default: current cwd)",
    "  --desplega-local-auth           Forward local Anthropic auth env vars",
    "  --desplega-install              Install the global schema Stop hook",
    "  --desplega-uninstall            Remove the global schema Stop hook",
    "  --desplega-<name>[=<value>]     Reserve a desplega wrapper option",
    "  --claude-help                   Show raw `claude -h` output",
    "  -v, --version                   Show claude-bridge and Claude CLI versions",
    "  -h, --help                      Show this help",
    "",
    "Blocked Claude options:",
    "  --tmux                          The wrapper owns tmux session creation",
    "  -w, --worktree                  The wrapper owns the per-run workdir",
    "  --replay-user-messages          Replay is not supported through this wrapper",
    "",
    "Claude subcommands are blocked. Use `claude <cmd>` directly for commands",
    "such as `doctor`, `mcp`, `plugin`, `update`, or `agents`.",
    "",
    "Most Claude interactive options are forwarded. Use `--claude-help` when",
    "you need the raw Claude help, but remember this wrapper owns print/output",
    "mode, schema extraction, tmux, workdir, and prompt delivery."
  ];
  return lines.join(`
`) + `
`;
}
function formatClaudeHelp(claudeHelp) {
  return [
    "Raw Claude help follows. Note: claude-bridge owns -p/--print, --output-format,",
    "--json-schema, --tmux, --worktree, and prompt delivery; use `claude <cmd>`",
    "directly for Claude subcommands.",
    "",
    claudeHelp.trimEnd(),
    ""
  ].join(`
`);
}
function claudeResumeSessionId(args) {
  for (let i = 0;i < args.length; i++) {
    const arg = args[i];
    if (arg.startsWith("--resume=")) {
      return arg.slice("--resume=".length) || undefined;
    }
    if (arg === "--resume" || arg === "-r") {
      const next = args[i + 1];
      return next && !looksLikeOption(next) ? next : undefined;
    }
  }
  return;
}
function extractClaudeArgs(args) {
  const claudeArgs = [];
  const promptParts = [];
  let version = false;
  let print = false;
  let outputFormat = "text";
  let outputFormatSpecified = false;
  let jsonSchema;
  for (let i = 0;i < args.length; i++) {
    const arg = args[i];
    if (arg === "--") {
      promptParts.push(...args.slice(i + 1));
      break;
    }
    const blocked = blockedOption(arg);
    if (blocked) {
      return {
        ok: false,
        error: {
          arg,
          message: `Unsupported Claude option ${blocked.label}: ${blocked.reason}`,
          hint: `Run \`claude ${argName(arg)}\` directly if you need that mode.`
        }
      };
    }
    if (arg === "-p" || arg === "--print") {
      print = true;
      continue;
    }
    if (arg === "-v" || arg === "--version") {
      version = true;
      continue;
    }
    if (!arg.startsWith("--") && arg.startsWith("-p")) {
      return {
        ok: false,
        error: {
          arg,
          message: "`-p` does not accept an attached value.",
          hint: 'Use `-p "your prompt"` or `--print "your prompt"`.'
        }
      };
    }
    if (arg.startsWith("--print=")) {
      return {
        ok: false,
        error: {
          arg,
          message: "`--print` does not accept a value.",
          hint: "Use `-p` or `--print`, then provide the prompt as a separate argument."
        }
      };
    }
    if (arg === "--output-format" || arg.startsWith("--output-format=")) {
      const parsed = parseOutputFormat(args, i);
      if (!parsed.ok)
        return parsed;
      outputFormat = parsed.value;
      outputFormatSpecified = true;
      i = parsed.nextIndex;
      continue;
    }
    if (arg === "--json-schema" || arg.startsWith("--json-schema=")) {
      const parsed = parseRequiredValue(args, i, "--json-schema");
      if (!parsed.ok)
        return parsed;
      if (jsonSchema !== undefined) {
        return {
          ok: false,
          error: {
            arg,
            message: "`--json-schema` can only be provided once.",
            hint: "Pass one JSON Schema object or one schema file path."
          }
        };
      }
      jsonSchema = parsed.value;
      i = parsed.nextIndex;
      continue;
    }
    if (looksLikeOption(arg)) {
      claudeArgs.push(arg);
      const name = argName(arg);
      if (arg.includes("="))
        continue;
      const arity = OPTION_ARITY.get(name);
      if (arity === "required") {
        if (i + 1 < args.length)
          claudeArgs.push(args[++i]);
      } else if (arity === "optional") {
        if (i + 1 < args.length && !looksLikeOption(args[i + 1])) {
          claudeArgs.push(args[++i]);
        }
      } else if (arity === "variadic") {
        while (i + 1 < args.length && !looksLikeOption(args[i + 1])) {
          claudeArgs.push(args[++i]);
        }
      }
      continue;
    }
    if (promptParts.length === 0) {
      const command = CLAUDE_COMMANDS.get(arg);
      if (command) {
        return {
          ok: false,
          error: {
            arg,
            message: `Claude command \`${arg}\` is not supported by this wrapper.`,
            hint: `Use \`claude ${arg}\` directly.`
          }
        };
      }
    }
    promptParts.push(arg);
  }
  if (outputFormatSpecified && !print) {
    return {
      ok: false,
      error: {
        arg: "--output-format",
        message: "`--output-format` requires `-p`/`--print`.",
        hint: 'Use `-p "your prompt" --output-format stream-json`.'
      }
    };
  }
  if (jsonSchema !== undefined && !print) {
    return {
      ok: false,
      error: {
        arg: "--json-schema",
        message: "`--json-schema` requires `-p`/`--print`.",
        hint: 'Use `-p "your prompt" --json-schema schema.json --output-format json`.'
      }
    };
  }
  return {
    ok: true,
    parsed: {
      print,
      version,
      outputFormat,
      jsonSchema,
      claudeArgs,
      initialPrompt: promptParts.length ? promptParts.join(" ") : undefined
    }
  };
}
function parseRequiredValue(args, index, name) {
  const arg = args[index];
  const value = arg.includes("=") ? arg.slice(arg.indexOf("=") + 1) : args[index + 1];
  if (!value) {
    return {
      ok: false,
      error: {
        arg,
        message: `\`${name}\` requires a value.`,
        hint: `Use \`${name} schema.json\` or \`${name} '{"type":"object"}'\`.`
      }
    };
  }
  return {
    ok: true,
    value,
    nextIndex: arg.includes("=") ? index : index + 1
  };
}
function parseOutputFormat(args, index) {
  const parsed = parseRequiredValue(args, index, "--output-format");
  if (!parsed.ok) {
    return {
      ok: false,
      error: {
        ...parsed.error,
        hint: "Use `--output-format text`, `--output-format json`, or `--output-format stream-json`."
      }
    };
  }
  const arg = args[index];
  const value = parsed.value;
  if (!isOutputFormat(value)) {
    return {
      ok: false,
      error: {
        arg,
        message: `Unsupported output format \`${value}\`.`,
        hint: "Use `text`, `json`, or `stream-json`."
      }
    };
  }
  return {
    ok: true,
    value,
    nextIndex: parsed.nextIndex
  };
}
function isOutputFormat(value) {
  return value === "text" || value === "json" || value === "stream-json";
}
function parseDesplegaArg(arg) {
  const body = arg.slice("--desplega-".length);
  if (!body)
    return null;
  const eq = body.indexOf("=");
  const name = eq === -1 ? body : body.slice(0, eq);
  if (!name)
    return null;
  return {
    raw: arg,
    name,
    value: eq === -1 ? true : parseDesplegaValue(body.slice(eq + 1))
  };
}
function parseDesplegaValue(value) {
  const normalized = value.trim().toLowerCase();
  if (["", "1", "true", "yes", "on"].includes(normalized))
    return true;
  if (["0", "false", "no", "off"].includes(normalized))
    return false;
  return value;
}
function hasDesplegaVerbose(args) {
  const flag = args.find((arg) => arg.name === "verbose");
  return flag ? flag.value !== false : false;
}
function containsHelp(args) {
  for (const arg of claudeArgsBeforeDashDash(args)) {
    if (arg === "-h" || arg === "--help")
      return true;
  }
  return false;
}
function* claudeArgsBeforeDashDash(args) {
  for (const arg of args) {
    if (arg === "--")
      return;
    yield arg;
  }
}
function blockedOption(arg) {
  const name = argName(arg);
  if (name.startsWith("--replay")) {
    return BLOCKED_OPTIONS.get("--replay-user-messages");
  }
  if (!arg.startsWith("--") && arg.startsWith("-w"))
    return BLOCKED_OPTIONS.get("-w");
  return BLOCKED_OPTIONS.get(name);
}
function looksLikeOption(arg) {
  return arg !== "-" && arg.startsWith("-");
}
function argName(arg) {
  const eq = arg.indexOf("=");
  return eq === -1 ? arg : arg.slice(0, eq);
}

// src/json-schema.ts
import { existsSync as existsSync2, readFileSync as readFileSync3 } from "fs";
import { resolve as resolve3 } from "path";

// node_modules/zod/v4/core/util.js
function getEnumValues(entries) {
  const numericValues = Object.values(entries).filter((v) => typeof v === "number");
  const values = Object.entries(entries).filter(([k, _]) => numericValues.indexOf(+k) === -1).map(([_, v]) => v);
  return values;
}
function joinValues(array, separator = "|") {
  return array.map((val) => stringifyPrimitive(val)).join(separator);
}
function jsonStringifyReplacer(_, value) {
  if (typeof value === "bigint")
    return value.toString();
  return value;
}

class Cached {
  constructor(getter) {
    this._getter = getter;
    this._value = undefined;
  }
  get value() {
    const getter = this._getter;
    if (getter !== undefined) {
      this._value = getter();
      this._getter = undefined;
    }
    return this._value;
  }
}
function cached(getter) {
  return new Cached(getter);
}
function nullish(input) {
  return input === null || input === undefined;
}
function cleanRegex(source) {
  const start = source.startsWith("^") ? 1 : 0;
  const end = source.endsWith("$") ? source.length - 1 : source.length;
  return source.slice(start, end);
}
function floatSafeRemainder(val, step) {
  const ratio = val / step;
  const roundedRatio = Math.round(ratio);
  const tolerance = 4 * Number.EPSILON * Math.max(Math.abs(ratio), 1);
  if (Math.abs(ratio - roundedRatio) < tolerance)
    return 0;
  return ratio - roundedRatio;
}
var EVALUATING = /* @__PURE__ */ Symbol("evaluating");
function defineLazy(object, key, getter) {
  let value = undefined;
  Object.defineProperty(object, key, {
    get() {
      if (value === EVALUATING) {
        return;
      }
      if (value === undefined) {
        value = EVALUATING;
        value = getter();
      }
      return value;
    },
    set(v) {
      Object.defineProperty(object, key, {
        value: v
      });
    },
    configurable: true
  });
}
function assignProp(target, prop, value) {
  Object.defineProperty(target, prop, {
    value,
    writable: true,
    enumerable: true,
    configurable: true
  });
}
function rawShape(def) {
  const desc = Object.getOwnPropertyDescriptor(def, "shape");
  return desc?.get ? desc.get.raw : desc?.value;
}
function sourceShape(schema) {
  return rawShape(schema._zod.def) ?? schema._zod.def.shape;
}
function deferProp(target, key, getter) {
  Object.defineProperty(target, key, {
    get() {
      const value = getter();
      assignProp(this, key, value);
      return value;
    },
    enumerable: true,
    configurable: true
  });
}
function putProp(target, key, value) {
  if (key in target)
    assignProp(target, key, value);
  else
    target[key] = value;
}
function mirrorShape(target, source, keys, wrap) {
  const raw = sourceShape(source);
  for (const key of keys) {
    const desc = Object.getOwnPropertyDescriptor(raw, key);
    if (!desc.enumerable)
      continue;
    if (desc.get) {
      deferProp(target, key, () => {
        const value = source._zod.def.shape[key];
        return wrap ? wrap(value, key) : value;
      });
    } else
      putProp(target, key, wrap ? wrap(desc.value, key) : desc.value);
  }
}
function mirrorProps(target, source) {
  for (const key of Reflect.ownKeys(source)) {
    const desc = Object.getOwnPropertyDescriptor(source, key);
    if (!desc.enumerable)
      continue;
    if (desc.get)
      deferProp(target, key, () => source[key]);
    else
      putProp(target, key, desc.value);
  }
}
function mergeDefs(...defs) {
  const mergedDescriptors = {};
  for (const def of defs) {
    const descriptors = Object.getOwnPropertyDescriptors(def);
    Object.assign(mergedDescriptors, descriptors);
  }
  return Object.defineProperties({}, mergedDescriptors);
}
function esc(str) {
  return JSON.stringify(str);
}
function slugify(input) {
  return input.toLowerCase().trim().replace(/[^\w\s-]/g, "").replace(/[\s_-]+/g, "-").replace(/^-+|-+$/g, "");
}
var captureStackTrace = "captureStackTrace" in Error ? Error.captureStackTrace : (..._args) => {};
function isObject(data) {
  return typeof data === "object" && data !== null && !Array.isArray(data);
}
var allowsEval = /* @__PURE__ */ cached(() => {
  if (globalConfig.jitless) {
    return false;
  }
  if (typeof navigator !== "undefined" && navigator?.userAgent?.includes("Cloudflare")) {
    return false;
  }
  try {
    const F = Function;
    new F("");
    return true;
  } catch (_) {
    return false;
  }
});
function isPlainObject(o) {
  if (isObject(o) === false)
    return false;
  const ctor = o.constructor;
  if (ctor === undefined)
    return true;
  if (typeof ctor !== "function")
    return true;
  const prot = ctor.prototype;
  if (isObject(prot) === false)
    return false;
  if (Object.prototype.hasOwnProperty.call(prot, "isPrototypeOf") === false) {
    return false;
  }
  return true;
}
function shallowClone(o) {
  if (isPlainObject(o))
    return { ...o };
  if (Array.isArray(o))
    return [...o];
  if (o instanceof Map)
    return new Map(o);
  if (o instanceof Set)
    return new Set(o);
  return o;
}
var propertyKeyTypes = /* @__PURE__ */ new Set(["string", "number", "symbol"]);
var primitiveTypes = /* @__PURE__ */ new Set([
  "string",
  "number",
  "bigint",
  "boolean",
  "symbol",
  "undefined"
]);
function escapeRegex(str) {
  return str.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
function clone(inst, def, params) {
  const cl = new inst._zod.constr(def ?? inst._zod.def);
  if (!def || params?.parent)
    cl._zod.parent = inst;
  return cl;
}
function normalizeParams(_params) {
  const params = _params;
  if (!params)
    return {};
  if (typeof params === "string")
    return { error: () => params };
  if (params?.message !== undefined) {
    if (params?.error !== undefined)
      throw new Error("Cannot specify both `message` and `error` params");
    params.error = params.message;
  }
  delete params.message;
  if (typeof params.error === "string")
    return { ...params, error: () => params.error };
  return params;
}
function stringifyPrimitive(value) {
  if (typeof value === "bigint")
    return value.toString() + "n";
  if (typeof value === "string")
    return `"${value}"`;
  return `${value}`;
}
function optionalKeys(shape) {
  return Object.keys(shape).filter((k) => {
    return shape[k]._zod.optin !== undefined && shape[k]._zod.optout === "optional";
  });
}
var NUMBER_FORMAT_RANGES = /* @__PURE__ */ (() => ({
  safeint: [Number.MIN_SAFE_INTEGER, Number.MAX_SAFE_INTEGER],
  int32: [-2147483648, 2147483647],
  uint32: [0, 4294967295],
  float32: [-340282346638528860000000000000000000000, 340282346638528860000000000000000000000],
  float64: [-Number.MAX_VALUE, Number.MAX_VALUE]
}))();
var BIGINT_FORMAT_RANGES = {
  int64: [/* @__PURE__ */ BigInt("-9223372036854775808"), /* @__PURE__ */ BigInt("9223372036854775807")],
  uint64: [/* @__PURE__ */ BigInt(0), /* @__PURE__ */ BigInt("18446744073709551615")]
};
function pick(schema, mask) {
  const currDef = schema._zod.def;
  const checks = currDef.checks;
  const hasChecks = checks && checks.length > 0;
  if (hasChecks) {
    throw new Error(".pick() cannot be used on object schemas containing refinements");
  }
  const newShape = {};
  mirrorShape(newShape, schema, maskedKeys(schema, mask));
  return clone(schema, mergeDefs(currDef, { shape: newShape, checks: [] }));
}
function maskedKeys(schema, mask) {
  const raw = sourceShape(schema);
  const keys = [];
  for (const key of Reflect.ownKeys(mask)) {
    if (!Object.getOwnPropertyDescriptor(raw, key)?.enumerable) {
      throw new Error(`Unrecognized key: "${String(key)}"`);
    }
    if (mask[key])
      keys.push(key);
  }
  return keys;
}
function omit(schema, mask) {
  const currDef = schema._zod.def;
  const checks = currDef.checks;
  const hasChecks = checks && checks.length > 0;
  if (hasChecks) {
    throw new Error(".omit() cannot be used on object schemas containing refinements");
  }
  const omitted = new Set(maskedKeys(schema, mask));
  const newShape = {};
  mirrorShape(newShape, schema, Reflect.ownKeys(sourceShape(schema)).filter((key) => !omitted.has(key)));
  return clone(schema, mergeDefs(currDef, { shape: newShape, checks: [] }));
}
function extend(schema, shape) {
  if (!isPlainObject(shape)) {
    throw new Error("Invalid input to extend: expected a plain object");
  }
  const checks = schema._zod.def.checks;
  const hasChecks = checks && checks.length > 0;
  if (hasChecks) {
    const existingShape = sourceShape(schema);
    for (const key of Reflect.ownKeys(shape)) {
      if (Object.getOwnPropertyDescriptor(existingShape, key) !== undefined) {
        throw new Error("Cannot overwrite keys on object schemas containing refinements. Use `.safeExtend()` instead.");
      }
    }
  }
  return clone(schema, mergeDefs(schema._zod.def, { shape: extended(schema, shape) }));
}
function extended(schema, shape) {
  const newShape = {};
  mirrorShape(newShape, schema, Reflect.ownKeys(sourceShape(schema)));
  mirrorProps(newShape, shape);
  return newShape;
}
function safeExtend(schema, shape) {
  if (!isPlainObject(shape)) {
    throw new Error("Invalid input to safeExtend: expected a plain object");
  }
  return clone(schema, mergeDefs(schema._zod.def, { shape: extended(schema, shape) }));
}
function merge(a, b) {
  if (!b?._zod?.def) {
    throw new Error("Invalid input to merge: expected an object schema. To merge a plain shape, use `.extend()`.");
  }
  if (a._zod.def.checks?.length) {
    throw new Error(".merge() cannot be used on object schemas containing refinements. Use .safeExtend() instead.");
  }
  const newShape = {};
  mirrorShape(newShape, a, Reflect.ownKeys(sourceShape(a)));
  mirrorShape(newShape, b, Reflect.ownKeys(sourceShape(b)));
  const def = mergeDefs(a._zod.def, {
    shape: newShape,
    get catchall() {
      return b._zod.def.catchall;
    },
    checks: b._zod.def.checks ?? []
  });
  return clone(a, def);
}
function partial(Class, schema, mask, name = "partial") {
  const currDef = schema._zod.def;
  const checks = currDef.checks;
  const hasChecks = checks && checks.length > 0;
  if (hasChecks) {
    throw new Error(`.${name}() cannot be used on object schemas containing refinements`);
  }
  const selected = mask ? new Set(maskedKeys(schema, mask)) : undefined;
  const newShape = {};
  mirrorShape(newShape, schema, Reflect.ownKeys(sourceShape(schema)), Class && ((value, key) => selected && !selected.has(key) ? value : new Class({ type: "optional", innerType: value })));
  return clone(schema, mergeDefs(schema._zod.def, { shape: newShape, checks: [] }));
}
function required(Class, schema, mask) {
  const selected = mask ? new Set(maskedKeys(schema, mask)) : undefined;
  const newShape = {};
  mirrorShape(newShape, schema, Reflect.ownKeys(sourceShape(schema)), (value, key) => selected && !selected.has(key) ? value : new Class({ type: "nonoptional", innerType: value }));
  return clone(schema, mergeDefs(schema._zod.def, { shape: newShape }));
}
function aborted(x, startIndex = 0) {
  if (x.aborted === true)
    return true;
  for (let i = startIndex;i < x.issues.length; i++) {
    if (x.issues[i]?.continue !== true) {
      return true;
    }
  }
  return false;
}
function explicitlyAborted(x, startIndex = 0) {
  if (x.aborted === true)
    return true;
  for (let i = startIndex;i < x.issues.length; i++) {
    if (x.issues[i]?.continue === false) {
      return true;
    }
  }
  return false;
}
function prefixIssues(path, issues) {
  return issues.map((iss) => {
    var _a;
    (_a = iss).path ?? (_a.path = []);
    iss.path.unshift(path);
    return iss;
  });
}
function unwrapMessage(message) {
  return typeof message === "string" ? message : message?.message;
}
function sizeCaps(defChecks) {
  const caps = [];
  for (const ch of defChecks ?? []) {
    const d = ch._zod.def;
    if (!/_(length|size)$|_equals$/.test(d.check))
      break;
    const max = d.maximum ?? d.length ?? d.size;
    if (max >= 0)
      caps.push([max, ch]);
  }
  return caps;
}
function attachSchema(issues, start, inst) {
  var _a;
  for (let i = start;i < issues.length; i++) {
    (_a = issues[i]).schema ?? (_a.schema = inst);
  }
}
function finalizeIssue(iss, ctx, config) {
  var _a;
  const traits = iss.inst?._zod?.traits;
  if (traits?.has("$ZodType")) {
    if (traits.has("$ZodCheck"))
      (_a = iss).schema ?? (_a.schema = iss.inst);
    else
      iss.schema = iss.inst;
  }
  const schemaError = iss.schema !== iss.inst ? iss.schema?._zod.def?.error : undefined;
  const message = iss.message ? iss.message : unwrapMessage(iss.inst?._zod.def?.error?.(iss)) ?? unwrapMessage(schemaError?.(iss)) ?? unwrapMessage(ctx?.error?.(iss)) ?? unwrapMessage(config.customError?.(iss)) ?? unwrapMessage(config.localeError?.(iss)) ?? "Invalid input";
  const full = {};
  for (const k of Object.keys(iss)) {
    if (k === "inst" || k === "schema" || k === "continue" || k === "input" || k === "__proto__")
      continue;
    full[k] = iss[k];
  }
  full.path ?? (full.path = []);
  full.message = message;
  if (ctx?.reportInput) {
    full.input = iss.input;
  }
  return full;
}
function getSizableOrigin(input) {
  if (input instanceof Set)
    return "set";
  if (input instanceof Map)
    return "map";
  if (input instanceof File)
    return "file";
  return "unknown";
}
var highSurrogate = /[\uD800-\uDBFF]/;
function codePointLength(str) {
  const units = str.length;
  if (!highSurrogate.test(str))
    return units;
  let count = units;
  for (let i = 0;i < units - 1; i++) {
    if ((str.charCodeAt(i) & 64512) === 55296 && (str.charCodeAt(i + 1) & 64512) === 56320) {
      count--;
      i++;
    }
  }
  return count;
}
function getLengthableOrigin(input) {
  if (Array.isArray(input))
    return "array";
  if (typeof input === "string")
    return "string";
  return "unknown";
}
function parsedType(data) {
  const t = typeof data;
  switch (t) {
    case "number": {
      return Number.isNaN(data) ? "nan" : "number";
    }
    case "object": {
      if (data === null) {
        return "null";
      }
      if (Array.isArray(data)) {
        return "array";
      }
      const obj = data;
      if (obj && Object.getPrototypeOf(obj) !== Object.prototype && "constructor" in obj && obj.constructor) {
        return obj.constructor.name;
      }
    }
  }
  return t;
}
function issue(...args) {
  const [iss, input, inst] = args;
  if (typeof iss === "string") {
    return {
      message: iss,
      code: "custom",
      input,
      inst
    };
  }
  return { ...iss };
}
function members(proto, table) {
  for (const key in table) {
    const desc = Object.getOwnPropertyDescriptor(table, key);
    if (desc.get)
      Object.defineProperty(proto, key, { ...desc, enumerable: false });
    else
      defineBound(proto, key, desc.value);
  }
}
function own(inst, key, value, enumerable = true) {
  Object.defineProperty(inst, key, { configurable: true, writable: true, enumerable, value });
  return value;
}
function hide(inst, key, value) {
  return own(inst, key, value, false);
}
function derived(computes, table) {
  for (const key in computes) {
    const compute = computes[key];
    Object.defineProperty(table, key, {
      configurable: true,
      enumerable: true,
      get() {
        return own(this, key, compute(this));
      },
      set(value) {
        own(this, key, value);
      }
    });
  }
  return table;
}
function defineBound(proto, key, fn) {
  Object.defineProperty(proto, key, {
    configurable: true,
    get() {
      return this == null ? fn : own(this, key, fn.bind(this));
    },
    set(value) {
      own(this, key, value);
    }
  });
}
function claim(inst, sentinel) {
  const proto = Object.getPrototypeOf(inst);
  return sentinel in proto ? undefined : proto;
}
var installing;
var broke = false;
var breaker = {
  configurable: true,
  get() {
    broke = true;
    return;
  }
};
function defineLazyInternal(inst, key, compute) {
  const proto = Object.getPrototypeOf(inst._zod);
  if (key in proto && installing !== inst._zod) {
    installing = undefined;
    return;
  }
  installing = inst._zod;
  Object.defineProperty(proto, key, {
    configurable: true,
    get() {
      Object.defineProperty(this, key, breaker);
      const outer = broke;
      broke = false;
      try {
        const value = compute(this);
        if (broke)
          delete this[key];
        else
          Object.defineProperty(this, key, { configurable: true, writable: true, value });
        broke = broke || outer;
        return value;
      } catch (err) {
        delete this[key];
        broke = broke || outer;
        throw err;
      }
    },
    set(value) {
      Object.defineProperty(this, key, { configurable: true, writable: true, value });
    }
  });
}
function installLazyProp(inst, key, make, enumerable) {
  const proto = claim(inst, key);
  if (!proto)
    return;
  Object.defineProperty(proto, key, {
    configurable: true,
    get() {
      const desc = { configurable: true, writable: true, enumerable, value: undefined };
      Object.defineProperty(this, key, desc);
      desc.value = make(this);
      Object.defineProperty(this, key, desc);
      return desc.value;
    },
    set(value) {
      Object.defineProperty(this, key, { configurable: true, writable: true, enumerable, value });
    }
  });
}
var CONSTANT_CATCH = "~constantCatch";
function constantCatch(value) {
  const fn = () => value;
  fn[CONSTANT_CATCH] = true;
  return fn;
}

// node_modules/zod/v4/core/core.js
var _a;
var _zodDesc = { value: undefined, enumerable: false };
var _E = "captureStackTrace" in Error ? Error : null;
function newError(Definition) {
  const E = _E;
  if (E) {
    const saved = E.stackTraceLimit;
    if (typeof saved === "number") {
      try {
        E.stackTraceLimit = 0;
      } catch {
        _E = null;
        return new Definition;
      }
      try {
        return new Definition;
      } finally {
        E.stackTraceLimit = saved;
      }
    }
  }
  return new Definition;
}
function $constructor(name, initializer, proto, params) {
  const zodProto = {};
  function Internals(def) {
    this.def = def;
    this.constr = _;
    this.traits = new Set;
  }
  Internals.prototype = zodProto;
  const protoMembers = proto;
  const initialized = protoMembers && new WeakSet;
  function init(inst, def) {
    if (!inst._zod) {
      _zodDesc.value = new Internals(def);
      try {
        Object.defineProperty(inst, "_zod", _zodDesc);
      } finally {
        _zodDesc.value = undefined;
      }
    } else if (inst._zod.traits.has(name)) {
      return;
    }
    inst._zod.traits.add(name);
    initializer(inst, def);
    if (initialized) {
      const own = Object.getPrototypeOf(inst);
      const ctorProto = inst._zod.constr.prototype;
      let up = own;
      while (up && up !== ctorProto)
        up = Object.getPrototypeOf(up);
      const target = up ?? own;
      if (!initialized.has(target)) {
        initialized.add(target);
        members(target, protoMembers);
      }
    }
    const proto = _.prototype;
    for (const k in proto) {
      if (!Object.prototype.hasOwnProperty.call(proto, k))
        continue;
      if (!(k in inst)) {
        inst[k] = proto[k].bind(inst);
      }
    }
  }
  const Parent = params?.Parent ?? Object;

  class Definition extends Parent {
  }
  Object.defineProperty(Definition, "name", { value: name });
  function _(def) {
    const inst = params?.Parent ? newError(Definition) : this;
    init(inst, def);
    const deferred = inst._zod.deferred;
    if (deferred) {
      for (const fn of deferred) {
        fn();
      }
      inst._zod.deferred = undefined;
    }
    const pp = globalThis.__zod_globalConfig?.postProcessor;
    if (pp)
      pp(inst);
    return inst;
  }
  Object.defineProperty(_, "init", { value: init });
  Object.defineProperty(_, Symbol.hasInstance, {
    value: (inst) => {
      if (params?.Parent && inst instanceof params.Parent)
        return true;
      return inst?._zod?.traits?.has(name);
    }
  });
  Object.defineProperty(_, "name", { value: name });
  return _;
}
class $ZodAsyncError extends Error {
  constructor() {
    super(`Encountered Promise during synchronous parse. Use .parseAsync() instead.`);
  }
}

class $ZodEncodeError extends Error {
  constructor(name) {
    super(`Encountered unidirectional transform during encode: ${name}`);
    this.name = "ZodEncodeError";
  }
}
(_a = globalThis).__zod_globalConfig ?? (_a.__zod_globalConfig = {});
var globalConfig = globalThis.__zod_globalConfig;
function config(newConfig) {
  if (newConfig)
    Object.assign(globalConfig, newConfig);
  return globalConfig;
}
// node_modules/zod/v4/core/errors.js
function _getMessage() {
  const internals = this._zod;
  internals.message ?? (internals.message = JSON.stringify(internals.def, jsonStringifyReplacer, 2));
  return internals.message;
}
function _setMessage(value) {
  this._zod.message = value;
}
var _messageDesc = {
  get: _getMessage,
  set: _setMessage,
  enumerable: true,
  configurable: true
};
var _issuesDesc = { value: undefined, enumerable: false };
var _installedToString = /* @__PURE__ */ new WeakSet([Object.prototype, Error.prototype]);
var initializer = (inst, def) => {
  inst.name = "$ZodError";
  _issuesDesc.value = def;
  Object.defineProperty(inst, "issues", _issuesDesc);
  _issuesDesc.value = undefined;
  Object.defineProperty(inst, "message", _messageDesc);
  const proto = Object.getPrototypeOf(inst);
  if (!_installedToString.has(proto)) {
    _installedToString.add(proto);
    Object.defineProperty(proto, "toString", {
      configurable: true,
      enumerable: false,
      get() {
        const value = () => this.message;
        Object.defineProperty(this, "toString", { value, configurable: true, writable: true });
        return value;
      },
      set(value) {
        Object.defineProperty(this, "toString", { value, configurable: true, writable: true });
      }
    });
  }
};
var $ZodError = $constructor("$ZodError", initializer);
var $ZodRealError = $constructor("$ZodError", initializer, undefined, {
  Parent: Error
});
function node(obj, key, make) {
  if (!Object.prototype.hasOwnProperty.call(obj, key)) {
    if (key === "__proto__") {
      Object.defineProperty(obj, key, { value: make(), writable: true, enumerable: true, configurable: true });
    } else {
      obj[key] = make();
    }
  }
  return obj[key];
}
function flattenError(error, mapper = (issue) => issue.message) {
  const fieldErrors = {};
  const formErrors = [];
  for (const sub of error.issues) {
    if (sub.path.length > 0) {
      node(fieldErrors, sub.path[0], () => []).push(mapper(sub));
    } else {
      formErrors.push(mapper(sub));
    }
  }
  return { formErrors, fieldErrors };
}
function formatError(error, mapper = (issue) => issue.message) {
  const fieldErrors = { _errors: [] };
  const processError = (error, path = []) => {
    for (const issue of error.issues) {
      if (issue.code === "invalid_union" && issue.errors.length) {
        issue.errors.map((issues) => processError({ issues }, [...path, ...issue.path]));
      } else if (issue.code === "invalid_key") {
        processError({ issues: issue.issues }, [...path, ...issue.path]);
      } else if (issue.code === "invalid_element") {
        processError({ issues: issue.issues }, [...path, ...issue.path]);
      } else {
        const fullpath = [...path, ...issue.path];
        if (fullpath.length === 0) {
          fieldErrors._errors.push(mapper(issue));
        } else {
          let curr = fieldErrors;
          let i = 0;
          while (i < fullpath.length) {
            const el = fullpath[i];
            const terminal = i === fullpath.length - 1;
            if (el === "_errors") {
              if (terminal)
                curr._errors.push(mapper(issue));
              i++;
              continue;
            }
            if (!Object.prototype.hasOwnProperty.call(curr, el)) {
              Object.defineProperty(curr, el, {
                value: { _errors: [] },
                enumerable: true,
                writable: true,
                configurable: true
              });
            }
            const node = curr[el];
            if (terminal) {
              node._errors.push(mapper(issue));
            }
            curr = node;
            i++;
          }
        }
      }
    }
  };
  processError(error);
  return fieldErrors;
}

// node_modules/zod/v4/core/parse.js
function finalizeParams(callee, params) {
  return { callee: params?.callee ?? callee, Err: params?.Err };
}
var _parse = (_Err) => {
  const fn = (schema, value, _ctx, _params) => {
    const ctx = _ctx ? { ..._ctx, async: false } : { async: false };
    const result = schema._zod.run({ value, issues: [] }, ctx);
    if (result instanceof Promise) {
      throw new $ZodAsyncError;
    }
    if (result.issues.length) {
      const e = new (_params?.Err ?? _Err)(result.issues.map((iss) => finalizeIssue(iss, ctx, config())));
      captureStackTrace(e, _params?.callee ?? fn);
      throw e;
    }
    return result.value;
  };
  return fn;
};
var parse = /* @__PURE__ */ _parse($ZodRealError);
var _parseAsync = (_Err) => {
  const fn = async (schema, value, _ctx, params) => {
    const ctx = _ctx ? { ..._ctx, async: true } : { async: true };
    let result = schema._zod.run({ value, issues: [] }, ctx);
    if (result instanceof Promise)
      result = await result;
    if (result.issues.length) {
      const e = new (params?.Err ?? _Err)(result.issues.map((iss) => finalizeIssue(iss, ctx, config())));
      captureStackTrace(e, params?.callee ?? fn);
      throw e;
    }
    return result.value;
  };
  return fn;
};
var parseAsync = /* @__PURE__ */ _parseAsync($ZodRealError);
var _safeParse = (_Err) => (schema, value, _ctx) => {
  const ctx = _ctx ? { ..._ctx, async: false } : { async: false };
  const result = schema._zod.run({ value, issues: [] }, ctx);
  if (result instanceof Promise) {
    throw new $ZodAsyncError;
  }
  return result.issues.length ? failure(_Err, result.issues, ctx) : { success: true, data: result.value };
};
function failure(Err, issues, ctx) {
  let error;
  return {
    success: false,
    get error() {
      if (!error) {
        error = new Err(issues.map((iss) => finalizeIssue(iss, ctx, config())));
        issues = undefined;
        ctx = undefined;
      }
      return error;
    },
    set error(e) {
      error = e;
      issues = undefined;
      ctx = undefined;
    }
  };
}
var _safeParseAsync = (_Err) => async (schema, value, _ctx) => {
  const ctx = _ctx ? { ..._ctx, async: true } : { async: true };
  let result = schema._zod.run({ value, issues: [] }, ctx);
  if (result instanceof Promise)
    result = await result;
  return result.issues.length ? failure(_Err, result.issues, ctx) : { success: true, data: result.value };
};
var COMPILE_INVALID = /* @__PURE__ */ Symbol.for("zod.compile.invalid");
var COMPILE_FALLBACK = /* @__PURE__ */ Symbol.for("zod.compile.fallback");
var validate = (schema, value, _ctx) => {
  const validator = schema._zod.bag.validator;
  if (validator !== undefined) {
    if (validator(value) !== COMPILE_INVALID)
      return true;
    if (validator.definite === true && _ctx === undefined)
      return false;
  }
  return validateFallback(schema, value, _ctx);
};
function validateFallback(schema, value, _ctx) {
  const ctx = _ctx ? { ..._ctx, async: false, abortEarly: true } : { async: false, abortEarly: true };
  const fallbackRun = schema._zod.bag.fallbackRun;
  let result;
  if (fallbackRun) {
    ctx[COMPILE_FALLBACK] = true;
    result = fallbackRun({ value, issues: [] }, ctx);
  } else {
    result = schema._zod.run({ value, issues: [] }, ctx);
  }
  if (result instanceof Promise) {
    throw new $ZodAsyncError;
  }
  return result.issues.length === 0;
}
var validateAsync = async (schema, value, _ctx) => {
  const ctx = _ctx ? { ..._ctx, async: true, abortEarly: true } : { async: true, abortEarly: true };
  let result = schema._zod.run({ value, issues: [] }, ctx);
  if (result instanceof Promise)
    result = await result;
  return result.issues.length === 0;
};
var _encode = (_Err) => {
  const parse = _parse(_Err);
  const fn = (schema, value, _ctx, _params) => {
    const ctx = _ctx ? { ..._ctx, direction: "backward" } : { direction: "backward" };
    return parse(schema, value, ctx, finalizeParams(fn, _params));
  };
  return fn;
};
var _decode = (_Err) => {
  const parse = _parse(_Err);
  const fn = (schema, value, _ctx, _params) => {
    return parse(schema, value, _ctx, finalizeParams(fn, _params));
  };
  return fn;
};
var _encodeAsync = (_Err) => {
  const parseAsync = _parseAsync(_Err);
  const fn = async (schema, value, _ctx, _params) => {
    const ctx = _ctx ? { ..._ctx, direction: "backward" } : { direction: "backward" };
    return await parseAsync(schema, value, ctx, finalizeParams(fn, _params));
  };
  return fn;
};
var _decodeAsync = (_Err) => {
  const parseAsync = _parseAsync(_Err);
  const fn = async (schema, value, _ctx, _params) => {
    return await parseAsync(schema, value, _ctx, finalizeParams(fn, _params));
  };
  return fn;
};
var _safeEncode = (_Err) => (schema, value, _ctx) => {
  const ctx = _ctx ? { ..._ctx, direction: "backward" } : { direction: "backward" };
  return _safeParse(_Err)(schema, value, ctx);
};
var _safeDecode = (_Err) => (schema, value, _ctx) => {
  return _safeParse(_Err)(schema, value, _ctx);
};
var _safeEncodeAsync = (_Err) => async (schema, value, _ctx) => {
  const ctx = _ctx ? { ..._ctx, direction: "backward" } : { direction: "backward" };
  return _safeParseAsync(_Err)(schema, value, ctx);
};
var _safeDecodeAsync = (_Err) => async (schema, value, _ctx) => {
  return _safeParseAsync(_Err)(schema, value, _ctx);
};
// node_modules/zod/v4/core/regexes.js
var cuid = /^[cC][0-9a-z]{6,}$/;
var cuid2 = /^[0-9a-z]+$/;
var ulid = /^[0-7][0-9A-HJKMNP-TV-Za-hjkmnp-tv-z]{25}$/;
var xid = /^[0-9a-vA-V]{20}$/;
var ksuid = /^[A-Za-z0-9]{27}$/;
var nanoid = /^[a-zA-Z0-9_-]{21}$/;
function nanoidOfLength(length) {
  return new RegExp(`^[a-zA-Z0-9_-]{${length}}$`);
}
var duration = /^P(?:(\d+W)|(?!.*W)(?=\d|T\d)(\d+Y)?(\d+M)?(\d+D)?(T(?=\d)(\d+H)?(\d+M)?(\d+([.,]\d+)?S)?)?)$/;
var extendedDuration = /^[-+]?P(?!$)(?:(?:[-+]?\d+Y)|(?:[-+]?\d+[.,]\d+Y$))?(?:(?:[-+]?\d+M)|(?:[-+]?\d+[.,]\d+M$))?(?:(?:[-+]?\d+W)|(?:[-+]?\d+[.,]\d+W$))?(?:(?:[-+]?\d+D)|(?:[-+]?\d+[.,]\d+D$))?(?:T(?=[\d+-])(?:(?:[-+]?\d+H)|(?:[-+]?\d+[.,]\d+H$))?(?:(?:[-+]?\d+M)|(?:[-+]?\d+[.,]\d+M$))?(?:[-+]?\d+(?:[.,]\d+)?S)?)??$/;
var guid = /^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})$/;
var uuid = (version) => {
  if (!version)
    return /^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-8][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}|00000000-0000-0000-0000-000000000000|ffffffff-ffff-ffff-ffff-ffffffffffff)$/;
  return new RegExp(`^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-${version}[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12})$`);
};
var uuid4 = /* @__PURE__ */ uuid(4);
var uuid6 = /* @__PURE__ */ uuid(6);
var uuid7 = /* @__PURE__ */ uuid(7);
var email = /^(?:[A-Za-z0-9_'+\-]+\.)*[A-Za-z0-9_'+\-]*[A-Za-z0-9_+-]@(?:[A-Za-z0-9][A-Za-z0-9\-]*\.)+[A-Za-z]{2,}$/;
var html5Email = /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/;
var rfc5322Email = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
var unicodeEmail = /^[^\s@"]{1,64}@[^\s@]{1,255}$/u;
var idnEmail = unicodeEmail;
var browserEmail = /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/;
var _emoji = `^(?=[\\s\\S]*[\\p{Extended_Pictographic}\\p{Regional_Indicator}\\u20E3])[\\p{Extended_Pictographic}\\p{Emoji_Component}]+$`;
function emoji() {
  return new RegExp(_emoji, "u");
}
var ipv4 = /^(?:(?:25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\.){3}(?:25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])$/;
var ipv6 = /^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:))$/;
var mac = (delimiter) => {
  const escapedDelim = escapeRegex(delimiter ?? ":");
  return new RegExp(`^(?:[0-9A-F]{2}${escapedDelim}){5}[0-9A-F]{2}$|^(?:[0-9a-f]{2}${escapedDelim}){5}[0-9a-f]{2}$`);
};
var cidrv4 = /^((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\/([0-9]|[1-2][0-9]|3[0-2])$/;
var cidrv6 = /^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:))\/(12[0-8]|1[01][0-9]|[1-9]?[0-9])$/;
var base64 = /^$|^(?:[0-9a-zA-Z+/]{4})*(?:(?:[0-9a-zA-Z+/]{2}==)|(?:[0-9a-zA-Z+/]{3}=))?$/;
var base64url = /^(?:[A-Za-z0-9_-]{4})*(?:[A-Za-z0-9_-]{2,3})?$/;
var hostname = /^(?=.{1,253}\.?$)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[-0-9a-zA-Z]{0,61}[0-9a-zA-Z])?)*\.?$/;
var domain = /^(?=.{1,253}$)([a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,63}$/;
var httpProtocol = /^https?$/;
var e164 = /^\+[1-9]\d{6,14}$/;
var creditCard = /^\d(?:[ -]?\d){11,18}$/;
var currencyCode = /^(?:AED|AFN|ALL|AMD|AOA|ARS|AUD|AWG|AZN|BAM|BBD|BDT|BHD|BIF|BMD|BND|BOB|BOV|BRL|BSD|BTN|BWP|BYN|BZD|CAD|CDF|CHE|CHF|CHW|CLF|CLP|CNY|COP|COU|CRC|CUP|CVE|CZK|DJF|DKK|DOP|DZD|EGP|ERN|ETB|EUR|FJD|FKP|GBP|GEL|GHS|GIP|GMD|GNF|GTQ|GYD|HKD|HNL|HTG|HUF|IDR|ILS|INR|IQD|IRR|ISK|JMD|JOD|JPY|KES|KGS|KHR|KMF|KPW|KRW|KWD|KYD|KZT|LAK|LBP|LKR|LRD|LSL|LYD|MAD|MDL|MGA|MKD|MMK|MNT|MOP|MRU|MUR|MVR|MWK|MXN|MXV|MYR|MZN|NAD|NGN|NIO|NOK|NPR|NZD|OMR|PAB|PEN|PGK|PHP|PKR|PLN|PYG|QAR|RON|RSD|RUB|RWF|SAR|SBD|SCR|SDG|SEK|SGD|SHP|SLE|SOS|SRD|SSP|STN|SVC|SYP|SZL|THB|TJS|TMT|TND|TOP|TRY|TTD|TWD|TZS|UAH|UGX|USD|USN|UYI|UYU|UYW|UZS|VED|VES|VND|VUV|WST|XAD|XAF|XAG|XAU|XBA|XBB|XBC|XBD|XCD|XCG|XDR|XOF|XPD|XPF|XPT|XSU|XTS|XUA|XXX|YER|ZAR|ZMW|ZWG)$/;
var iban = /^[A-Z]{2}(?!00|01|99)\d{2}[A-Z0-9]{11,30}$/;
var dateSource = `(?:(?:\\d\\d[2468][048]|\\d\\d[13579][26]|\\d\\d0[48]|[02468][048]00|[13579][26]00)-02-29|\\d{4}-(?:(?:0[13578]|1[02])-(?:0[1-9]|[12]\\d|3[01])|(?:0[469]|11)-(?:0[1-9]|[12]\\d|30)|(?:02)-(?:0[1-9]|1\\d|2[0-8])))`;
function anchor(source) {
  return new RegExp(`^${source}$`);
}
var date = /* @__PURE__ */ anchor(dateSource);
function timeSource(args) {
  const hhmm = `(?:[01]\\d|2[0-3]):[0-5]\\d`;
  const regex = typeof args.precision === "number" ? args.precision === -1 ? `${hhmm}` : args.precision === 0 ? `${hhmm}:[0-5]\\d` : `${hhmm}:[0-5]\\d\\.\\d{${args.precision}}` : args.seconds ? `${hhmm}:[0-5]\\d(?:\\.\\d+)?` : `${hhmm}(?::[0-5]\\d(?:\\.\\d+)?)?`;
  return regex;
}
function time2(args) {
  return new RegExp(`^${timeSource(args)}$`);
}
function datetime(args) {
  const opts = ["Z"];
  if (args.offset)
    opts.push(`([+-](?:[01]\\d|2[0-3]):[0-5]\\d)`);
  const qualified = `${timeSource({ precision: args.precision, seconds: true })}(?:${opts.join("|")})`;
  const timeRegex = args.local ? `${qualified}|${timeSource({ precision: args.precision })}` : qualified;
  return new RegExp(`^${dateSource}T(?:${timeRegex})$`);
}
var anyString = /^[\s\S]{0,}$/;
var string = (params) => {
  const regex = params ? `[\\s\\S]{${params?.minimum ?? 0},${params?.maximum ?? ""}}` : `[\\s\\S]*`;
  return new RegExp(`^${regex}$`);
};
var bigint = /^-?\d+n?$/;
var integer = /^-?\d+$/;
var number = /^-?\d+(?:\.\d+)?$/;
var boolean = /^(?:true|false)$/i;
var _null = /^null$/i;
var _undefined = /^undefined$/i;
var lowercase = /^[^A-Z]*$/;
var uppercase = /^[^a-z]*$/;
var hex = /^[0-9a-fA-F]*$/;
function fixedBase64(bodyLength, padding) {
  return new RegExp(`^[A-Za-z0-9+/]{${bodyLength}}${padding}$`);
}
function fixedBase64url(length) {
  return new RegExp(`^[A-Za-z0-9_-]{${length}}$`);
}
var md5_hex = /^[0-9a-fA-F]{32}$/;
var md5_base64 = /* @__PURE__ */ fixedBase64(22, "==");
var md5_base64url = /* @__PURE__ */ fixedBase64url(22);
var sha1_hex = /^[0-9a-fA-F]{40}$/;
var sha1_base64 = /* @__PURE__ */ fixedBase64(27, "=");
var sha1_base64url = /* @__PURE__ */ fixedBase64url(27);
var sha256_hex = /^[0-9a-fA-F]{64}$/;
var sha256_base64 = /* @__PURE__ */ fixedBase64(43, "=");
var sha256_base64url = /* @__PURE__ */ fixedBase64url(43);
var sha384_hex = /^[0-9a-fA-F]{96}$/;
var sha384_base64 = /* @__PURE__ */ fixedBase64(64, "");
var sha384_base64url = /* @__PURE__ */ fixedBase64url(64);
var sha512_hex = /^[0-9a-fA-F]{128}$/;
var sha512_base64 = /* @__PURE__ */ fixedBase64(86, "==");
var sha512_base64url = /* @__PURE__ */ fixedBase64url(86);

// node_modules/zod/v4/core/checks.js
var $ZodCheck = /* @__PURE__ */ $constructor("$ZodCheck", (inst, def) => {
  var _a;
  inst._zod ?? (inst._zod = {});
  inst._zod.def = def;
  (_a = inst._zod).onattach ?? (_a.onattach = []);
});
var _whenHasSize = (payload) => {
  const val = payload.value;
  return !nullish(val) && val.size !== undefined;
};
var _whenHasLength = (payload) => {
  const val = payload.value;
  return !nullish(val) && val.length !== undefined;
};
var numericOriginMap = {
  number: "number",
  bigint: "bigint",
  object: "date"
};
var $ZodCheckLessThan = /* @__PURE__ */ $constructor("$ZodCheckLessThan", (inst, def) => {
  $ZodCheck.init(inst, def);
  const origin = numericOriginMap[typeof def.value];
  inst._zod.check = (payload) => {
    if (def.inclusive ? payload.value <= def.value : payload.value < def.value) {
      return;
    }
    payload.issues.push({
      origin: numericOriginMap[typeof payload.value] ?? origin,
      code: "too_big",
      maximum: typeof def.value === "object" ? def.value.getTime() : def.value,
      input: payload.value,
      inclusive: def.inclusive,
      inst,
      continue: !def.abort
    });
  };
});
var $ZodCheckGreaterThan = /* @__PURE__ */ $constructor("$ZodCheckGreaterThan", (inst, def) => {
  $ZodCheck.init(inst, def);
  const origin = numericOriginMap[typeof def.value];
  inst._zod.check = (payload) => {
    if (def.inclusive ? payload.value >= def.value : payload.value > def.value) {
      return;
    }
    payload.issues.push({
      origin: numericOriginMap[typeof payload.value] ?? origin,
      code: "too_small",
      minimum: typeof def.value === "object" ? def.value.getTime() : def.value,
      input: payload.value,
      inclusive: def.inclusive,
      inst,
      continue: !def.abort
    });
  };
});
var $ZodCheckMultipleOf = /* @__PURE__ */ $constructor("$ZodCheckMultipleOf", (inst, def) => {
  $ZodCheck.init(inst, def);
  inst._zod.check = (payload) => {
    if (typeof payload.value !== typeof def.value)
      throw new Error("Cannot mix number and bigint in multiple_of check.");
    const isMultiple = typeof payload.value === "bigint" ? def.value !== BigInt(0) && payload.value % def.value === BigInt(0) : floatSafeRemainder(payload.value, def.value) === 0;
    if (isMultiple)
      return;
    payload.issues.push({
      origin: typeof payload.value,
      code: "not_multiple_of",
      divisor: def.value,
      input: payload.value,
      inst,
      continue: !def.abort
    });
  };
});
var $ZodCheckNumberFormat = /* @__PURE__ */ $constructor("$ZodCheckNumberFormat", (inst, def) => {
  $ZodCheck.init(inst, def);
  def.format = def.format || "float64";
  const isInt = def.format?.includes("int");
  const origin = isInt ? "int" : "number";
  const [minimum, maximum] = NUMBER_FORMAT_RANGES[def.format];
  inst._zod.check = (payload) => {
    const input = payload.value;
    if (isInt) {
      if (!Number.isInteger(input)) {
        payload.issues.push({
          expected: origin,
          format: def.format,
          code: "invalid_type",
          continue: false,
          input,
          inst
        });
        return;
      }
      if (!Number.isSafeInteger(input)) {
        if (input > 0) {
          payload.issues.push({
            input,
            code: "too_big",
            maximum: Number.MAX_SAFE_INTEGER,
            note: "Integers must be within the safe integer range.",
            inst,
            origin,
            inclusive: true,
            continue: !def.abort
          });
        } else {
          payload.issues.push({
            input,
            code: "too_small",
            minimum: Number.MIN_SAFE_INTEGER,
            note: "Integers must be within the safe integer range.",
            inst,
            origin,
            inclusive: true,
            continue: !def.abort
          });
        }
        return;
      }
    }
    if (input < minimum) {
      payload.issues.push({
        origin: "number",
        input,
        code: "too_small",
        minimum,
        inclusive: true,
        inst,
        continue: !def.abort
      });
    }
    if (input > maximum) {
      payload.issues.push({
        origin: "number",
        input,
        code: "too_big",
        maximum,
        inclusive: true,
        inst,
        continue: !def.abort
      });
    }
  };
});
var $ZodCheckBigIntFormat = /* @__PURE__ */ $constructor("$ZodCheckBigIntFormat", (inst, def) => {
  $ZodCheck.init(inst, def);
  const [minimum, maximum] = BIGINT_FORMAT_RANGES[def.format];
  inst._zod.check = (payload) => {
    const input = payload.value;
    if (input < minimum) {
      payload.issues.push({
        origin: "bigint",
        input,
        code: "too_small",
        minimum,
        inclusive: true,
        inst,
        continue: !def.abort
      });
    }
    if (input > maximum) {
      payload.issues.push({
        origin: "bigint",
        input,
        code: "too_big",
        maximum,
        inclusive: true,
        inst,
        continue: !def.abort
      });
    }
  };
});
var $ZodCheckMaxSize = /* @__PURE__ */ $constructor("$ZodCheckMaxSize", (inst, def) => {
  var _a;
  $ZodCheck.init(inst, def);
  (_a = inst._zod.def).when ?? (_a.when = _whenHasSize);
  inst._zod.check = (payload) => {
    const input = payload.value;
    const size = input.size;
    if (size <= def.maximum)
      return;
    payload.issues.push({
      origin: getSizableOrigin(input),
      code: "too_big",
      maximum: def.maximum,
      inclusive: true,
      input,
      inst,
      continue: !def.abort
    });
  };
});
var $ZodCheckMinSize = /* @__PURE__ */ $constructor("$ZodCheckMinSize", (inst, def) => {
  var _a;
  $ZodCheck.init(inst, def);
  (_a = inst._zod.def).when ?? (_a.when = _whenHasSize);
  inst._zod.check = (payload) => {
    const input = payload.value;
    const size = input.size;
    if (size >= def.minimum)
      return;
    payload.issues.push({
      origin: getSizableOrigin(input),
      code: "too_small",
      minimum: def.minimum,
      inclusive: true,
      input,
      inst,
      continue: !def.abort
    });
  };
});
var $ZodCheckSizeEquals = /* @__PURE__ */ $constructor("$ZodCheckSizeEquals", (inst, def) => {
  var _a;
  $ZodCheck.init(inst, def);
  (_a = inst._zod.def).when ?? (_a.when = _whenHasSize);
  inst._zod.check = (payload) => {
    const input = payload.value;
    const size = input.size;
    if (size === def.size)
      return;
    const tooBig = size > def.size;
    payload.issues.push({
      origin: getSizableOrigin(input),
      ...tooBig ? { code: "too_big", maximum: def.size } : { code: "too_small", minimum: def.size },
      inclusive: true,
      exact: true,
      input: payload.value,
      inst,
      continue: !def.abort
    });
  };
});
var $ZodCheckMaxLength = /* @__PURE__ */ $constructor("$ZodCheckMaxLength", (inst, def) => {
  var _a;
  $ZodCheck.init(inst, def);
  (_a = inst._zod.def).when ?? (_a.when = _whenHasLength);
  inst._zod.check = (payload) => {
    const input = payload.value;
    const units = input.length;
    const length = typeof input === "string" && units > def.maximum ? codePointLength(input) : units;
    if (length <= def.maximum)
      return;
    const origin = getLengthableOrigin(input);
    payload.issues.push({
      origin,
      code: "too_big",
      maximum: def.maximum,
      inclusive: true,
      input,
      inst,
      continue: !def.abort
    });
  };
});
var $ZodCheckMinLength = /* @__PURE__ */ $constructor("$ZodCheckMinLength", (inst, def) => {
  var _a;
  $ZodCheck.init(inst, def);
  (_a = inst._zod.def).when ?? (_a.when = _whenHasLength);
  inst._zod.check = (payload) => {
    const input = payload.value;
    const units = input.length;
    const length = typeof input === "string" && units >= def.minimum && units < def.minimum * 2 ? codePointLength(input) : units;
    if (length >= def.minimum)
      return;
    const origin = getLengthableOrigin(input);
    payload.issues.push({
      origin,
      code: "too_small",
      minimum: def.minimum,
      inclusive: true,
      input,
      inst,
      continue: !def.abort
    });
  };
});
var $ZodCheckLengthEquals = /* @__PURE__ */ $constructor("$ZodCheckLengthEquals", (inst, def) => {
  var _a;
  $ZodCheck.init(inst, def);
  (_a = inst._zod.def).when ?? (_a.when = _whenHasLength);
  inst._zod.check = (payload) => {
    const input = payload.value;
    const units = input.length;
    const length = typeof input === "string" && units >= def.length && units <= def.length * 2 ? codePointLength(input) : units;
    if (length === def.length)
      return;
    const origin = getLengthableOrigin(input);
    const tooBig = length > def.length;
    payload.issues.push({
      origin,
      ...tooBig ? { code: "too_big", maximum: def.length } : { code: "too_small", minimum: def.length },
      inclusive: true,
      exact: true,
      input: payload.value,
      inst,
      continue: !def.abort
    });
  };
});
var $ZodCheckStringFormat = /* @__PURE__ */ $constructor("$ZodCheckStringFormat", (inst, def) => {
  var _a, _b;
  $ZodCheck.init(inst, def);
  if (def.pattern)
    (_a = inst._zod).check ?? (_a.check = (payload) => {
      def.pattern.lastIndex = 0;
      if (def.pattern.test(payload.value))
        return;
      payload.issues.push({
        origin: "string",
        code: "invalid_format",
        format: def.format,
        input: payload.value,
        ...def.pattern ? { pattern: def.pattern.toString() } : {},
        inst,
        continue: !def.abort
      });
    });
  else
    (_b = inst._zod).check ?? (_b.check = () => {});
});
var $ZodCheckRegex = /* @__PURE__ */ $constructor("$ZodCheckRegex", (inst, def) => {
  $ZodCheckStringFormat.init(inst, def);
  inst._zod.check = (payload) => {
    def.pattern.lastIndex = 0;
    if (def.pattern.test(payload.value))
      return;
    payload.issues.push({
      origin: "string",
      code: "invalid_format",
      format: "regex",
      input: payload.value,
      pattern: def.pattern.toString(),
      inst,
      continue: !def.abort
    });
  };
});
var $ZodCheckLowerCase = /* @__PURE__ */ $constructor("$ZodCheckLowerCase", (inst, def) => {
  def.pattern ?? (def.pattern = lowercase);
  $ZodCheckStringFormat.init(inst, def);
});
var $ZodCheckUpperCase = /* @__PURE__ */ $constructor("$ZodCheckUpperCase", (inst, def) => {
  def.pattern ?? (def.pattern = uppercase);
  $ZodCheckStringFormat.init(inst, def);
});
var $ZodCheckIncludes = /* @__PURE__ */ $constructor("$ZodCheckIncludes", (inst, def) => {
  $ZodCheck.init(inst, def);
  const escapedRegex = escapeRegex(def.includes);
  const pattern = new RegExp(typeof def.position === "number" ? `^.{${def.position},}${escapedRegex}` : escapedRegex);
  def.pattern = pattern;
  inst._zod.check = (payload) => {
    if (payload.value.includes(def.includes, def.position))
      return;
    payload.issues.push({
      origin: "string",
      code: "invalid_format",
      format: "includes",
      includes: def.includes,
      input: payload.value,
      inst,
      continue: !def.abort
    });
  };
});
var $ZodCheckStartsWith = /* @__PURE__ */ $constructor("$ZodCheckStartsWith", (inst, def) => {
  $ZodCheck.init(inst, def);
  const pattern = new RegExp(`^${escapeRegex(def.prefix)}.*`);
  def.pattern ?? (def.pattern = pattern);
  inst._zod.check = (payload) => {
    if (payload.value.startsWith(def.prefix))
      return;
    payload.issues.push({
      origin: "string",
      code: "invalid_format",
      format: "starts_with",
      prefix: def.prefix,
      input: payload.value,
      inst,
      continue: !def.abort
    });
  };
});
var $ZodCheckEndsWith = /* @__PURE__ */ $constructor("$ZodCheckEndsWith", (inst, def) => {
  $ZodCheck.init(inst, def);
  const pattern = new RegExp(`.*${escapeRegex(def.suffix)}$`);
  def.pattern ?? (def.pattern = pattern);
  inst._zod.check = (payload) => {
    if (payload.value.endsWith(def.suffix))
      return;
    payload.issues.push({
      origin: "string",
      code: "invalid_format",
      format: "ends_with",
      suffix: def.suffix,
      input: payload.value,
      inst,
      continue: !def.abort
    });
  };
});
function handleCheckPropertyResult(result, payload, property) {
  if (result.issues.length) {
    payload.issues.push(...prefixIssues(property, result.issues));
  }
}
var $ZodCheckProperty = /* @__PURE__ */ $constructor("$ZodCheckProperty", (inst, def) => {
  $ZodCheck.init(inst, def);
  inst._zod.check = (payload) => {
    const result = def.schema._zod.run({
      value: payload.value[def.property],
      issues: []
    }, {});
    if (result instanceof Promise) {
      return result.then((result) => handleCheckPropertyResult(result, payload, def.property));
    }
    handleCheckPropertyResult(result, payload, def.property);
    return;
  };
});
var $ZodCheckProperties = /* @__PURE__ */ $constructor("$ZodCheckProperties", (inst, def) => {
  $ZodCheck.init(inst, def);
  hide(inst, Symbol.iterator, function* () {
    yield inst;
  });
  let entries;
  inst._zod.check = (payload) => {
    if (payload.value == null) {
      payload.issues.push({ expected: "object", code: "invalid_type", input: payload.value, inst });
      return;
    }
    entries ?? (entries = Reflect.ownKeys(def.shape).map((key) => [key, def.shape[key]]));
    const input = payload.value;
    let proms;
    for (const [key, schema] of entries) {
      const result = schema._zod.run({ value: input[key], issues: [] }, {});
      if (result instanceof Promise) {
        proms ?? (proms = []);
        proms.push(result.then((result) => handleCheckPropertyResult(result, payload, key)));
      } else {
        handleCheckPropertyResult(result, payload, key);
      }
    }
    if (proms)
      return Promise.all(proms).then(() => {
        return;
      });
    return;
  };
});
var $ZodCheckMimeType = /* @__PURE__ */ $constructor("$ZodCheckMimeType", (inst, def) => {
  $ZodCheck.init(inst, def);
  const mimeSet = new Set(def.mime);
  inst._zod.check = (payload) => {
    if (mimeSet.has(payload.value.type))
      return;
    payload.issues.push({
      code: "invalid_value",
      values: def.mime,
      input: payload.value.type,
      inst,
      continue: !def.abort
    });
  };
});
var $ZodCheckOverwrite = /* @__PURE__ */ $constructor("$ZodCheckOverwrite", (inst, def) => {
  $ZodCheck.init(inst, def);
  inst._zod.check = (payload) => {
    payload.value = def.tx(payload.value);
  };
});

// node_modules/zod/v4/core/doc.js
class Doc {
  constructor(args = [], closed = {}) {
    this.content = [];
    this.indent = 0;
    this.args = args;
    this.closed = closed;
  }
  indented(fn) {
    this.indent += 1;
    try {
      fn(this);
    } finally {
      this.indent -= 1;
    }
  }
  write(arg) {
    if (typeof arg === "function") {
      arg(this, { execution: "sync" });
      arg(this, { execution: "async" });
      return;
    }
    const content = arg;
    const lines = content.split(`
`).filter((x) => x);
    const minIndent = Math.min(...lines.map((x) => x.length - x.trimStart().length));
    const dedented = lines.map((x) => x.slice(minIndent)).map((x) => " ".repeat(this.indent * 2) + x);
    for (const line of dedented) {
      this.content.push(line);
    }
  }
  compile() {
    const F = Function;
    const content = this?.content ?? [``];
    const factory = new F(...Object.keys(this.closed), `return function (${this.args.join(", ")}) {
${content.join(`
`)}
};`);
    return factory(...Object.values(this.closed));
  }
}

// node_modules/zod/v4/core/versions.js
var version = {
  major: 4,
  minor: 6,
  patch: 6
};

// node_modules/zod/v4/core/schemas.js
var $ZodType = /* @__PURE__ */ $constructor("$ZodType", (inst, def) => {
  var _a;
  inst ?? (inst = {});
  inst._zod.def = def;
  inst._zod.bag = inst._zod.bag || {};
  inst._zod.version = version;
  const defChecks = inst._zod.def.checks;
  const checks = inst._zod.traits.has("$ZodCheck") ? [inst, ...defChecks ?? []] : defChecks?.length ? [...defChecks] : [];
  for (const ch of checks) {
    for (const fn of ch._zod.onattach) {
      fn(inst);
    }
  }
  if (checks.length === 0) {
    (_a = inst._zod).deferred ?? (_a.deferred = []);
    inst._zod.deferred?.push(() => {
      inst._zod.run = inst._zod.parse;
    });
  } else {
    const runChecks = (payload, checks, ctx) => {
      if (payload.memo)
        return payload;
      let isAborted = aborted(payload);
      let asyncResult;
      for (const ch of checks) {
        if (ch._zod.def.when) {
          if (explicitlyAborted(payload))
            continue;
          const shouldRun = ch._zod.def.when(payload);
          if (!shouldRun)
            continue;
        } else if (isAborted) {
          continue;
        }
        const currLen = payload.issues.length;
        const _ = ch._zod.check(payload);
        if (_ instanceof Promise && ctx?.async === false) {
          throw new $ZodAsyncError;
        }
        if (asyncResult || _ instanceof Promise) {
          asyncResult = (asyncResult ?? Promise.resolve()).then(async () => {
            await _;
            const nextLen = payload.issues.length;
            if (nextLen === currLen)
              return;
            attachSchema(payload.issues, currLen, inst);
            if (!isAborted)
              isAborted = aborted(payload, currLen);
          });
        } else {
          const nextLen = payload.issues.length;
          if (nextLen === currLen)
            continue;
          attachSchema(payload.issues, currLen, inst);
          if (!isAborted)
            isAborted = aborted(payload, currLen);
        }
      }
      if (asyncResult) {
        return asyncResult.then(() => {
          return payload;
        });
      }
      return payload;
    };
    const handleCanaryResult = (canary, payload, ctx) => {
      if (aborted(canary)) {
        canary.aborted = true;
        return canary;
      }
      const checkResult = runChecks(payload, checks, ctx);
      if (checkResult instanceof Promise) {
        if (ctx.async === false)
          throw new $ZodAsyncError;
        return checkResult.then((checkResult) => inst._zod.parse(checkResult, ctx));
      }
      return inst._zod.parse(checkResult, ctx);
    };
    inst._zod.run = (payload, ctx) => {
      if (ctx.skipChecks) {
        return inst._zod.parse(payload, ctx);
      }
      if (ctx.direction === "backward") {
        const canary = inst._zod.parse({ value: payload.value, issues: [] }, { ...ctx, skipChecks: true });
        if (canary instanceof Promise) {
          return canary.then((canary) => {
            return handleCanaryResult(canary, payload, ctx);
          });
        }
        return handleCanaryResult(canary, payload, ctx);
      }
      const result = inst._zod.parse(payload, ctx);
      if (result instanceof Promise) {
        if (ctx.async === false)
          throw new $ZodAsyncError;
        return result.then((result) => runChecks(result, checks, ctx));
      }
      return runChecks(result, checks, ctx);
    };
  }
}, {
  get "~standard"() {
    return hide(this, "~standard", standardProps(this));
  },
  set "~standard"(value) {
    own(this, "~standard", value);
  }
});
var toStandardResult = (r, ctx) => r.issues.length ? { issues: r.issues.map((iss) => finalizeIssue(iss, ctx, config())) } : { value: r.value };
async function validateAsync2(inst, value) {
  const ctx = { async: true };
  return toStandardResult(await inst._zod.run({ value, issues: [] }, ctx), ctx);
}
function standardProps(inst) {
  return {
    validate: (value) => {
      const ctx = { async: false };
      try {
        const r = inst._zod.run({ value, issues: [] }, ctx);
        if (!(r instanceof Promise))
          return toStandardResult(r, ctx);
      } catch (_) {}
      return validateAsync2(inst, value);
    },
    vendor: "zod",
    version: 1
  };
}
var $ZodString = /* @__PURE__ */ $constructor("$ZodString", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.pattern = def.pattern ?? anyString;
  inst._zod.parse = (payload, _) => {
    if (def.coerce)
      try {
        payload.value = String(payload.value);
      } catch (_) {}
    if (typeof payload.value === "string")
      return payload;
    payload.issues.push({
      expected: "string",
      code: "invalid_type",
      input: payload.value,
      inst
    });
    return payload;
  };
});
var $ZodStringFormat = /* @__PURE__ */ $constructor("$ZodStringFormat", (inst, def) => {
  $ZodCheckStringFormat.init(inst, def);
  $ZodString.init(inst, def);
});
var $ZodGUID = /* @__PURE__ */ $constructor("$ZodGUID", (inst, def) => {
  def.pattern ?? (def.pattern = guid);
  $ZodStringFormat.init(inst, def);
});
var $ZodUUID = /* @__PURE__ */ $constructor("$ZodUUID", (inst, def) => {
  if (def.version) {
    const versionMap = {
      v1: 1,
      v2: 2,
      v3: 3,
      v4: 4,
      v5: 5,
      v6: 6,
      v7: 7,
      v8: 8
    };
    const v = versionMap[def.version];
    if (v === undefined)
      throw new Error(`Invalid UUID version: "${def.version}"`);
    def.pattern ?? (def.pattern = uuid(v));
  } else
    def.pattern ?? (def.pattern = uuid());
  $ZodStringFormat.init(inst, def);
});
var $ZodEmail = /* @__PURE__ */ $constructor("$ZodEmail", (inst, def) => {
  def.pattern ?? (def.pattern = email);
  $ZodStringFormat.init(inst, def);
});
var URL_BAD_FORMAT = 1;
var URL_UNPARSEABLE = 2;
function canParseURL(input) {
  try {
    if (typeof URL !== "undefined" && typeof URL.canParse === "function")
      return URL.canParse(input);
    new URL(input);
    return true;
  } catch {
    return false;
  }
}
function validateURL(trimmed, def) {
  if (!("normalize" in def) && !("hostname" in def) && !("protocol" in def)) {
    return canParseURL(trimmed) || URL_UNPARSEABLE;
  }
  return parseURLObject(trimmed, def);
}
function parseURLObject(trimmed, def) {
  if (!def.normalize && def.protocol?.source === httpProtocol.source && !/^https?:\/\//i.test(trimmed)) {
    return URL_BAD_FORMAT;
  }
  try {
    if (typeof URL !== "undefined") {
      const URLStatic = URL;
      if (typeof URLStatic.parse === "function")
        return URLStatic.parse(trimmed) ?? URL_UNPARSEABLE;
    }
    return new URL(trimmed);
  } catch {
    return URL_UNPARSEABLE;
  }
}
var asciiTabOrNewline = /[\t\n\r]/g;
function stripTabAndNewline(value) {
  return value.replace(asciiTabOrNewline, "");
}
function urlHostnameOk(url, hostname) {
  hostname.lastIndex = 0;
  return hostname.test(url.hostname);
}
function urlProtocolOk(url, protocol) {
  protocol.lastIndex = 0;
  return protocol.test(url.protocol.endsWith(":") ? url.protocol.slice(0, -1) : url.protocol);
}
var $ZodURL = /* @__PURE__ */ $constructor("$ZodURL", (inst, def) => {
  $ZodStringFormat.init(inst, def);
  inst._zod.check = (payload) => {
    try {
      const trimmed = payload.value.trim();
      const url = validateURL(trimmed, def);
      if (url === URL_BAD_FORMAT) {
        payload.issues.push({
          code: "invalid_format",
          format: "url",
          note: "Invalid URL format",
          input: payload.value,
          inst,
          continue: !def.abort
        });
        return;
      }
      if (url === URL_UNPARSEABLE) {
        payload.issues.push({
          code: "invalid_format",
          format: "url",
          input: payload.value,
          inst,
          continue: !def.abort
        });
        return;
      }
      if (url === true) {
        payload.value = stripTabAndNewline(trimmed);
        return;
      }
      if (def.hostname && !urlHostnameOk(url, def.hostname)) {
        payload.issues.push({
          code: "invalid_format",
          format: "url",
          note: "Invalid hostname",
          pattern: def.hostname.source,
          input: payload.value,
          inst,
          continue: !def.abort
        });
      }
      if (def.protocol && !urlProtocolOk(url, def.protocol)) {
        payload.issues.push({
          code: "invalid_format",
          format: "url",
          note: "Invalid protocol",
          pattern: def.protocol.source,
          input: payload.value,
          inst,
          continue: !def.abort
        });
      }
      payload.value = def.normalize ? url.href : stripTabAndNewline(trimmed);
      return;
    } catch (_) {
      payload.issues.push({
        code: "invalid_format",
        format: "url",
        input: payload.value,
        inst,
        continue: !def.abort
      });
    }
  };
});
var $ZodEmoji = /* @__PURE__ */ $constructor("$ZodEmoji", (inst, def) => {
  def.pattern ?? (def.pattern = emoji());
  $ZodStringFormat.init(inst, def);
});
var $ZodNanoID = /* @__PURE__ */ $constructor("$ZodNanoID", (inst, def) => {
  if (def.length !== undefined && (!Number.isInteger(def.length) || def.length < 1))
    throw new Error(`Invalid nanoid length: ${def.length}`);
  def.pattern ?? (def.pattern = def.length === undefined ? nanoid : nanoidOfLength(def.length));
  $ZodStringFormat.init(inst, def);
});
var $ZodCUID = /* @__PURE__ */ $constructor("$ZodCUID", (inst, def) => {
  def.pattern ?? (def.pattern = cuid);
  $ZodStringFormat.init(inst, def);
});
var $ZodCUID2 = /* @__PURE__ */ $constructor("$ZodCUID2", (inst, def) => {
  def.pattern ?? (def.pattern = cuid2);
  $ZodStringFormat.init(inst, def);
});
var $ZodULID = /* @__PURE__ */ $constructor("$ZodULID", (inst, def) => {
  def.pattern ?? (def.pattern = ulid);
  $ZodStringFormat.init(inst, def);
});
var $ZodXID = /* @__PURE__ */ $constructor("$ZodXID", (inst, def) => {
  def.pattern ?? (def.pattern = xid);
  $ZodStringFormat.init(inst, def);
});
var $ZodKSUID = /* @__PURE__ */ $constructor("$ZodKSUID", (inst, def) => {
  def.pattern ?? (def.pattern = ksuid);
  $ZodStringFormat.init(inst, def);
});
var $ZodISODateTime = /* @__PURE__ */ $constructor("$ZodISODateTime", (inst, def) => {
  def.pattern ?? (def.pattern = datetime(def));
  $ZodStringFormat.init(inst, def);
});
var $ZodISODate = /* @__PURE__ */ $constructor("$ZodISODate", (inst, def) => {
  def.pattern ?? (def.pattern = date);
  $ZodStringFormat.init(inst, def);
});
var $ZodISOTime = /* @__PURE__ */ $constructor("$ZodISOTime", (inst, def) => {
  def.pattern ?? (def.pattern = time2(def));
  $ZodStringFormat.init(inst, def);
});
var $ZodISODuration = /* @__PURE__ */ $constructor("$ZodISODuration", (inst, def) => {
  def.pattern ?? (def.pattern = duration);
  $ZodStringFormat.init(inst, def);
});
var $ZodIPv4 = /* @__PURE__ */ $constructor("$ZodIPv4", (inst, def) => {
  def.pattern ?? (def.pattern = ipv4);
  $ZodStringFormat.init(inst, def);
});
var ipv6Alphabet = /^[0-9a-fA-F:.]+$/;
function isValidIPv6(value) {
  if (!ipv6Alphabet.test(value))
    return false;
  return canParseURL(`http://[${value}]`);
}
var $ZodIPv6 = /* @__PURE__ */ $constructor("$ZodIPv6", (inst, def) => {
  def.pattern ?? (def.pattern = ipv6);
  $ZodStringFormat.init(inst, def);
  inst._zod.check = (payload) => {
    if (!isValidIPv6(payload.value)) {
      payload.issues.push({
        code: "invalid_format",
        format: "ipv6",
        input: payload.value,
        inst,
        continue: !def.abort
      });
    }
  };
});
var $ZodMAC = /* @__PURE__ */ $constructor("$ZodMAC", (inst, def) => {
  def.pattern ?? (def.pattern = mac(def.delimiter));
  $ZodStringFormat.init(inst, def);
});
var $ZodCIDRv4 = /* @__PURE__ */ $constructor("$ZodCIDRv4", (inst, def) => {
  def.pattern ?? (def.pattern = cidrv4);
  $ZodStringFormat.init(inst, def);
});
function isValidCIDRv6(value) {
  const parts = value.split("/");
  if (parts.length !== 2)
    return false;
  const [address, prefix] = parts;
  if (!prefix)
    return false;
  const prefixNum = Number(prefix);
  if (`${prefixNum}` !== prefix)
    return false;
  if (prefixNum < 0 || prefixNum > 128)
    return false;
  return isValidIPv6(address);
}
var $ZodCIDRv6 = /* @__PURE__ */ $constructor("$ZodCIDRv6", (inst, def) => {
  def.pattern ?? (def.pattern = cidrv6);
  $ZodStringFormat.init(inst, def);
  inst._zod.check = (payload) => {
    if (!isValidCIDRv6(payload.value)) {
      payload.issues.push({
        code: "invalid_format",
        format: "cidrv6",
        input: payload.value,
        inst,
        continue: !def.abort
      });
    }
  };
});
function isValidBase64(data) {
  if (data === "")
    return true;
  if (/\s/.test(data))
    return false;
  if (data.length % 4 !== 0)
    return false;
  try {
    atob(data);
    return true;
  } catch {
    return false;
  }
}
var base64Charset = /^[0-9a-zA-Z+/]*={0,2}$/;
var $ZodBase64 = /* @__PURE__ */ $constructor("$ZodBase64", (inst, def) => {
  def.pattern ?? (def.pattern = base64Charset);
  $ZodStringFormat.init(inst, def);
  inst._zod.check = (payload) => {
    if (isValidBase64(payload.value))
      return;
    payload.issues.push({
      code: "invalid_format",
      format: "base64",
      input: payload.value,
      inst,
      continue: !def.abort
    });
  };
});
var base64urlCharset = /^[A-Za-z0-9_-]*$/;
function isValidBase64URL(data) {
  if (!base64urlCharset.test(data))
    return false;
  const base64 = data.replace(/[-_]/g, (c) => c === "-" ? "+" : "/");
  const padded = base64.padEnd(Math.ceil(base64.length / 4) * 4, "=");
  return isValidBase64(padded);
}
var $ZodBase64URL = /* @__PURE__ */ $constructor("$ZodBase64URL", (inst, def) => {
  def.pattern ?? (def.pattern = base64urlCharset);
  $ZodStringFormat.init(inst, def);
  inst._zod.check = (payload) => {
    if (isValidBase64URL(payload.value))
      return;
    payload.issues.push({
      code: "invalid_format",
      format: "base64url",
      input: payload.value,
      inst,
      continue: !def.abort
    });
  };
});
var $ZodE164 = /* @__PURE__ */ $constructor("$ZodE164", (inst, def) => {
  def.pattern ?? (def.pattern = e164);
  $ZodStringFormat.init(inst, def);
});
var CC_SANITIZE = /[- ]/g;
function isLuhnAlgo(digits) {
  let length = digits.length;
  let bit = 1;
  let sum = 0;
  while (length) {
    const value = digits.charCodeAt(--length) - 48;
    bit ^= 1;
    sum += bit ? [0, 2, 4, 6, 8, 1, 3, 5, 7, 9][value] : value;
  }
  return sum % 10 === 0;
}
function isValidCreditCard(input) {
  if (!creditCard.test(input))
    return false;
  return isLuhnAlgo(input.replace(CC_SANITIZE, ""));
}
var $ZodCreditCard = /* @__PURE__ */ $constructor("$ZodCreditCard", (inst, def) => {
  def.pattern ?? (def.pattern = creditCard);
  $ZodStringFormat.init(inst, def);
  inst._zod.check = (payload) => {
    if (isValidCreditCard(payload.value))
      return;
    payload.issues.push({
      code: "invalid_format",
      format: "credit_card",
      input: payload.value,
      inst,
      continue: !def.abort
    });
  };
});
function isIso7064Mod97(iban2) {
  let remainder = 0;
  const len = iban2.length;
  for (let i = 4;i < len; i++) {
    const code = iban2.charCodeAt(i);
    remainder = (code >= 65 ? remainder * 100 + (code - 55) : remainder * 10 + (code - 48)) % 97;
  }
  for (let i = 0;i < 4; i++) {
    const code = iban2.charCodeAt(i);
    remainder = (code >= 65 ? remainder * 100 + (code - 55) : remainder * 10 + (code - 48)) % 97;
  }
  return remainder === 1;
}
function isValidIBAN(input) {
  if (!iban.test(input))
    return false;
  return isIso7064Mod97(input);
}
var $ZodIBAN = /* @__PURE__ */ $constructor("$ZodIBAN", (inst, def) => {
  def.pattern ?? (def.pattern = iban);
  $ZodStringFormat.init(inst, def);
  inst._zod.check = (payload) => {
    if (isValidIBAN(payload.value))
      return;
    payload.issues.push({
      code: "invalid_format",
      format: "iban",
      input: payload.value,
      inst,
      continue: !def.abort
    });
  };
});
function isValidJWT(token, algorithm = null) {
  try {
    const tokensParts = token.split(".");
    if (tokensParts.length !== 3)
      return false;
    const [header] = tokensParts;
    if (!header)
      return false;
    const parsedHeader = JSON.parse(atob(header));
    if ("typ" in parsedHeader && parsedHeader?.typ !== "JWT")
      return false;
    if (!parsedHeader.alg)
      return false;
    if (algorithm && (!("alg" in parsedHeader) || parsedHeader.alg !== algorithm))
      return false;
    return true;
  } catch {
    return false;
  }
}
var $ZodJWT = /* @__PURE__ */ $constructor("$ZodJWT", (inst, def) => {
  $ZodStringFormat.init(inst, def);
  inst._zod.check = (payload) => {
    if (isValidJWT(payload.value, def.alg))
      return;
    payload.issues.push({
      code: "invalid_format",
      format: "jwt",
      input: payload.value,
      inst,
      continue: !def.abort
    });
  };
});
var $ZodCustomStringFormat = /* @__PURE__ */ $constructor("$ZodCustomStringFormat", (inst, def) => {
  $ZodStringFormat.init(inst, def);
  inst._zod.check = (payload) => {
    if (def.fn(payload.value))
      return;
    payload.issues.push({
      code: "invalid_format",
      format: def.format,
      input: payload.value,
      inst,
      continue: !def.abort
    });
  };
});
var $ZodNumber = /* @__PURE__ */ $constructor("$ZodNumber", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.pattern = number;
  inst._zod.parse = (payload, _ctx) => {
    if (def.coerce)
      try {
        payload.value = Number(payload.value);
      } catch (_) {}
    const input = payload.value;
    if (typeof input === "number" && !Number.isNaN(input) && Number.isFinite(input)) {
      return payload;
    }
    const received = typeof input === "number" ? Number.isNaN(input) ? "NaN" : !Number.isFinite(input) ? String(input) : undefined : undefined;
    payload.issues.push({
      expected: "number",
      code: "invalid_type",
      input,
      inst,
      ...received ? { received } : {}
    });
    return payload;
  };
});
var $ZodNumberFormat = /* @__PURE__ */ $constructor("$ZodNumberFormat", (inst, def) => {
  $ZodCheckNumberFormat.init(inst, def);
  $ZodNumber.init(inst, def);
});
var $ZodBoolean = /* @__PURE__ */ $constructor("$ZodBoolean", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.pattern = boolean;
  inst._zod.parse = (payload, _ctx) => {
    if (def.coerce)
      try {
        payload.value = Boolean(payload.value);
      } catch (_) {}
    const input = payload.value;
    if (typeof input === "boolean")
      return payload;
    payload.issues.push({
      expected: "boolean",
      code: "invalid_type",
      input,
      inst
    });
    return payload;
  };
});
var $ZodBigInt = /* @__PURE__ */ $constructor("$ZodBigInt", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.pattern = bigint;
  inst._zod.parse = (payload, _ctx) => {
    if (def.coerce)
      try {
        payload.value = BigInt(payload.value);
      } catch (_) {}
    if (typeof payload.value === "bigint")
      return payload;
    payload.issues.push({
      expected: "bigint",
      code: "invalid_type",
      input: payload.value,
      inst
    });
    return payload;
  };
});
var $ZodBigIntFormat = /* @__PURE__ */ $constructor("$ZodBigIntFormat", (inst, def) => {
  $ZodCheckBigIntFormat.init(inst, def);
  $ZodBigInt.init(inst, def);
});
var $ZodSymbol = /* @__PURE__ */ $constructor("$ZodSymbol", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.parse = (payload, _ctx) => {
    const input = payload.value;
    if (typeof input === "symbol")
      return payload;
    payload.issues.push({
      expected: "symbol",
      code: "invalid_type",
      input,
      inst
    });
    return payload;
  };
});
var $ZodUndefined = /* @__PURE__ */ $constructor("$ZodUndefined", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.pattern = _undefined;
  inst._zod.values = new Set([undefined]);
  inst._zod.parse = (payload, _ctx) => {
    const input = payload.value;
    if (typeof input === "undefined")
      return payload;
    payload.issues.push({
      expected: "undefined",
      code: "invalid_type",
      input,
      inst
    });
    return payload;
  };
});
var $ZodNull = /* @__PURE__ */ $constructor("$ZodNull", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.pattern = _null;
  inst._zod.values = new Set([null]);
  inst._zod.parse = (payload, _ctx) => {
    const input = payload.value;
    if (input === null)
      return payload;
    payload.issues.push({
      expected: "null",
      code: "invalid_type",
      input,
      inst
    });
    return payload;
  };
});
var $ZodAny = /* @__PURE__ */ $constructor("$ZodAny", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.parse = (payload) => payload;
});
var $ZodUnknown = /* @__PURE__ */ $constructor("$ZodUnknown", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.parse = (payload) => payload;
});
var $ZodNever = /* @__PURE__ */ $constructor("$ZodNever", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.parse = (payload, _ctx) => {
    payload.issues.push({
      expected: "never",
      code: "invalid_type",
      input: payload.value,
      inst
    });
    return payload;
  };
});
var $ZodVoid = /* @__PURE__ */ $constructor("$ZodVoid", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.parse = (payload, _ctx) => {
    const input = payload.value;
    if (typeof input === "undefined")
      return payload;
    payload.issues.push({
      expected: "void",
      code: "invalid_type",
      input,
      inst
    });
    return payload;
  };
});
var $ZodDate = /* @__PURE__ */ $constructor("$ZodDate", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.parse = (payload, _ctx) => {
    if (def.coerce) {
      try {
        payload.value = new Date(payload.value);
      } catch (_err) {}
    }
    const input = payload.value;
    const isDate = input instanceof Date;
    const isValidDate = isDate && !Number.isNaN(input.getTime());
    if (isValidDate)
      return payload;
    payload.issues.push({
      expected: "date",
      code: "invalid_type",
      input,
      ...isDate ? { received: "Invalid Date" } : {},
      inst
    });
    return payload;
  };
});
function capSize(inst, size) {
  const caps = sizeCaps(inst._zod.def.checks);
  if (!caps.length)
    return;
  const cap = Math.min(...caps.map((c) => c[0]));
  const parse = inst._zod.parse;
  inst._zod.parse = (payload, ctx) => {
    const n = size(payload.value);
    if (n > cap && !ctx?.skipChecks && ctx?.direction !== "backward") {
      for (const [max, ch] of sizeCaps(inst._zod.def.checks)) {
        if (n > max && ch._zod.def.when(payload)) {
          const start = payload.issues.length;
          ch._zod.check(payload);
          payload.issues[start].continue = false;
          attachSchema(payload.issues, start, inst);
          return payload;
        }
      }
    }
    return parse(payload, ctx);
  };
}
var arraySize = (input) => Array.isArray(input) ? input.length : -1;
var mapSize = (input) => input instanceof Map ? input.size : -1;
var setSize = (input) => input instanceof Set ? input.size : -1;
function handleArrayResult(result, final, index) {
  if (result.issues.length) {
    final.issues.push(...prefixIssues(index, result.issues));
  }
  final.value[index] = result.value;
}
var $ZodArray = /* @__PURE__ */ $constructor("$ZodArray", (inst, def) => {
  $ZodType.init(inst, def);
  const memo = globalConfig.memoizer;
  memo?.attach(inst);
  inst._zod.parse = (payload, ctx) => {
    const input = payload.value;
    if (!Array.isArray(input)) {
      payload.issues.push({
        expected: "array",
        code: "invalid_type",
        input,
        inst
      });
      return payload;
    }
    payload.value = memo ? memo.alloc(inst, payload, Array(input.length), ctx) : Array(input.length);
    const proms = [];
    const abortEarly = ctx?.abortEarly;
    for (let i = 0;i < input.length; i++) {
      const item = input[i];
      const result = def.element._zod.run({
        value: item,
        issues: []
      }, ctx);
      if (result instanceof Promise) {
        proms.push(result.then((result) => handleArrayResult(result, payload, i)));
      } else {
        handleArrayResult(result, payload, i);
        if (abortEarly && result.issues.length !== 0 && aborted(result))
          break;
      }
    }
    if (proms.length) {
      return Promise.all(proms).then(() => payload);
    }
    return payload;
  };
  capSize(inst, arraySize);
});
function handlePropertyResult(result, final, key, input, optin, optout) {
  const isPresent = key in input;
  const isOptionalOut = optout === "optional";
  if (!isPresent && isOptionalOut && optin === "optional") {
    return;
  }
  if (result.issues.length) {
    if (optin !== undefined && isOptionalOut && !isPresent) {
      return;
    }
    final.issues.push(...prefixIssues(key, result.issues));
  }
  if (!isPresent && optin === undefined) {
    if (!result.issues.length) {
      final.issues.push({
        code: "invalid_type",
        expected: "nonoptional",
        input: undefined,
        path: [key]
      });
    }
    return;
  }
  if (result.value === undefined) {
    if (isPresent || optin === "defaulted" && !isOptionalOut) {
      final.value[key] = undefined;
    }
  } else {
    final.value[key] = result.value;
  }
}
var NO_SYMBOL_KEYS = [];
function normalizeDef(def) {
  const keys = Object.keys(def.shape);
  const ownSymbols = Object.getOwnPropertySymbols(def.shape);
  const symbolKeys = ownSymbols.length ? ownSymbols : NO_SYMBOL_KEYS;
  const allKeys = symbolKeys.length ? [...keys, ...symbolKeys] : keys;
  for (const k of allKeys) {
    if (!def.shape?.[k]?._zod?.traits?.has("$ZodType")) {
      throw new Error(`Invalid element at key "${String(k)}": expected a Zod schema`);
    }
  }
  const okeys = optionalKeys(def.shape);
  return {
    ...def,
    allKeys,
    symbolKeys,
    keySet: new Set(keys),
    numKeys: keys.length,
    optionalKeys: new Set(okeys)
  };
}
function handleCatchall(proms, input, payload, ctx, def, inst, abortEarly) {
  const unrecognized = [];
  const keySet = def.keySet;
  const _catchall = def.catchall._zod;
  const t = _catchall.def.type;
  const optin = _catchall.optin;
  const optout = _catchall.optout;
  let seen = 0;
  for (const key in input) {
    if (abortEarly && payload.issues.length !== seen) {
      if (aborted(payload, seen))
        break;
      seen = payload.issues.length;
    }
    if (keySet.has(key))
      continue;
    if (key === "__proto__") {
      if (t === "never")
        unrecognized.push(key);
      continue;
    }
    if (t === "never") {
      unrecognized.push(key);
      continue;
    }
    const r = _catchall.run({ value: input[key], issues: [] }, ctx);
    if (r instanceof Promise) {
      proms.push(r.then((r) => handlePropertyResult(r, payload, key, input, optin, optout)));
    } else {
      handlePropertyResult(r, payload, key, input, optin, optout);
    }
  }
  if (unrecognized.length) {
    payload.issues.push({
      code: "unrecognized_keys",
      keys: unrecognized,
      input,
      inst,
      continue: true
    });
  }
  if (!proms.length)
    return payload;
  return Promise.all(proms).then(() => {
    return payload;
  });
}
var $ZodObject = /* @__PURE__ */ $constructor("$ZodObject", (inst, def) => {
  $ZodType.init(inst, def);
  const desc = Object.getOwnPropertyDescriptor(def, "shape");
  const sh = desc?.get ? desc.get.raw : def.shape ?? {};
  if (sh) {
    const get = () => {
      const newSh = { ...sh };
      Object.defineProperty(def, "shape", { value: newSh });
      get.raw = newSh;
      return newSh;
    };
    get.raw = sh;
    Object.defineProperty(def, "shape", { get });
  }
  const _normalized = cached(() => normalizeDef(def));
  defineLazyInternal(inst, "propValues", (zod) => {
    const shape = zod.def.shape;
    const propValues = {};
    for (const key in shape) {
      const field = shape[key]._zod;
      if (field.values) {
        if (!Object.prototype.hasOwnProperty.call(propValues, key)) {
          assignProp(propValues, key, new Set);
        }
        for (const v of field.values)
          propValues[key].add(v);
        if (field.optin !== undefined)
          propValues[key].add(undefined);
      }
    }
    return propValues;
  });
  const isObject2 = isObject;
  const catchall = def.catchall;
  let value;
  const memo = globalConfig.memoizer;
  memo?.attach(inst);
  inst._zod.parse = (payload, ctx) => {
    value ?? (value = _normalized.value);
    const input = payload.value;
    if (!isObject2(input)) {
      payload.issues.push({
        expected: "object",
        code: "invalid_type",
        input,
        inst
      });
      return payload;
    }
    payload.value = memo ? memo.alloc(inst, payload, {}, ctx) : {};
    const proms = [];
    const shape = value.shape;
    const abortEarly = ctx?.abortEarly;
    let seen = payload.issues.length;
    for (const key of value.allKeys) {
      if (abortEarly && payload.issues.length !== seen) {
        if (aborted(payload, seen))
          break;
        seen = payload.issues.length;
      }
      if (key === "__proto__")
        continue;
      const el = shape[key];
      const optin = el._zod.optin;
      const optout = el._zod.optout;
      const r = el._zod.run({ value: input[key], issues: [] }, ctx);
      if (r instanceof Promise) {
        proms.push(r.then((r) => handlePropertyResult(r, payload, key, input, optin, optout)));
      } else {
        handlePropertyResult(r, payload, key, input, optin, optout);
      }
    }
    if (!catchall) {
      return proms.length ? Promise.all(proms).then(() => payload) : payload;
    }
    return handleCatchall(proms, input, payload, ctx, _normalized.value, inst, abortEarly === true);
  };
});
var $ZodObjectJIT = /* @__PURE__ */ $constructor("$ZodObjectJIT", (inst, def) => {
  $ZodObject.init(inst, def);
  const superParse = inst._zod.parse;
  const _normalized = cached(() => normalizeDef(def));
  const memo = globalConfig.memoizer;
  const generateFastpass = (shape) => {
    const normalized = _normalized.value;
    const syms = normalized.symbolKeys;
    const doc = new Doc(["payload", "ctx"], { shape, inst, memo, syms });
    const parseStr = (k) => `shape[${k}]._zod.run({ value: input[${k}], issues: [] }, ctx)`;
    const prefixStr = (id, k) => `
          let ${id}_ab = false;
          for (let i = 0; i < ${id}.issues.length; i++) {
            const iss = ${id}.issues[i];
            iss.path = iss.path ? [${k}, ...iss.path] : [${k}];
            payload.issues.push(iss);
            if (iss.continue !== true) ${id}_ab = true;
          }
          if (${id}_ab && ctx && ctx.abortEarly) {
            payload.value = newResult;
            return payload;
          }`;
    doc.write(`const input = payload.value;`);
    const ids = Object.create(null);
    let counter = 0;
    for (const key of normalized.allKeys) {
      ids[key] = `key_${counter++}`;
    }
    doc.write(memo ? `const newResult = memo.alloc(inst, payload, {}, ctx);` : `const newResult = {};`);
    for (const key of normalized.allKeys) {
      if (key === "__proto__")
        continue;
      const id = ids[key];
      const k = typeof key === "symbol" ? `syms[${syms.indexOf(key)}]` : esc(key);
      const isPresent = `${k} in input`;
      const schema = shape[key];
      const optin = schema?._zod?.optin;
      const isOptionalIn = optin !== undefined;
      const isOptionalOut = schema?._zod?.optout === "optional";
      doc.write(`const ${id} = ${parseStr(k)};`);
      if (isOptionalIn && isOptionalOut) {
        const assign = optin === "optional" ? `${id}_present` : `${id}.value !== undefined || ${id}_present`;
        doc.write(`
        const ${id}_present = ${isPresent};
        if (!${id}.issues.length || ${id}_present) {
          if (${id}.issues.length) {${prefixStr(id, k)}
          }

          if (${assign}) {
            newResult[${k}] = ${id}.value;
          }
        }

      `);
      } else if (!isOptionalIn) {
        doc.write(`
        const ${id}_present = ${isPresent};
        if (${id}.issues.length) {${prefixStr(id, k)}
        }
        if (!${id}_present && !${id}.issues.length) {
          payload.issues.push({
            code: "invalid_type",
            expected: "nonoptional",
            input: undefined,
            path: [${k}]
          });
          if (ctx && ctx.abortEarly) {
            payload.value = newResult;
            return payload;
          }
        }

        if (${id}_present) {
          newResult[${k}] = ${id}.value;
        }

      `);
      } else {
        doc.write(`
        if (${id}.issues.length) {${prefixStr(id, k)}
        }
      `);
        if (optin === "defaulted") {
          doc.write(`newResult[${k}] = ${id}.value;`);
        } else {
          doc.write(`
        if (${id}.value !== undefined || ${isPresent}) {
          newResult[${k}] = ${id}.value;
        }
      `);
        }
      }
    }
    doc.write(`payload.value = newResult;`);
    doc.write(`return payload;`);
    return doc.compile();
  };
  let fastpass;
  const isObject2 = isObject;
  const jit = !globalConfig.jitless;
  const allowsEval2 = allowsEval;
  const fastEnabled = jit && allowsEval2.value;
  const catchall = def.catchall;
  let value;
  inst._zod.parse = (payload, ctx) => {
    value ?? (value = _normalized.value);
    const input = payload.value;
    if (!isObject2(input)) {
      payload.issues.push({
        expected: "object",
        code: "invalid_type",
        input,
        inst
      });
      return payload;
    }
    if (jit && fastEnabled && ctx?.async === false && ctx.jitless !== true) {
      if (!fastpass)
        fastpass = generateFastpass(def.shape);
      payload = fastpass(payload, ctx);
      if (!catchall)
        return payload;
      return handleCatchall([], input, payload, ctx, value, inst, ctx?.abortEarly === true);
    }
    return superParse(payload, ctx);
  };
});
function handleUnionResults(results, final, inst, ctx) {
  for (const result of results) {
    if (result.issues.length === 0) {
      final.value = result.value;
      return final;
    }
  }
  const nonaborted = results.filter((r) => !aborted(r));
  if (nonaborted.length === 1) {
    final.value = nonaborted[0].value;
    return nonaborted[0];
  }
  final.issues.push({
    code: "invalid_union",
    input: final.value,
    inst,
    errors: results.map((result) => result.issues.map((iss) => finalizeIssue(iss, ctx, config())))
  });
  return final;
}
var $ZodUnion = /* @__PURE__ */ $constructor("$ZodUnion", (inst, def) => {
  $ZodType.init(inst, def);
  defineLazyInternal(inst, "optin", (zod) => zod.def.options.some((o) => o._zod.optin === "defaulted") ? "defaulted" : zod.def.options.some((o) => o._zod.optin !== undefined) ? "optional" : undefined);
  defineLazyInternal(inst, "optout", (zod) => zod.def.options.some((o) => o._zod.optout === "optional") ? "optional" : undefined);
  defineLazyInternal(inst, "values", (zod) => {
    if (zod.def.options.every((o) => o._zod.values)) {
      return new Set(zod.def.options.flatMap((option) => Array.from(option._zod.values)));
    }
    return;
  });
  defineLazyInternal(inst, "pattern", (zod) => {
    if (zod.def.options.every((o) => o._zod.pattern)) {
      const patterns = zod.def.options.map((o) => o._zod.pattern);
      return new RegExp(`^(${patterns.map((p) => cleanRegex(p.source)).join("|")})$`);
    }
    return;
  });
  const first = def.options.length === 1 ? def.options[0]._zod.run : null;
  inst._zod.parse = (payload, ctx) => {
    if (first) {
      return first(payload, ctx);
    }
    let async = false;
    const results = [];
    for (const option of def.options) {
      const result = option._zod.run({
        value: payload.value,
        issues: []
      }, ctx);
      if (result instanceof Promise) {
        results.push(result);
        async = true;
      } else {
        if (result.issues.length === 0)
          return result;
        results.push(result);
      }
    }
    if (!async)
      return handleUnionResults(results, payload, inst, ctx);
    return Promise.all(results).then((results) => {
      return handleUnionResults(results, payload, inst, ctx);
    });
  };
});
function handleExclusiveUnionResults(results, final, inst, ctx) {
  const matches = [];
  for (let i = 0;i < results.length; i++) {
    if (results[i].issues.length === 0)
      matches.push(i);
  }
  if (matches.length === 1) {
    final.value = results[matches[0]].value;
    return final;
  }
  if (matches.length === 0) {
    final.issues.push({
      code: "invalid_union",
      input: final.value,
      inst,
      errors: results.map((result) => result.issues.map((iss) => finalizeIssue(iss, ctx, config())))
    });
  } else {
    final.issues.push({
      code: "invalid_union",
      input: final.value,
      inst,
      errors: [],
      inclusive: false,
      matches
    });
  }
  return final;
}
var $ZodXor = /* @__PURE__ */ $constructor("$ZodXor", (inst, def) => {
  $ZodUnion.init(inst, def);
  def.inclusive = false;
  const first = def.options.length === 1 ? def.options[0]._zod.run : null;
  inst._zod.parse = (payload, ctx) => {
    if (first) {
      return first(payload, ctx);
    }
    let async = false;
    const results = [];
    for (const option of def.options) {
      const result = option._zod.run({
        value: payload.value,
        issues: []
      }, ctx);
      if (result instanceof Promise) {
        results.push(result);
        async = true;
      } else {
        results.push(result);
      }
    }
    if (!async)
      return handleExclusiveUnionResults(results, payload, inst, ctx);
    return Promise.all(results).then((results) => {
      return handleExclusiveUnionResults(results, payload, inst, ctx);
    });
  };
});
function discriminatorMap(def) {
  const map = new Map;
  for (const option of def.options) {
    const values = option._zod.propValues?.[def.discriminator];
    if (!values || values.size === 0)
      throw new Error(`Invalid discriminated union option at index "${def.options.indexOf(option)}"`);
    for (const value of values) {
      if (map.has(value)) {
        if (value !== undefined)
          throw new Error(`Duplicate discriminator value "${String(value)}"`);
        map.set(value, null);
      } else {
        map.set(value, option);
      }
    }
  }
  return map;
}
var $ZodDiscriminatedUnion = /* @__PURE__ */ $constructor("$ZodDiscriminatedUnion", (inst, def) => {
  def.inclusive = false;
  $ZodUnion.init(inst, def);
  const _super = inst._zod.parse;
  defineLazyInternal(inst, "propValues", (zod) => {
    const propValues = {};
    let undefinedCount = 0;
    for (const option of zod.def.options) {
      const pv = option._zod.propValues;
      if (!pv || Object.keys(pv).length === 0)
        throw new Error(`Invalid discriminated union option at index "${zod.def.options.indexOf(option)}"`);
      if (pv[zod.def.discriminator]?.has(undefined))
        undefinedCount++;
      for (const [k, v] of Object.entries(pv)) {
        if (!Object.prototype.hasOwnProperty.call(propValues, k)) {
          assignProp(propValues, k, new Set);
        }
        for (const val of v) {
          propValues[k].add(val);
        }
      }
    }
    if (!zod.def.unionFallback && undefinedCount > 1)
      propValues[zod.def.discriminator]?.delete(undefined);
    return propValues;
  });
  def.options.forEach((option, i) => {
    const propShape = rawShape(option._zod.def);
    if (propShape && !Object.prototype.hasOwnProperty.call(propShape, def.discriminator)) {
      throw new Error(`Invalid discriminated union option at index "${i}"`);
    }
  });
  const disc = cached(() => discriminatorMap(def));
  inst._zod.parse = (payload, ctx) => {
    const input = payload.value;
    if (!isObject(input)) {
      payload.issues.push({
        code: "invalid_type",
        expected: "object",
        input,
        inst
      });
      return payload;
    }
    const value = input?.[def.discriminator];
    const opt = disc.value.get(value);
    if (opt && (value !== undefined || ctx.direction !== "backward")) {
      return opt._zod.run(payload, ctx);
    }
    if (def.unionFallback || ctx.direction === "backward") {
      return _super(payload, ctx);
    }
    payload.issues.push({
      code: "invalid_union",
      errors: [],
      note: "No matching discriminator",
      discriminator: def.discriminator,
      options: Array.from(disc.value.keys()).filter((value) => disc.value.get(value) !== null),
      input,
      path: [def.discriminator],
      inst
    });
    return payload;
  };
});
var $ZodIntersection = /* @__PURE__ */ $constructor("$ZodIntersection", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.parse = (payload, ctx) => {
    const input = payload.value;
    const left = def.left._zod.run({ value: input, issues: [] }, ctx);
    const right = def.right._zod.run({ value: input, issues: [] }, ctx);
    const async = left instanceof Promise || right instanceof Promise;
    if (async) {
      return Promise.all([left, right]).then(([left, right]) => {
        return handleIntersectionResults(payload, left, right);
      });
    }
    return handleIntersectionResults(payload, left, right);
  };
});
function mergeValues(a, b) {
  if (a === b) {
    return { valid: true, data: a };
  }
  if (a instanceof Date && b instanceof Date && +a === +b) {
    return { valid: true, data: a };
  }
  if (isPlainObject(a) && isPlainObject(b)) {
    const bKeys = Object.keys(b);
    const sharedKeys = Object.keys(a).filter((key) => bKeys.indexOf(key) !== -1);
    const newObj = { ...a, ...b };
    if (Object.prototype.hasOwnProperty.call(newObj, "__proto__"))
      delete newObj.__proto__;
    for (const key of sharedKeys) {
      if (key === "__proto__")
        continue;
      const sharedValue = mergeValues(a[key], b[key]);
      if (!sharedValue.valid) {
        return {
          valid: false,
          mergeErrorPath: [key, ...sharedValue.mergeErrorPath]
        };
      }
      newObj[key] = sharedValue.data;
    }
    return { valid: true, data: newObj };
  }
  if (Array.isArray(a) && Array.isArray(b)) {
    if (a.length !== b.length) {
      return { valid: false, mergeErrorPath: [] };
    }
    const newArray = [];
    for (let index = 0;index < a.length; index++) {
      const itemA = a[index];
      const itemB = b[index];
      const sharedValue = mergeValues(itemA, itemB);
      if (!sharedValue.valid) {
        return {
          valid: false,
          mergeErrorPath: [index, ...sharedValue.mergeErrorPath]
        };
      }
      newArray.push(sharedValue.data);
    }
    return { valid: true, data: newArray };
  }
  return { valid: false, mergeErrorPath: [] };
}
function handleIntersectionResults(result, left, right) {
  const unrecKeys = new Map;
  let unrecIssue;
  const keyIssues = new Map;
  const collect = (iss, side) => {
    let keys;
    if (iss.code === "unrecognized_keys" && !iss.path?.length) {
      unrecIssue ?? (unrecIssue = iss);
      keys = iss.keys;
    } else if (iss.code === "invalid_key" && iss.origin === "record" && iss.path?.length === 1) {
      const k = String(iss.path[0]);
      if (!keyIssues.has(k))
        keyIssues.set(k, iss);
      keys = [k];
    } else {
      return false;
    }
    for (const k of keys) {
      if (!unrecKeys.has(k))
        unrecKeys.set(k, {});
      unrecKeys.get(k)[side] = true;
    }
    return true;
  };
  for (const iss of left.issues) {
    if (!collect(iss, "l"))
      result.issues.push(iss);
  }
  for (const iss of right.issues) {
    if (!collect(iss, "r"))
      result.issues.push(iss);
  }
  const bothKeys = [...unrecKeys].filter(([, f]) => f.l && f.r).map(([k]) => k);
  if (bothKeys.length) {
    const aggregated = unrecIssue ? bothKeys.filter((k) => unrecIssue.keys.includes(k)) : [];
    if (aggregated.length)
      result.issues.push({ ...unrecIssue, keys: aggregated });
    for (const k of bothKeys) {
      if (!aggregated.includes(k) && keyIssues.has(k))
        result.issues.push(keyIssues.get(k));
    }
  }
  const merged = mergeValues(left.value, right.value);
  if (!merged.valid) {
    if (aborted(result))
      return result;
    throw new Error(`Unmergable intersection. Error path: ` + `${JSON.stringify(merged.mergeErrorPath)}`);
  }
  result.value = merged.data;
  return result;
}
var $ZodTuple = /* @__PURE__ */ $constructor("$ZodTuple", (inst, def) => {
  $ZodType.init(inst, def);
  const items = def.items;
  const memo = globalConfig.memoizer;
  memo?.attach(inst);
  inst._zod.parse = (payload, ctx) => {
    const input = payload.value;
    if (!Array.isArray(input)) {
      payload.issues.push({
        input,
        inst,
        expected: "tuple",
        code: "invalid_type"
      });
      return payload;
    }
    payload.value = memo ? memo.alloc(inst, payload, [], ctx) : [];
    const proms = [];
    const optinStart = getTupleOptStart(items, "optin");
    const optoutStart = getTupleOptStart(items, "optout");
    if (!def.rest) {
      if (input.length < optinStart) {
        payload.issues.push({
          code: "too_small",
          minimum: optinStart,
          inclusive: true,
          input,
          inst,
          origin: "array"
        });
        return payload;
      }
      if (input.length > items.length) {
        payload.issues.push({
          code: "too_big",
          maximum: items.length,
          inclusive: true,
          input,
          inst,
          origin: "array"
        });
      }
    }
    const itemResults = new Array(items.length);
    const abortEarly = def.rest ? ctx?.abortEarly : undefined;
    let itemAborted = false;
    for (let i = 0;i < items.length; i++) {
      const r = items[i]._zod.run({ value: input[i], issues: [] }, ctx);
      if (r instanceof Promise) {
        proms.push(r.then((rr) => {
          itemResults[i] = rr;
        }));
      } else {
        itemResults[i] = r;
        if (abortEarly && !itemAborted && r.issues.length)
          itemAborted = aborted(r);
      }
    }
    if (def.rest && !itemAborted) {
      let i = items.length - 1;
      const rest = input.slice(items.length);
      let seen = payload.issues.length;
      for (const el of rest) {
        if (abortEarly && payload.issues.length !== seen) {
          if (aborted(payload, seen))
            break;
          seen = payload.issues.length;
        }
        i++;
        const result = def.rest._zod.run({ value: el, issues: [] }, ctx);
        if (result instanceof Promise) {
          proms.push(result.then((r) => handleTupleResult(r, payload, i)));
        } else {
          handleTupleResult(result, payload, i);
        }
      }
    }
    if (proms.length) {
      return Promise.all(proms).then(() => handleTupleResults(itemResults, payload, items, input, optoutStart));
    }
    return handleTupleResults(itemResults, payload, items, input, optoutStart);
  };
});
function getTupleOptStart(items, key) {
  for (let i = items.length - 1;i >= 0; i--) {
    const omittable = key === "optin" ? items[i]._zod.optin !== undefined : items[i]._zod.optout === "optional";
    if (!omittable)
      return i + 1;
  }
  return 0;
}
function handleTupleResult(result, final, index) {
  if (result.issues.length) {
    final.issues.push(...prefixIssues(index, result.issues));
  }
  final.value[index] = result.value;
}
function handleTupleResults(itemResults, final, items, input, optoutStart) {
  for (let i = 0;i < items.length; i++) {
    const r = itemResults[i];
    const isPresent = i < input.length;
    if (!isPresent && i >= optoutStart && items[i]._zod.optin === "optional") {
      final.value.length = i;
      break;
    }
    if (r.issues.length) {
      if (!isPresent && i >= optoutStart) {
        final.value.length = i;
        break;
      }
      final.issues.push(...prefixIssues(i, r.issues));
    }
    final.value[i] = r.value;
  }
  for (let i = final.value.length - 1;i >= input.length; i--) {
    if (items[i]._zod.optout === "optional" && final.value[i] === undefined) {
      final.value.length = i;
    } else {
      break;
    }
  }
  return final;
}
var $ZodRecord = /* @__PURE__ */ $constructor("$ZodRecord", (inst, def) => {
  $ZodType.init(inst, def);
  const memo = globalConfig.memoizer;
  memo?.attach(inst);
  inst._zod.parse = (payload, ctx) => {
    const input = payload.value;
    if (!isPlainObject(input)) {
      payload.issues.push({
        expected: "record",
        code: "invalid_type",
        input,
        inst
      });
      return payload;
    }
    const proms = [];
    const values = def.keyType._zod.values;
    if (values && !def.partial) {
      payload.value = memo ? memo.alloc(inst, payload, {}, ctx) : {};
      const recordKeys = new Set;
      for (const key of values) {
        if (typeof key === "string" || typeof key === "number" || typeof key === "symbol") {
          recordKeys.add(typeof key === "number" ? key.toString() : key);
          if (key === "__proto__")
            continue;
          const keyResult = def.keyType._zod.run({ value: key, issues: [] }, ctx);
          if (keyResult instanceof Promise) {
            throw new Error("Async schemas not supported in object keys currently");
          }
          if (keyResult.issues.length) {
            payload.issues.push({
              code: "invalid_key",
              origin: "record",
              issues: keyResult.issues.map((iss) => finalizeIssue(iss, ctx, config())),
              input: key,
              path: [key],
              inst
            });
            continue;
          }
          const outKey = keyResult.value;
          if (outKey === "__proto__")
            continue;
          const result = def.valueType._zod.run({ value: input[key], issues: [] }, ctx);
          if (result instanceof Promise) {
            proms.push(result.then((result) => {
              if (result.issues.length) {
                payload.issues.push(...prefixIssues(key, result.issues));
              }
              payload.value[outKey] = result.value;
            }));
          } else {
            if (result.issues.length) {
              payload.issues.push(...prefixIssues(key, result.issues));
            }
            payload.value[outKey] = result.value;
          }
        }
      }
      let unrecognized;
      for (const key in input) {
        if (!recordKeys.has(key)) {
          if (def.mode === "loose") {
            if (key === "__proto__")
              continue;
            payload.value[key] = input[key];
          } else {
            unrecognized = unrecognized ?? [];
            unrecognized.push(key);
          }
        }
      }
      if (unrecognized && unrecognized.length > 0) {
        payload.issues.push({
          code: "unrecognized_keys",
          input,
          inst,
          keys: unrecognized,
          continue: true
        });
      }
    } else {
      payload.value = memo ? memo.alloc(inst, payload, {}, ctx) : {};
      let unrecognized;
      for (const key of Reflect.ownKeys(input)) {
        if (key === "__proto__")
          continue;
        if (!Object.prototype.propertyIsEnumerable.call(input, key))
          continue;
        let keyResult = def.keyType._zod.run({ value: key, issues: [] }, ctx);
        if (keyResult instanceof Promise) {
          throw new Error("Async schemas not supported in object keys currently");
        }
        const checkNumericKey = typeof key === "string" && number.test(key) && keyResult.issues.length;
        if (checkNumericKey) {
          const retryResult = def.keyType._zod.run({ value: Number(key), issues: [] }, ctx);
          if (retryResult instanceof Promise) {
            throw new Error("Async schemas not supported in object keys currently");
          }
          if (retryResult.issues.length === 0) {
            keyResult = retryResult;
          }
        }
        if (keyResult.issues.length) {
          if (def.mode === "loose") {
            payload.value[key] = input[key];
          } else if (values) {
            unrecognized = unrecognized ?? [];
            unrecognized.push(key);
          } else {
            payload.issues.push({
              code: "invalid_key",
              origin: "record",
              issues: keyResult.issues.map((iss) => finalizeIssue(iss, ctx, config())),
              input: key,
              path: [key],
              inst
            });
          }
          continue;
        }
        const outKey = keyResult.value;
        if (outKey === "__proto__")
          continue;
        const result = def.valueType._zod.run({ value: input[key], issues: [] }, ctx);
        if (result instanceof Promise) {
          proms.push(result.then((result) => {
            if (result.issues.length) {
              payload.issues.push(...prefixIssues(key, result.issues));
            }
            payload.value[outKey] = result.value;
          }));
        } else {
          if (result.issues.length) {
            payload.issues.push(...prefixIssues(key, result.issues));
          }
          payload.value[outKey] = result.value;
        }
      }
      if (unrecognized && unrecognized.length > 0) {
        payload.issues.push({
          code: "unrecognized_keys",
          input,
          inst,
          keys: unrecognized,
          continue: true
        });
      }
    }
    if (proms.length) {
      return Promise.all(proms).then(() => payload);
    }
    return payload;
  };
});
var $ZodMap = /* @__PURE__ */ $constructor("$ZodMap", (inst, def) => {
  $ZodType.init(inst, def);
  const memo = globalConfig.memoizer;
  memo?.attach(inst);
  inst._zod.parse = (payload, ctx) => {
    const input = payload.value;
    if (!(input instanceof Map)) {
      payload.issues.push({
        expected: "map",
        code: "invalid_type",
        input,
        inst
      });
      return payload;
    }
    const proms = [];
    payload.value = memo ? memo.alloc(inst, payload, new Map, ctx) : new Map;
    const abortEarly = ctx?.abortEarly;
    let seen = payload.issues.length;
    for (const [key, value] of input) {
      if (abortEarly && payload.issues.length !== seen) {
        if (aborted(payload, seen))
          break;
        seen = payload.issues.length;
      }
      const keyResult = def.keyType._zod.run({ value: key, issues: [] }, ctx);
      const valueResult = def.valueType._zod.run({ value, issues: [] }, ctx);
      if (keyResult instanceof Promise || valueResult instanceof Promise) {
        proms.push(Promise.all([keyResult, valueResult]).then(([keyResult, valueResult]) => {
          handleMapResult(keyResult, valueResult, payload, key, input, inst, ctx);
        }));
      } else {
        handleMapResult(keyResult, valueResult, payload, key, input, inst, ctx);
      }
    }
    if (proms.length)
      return Promise.all(proms).then(() => payload);
    return payload;
  };
  capSize(inst, mapSize);
});
function handleMapResult(keyResult, valueResult, final, key, input, inst, ctx) {
  if (keyResult.issues.length) {
    if (propertyKeyTypes.has(typeof key)) {
      final.issues.push(...prefixIssues(key, keyResult.issues));
    } else {
      final.issues.push({
        code: "invalid_key",
        origin: "map",
        input,
        inst,
        issues: keyResult.issues.map((iss) => finalizeIssue(iss, ctx, config()))
      });
    }
  }
  if (valueResult.issues.length) {
    if (propertyKeyTypes.has(typeof key)) {
      final.issues.push(...prefixIssues(key, valueResult.issues));
    } else {
      final.issues.push({
        origin: "map",
        code: "invalid_element",
        input,
        inst,
        key,
        issues: valueResult.issues.map((iss) => finalizeIssue(iss, ctx, config()))
      });
    }
  }
  final.value.set(keyResult.value, valueResult.value);
}
var $ZodSet = /* @__PURE__ */ $constructor("$ZodSet", (inst, def) => {
  $ZodType.init(inst, def);
  const memo = globalConfig.memoizer;
  memo?.attach(inst);
  inst._zod.parse = (payload, ctx) => {
    const input = payload.value;
    if (!(input instanceof Set)) {
      payload.issues.push({
        input,
        inst,
        expected: "set",
        code: "invalid_type"
      });
      return payload;
    }
    const proms = [];
    payload.value = memo ? memo.alloc(inst, payload, new Set, ctx) : new Set;
    const abortEarly = ctx?.abortEarly;
    let seen = payload.issues.length;
    for (const item of input) {
      if (abortEarly && payload.issues.length !== seen) {
        if (aborted(payload, seen))
          break;
        seen = payload.issues.length;
      }
      const result = def.valueType._zod.run({ value: item, issues: [] }, ctx);
      if (result instanceof Promise) {
        proms.push(result.then((result) => handleSetResult(result, payload)));
      } else
        handleSetResult(result, payload);
    }
    if (proms.length)
      return Promise.all(proms).then(() => payload);
    return payload;
  };
  capSize(inst, setSize);
});
function handleSetResult(result, final) {
  if (result.issues.length) {
    final.issues.push(...result.issues);
  }
  final.value.add(result.value);
}
var $ZodEnum = /* @__PURE__ */ $constructor("$ZodEnum", (inst, def) => {
  $ZodType.init(inst, def);
  const values = getEnumValues(def.entries);
  const valuesSet = new Set(values);
  inst._zod.values = valuesSet;
  defineLazyInternal(inst, "pattern", (zod) => {
    const patternValues = getEnumValues(zod.def.entries).filter((k) => propertyKeyTypes.has(typeof k));
    return new RegExp(patternValues.length ? `^(${patternValues.map((o) => escapeRegex(o.toString())).join("|")})$` : "^[^\\s\\S]$");
  });
  inst._zod.parse = (payload, _ctx) => {
    const input = payload.value;
    if (valuesSet.has(input)) {
      return payload;
    }
    payload.issues.push({
      code: "invalid_value",
      values,
      input,
      inst
    });
    return payload;
  };
});
var $ZodLiteral = /* @__PURE__ */ $constructor("$ZodLiteral", (inst, def) => {
  $ZodType.init(inst, def);
  const values = new Set(def.values);
  inst._zod.values = values;
  defineLazyInternal(inst, "pattern", (zod) => {
    const vals = zod.def.values;
    return new RegExp(vals.length ? `^(${vals.map((o) => typeof o === "string" ? escapeRegex(o) : o ? escapeRegex(o.toString()) : String(o)).join("|")})$` : "^[^\\s\\S]$");
  });
  inst._zod.parse = (payload, _ctx) => {
    const input = payload.value;
    if (values.has(input)) {
      return payload;
    }
    payload.issues.push({
      code: "invalid_value",
      values: def.values,
      input,
      inst
    });
    return payload;
  };
});
var $ZodFile = /* @__PURE__ */ $constructor("$ZodFile", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.parse = (payload, _ctx) => {
    const input = payload.value;
    if (input instanceof File)
      return payload;
    payload.issues.push({
      expected: "file",
      code: "invalid_type",
      input,
      inst
    });
    return payload;
  };
});
var $ZodTransform = /* @__PURE__ */ $constructor("$ZodTransform", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.optin = "optional";
  globalConfig.memoizer?.guard(inst);
  inst._zod.parse = (payload, ctx) => {
    if (ctx.direction === "backward") {
      throw new $ZodEncodeError(inst.constructor.name);
    }
    const _out = def.transform(payload.value, payload);
    if (ctx.async) {
      const output = _out instanceof Promise ? _out : Promise.resolve(_out);
      return output.then((output) => {
        payload.value = output;
        return payload;
      });
    }
    if (_out instanceof Promise) {
      throw new $ZodAsyncError;
    }
    payload.value = _out;
    return payload;
  };
});
function handleOptionalResult(payload, result) {
  payload.value = result.issues.length ? undefined : result.value;
  return payload;
}
var $ZodOptional = /* @__PURE__ */ $constructor("$ZodOptional", (inst, def) => {
  $ZodType.init(inst, def);
  defineLazyInternal(inst, "optin", (zod) => zod.def.innerType._zod.optin === "defaulted" ? "defaulted" : "optional");
  inst._zod.optout = "optional";
  defineLazyInternal(inst, "values", (zod) => {
    const values = zod.def.innerType._zod.values;
    return values ? new Set([...values, undefined]) : undefined;
  });
  defineLazyInternal(inst, "pattern", (zod) => {
    const pattern = zod.def.innerType._zod.pattern;
    return pattern ? new RegExp(`^(${cleanRegex(pattern.source)})?$`) : undefined;
  });
  inst._zod.parse = (payload, ctx) => {
    if (payload.value === undefined) {
      if (def.innerType._zod.optin !== "defaulted")
        return payload;
      const result = def.innerType._zod.run({ value: payload.value, issues: [] }, ctx);
      if (result instanceof Promise)
        return result.then((result) => handleOptionalResult(payload, result));
      return handleOptionalResult(payload, result);
    }
    return def.innerType._zod.run(payload, ctx);
  };
});
var $ZodExactOptional = /* @__PURE__ */ $constructor("$ZodExactOptional", (inst, def) => {
  $ZodOptional.init(inst, def);
  defineLazyInternal(inst, "values", (zod) => zod.def.innerType._zod.values);
  defineLazyInternal(inst, "pattern", (zod) => zod.def.innerType._zod.pattern);
  inst._zod.parse = (payload, ctx) => {
    return def.innerType._zod.run(payload, ctx);
  };
});
var $ZodNullable = /* @__PURE__ */ $constructor("$ZodNullable", (inst, def) => {
  $ZodType.init(inst, def);
  defineLazyInternal(inst, "optin", (zod) => zod.def.innerType._zod.optin);
  defineLazyInternal(inst, "optout", (zod) => zod.def.innerType._zod.optout);
  defineLazyInternal(inst, "pattern", (zod) => {
    const pattern = zod.def.innerType._zod.pattern;
    return pattern ? new RegExp(`^(${cleanRegex(pattern.source)}|null)$`) : undefined;
  });
  defineLazyInternal(inst, "values", (zod) => {
    return zod.def.innerType._zod.values ? new Set([...zod.def.innerType._zod.values, null]) : undefined;
  });
  inst._zod.parse = (payload, ctx) => {
    if (payload.value === null)
      return payload;
    return def.innerType._zod.run(payload, ctx);
  };
});
var $ZodDefault = /* @__PURE__ */ $constructor("$ZodDefault", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.optin = "defaulted";
  defineLazyInternal(inst, "values", (zod) => zod.def.innerType._zod.values);
  inst._zod.parse = (payload, ctx) => {
    if (ctx.direction === "backward") {
      return def.innerType._zod.run(payload, ctx);
    }
    if (payload.value === undefined) {
      payload.value = def.defaultValue;
      return payload;
    }
    const result = def.innerType._zod.run(payload, ctx);
    if (result instanceof Promise) {
      return result.then((result) => handleDefaultResult(result, def));
    }
    return handleDefaultResult(result, def);
  };
});
function handleDefaultResult(payload, def) {
  if (payload.value === undefined) {
    payload.value = def.defaultValue;
  }
  return payload;
}
var $ZodPrefault = /* @__PURE__ */ $constructor("$ZodPrefault", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.optin = "defaulted";
  defineLazyInternal(inst, "values", (zod) => zod.def.innerType._zod.values);
  inst._zod.parse = (payload, ctx) => {
    if (ctx.direction === "backward") {
      return def.innerType._zod.run(payload, ctx);
    }
    if (payload.value === undefined) {
      payload.value = def.defaultValue;
    }
    return def.innerType._zod.run(payload, ctx);
  };
});
var $ZodNonOptional = /* @__PURE__ */ $constructor("$ZodNonOptional", (inst, def) => {
  $ZodType.init(inst, def);
  defineLazyInternal(inst, "values", (zod) => {
    const v = zod.def.innerType._zod.values;
    return v ? new Set([...v].filter((x) => x !== undefined)) : undefined;
  });
  inst._zod.parse = (payload, ctx) => {
    const result = def.innerType._zod.run(payload, ctx);
    if (result instanceof Promise) {
      return result.then((result) => handleNonOptionalResult(result, inst));
    }
    return handleNonOptionalResult(result, inst);
  };
});
function handleNonOptionalResult(payload, inst) {
  if (!payload.issues.length && payload.value === undefined) {
    payload.issues.push({
      code: "invalid_type",
      expected: "nonoptional",
      input: payload.value,
      inst
    });
  }
  return payload;
}
var $ZodSuccess = /* @__PURE__ */ $constructor("$ZodSuccess", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.parse = (payload, ctx) => {
    if (ctx.direction === "backward") {
      throw new $ZodEncodeError("ZodSuccess");
    }
    const result = def.innerType._zod.run(payload, ctx);
    if (result instanceof Promise) {
      return result.then((result) => {
        payload.value = result.issues.length === 0;
        return payload;
      });
    }
    payload.value = result.issues.length === 0;
    return payload;
  };
});
function handleCatchResult(payload, result, def, ctx) {
  if (!result.issues.length) {
    payload.value = result.value;
    if (result.memo)
      payload.memo = true;
    return payload;
  }
  payload.value = def.catchValue({
    ...result,
    value: payload.value,
    error: {
      issues: result.issues.map((iss) => finalizeIssue(iss, ctx, config()))
    },
    input: payload.value
  });
  return payload;
}
var $ZodCatch = /* @__PURE__ */ $constructor("$ZodCatch", (inst, def) => {
  $ZodType.init(inst, def);
  defineLazyInternal(inst, "optin", (zod) => zod.def.innerType._zod.optin === "defaulted" ? "defaulted" : "optional");
  defineLazyInternal(inst, "optout", (zod) => zod.def.innerType._zod.optout);
  defineLazyInternal(inst, "values", (zod) => zod.def.innerType._zod.values);
  inst._zod.parse = (payload, ctx) => {
    if (ctx.direction === "backward") {
      return def.innerType._zod.run(payload, ctx);
    }
    const result = def.innerType._zod.run({ value: payload.value, issues: [] }, ctx);
    if (result instanceof Promise) {
      return result.then((result) => handleCatchResult(payload, result, def, ctx));
    }
    return handleCatchResult(payload, result, def, ctx);
  };
});
var $ZodNaN = /* @__PURE__ */ $constructor("$ZodNaN", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.parse = (payload, _ctx) => {
    if (typeof payload.value !== "number" || !Number.isNaN(payload.value)) {
      payload.issues.push({
        input: payload.value,
        inst,
        expected: "nan",
        code: "invalid_type"
      });
      return payload;
    }
    return payload;
  };
});
var $ZodPipe = /* @__PURE__ */ $constructor("$ZodPipe", (inst, def) => {
  $ZodType.init(inst, def);
  defineLazyInternal(inst, "values", (zod) => zod.def.in._zod.values);
  defineLazyInternal(inst, "optin", (zod) => zod.def.in._zod.optin);
  defineLazyInternal(inst, "optout", (zod) => zod.def.out._zod.optout);
  defineLazyInternal(inst, "propValues", (zod) => zod.def.in._zod.propValues);
  inst._zod.parse = (payload, ctx) => {
    if (ctx.direction === "backward") {
      const right = def.out._zod.run(payload, ctx);
      if (right instanceof Promise) {
        return right.then((right) => handlePipeResult(right, def.in, ctx));
      }
      return handlePipeResult(right, def.in, ctx);
    }
    const left = def.in._zod.run(payload, ctx);
    if (left instanceof Promise) {
      return left.then((left) => handlePipeResult(left, def.out, ctx));
    }
    return handlePipeResult(left, def.out, ctx);
  };
});
function handlePipeResult(left, next, ctx) {
  if (left.issues.some((iss) => iss.code !== "unrecognized_keys")) {
    left.aborted = true;
    return left;
  }
  return next._zod.run({ value: left.value, issues: left.issues }, ctx);
}
var $ZodCodec = /* @__PURE__ */ $constructor("$ZodCodec", (inst, def) => {
  $ZodType.init(inst, def);
  defineLazyInternal(inst, "values", (zod) => zod.def.in._zod.values);
  defineLazyInternal(inst, "optin", (zod) => zod.def.in._zod.optin);
  defineLazyInternal(inst, "optout", (zod) => zod.def.out._zod.optout);
  defineLazyInternal(inst, "propValues", (zod) => zod.def.in._zod.propValues);
  inst._zod.parse = (payload, ctx) => {
    const direction = ctx.direction || "forward";
    if (direction === "forward") {
      const left = def.in._zod.run(payload, ctx);
      if (left instanceof Promise) {
        return left.then((left) => handleCodecAResult(left, def, ctx));
      }
      return handleCodecAResult(left, def, ctx);
    } else {
      const right = def.out._zod.run(payload, ctx);
      if (right instanceof Promise) {
        return right.then((right) => handleCodecAResult(right, def, ctx));
      }
      return handleCodecAResult(right, def, ctx);
    }
  };
});
function handleCodecAResult(result, def, ctx) {
  if (result.issues.length) {
    result.aborted = true;
    return result;
  }
  const direction = ctx.direction || "forward";
  if (direction === "forward") {
    const transformed = def.transform(result.value, result);
    if (transformed instanceof Promise) {
      return transformed.then((value) => handleCodecTxResult(result, value, def.out, ctx));
    }
    return handleCodecTxResult(result, transformed, def.out, ctx);
  } else {
    const transformed = def.reverseTransform(result.value, result);
    if (transformed instanceof Promise) {
      return transformed.then((value) => handleCodecTxResult(result, value, def.in, ctx));
    }
    return handleCodecTxResult(result, transformed, def.in, ctx);
  }
}
function handleCodecTxResult(left, value, nextSchema, ctx) {
  if (left.issues.length) {
    left.aborted = true;
    return left;
  }
  return nextSchema._zod.run({ value, issues: left.issues }, ctx);
}
var $ZodPreprocess = /* @__PURE__ */ $constructor("$ZodPreprocess", (inst, def) => {
  $ZodPipe.init(inst, def);
});
var $ZodReadonly = /* @__PURE__ */ $constructor("$ZodReadonly", (inst, def) => {
  $ZodType.init(inst, def);
  defineLazyInternal(inst, "propValues", (zod) => zod.def.innerType._zod.propValues);
  defineLazyInternal(inst, "values", (zod) => zod.def.innerType._zod.values);
  defineLazyInternal(inst, "optin", (zod) => zod.def.innerType?._zod?.optin);
  defineLazyInternal(inst, "optout", (zod) => zod.def.innerType?._zod?.optout);
  inst._zod.parse = (payload, ctx) => {
    if (ctx.direction === "backward") {
      return def.innerType._zod.run(payload, ctx);
    }
    const result = def.innerType._zod.run(payload, ctx);
    if (result instanceof Promise) {
      return result.then(handleReadonlyResult);
    }
    return handleReadonlyResult(result);
  };
});
function handleReadonlyResult(payload) {
  if (!payload.memo)
    payload.value = Object.freeze(payload.value);
  return payload;
}
function leafPattern(schema) {
  const def = schema._zod.def;
  let pattern = def.pattern;
  let isInt = !!def.format?.includes("int");
  let minimum;
  let maximum;
  for (const ch of def.checks ?? []) {
    const d = ch._zod.def;
    if (d.pattern)
      pattern = d.pattern;
    isInt || (isInt = !!d.format?.includes("int"));
    const lo = d.minimum ?? d.length;
    const hi = d.maximum ?? d.length;
    if (lo !== undefined && (minimum === undefined || lo > minimum))
      minimum = lo;
    if (hi !== undefined && (maximum === undefined || hi < maximum))
      maximum = hi;
  }
  if (pattern)
    return pattern.source;
  if (minimum !== undefined && maximum !== undefined && minimum > maximum)
    return "(?!)";
  if (minimum !== undefined || maximum !== undefined)
    return string({ minimum, maximum }).source;
  const own = schema._zod.pattern;
  return (isInt && own === number ? integer : own)?.source;
}
function partPattern(schema) {
  const def = schema._zod.def;
  const own = schema._zod.pattern?.source;
  const inner = def.innerType ?? schema._zod.innerType;
  if (inner) {
    const before = inner._zod.pattern?.source;
    const after = partPattern(inner);
    if (own && before && after && after !== before) {
      return own.replace(cleanRegex(before), () => cleanRegex(after));
    }
    return own;
  }
  if (def.options) {
    const sources = def.options.map(partPattern);
    if (sources.every(Boolean))
      return `^(${sources.map((s) => cleanRegex(s)).join("|")})$`;
  }
  return leafPattern(schema);
}
var $ZodTemplateLiteral = /* @__PURE__ */ $constructor("$ZodTemplateLiteral", (inst, def) => {
  $ZodType.init(inst, def);
  const regexParts = [];
  for (const part of def.parts) {
    if (typeof part === "object" && part !== null) {
      const source = partPattern(part);
      if (!source) {
        throw new Error(`Invalid template literal part, no pattern found: ${[...part._zod.traits].shift()}`);
      }
      regexParts.push(cleanRegex(source));
    } else if (part === null || primitiveTypes.has(typeof part)) {
      regexParts.push(escapeRegex(`${part}`));
    } else {
      throw new Error(`Invalid template literal part: ${part}`);
    }
  }
  inst._zod.pattern = new RegExp(`^${regexParts.join("")}$`);
  inst._zod.parse = (payload, _ctx) => {
    if (typeof payload.value !== "string") {
      payload.issues.push({
        input: payload.value,
        inst,
        expected: "string",
        code: "invalid_type"
      });
      return payload;
    }
    inst._zod.pattern.lastIndex = 0;
    if (!inst._zod.pattern.test(payload.value)) {
      payload.issues.push({
        input: payload.value,
        inst,
        code: "invalid_format",
        format: def.format ?? "template_literal",
        pattern: inst._zod.pattern.source
      });
      return payload;
    }
    return payload;
  };
});
var $ZodFunction = /* @__PURE__ */ $constructor("$ZodFunction", (inst, def) => {
  $ZodType.init(inst, def);
  Object.defineProperty(inst, "_def", { value: def });
  inst._zod.def = def;
  inst.implement = (func) => {
    if (typeof func !== "function") {
      throw new Error("implement() must be called with a function");
    }
    return Object.defineProperty(function(...args) {
      const parsedArgs = inst._def.input ? parse(inst._def.input, args) : args;
      const result = Reflect.apply(func, this, parsedArgs);
      if (inst._def.output) {
        return parse(inst._def.output, result);
      }
      return result;
    }, "_zod", { value: inst._zod, enumerable: false });
  };
  inst.implementAsync = (func) => {
    if (typeof func !== "function") {
      throw new Error("implementAsync() must be called with a function");
    }
    return Object.defineProperty(async function(...args) {
      const parsedArgs = inst._def.input ? await parseAsync(inst._def.input, args) : args;
      const result = await Reflect.apply(func, this, parsedArgs);
      if (inst._def.output) {
        return await parseAsync(inst._def.output, result);
      }
      return result;
    }, "_zod", { value: inst._zod, enumerable: false });
  };
  inst._zod.parse = (payload, _ctx) => {
    if (typeof payload.value !== "function") {
      payload.issues.push({
        code: "invalid_type",
        expected: "function",
        input: payload.value,
        inst
      });
      return payload;
    }
    const hasPromiseOutput = inst._def.output && inst._def.output._zod.def.type === "promise";
    if (hasPromiseOutput) {
      payload.value = inst.implementAsync(payload.value);
    } else {
      payload.value = inst.implement(payload.value);
    }
    return payload;
  };
  inst.input = (...args) => {
    const F = inst.constructor;
    if (Array.isArray(args[0])) {
      return new F({
        type: "function",
        input: new $ZodTuple({
          type: "tuple",
          items: args[0],
          rest: args[1]
        }),
        output: inst._def.output
      });
    }
    return new F({
      type: "function",
      input: args[0],
      output: inst._def.output
    });
  };
  inst.output = (output) => {
    const F = inst.constructor;
    return new F({
      type: "function",
      input: inst._def.input,
      output
    });
  };
  return inst;
});
var $ZodPromise = /* @__PURE__ */ $constructor("$ZodPromise", (inst, def) => {
  $ZodType.init(inst, def);
  inst._zod.parse = (payload, ctx) => {
    return Promise.resolve(payload.value).then((inner) => def.innerType._zod.run({ value: inner, issues: [] }, ctx));
  };
});
var $ZodLazy = /* @__PURE__ */ $constructor("$ZodLazy", (inst, def) => {
  $ZodType.init(inst, def);
  defineLazy(inst._zod, "innerType", () => {
    const d = def;
    if (!d._cachedInner)
      d._cachedInner = def.getter();
    return d._cachedInner;
  });
  defineLazyInternal(inst, "pattern", (zod) => zod.innerType?._zod?.pattern);
  defineLazyInternal(inst, "propValues", (zod) => zod.innerType?._zod?.propValues);
  defineLazyInternal(inst, "optin", (zod) => zod.innerType?._zod?.optin ?? undefined);
  defineLazyInternal(inst, "optout", (zod) => zod.innerType?._zod?.optout ?? undefined);
  inst._zod.parse = (payload, ctx) => {
    const inner = inst._zod.innerType;
    return inner._zod.run(payload, ctx);
  };
});
var $ZodCustom = /* @__PURE__ */ $constructor("$ZodCustom", (inst, def) => {
  $ZodCheck.init(inst, def);
  $ZodType.init(inst, def);
  inst._zod.parse = (payload, _) => {
    return payload;
  };
  inst._zod.check = (payload) => {
    const input = payload.value;
    const r = def.fn(input);
    if (r instanceof Promise) {
      return r.then((r) => handleRefineResult(r, payload, input, inst));
    }
    handleRefineResult(r, payload, input, inst);
    return;
  };
});
function handleRefineResult(result, payload, input, inst) {
  if (!result) {
    const _iss = {
      code: "custom",
      input,
      inst,
      path: [...inst._zod.def.path ?? []],
      continue: !inst._zod.def.abort
    };
    if (inst._zod.def.params)
      _iss.params = inst._zod.def.params;
    payload.issues.push(issue(_iss));
  }
}
// node_modules/zod/v4/core/memoizer.js
class $ZodCyclicError extends Error {
  constructor() {
    super(`Cannot parse a reference cycle that closes through a transform`);
    this.name = "ZodCyclicError";
  }
}
var STATE = "~memo";
var NO_ISSUES = [];
function isRef(value) {
  return value !== null && typeof value === "object";
}
function cloneIssues(issues) {
  return issues.map((iss) => iss.path ? { ...iss, path: iss.path.slice() } : { ...iss });
}
var recursive = /* @__PURE__ */ new WeakMap;
var NONE = 0;
var ASSUMED = 1;
var PROVEN = 2;
function isRecursive(inst, stack, resolve) {
  const cached = recursive.get(inst);
  if (cached !== undefined)
    return cached ? PROVEN : NONE;
  if (stack.has(inst))
    return PROVEN;
  stack.add(inst);
  let result = NONE;
  const check = (child) => {
    if (result !== PROVEN && child?._zod) {
      const answer = isRecursive(child, stack, resolve);
      if (answer > result)
        result = answer;
    }
  };
  const shape = (sh, spread) => {
    let answer = NONE;
    for (const key of Reflect.ownKeys(sh)) {
      const desc = Object.getOwnPropertyDescriptor(sh, key);
      if (spread && !desc.enumerable)
        continue;
      const child = desc.get ? ASSUMED : desc.value?._zod ? isRecursive(desc.value, stack, resolve) : NONE;
      if (child > answer)
        answer = child;
    }
    return answer;
  };
  const merge = (answer) => {
    if (answer > result)
      result = answer;
  };
  const def = inst._zod.def;
  const kind = def.type;
  switch (kind) {
    case "object": {
      const raw = rawShape(def);
      merge(raw ? shape(raw, true) : ASSUMED);
      check(def.catchall);
      break;
    }
    case "array":
      check(def.element);
      break;
    case "tuple":
      for (const el of def.items)
        check(el);
      check(def.rest);
      break;
    case "record":
    case "map":
      check(def.keyType);
      check(def.valueType);
      break;
    case "set":
      check(def.valueType);
      break;
    case "union":
      for (const el of def.options)
        check(el);
      break;
    case "intersection":
      check(def.left);
      check(def.right);
      break;
    case "optional":
    case "nullable":
    case "default":
    case "prefault":
    case "catch":
    case "readonly":
    case "nonoptional":
    case "promise":
    case "success":
      check(def.innerType);
      break;
    case "pipe":
      check(def.in);
      check(def.out);
      break;
    case "function":
      check(def.input);
      check(def.output);
      break;
    case "lazy": {
      const inner = def._cachedInner ?? (resolve ? inst._zod.innerType : undefined);
      merge(inner ? isRecursive(inner, stack, false) : ASSUMED);
      break;
    }
    case "template_literal":
    case "string":
    case "number":
    case "int":
    case "boolean":
    case "bigint":
    case "symbol":
    case "undefined":
    case "null":
    case "void":
    case "never":
    case "any":
    case "unknown":
    case "date":
    case "nan":
    case "enum":
    case "literal":
    case "file":
    case "transform":
    case "custom":
      break;
    default: {
      for (const key in def) {
        const desc = Object.getOwnPropertyDescriptor(def, key);
        if (!desc || desc.get)
          continue;
        const value = desc.value;
        if (!value || typeof value !== "object")
          continue;
        if (value._zod)
          check(value);
        else if (Array.isArray(value))
          for (const el of value)
            check(el);
      }
    }
  }
  stack.delete(inst);
  return settle(inst, result);
}
function settle(inst, answer) {
  if (answer !== ASSUMED)
    recursive.set(inst, answer === PROVEN);
  return answer;
}
function bucketFor(state, inst) {
  let bucket = state.buckets.get(inst);
  if (!bucket) {
    bucket = new WeakMap;
    state.buckets.set(inst, bucket);
  }
  return bucket;
}
var handoff;
var open = [];
var memo = {
  alloc(_inst, payload, empty) {
    const bucket = handoff;
    if (!bucket)
      return empty;
    handoff = undefined;
    const entry = { value: empty, issues: null };
    bucket.set(payload.value, entry);
    open.push(entry);
    return empty;
  },
  guard(inst) {
    var _a;
    (_a = inst._zod).deferred ?? (_a.deferred = []);
    inst._zod.deferred.push(() => {
      const base = inst._zod.parse;
      const wrapped = (payload, ctx) => {
        if (ctx.direction !== "backward" && isBackEdge(ctx, payload.value))
          throw new $ZodCyclicError;
        return base(payload, ctx);
      };
      inst._zod.parse = wrapped;
      if (inst._zod.run === base)
        inst._zod.run = wrapped;
    });
  },
  attach(inst) {
    var _a;
    let isRecursiveInst;
    let rechecked = false;
    let lastCtx;
    let lastBucket;
    (_a = inst._zod).deferred ?? (_a.deferred = []);
    inst._zod.deferred.push(() => {
      const base = inst._zod.parse;
      const wrapped = (payload, ctx) => {
        if (isRecursiveInst === undefined) {
          const walked = isRecursive(inst, new Set, false);
          if (walked === NONE) {
            inst._zod.parse = base;
            if (inst._zod.run === wrapped)
              inst._zod.run = base;
            return base(payload, ctx);
          }
          if (walked === PROVEN || rechecked)
            isRecursiveInst = true;
          else
            rechecked = true;
        }
        const input = payload.value;
        if (!isRef(input))
          return base(payload, ctx);
        let state = ctx[STATE];
        if (!state) {
          state = { buckets: new WeakMap, backEdges: undefined };
          ctx[STATE] = state;
        }
        let bucket;
        if (lastCtx === ctx) {
          bucket = lastBucket;
        } else {
          bucket = bucketFor(state, inst);
          lastCtx = ctx;
          lastBucket = bucket;
        }
        const hit = bucket.get(input);
        if (hit) {
          payload.value = hit.value;
          if (hit.issues) {
            if (hit.issues.length)
              payload.issues.push(...cloneIssues(hit.issues));
          } else {
            payload.memo = true;
            state.backEdges ?? (state.backEdges = new WeakSet);
            state.backEdges.add(hit.value);
          }
          return payload;
        }
        handoff = bucket;
        const depth = open.length;
        const result = base(payload, ctx);
        handoff = undefined;
        const entry = open.length > depth ? open.pop() : undefined;
        if (result instanceof Promise) {
          return result.then((r) => {
            if (entry)
              entry.issues = r.issues.length ? cloneIssues(r.issues) : NO_ISSUES;
            return r;
          });
        }
        if (entry)
          entry.issues = result.issues.length ? cloneIssues(result.issues) : NO_ISSUES;
        return result;
      };
      inst._zod.parse = wrapped;
      if (inst._zod.run === base)
        inst._zod.run = wrapped;
    });
  }
};
function memoizer() {
  return memo;
}
function isBackEdge(ctx, value) {
  const backEdges = ctx[STATE]?.backEdges;
  return backEdges !== undefined && isRef(value) && backEdges.has(value);
}
// node_modules/zod/v4/locales/en.js
var error = () => {
  const Sizable = {
    string: { unit: "characters", verb: "to have" },
    file: { unit: "bytes", verb: "to have" },
    array: { unit: "items", verb: "to have" },
    set: { unit: "items", verb: "to have" },
    map: { unit: "entries", verb: "to have" }
  };
  function getSizing(origin) {
    return Sizable[origin] ?? null;
  }
  const FormatDictionary = {
    regex: "input",
    email: "email address",
    url: "URL",
    emoji: "emoji",
    uuid: "UUID",
    uuidv4: "UUIDv4",
    uuidv6: "UUIDv6",
    nanoid: "nanoid",
    guid: "GUID",
    cuid: "cuid",
    cuid2: "cuid2",
    ulid: "ULID",
    xid: "XID",
    ksuid: "KSUID",
    datetime: "ISO datetime",
    date: "ISO date",
    time: "ISO time",
    duration: "ISO duration",
    ipv4: "IPv4 address",
    ipv6: "IPv6 address",
    mac: "MAC address",
    cidrv4: "IPv4 range",
    cidrv6: "IPv6 range",
    base64: "base64-encoded string",
    base64url: "base64url-encoded string",
    json_string: "JSON string",
    e164: "E.164 number",
    currency_code: "currency code",
    credit_card: "credit card number",
    iban: "IBAN",
    jwt: "JWT",
    template_literal: "input"
  };
  const TypeDictionary = {
    nan: "NaN"
  };
  function getTypeName(type, input) {
    if (type === "number" && typeof input === "number" && !Number.isFinite(input)) {
      return String(input);
    }
    return TypeDictionary[type] ?? type;
  }
  return (issue) => {
    switch (issue.code) {
      case "invalid_type": {
        const expected = getTypeName(issue.expected);
        const receivedType = parsedType(issue.input);
        const received = getTypeName(receivedType, issue.input);
        return `Invalid input: expected ${expected}, received ${received}`;
      }
      case "invalid_value":
        if (issue.values.length === 1)
          return `Invalid input: expected ${stringifyPrimitive(issue.values[0])}`;
        return `Invalid option: expected one of ${joinValues(issue.values, "|")}`;
      case "too_big": {
        const adj = issue.exact ? "exactly " : issue.inclusive ? "<=" : "<";
        const sizing = getSizing(issue.origin);
        if (sizing)
          return `Too big: expected ${issue.origin ?? "value"} to have ${adj}${issue.maximum.toString()} ${sizing.unit ?? "elements"}`;
        return `Too big: expected ${issue.origin ?? "value"} to be ${adj}${issue.maximum.toString()}`;
      }
      case "too_small": {
        const adj = issue.exact ? "exactly " : issue.inclusive ? ">=" : ">";
        const sizing = getSizing(issue.origin);
        if (sizing) {
          return `Too small: expected ${issue.origin} to have ${adj}${issue.minimum.toString()} ${sizing.unit}`;
        }
        return `Too small: expected ${issue.origin} to be ${adj}${issue.minimum.toString()}`;
      }
      case "invalid_format": {
        const _issue = issue;
        if (_issue.format === "starts_with") {
          return `Invalid string: must start with "${_issue.prefix}"`;
        }
        if (_issue.format === "ends_with")
          return `Invalid string: must end with "${_issue.suffix}"`;
        if (_issue.format === "includes")
          return `Invalid string: must include "${_issue.includes}"`;
        if (_issue.format === "regex")
          return `Invalid string: must match pattern ${_issue.pattern}`;
        return `Invalid ${FormatDictionary[_issue.format] ?? issue.format}`;
      }
      case "not_multiple_of":
        return `Invalid number: must be a multiple of ${issue.divisor}`;
      case "unrecognized_keys":
        return `Unrecognized key${issue.keys.length > 1 ? "s" : ""}: ${joinValues(issue.keys, ", ")}`;
      case "invalid_key":
        return `Invalid key in ${issue.origin}`;
      case "invalid_union":
        if (issue.options && Array.isArray(issue.options) && issue.options.length > 0) {
          const opts = issue.options.map((o) => `'${o}'`).join(" | ");
          return `Invalid discriminator value. Expected ${opts}`;
        }
        if (issue.inclusive === false) {
          return "Invalid input: more than one option matched";
        }
        return "Invalid input";
      case "invalid_element":
        return `Invalid value in ${issue.origin}`;
      default:
        return `Invalid input`;
    }
  };
};
function en_default() {
  return {
    localeError: error()
  };
}
// node_modules/zod/v4/core/registries.js
var _a2;
class $ZodRegistry {
  constructor() {
    this._map = new WeakMap;
    this._idmap = new Map;
  }
  add(schema, ..._meta) {
    const meta = _meta[0];
    this._map.set(schema, meta);
    if (meta && typeof meta === "object" && "id" in meta) {
      this._idmap.set(meta.id, schema);
    }
    return this;
  }
  clear() {
    this._map = new WeakMap;
    this._idmap = new Map;
    return this;
  }
  remove(schema) {
    const meta = this._map.get(schema);
    if (meta && typeof meta === "object" && "id" in meta) {
      this._idmap.delete(meta.id);
    }
    this._map.delete(schema);
    return this;
  }
  get(schema) {
    const p = schema._zod.parent;
    if (p) {
      const pm = { ...this.get(p) ?? {} };
      delete pm.id;
      const f = { ...pm, ...this._map.get(schema) };
      return Object.keys(f).length ? f : undefined;
    }
    return this._map.get(schema);
  }
  has(schema) {
    return this._map.has(schema);
  }
}
function registry() {
  return new $ZodRegistry;
}
(_a2 = globalThis).__zod_globalRegistry ?? (_a2.__zod_globalRegistry = registry());
var globalRegistry = globalThis.__zod_globalRegistry;
// node_modules/zod/v4/core/api.js
function snapshotChecks(def) {
  if (def.checks)
    def.checks = [...def.checks];
  return def;
}
function _string(Class, params) {
  return new Class(snapshotChecks({ type: "string", ...normalizeParams(params) }));
}
function _coercedString(Class, params) {
  return new Class(snapshotChecks({ type: "string", coerce: true, ...normalizeParams(params) }));
}
function _email(Class, params) {
  return new Class({
    type: "string",
    format: "email",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _guid(Class, params) {
  return new Class({
    type: "string",
    format: "guid",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _uuid(Class, params) {
  return new Class({
    type: "string",
    format: "uuid",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _uuidv4(Class, params) {
  return new Class({
    type: "string",
    format: "uuid",
    check: "string_format",
    abort: false,
    version: "v4",
    ...normalizeParams(params)
  });
}
function _uuidv6(Class, params) {
  return new Class({
    type: "string",
    format: "uuid",
    check: "string_format",
    abort: false,
    version: "v6",
    ...normalizeParams(params)
  });
}
function _uuidv7(Class, params) {
  return new Class({
    type: "string",
    format: "uuid",
    check: "string_format",
    abort: false,
    version: "v7",
    ...normalizeParams(params)
  });
}
function _url(Class, params) {
  return new Class({
    type: "string",
    format: "url",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _emoji2(Class, params) {
  return new Class({
    type: "string",
    format: "emoji",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _nanoid(Class, params) {
  return new Class({
    type: "string",
    format: "nanoid",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _cuid(Class, params) {
  return new Class({
    type: "string",
    format: "cuid",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _cuid2(Class, params) {
  return new Class({
    type: "string",
    format: "cuid2",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _ulid(Class, params) {
  return new Class({
    type: "string",
    format: "ulid",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _xid(Class, params) {
  return new Class({
    type: "string",
    format: "xid",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _ksuid(Class, params) {
  return new Class({
    type: "string",
    format: "ksuid",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _ipv4(Class, params) {
  return new Class({
    type: "string",
    format: "ipv4",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _ipv6(Class, params) {
  return new Class({
    type: "string",
    format: "ipv6",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _mac(Class, params) {
  return new Class({
    type: "string",
    format: "mac",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _cidrv4(Class, params) {
  return new Class({
    type: "string",
    format: "cidrv4",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _cidrv6(Class, params) {
  return new Class({
    type: "string",
    format: "cidrv6",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _base64(Class, params) {
  return new Class({
    type: "string",
    format: "base64",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _base64url(Class, params) {
  return new Class({
    type: "string",
    format: "base64url",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _e164(Class, params) {
  return new Class({
    type: "string",
    format: "e164",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _creditCard(Class, params) {
  return new Class({
    type: "string",
    format: "credit_card",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _iban(Class, params) {
  return new Class({
    type: "string",
    format: "iban",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _jwt(Class, params) {
  return new Class({
    type: "string",
    format: "jwt",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
function _isoDateTime(Class, params) {
  return new Class({
    type: "string",
    format: "datetime",
    check: "string_format",
    offset: false,
    local: false,
    precision: null,
    ...normalizeParams(params)
  });
}
function _isoDate(Class, params) {
  return new Class({
    type: "string",
    format: "date",
    check: "string_format",
    ...normalizeParams(params)
  });
}
function _isoTime(Class, params) {
  return new Class({
    type: "string",
    format: "time",
    check: "string_format",
    precision: null,
    ...normalizeParams(params)
  });
}
function _isoDuration(Class, params) {
  return new Class({
    type: "string",
    format: "duration",
    check: "string_format",
    ...normalizeParams(params)
  });
}
function _number(Class, params) {
  return new Class(snapshotChecks({ type: "number", checks: [], ...normalizeParams(params) }));
}
function _coercedNumber(Class, params) {
  return new Class(snapshotChecks({ type: "number", coerce: true, checks: [], ...normalizeParams(params) }));
}
function _int(Class, params) {
  return new Class({
    type: "number",
    check: "number_format",
    abort: false,
    format: "safeint",
    ...normalizeParams(params)
  });
}
function _float32(Class, params) {
  return new Class({
    type: "number",
    check: "number_format",
    abort: false,
    format: "float32",
    ...normalizeParams(params)
  });
}
function _float64(Class, params) {
  return new Class({
    type: "number",
    check: "number_format",
    abort: false,
    format: "float64",
    ...normalizeParams(params)
  });
}
function _int32(Class, params) {
  return new Class({
    type: "number",
    check: "number_format",
    abort: false,
    format: "int32",
    ...normalizeParams(params)
  });
}
function _uint32(Class, params) {
  return new Class({
    type: "number",
    check: "number_format",
    abort: false,
    format: "uint32",
    ...normalizeParams(params)
  });
}
function _boolean(Class, params) {
  return new Class({
    type: "boolean",
    ...normalizeParams(params)
  });
}
function _coercedBoolean(Class, params) {
  return new Class({
    type: "boolean",
    coerce: true,
    ...normalizeParams(params)
  });
}
function _bigint(Class, params) {
  return new Class({
    type: "bigint",
    ...normalizeParams(params)
  });
}
function _coercedBigint(Class, params) {
  return new Class({
    type: "bigint",
    coerce: true,
    ...normalizeParams(params)
  });
}
function _int64(Class, params) {
  return new Class({
    type: "bigint",
    check: "bigint_format",
    abort: false,
    format: "int64",
    ...normalizeParams(params)
  });
}
function _uint64(Class, params) {
  return new Class({
    type: "bigint",
    check: "bigint_format",
    abort: false,
    format: "uint64",
    ...normalizeParams(params)
  });
}
function _symbol(Class, params) {
  return new Class({
    type: "symbol",
    ...normalizeParams(params)
  });
}
function _undefined2(Class, params) {
  return new Class({
    type: "undefined",
    ...normalizeParams(params)
  });
}
function _null2(Class, params) {
  return new Class({
    type: "null",
    ...normalizeParams(params)
  });
}
function _any(Class) {
  return new Class({
    type: "any"
  });
}
function _unknown(Class) {
  return new Class({
    type: "unknown"
  });
}
function _never(Class, params) {
  return new Class({
    type: "never",
    ...normalizeParams(params)
  });
}
function _void(Class, params) {
  return new Class({
    type: "void",
    ...normalizeParams(params)
  });
}
function _date(Class, params) {
  return new Class({
    type: "date",
    ...normalizeParams(params)
  });
}
function _coercedDate(Class, params) {
  return new Class({
    type: "date",
    coerce: true,
    ...normalizeParams(params)
  });
}
function _nan(Class, params) {
  return new Class({
    type: "nan",
    ...normalizeParams(params)
  });
}
function _lt(value, params) {
  return new $ZodCheckLessThan({
    check: "less_than",
    ...normalizeParams(params),
    value,
    inclusive: false
  });
}
function _lte(value, params) {
  return new $ZodCheckLessThan({
    check: "less_than",
    ...normalizeParams(params),
    value,
    inclusive: true
  });
}
function _gt(value, params) {
  return new $ZodCheckGreaterThan({
    check: "greater_than",
    ...normalizeParams(params),
    value,
    inclusive: false
  });
}
function _gte(value, params) {
  return new $ZodCheckGreaterThan({
    check: "greater_than",
    ...normalizeParams(params),
    value,
    inclusive: true
  });
}
function _positive(params) {
  return _gt(0, params);
}
function _negative(params) {
  return _lt(0, params);
}
function _nonpositive(params) {
  return _lte(0, params);
}
function _nonnegative(params) {
  return _gte(0, params);
}
function _multipleOf(value, params) {
  return new $ZodCheckMultipleOf({
    check: "multiple_of",
    ...normalizeParams(params),
    value
  });
}
function _maxSize(maximum, params) {
  return new $ZodCheckMaxSize({
    check: "max_size",
    ...normalizeParams(params),
    maximum
  });
}
function _minSize(minimum, params) {
  return new $ZodCheckMinSize({
    check: "min_size",
    ...normalizeParams(params),
    minimum
  });
}
function _size(size, params) {
  return new $ZodCheckSizeEquals({
    check: "size_equals",
    ...normalizeParams(params),
    size
  });
}
function _maxLength(maximum, params) {
  const ch = new $ZodCheckMaxLength({
    check: "max_length",
    ...normalizeParams(params),
    maximum
  });
  return ch;
}
function _minLength(minimum, params) {
  return new $ZodCheckMinLength({
    check: "min_length",
    ...normalizeParams(params),
    minimum
  });
}
function _length(length, params) {
  return new $ZodCheckLengthEquals({
    check: "length_equals",
    ...normalizeParams(params),
    length
  });
}
function _regex(pattern, params) {
  return new $ZodCheckRegex({
    check: "string_format",
    format: "regex",
    ...normalizeParams(params),
    pattern
  });
}
function _lowercase(params) {
  return new $ZodCheckLowerCase({
    check: "string_format",
    format: "lowercase",
    ...normalizeParams(params)
  });
}
function _uppercase(params) {
  return new $ZodCheckUpperCase({
    check: "string_format",
    format: "uppercase",
    ...normalizeParams(params)
  });
}
function _includes(includes, params) {
  return new $ZodCheckIncludes({
    check: "string_format",
    format: "includes",
    ...normalizeParams(params),
    includes
  });
}
function _startsWith(prefix, params) {
  return new $ZodCheckStartsWith({
    check: "string_format",
    format: "starts_with",
    ...normalizeParams(params),
    prefix
  });
}
function _endsWith(suffix, params) {
  return new $ZodCheckEndsWith({
    check: "string_format",
    format: "ends_with",
    ...normalizeParams(params),
    suffix
  });
}
function _property(property, schema, params) {
  return new $ZodCheckProperty({
    check: "property",
    property,
    schema,
    ...normalizeParams(params)
  });
}
function _properties(shape, params) {
  return new $ZodCheckProperties({
    check: "properties",
    shape,
    ...normalizeParams(params)
  });
}
function _mime(types, params) {
  return new $ZodCheckMimeType({
    check: "mime_type",
    mime: types,
    ...normalizeParams(params)
  });
}
function _overwrite(tx) {
  return new $ZodCheckOverwrite({
    check: "overwrite",
    tx
  });
}
function _normalize(form) {
  return _overwrite((input) => input.normalize(form));
}
function _trim() {
  return _overwrite((input) => input.trim());
}
function _toLowerCase() {
  return _overwrite((input) => input.toLowerCase());
}
function _toUpperCase() {
  return _overwrite((input) => input.toUpperCase());
}
function _slugify() {
  return _overwrite((input) => slugify(input));
}
function _array(Class, element, params) {
  return new Class({
    type: "array",
    element,
    ...normalizeParams(params)
  });
}
function _file(Class, params) {
  return new Class({
    type: "file",
    ...normalizeParams(params)
  });
}
function _custom(Class, fn, _params) {
  const norm = normalizeParams(_params);
  norm.abort ?? (norm.abort = true);
  const schema = new Class({
    type: "custom",
    check: "custom",
    fn,
    ...norm
  });
  return schema;
}
function _refine(Class, fn, _params) {
  const schema = new Class({
    type: "custom",
    check: "custom",
    fn,
    ...normalizeParams(_params)
  });
  return schema;
}
function _superRefine(fn, params) {
  const ch = _check((payload) => {
    payload.addIssue = (issue2) => {
      if (typeof issue2 === "string") {
        payload.issues.push(issue(issue2, payload.value, ch._zod.def));
      } else {
        const _issue = issue2;
        if (_issue.fatal)
          _issue.continue = false;
        _issue.code ?? (_issue.code = "custom");
        if (!("input" in _issue))
          _issue.input = payload.value;
        _issue.inst ?? (_issue.inst = ch);
        _issue.continue ?? (_issue.continue = !ch._zod.def.abort);
        payload.issues.push(issue(_issue));
      }
    };
    return fn(payload.value, payload);
  }, params);
  return ch;
}
function _check(fn, params) {
  const ch = new $ZodCheck({
    check: "custom",
    ...normalizeParams(params)
  });
  ch._zod.check = fn;
  return ch;
}
function describe(description) {
  const ch = new $ZodCheck({ check: "describe" });
  ch._zod.onattach = [
    (inst) => {
      const existing = globalRegistry.get(inst) ?? {};
      globalRegistry.add(inst, { ...existing, description });
    }
  ];
  ch._zod.check = () => {};
  return ch;
}
function meta(metadata) {
  const ch = new $ZodCheck({ check: "meta" });
  ch._zod.onattach = [
    (inst) => {
      const existing = globalRegistry.get(inst) ?? {};
      globalRegistry.add(inst, { ...existing, ...metadata });
    }
  ];
  ch._zod.check = () => {};
  return ch;
}
function _stringbool(Classes, _params) {
  const params = normalizeParams(_params);
  let truthyArray = params.truthy ?? ["true", "1", "yes", "on", "y", "enabled"];
  let falsyArray = params.falsy ?? ["false", "0", "no", "off", "n", "disabled"];
  if (params.case !== "sensitive") {
    truthyArray = truthyArray.map((v) => typeof v === "string" ? v.toLowerCase() : v);
    falsyArray = falsyArray.map((v) => typeof v === "string" ? v.toLowerCase() : v);
  }
  const truthySet = new Set(truthyArray);
  const falsySet = new Set(falsyArray);
  const _Codec = Classes.Codec ?? $ZodCodec;
  const _Boolean = Classes.Boolean ?? $ZodBoolean;
  const _String = Classes.String ?? $ZodString;
  const stringSchema = new _String({ type: "string", error: params.error });
  const booleanSchema = new _Boolean({ type: "boolean", error: params.error });
  const codec = new _Codec({
    type: "pipe",
    in: stringSchema,
    out: booleanSchema,
    transform: (input, payload) => {
      let data = input;
      if (params.case !== "sensitive")
        data = data.toLowerCase();
      if (truthySet.has(data)) {
        return true;
      } else if (falsySet.has(data)) {
        return false;
      } else {
        payload.issues.push({
          code: "invalid_value",
          expected: "stringbool",
          values: [...truthySet, ...falsySet],
          input: payload.value,
          inst: codec,
          continue: false
        });
        return {};
      }
    },
    reverseTransform: (input, _payload) => {
      if (input === true) {
        return truthyArray[0] || "true";
      } else {
        return falsyArray[0] || "false";
      }
    },
    error: params.error
  });
  codec._zod.bag.truthy = truthyArray;
  codec._zod.bag.falsy = falsyArray;
  codec._zod.bag.case = params.case ?? "insensitive";
  return codec;
}
function _stringFormat(Class, format, fnOrRegex, _params = {}) {
  const params = normalizeParams(_params);
  const def = {
    check: "string_format",
    type: "string",
    format,
    fn: typeof fnOrRegex === "function" ? fnOrRegex : (val) => fnOrRegex.test(val),
    ...params
  };
  if (fnOrRegex instanceof RegExp) {
    def.pattern = fnOrRegex;
  }
  const inst = new Class(def);
  return inst;
}
// node_modules/zod/v4/core/to-json-schema.js
function assignProps(target, ...sources) {
  for (const source of sources) {
    for (const key of Reflect.ownKeys(source)) {
      if (Object.prototype.propertyIsEnumerable.call(source, key)) {
        assignProp(target, key, source[key]);
      }
    }
  }
  return target;
}
function initializeContext(params) {
  let target = params?.target ?? "draft-2020-12";
  if (target === "draft-4")
    target = "draft-04";
  if (target === "draft-7")
    target = "draft-07";
  return {
    processors: params.processors ?? {},
    metadataRegistry: params?.metadata ?? globalRegistry,
    target,
    unrepresentable: params?.unrepresentable ?? "throw",
    override: params?.override ?? (() => {}),
    io: params?.io ?? "output",
    counter: 0,
    seen: new Map,
    sharedDefsExtractedFor: undefined,
    sharedEmitDoneFor: undefined,
    cycles: params?.cycles ?? "ref",
    reused: params?.reused ?? "inline",
    intersections: [],
    deferred: [],
    external: params?.external ?? undefined
  };
}
function handleUnrepresentable(schema, ctx, json, params, message) {
  const result = typeof ctx.unrepresentable === "function" ? ctx.unrepresentable({ zodSchema: schema, path: params.path, message }) : ctx.unrepresentable;
  if (result === "any")
    return false;
  if (result === undefined || result === "throw")
    throw new Error(message);
  Object.assign(json, result);
  return true;
}
function processSchema(schema, ctx, _params = { path: [], schemaPath: [] }) {
  var _a;
  const def = schema._zod.def;
  const seen = ctx.seen.get(schema);
  if (seen) {
    seen.count++;
    const isCycle = _params.schemaPath.includes(schema);
    if (isCycle) {
      seen.cycle = _params.path;
    }
    return seen.schema;
  }
  const result = { schema: {}, count: 1, cycle: undefined, path: _params.path };
  ctx.seen.set(schema, result);
  ctx.sharedDefsExtractedFor = undefined;
  ctx.sharedEmitDoneFor = undefined;
  const overrideSchema = schema._zod.toJSONSchema?.();
  if (overrideSchema) {
    result.schema = overrideSchema;
  } else {
    const params = {
      ..._params,
      schemaPath: [..._params.schemaPath, schema],
      path: _params.path
    };
    if (schema._zod.processJSONSchema) {
      schema._zod.processJSONSchema(ctx, result.schema, params);
    } else {
      const _json = result.schema;
      const processor = ctx.processors[def.type];
      if (!processor) {
        throw new Error(`[toJSONSchema]: Non-representable type encountered: ${def.type}`);
      }
      processor(schema, ctx, _json, params);
    }
    const parent = schema._zod.parent;
    if (parent) {
      if (!result.ref)
        result.ref = parent;
      processSchema(parent, ctx, params);
      ctx.seen.get(parent).isParent = true;
    }
  }
  const meta = ctx.metadataRegistry.get(schema);
  if (meta)
    assignProps(result.schema, meta);
  if (ctx.io === "input" && isTransforming(schema)) {
    delete result.schema.examples;
    delete result.schema.default;
  }
  if (ctx.io === "input" && "_prefault" in result.schema)
    (_a = result.schema).default ?? (_a.default = result.schema._prefault);
  delete result.schema._prefault;
  const _result = ctx.seen.get(schema);
  return _result.schema;
}
function encodeJSONPointerSegment(segment) {
  return segment.replace(/~/g, "~0").replace(/\//g, "~1");
}
function extractDefs(ctx, schema) {
  const root = ctx.seen.get(schema);
  if (!root)
    throw new Error("Unprocessed schema. This is a bug in Zod.");
  if (ctx.external && ctx.sharedDefsExtractedFor === ctx.external)
    return;
  const idToSchema = new Map;
  for (const entry of ctx.seen.entries()) {
    const id = ctx.metadataRegistry.get(entry[0])?.id;
    if (id) {
      const existing = idToSchema.get(id);
      if (existing && existing !== entry[0]) {
        throw new Error(`Duplicate schema id "${id}" detected during JSON Schema conversion. Two different schemas cannot share the same id when converted together.`);
      }
      idToSchema.set(id, entry[0]);
    }
  }
  const makeURI = (entry) => {
    const defsSegment = ctx.target === "draft-2020-12" ? "$defs" : "definitions";
    if (ctx.external) {
      const externalId = ctx.external.registry.get(entry[0])?.id;
      const uriGenerator = ctx.external.uri ?? ((id) => id);
      if (externalId) {
        return { ref: uriGenerator(externalId) };
      }
      const id = entry[1].defId ?? entry[1].schema.id ?? `schema${ctx.counter++}`;
      entry[1].defId = id;
      return { defId: id, ref: `${uriGenerator("__shared")}#/${defsSegment}/${encodeJSONPointerSegment(id)}` };
    }
    const uriPrefix = `#`;
    const defUriPrefix = `${uriPrefix}/${defsSegment}/`;
    if (entry[1] === root && !entry[1].schema.id) {
      return { ref: uriPrefix };
    }
    const defId = entry[1].schema.id ?? `__schema${ctx.counter++}`;
    return { defId, ref: defUriPrefix + encodeJSONPointerSegment(defId) };
  };
  const extractToDef = (entry) => {
    if (entry[1].schema.$ref) {
      return;
    }
    const seen = entry[1];
    const { ref, defId } = makeURI(entry);
    seen.def = { ...seen.schema };
    if (defId)
      seen.defId = defId;
    const schema = seen.schema;
    for (const key in schema) {
      delete schema[key];
    }
    schema.$ref = ref;
  };
  if (ctx.cycles === "throw") {
    for (const entry of ctx.seen.entries()) {
      const seen = entry[1];
      if (seen.cycle) {
        throw new Error("Cycle detected: " + `#/${seen.cycle?.join("/")}/<root>` + '\n\nSet the `cycles` parameter to `"ref"` to resolve cyclical schemas with defs.');
      }
    }
  }
  for (const entry of ctx.seen.entries()) {
    const seen = entry[1];
    if (schema === entry[0]) {
      extractToDef(entry);
      continue;
    }
    if (ctx.external) {
      const ext = ctx.external.registry.get(entry[0])?.id;
      if (schema !== entry[0] && ext) {
        extractToDef(entry);
        continue;
      }
    }
    const id = ctx.metadataRegistry.get(entry[0])?.id;
    if (id) {
      extractToDef(entry);
      continue;
    }
    if (seen.cycle) {
      extractToDef(entry);
      continue;
    }
    if (seen.count > 1) {
      if (ctx.reused === "ref") {
        extractToDef(entry);
      }
    }
  }
  if (ctx.external)
    ctx.sharedDefsExtractedFor = ctx.external;
}
function compactTypeUnion(schema) {
  const options = schema.anyOf;
  if (!Array.isArray(options) || options.length === 0 || schema.type !== undefined)
    return;
  const types = [];
  for (const option of options) {
    if (!option || typeof option !== "object")
      return;
    compactTypeUnion(option);
    const keys = Object.keys(option);
    if (keys.length !== 1 || keys[0] !== "type")
      return;
    const type = option.type;
    for (const member of Array.isArray(type) ? type : [type]) {
      if (typeof member !== "string")
        return;
      if (!types.includes(member))
        types.push(member);
    }
  }
  delete schema.anyOf;
  schema.type = types.length === 1 ? types[0] : types;
}
var FOLDABLE_KEYS = new Set(["type", "properties", "required", "additionalProperties"]);
var UNION_KEYS = ["oneOf", "anyOf"];
function undeclaredConstraint(member) {
  const extra = member.additionalProperties;
  if (extra === undefined || extra === false || typeof extra !== "object" || extra === null)
    return null;
  return Object.keys(extra).length ? extra : null;
}
function foldObjects(members) {
  const objects = [];
  for (const member of members) {
    if (typeof member !== "object" || member.type !== "object")
      return null;
    for (const key in member) {
      if (!FOLDABLE_KEYS.has(key))
        return null;
    }
    objects.push(member);
  }
  const properties = {};
  const required = new Set;
  for (const object of objects) {
    for (const key in object.properties) {
      if (Object.prototype.hasOwnProperty.call(properties, key))
        continue;
      const parts = [];
      for (const other of objects) {
        const part = other.properties?.[key] ?? undeclaredConstraint(other);
        if (part === null || part === undefined)
          continue;
        if (!parts.some((seen) => JSON.stringify(seen) === JSON.stringify(part)))
          parts.push(part);
      }
      const merged = parts.length === 1 ? parts[0] : foldObjects(parts) ?? { allOf: parts };
      assignProp(properties, key, merged);
    }
    for (const key of object.required ?? [])
      required.add(key);
  }
  const folded = { type: "object", properties };
  if (required.size)
    folded.required = [...required];
  if (objects.every((object) => object.additionalProperties === false)) {
    folded.additionalProperties = false;
  } else {
    const constraints = [];
    for (const object of objects) {
      const constraint = undeclaredConstraint(object);
      if (constraint && !constraints.some((seen) => JSON.stringify(seen) === JSON.stringify(constraint)))
        constraints.push(constraint);
    }
    if (constraints.length === 1)
      folded.additionalProperties = constraints[0];
    else if (constraints.length > 1)
      folded.additionalProperties = { allOf: constraints };
  }
  return folded;
}
function foldIntersection(json) {
  const allOf = json.allOf;
  if (!Array.isArray(allOf) || allOf.length < 2)
    return;
  for (const key of FOLDABLE_KEYS)
    if (key in json)
      return;
  const unions = allOf.filter((m) => UNION_KEYS.some((k) => Array.isArray(m[k])));
  let folded = null;
  if (!unions.length) {
    folded = foldObjects(allOf);
  } else {
    const union = unions[0];
    const keyword = UNION_KEYS.find((k) => Array.isArray(union[k]));
    if (Object.keys(union).length !== 1)
      return;
    const rest = allOf.filter((m) => m !== union);
    const branches = union[keyword].map((branch) => foldObjects([...rest, branch]));
    if (branches.some((b) => !b))
      return;
    folded = { [keyword]: branches };
  }
  if (!folded)
    return;
  delete json.allOf;
  assignProps(json, folded);
}
function finalize(ctx, schema) {
  const root = ctx.seen.get(schema);
  if (!root)
    throw new Error("Unprocessed schema. This is a bug in Zod.");
  const flattenRef = (zodSchema) => {
    const seen = ctx.seen.get(zodSchema);
    if (seen.ref === null)
      return;
    const schema = seen.def ?? seen.schema;
    const _cached = { ...schema };
    const ref = seen.ref;
    seen.ref = null;
    if (ref) {
      flattenRef(ref);
      const refSeen = ctx.seen.get(ref);
      const refSchema = refSeen.schema;
      if (refSchema.$ref && (ctx.target === "draft-07" || ctx.target === "draft-04" || ctx.target === "openapi-3.0")) {
        schema.allOf = schema.allOf ?? [];
        schema.allOf.push(refSchema);
      } else {
        assignProps(schema, refSchema);
      }
      assignProps(schema, _cached);
      const isParentRef = zodSchema._zod.parent === ref;
      if (isParentRef) {
        for (const key in schema) {
          if (key === "$ref" || key === "allOf")
            continue;
          if (!(key in _cached)) {
            delete schema[key];
          }
        }
      }
      if (refSchema.$ref && refSeen.def) {
        for (const key in schema) {
          if (key === "$ref" || key === "allOf")
            continue;
          if (key in refSeen.def && JSON.stringify(schema[key]) === JSON.stringify(refSeen.def[key])) {
            delete schema[key];
          }
        }
      }
    }
    const parent = zodSchema._zod.parent;
    if (parent && parent !== ref) {
      flattenRef(parent);
      const parentSeen = ctx.seen.get(parent);
      if (parentSeen?.schema.$ref) {
        schema.$ref = parentSeen.schema.$ref;
        if (parentSeen.def) {
          for (const key in schema) {
            if (key === "$ref" || key === "allOf")
              continue;
            if (key in parentSeen.def && JSON.stringify(schema[key]) === JSON.stringify(parentSeen.def[key])) {
              delete schema[key];
            }
          }
        }
      }
    }
    ctx.override({
      zodSchema,
      jsonSchema: schema,
      path: seen.path ?? []
    });
  };
  if (!ctx.external || ctx.sharedEmitDoneFor !== ctx.external) {
    for (const entry of [...ctx.seen.entries()].reverse()) {
      flattenRef(entry[0]);
    }
    if (ctx.target !== "openapi-3.0") {
      for (const entry of ctx.seen.entries()) {
        compactTypeUnion(entry[1].def ?? entry[1].schema);
      }
    }
    for (const rewrite of ctx.deferred)
      rewrite();
    if (ctx.intersections.length) {
      const carriers = new Map;
      for (const seen of ctx.seen.values()) {
        for (const json of [seen.schema, seen.def]) {
          const allOf = json?.allOf;
          if (!Array.isArray(allOf))
            continue;
          const existing = carriers.get(allOf);
          if (existing)
            existing.push(json);
          else
            carriers.set(allOf, [json]);
        }
      }
      for (const allOf of ctx.intersections) {
        for (const json of carriers.get(allOf) ?? [])
          foldIntersection(json);
      }
    }
  }
  const result = {};
  if (ctx.target === "draft-2020-12") {
    result.$schema = "https://json-schema.org/draft/2020-12/schema";
  } else if (ctx.target === "draft-07") {
    result.$schema = "http://json-schema.org/draft-07/schema#";
  } else if (ctx.target === "draft-04") {
    result.$schema = "http://json-schema.org/draft-04/schema#";
  } else if (ctx.target === "openapi-3.0") {}
  if (ctx.external?.uri) {
    const id = ctx.external.registry.get(schema)?.id;
    if (!id)
      throw new Error("Schema is missing an `id` property");
    result.$id = ctx.external.uri(id);
  }
  assignProps(result, root.defId ? root.schema : root.def ?? root.schema);
  const rootMetaId = ctx.metadataRegistry.get(schema)?.id;
  if (rootMetaId !== undefined && result.id === rootMetaId)
    delete result.id;
  const defs = ctx.external?.defs ?? {};
  if (!ctx.external || ctx.sharedEmitDoneFor !== ctx.external) {
    for (const entry of ctx.seen.entries()) {
      const seen = entry[1];
      if (seen.def && seen.defId) {
        if (seen.def.id === seen.defId)
          delete seen.def.id;
        assignProp(defs, seen.defId, seen.def);
      }
    }
  }
  if (ctx.external)
    ctx.sharedEmitDoneFor = ctx.external;
  if (ctx.external) {} else {
    if (Object.keys(defs).length > 0) {
      if (ctx.target === "draft-2020-12") {
        result.$defs = defs;
      } else {
        result.definitions = defs;
      }
    }
  }
  try {
    const finalized = JSON.parse(JSON.stringify(result));
    Object.defineProperty(finalized, "~standard", {
      value: {
        ...schema["~standard"],
        jsonSchema: {
          input: createStandardJSONSchemaMethod(schema, "input", ctx.processors),
          output: createStandardJSONSchemaMethod(schema, "output", ctx.processors)
        }
      },
      enumerable: false,
      writable: false
    });
    return finalized;
  } catch (_err) {
    throw new Error("Error converting schema to JSON.");
  }
}
function isTransforming(_schema, _ctx) {
  const ctx = _ctx ?? { seen: new Set };
  if (ctx.seen.has(_schema))
    return false;
  ctx.seen.add(_schema);
  const def = _schema._zod.def;
  if (def.type === "transform")
    return true;
  if (def.type === "array")
    return isTransforming(def.element, ctx);
  if (def.type === "set")
    return isTransforming(def.valueType, ctx);
  if (def.type === "lazy")
    return isTransforming(def.getter(), ctx);
  if (def.type === "promise" || def.type === "optional" || def.type === "nonoptional" || def.type === "nullable" || def.type === "readonly" || def.type === "default" || def.type === "prefault" || def.type === "catch") {
    return isTransforming(def.innerType, ctx);
  }
  if (def.type === "intersection") {
    return isTransforming(def.left, ctx) || isTransforming(def.right, ctx);
  }
  if (def.type === "record" || def.type === "map") {
    return isTransforming(def.keyType, ctx) || isTransforming(def.valueType, ctx);
  }
  if (def.type === "pipe") {
    if (_schema._zod.traits.has("$ZodCodec"))
      return true;
    return isTransforming(def.in, ctx) || isTransforming(def.out, ctx);
  }
  if (def.type === "object") {
    for (const key in def.shape) {
      if (isTransforming(def.shape[key], ctx))
        return true;
    }
    return false;
  }
  if (def.type === "union") {
    for (const option of def.options) {
      if (isTransforming(option, ctx))
        return true;
    }
    return false;
  }
  if (def.type === "tuple") {
    for (const item of def.items) {
      if (isTransforming(item, ctx))
        return true;
    }
    if (def.rest && isTransforming(def.rest, ctx))
      return true;
    return false;
  }
  return false;
}
var createToJSONSchemaMethod = (schema, processors = {}) => (params) => {
  const ctx = initializeContext({ ...params, processors });
  processSchema(schema, ctx);
  extractDefs(ctx, schema);
  return finalize(ctx, schema);
};
var createStandardJSONSchemaMethod = (schema, io, processors = {}) => (params) => {
  const { libraryOptions, target } = params ?? {};
  const ctx = initializeContext({ ...libraryOptions ?? {}, target, io, processors });
  processSchema(schema, ctx);
  extractDefs(ctx, schema);
  return finalize(ctx, schema);
};
// node_modules/zod/v4/core/json-schema-processors.js
var narrowMin = (agg, key, value) => {
  if (agg[key] === undefined || value > agg[key])
    agg[key] = value;
};
var narrowMax = (agg, key, value) => {
  if (agg[key] === undefined || value < agg[key])
    agg[key] = value;
};
var narrowBoth = (agg, value) => {
  narrowMin(agg, "minimum", value);
  narrowMax(agg, "maximum", value);
};
var addDivisor = (agg, value) => {
  agg.multipleOf ?? (agg.multipleOf = []);
  if (!agg.multipleOf.includes(value))
    agg.multipleOf.push(value);
};
var addPattern = (agg, pattern) => {
  agg.patterns ?? (agg.patterns = new Set);
  agg.patterns.add(pattern);
};
var intersectMime = (agg, mime) => {
  agg.mime = agg.mime ? agg.mime.filter((m) => mime.includes(m)) : [...mime];
};
var setFormat = (agg, format) => {
  agg.format = format;
  if (format.includes("int"))
    agg.isInt = true;
};
var minContributor = (agg, def) => narrowMin(agg, "minimum", def.minimum);
var maxContributor = (agg, def) => narrowMax(agg, "maximum", def.maximum);
var formatContributor = (ranges) => (agg, def) => {
  setFormat(agg, def.format);
  const [minimum, maximum] = ranges[def.format];
  narrowMin(agg, "minimum", minimum);
  narrowMax(agg, "maximum", maximum);
};
var contributors = {
  greater_than: (agg, def) => narrowMin(agg, def.inclusive ? "minimum" : "exclusiveMinimum", def.value),
  less_than: (agg, def) => narrowMax(agg, def.inclusive ? "maximum" : "exclusiveMaximum", def.value),
  multiple_of: (agg, def) => addDivisor(agg, def.value),
  number_format: formatContributor(NUMBER_FORMAT_RANGES),
  bigint_format: formatContributor(BIGINT_FORMAT_RANGES),
  min_length: minContributor,
  max_length: maxContributor,
  length_equals: (agg, def) => narrowBoth(agg, def.length),
  min_size: minContributor,
  max_size: maxContributor,
  size_equals: (agg, def) => narrowBoth(agg, def.size),
  string_format: (agg, def) => {
    setFormat(agg, def.format);
    if (def.pattern)
      addPattern(agg, def.pattern);
    if (def.format === "base64" || def.format === "base64url")
      agg.contentEncoding = def.format;
    if (def.local || def.precision === -1)
      agg.laxFormat = true;
  },
  mime_type: (agg, def) => intersectMime(agg, def.mime)
};
function aggregateChecks(schema) {
  const agg = {};
  const def = schema._zod.def;
  const list = schema._zod.traits.has("$ZodCheck") ? [schema, ...def.checks ?? []] : def.checks ?? [];
  for (const ch of list)
    contributors[ch._zod.def.check]?.(agg, ch._zod.def);
  const bag = schema._zod.bag;
  if (bag.minimum !== undefined)
    narrowMin(agg, "minimum", bag.minimum);
  if (bag.exclusiveMinimum !== undefined)
    narrowMin(agg, "exclusiveMinimum", bag.exclusiveMinimum);
  if (bag.maximum !== undefined)
    narrowMax(agg, "maximum", bag.maximum);
  if (bag.exclusiveMaximum !== undefined)
    narrowMax(agg, "exclusiveMaximum", bag.exclusiveMaximum);
  if (bag.multipleOf !== undefined)
    addDivisor(agg, bag.multipleOf);
  if (bag.format !== undefined) {
    agg.format ?? (agg.format = bag.format);
    if (bag.format.includes("int"))
      agg.isInt = true;
  }
  if (bag.mime)
    intersectMime(agg, bag.mime);
  for (const pattern of bag.patterns ?? [])
    addPattern(agg, pattern);
  return agg;
}
var formatMap = {
  guid: "uuid",
  url: "uri",
  datetime: "date-time",
  json_string: "json-string",
  regex: ""
};
var exactPatterns = new Map([
  [base64Charset, base64],
  [base64urlCharset, base64url]
]);
var exactPattern = (p) => exactPatterns.get(p) ?? p;
var stringProcessor = (schema, ctx, _json, _params) => {
  const json = _json;
  json.type = "string";
  const { minimum, maximum, format, patterns, contentEncoding, laxFormat } = aggregateChecks(schema);
  if (typeof minimum === "number")
    json.minLength = minimum;
  if (typeof maximum === "number")
    json.maxLength = maximum;
  if (format) {
    json.format = formatMap[format] ?? format;
    if (json.format === "")
      delete json.format;
    if (format === "time" || laxFormat) {
      delete json.format;
    }
  }
  if (contentEncoding)
    json.contentEncoding = contentEncoding;
  if (patterns && patterns.size > 0) {
    const patternList = [...patterns].map(exactPattern);
    if (patternList.length === 1)
      json.pattern = patternList[0].source;
    else if (patternList.length > 1) {
      json.allOf = [
        ...patternList.map((regex) => ({
          ...ctx.target === "draft-07" || ctx.target === "draft-04" || ctx.target === "openapi-3.0" ? { type: "string" } : {},
          pattern: regex.source
        }))
      ];
    }
  }
};
var numberProcessor = (schema, ctx, _json, params) => {
  const json = _json;
  const { minimum, maximum, multipleOf, exclusiveMaximum, exclusiveMinimum, isInt } = aggregateChecks(schema);
  json.type = isInt ? "integer" : "number";
  const exMin = typeof exclusiveMinimum === "number" && exclusiveMinimum >= (minimum ?? Number.NEGATIVE_INFINITY);
  const exMax = typeof exclusiveMaximum === "number" && exclusiveMaximum <= (maximum ?? Number.POSITIVE_INFINITY);
  const legacy = ctx.target === "draft-04" || ctx.target === "openapi-3.0";
  if (exMin) {
    if (legacy) {
      json.minimum = exclusiveMinimum;
      json.exclusiveMinimum = true;
    } else {
      json.exclusiveMinimum = exclusiveMinimum;
    }
  } else if (typeof minimum === "number") {
    json.minimum = minimum;
  }
  if (exMax) {
    if (legacy) {
      json.maximum = exclusiveMaximum;
      json.exclusiveMaximum = true;
    } else {
      json.exclusiveMaximum = exclusiveMaximum;
    }
  } else if (typeof maximum === "number") {
    json.maximum = maximum;
  }
  if (multipleOf) {
    const divisors = new Set;
    for (const divisor of multipleOf) {
      if (Number.isFinite(divisor) && divisor !== 0)
        divisors.add(Math.abs(divisor));
      else
        handleUnrepresentable(schema, ctx, json, params, `A multipleOf divisor of ${divisor} cannot be represented in JSON Schema`);
    }
    const [first, ...rest] = divisors;
    if (first !== undefined)
      json.multipleOf = first;
    if (rest.length)
      json.allOf = [...json.allOf ?? [], ...rest.map((m) => ({ multipleOf: m }))];
  }
};
var booleanProcessor = (_schema, _ctx, json, _params) => {
  json.type = "boolean";
};
var bigintProcessor = (schema, ctx, json, params) => {
  handleUnrepresentable(schema, ctx, json, params, "BigInt cannot be represented in JSON Schema");
};
var symbolProcessor = (schema, ctx, json, params) => {
  handleUnrepresentable(schema, ctx, json, params, "Symbols cannot be represented in JSON Schema");
};
var nullProcessor = (_schema, ctx, json, _params) => {
  if (ctx.target === "openapi-3.0") {
    json.type = "string";
    json.nullable = true;
    json.enum = [null];
  } else {
    json.type = "null";
  }
};
var undefinedProcessor = (schema, ctx, json, params) => {
  handleUnrepresentable(schema, ctx, json, params, "Undefined cannot be represented in JSON Schema");
};
var voidProcessor = (schema, ctx, json, params) => {
  handleUnrepresentable(schema, ctx, json, params, "Void cannot be represented in JSON Schema");
};
var neverProcessor = (_schema, _ctx, json, _params) => {
  json.not = {};
};
var anyProcessor = (_schema, _ctx, _json, _params) => {};
var unknownProcessor = (_schema, _ctx, _json, _params) => {};
var dateProcessor = (schema, ctx, json, params) => {
  handleUnrepresentable(schema, ctx, json, params, "Date cannot be represented in JSON Schema");
};
var enumProcessor = (schema, _ctx, json, _params) => {
  const def = schema._zod.def;
  const values = getEnumValues(def.entries);
  if (values.length === 0) {
    json.not = {};
    return;
  }
  if (values.every((v) => typeof v === "number"))
    json.type = "number";
  if (values.every((v) => typeof v === "string"))
    json.type = "string";
  json.enum = values;
};
var literalProcessor = (schema, ctx, json, params) => {
  const def = schema._zod.def;
  if (def.values.length === 0) {
    json.not = {};
    return;
  }
  const vals = [];
  for (const val of def.values) {
    if (val === undefined) {
      if (handleUnrepresentable(schema, ctx, json, params, "Literal `undefined` cannot be represented in JSON Schema"))
        return;
    } else if (typeof val === "bigint") {
      if (handleUnrepresentable(schema, ctx, json, params, "BigInt literals cannot be represented in JSON Schema"))
        return;
      vals.push(Number(val));
    } else {
      vals.push(val);
    }
  }
  if (vals.length === 0) {} else if (vals.length === 1) {
    const val = vals[0];
    json.type = val === null ? "null" : typeof val;
    if (ctx.target === "draft-04" || ctx.target === "openapi-3.0") {
      json.enum = [val];
    } else {
      json.const = val;
    }
  } else {
    if (vals.every((v) => typeof v === "number"))
      json.type = "number";
    if (vals.every((v) => typeof v === "string"))
      json.type = "string";
    if (vals.every((v) => typeof v === "boolean"))
      json.type = "boolean";
    if (vals.every((v) => v === null))
      json.type = "null";
    json.enum = vals;
  }
};
var nanProcessor = (schema, ctx, json, params) => {
  handleUnrepresentable(schema, ctx, json, params, "NaN cannot be represented in JSON Schema");
};
var templateLiteralProcessor = (schema, _ctx, json, _params) => {
  const _json = json;
  const pattern = schema._zod.pattern;
  if (!pattern)
    throw new Error("Pattern not found in template literal");
  _json.type = "string";
  _json.pattern = pattern.source;
};
var fileProcessor = (schema, _ctx, json, _params) => {
  const _json = json;
  _json.type = "string";
  _json.format = "binary";
  _json.contentEncoding = "binary";
  const { minimum, maximum, mime } = aggregateChecks(schema);
  if (minimum !== undefined)
    _json.minLength = minimum;
  if (maximum !== undefined)
    _json.maxLength = maximum;
  if (!mime)
    return;
  if (mime.length === 0)
    _json.not = {};
  else if (mime.length === 1)
    _json.contentMediaType = mime[0];
  else
    _json.anyOf = mime.map((m) => ({ contentMediaType: m }));
};
var successProcessor = (_schema, _ctx, json, _params) => {
  json.type = "boolean";
};
var customProcessor = (schema, ctx, json, params) => {
  handleUnrepresentable(schema, ctx, json, params, "Custom types cannot be represented in JSON Schema");
};
var functionProcessor = (schema, ctx, json, params) => {
  handleUnrepresentable(schema, ctx, json, params, "Function types cannot be represented in JSON Schema");
};
var transformProcessor = (schema, ctx, json, params) => {
  handleUnrepresentable(schema, ctx, json, params, "Transforms cannot be represented in JSON Schema");
};
var mapProcessor = (schema, ctx, json, params) => {
  handleUnrepresentable(schema, ctx, json, params, "Map cannot be represented in JSON Schema");
};
var setProcessor = (schema, ctx, json, params) => {
  handleUnrepresentable(schema, ctx, json, params, "Set cannot be represented in JSON Schema");
};
var arrayProcessor = (schema, ctx, _json, params) => {
  const json = _json;
  const def = schema._zod.def;
  const { minimum, maximum } = aggregateChecks(schema);
  if (typeof minimum === "number")
    json.minItems = minimum;
  if (typeof maximum === "number")
    json.maxItems = maximum;
  json.type = "array";
  json.items = processSchema(def.element, ctx, {
    ...params,
    path: [...params.path, "items"]
  });
};
function inputOptin(schema) {
  const def = schema._zod.def;
  if (def.type === "pipe" && def.in._zod.traits.has("$ZodTransform")) {
    return inputOptin(def.out);
  }
  if (def.type === "catch") {
    return inputOptin(def.innerType);
  }
  return schema._zod.optin;
}
var objectProcessor = (schema, ctx, _json, params) => {
  const json = _json;
  const def = schema._zod.def;
  const shape = def.shape;
  const symbolKeys = Object.getOwnPropertySymbols(shape);
  if (symbolKeys.length && handleUnrepresentable(schema, ctx, json, params, "Symbol keys cannot be represented in JSON Schema")) {
    return;
  }
  json.type = "object";
  json.properties = {};
  for (const key in shape) {
    assignProp(json.properties, key, processSchema(shape[key], ctx, {
      ...params,
      path: [...params.path, "properties", key]
    }));
  }
  const requiredKeys = [];
  for (const key of Object.keys(shape)) {
    const field = def.shape[key];
    if (ctx.io === "input" ? inputOptin(field) === undefined : field._zod.optout === undefined) {
      requiredKeys.push(key);
    }
  }
  if (requiredKeys.length > 0) {
    json.required = requiredKeys;
  }
  if (def.catchall?._zod.def.type === "never") {
    json.additionalProperties = false;
  } else if (!def.catchall) {
    if (ctx.io === "output")
      json.additionalProperties = false;
  } else if (def.catchall) {
    json.additionalProperties = processSchema(def.catchall, ctx, {
      ...params,
      path: [...params.path, "additionalProperties"]
    });
  }
};
var unionProcessor = (schema, ctx, json, params) => {
  const def = schema._zod.def;
  const isExclusive = def.inclusive === false;
  const options = def.options.map((x, i) => processSchema(x, ctx, {
    ...params,
    path: [...params.path, isExclusive ? "oneOf" : "anyOf", i]
  }));
  if (isExclusive) {
    json.oneOf = options;
  } else {
    json.anyOf = options;
  }
};
var intersectionProcessor = (schema, ctx, json, params) => {
  const def = schema._zod.def;
  const a = processSchema(def.left, ctx, {
    ...params,
    path: [...params.path, "allOf", 0]
  });
  const b = processSchema(def.right, ctx, {
    ...params,
    path: [...params.path, "allOf", 1]
  });
  const isSimpleIntersection = (val) => ("allOf" in val) && Object.keys(val).length === 1;
  const allOf = [
    ...isSimpleIntersection(a) ? a.allOf : [a],
    ...isSimpleIntersection(b) ? b.allOf : [b]
  ];
  json.allOf = allOf;
  ctx.intersections.push(allOf);
};
var tupleProcessor = (schema, ctx, _json, params) => {
  const json = _json;
  const def = schema._zod.def;
  json.type = "array";
  const prefixPath = ctx.target === "draft-2020-12" ? "prefixItems" : "items";
  const restPath = ctx.target === "draft-2020-12" ? "items" : ctx.target === "openapi-3.0" ? "items" : "additionalItems";
  const prefixItems = def.items.map((x, i) => processSchema(x, ctx, {
    ...params,
    path: [...params.path, prefixPath, i]
  }));
  const rest = def.rest ? processSchema(def.rest, ctx, {
    ...params,
    path: [...params.path, restPath, ...ctx.target === "openapi-3.0" ? [def.items.length] : []]
  }) : null;
  let minItems = def.items.length;
  while (minItems > 0) {
    const item = def.items[minItems - 1];
    const optional = ctx.io === "input" ? inputOptin(item) !== undefined : item._zod.optout === "optional";
    if (!optional)
      break;
    minItems--;
  }
  const maxItems = def.items.length;
  const isClosed = !def.rest;
  if (ctx.target === "draft-2020-12") {
    json.prefixItems = prefixItems;
    if (isClosed) {
      json.items = false;
    } else if (rest) {
      json.items = rest;
    }
    if (minItems > 0)
      json.minItems = minItems;
    if (isClosed)
      json.maxItems = maxItems;
  } else if (ctx.target === "openapi-3.0") {
    json.items = {
      anyOf: prefixItems
    };
    if (rest) {
      json.items.anyOf.push(rest);
    }
    if (minItems > 0)
      json.minItems = minItems;
    if (isClosed)
      json.maxItems = maxItems;
  } else {
    json.items = prefixItems;
    if (isClosed) {
      json.additionalItems = false;
    } else if (rest) {
      json.additionalItems = rest;
    }
    if (minItems > 0)
      json.minItems = minItems;
    if (isClosed)
      json.maxItems = maxItems;
  }
  const { minimum, maximum } = aggregateChecks(schema);
  if (typeof minimum === "number")
    json.minItems = minimum;
  if (typeof maximum === "number")
    json.maxItems = maximum;
};
function stringifyKeyNames(bySchema, json, visited) {
  if (json.$ref) {
    if (visited.has(json))
      return json;
    visited.add(json);
    const def = bySchema.get(json)?.def;
    if (!def)
      return json;
    const inlined = stringifyKeyNames(bySchema, def, visited);
    return inlined === def ? json : inlined;
  }
  for (const keyword of ["anyOf", "oneOf"]) {
    const branches = json[keyword];
    if (!Array.isArray(branches))
      continue;
    const mapped = branches.map((branch) => stringifyKeyNames(bySchema, branch, visited));
    if (mapped.some((branch, i) => branch !== branches[i]))
      json = { ...json, [keyword]: mapped };
  }
  const types = Array.isArray(json.type) ? json.type : [json.type];
  const numericType = !types.includes("string") && types.some((t) => t === "number" || t === "integer");
  const values = json.enum ?? (json.const !== undefined ? [json.const] : undefined);
  if (!numericType && !values?.some((v) => typeof v === "number"))
    return json;
  const { minimum, maximum, exclusiveMinimum, exclusiveMaximum, multipleOf, format, id, ...rest } = json;
  if (rest.enum)
    rest.enum = rest.enum.map((v) => typeof v === "number" ? String(v) : v);
  else if (typeof rest.const === "number")
    rest.const = String(rest.const);
  if (!numericType)
    return rest;
  rest.type = "string";
  if (!values)
    rest.pattern = (types.includes("number") ? number : integer).source;
  return rest;
}
var pendingRecords = new WeakMap;
function rewriteKeyNames(ctx) {
  const bySchema = new Map;
  for (const entry of ctx.seen.values()) {
    if (entry.def && !bySchema.has(entry.schema))
      bySchema.set(entry.schema, entry);
  }
  const rewrites = new Map;
  for (const record of pendingRecords.get(ctx) ?? []) {
    const seen = ctx.seen.get(record);
    const names = (seen?.def ?? seen?.schema)?.propertyNames;
    if (!names || names === true || rewrites.has(names))
      continue;
    const rewritten = stringifyKeyNames(bySchema, names, new Set);
    if (rewritten !== names)
      rewrites.set(names, rewritten);
  }
  if (!rewrites.size)
    return;
  for (const entry of ctx.seen.values()) {
    for (const carrier of [entry.schema, entry.def]) {
      const rewritten = carrier && rewrites.get(carrier.propertyNames);
      if (rewritten)
        carrier.propertyNames = rewritten;
    }
  }
}
var recordProcessor = (schema, ctx, _json, params) => {
  const json = _json;
  const def = schema._zod.def;
  json.type = "object";
  const keyType = def.keyType;
  const patterns = aggregateChecks(keyType).patterns;
  if (def.mode === "loose" && patterns && patterns.size > 0) {
    const valueSchema = processSchema(def.valueType, ctx, {
      ...params,
      path: [...params.path, "patternProperties", "*"]
    });
    json.patternProperties = {};
    for (const pattern of patterns) {
      assignProp(json.patternProperties, exactPattern(pattern).source, valueSchema);
    }
  } else {
    if (ctx.target === "draft-07" || ctx.target === "draft-2020-12") {
      json.propertyNames = processSchema(def.keyType, ctx, {
        ...params,
        path: [...params.path, "propertyNames"]
      });
      let pending = pendingRecords.get(ctx);
      if (!pending) {
        pending = [];
        pendingRecords.set(ctx, pending);
        ctx.deferred.push(() => rewriteKeyNames(ctx));
      }
      pending.push(schema);
    }
    json.additionalProperties = processSchema(def.valueType, ctx, {
      ...params,
      path: [...params.path, "additionalProperties"]
    });
  }
  const keyValues = keyType._zod.values;
  const omittableOnInput = ctx.io === "input" && inputOptin(def.valueType) !== undefined;
  if (keyValues && !def.partial && !omittableOnInput) {
    const validKeyValues = [...keyValues].filter((v) => typeof v === "string" || typeof v === "number");
    if (validKeyValues.length > 0) {
      json.required = validKeyValues.map(String);
    }
  }
};
var nullableProcessor = (schema, ctx, json, params) => {
  const def = schema._zod.def;
  const inner = processSchema(def.innerType, ctx, params);
  const seen = ctx.seen.get(schema);
  if (ctx.target === "openapi-3.0") {
    seen.ref = def.innerType;
    json.nullable = true;
  } else {
    json.anyOf = [inner, { type: "null" }];
  }
};
var nonoptionalProcessor = (schema, ctx, _json, params) => {
  const def = schema._zod.def;
  processSchema(def.innerType, ctx, params);
  const seen = ctx.seen.get(schema);
  seen.ref = def.innerType;
};
var UNREPRESENTABLE_DEFAULT = Symbol();
function serializeDefaultValue(value, schema, ctx, json, params) {
  let unrepresentable = false;
  const serialized = JSON.stringify(value, (_, val) => {
    if (typeof val !== "bigint")
      return val;
    unrepresentable = true;
    return null;
  });
  if (!unrepresentable)
    return JSON.parse(serialized);
  handleUnrepresentable(schema, ctx, json, params, "BigInt defaults cannot be represented in JSON Schema");
  return UNREPRESENTABLE_DEFAULT;
}
var defaultProcessor = (schema, ctx, json, params) => {
  const def = schema._zod.def;
  processSchema(def.innerType, ctx, params);
  const seen = ctx.seen.get(schema);
  seen.ref = def.innerType;
  const value = serializeDefaultValue(def.defaultValue, schema, ctx, json, params);
  if (value !== UNREPRESENTABLE_DEFAULT)
    json.default = value;
};
var prefaultProcessor = (schema, ctx, json, params) => {
  const def = schema._zod.def;
  processSchema(def.innerType, ctx, params);
  const seen = ctx.seen.get(schema);
  seen.ref = def.innerType;
  if (ctx.io !== "input")
    return;
  const value = serializeDefaultValue(def.defaultValue, schema, ctx, json, params);
  if (value !== UNREPRESENTABLE_DEFAULT)
    json._prefault = value;
};
var catchProcessor = (schema, ctx, json, params) => {
  const def = schema._zod.def;
  processSchema(def.innerType, ctx, params);
  const seen = ctx.seen.get(schema);
  seen.ref = def.innerType;
  let catchValue;
  try {
    catchValue = def.catchValue(undefined);
  } catch {
    handleUnrepresentable(schema, ctx, json, params, "Dynamic catch values are not supported in JSON Schema");
    return;
  }
  json.default = catchValue;
};
var pipeProcessor = (schema, ctx, _json, params) => {
  const def = schema._zod.def;
  const inIsTransform = def.in._zod.traits.has("$ZodTransform");
  const innerType = ctx.io === "input" ? inIsTransform ? def.out : def.in : def.out;
  processSchema(innerType, ctx, params);
  const seen = ctx.seen.get(schema);
  seen.ref = innerType;
};
var readonlyProcessor = (schema, ctx, json, params) => {
  const def = schema._zod.def;
  processSchema(def.innerType, ctx, params);
  const seen = ctx.seen.get(schema);
  seen.ref = def.innerType;
  json.readOnly = true;
};
var promiseProcessor = (schema, ctx, _json, params) => {
  const def = schema._zod.def;
  processSchema(def.innerType, ctx, params);
  const seen = ctx.seen.get(schema);
  seen.ref = def.innerType;
};
var optionalProcessor = (schema, ctx, _json, params) => {
  const def = schema._zod.def;
  processSchema(def.innerType, ctx, params);
  const seen = ctx.seen.get(schema);
  seen.ref = def.innerType;
};
var lazyProcessor = (schema, ctx, _json, params) => {
  const innerType = schema._zod.innerType;
  processSchema(innerType, ctx, params);
  const seen = ctx.seen.get(schema);
  seen.ref = innerType;
};
// node_modules/zod/v4/classic/errors.js
var _installedErrorProtos = /* @__PURE__ */ new WeakSet([Object.prototype, Error.prototype]);
function _lazyMethod(proto, key, make) {
  Object.defineProperty(proto, key, {
    configurable: true,
    enumerable: false,
    get() {
      const value = make(this);
      Object.defineProperty(this, key, { value, configurable: true, writable: true });
      return value;
    },
    set(value) {
      Object.defineProperty(this, key, { value, configurable: true, writable: true });
    }
  });
}
var initializer2 = (inst, issues) => {
  $ZodError.init(inst, issues);
  inst.name = "ZodError";
  const proto = Object.getPrototypeOf(inst);
  if (_installedErrorProtos.has(proto))
    return;
  _installedErrorProtos.add(proto);
  _lazyMethod(proto, "format", (self) => (mapper) => formatError(self, mapper));
  _lazyMethod(proto, "flatten", (self) => (mapper) => flattenError(self, mapper));
  _lazyMethod(proto, "addIssue", (self) => (issue) => {
    self.issues.push(issue);
    self.message = JSON.stringify(self.issues, jsonStringifyReplacer, 2);
  });
  _lazyMethod(proto, "addIssues", (self) => (issues) => {
    self.issues.push(...issues);
    self.message = JSON.stringify(self.issues, jsonStringifyReplacer, 2);
  });
  Object.defineProperty(proto, "isEmpty", {
    configurable: true,
    enumerable: false,
    get() {
      return this.issues.length === 0;
    }
  });
};
var ZodRealError = /* @__PURE__ */ $constructor("ZodError", initializer2, undefined, {
  Parent: Error
});

// node_modules/zod/v4/classic/parse.js
var parse3 = /* @__PURE__ */ _parse(ZodRealError);
var parseAsync2 = /* @__PURE__ */ _parseAsync(ZodRealError);
var safeParse = /* @__PURE__ */ _safeParse(ZodRealError);
var safeParseAsync = /* @__PURE__ */ _safeParseAsync(ZodRealError);
var encode = /* @__PURE__ */ _encode(ZodRealError);
var decode = /* @__PURE__ */ _decode(ZodRealError);
var encodeAsync = /* @__PURE__ */ _encodeAsync(ZodRealError);
var decodeAsync = /* @__PURE__ */ _decodeAsync(ZodRealError);
var safeEncode = /* @__PURE__ */ _safeEncode(ZodRealError);
var safeDecode = /* @__PURE__ */ _safeDecode(ZodRealError);
var safeEncodeAsync = /* @__PURE__ */ _safeEncodeAsync(ZodRealError);
var safeDecodeAsync = /* @__PURE__ */ _safeDecodeAsync(ZodRealError);

// node_modules/zod/v4/classic/schemas.js
function _ensureDefaultLocale() {
  if (!globalConfig.localeError)
    config(en_default());
}
function _ensureDefaultMemoizer() {
  if (!globalConfig.memoizer)
    config({ memoizer: memoizer() });
}
var ZodType = /* @__PURE__ */ $constructor("ZodType", (inst, def) => {
  _ensureDefaultLocale();
  $ZodType.init(inst, def);
  inst.def = def;
  inst.type = def.type;
  return inst;
}, {
  check(...chks) {
    const def = this.def;
    return this.clone(mergeDefs(def, {
      checks: [
        ...def.checks ?? [],
        ...chks.map((ch) => typeof ch === "function" ? { _zod: { check: ch, def: { check: "custom" }, onattach: [] } } : ch)
      ]
    }), { parent: true });
  },
  with(...chks) {
    return this.check(...chks);
  },
  clone(def, params) {
    return clone(this, def, params);
  },
  brand() {
    return this;
  },
  register(reg, meta) {
    reg.add(this, meta);
    return this;
  },
  refine(check2, params) {
    return this.check(refine(check2, params));
  },
  superRefine(refinement, params) {
    return this.check(superRefine(refinement, params));
  },
  overwrite(fn) {
    return this.check(_overwrite(fn));
  },
  optional() {
    return optional(this);
  },
  exactOptional() {
    return exactOptional(this);
  },
  nullable() {
    return nullable(this);
  },
  nullish() {
    return optional(nullable(this));
  },
  nonoptional(params) {
    return nonoptional(this, params);
  },
  array() {
    return array(this);
  },
  or(arg) {
    return union([this, arg]);
  },
  and(arg) {
    return intersection(this, arg);
  },
  transform(tx) {
    return pipe(this, transform(tx));
  },
  default(d) {
    return _default(this, d);
  },
  prefault(d) {
    return prefault(this, d);
  },
  catch(params) {
    return _catch(this, params);
  },
  pipe(target) {
    return pipe(this, target);
  },
  readonly() {
    return readonly(this);
  },
  describe(description) {
    const cl = this.clone();
    globalRegistry.add(cl, { description });
    return cl;
  },
  meta(...args) {
    if (args.length === 0)
      return globalRegistry.get(this);
    const cl = this.clone();
    globalRegistry.add(cl, args[0]);
    return cl;
  },
  isOptional() {
    return this.safeParse(undefined).success;
  },
  isNullable() {
    return this.safeParse(null).success;
  },
  apply(fn, ...args) {
    return args.length === 0 ? fn(this) : fn(this, ...args);
  },
  get "~standard"() {
    return hide(this, "~standard", {
      ...standardProps(this),
      jsonSchema: {
        input: createStandardJSONSchemaMethod(this, "input"),
        output: createStandardJSONSchemaMethod(this, "output")
      }
    });
  },
  set "~standard"(value) {
    own(this, "~standard", value);
  },
  parse: function _parse(data, params) {
    return parse3(this, data, params, { callee: _parse });
  },
  parseAsync: async function _parseAsync(data, params) {
    return await parseAsync2(this, data, params, { callee: _parseAsync });
  },
  safeParse(data, params) {
    return safeParse(this, data, params);
  },
  async safeParseAsync(data, params) {
    return safeParseAsync(this, data, params);
  },
  get spa() {
    return this?.safeParseAsync;
  },
  set spa(value) {
    own(this, "spa", value);
  },
  validate(data, params) {
    return validate(this, data, params);
  },
  validateAsync(data, params) {
    return validateAsync(this, data, params);
  },
  encode: function _encode(data, params) {
    return encode(this, data, params, { callee: _encode });
  },
  decode: function _decode(data, params) {
    return decode(this, data, params, { callee: _decode });
  },
  encodeAsync: async function _encodeAsync(data, params) {
    return await encodeAsync(this, data, params, { callee: _encodeAsync });
  },
  decodeAsync: async function _decodeAsync(data, params) {
    return await decodeAsync(this, data, params, { callee: _decodeAsync });
  },
  safeEncode(data, params) {
    return safeEncode(this, data, params);
  },
  safeDecode(data, params) {
    return safeDecode(this, data, params);
  },
  async safeEncodeAsync(data, params) {
    return safeEncodeAsync(this, data, params);
  },
  async safeDecodeAsync(data, params) {
    return safeDecodeAsync(this, data, params);
  },
  toJSONSchema(params) {
    return createToJSONSchemaMethod(this, {})(params);
  },
  get description() {
    return globalRegistry.get(this)?.description;
  },
  get _def() {
    return this._zod.def;
  }
});
var _ZodString = /* @__PURE__ */ $constructor("_ZodString", (inst, def) => {
  $ZodString.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => stringProcessor(inst, ctx, json2, params);
}, /* @__PURE__ */ derived({
  format: (inst) => aggregateChecks(inst).format ?? null,
  minLength: (inst) => aggregateChecks(inst).minimum ?? null,
  maxLength: (inst) => aggregateChecks(inst).maximum ?? null
}, {
  regex(...args) {
    return this.check(_regex(...args));
  },
  includes(...args) {
    return this.check(_includes(...args));
  },
  startsWith(...args) {
    return this.check(_startsWith(...args));
  },
  endsWith(...args) {
    return this.check(_endsWith(...args));
  },
  min(...args) {
    return this.check(_minLength(...args));
  },
  max(...args) {
    return this.check(_maxLength(...args));
  },
  length(...args) {
    return this.check(_length(...args));
  },
  nonempty(...args) {
    return this.check(_minLength(1, ...args));
  },
  lowercase(params) {
    return this.check(_lowercase(params));
  },
  uppercase(params) {
    return this.check(_uppercase(params));
  },
  trim() {
    return this.check(_trim());
  },
  normalize(...args) {
    return this.check(_normalize(...args));
  },
  toLowerCase() {
    return this.check(_toLowerCase());
  },
  toUpperCase() {
    return this.check(_toUpperCase());
  },
  slugify() {
    return this.check(_slugify());
  }
}));
var ZodString = /* @__PURE__ */ $constructor("ZodString", (inst, def) => {
  $ZodString.init(inst, def);
  _ZodString.init(inst, def);
}, {
  email(params) {
    return this.check(_email(ZodEmail, params));
  },
  url(params) {
    return this.check(_url(ZodURL, params));
  },
  jwt(params) {
    return this.check(_jwt(ZodJWT, params));
  },
  emoji(params) {
    return this.check(_emoji2(ZodEmoji, params));
  },
  guid(params) {
    return this.check(_guid(ZodGUID, params));
  },
  uuid(params) {
    return this.check(_uuid(ZodUUID, params));
  },
  uuidv4(params) {
    return this.check(_uuidv4(ZodUUID, params));
  },
  uuidv6(params) {
    return this.check(_uuidv6(ZodUUID, params));
  },
  uuidv7(params) {
    return this.check(_uuidv7(ZodUUID, params));
  },
  nanoid(params) {
    return this.check(_nanoid(ZodNanoID, params));
  },
  cuid(params) {
    return this.check(_cuid(ZodCUID, params));
  },
  cuid2(params) {
    return this.check(_cuid2(ZodCUID2, params));
  },
  ulid(params) {
    return this.check(_ulid(ZodULID, params));
  },
  base64(params) {
    return this.check(_base64(ZodBase64, params));
  },
  base64url(params) {
    return this.check(_base64url(ZodBase64URL, params));
  },
  xid(params) {
    return this.check(_xid(ZodXID, params));
  },
  ksuid(params) {
    return this.check(_ksuid(ZodKSUID, params));
  },
  ipv4(params) {
    return this.check(_ipv4(ZodIPv4, params));
  },
  ipv6(params) {
    return this.check(_ipv6(ZodIPv6, params));
  },
  cidrv4(params) {
    return this.check(_cidrv4(ZodCIDRv4, params));
  },
  cidrv6(params) {
    return this.check(_cidrv6(ZodCIDRv6, params));
  },
  e164(params) {
    return this.check(_e164(ZodE164, params));
  },
  datetime(params) {
    return this.check(_isoDateTime(ZodISODateTime, params));
  },
  date(params) {
    return this.check(_isoDate(ZodISODate, params));
  },
  time(params) {
    return this.check(_isoTime(ZodISOTime, params));
  },
  duration(params) {
    return this.check(_isoDuration(ZodISODuration, params));
  }
});
function string2(params) {
  return _string(ZodString, params);
}
var ZodStringFormat = /* @__PURE__ */ $constructor("ZodStringFormat", (inst, def) => {
  $ZodStringFormat.init(inst, def);
  _ZodString.init(inst, def);
});
var ZodISODateTime = /* @__PURE__ */ $constructor("ZodISODateTime", (inst, def) => {
  $ZodISODateTime.init(inst, def);
  ZodStringFormat.init(inst, def);
});
var ZodISODate = /* @__PURE__ */ $constructor("ZodISODate", (inst, def) => {
  $ZodISODate.init(inst, def);
  ZodStringFormat.init(inst, def);
});
var ZodISOTime = /* @__PURE__ */ $constructor("ZodISOTime", (inst, def) => {
  $ZodISOTime.init(inst, def);
  ZodStringFormat.init(inst, def);
});
var ZodISODuration = /* @__PURE__ */ $constructor("ZodISODuration", (inst, def) => {
  $ZodISODuration.init(inst, def);
  ZodStringFormat.init(inst, def);
});
var ZodEmail = /* @__PURE__ */ $constructor("ZodEmail", (inst, def) => {
  $ZodEmail.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function email2(params) {
  return _email(ZodEmail, params);
}
var ZodGUID = /* @__PURE__ */ $constructor("ZodGUID", (inst, def) => {
  $ZodGUID.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function guid2(params) {
  return _guid(ZodGUID, params);
}
var ZodUUID = /* @__PURE__ */ $constructor("ZodUUID", (inst, def) => {
  $ZodUUID.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function uuid2(params) {
  return _uuid(ZodUUID, params);
}
function uuidv4(params) {
  return _uuidv4(ZodUUID, params);
}
function uuidv6(params) {
  return _uuidv6(ZodUUID, params);
}
function uuidv7(params) {
  return _uuidv7(ZodUUID, params);
}
var ZodURL = /* @__PURE__ */ $constructor("ZodURL", (inst, def) => {
  $ZodURL.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function url(params) {
  return _url(ZodURL, params);
}
function httpUrl(params) {
  return _url(ZodURL, {
    protocol: httpProtocol,
    hostname: domain,
    ...normalizeParams(params)
  });
}
var ZodEmoji = /* @__PURE__ */ $constructor("ZodEmoji", (inst, def) => {
  $ZodEmoji.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function emoji2(params) {
  return _emoji2(ZodEmoji, params);
}
var ZodNanoID = /* @__PURE__ */ $constructor("ZodNanoID", (inst, def) => {
  $ZodNanoID.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function nanoid2(params) {
  return _nanoid(ZodNanoID, params);
}
var ZodCUID = /* @__PURE__ */ $constructor("ZodCUID", (inst, def) => {
  $ZodCUID.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function cuid3(params) {
  return _cuid(ZodCUID, params);
}
var ZodCUID2 = /* @__PURE__ */ $constructor("ZodCUID2", (inst, def) => {
  $ZodCUID2.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function cuid22(params) {
  return _cuid2(ZodCUID2, params);
}
var ZodULID = /* @__PURE__ */ $constructor("ZodULID", (inst, def) => {
  $ZodULID.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function ulid2(params) {
  return _ulid(ZodULID, params);
}
var ZodXID = /* @__PURE__ */ $constructor("ZodXID", (inst, def) => {
  $ZodXID.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function xid2(params) {
  return _xid(ZodXID, params);
}
var ZodKSUID = /* @__PURE__ */ $constructor("ZodKSUID", (inst, def) => {
  $ZodKSUID.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function ksuid2(params) {
  return _ksuid(ZodKSUID, params);
}
var ZodIPv4 = /* @__PURE__ */ $constructor("ZodIPv4", (inst, def) => {
  $ZodIPv4.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function ipv42(params) {
  return _ipv4(ZodIPv4, params);
}
var ZodMAC = /* @__PURE__ */ $constructor("ZodMAC", (inst, def) => {
  $ZodMAC.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function mac2(params) {
  return _mac(ZodMAC, params);
}
var ZodIPv6 = /* @__PURE__ */ $constructor("ZodIPv6", (inst, def) => {
  $ZodIPv6.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function ipv62(params) {
  return _ipv6(ZodIPv6, params);
}
var ZodCIDRv4 = /* @__PURE__ */ $constructor("ZodCIDRv4", (inst, def) => {
  $ZodCIDRv4.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function cidrv42(params) {
  return _cidrv4(ZodCIDRv4, params);
}
var ZodCIDRv6 = /* @__PURE__ */ $constructor("ZodCIDRv6", (inst, def) => {
  $ZodCIDRv6.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function cidrv62(params) {
  return _cidrv6(ZodCIDRv6, params);
}
var ZodBase64 = /* @__PURE__ */ $constructor("ZodBase64", (inst, def) => {
  $ZodBase64.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function base642(params) {
  return _base64(ZodBase64, params);
}
var ZodBase64URL = /* @__PURE__ */ $constructor("ZodBase64URL", (inst, def) => {
  $ZodBase64URL.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function base64url2(params) {
  return _base64url(ZodBase64URL, params);
}
var ZodE164 = /* @__PURE__ */ $constructor("ZodE164", (inst, def) => {
  $ZodE164.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function e1642(params) {
  return _e164(ZodE164, params);
}
var ZodCreditCard = /* @__PURE__ */ $constructor("ZodCreditCard", (inst, def) => {
  $ZodCreditCard.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function creditCard2(params) {
  return _creditCard(ZodCreditCard, params);
}
var ZodIBAN = /* @__PURE__ */ $constructor("ZodIBAN", (inst, def) => {
  $ZodIBAN.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function iban2(params) {
  return _iban(ZodIBAN, params);
}
var ZodJWT = /* @__PURE__ */ $constructor("ZodJWT", (inst, def) => {
  $ZodJWT.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function jwt(params) {
  return _jwt(ZodJWT, params);
}
var ZodCustomStringFormat = /* @__PURE__ */ $constructor("ZodCustomStringFormat", (inst, def) => {
  $ZodCustomStringFormat.init(inst, def);
  ZodStringFormat.init(inst, def);
});
function stringFormat(format, fnOrRegex, _params = {}) {
  return _stringFormat(ZodCustomStringFormat, format, fnOrRegex, _params);
}
function hostname2(_params) {
  return _stringFormat(ZodCustomStringFormat, "hostname", hostname, _params);
}
function hex2(_params) {
  return _stringFormat(ZodCustomStringFormat, "hex", hex, _params);
}
function currencyCode2(_params) {
  return _stringFormat(ZodCustomStringFormat, "currency_code", currencyCode, _params);
}
function hash(alg, params) {
  const enc = params?.enc ?? "hex";
  const format = `${alg}_${enc}`;
  const regex = exports_regexes[format];
  if (!regex)
    throw new Error(`Unrecognized hash format: ${format}`);
  return _stringFormat(ZodCustomStringFormat, format, regex, params);
}
var ZodNumber = /* @__PURE__ */ $constructor("ZodNumber", (inst, def) => {
  $ZodNumber.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => numberProcessor(inst, ctx, json2, params);
  inst.isFinite = true;
}, /* @__PURE__ */ derived({
  minValue: (inst) => {
    const { minimum, exclusiveMinimum } = aggregateChecks(inst);
    return Math.max(minimum ?? Number.NEGATIVE_INFINITY, exclusiveMinimum ?? Number.NEGATIVE_INFINITY);
  },
  maxValue: (inst) => {
    const { maximum, exclusiveMaximum } = aggregateChecks(inst);
    return Math.min(maximum ?? Number.POSITIVE_INFINITY, exclusiveMaximum ?? Number.POSITIVE_INFINITY);
  },
  isInt: (inst) => {
    const { isInt, multipleOf } = aggregateChecks(inst);
    return !!isInt || !!multipleOf?.some(Number.isSafeInteger);
  },
  format: (inst) => aggregateChecks(inst).format ?? null
}, {
  gt(value, params) {
    return this.check(_gt(value, params));
  },
  gte(value, params) {
    return this.check(_gte(value, params));
  },
  min(value, params) {
    return this.check(_gte(value, params));
  },
  lt(value, params) {
    return this.check(_lt(value, params));
  },
  lte(value, params) {
    return this.check(_lte(value, params));
  },
  max(value, params) {
    return this.check(_lte(value, params));
  },
  int(params) {
    return this.check(int(params));
  },
  safe(params) {
    return this.check(int(params));
  },
  positive(params) {
    return this.check(_gt(0, params));
  },
  nonnegative(params) {
    return this.check(_gte(0, params));
  },
  negative(params) {
    return this.check(_lt(0, params));
  },
  nonpositive(params) {
    return this.check(_lte(0, params));
  },
  multipleOf(value, params) {
    return this.check(_multipleOf(value, params));
  },
  step(value, params) {
    return this.check(_multipleOf(value, params));
  },
  finite() {
    return this;
  }
}));
function number2(params) {
  return _number(ZodNumber, params);
}
var ZodNumberFormat = /* @__PURE__ */ $constructor("ZodNumberFormat", (inst, def) => {
  $ZodNumberFormat.init(inst, def);
  ZodNumber.init(inst, def);
});
function int(params) {
  return _int(ZodNumberFormat, params);
}
function float32(params) {
  return _float32(ZodNumberFormat, params);
}
function float64(params) {
  return _float64(ZodNumberFormat, params);
}
function int32(params) {
  return _int32(ZodNumberFormat, params);
}
function uint32(params) {
  return _uint32(ZodNumberFormat, params);
}
var ZodBoolean = /* @__PURE__ */ $constructor("ZodBoolean", (inst, def) => {
  $ZodBoolean.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => booleanProcessor(inst, ctx, json2, params);
});
function boolean2(params) {
  return _boolean(ZodBoolean, params);
}
var ZodBigInt = /* @__PURE__ */ $constructor("ZodBigInt", (inst, def) => {
  $ZodBigInt.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => bigintProcessor(inst, ctx, json2, params);
}, /* @__PURE__ */ derived({
  minValue: (inst) => aggregateChecks(inst).minimum ?? null,
  maxValue: (inst) => aggregateChecks(inst).maximum ?? null,
  format: (inst) => aggregateChecks(inst).format ?? null
}, {
  gte(value, params) {
    return this.check(_gte(value, params));
  },
  min(value, params) {
    return this.check(_gte(value, params));
  },
  gt(value, params) {
    return this.check(_gt(value, params));
  },
  lt(value, params) {
    return this.check(_lt(value, params));
  },
  lte(value, params) {
    return this.check(_lte(value, params));
  },
  max(value, params) {
    return this.check(_lte(value, params));
  },
  positive(params) {
    return this.check(_gt(BigInt(0), params));
  },
  negative(params) {
    return this.check(_lt(BigInt(0), params));
  },
  nonpositive(params) {
    return this.check(_lte(BigInt(0), params));
  },
  nonnegative(params) {
    return this.check(_gte(BigInt(0), params));
  },
  multipleOf(value, params) {
    return this.check(_multipleOf(value, params));
  }
}));
function bigint2(params) {
  return _bigint(ZodBigInt, params);
}
var ZodBigIntFormat = /* @__PURE__ */ $constructor("ZodBigIntFormat", (inst, def) => {
  $ZodBigIntFormat.init(inst, def);
  ZodBigInt.init(inst, def);
});
function int64(params) {
  return _int64(ZodBigIntFormat, params);
}
function uint64(params) {
  return _uint64(ZodBigIntFormat, params);
}
var ZodSymbol = /* @__PURE__ */ $constructor("ZodSymbol", (inst, def) => {
  $ZodSymbol.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => symbolProcessor(inst, ctx, json2, params);
});
function symbol(params) {
  return _symbol(ZodSymbol, params);
}
var ZodUndefined = /* @__PURE__ */ $constructor("ZodUndefined", (inst, def) => {
  $ZodUndefined.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => undefinedProcessor(inst, ctx, json2, params);
});
function _undefined3(params) {
  return _undefined2(ZodUndefined, params);
}
var ZodNull = /* @__PURE__ */ $constructor("ZodNull", (inst, def) => {
  $ZodNull.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => nullProcessor(inst, ctx, json2, params);
});
function _null3(params) {
  return _null2(ZodNull, params);
}
var ZodAny = /* @__PURE__ */ $constructor("ZodAny", (inst, def) => {
  $ZodAny.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => anyProcessor(inst, ctx, json2, params);
});
function any() {
  return _any(ZodAny);
}
var ZodUnknown = /* @__PURE__ */ $constructor("ZodUnknown", (inst, def) => {
  $ZodUnknown.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => unknownProcessor(inst, ctx, json2, params);
});
function unknown() {
  return _unknown(ZodUnknown);
}
var ZodNever = /* @__PURE__ */ $constructor("ZodNever", (inst, def) => {
  $ZodNever.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => neverProcessor(inst, ctx, json2, params);
});
function never(params) {
  return _never(ZodNever, params);
}
var ZodVoid = /* @__PURE__ */ $constructor("ZodVoid", (inst, def) => {
  $ZodVoid.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => voidProcessor(inst, ctx, json2, params);
});
function _void2(params) {
  return _void(ZodVoid, params);
}
var ZodDate = /* @__PURE__ */ $constructor("ZodDate", (inst, def) => {
  $ZodDate.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => dateProcessor(inst, ctx, json2, params);
  inst.min = (value, params) => inst.check(_gte(value, params));
  inst.max = (value, params) => inst.check(_lte(value, params));
}, /* @__PURE__ */ derived({
  minDate: (inst) => {
    const { minimum } = aggregateChecks(inst);
    return minimum ? new Date(minimum) : null;
  },
  maxDate: (inst) => {
    const { maximum } = aggregateChecks(inst);
    return maximum ? new Date(maximum) : null;
  }
}, {}));
function date2(params) {
  return _date(ZodDate, params);
}
var ZodArray = /* @__PURE__ */ $constructor("ZodArray", (inst, def) => {
  _ensureDefaultMemoizer();
  $ZodArray.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => arrayProcessor(inst, ctx, json2, params);
  inst.element = def.element;
}, {
  min(n, params) {
    return this.check(_minLength(n, params));
  },
  nonempty(params) {
    return this.check(_minLength(1, params));
  },
  max(n, params) {
    return this.check(_maxLength(n, params));
  },
  length(n, params) {
    return this.check(_length(n, params));
  },
  unwrap() {
    return this.element;
  }
});
function array(element, params) {
  return _array(ZodArray, element, params);
}
function keyof(schema) {
  const shape = schema._zod.def.shape;
  return _enum(Object.keys(shape));
}
var ZodObject = /* @__PURE__ */ $constructor("ZodObject", (inst, def) => {
  _ensureDefaultMemoizer();
  $ZodObjectJIT.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => objectProcessor(inst, ctx, json2, params);
  installLazyProp(inst, "shape", (self) => self._zod.def.shape, false);
}, {
  keyof() {
    return _enum(Object.keys(this._zod.def.shape));
  },
  catchall(catchall) {
    return this.clone(mergeDefs(this._zod.def, { catchall }));
  },
  passthrough() {
    return this.clone(mergeDefs(this._zod.def, { catchall: unknown() }));
  },
  loose() {
    return this.clone(mergeDefs(this._zod.def, { catchall: unknown() }));
  },
  strict() {
    return this.clone(mergeDefs(this._zod.def, { catchall: never() }));
  },
  strip() {
    return this.clone(mergeDefs(this._zod.def, { catchall: undefined }));
  },
  extend(incoming) {
    return extend(this, incoming);
  },
  safeExtend(incoming) {
    return safeExtend(this, incoming);
  },
  merge(other) {
    return merge(this, other);
  },
  pick(mask) {
    return pick(this, mask);
  },
  omit(mask) {
    return omit(this, mask);
  },
  partial(...args) {
    return partial(ZodOptional, this, args[0]);
  },
  exactPartial(...args) {
    return partial(ZodExactOptional, this, args[0], "exactPartial");
  },
  required(...args) {
    return required(ZodNonOptional, this, args[0]);
  }
});
function object(shape, params) {
  const def = {
    type: "object",
    shape: shape ?? {},
    ...normalizeParams(params)
  };
  return new ZodObject(def);
}
function strictObject(shape, params) {
  return new ZodObject({
    type: "object",
    shape,
    catchall: never(),
    ...normalizeParams(params)
  });
}
function looseObject(shape, params) {
  return new ZodObject({
    type: "object",
    shape,
    catchall: unknown(),
    ...normalizeParams(params)
  });
}
var ZodUnion = /* @__PURE__ */ $constructor("ZodUnion", (inst, def) => {
  $ZodUnion.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => unionProcessor(inst, ctx, json2, params);
  inst.options = def.options;
});
function union(options, params) {
  return new ZodUnion({
    type: "union",
    options,
    ...normalizeParams(params)
  });
}
var ZodXor = /* @__PURE__ */ $constructor("ZodXor", (inst, def) => {
  ZodUnion.init(inst, def);
  $ZodXor.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => unionProcessor(inst, ctx, json2, params);
  inst.options = def.options;
});
function xor(options, params) {
  return new ZodXor({
    type: "union",
    options,
    inclusive: false,
    ...normalizeParams(params)
  });
}
var ZodDiscriminatedUnion = /* @__PURE__ */ $constructor("ZodDiscriminatedUnion", (inst, def) => {
  ZodUnion.init(inst, def);
  $ZodDiscriminatedUnion.init(inst, def);
});
function discriminatedUnion(discriminator, options, params) {
  return new ZodDiscriminatedUnion({
    type: "union",
    options,
    discriminator,
    ...normalizeParams(params)
  });
}
var ZodIntersection = /* @__PURE__ */ $constructor("ZodIntersection", (inst, def) => {
  $ZodIntersection.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => intersectionProcessor(inst, ctx, json2, params);
});
function intersection(left, right) {
  return new ZodIntersection({
    type: "intersection",
    left,
    right
  });
}
var ZodTuple = /* @__PURE__ */ $constructor("ZodTuple", (inst, def) => {
  _ensureDefaultMemoizer();
  $ZodTuple.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => tupleProcessor(inst, ctx, json2, params);
}, {
  rest(rest) {
    return this.clone({
      ...this._zod.def,
      rest
    });
  },
  partial() {
    const def = this._zod.def;
    if (def.checks?.length)
      throw new Error(".partial() cannot be used on tuple schemas containing refinements");
    return this.clone({
      ...def,
      items: def.items.map((item) => new ZodOptional({ type: "optional", innerType: item }))
    });
  }
});
function tuple(items, _paramsOrRest, _params) {
  const hasRest = _paramsOrRest instanceof $ZodType;
  const params = hasRest ? _params : _paramsOrRest;
  const rest = hasRest ? _paramsOrRest : null;
  return new ZodTuple({
    type: "tuple",
    items,
    rest,
    ...normalizeParams(params)
  });
}
var ZodRecord = /* @__PURE__ */ $constructor("ZodRecord", (inst, def) => {
  _ensureDefaultMemoizer();
  $ZodRecord.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => recordProcessor(inst, ctx, json2, params);
  inst.keyType = def.keyType;
  inst.valueType = def.valueType;
});
function record(keyType, valueType, params) {
  if (!valueType || !valueType._zod) {
    return new ZodRecord({
      type: "record",
      keyType: string2(),
      valueType: keyType,
      ...normalizeParams(valueType)
    });
  }
  return new ZodRecord({
    type: "record",
    keyType,
    valueType,
    ...normalizeParams(params)
  });
}
function partialRecord(keyType, valueType, params) {
  return new ZodRecord({
    type: "record",
    keyType,
    valueType,
    ...normalizeParams(params),
    partial: true
  });
}
function looseRecord(keyType, valueType, params) {
  return new ZodRecord({
    type: "record",
    keyType,
    valueType,
    mode: "loose",
    ...normalizeParams(params)
  });
}
var ZodMap = /* @__PURE__ */ $constructor("ZodMap", (inst, def) => {
  _ensureDefaultMemoizer();
  $ZodMap.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => mapProcessor(inst, ctx, json2, params);
  inst.keyType = def.keyType;
  inst.valueType = def.valueType;
  inst.min = (...args) => inst.check(_minSize(...args));
  inst.nonempty = (params) => inst.check(_minSize(1, params));
  inst.max = (...args) => inst.check(_maxSize(...args));
  inst.size = (...args) => inst.check(_size(...args));
});
function map(keyType, valueType, params) {
  return new ZodMap({
    type: "map",
    keyType,
    valueType,
    ...normalizeParams(params)
  });
}
var ZodSet = /* @__PURE__ */ $constructor("ZodSet", (inst, def) => {
  _ensureDefaultMemoizer();
  $ZodSet.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => setProcessor(inst, ctx, json2, params);
  inst.min = (...args) => inst.check(_minSize(...args));
  inst.nonempty = (params) => inst.check(_minSize(1, params));
  inst.max = (...args) => inst.check(_maxSize(...args));
  inst.size = (...args) => inst.check(_size(...args));
});
function set(valueType, params) {
  return new ZodSet({
    type: "set",
    valueType,
    ...normalizeParams(params)
  });
}
var ZodEnum = /* @__PURE__ */ $constructor("ZodEnum", (inst, def) => {
  $ZodEnum.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => enumProcessor(inst, ctx, json2, params);
  inst.enum = def.entries;
  inst.options = [...inst._zod.values];
  const keys = new Set(Object.keys(def.entries));
  inst.extract = (values, params) => {
    const newEntries = {};
    for (const value of values) {
      if (keys.has(value)) {
        newEntries[value] = def.entries[value];
      } else
        throw new Error(`Key ${value} not found in enum`);
    }
    return new ZodEnum({
      ...def,
      checks: [],
      ...normalizeParams(params),
      entries: newEntries
    });
  };
  inst.exclude = (values, params) => {
    const newEntries = { ...def.entries };
    for (const value of values) {
      if (keys.has(value)) {
        delete newEntries[value];
      } else
        throw new Error(`Key ${value} not found in enum`);
    }
    return new ZodEnum({
      ...def,
      checks: [],
      ...normalizeParams(params),
      entries: newEntries
    });
  };
});
function _enum(values, params) {
  const entries = Array.isArray(values) ? Object.fromEntries(values.map((v) => [v, v])) : values;
  return new ZodEnum({
    type: "enum",
    entries,
    ...normalizeParams(params)
  });
}
function nativeEnum(entries, params) {
  return new ZodEnum({
    type: "enum",
    entries,
    ...normalizeParams(params)
  });
}
var ZodLiteral = /* @__PURE__ */ $constructor("ZodLiteral", (inst, def) => {
  $ZodLiteral.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => literalProcessor(inst, ctx, json2, params);
  inst.values = new Set(def.values);
  Object.defineProperty(inst, "value", {
    get() {
      if (def.values.length > 1) {
        throw new Error("This schema contains multiple valid literal values. Use `.values` instead.");
      }
      return def.values[0];
    }
  });
});
function literal(value, params) {
  return new ZodLiteral({
    type: "literal",
    values: Array.isArray(value) ? value : [value],
    ...normalizeParams(params)
  });
}
var ZodFile = /* @__PURE__ */ $constructor("ZodFile", (inst, def) => {
  $ZodFile.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => fileProcessor(inst, ctx, json2, params);
  inst.min = (size, params) => inst.check(_minSize(size, params));
  inst.max = (size, params) => inst.check(_maxSize(size, params));
  inst.mime = (types, params) => inst.check(_mime(Array.isArray(types) ? types : [types], params));
});
function file(params) {
  return _file(ZodFile, params);
}
var ZodTransform = /* @__PURE__ */ $constructor("ZodTransform", (inst, def) => {
  _ensureDefaultMemoizer();
  $ZodTransform.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => transformProcessor(inst, ctx, json2, params);
  inst._zod.parse = (payload, _ctx) => {
    if (_ctx.direction === "backward") {
      throw new $ZodEncodeError(inst.constructor.name);
    }
    payload.addIssue = (issue2) => {
      if (typeof issue2 === "string") {
        payload.issues.push(issue(issue2, payload.value, def));
      } else {
        const _issue = issue2;
        if (_issue.fatal)
          _issue.continue = false;
        _issue.code ?? (_issue.code = "custom");
        if (!("input" in _issue))
          _issue.input = payload.value;
        _issue.inst ?? (_issue.inst = inst);
        payload.issues.push(issue(_issue));
      }
    };
    const output = def.transform(payload.value, payload);
    if (output instanceof Promise) {
      return output.then((output) => {
        payload.value = output;
        return payload;
      });
    }
    payload.value = output;
    return payload;
  };
});
function transform(fn) {
  return new ZodTransform({
    type: "transform",
    transform: fn
  });
}
var ZodOptional = /* @__PURE__ */ $constructor("ZodOptional", (inst, def) => {
  $ZodOptional.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => optionalProcessor(inst, ctx, json2, params);
  inst.unwrap = () => inst._zod.def.innerType;
});
function optional(innerType) {
  return new ZodOptional({
    type: "optional",
    innerType
  });
}
var ZodExactOptional = /* @__PURE__ */ $constructor("ZodExactOptional", (inst, def) => {
  $ZodExactOptional.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => optionalProcessor(inst, ctx, json2, params);
  inst.unwrap = () => inst._zod.def.innerType;
});
function exactOptional(innerType) {
  return new ZodExactOptional({
    type: "optional",
    innerType
  });
}
var ZodNullable = /* @__PURE__ */ $constructor("ZodNullable", (inst, def) => {
  $ZodNullable.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => nullableProcessor(inst, ctx, json2, params);
  inst.unwrap = () => inst._zod.def.innerType;
});
function nullable(innerType) {
  return new ZodNullable({
    type: "nullable",
    innerType
  });
}
function nullish2(innerType) {
  return optional(nullable(innerType));
}
var ZodDefault = /* @__PURE__ */ $constructor("ZodDefault", (inst, def) => {
  $ZodDefault.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => defaultProcessor(inst, ctx, json2, params);
  inst.unwrap = () => inst._zod.def.innerType;
  inst.removeDefault = inst.unwrap;
});
function _default(innerType, defaultValue) {
  return new ZodDefault({
    type: "default",
    innerType,
    get defaultValue() {
      return typeof defaultValue === "function" ? defaultValue() : shallowClone(defaultValue);
    }
  });
}
var ZodPrefault = /* @__PURE__ */ $constructor("ZodPrefault", (inst, def) => {
  $ZodPrefault.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => prefaultProcessor(inst, ctx, json2, params);
  inst.unwrap = () => inst._zod.def.innerType;
});
function prefault(innerType, defaultValue) {
  return new ZodPrefault({
    type: "prefault",
    innerType,
    get defaultValue() {
      return typeof defaultValue === "function" ? defaultValue() : shallowClone(defaultValue);
    }
  });
}
var ZodNonOptional = /* @__PURE__ */ $constructor("ZodNonOptional", (inst, def) => {
  $ZodNonOptional.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => nonoptionalProcessor(inst, ctx, json2, params);
  inst.unwrap = () => inst._zod.def.innerType;
});
function nonoptional(innerType, params) {
  return new ZodNonOptional({
    type: "nonoptional",
    innerType,
    ...normalizeParams(params)
  });
}
var ZodSuccess = /* @__PURE__ */ $constructor("ZodSuccess", (inst, def) => {
  $ZodSuccess.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => successProcessor(inst, ctx, json2, params);
  inst.unwrap = () => inst._zod.def.innerType;
});
function success(innerType) {
  return new ZodSuccess({
    type: "success",
    innerType
  });
}
var ZodCatch = /* @__PURE__ */ $constructor("ZodCatch", (inst, def) => {
  $ZodCatch.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => catchProcessor(inst, ctx, json2, params);
  inst.unwrap = () => inst._zod.def.innerType;
  inst.removeCatch = inst.unwrap;
});
function _catch(innerType, catchValue) {
  return new ZodCatch({
    type: "catch",
    innerType,
    catchValue: typeof catchValue === "function" ? catchValue : constantCatch(catchValue)
  });
}
var ZodNaN = /* @__PURE__ */ $constructor("ZodNaN", (inst, def) => {
  $ZodNaN.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => nanProcessor(inst, ctx, json2, params);
});
function nan(params) {
  return _nan(ZodNaN, params);
}
var ZodPipe = /* @__PURE__ */ $constructor("ZodPipe", (inst, def) => {
  $ZodPipe.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => pipeProcessor(inst, ctx, json2, params);
  inst.in = def.in;
  inst.out = def.out;
});
function pipe(in_, out) {
  return new ZodPipe({
    type: "pipe",
    in: in_,
    out
  });
}
var ZodCodec = /* @__PURE__ */ $constructor("ZodCodec", (inst, def) => {
  ZodPipe.init(inst, def);
  $ZodCodec.init(inst, def);
});
function codec(in_, out, params) {
  return new ZodCodec({
    type: "pipe",
    in: in_,
    out,
    transform: params.decode,
    reverseTransform: params.encode
  });
}
function invertCodec(codec2) {
  const def = codec2._zod.def;
  return new ZodCodec({
    type: "pipe",
    in: def.out,
    out: def.in,
    transform: def.reverseTransform,
    reverseTransform: def.transform
  });
}
var ZodPreprocess = /* @__PURE__ */ $constructor("ZodPreprocess", (inst, def) => {
  ZodPipe.init(inst, def);
  $ZodPreprocess.init(inst, def);
});
var ZodReadonly = /* @__PURE__ */ $constructor("ZodReadonly", (inst, def) => {
  $ZodReadonly.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => readonlyProcessor(inst, ctx, json2, params);
  inst.unwrap = () => inst._zod.def.innerType;
});
function readonly(innerType) {
  return new ZodReadonly({
    type: "readonly",
    innerType
  });
}
var ZodTemplateLiteral = /* @__PURE__ */ $constructor("ZodTemplateLiteral", (inst, def) => {
  $ZodTemplateLiteral.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => templateLiteralProcessor(inst, ctx, json2, params);
});
function templateLiteral(parts, params) {
  return new ZodTemplateLiteral({
    type: "template_literal",
    parts,
    ...normalizeParams(params)
  });
}
var ZodLazy = /* @__PURE__ */ $constructor("ZodLazy", (inst, def) => {
  $ZodLazy.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => lazyProcessor(inst, ctx, json2, params);
  inst.unwrap = () => inst._zod.def.getter();
});
function lazy(getter) {
  return new ZodLazy({
    type: "lazy",
    getter
  });
}
var ZodPromise = /* @__PURE__ */ $constructor("ZodPromise", (inst, def) => {
  $ZodPromise.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => promiseProcessor(inst, ctx, json2, params);
  inst.unwrap = () => inst._zod.def.innerType;
});
function promise(innerType) {
  return new ZodPromise({
    type: "promise",
    innerType
  });
}
var ZodFunction = /* @__PURE__ */ $constructor("ZodFunction", (inst, def) => {
  $ZodFunction.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => functionProcessor(inst, ctx, json2, params);
});
function _function(params) {
  return new ZodFunction({
    type: "function",
    input: Array.isArray(params?.input) ? tuple(params?.input) : params?.input ?? array(unknown()),
    output: params?.output ?? unknown()
  });
}
var ZodCustom = /* @__PURE__ */ $constructor("ZodCustom", (inst, def) => {
  $ZodCustom.init(inst, def);
  ZodType.init(inst, def);
  inst._zod.processJSONSchema = (ctx, json2, params) => customProcessor(inst, ctx, json2, params);
});
function check(fn) {
  const ch = new $ZodCheck({
    check: "custom"
  });
  ch._zod.check = fn;
  return ch;
}
function custom(fn, _params) {
  return _custom(ZodCustom, fn ?? (() => true), _params);
}
function refine(fn, _params = {}) {
  return _refine(ZodCustom, fn, _params);
}
function superRefine(fn, params) {
  return _superRefine(fn, params);
}
var describe2 = describe;
var meta2 = meta;
var ZodInstanceOf = /* @__PURE__ */ $constructor("ZodInstanceOf", (inst, def) => {
  ZodCustom.init(inst, def);
}, {
  properties(shape, params) {
    return this.check(_properties(shape, params));
  }
});
function _instanceof(cls, params = {}) {
  const inst = new ZodInstanceOf({
    type: "custom",
    check: "custom",
    fn: (data) => data instanceof cls,
    abort: true,
    ...normalizeParams(params)
  });
  inst._zod.bag.Class = cls;
  inst._zod.check = (payload) => {
    if (!(payload.value instanceof cls)) {
      payload.issues.push({
        code: "invalid_type",
        expected: cls.name,
        input: payload.value,
        inst,
        path: [...inst._zod.def.path ?? []]
      });
    }
  };
  return inst;
}
var stringbool = (...args) => _stringbool({
  Codec: ZodCodec,
  Boolean: ZodBoolean,
  String: ZodString
}, ...args);
function json(params) {
  const jsonSchema = lazy(() => {
    return union([string2(params), number2(), boolean2(), _null3(), array(jsonSchema), record(string2(), jsonSchema)]);
  });
  return jsonSchema;
}
function preprocess(fn, schema) {
  return new ZodPreprocess({
    type: "pipe",
    in: transform(fn),
    out: schema
  });
}
// node_modules/zod/v4/classic/iso.js
function datetime2(params) {
  return _isoDateTime(ZodISODateTime, params);
}
function date3(params) {
  return _isoDate(ZodISODate, params);
}
function time3(params) {
  return _isoTime(ZodISOTime, params);
}
function duration2(params) {
  return _isoDuration(ZodISODuration, params);
}

// node_modules/zod/v4/classic/from-json-schema.js
var z = {
  ...exports_schemas2,
  ...exports_checks2,
  iso: exports_iso
};
var RECOGNIZED_KEYS = /* @__PURE__ */ new Set([
  "$schema",
  "$ref",
  "$defs",
  "definitions",
  "$id",
  "id",
  "$comment",
  "$anchor",
  "$vocabulary",
  "$dynamicRef",
  "$dynamicAnchor",
  "type",
  "enum",
  "const",
  "anyOf",
  "oneOf",
  "allOf",
  "not",
  "properties",
  "required",
  "additionalProperties",
  "patternProperties",
  "propertyNames",
  "minProperties",
  "maxProperties",
  "items",
  "prefixItems",
  "additionalItems",
  "minItems",
  "maxItems",
  "uniqueItems",
  "contains",
  "minContains",
  "maxContains",
  "minLength",
  "maxLength",
  "pattern",
  "format",
  "minimum",
  "maximum",
  "exclusiveMinimum",
  "exclusiveMaximum",
  "multipleOf",
  "description",
  "default",
  "contentEncoding",
  "contentMediaType",
  "contentSchema",
  "unevaluatedItems",
  "unevaluatedProperties",
  "if",
  "then",
  "else",
  "dependentSchemas",
  "dependentRequired",
  "nullable",
  "readOnly"
]);
function detectVersion(schema, defaultTarget) {
  const $schema = schema.$schema;
  if ($schema === "https://json-schema.org/draft/2020-12/schema") {
    return "draft-2020-12";
  }
  if ($schema === "http://json-schema.org/draft-07/schema#") {
    return "draft-7";
  }
  if ($schema === "http://json-schema.org/draft-04/schema#") {
    return "draft-4";
  }
  return defaultTarget ?? "draft-2020-12";
}
function applyMinItems(items, minItems) {
  return items.map((item, index) => index < minItems ? item : item.optional());
}
function decodeJSONPointerSegment(segment) {
  return segment.replace(/~1/g, "/").replace(/~0/g, "~");
}
function resolveRef(ref, ctx) {
  if (!ref.startsWith("#")) {
    throw new Error("External $ref is not supported, only local refs (#/...) are allowed");
  }
  const path = ref.slice(1).split("/").filter(Boolean);
  if (path.length === 0) {
    return ctx.rootSchema;
  }
  const defsKey = ctx.version === "draft-2020-12" ? "$defs" : "definitions";
  if (path[0] === defsKey) {
    const key = path[1] === undefined ? undefined : decodeJSONPointerSegment(path[1]);
    if (!key || !ctx.defs[key]) {
      throw new Error(`Reference not found: ${ref}`);
    }
    return ctx.defs[key];
  }
  throw new Error(`Reference not found: ${ref}`);
}
function checkObjectGuards(objectSchema, guards) {
  const guard = z.transform((value) => value).check((payload) => {
    const value = payload.value;
    if (typeof value !== "object" || value === null || Array.isArray(value))
      return;
    const keys = Object.getOwnPropertyNames(value);
    if (guards.minProperties !== undefined && keys.length < guards.minProperties) {
      payload.issues.push({
        origin: "object",
        code: "too_small",
        minimum: guards.minProperties,
        inclusive: true,
        message: `Too small: expected object to have >=${guards.minProperties} properties`,
        input: value,
        inst: objectSchema,
        continue: true
      });
    }
    if (guards.maxProperties !== undefined && keys.length > guards.maxProperties) {
      payload.issues.push({
        origin: "object",
        code: "too_big",
        maximum: guards.maxProperties,
        inclusive: true,
        message: `Too big: expected object to have <=${guards.maxProperties} properties`,
        input: value,
        inst: objectSchema,
        continue: true
      });
    }
    if (guards.keySchema) {
      for (const key of keys) {
        const result = guards.keySchema.safeParse(key);
        if (result.success)
          continue;
        payload.issues.push({
          code: "invalid_key",
          origin: "record",
          issues: result.error.issues,
          input: key,
          path: [key],
          continue: true
        });
      }
    }
  });
  return guard.pipe(objectSchema);
}
function canonicalKey(value, seen) {
  if (value === null)
    return "z";
  const type = typeof value;
  if (type !== "object") {
    if (type === "number" && Number.isNaN(value))
      return null;
    const raw = String(value);
    return `${type[0]}${raw.length}:${raw}`;
  }
  if (seen.has(value))
    return null;
  seen.add(value);
  try {
    if (Array.isArray(value)) {
      const parts = [];
      for (const item of value) {
        const key = canonicalKey(item, seen);
        if (key === null)
          return null;
        parts.push(key);
      }
      return `a${parts.length}:[${parts.join(",")}]`;
    }
    const keys = Object.keys(value).sort();
    const parts = [];
    for (const k of keys) {
      const key = canonicalKey(value[k], seen);
      if (key === null)
        return null;
      parts.push(`${k.length}:${k}=${key}`);
    }
    return `o${parts.length}:{${parts.join(",")}}`;
  } finally {
    seen.delete(value);
  }
}
var SCHEMA_KEYWORDS = /* @__PURE__ */ new Set([
  "items",
  "prefixItems",
  "additionalItems",
  "additionalProperties",
  "contains",
  "propertyNames",
  "not",
  "if",
  "then",
  "else",
  "allOf",
  "anyOf",
  "oneOf",
  "unevaluatedItems",
  "unevaluatedProperties",
  "contentSchema"
]);
var SCHEMA_MAP_KEYWORDS = /* @__PURE__ */ new Set([
  "properties",
  "patternProperties",
  "dependentSchemas",
  "dependencies",
  "$defs",
  "definitions"
]);
function containsRef(value) {
  if (typeof value !== "object" || value === null)
    return false;
  if (Array.isArray(value))
    return value.some(containsRef);
  if (typeof value.$ref === "string")
    return true;
  return Object.entries(value).some(([key, sub]) => {
    if (SCHEMA_KEYWORDS.has(key))
      return containsRef(sub);
    if (!SCHEMA_MAP_KEYWORDS.has(key) || typeof sub !== "object" || sub === null)
      return false;
    return Object.values(sub).some(containsRef);
  });
}
function plural(n) {
  return n === 1 ? "element" : "elements";
}
function checkArrayGuards(arraySchema, guards) {
  const guard = z.transform((value) => value).check((payload) => {
    const items = payload.value;
    if (!Array.isArray(items))
      return;
    if (guards.uniqueItems === true) {
      const firstSeen = new Map;
      for (let i = 0;i < items.length; i++) {
        const key = canonicalKey(items[i], new Set);
        if (key === null)
          continue;
        const first = firstSeen.get(key);
        if (first === undefined) {
          firstSeen.set(key, i);
          continue;
        }
        payload.issues.push({
          code: "custom",
          message: `Array items must be unique: element at index ${i} duplicates the one at index ${first}`,
          input: items,
          path: [i],
          continue: true
        });
      }
    }
    if (guards.containsSchema) {
      const minContains = guards.minContains ?? 1;
      const ceiling = guards.maxContains !== undefined ? guards.maxContains + 1 : Number.POSITIVE_INFINITY;
      let matches = 0;
      for (const item of items) {
        if (guards.containsSchema.safeParse(item).success && ++matches >= ceiling)
          break;
      }
      if (matches < minContains) {
        payload.issues.push({
          code: "custom",
          message: `Array must contain at least ${minContains} matching ${plural(minContains)}; found ${matches}`,
          input: items,
          continue: true
        });
      }
      if (guards.maxContains !== undefined && matches > guards.maxContains) {
        payload.issues.push({
          code: "custom",
          message: `Array must contain at most ${guards.maxContains} matching ${plural(guards.maxContains)}`,
          input: items,
          continue: true
        });
      }
    }
  });
  return guard.pipe(arraySchema);
}
function getTupleRest(restSchema, ctx) {
  if (restSchema === false) {
    return;
  }
  if (restSchema === undefined || restSchema === true) {
    return z.any();
  }
  return convertSchema(restSchema, ctx);
}
var fullTime = /^(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d(?:\.\d+)?(?:Z|[+-](?:[01]\d|2[0-3]):[0-5]\d)$/;
function convertBaseSchema(schema, ctx) {
  if (schema.not !== undefined) {
    if (typeof schema.not === "object" && Object.keys(schema.not).length === 0) {
      return z.never();
    }
    throw new Error("not is not supported in Zod (except { not: {} } for never)");
  }
  if (schema.unevaluatedItems !== undefined) {
    throw new Error("unevaluatedItems is not supported");
  }
  if (schema.unevaluatedProperties !== undefined) {
    throw new Error("unevaluatedProperties is not supported");
  }
  if (schema.if !== undefined || schema.then !== undefined || schema.else !== undefined) {
    throw new Error("Conditional schemas (if/then/else) are not supported");
  }
  if (schema.dependentSchemas !== undefined || schema.dependentRequired !== undefined) {
    throw new Error("dependentSchemas and dependentRequired are not supported");
  }
  if (schema.$ref) {
    const refPath = schema.$ref;
    if (ctx.refs.has(refPath)) {
      return ctx.refs.get(refPath);
    }
    if (ctx.processing.has(refPath)) {
      return z.lazy(() => {
        if (!ctx.refs.has(refPath)) {
          throw new Error(`Circular reference not resolved: ${refPath}`);
        }
        return ctx.refs.get(refPath);
      });
    }
    ctx.processing.add(refPath);
    const resolved = resolveRef(refPath, ctx);
    const zodSchema = convertSchema(resolved, ctx);
    ctx.refs.set(refPath, zodSchema);
    ctx.processing.delete(refPath);
    return zodSchema;
  }
  if (schema.enum !== undefined) {
    const enumValues = schema.enum;
    if (ctx.version === "openapi-3.0" && schema.nullable === true && enumValues.length === 1 && enumValues[0] === null) {
      return z.null();
    }
    if (enumValues.length === 0) {
      return z.never();
    }
    if (enumValues.length === 1) {
      return z.literal(enumValues[0]);
    }
    if (enumValues.every((v) => typeof v === "string")) {
      return z.enum(enumValues);
    }
    const literalSchemas = enumValues.map((v) => z.literal(v));
    if (literalSchemas.length < 2) {
      return literalSchemas[0];
    }
    return z.union([literalSchemas[0], literalSchemas[1], ...literalSchemas.slice(2)]);
  }
  if (schema.const !== undefined) {
    return z.literal(schema.const);
  }
  const type = schema.type;
  if (Array.isArray(type)) {
    const typeSchemas = type.map((t) => {
      const typeSchema = { ...schema, type: t };
      return convertBaseSchema(typeSchema, ctx);
    });
    if (typeSchemas.length === 0) {
      return z.never();
    }
    if (typeSchemas.length === 1) {
      return typeSchemas[0];
    }
    return z.union(typeSchemas);
  }
  if (!type) {
    return z.any();
  }
  let zodSchema;
  switch (type) {
    case "string": {
      let stringSchema = z.string();
      if (schema.format) {
        const format = schema.format;
        if (format === "email") {
          stringSchema = stringSchema.check(z.email());
        } else if (format === "uri" || format === "uri-reference") {
          stringSchema = stringSchema.check(z.url());
        } else if (format === "uuid" || format === "guid") {
          stringSchema = stringSchema.check(z.uuid());
        } else if (format === "date-time") {
          stringSchema = stringSchema.check(z.iso.datetime({ offset: true }));
        } else if (format === "date") {
          stringSchema = stringSchema.check(z.iso.date());
        } else if (format === "time") {
          stringSchema = stringSchema.check(z.regex(fullTime));
        } else if (format === "duration") {
          stringSchema = stringSchema.check(z.iso.duration());
        } else if (format === "hostname") {
          stringSchema = stringSchema.check(z.hostname());
        } else if (format === "ipv4") {
          stringSchema = stringSchema.check(z.ipv4());
        } else if (format === "ipv6") {
          stringSchema = stringSchema.check(z.ipv6());
        } else if (format === "mac") {
          stringSchema = stringSchema.check(z.mac());
        } else if (format === "cidr") {
          stringSchema = stringSchema.check(z.cidrv4());
        } else if (format === "cidr-v6") {
          stringSchema = stringSchema.check(z.cidrv6());
        } else if (format === "base64") {
          stringSchema = stringSchema.check(z.base64());
        } else if (format === "base64url") {
          stringSchema = stringSchema.check(z.base64url());
        } else if (format === "e164") {
          stringSchema = stringSchema.check(z.e164());
        } else if (format === "credit_card") {
          stringSchema = stringSchema.check(z.creditCard());
        } else if (format === "iban") {
          stringSchema = stringSchema.check(z.iban());
        } else if (format === "jwt") {
          stringSchema = stringSchema.check(z.jwt());
        } else if (format === "emoji") {
          stringSchema = stringSchema.check(z.emoji());
        } else if (format === "nanoid") {
          stringSchema = stringSchema.check(z.nanoid());
        } else if (format === "cuid") {
          stringSchema = stringSchema.check(z.cuid());
        } else if (format === "cuid2") {
          stringSchema = stringSchema.check(z.cuid2());
        } else if (format === "ulid") {
          stringSchema = stringSchema.check(z.ulid());
        } else if (format === "xid") {
          stringSchema = stringSchema.check(z.xid());
        } else if (format === "ksuid") {
          stringSchema = stringSchema.check(z.ksuid());
        }
      }
      if (typeof schema.minLength === "number") {
        stringSchema = stringSchema.min(schema.minLength);
      }
      if (typeof schema.maxLength === "number") {
        stringSchema = stringSchema.max(schema.maxLength);
      }
      if (schema.pattern) {
        stringSchema = stringSchema.regex(new RegExp(schema.pattern));
      }
      zodSchema = stringSchema;
      break;
    }
    case "number":
    case "integer": {
      let numberSchema = type === "integer" ? z.number().int() : z.number();
      if (typeof schema.minimum === "number" && schema.exclusiveMinimum !== true) {
        numberSchema = numberSchema.min(schema.minimum);
      }
      if (typeof schema.maximum === "number" && schema.exclusiveMaximum !== true) {
        numberSchema = numberSchema.max(schema.maximum);
      }
      if (typeof schema.exclusiveMinimum === "number") {
        numberSchema = numberSchema.gt(schema.exclusiveMinimum);
      } else if (schema.exclusiveMinimum === true && typeof schema.minimum === "number") {
        numberSchema = numberSchema.gt(schema.minimum);
      }
      if (typeof schema.exclusiveMaximum === "number") {
        numberSchema = numberSchema.lt(schema.exclusiveMaximum);
      } else if (schema.exclusiveMaximum === true && typeof schema.maximum === "number") {
        numberSchema = numberSchema.lt(schema.maximum);
      }
      if (typeof schema.multipleOf === "number") {
        numberSchema = numberSchema.multipleOf(schema.multipleOf);
      }
      zodSchema = numberSchema;
      break;
    }
    case "boolean": {
      zodSchema = z.boolean();
      break;
    }
    case "null": {
      zodSchema = z.null();
      break;
    }
    case "object": {
      const shape = {};
      const properties = schema.properties || {};
      const requiredSet = new Set(schema.required || []);
      const additionalSchema = typeof schema.additionalProperties === "object" ? convertSchema(schema.additionalProperties, ctx) : undefined;
      for (const [key, propSchema] of Object.entries(properties)) {
        const propZodSchema = convertSchema(propSchema, ctx);
        assignProp(shape, key, requiredSet.has(key) ? propZodSchema : propZodSchema.optional());
      }
      if (schema.patternProperties) {
        const patternProps = schema.patternProperties;
        const patternKeys = Object.keys(patternProps);
        const looseRecords = [];
        for (const pattern of patternKeys) {
          const patternValue = convertSchema(patternProps[pattern], ctx);
          const keySchema = z.string().regex(new RegExp(pattern));
          looseRecords.push(z.looseRecord(keySchema, patternValue));
        }
        const schemasToIntersect = [];
        if (Object.keys(shape).length > 0) {
          schemasToIntersect.push(z.object(shape).passthrough());
        }
        schemasToIntersect.push(...looseRecords);
        if (schemasToIntersect.length === 0) {
          zodSchema = z.object({}).passthrough();
        } else if (schemasToIntersect.length === 1) {
          zodSchema = schemasToIntersect[0];
        } else {
          let result = z.intersection(schemasToIntersect[0], schemasToIntersect[1]);
          for (let i = 2;i < schemasToIntersect.length; i++) {
            result = z.intersection(result, schemasToIntersect[i]);
          }
          zodSchema = result;
        }
        if (schema.additionalProperties === false) {
          const propertyKeys = Object.keys(shape);
          const patterns = patternKeys.map((p) => new RegExp(p));
          const basePatternSchema = zodSchema;
          zodSchema = zodSchema.check((payload) => {
            if (!isPlainObject(payload.value))
              return;
            const unrecognized = [];
            for (const key of Object.keys(payload.value)) {
              if (propertyKeys.includes(key))
                continue;
              if (patterns.some((regex) => regex.test(key)))
                continue;
              unrecognized.push(key);
            }
            if (unrecognized.length) {
              payload.issues.push({
                code: "unrecognized_keys",
                keys: unrecognized,
                input: payload.value,
                inst: basePatternSchema
              });
            }
          });
        }
      } else {
        const objectSchema = z.object(shape);
        if (schema.additionalProperties === false) {
          zodSchema = objectSchema.strict();
        } else if (additionalSchema) {
          zodSchema = objectSchema.catchall(additionalSchema);
        } else {
          zodSchema = objectSchema.passthrough();
        }
      }
      const hasKeyGuard = schema.propertyNames !== undefined && schema.propertyNames !== true;
      const minProperties = typeof schema.minProperties === "number" ? schema.minProperties : undefined;
      const maxProperties = typeof schema.maxProperties === "number" ? schema.maxProperties : undefined;
      if (hasKeyGuard || minProperties !== undefined || maxProperties !== undefined) {
        let keySchema;
        if (hasKeyGuard) {
          const keyJSONSchema = typeof schema.propertyNames === "object" && schema.propertyNames.type === undefined ? { type: "string", ...schema.propertyNames } : schema.propertyNames;
          keySchema = convertSchema(keyJSONSchema, ctx);
        }
        zodSchema = checkObjectGuards(zodSchema, { keySchema, minProperties, maxProperties });
      }
      break;
    }
    case "array": {
      const prefixItems = schema.prefixItems;
      const items = schema.items;
      if (prefixItems && Array.isArray(prefixItems)) {
        const minItems = typeof schema.minItems === "number" ? schema.minItems : 0;
        const tupleItems = prefixItems.map((item) => convertSchema(item, ctx));
        const positionalItems = applyMinItems(tupleItems, minItems);
        const rest = !Array.isArray(items) ? getTupleRest(items, ctx) : undefined;
        const tupleSchema = z.tuple(positionalItems);
        zodSchema = rest ? tupleSchema.rest(rest) : tupleSchema;
        if (typeof schema.minItems === "number") {
          zodSchema = zodSchema.check(z.minLength(schema.minItems));
        }
        if (typeof schema.maxItems === "number") {
          zodSchema = zodSchema.check(z.maxLength(schema.maxItems));
        }
      } else if (Array.isArray(items)) {
        const minItems = typeof schema.minItems === "number" ? schema.minItems : 0;
        const tupleItems = items.map((item) => convertSchema(item, ctx));
        const positionalItems = applyMinItems(tupleItems, minItems);
        const rest = getTupleRest(schema.additionalItems, ctx);
        const tupleSchema = z.tuple(positionalItems);
        zodSchema = rest ? tupleSchema.rest(rest) : tupleSchema;
        if (typeof schema.minItems === "number") {
          zodSchema = zodSchema.check(z.minLength(schema.minItems));
        }
        if (typeof schema.maxItems === "number") {
          zodSchema = zodSchema.check(z.maxLength(schema.maxItems));
        }
      } else if (items !== undefined) {
        const element = convertSchema(items, ctx);
        let arraySchema = z.array(element);
        if (typeof schema.minItems === "number") {
          arraySchema = arraySchema.min(schema.minItems);
        }
        if (typeof schema.maxItems === "number") {
          arraySchema = arraySchema.max(schema.maxItems);
        }
        zodSchema = arraySchema;
      } else {
        zodSchema = z.array(z.any());
      }
      if (schema.uniqueItems === true || schema.contains !== undefined) {
        zodSchema = checkArrayGuards(zodSchema, {
          uniqueItems: schema.uniqueItems === true,
          containsSchema: schema.contains !== undefined ? convertSchema(schema.contains, ctx) : undefined,
          minContains: typeof schema.minContains === "number" ? schema.minContains : undefined,
          maxContains: typeof schema.maxContains === "number" ? schema.maxContains : undefined
        });
      }
      break;
    }
    default:
      throw new Error(`Unsupported type: ${type}`);
  }
  return zodSchema;
}
function convertSchema(schema, ctx) {
  if (typeof schema === "boolean") {
    return schema ? z.any() : z.never();
  }
  let baseSchema = convertBaseSchema(schema, ctx);
  const hasExplicitType = schema.type || schema.enum !== undefined || schema.const !== undefined;
  if (schema.anyOf && Array.isArray(schema.anyOf)) {
    const options = schema.anyOf.map((s) => convertSchema(s, ctx));
    const anyOfUnion = z.union(options);
    baseSchema = hasExplicitType ? z.intersection(baseSchema, anyOfUnion) : anyOfUnion;
  }
  if (schema.oneOf && Array.isArray(schema.oneOf)) {
    const options = schema.oneOf.map((s) => convertSchema(s, ctx));
    const oneOfUnion = z.xor(options);
    baseSchema = hasExplicitType ? z.intersection(baseSchema, oneOfUnion) : oneOfUnion;
  }
  if (schema.allOf && Array.isArray(schema.allOf)) {
    if (schema.allOf.length === 0) {
      baseSchema = hasExplicitType ? baseSchema : z.any();
    } else {
      let result = hasExplicitType ? baseSchema : convertSchema(schema.allOf[0], ctx);
      const startIdx = hasExplicitType ? 0 : 1;
      for (let i = startIdx;i < schema.allOf.length; i++) {
        result = z.intersection(result, convertSchema(schema.allOf[i], ctx));
      }
      baseSchema = result;
    }
  }
  if (schema.nullable === true && ctx.version === "openapi-3.0") {
    baseSchema = z.nullable(baseSchema);
  }
  if (schema.readOnly === true) {
    baseSchema = z.readonly(baseSchema);
  }
  if (schema.default !== undefined) {
    baseSchema = baseSchema.default(schema.default);
  }
  const extraMeta = {};
  const coreMetadataKeys = ["$id", "id", "$comment", "$anchor", "$vocabulary", "$dynamicRef", "$dynamicAnchor"];
  for (const key of coreMetadataKeys) {
    if (key in schema) {
      extraMeta[key] = schema[key];
    }
  }
  const contentMetadataKeys = ["contentEncoding", "contentMediaType", "contentSchema"];
  for (const key of contentMetadataKeys) {
    if (key in schema) {
      extraMeta[key] = schema[key];
    }
  }
  if (schema.type === "object" && schema.$ref === undefined) {
    if (schema.propertyNames !== undefined && !containsRef(schema.propertyNames)) {
      extraMeta.propertyNames = schema.propertyNames;
    }
    for (const key of ["minProperties", "maxProperties"]) {
      if (schema[key] !== undefined)
        extraMeta[key] = schema[key];
    }
  }
  if (schema.type === "array" && schema.$ref === undefined) {
    if (schema.contains !== undefined && !containsRef(schema.contains)) {
      extraMeta.contains = schema.contains;
    }
    for (const key of ["uniqueItems", "minContains", "maxContains"]) {
      if (schema[key] !== undefined)
        extraMeta[key] = schema[key];
    }
  }
  for (const key of Object.keys(schema)) {
    if (!RECOGNIZED_KEYS.has(key)) {
      assignProp(extraMeta, key, schema[key]);
    }
  }
  if (Object.keys(extraMeta).length > 0) {
    ctx.registry.add(baseSchema, extraMeta);
  }
  if (schema.description) {
    baseSchema = baseSchema.describe(schema.description);
  }
  return baseSchema;
}
function fromJSONSchema(schema, params) {
  if (typeof schema === "boolean") {
    return schema ? z.any() : z.never();
  }
  let normalized;
  try {
    normalized = JSON.parse(JSON.stringify(schema));
  } catch {
    throw new Error("fromJSONSchema input is not valid JSON (possibly cyclic); use $defs/$ref for recursive schemas");
  }
  const version = detectVersion(normalized, params?.defaultTarget);
  const defs = normalized.$defs || normalized.definitions || {};
  const ctx = {
    version,
    defs,
    refs: new Map,
    processing: new Set,
    rootSchema: normalized,
    registry: params?.registry ?? globalRegistry
  };
  return convertSchema(normalized, ctx);
}
// node_modules/zod/v4/classic/coerce.js
function string3(params) {
  return _coercedString(ZodString, params);
}
function number3(params) {
  return _coercedNumber(ZodNumber, params);
}
function boolean3(params) {
  return _coercedBoolean(ZodBoolean, params);
}
function bigint3(params) {
  return _coercedBigint(ZodBigInt, params);
}
function date4(params) {
  return _coercedDate(ZodDate, params);
}
// src/json-schema.ts
var DEFAULT_JSON_SCHEMA_MAX_TOKENS = 15000;
var JSON_LIMITS = {
  schemaBytes: 256 << 10,
  outputBytes: 4 << 20,
  depth: 64,
  items: 1e4,
  nodes: 1e5,
  issues: 8
};
function jsonShapeError(value, what) {
  const stack = [[value, 1]];
  let nodes = 0;
  while (stack.length) {
    const [v, depth] = stack.pop();
    if (++nodes > JSON_LIMITS.nodes)
      return `${what} has more than ${JSON_LIMITS.nodes} values`;
    if (depth > JSON_LIMITS.depth)
      return `${what} is nested deeper than ${JSON_LIMITS.depth}`;
    if (v === null || typeof v !== "object")
      continue;
    const children = Array.isArray(v) ? v : Object.values(v);
    if (children.length > JSON_LIMITS.items) {
      return `${what} has a list or object of more than ${JSON_LIMITS.items} items`;
    }
    for (const c of children)
      stack.push([c, depth + 1]);
  }
  return null;
}
var JsonValueSchema = lazy(() => union([
  string2(),
  number2(),
  boolean2(),
  _null3(),
  array(JsonValueSchema),
  record(string2(), JsonValueSchema)
]));
var JsonSchemaObjectSchema = record(string2(), JsonValueSchema);
var PositiveIntSchema = exports_coerce.number().int().positive();
function resolveJsonSchemaMaxTokens(desplegaArgs, env = process.env) {
  const flag = desplegaArgs.find((arg) => ["json-schema-max-tokens", "schema-max-tokens"].includes(arg.name));
  const raw = flag?.value === undefined || flag.value === true ? env.CLAUDE_BRIDGE_JSON_SCHEMA_MAX_TOKENS : flag.value;
  if (raw === undefined || raw === false || raw === "")
    return DEFAULT_JSON_SCHEMA_MAX_TOKENS;
  const parsed = PositiveIntSchema.safeParse(raw);
  if (!parsed.success) {
    throw new Error(`Invalid JSON schema token limit \`${String(raw)}\`; expected a positive integer.`);
  }
  return parsed.data;
}
function loadJsonSchema(input, cwd, maxTokens) {
  const rawText = readSchemaInput(input, cwd);
  return loadJsonSchemaFromText(rawText, maxTokens);
}
function loadJsonSchemaFromText(rawText, maxTokens) {
  if (Buffer.byteLength(rawText, "utf8") > JSON_LIMITS.schemaBytes) {
    throw new Error(`--json-schema is larger than ${JSON_LIMITS.schemaBytes >> 10} KiB.`);
  }
  let parsed;
  try {
    parsed = JSON.parse(rawText);
  } catch (err) {
    throw new Error(`Unable to parse --json-schema as JSON: ${err.message}`);
  }
  const shape = jsonShapeError(parsed, "--json-schema");
  if (shape)
    throw new Error(`${shape}.`);
  const schema = JsonSchemaObjectSchema.safeParse(parsed);
  if (!schema.success || Array.isArray(parsed)) {
    throw new Error("--json-schema must resolve to a JSON Schema object.");
  }
  const compact = JSON.stringify(schema.data);
  const estimatedTokens = estimateTokens(compact);
  if (estimatedTokens > maxTokens) {
    throw new Error(`--json-schema is too large: estimated ${estimatedTokens} tokens exceeds limit ${maxTokens}.`);
  }
  return { schema: schema.data, compact, estimatedTokens, maxTokens };
}
function makeJsonSchemaSystemPrompt(compactSchema) {
  return [
    '<critical when="before replying to the user">',
    "Your final assistant reply must end with valid JSON matching this schema. No prose after the JSON.",
    compactSchema,
    "</critical>"
  ].join(`
`);
}
function mergeAppendSystemPrompt(args, addition) {
  const kept = [];
  const prompts = [];
  for (let i = 0;i < args.length; i++) {
    const arg = args[i];
    if (arg === "--append-system-prompt") {
      if (i + 1 < args.length)
        prompts.push(args[++i]);
      continue;
    }
    if (arg.startsWith("--append-system-prompt=")) {
      prompts.push(arg.slice("--append-system-prompt=".length));
      continue;
    }
    kept.push(arg);
  }
  const merged = [...prompts, addition].filter(Boolean).join(`

`);
  return [...kept, "--append-system-prompt", merged];
}
function extractJsonValue(text) {
  const trimmed = text.trim();
  if (trimmed) {
    const parsed = parseJson(trimmed);
    if (parsed.ok)
      return { value: parsed.value, json: trimmed, source: "full" };
  }
  const fenced = extractLastFencedJson(text);
  if (fenced)
    return fenced;
  return extractLastBalancedJson(text);
}
function extractAndValidateStructuredOutput(text, schema) {
  if (Buffer.byteLength(text, "utf8") > JSON_LIMITS.outputBytes) {
    return {
      ok: false,
      message: `Claude reply is larger than ${JSON_LIMITS.outputBytes >> 20} MiB; it was not checked against --json-schema.`,
      rawResponse: ""
    };
  }
  const extracted = extractJsonValue(text);
  if (!extracted) {
    return {
      ok: false,
      message: "Claude reply did not contain a valid JSON value for --json-schema. Expected full JSON, a fenced json block, or a final balanced JSON object/array.",
      rawResponse: text
    };
  }
  const shape = jsonShapeError(extracted.value, "The reply's JSON");
  if (shape)
    return { ok: false, message: `${shape}; it was not checked against --json-schema.`, rawResponse: text };
  const validated = validateJsonAgainstSchema(extracted.value, schema);
  if (!validated.ok) {
    return {
      ok: false,
      message: `${validated.message} ${validated.issues.join("; ")}`,
      rawResponse: text
    };
  }
  return { ok: true, value: validated.value, source: extracted.source };
}
function validateJsonAgainstSchema(value, schema) {
  let zodSchema;
  try {
    zodSchema = fromJSONSchema(schema);
  } catch (err) {
    return {
      ok: false,
      message: "Unable to convert --json-schema to a Zod schema.",
      issues: [err.message]
    };
  }
  const parsed = zodSchema.safeParse(value);
  if (parsed.success)
    return { ok: true, value: parsed.data };
  return {
    ok: false,
    message: "Extracted JSON does not conform to --json-schema.",
    issues: [
      ...parsed.error.issues.slice(0, JSON_LIMITS.issues).map((issue) => {
        const path = issue.path.length ? issue.path.join(".") : "<root>";
        return `${path}: ${issue.message}`;
      }),
      ...parsed.error.issues.length > JSON_LIMITS.issues ? [`and ${parsed.error.issues.length - JSON_LIMITS.issues} more`] : []
    ]
  };
}
function readSchemaInput(input, cwd) {
  const trimmed = input.trim();
  if (trimmed.startsWith("{"))
    return trimmed;
  const path = resolve3(cwd, input);
  if (existsSync2(path))
    return readFileSync3(path, "utf8");
  return trimmed;
}
function estimateTokens(value) {
  return Math.ceil(value.length / 4);
}
function extractLastFencedJson(text) {
  const fenceRe = /```(?:json|JSON)?\s*([\s\S]*?)```/g;
  let match;
  let last = null;
  while (match = fenceRe.exec(text)) {
    const candidate = (match[1] ?? "").trim();
    if (!candidate)
      continue;
    const parsed = parseJson(candidate);
    if (parsed.ok)
      last = { value: parsed.value, json: candidate, source: "fenced" };
  }
  return last;
}
function extractLastBalancedJson(text) {
  let best = null;
  for (let start = 0;start < text.length; start++) {
    const ch = text[start];
    if (ch !== "{" && ch !== "[")
      continue;
    const json = balancedSliceFrom(text, start);
    if (!json)
      continue;
    const parsed = parseJson(json);
    if (!parsed.ok)
      continue;
    const end = start + json.length;
    if (!best || end > best.end || end === best.end && start < best.start) {
      best = { start, end, value: parsed.value, json };
    }
  }
  return best ? { value: best.value, json: best.json, source: "balanced" } : null;
}
function balancedSliceFrom(text, start) {
  const stack = [];
  let inString = false;
  let escaped = false;
  for (let i = start;i < text.length; i++) {
    const ch = text[i];
    if (inString) {
      if (escaped) {
        escaped = false;
      } else if (ch === "\\") {
        escaped = true;
      } else if (ch === '"') {
        inString = false;
      }
      continue;
    }
    if (ch === '"') {
      inString = true;
      continue;
    }
    if (ch === "{") {
      stack.push("}");
      continue;
    }
    if (ch === "[") {
      stack.push("]");
      continue;
    }
    if (ch === "}" || ch === "]") {
      if (stack.pop() !== ch)
        return null;
      if (stack.length === 0)
        return text.slice(start, i + 1);
    }
  }
  return null;
}
function parseJson(value) {
  try {
    return { ok: true, value: JSON.parse(value) };
  } catch {
    return { ok: false };
  }
}

// src/hook-install.ts
import { copyFileSync as copyFileSync2, existsSync as existsSync3, mkdirSync as mkdirSync2, readFileSync as readFileSync4 } from "fs";
import { dirname as dirname4, resolve as resolve4 } from "path";
var JSON_SCHEMA_STOP_HOOK_ARG = "--desplega-internal-json-schema-stop-hook";
var RUNTIME_HOOK_ARG = "--desplega-internal-runtime-hook";
function buildJsonSchemaStopHookCommand(argv = process.argv, bun = process.execPath) {
  return buildInternalHookCommand(JSON_SCHEMA_STOP_HOOK_ARG, argv, bun);
}
function buildRuntimeHookCommand(argv = process.argv, bun = process.execPath) {
  return buildInternalHookCommand(RUNTIME_HOOK_ARG, argv, bun);
}
function buildInternalHookCommand(arg, argv = process.argv, bun = process.execPath) {
  const entry = argv[1] ? resolve4(process.cwd(), argv[1]) : resolve4("src/cli.ts");
  return [bun, entry, arg].map(shellQuote).join(" ");
}
function installRuntimeHooks(opts) {
  const settingsPath = opts.settingsPath;
  const settings = readSettings(settingsPath);
  settings.hooks ??= {};
  const before = JSON.stringify(settings.hooks);
  settings.hooks.Stop = upsertBridgeHook(settings.hooks.Stop ?? [], opts.command, RUNTIME_HOOK_ARG);
  settings.hooks.StopFailure = upsertBridgeHook(settings.hooks.StopFailure ?? [], opts.command, RUNTIME_HOOK_ARG);
  settings.hooks.MessageDisplay = upsertBridgeHook(settings.hooks.MessageDisplay ?? [], opts.command, RUNTIME_HOOK_ARG);
  settings.hooks.UserPromptSubmit = upsertBridgeHook(settings.hooks.UserPromptSubmit ?? [], opts.command, RUNTIME_HOOK_ARG);
  settings.hooks.SessionStart = upsertBridgeHook(settings.hooks.SessionStart ?? [], opts.command, RUNTIME_HOOK_ARG);
  settings.hooks.PreToolUse = upsertBridgeHook(settings.hooks.PreToolUse ?? [], opts.command, RUNTIME_HOOK_ARG);
  const changed = before !== JSON.stringify(settings.hooks);
  if (changed)
    writeSettings(settingsPath, settings);
  return { settingsPath, changed };
}
function installUserContextHooks(opts) {
  const settings = readSettings(opts.settingsPath);
  settings.hooks ??= {};
  for (const event of ["PostToolUse", "PostToolUseFailure"]) {
    settings.hooks[event] = upsertBridgeHook(settings.hooks[event] ?? [], opts.command, RUNTIME_HOOK_ARG);
  }
  writeSettings(opts.settingsPath, settings);
}
function installJsonSchemaStopHook(opts) {
  const settingsPath = opts.settingsPath;
  const settings = readSettings(settingsPath);
  settings.hooks ??= {};
  const stopGroups = settings.hooks.Stop ?? [];
  const current = removeStaleJsonSchemaStopHookGroups(stopGroups, opts.command);
  const alreadyInstalled = current.some((group) => (group.hooks ?? []).some((handler) => handler.command === opts.command));
  settings.hooks.Stop = alreadyInstalled ? current : [
    ...current,
    {
      matcher: "*",
      hooks: [
        {
          type: "command",
          command: opts.command,
          timeout: 30
        }
      ]
    }
  ];
  const changed = stopGroups.length !== settings.hooks.Stop.length || !alreadyInstalled || JSON.stringify(stopGroups) !== JSON.stringify(settings.hooks.Stop);
  if (changed)
    writeSettings(settingsPath, settings);
  return { settingsPath, changed };
}
function readSettings(settingsPath) {
  if (!existsSync3(settingsPath))
    return {};
  const raw = readFileSync4(settingsPath, "utf8").trim();
  if (!raw)
    return {};
  return JSON.parse(raw);
}
function writeSettings(settingsPath, settings) {
  mkdirSync2(dirname4(settingsPath), { recursive: true });
  const backup = settingsPath + ".claude-bridge-backup";
  if (existsSync3(settingsPath) && !existsSync3(backup))
    copyFileSync2(settingsPath, backup);
  writePrivateAtomic(settingsPath, JSON.stringify(settings, null, 2) + `
`);
}
function removeStaleJsonSchemaStopHookGroups(groups, keepCommand) {
  return groups.map((group) => ({
    ...group,
    hooks: (group.hooks ?? []).filter((handler) => !handler.command?.includes(JSON_SCHEMA_STOP_HOOK_ARG) || keepCommand !== undefined && handler.command === keepCommand && group.matcher === "*")
  })).filter((group) => (group.hooks ?? []).length > 0);
}
function upsertBridgeHook(groups, command, marker) {
  const current = groups.map((group) => ({
    ...group,
    hooks: (group.hooks ?? []).filter((handler) => !handler.command?.includes(marker) || handler.command === command && group.matcher === "*")
  })).filter((group) => (group.hooks ?? []).length > 0);
  const alreadyInstalled = current.some((group) => (group.hooks ?? []).some((handler) => handler.command === command));
  if (alreadyInstalled)
    return current;
  return [
    ...current,
    {
      matcher: "*",
      hooks: [
        {
          type: "command",
          command,
          timeout: 30
        }
      ]
    }
  ];
}
function shellQuote(value) {
  return `'${value.replace(/'/g, `'\\''`)}'`;
}

// src/print-result.ts
function makePrintErrorResult(message, options = {}) {
  return {
    type: "result",
    subtype: "error",
    is_error: true,
    error: message,
    ...options.rawResponse !== undefined ? { raw_response: options.rawResponse } : {},
    ...options.runState ? { run_state: options.runState } : {},
    ...options.paneTail ? { pane_tail: options.paneTail } : {},
    ...options.stderrLog ? { stderr_log: options.stderrLog } : {},
    ...options.sessionId ? { session_id: options.sessionId } : {},
    ...options.debug?.length ? { debug: options.debug } : {}
  };
}

// src/stop-hook.ts
import { appendFileSync, existsSync as existsSync4, mkdirSync as mkdirSync3, readFileSync as readFileSync6, writeFileSync as writeFileSync2 } from "fs";
import { join as join4 } from "path";

// src/user-context.ts
import { lstatSync, readFileSync as readFileSync5, renameSync as renameSync2, writeFileSync } from "fs";
function consumeUserContext(input, env) {
  const event = input.hook_event_name;
  const path = env.CLAUDE_BRIDGE_USER_CONTEXT_PATH;
  if (!path || env.CLAUDE_BRIDGE_RUNTIME_HOOK !== "1" || !["PostToolUse", "PostToolUseFailure"].includes(String(event)))
    return null;
  const prefix = `mcp__${process.env.CLAUDE_BRIDGE_MCP_SERVER || "magpie"}__`;
  if (typeof input.tool_name !== "string" || !input.tool_name.startsWith(prefix) || typeof input.tool_use_id !== "string")
    return null;
  let payload;
  try {
    const stat = lstatSync(path);
    if (!stat.isFile() || stat.size > 1100000 || (stat.mode & 63) !== 0)
      return null;
    payload = JSON.parse(readFileSync5(path, "utf8"));
    if (!payload || typeof payload.id !== "string" || !/^[a-f0-9]{16,64}$/.test(payload.id) || typeof payload.text !== "string" || payload.text.length === 0 || Buffer.byteLength(payload.text) > 1048576 || !Array.isArray(payload.tool_call_ids) || !payload.tool_call_ids.every((id) => typeof id === "string" && id !== "") || !payload.tool_call_ids.includes(input.tool_use_id))
      return null;
    renameSync2(path, path + ".claimed-" + payload.id);
  } catch {
    return null;
  }
  const context = `The outer client sent this new user message while the external tool was running. It is user input, separate from the tool's returned data:

` + payload.text;
  writeFileSync(path + ".ack", JSON.stringify({ id: payload.id, tool_use_id: input.tool_use_id }) + `
`, { mode: 384 });
  return { hookSpecificOutput: { hookEventName: event, additionalContext: context } };
}

// src/stop-hook.ts
var SCHEMA_STOP_FEEDBACK_MARKER = "claude-bridge schema output required";
var DEFAULT_MAX_SCHEMA_STOP_BLOCKS = 3;
async function runJsonSchemaStopHook() {
  const input = await new Response(Bun.stdin.stream()).text();
  const decision = evaluateJsonSchemaStopHook(input, process.env);
  if (decision)
    process.stdout.write(JSON.stringify(decision) + `
`);
}
async function runRuntimeHook() {
  const input = await new Response(Bun.stdin.stream()).text();
  recordRuntimeHook(input, process.env);
  const parsed = parseHookInput(input);
  const context = parsed ? consumeUserContext(parsed, process.env) : null;
  if (context)
    process.stdout.write(JSON.stringify(context) + `
`);
}
function recordRuntimeHook(inputText, env = process.env) {
  if (env.CLAUDE_BRIDGE_RUNTIME_HOOK !== "1" || !env.CLAUDE_BRIDGE_RUN_DIR)
    return;
  const input = parseHookInput(inputText);
  if (!input?.hook_event_name)
    return;
  mkdirSync3(env.CLAUDE_BRIDGE_RUN_DIR, { recursive: true });
  if (typeof input.transcript_path === "string" && input.transcript_path) {
    const transcriptEventPath = join4(env.CLAUDE_BRIDGE_RUN_DIR, "transcript-event.json");
    if (!existsSync4(transcriptEventPath))
      writeFileSync2(transcriptEventPath, inputText.trim() + `
`);
  }
  if (input.hook_event_name === "Stop") {
    if (evaluateJsonSchemaStopHook(inputText, env)?.decision === "block")
      return;
    writeFileSync2(join4(env.CLAUDE_BRIDGE_RUN_DIR, "stop-event.json"), inputText.trim() + `
`);
    return;
  }
  if (input.hook_event_name === "StopFailure") {
    writeFileSync2(join4(env.CLAUDE_BRIDGE_RUN_DIR, "stop-event.json"), inputText.trim() + `
`);
    return;
  }
  if (input.hook_event_name === "MessageDisplay") {
    appendFileSync(join4(env.CLAUDE_BRIDGE_RUN_DIR, "message-display.jsonl"), JSON.stringify(normalizeMessageDisplay(input)) + `
`);
  }
}
function evaluateJsonSchemaStopHook(inputText, env = process.env) {
  if (env.CLAUDE_BRIDGE_SCHEMA_STOP_HOOK !== "1" || !env.CLAUDE_BRIDGE_JSON_SCHEMA_PATH) {
    return null;
  }
  const input = parseHookInput(inputText);
  if (!input || input.hook_event_name !== "Stop")
    return null;
  const schema = loadJsonSchemaFromText(readFileSync6(env.CLAUDE_BRIDGE_JSON_SCHEMA_PATH, "utf8"), Number.MAX_SAFE_INTEGER).schema;
  const candidates = collectCandidateReplies(input);
  for (const candidate of candidates) {
    const result = extractAndValidateStructuredOutput(candidate, schema);
    if (result.ok)
      return null;
  }
  const feedbackCount = input.transcript_path ? countSchemaStopFeedbacks(input.transcript_path) : 0;
  if (feedbackCount >= resolveMaxSchemaStopBlocks(env))
    return null;
  return {
    decision: "block",
    reason: [
      `${SCHEMA_STOP_FEEDBACK_MARKER}: reply again with valid JSON matching the provided JSON Schema.`,
      "No prose after the JSON."
    ].join(" ")
  };
}
function parseHookInput(inputText) {
  try {
    return JSON.parse(inputText);
  } catch {
    return null;
  }
}
function normalizeMessageDisplay(input) {
  const delta = typeof input.delta === "string" ? input.delta : typeof input.message === "string" ? input.message : typeof input.last_assistant_message === "string" ? input.last_assistant_message : "";
  return {
    type: "message_display",
    delta,
    raw: input
  };
}
function collectCandidateReplies(input) {
  const candidates = [];
  const transcriptReply = input.transcript_path ? latestAssistantTextFromTranscript(input.transcript_path) : null;
  if (transcriptReply)
    candidates.push(transcriptReply);
  if (input.last_assistant_message)
    candidates.push(input.last_assistant_message);
  return candidates;
}
function latestAssistantTextFromTranscript(transcriptPath) {
  try {
    const rows = readTranscriptRows(transcriptPath);
    for (let i = rows.length - 1;i >= 0; i--) {
      if (rows[i]?.type !== "assistant")
        continue;
      const text = textFromMessageContent(rows[i].message?.content).trim();
      if (text)
        return text;
    }
  } catch {
    return null;
  }
  return null;
}
function countSchemaStopFeedbacks(transcriptPath) {
  try {
    return readTranscriptRows(transcriptPath).filter((row) => textFromMessageContent(row.message?.content).includes(SCHEMA_STOP_FEEDBACK_MARKER)).length;
  } catch {
    return 0;
  }
}
function readTranscriptRows(transcriptPath) {
  return readFileSync6(transcriptPath, "utf8").split(`
`).filter(Boolean).map((line) => JSON.parse(line));
}
function textFromMessageContent(content) {
  if (typeof content === "string")
    return content;
  if (!Array.isArray(content))
    return "";
  return content.map((part) => {
    if (!part || typeof part !== "object")
      return "";
    const block = part;
    return block.type === "text" ? String(block.text ?? "") : "";
  }).join(`
`);
}
function resolveMaxSchemaStopBlocks(env) {
  const value = Number(env.CLAUDE_BRIDGE_SCHEMA_STOP_HOOK_MAX_BLOCKS ?? DEFAULT_MAX_SCHEMA_STOP_BLOCKS);
  if (!Number.isFinite(value) || value < 1)
    return DEFAULT_MAX_SCHEMA_STOP_BLOCKS;
  return Math.floor(value);
}

// src/native-api-error.ts
class NativeApiErrorTurn {
  pending = null;
  observe(row) {
    if (row.type === "assistant") {
      const message = row.message && typeof row.message === "object" ? row.message : {};
      if (row.isApiErrorMessage === true && message.model === "<synthetic>") {
        const blocks = Array.isArray(message.content) ? message.content : [];
        const text = blocks.filter((block) => Boolean(block && typeof block === "object")).filter((block) => block.type === "text" && typeof block.text === "string").map((block) => String(block.text).trim()).filter(Boolean).join(`
`);
        this.pending = text ? {
          message: text,
          ...row.error === "invalid_request" ? { status: 400 } : {}
        } : null;
      } else if (message.model !== "<synthetic>") {
        this.pending = null;
      }
    }
    if (row.type === "system" && row.subtype === "turn_duration" && this.pending) {
      const failure = this.pending;
      this.pending = null;
      return failure;
    }
    return null;
  }
}

// src/launch-command.ts
function interactiveMcpBackgroundEnvArgs() {
  return [
    "CLAUDE_CODE_MCP_AUTO_BACKGROUND_MS=0",
    "CLAUDE_CODE_MCP_ALWAYS_LOAD_JOIN_BEFORE_INPUT=1",
    "CLAUDE_CODE_MCP_STARTUP_WAIT_MS=30000"
  ];
}
function nativeAutoCompactionLaunchEnv(env) {
  if (env.DISABLE_AUTO_COMPACT === "1") {
    return { unsetEnvArgs: [], envArgs: ["DISABLE_AUTO_COMPACT=1"] };
  }
  return { unsetEnvArgs: ["-u", "DISABLE_AUTO_COMPACT"], envArgs: [] };
}
function shellQuote2(value) {
  if (value.length === 0)
    return "''";
  return `'${value.replace(/'/g, `'\\''`)}'`;
}
function buildClaudeLaunchCommand(opts) {
  const envCommand = [
    "env",
    ...opts.unsetEnvArgs,
    ...opts.envArgs,
    opts.claudePath,
    ...opts.claudeArgs
  ].map(shellQuote2).join(" ");
  const holdSeconds = Math.max(1, Math.ceil(opts.holdMs / 1000));
  const script = [
    envCommand,
    "status=$?",
    `printf '%s\\n' "$status" > ${shellQuote2(opts.exitStatusPath)}`,
    `printf '\\n[claude-bridge] claude exited with status %s\\n' "$status" >&2`,
    `sleep ${holdSeconds}`,
    `exit "$status"`
  ].join(`
`);
  return `/bin/sh -lc ${shellQuote2(script)}`;
}

// src/native-tmux-guard.ts
import { closeSync as closeSync2, fsyncSync as fsyncSync2, openSync as openSync2, renameSync as renameSync3, unlinkSync as unlinkSync2, writeSync as writeSync2 } from "fs";
import { dirname as dirname5, isAbsolute, join as join5 } from "path";
import { createHash, randomUUID as randomUUID2 } from "crypto";
function socketEnvironmentDigest() {
  return createHash("sha256").update(`${process.env.TMUX_TMPDIR ?? ""}\x00${process.env.TMUX ?? ""}`).digest("hex");
}
function writeNativeTmuxGuard(path, session, socket = "") {
  if (!isAbsolute(path) || !/^claude-bridge-[a-fA-F0-9]{8}$/.test(session)) {
    throw new Error("invalid native tmux ownership marker path/session");
  }
  const folder = dirname5(path);
  const tmp = join5(folder, `.native-tmux-${randomUUID2()}.tmp`);
  let fd;
  try {
    fd = openSync2(tmp, "wx", 384);
    const payload = Buffer.from(JSON.stringify({
      version: 2,
      session,
      socket_env_sha256: socketEnvironmentDigest(),
      ...socket ? { socket } : {}
    }) + `
`, "utf8");
    const n = writeSync2(fd, payload);
    if (n !== payload.length)
      throw new Error("incomplete native tmux marker write");
    fsyncSync2(fd);
    closeSync2(fd);
    fd = undefined;
    renameSync3(tmp, path);
    const directory = openSync2(folder, "r");
    try {
      fsyncSync2(directory);
    } finally {
      closeSync2(directory);
    }
  } catch (error) {
    if (fd !== undefined)
      closeSync2(fd);
    try {
      unlinkSync2(tmp);
    } catch {}
    throw error;
  }
}

// src/model-pricing-data.ts
var MODEL_COST_RATES = {
  "claude-3-5-haiku-20241022": { input: 0.8, output: 4, cacheRead: 0.08, cacheWrite5m: 1 },
  "claude-3-5-haiku-latest": { input: 0.8, output: 4, cacheRead: 0.08, cacheWrite5m: 1 },
  "claude-3-5-sonnet-20240620": { input: 3, output: 15, cacheRead: 0.3, cacheWrite5m: 3.75 },
  "claude-3-5-sonnet-20241022": { input: 3, output: 15, cacheRead: 0.3, cacheWrite5m: 3.75 },
  "claude-3-7-sonnet-20250219": { input: 3, output: 15, cacheRead: 0.3, cacheWrite5m: 3.75 },
  "claude-3-haiku-20240307": { input: 0.25, output: 1.25, cacheRead: 0.03, cacheWrite5m: 0.3 },
  "claude-3-opus-20240229": { input: 15, output: 75, cacheRead: 1.5, cacheWrite5m: 18.75 },
  "claude-3-sonnet-20240229": { input: 3, output: 15, cacheRead: 0.3, cacheWrite5m: 0.3 },
  "claude-fable-5": { input: 10, output: 50, cacheRead: 1, cacheWrite5m: 12.5 },
  "claude-haiku-4-5": { input: 1, output: 5, cacheRead: 0.1, cacheWrite5m: 1.25 },
  "claude-haiku-4-5-20251001": { input: 1, output: 5, cacheRead: 0.1, cacheWrite5m: 1.25 },
  "claude-opus-4-0": { input: 15, output: 75, cacheRead: 1.5, cacheWrite5m: 18.75 },
  "claude-opus-4-1": { input: 15, output: 75, cacheRead: 1.5, cacheWrite5m: 18.75 },
  "claude-opus-4-1-20250805": { input: 15, output: 75, cacheRead: 1.5, cacheWrite5m: 18.75 },
  "claude-opus-4-20250514": { input: 15, output: 75, cacheRead: 1.5, cacheWrite5m: 18.75 },
  "claude-opus-4-5": { input: 5, output: 25, cacheRead: 0.5, cacheWrite5m: 6.25 },
  "claude-opus-4-5-20251101": { input: 5, output: 25, cacheRead: 0.5, cacheWrite5m: 6.25 },
  "claude-opus-4-6": { input: 5, output: 25, cacheRead: 0.5, cacheWrite5m: 6.25 },
  "claude-opus-4-7": { input: 5, output: 25, cacheRead: 0.5, cacheWrite5m: 6.25 },
  "claude-opus-4-8": { input: 5, output: 25, cacheRead: 0.5, cacheWrite5m: 6.25 },
  "claude-sonnet-4-0": { input: 3, output: 15, cacheRead: 0.3, cacheWrite5m: 3.75 },
  "claude-sonnet-4-20250514": { input: 3, output: 15, cacheRead: 0.3, cacheWrite5m: 3.75 },
  "claude-sonnet-4-5": { input: 3, output: 15, cacheRead: 0.3, cacheWrite5m: 3.75 },
  "claude-sonnet-4-5-20250929": { input: 3, output: 15, cacheRead: 0.3, cacheWrite5m: 3.75 },
  "claude-sonnet-4-6": { input: 3, output: 15, cacheRead: 0.3, cacheWrite5m: 3.75 }
};
var MODEL_LIMITS = {
  "claude-3-5-haiku-20241022": { contextWindow: 200000, maxOutputTokens: 8192 },
  "claude-3-5-haiku-latest": { contextWindow: 200000, maxOutputTokens: 8192 },
  "claude-3-5-sonnet-20240620": { contextWindow: 200000, maxOutputTokens: 8192 },
  "claude-3-5-sonnet-20241022": { contextWindow: 200000, maxOutputTokens: 8192 },
  "claude-3-7-sonnet-20250219": { contextWindow: 200000, maxOutputTokens: 64000 },
  "claude-3-haiku-20240307": { contextWindow: 200000, maxOutputTokens: 4096 },
  "claude-3-opus-20240229": { contextWindow: 200000, maxOutputTokens: 4096 },
  "claude-3-sonnet-20240229": { contextWindow: 200000, maxOutputTokens: 4096 },
  "claude-fable-5": { contextWindow: 1e6, maxOutputTokens: 128000 },
  "claude-haiku-4-5": { contextWindow: 200000, maxOutputTokens: 64000 },
  "claude-haiku-4-5-20251001": { contextWindow: 200000, maxOutputTokens: 64000 },
  "claude-opus-4-0": { contextWindow: 200000, maxOutputTokens: 32000 },
  "claude-opus-4-1": { contextWindow: 200000, maxOutputTokens: 32000 },
  "claude-opus-4-1-20250805": { contextWindow: 200000, maxOutputTokens: 32000 },
  "claude-opus-4-20250514": { contextWindow: 200000, maxOutputTokens: 32000 },
  "claude-opus-4-5": { contextWindow: 200000, maxOutputTokens: 64000 },
  "claude-opus-4-5-20251101": { contextWindow: 200000, maxOutputTokens: 64000 },
  "claude-opus-4-6": { contextWindow: 1e6, maxOutputTokens: 128000 },
  "claude-opus-4-7": { contextWindow: 1e6, maxOutputTokens: 128000 },
  "claude-opus-4-8": { contextWindow: 1e6, maxOutputTokens: 128000 },
  "claude-sonnet-4-0": { contextWindow: 200000, maxOutputTokens: 64000 },
  "claude-sonnet-4-20250514": { contextWindow: 200000, maxOutputTokens: 64000 },
  "claude-sonnet-4-5": { contextWindow: 200000, maxOutputTokens: 64000 },
  "claude-sonnet-4-5-20250929": { contextWindow: 200000, maxOutputTokens: 64000 },
  "claude-sonnet-4-6": { contextWindow: 1e6, maxOutputTokens: 64000 }
};
var MODEL_ALIASES = {
  "claude-fable": "claude-fable-5",
  "claude-haiku": "claude-haiku-4-5",
  "claude-opus": "claude-opus-4-8",
  "claude-sonnet": "claude-sonnet-4-6",
  fable: "claude-fable-5",
  haiku: "claude-haiku-4-5",
  opus: "claude-opus-4-8",
  sonnet: "claude-sonnet-4-6"
};

// src/model-pricing.ts
var CACHE_WRITE_1H_MULTIPLIER = 2;
function resolveModelCostRates(model) {
  if (!model)
    return null;
  const key = model.trim().toLowerCase().replace(/\[[^\]]*\]\s*$/, "");
  if (!key)
    return null;
  const exact = MODEL_COST_RATES[key];
  if (exact)
    return exact;
  const aliased = MODEL_ALIASES[key];
  if (aliased && MODEL_COST_RATES[aliased])
    return MODEL_COST_RATES[aliased];
  for (const [alias, id] of Object.entries(MODEL_ALIASES)) {
    if (key.includes(alias) && MODEL_COST_RATES[id])
      return MODEL_COST_RATES[id];
  }
  return null;
}
function computeCostUsd(model, usage) {
  const rates = resolveModelCostRates(model);
  if (!rates || !usage)
    return null;
  const input = usage.input_tokens ?? 0;
  const output = usage.output_tokens ?? 0;
  const cacheRead = usage.cache_read_input_tokens ?? 0;
  const detail = usage.cache_creation;
  const hasSplit = detail != null && (detail.ephemeral_5m_input_tokens != null || detail.ephemeral_1h_input_tokens != null);
  const cacheWrite5m = hasSplit ? detail.ephemeral_5m_input_tokens ?? 0 : usage.cache_creation_input_tokens ?? 0;
  const cacheWrite1h = hasSplit ? detail.ephemeral_1h_input_tokens ?? 0 : 0;
  const microUsd = input * rates.input + output * rates.output + cacheRead * rates.cacheRead + cacheWrite5m * rates.cacheWrite5m + cacheWrite1h * rates.input * CACHE_WRITE_1H_MULTIPLIER;
  return microUsd / 1e6;
}
function resolveModelLimits(model) {
  if (!model)
    return null;
  const key = model.trim().toLowerCase().replace(/\[[^\]]*\]\s*$/, "");
  if (MODEL_LIMITS[key])
    return MODEL_LIMITS[key];
  const aliased = MODEL_ALIASES[key];
  if (aliased && MODEL_LIMITS[aliased])
    return MODEL_LIMITS[aliased];
  for (const [alias, id] of Object.entries(MODEL_ALIASES)) {
    if (key.includes(alias) && MODEL_LIMITS[id])
      return MODEL_LIMITS[id];
  }
  return null;
}
function buildModelUsage(model, usage) {
  if (!model || !usage)
    return null;
  const limits = resolveModelLimits(model);
  return {
    [model]: {
      inputTokens: usage.input_tokens ?? 0,
      outputTokens: usage.output_tokens ?? 0,
      cacheReadInputTokens: usage.cache_read_input_tokens ?? 0,
      cacheCreationInputTokens: usage.cache_creation_input_tokens ?? 0,
      webSearchRequests: usage.server_tool_use?.web_search_requests ?? 0,
      costUSD: computeCostUsd(model, usage) ?? 0,
      contextWindow: limits?.contextWindow ?? null,
      maxOutputTokens: limits?.maxOutputTokens ?? null
    }
  };
}

// src/claude-compat.ts
function asRecord(value) {
  return value && typeof value === "object" && !Array.isArray(value) ? value : undefined;
}
function firstString(...values) {
  for (const value of values)
    if (typeof value === "string" && value)
      return value;
  return;
}
function modelFromAssistantRow(row) {
  const model = asRecord(row.message)?.model;
  return typeof model === "string" ? model : undefined;
}
function stopReasonFromAssistantRow(row) {
  const reason = asRecord(row.message)?.stop_reason;
  return typeof reason === "string" ? reason : undefined;
}
function isToolResultUserRow(row) {
  const content = asRecord(row.message)?.content;
  return Array.isArray(content) && content.some((part) => asRecord(part)?.type === "tool_result");
}
function makeClaudeInitEvent(opts) {
  return {
    type: "system",
    subtype: "init",
    ...opts.cwd ? { cwd: opts.cwd } : {},
    ...opts.sessionId ? { session_id: opts.sessionId } : {},
    tools: [],
    mcp_servers: [],
    slash_commands: [],
    agents: [],
    skills: [],
    plugins: [],
    ...opts.model ? { model: opts.model } : {},
    permissionMode: opts.permissionMode ?? "default",
    apiKeySource: opts.apiKeySource ?? null,
    output_style: opts.outputStyle ?? "default",
    memory_paths: opts.memoryPaths ?? {},
    fast_mode_state: "off",
    ...opts.version ? { claude_code_version: opts.version } : {},
    ...opts.uuid ? { uuid: opts.uuid } : {}
  };
}
function reshapeAssistantEvent(row, sessionId) {
  const session = firstString(row.sessionId, sessionId);
  const requestId = firstString(row.requestId);
  const uuid = firstString(row.uuid);
  return {
    type: "assistant",
    message: row.message ?? {},
    parent_tool_use_id: firstString(row.toolUseID) ?? null,
    ...session ? { session_id: session } : {},
    ...uuid ? { uuid } : {},
    ...requestId ? { request_id: requestId } : {},
    ...row.isApiErrorMessage === true ? {
      isApiErrorMessage: true,
      ...row.error !== undefined ? { error: row.error } : {},
      ...typeof row.api_error_status === "number" ? { api_error_status: row.api_error_status } : {}
    } : {}
  };
}
function reshapeUserEvent(row, sessionId) {
  if (!isToolResultUserRow(row))
    return null;
  const session = firstString(row.sessionId, sessionId);
  const uuid = firstString(row.uuid);
  return {
    type: "user",
    message: row.message ?? {},
    parent_tool_use_id: firstString(row.toolUseID) ?? null,
    ...session ? { session_id: session } : {},
    ...uuid ? { uuid } : {}
  };
}
var SURFACED_SYSTEM_SUBTYPES = new Set(["api_error", "model_refusal_fallback"]);
var SYSTEM_EVENT_FIELDS = [
  "error",
  "cause",
  "retryAttempt",
  "retryInMs",
  "maxRetries",
  "level",
  "slug",
  "direction",
  "content",
  "trigger",
  "originalModel",
  "fallbackModel",
  "apiRefusalCategory",
  "apiRefusalExplanation"
];
function reshapeSystemEvent(row, sessionId) {
  const subtype = row.subtype;
  if (typeof subtype !== "string" || !SURFACED_SYSTEM_SUBTYPES.has(subtype))
    return null;
  const event = { type: "system", subtype };
  for (const key of SYSTEM_EVENT_FIELDS) {
    if (key in row)
      event[key] = row[key];
  }
  const session = firstString(row.sessionId, sessionId);
  if (session)
    event.session_id = session;
  const uuid = firstString(row.uuid);
  if (uuid)
    event.uuid = uuid;
  return event;
}
function transcriptRowToClaudeStreamEvent(row, sessionId) {
  if (row.type === "assistant")
    return reshapeAssistantEvent(row, sessionId);
  if (row.type === "user")
    return reshapeUserEvent(row, sessionId);
  if (row.type === "system")
    return reshapeSystemEvent(row, sessionId);
  return null;
}
function apiErrorStatusFromRow(row) {
  const error = row.error;
  const status = error && typeof error === "object" ? error.status : undefined;
  return typeof status === "number" ? status : null;
}
function makeClaudeResultEvent(opts) {
  const cost = computeCostUsd(opts.model, opts.usage);
  const modelUsage = buildModelUsage(opts.model, opts.usage);
  return {
    type: "result",
    subtype: "success",
    is_error: false,
    api_error_status: null,
    duration_ms: opts.durationMs ?? null,
    duration_api_ms: null,
    ttft_ms: null,
    ttft_stream_ms: null,
    time_to_request_ms: null,
    num_turns: opts.numTurns ?? 1,
    result: opts.resultText,
    stop_reason: opts.stopReason ?? null,
    session_id: opts.sessionId ?? null,
    total_cost_usd: cost,
    usage: opts.usage ?? null,
    modelUsage,
    permission_denials: [],
    terminal_reason: "completed",
    fast_mode_state: "off",
    uuid: opts.uuid ?? null,
    ...opts.structuredOutput?.present ? { structured_output: opts.structuredOutput.value } : {}
  };
}
function makeClaudeErrorResultEvent(opts) {
  return {
    type: "result",
    subtype: opts.subtype ?? "error_during_execution",
    is_error: true,
    api_error_status: opts.apiErrorStatus ?? null,
    duration_ms: opts.durationMs ?? null,
    duration_api_ms: null,
    ttft_ms: null,
    ttft_stream_ms: null,
    time_to_request_ms: null,
    num_turns: 1,
    result: null,
    error: opts.message,
    stop_reason: null,
    session_id: opts.sessionId ?? null,
    total_cost_usd: computeCostUsd(opts.model, opts.usage),
    usage: opts.usage ?? null,
    modelUsage: buildModelUsage(opts.model, opts.usage),
    permission_denials: [],
    terminal_reason: "error",
    fast_mode_state: "off",
    uuid: opts.uuid ?? null,
    ...opts.extras ?? {}
  };
}

// src/cli.ts
var HERE = dirname6(fileURLToPath(import.meta.url));
var REPO = resolve5(HERE, "..");
var parsedArgs = parseCliArgs(process.argv.slice(2));
if (!parsedArgs.ok) {
  process.stderr.write(`claude-bridge: ${parsedArgs.error.message}
`);
  process.stderr.write(`${parsedArgs.error.hint}
`);
  process.stderr.write("Run `bun ./src/cli.ts --help` for wrapper usage.\n");
  process.exit(2);
}
var args = parsedArgs.parsed;
if (args.help) {
  process.stdout.write(formatWrapperHelp());
  process.exit(0);
}
if (args.claudeHelp) {
  const claudeHelp = spawnSync("claude", ["-h"], { encoding: "utf8" });
  const helpText = (claudeHelp.stdout ?? "") + (claudeHelp.stderr ?? "");
  process.stdout.write(formatClaudeHelp(helpText));
  process.exit(0);
}
if (args.version) {
  process.stdout.write(formatVersionInfo());
  process.exit(0);
}
var desplegaArgs = args.desplegaArgs;
var desplegaVerbose = args.desplegaVerbose;
var printMode = args.print;
var outputFormat = args.outputFormat;
var desplegaFormat = hasDesplegaFlag("format");
if (hasDesplegaFlag("internal-json-schema-stop-hook")) {
  await runJsonSchemaStopHook();
  process.exit(0);
}
if (hasDesplegaFlag("internal-runtime-hook")) {
  await runRuntimeHook();
  process.exit(0);
}
var targetCwd = resolveTargetCwd();
var transcriptsBefore = await listAllTranscriptPaths();
var initialMessage = args.initialPrompt;
if (printMode && !initialMessage && !process.stdin.isTTY) {
  initialMessage = (await new Response(Bun.stdin.stream()).text()).trimEnd();
}
if (printMode && !initialMessage) {
  process.stderr.write(`claude-bridge: -p/--print requires a prompt argument or piped stdin.
`);
  process.stderr.write("Run `bun ./src/cli.ts --help` for wrapper usage.\n");
  process.exit(2);
}
var jsonSchema = loadPrintJsonSchema();
var forwardedClaudeArgs = jsonSchema ? mergeAppendSystemPrompt(args.claudeArgs, makeJsonSchemaSystemPrompt(jsonSchema.compact)) : args.claudeArgs;
var resumedSessionId = claudeResumeSessionId(forwardedClaudeArgs);
var forkedSession = Boolean(resumedSessionId && forwardedClaudeArgs.includes("--fork-session"));
var resumedTranscriptPathBeforeLaunch = null;
var resumedTranscriptBaselineRows = null;
if (resumedSessionId && !forkedSession) {
  try {
    resumedTranscriptPathBeforeLaunch = await findTranscriptBySessionId(resumedSessionId);
    if (resumedTranscriptPathBeforeLaunch) {
      resumedTranscriptBaselineRows = (await readTranscript(resumedTranscriptPathBeforeLaunch)).length;
    }
  } catch {}
}
var runId = new Date().toISOString().replace(/[:.]/g, "-") + "-" + randomUUID3().slice(0, 8);
var runDir = join6(resolveRunRoot(targetCwd), runId);
mkdirSync4(runDir, { recursive: true });
var traceAPI = process.env.MAGPIE_CLAUDE_TUI_DEBUG_API === "1";
if (traceAPI)
  chmodSync(runDir, 448);
var jsonSchemaPath = jsonSchema ? join6(runDir, "json-schema.json") : undefined;
var claudeExitStatusPath = join6(runDir, "claude-exit-status");
var sessionName = `claude-bridge-${runId.slice(-8)}`;
var CLAUDE_READY_TIMEOUT_MS = envDurationMs("CLAUDE_BRIDGE_CLAUDE_READY_TIMEOUT_MS", 180000);
var TMUX_SUBMIT_DELAY_MS = envDurationMs("CLAUDE_BRIDGE_TMUX_SUBMIT_DELAY_MS", 1000);
var TMUX_EXIT_HOLD_MS = envDurationMs("CLAUDE_BRIDGE_TMUX_EXIT_HOLD_MS", 30000);
var MAX_SESSION_MS = envDurationMs("CLAUDE_BRIDGE_MAX_SESSION_MS", 12 * 60 * 60000);
var TURN_END_FLUSH_TIMEOUT_MS = envDurationMs("CLAUDE_BRIDGE_TURN_END_FLUSH_TIMEOUT_MS", 5000);
var stopEventPath = join6(runDir, "stop-event.json");
var messageDisplayPath = join6(runDir, "message-display.jsonl");
var transcriptEventPath = join6(runDir, "transcript-event.json");
var localSettings = join6(targetCwd, ".claude", "settings.local.json");
if (jsonSchemaPath && jsonSchema) {
  writeFileSync3(jsonSchemaPath, jsonSchema.compact + `
`);
  installJsonSchemaStopHook({ settingsPath: localSettings, command: buildJsonSchemaStopHookCommand() });
}
var userContextPath = process.env.CLAUDE_BRIDGE_USER_CONTEXT_PATH;
installRuntimeHooks({ settingsPath: localSettings, command: buildRuntimeHookCommand() });
if (userContextPath) {
  installUserContextHooks({ settingsPath: localSettings, command: buildRuntimeHookCommand() });
  writeFileSync3(userContextPath + ".ready", `1
`, { mode: 384 });
}
if (tmuxSocket && !process.env.CLAUDE_CONFIG_DIR) {
  failBeforeRun("magpie runs the Claude TUI only on an account's own config folder (CLAUDE_CONFIG_DIR)");
}
preAcceptProject({ workdir: targetCwd, mcpServerNames: [] });
var runningAsRoot = typeof process.getuid === "function" && process.getuid() === 0;
var bypassPermissions = !runningAsRoot;
writeWorkdirSettings(targetCwd, { bypassPermissions });
var tmuxPath = process.env.CLAUDE_BRIDGE_TMUX_PATH || resolveRequiredCommand("tmux", "Required command `tmux` was not found on PATH. Install tmux on this host before running claude-bridge.");
var claudePath = process.env.CLAUDE_BRIDGE_CLAUDE_PATH || resolveRequiredCommand("claude", "Required command `claude` was not found on PATH. Install Claude Code on this host, or make sure PATH includes it before running claude-bridge.");
var tmuxSocket = process.env.CLAUDE_BRIDGE_TMUX_SOCKET || "";
var tmuxBase = tmuxSocket ? ["-S", tmuxSocket, "-f", "/dev/null"] : [];
var tmuxServer = null;
var ctrl = new AbortController;
var isTTY = !!process.stdout.isTTY && !printMode;
var showRaw = desplegaVerbose;
var inputPrompt = isTTY ? "> " : "";
var debugEvents = [];
var transcriptSessionId;
var lastAssistantText = "";
var lastAssistantRow = null;
var lastTurnDurationMs;
var pendingPrintResultText = "";
var pendingPrintFailure = null;
var nativeApiErrorTurn = new NativeApiErrorTurn;
var pendingNativeApiFailure = null;
var printDone = false;
var messageDisplayOffset = 0;
var claudeInitEmitted = false;
var claudeStreamEmittedCount = 0;
var claudeVersionFromTranscript;
var transcriptProjectFolder;
var lastApiErrorStatus = null;
var rlRef = null;
var stdoutLine = (obj) => {
  if (printMode) {
    if (outputFormat === "stream-json" && desplegaFormat) {
      process.stdout.write(JSON.stringify(obj) + `
`);
    }
    return;
  }
  let text;
  if (isTTY) {
    const line = formatEnvelope(obj, { showRaw });
    if (!line)
      return;
    text = line + `
`;
  } else {
    text = JSON.stringify(obj) + `
`;
  }
  if (rlRef && isTTY) {
    process.stdout.write("\r\x1B[2K" + text);
    rlRef.prompt(true);
  } else {
    process.stdout.write(text);
  }
};
var desplegaDebug = (message, data) => {
  if (!desplegaVerbose)
    return;
  const event = { type: "desplega_debug", message, ...data ? { data } : {} };
  if (printMode && outputFormat === "json" && desplegaFormat) {
    debugEvents.push(event);
    return;
  }
  if (printMode && outputFormat === "stream-json" && desplegaFormat) {
    stdoutLine(event);
    return;
  }
  if (printMode) {
    const suffix = data ? ` ${JSON.stringify(data)}` : "";
    process.stderr.write(`[desplega] ${message}${suffix}
`);
    return;
  }
  if (isTTY) {
    const suffix = data ? ` ${JSON.stringify(data)}` : "";
    process.stderr.write(`[desplega] ${message}${suffix}
`);
    return;
  }
  stdoutLine(event);
};
desplegaDebug("parsed arguments", {
  claudeArgs: forwardedClaudeArgs,
  desplegaArgs,
  jsonSchema: jsonSchema ? { estimatedTokens: jsonSchema.estimatedTokens, maxTokens: jsonSchema.maxTokens } : undefined
});
function formatVersionInfo() {
  const pkg = JSON.parse(readFileSync7(join6(REPO, "package.json"), "utf8"));
  const whichClaude = spawnSync("which", ["claude"], { encoding: "utf8" });
  const claudePath = whichClaude.status === 0 ? whichClaude.stdout.trim() : whichClaude.stderr.trim() || "not found";
  const claudeVersion = whichClaude.status === 0 ? commandOutput("claude", ["-v"]) || "unknown" : "unavailable";
  return [
    `${pkg.name ?? "claude-bridge"} ${pkg.version ?? "0.0.0"}`,
    `claude path: ${claudePath}`,
    `claude version: ${claudeVersion}`
  ].join(`
`) + `
`;
}
function commandOutput(command, args) {
  const res = spawnSync(command, args, { encoding: "utf8" });
  return `${res.stdout ?? ""}${res.stderr ?? ""}`.trim();
}
function envDurationMs(name, fallback) {
  const raw = process.env[name];
  if (!raw)
    return fallback;
  const value = Number(raw);
  if (!Number.isFinite(value) || value <= 0)
    return fallback;
  return Math.floor(value);
}
function loadPrintJsonSchema() {
  if (!args.jsonSchema)
    return null;
  try {
    const maxTokens = resolveJsonSchemaMaxTokens(desplegaArgs);
    return loadJsonSchema(args.jsonSchema, targetCwd, maxTokens);
  } catch (err) {
    failBeforeRun(err.message);
  }
}
function resolveTargetCwd() {
  const value = desplegaValue("cwd");
  const cwd = value === undefined || value === true ? process.cwd() : String(value);
  const resolved = resolve5(process.cwd(), cwd);
  if (!existsSync5(resolved) || !statSync2(resolved).isDirectory()) {
    process.stderr.write(`claude-bridge: --desplega-cwd must point to an existing directory: ${resolved}
`);
    process.exit(2);
  }
  return resolved;
}
function resolveRunRoot(cwd) {
  const value = desplegaValue("run-dir") ?? desplegaValue("state-dir");
  if (value !== undefined && value !== true)
    return resolve5(process.cwd(), String(value));
  return join6(cwd, ".claude-bridge", "runs");
}
function hasDesplegaFlag(name) {
  const arg = desplegaArgs.find((arg) => arg.name === name);
  return Boolean(arg && arg.value !== false);
}
function desplegaValue(name) {
  return desplegaArgs.find((arg) => arg.name === name)?.value;
}
function failBeforeRun(message) {
  if (printMode && outputFormat !== "text" && desplegaFormat) {
    process.stdout.write(JSON.stringify({ type: "result", subtype: "error", is_error: true, error: message }) + `
`);
  } else {
    process.stderr.write(`claude-bridge: ${message}
`);
  }
  process.exit(2);
}
function commandPath(command) {
  const res = spawnSync("which", [command], { encoding: "utf8" });
  if (res.status !== 0)
    return null;
  const first = res.stdout.trim().split(/\r?\n/)[0];
  return first || null;
}
function resolveRequiredCommand(command, message) {
  return commandPath(command) ?? failBeforeRun(message);
}
var claudeReady = false;
var pendingMessages = initialMessage ? [initialMessage] : [];
function sendUserMessage(text) {
  if (!claudeReady) {
    pendingMessages.push(text);
    desplegaDebug("queued message before Claude readiness", {
      length: text.length,
      claudeReady
    });
    return null;
  }
  const id = randomUUID3().slice(0, 8);
  const sent = sendPromptToTmux(text, id);
  if (!sent)
    return null;
  stdoutLine({ type: "push", id, content: text, transport: "tmux" });
  return id;
}
function sendPromptToTmux(text, id) {
  const bufferName = `claude-bridge-${id}`;
  const load = spawnSync(tmuxPath, [...tmuxBase, "load-buffer", "-b", bufferName, "-"], {
    input: text,
    encoding: "utf8"
  });
  if (load.status !== 0) {
    return failTmuxSend("load prompt buffer", load);
  }
  const paste = spawnSync(tmuxPath, [...tmuxBase, "paste-buffer", "-d", "-p", "-r", "-b", bufferName, "-t", sessionName], {
    encoding: "utf8"
  });
  if (paste.status !== 0) {
    return failTmuxSend("paste prompt buffer", paste);
  }
  const ready = waitForPendingPrompt(TMUX_SUBMIT_DELAY_MS + 9000);
  if (!ready) {
    desplegaDebug("prompt did not become visible before submit timeout", {
      id,
      length: text.length
    });
  }
  const enter = spawnSync(tmuxPath, [...tmuxBase, "send-keys", "-t", sessionName, "Enter"], {
    encoding: "utf8"
  });
  if (enter.status !== 0) {
    return failTmuxSend("submit prompt", enter);
  }
  for (const delayMs of [600, 1200, 2000]) {
    sleepSync(delayMs);
    if (!paneHasPendingInput(capturePane()))
      break;
    const retry = spawnSync(tmuxPath, [...tmuxBase, "send-keys", "-t", sessionName, "Enter"], {
      encoding: "utf8"
    });
    if (retry.status !== 0) {
      return failTmuxSend("retry prompt submission", retry);
    }
    desplegaDebug("retried prompt submission through tmux", {
      id,
      length: text.length,
      delayMs
    });
  }
  desplegaDebug("sent prompt through tmux", { id, length: text.length, ready });
  return true;
}
function waitForPendingPrompt(timeoutMs) {
  const deadline = Date.now() + timeoutMs;
  do {
    if (paneHasPendingInput(capturePane()))
      return true;
    sleepSync(200);
  } while (Date.now() < deadline);
  return false;
}
function paneHasPendingInput(pane) {
  const prompts = pane.split(`
`).filter((line) => /^\s*(?:\u276F|>)\s*/.test(line));
  if (prompts.length === 0)
    return false;
  return /^\s*(?:\u276F|>)\s*\S/.test(prompts[prompts.length - 1]);
}
function failTmuxSend(action, result) {
  const message = `Failed to ${action} with tmux.`;
  const stderr = String(result.stderr ?? "").trim();
  desplegaDebug("tmux send failure", { action, status: result.status, stderr });
  if (printMode) {
    failPrint(stderr ? `${message} ${stderr}` : message);
  } else {
    stdoutLine({
      type: "bridge_error",
      where: "tmux-send",
      message: stderr ? `${message} ${stderr}` : message
    });
  }
  return false;
}
function flushPendingMessages() {
  let sent = false;
  while (pendingMessages.length > 0) {
    const text = pendingMessages.shift();
    if (!text)
      continue;
    const id = sendUserMessage(text);
    if (!id)
      break;
    sent = true;
  }
  return sent;
}
function promptIfReady() {
  if (!printMode && claudeReady && isTTY)
    rlRef?.prompt();
}
function sleepSync(ms) {
  const sab = new SharedArrayBuffer(4);
  const int32 = new Int32Array(sab);
  Atomics.wait(int32, 0, 0, ms);
}
function finishPrintResult(resultText) {
  if (printDone)
    return;
  const structured = extractStructuredOutput(resultText);
  if (structured && !structured.ok) {
    failPrint(structured.message, { rawResponse: structured.rawResponse });
    return;
  }
  printDone = true;
  if (outputFormat === "text") {
    const text = structured ? JSON.stringify(structured.value) : resultText;
    process.stdout.write(text.endsWith(`
`) ? text : text + `
`);
  } else if (desplegaFormat) {
    const result = {
      type: "result",
      subtype: "success",
      is_error: false,
      result: resultText,
      ...structured ? { structured_output: structured.value, structured_output_source: structured.source } : {},
      ...transcriptSessionId ? { session_id: transcriptSessionId } : {},
      ...outputFormat === "json" && debugEvents.length ? { debug: debugEvents } : {}
    };
    process.stdout.write(JSON.stringify(result) + `
`);
  } else if (outputFormat === "json") {
    process.stdout.write(JSON.stringify(makeClaudeCompatResult(resultText, structured)) + `
`);
  }
  shutdown();
}
function makeClaudeCompatResult(resultText, structured) {
  return makeClaudeResultEvent({
    resultText,
    sessionId: transcriptSessionId,
    durationMs: lastTurnDurationMs,
    model: currentModel(),
    usage: lastAssistantUsage(),
    stopReason: lastAssistantRow ? stopReasonFromAssistantRow(lastAssistantRow) : undefined,
    structuredOutput: structured ? { present: true, value: structured.value } : undefined,
    uuid: randomUUID3()
  });
}
function currentModel() {
  return lastAssistantRow ? modelFromAssistantRow(lastAssistantRow) : undefined;
}
function lastAssistantUsage() {
  const msg = lastAssistantRow?.message ?? undefined;
  return msg?.usage;
}
function extractStructuredOutput(text) {
  if (!jsonSchema)
    return null;
  return extractAndValidateStructuredOutput(text, jsonSchema.schema);
}
function failPrint(message, details = {}) {
  if (printDone)
    return;
  printDone = true;
  const failureCapture = captureFailurePane();
  const messageWithPaneTail = appendPaneTail(message, failureCapture.paneTail);
  if (outputFormat === "text") {
    process.stderr.write(`claude-bridge: ${messageWithPaneTail}
`);
    if (details.rawResponse !== undefined) {
      process.stderr.write(`Raw Claude reply:
${details.rawResponse}
`);
    }
  } else if (desplegaFormat) {
    let stderrLog;
    if (printMode) {
      try {
        stderrLog = readFileSync7(join6(runDir, "stdout.stderr.log"), "utf8").trim() || undefined;
      } catch {}
    }
    const result = makePrintErrorResult(messageWithPaneTail, {
      rawResponse: details.rawResponse,
      sessionId: transcriptSessionId,
      runState: runDir,
      paneTail: failureCapture.paneTail,
      stderrLog,
      debug: outputFormat === "json" ? debugEvents : undefined
    });
    process.stdout.write(JSON.stringify(result) + `
`);
  } else {
    if (outputFormat === "stream-json")
      ensureClaudeInit();
    process.stdout.write(JSON.stringify(makeClaudeErrorResultEvent({
      message: messageWithPaneTail,
      apiErrorStatus: lastApiErrorStatus,
      sessionId: transcriptSessionId,
      durationMs: lastTurnDurationMs,
      model: currentModel(),
      usage: lastAssistantUsage(),
      uuid: randomUUID3(),
      extras: {
        run_state: runDir,
        ...details.rawResponse !== undefined ? { raw_response: details.rawResponse } : {},
        ...failureCapture.paneTail ? { pane_tail: failureCapture.paneTail } : {}
      }
    })) + `
`);
  }
  shutdown(1);
}
function startTmux() {
  const claudeArgs = [
    ...bypassPermissions ? ["--dangerously-skip-permissions"] : [],
    ...traceAPI ? ["--debug", "api", "--debug-file", join6(runDir, "claude-api.log")] : [],
    ...forwardedClaudeArgs
  ];
  const envArgs = [
    ...interactiveMcpBackgroundEnvArgs(),
    ...nativeAutoCompactionLaunchEnv(process.env).envArgs,
    ...claudeAuthEnvArgs(process.env, { localAuth: hasDesplegaFlag("local-auth") }),
    "CLAUDE_BRIDGE_RUNTIME_HOOK=1",
    ...process.env.CLAUDE_BRIDGE_MCP_SERVER ? [`CLAUDE_BRIDGE_MCP_SERVER=${process.env.CLAUDE_BRIDGE_MCP_SERVER}`] : [],
    `CLAUDE_BRIDGE_RUN_DIR=${runDir}`,
    ...userContextPath ? [`CLAUDE_BRIDGE_USER_CONTEXT_PATH=${userContextPath}`] : [],
    ...jsonSchemaPath ? [
      "CLAUDE_BRIDGE_SCHEMA_STOP_HOOK=1",
      `CLAUDE_BRIDGE_JSON_SCHEMA_PATH=${jsonSchemaPath}`,
      `CLAUDE_BRIDGE_RUN_ID=${runId}`
    ] : []
  ];
  const launchCommand = buildClaudeLaunchCommand({
    claudePath,
    claudeArgs,
    unsetEnvArgs: [...claudeUnsetEnvArgs(), ...nativeAutoCompactionLaunchEnv(process.env).unsetEnvArgs],
    envArgs,
    exitStatusPath: claudeExitStatusPath,
    holdMs: TMUX_EXIT_HOLD_MS
  });
  const nativeGuardPath = process.env.MAGPIE_CLAUDE_NATIVE_TMUX_GUARD_PATH;
  if (nativeGuardPath) {
    try {
      writeNativeTmuxGuard(nativeGuardPath, sessionName, tmuxSocket);
    } catch (error) {
      process.stderr.write(`[claude-bridge] cannot establish private native tmux ownership marker: ${String(error)}
`);
      process.exit(1);
    }
  }
  desplegaDebug("starting claude in tmux", {
    sessionName,
    claudePath,
    claudeArgs,
    bypassPermissions
  });
  const newSession = [
    "new-session",
    "-d",
    "-s",
    sessionName,
    "-x",
    "220",
    "-y",
    "50",
    "-c",
    targetCwd,
    launchCommand
  ];
  const res = startPrivateTmux(newSession);
  if (res.status !== 0) {
    process.stderr.write(`[claude-bridge] tmux new-session exited ${res.status}
`);
    process.exit(1);
  }
  autoAcceptStartupPrompts();
  waitForClaudeReady();
  startTranscriptTail();
  if (printMode)
    waitForPrintTurnEnd();
  if (!printMode)
    printBanner();
}
function startPrivateTmux(newSession) {
  if (!tmuxSocket)
    return spawnSync(tmuxPath, newSession, { stdio: "inherit" });
  tmuxServer = spawn(tmuxPath, [...tmuxBase, "-D"], { stdio: "ignore" });
  tmuxServer.on("exit", () => {
    tmuxServer = null;
  });
  let res = spawnSync(tmuxPath, [...tmuxBase, ...newSession], { stdio: "pipe" });
  for (let i = 0;res.status !== 0 && i < 100 && tmuxServer; i++) {
    Bun.sleepSync(50);
    res = spawnSync(tmuxPath, [...tmuxBase, ...newSession], { stdio: "pipe" });
  }
  if (res.status !== 0 && res.stderr)
    process.stderr.write(res.stderr);
  return res;
}
async function waitForPrintTurnEnd() {
  const POLL_MS = 100;
  const startedAt = Date.now();
  while (!ctrl.signal.aborted && Date.now() - startedAt < MAX_SESSION_MS) {
    emitMessageDisplayEvents();
    if (failCompletedNativeApiTurn())
      return;
    const stopEvent = readRuntimeStopEvent();
    if (stopEvent) {
      emitMessageDisplayEvents();
      if (stopEvent.hook_event_name === "Stop" && stopEvent.transcript_path) {
        const flushed = await waitForTranscriptTurnEnd(stopEvent.transcript_path, TURN_END_FLUSH_TIMEOUT_MS, ctrl.signal, currentTurnTranscriptRows);
        if (!flushed) {
          failPrint("Current Claude turn did not flush its terminal transcript row before the deadline.");
          return;
        }
        emitMessageDisplayEvents();
      }
      await hydrateTranscriptFromStopEvent(stopEvent);
      if (failCompletedNativeApiTurn())
        return;
      if (stopEvent.hook_event_name === "StopFailure") {
        failPrint(runtimeStopFailureMessage(stopEvent), { rawResponse: JSON.stringify(stopEvent) });
        return;
      }
      if (pendingPrintFailure) {
        failPrint(pendingPrintFailure.message, { rawResponse: pendingPrintFailure.rawResponse });
        return;
      }
      const stopAssistantText = typeof stopEvent.last_assistant_message === "string" && stopEvent.last_assistant_message !== "" ? stopEvent.last_assistant_message : undefined;
      const resultText = stopAssistantText ?? pendingPrintResultText ?? lastAssistantText;
      if (resultText === undefined || resultText === "") {
        failPrint("Claude Stop hook fired without assistant text.");
        return;
      }
      if (outputFormat === "stream-json" && !desplegaFormat) {
        if (!await emitClaudeStreamResult(stopEvent, resultText))
          return;
      }
      finishPrintResult(resultText);
      return;
    }
    const exitStatus = readClaudeExitStatus();
    if (exitStatus !== null && exitStatus !== "0") {
      failPrint(`Claude exited with status ${exitStatus} before the Stop hook produced a result.`);
      return;
    }
    if (!tmuxSessionExists()) {
      failPrint("tmux session/server died before the Stop hook produced a result.");
      return;
    }
    await sleep2(POLL_MS);
  }
  if (!printDone) {
    const elapsed = Math.round((Date.now() - startedAt) / 1000);
    failPrint(`Session safety ceiling reached after ${elapsed}s waiting for the Stop hook event.`);
  }
}
function failCompletedNativeApiTurn() {
  const failure = pendingNativeApiFailure;
  if (!failure)
    return false;
  if (failure.status)
    lastApiErrorStatus = failure.status;
  failPrint(`Claude Code API error: ${failure.message}`);
  return true;
}
function readRuntimeStopEvent() {
  if (!existsSync5(stopEventPath))
    return null;
  try {
    const event = JSON.parse(readFileSync7(stopEventPath, "utf8"));
    return event.hook_event_name === "Stop" || event.hook_event_name === "StopFailure" ? event : null;
  } catch {
    return null;
  }
}
function runtimeStopFailureMessage(stopEvent) {
  for (const field of [stopEvent.error, stopEvent.reason, stopEvent.message]) {
    if (typeof field === "string" && field.trim()) {
      return `Claude StopFailure hook fired: ${field.trim()}`;
    }
  }
  return "Claude StopFailure hook fired before the Stop hook produced a result.";
}
function currentTurnTranscriptRows(rows) {
  if (!forkedSession && resumedSessionId) {
    if (resumedTranscriptBaselineRows !== null) {
      return transcriptRowsAfterBaseline(rows, resumedTranscriptBaselineRows);
    }
    if (!initialMessage)
      return null;
    return transcriptRowsFromLastUserPrompt(rows, initialMessage);
  }
  if (!forkedSession)
    return rows;
  if (!initialMessage)
    return null;
  return transcriptRowsFromLastUserPrompt(rows, initialMessage);
}
async function hydrateTranscriptFromStopEvent(stopEvent) {
  const transcriptPath = stopEvent.transcript_path;
  if (!transcriptPath)
    return;
  transcriptSessionId = sessionIdFromPath(transcriptPath);
  try {
    const allRows = await readTranscript(transcriptPath);
    const rows = currentTurnTranscriptRows(allRows);
    if (!rows) {
      pendingPrintFailure = {
        message: "Fork transcript did not contain the submitted prompt; refusing inherited-history replay."
      };
      return;
    }
    rows.forEach((row, index) => handleTranscriptRow(row, JSON.stringify(row)));
    desplegaDebug("hydrated final transcript from Stop hook", {
      transcriptPath,
      rows: rows.length,
      inheritedRowsSkipped: allRows.length - rows.length
    });
  } catch (err) {
    desplegaDebug("failed to hydrate transcript from Stop hook", {
      transcriptPath,
      error: err.message
    });
  }
}
function ensureClaudeInit(model) {
  if (claudeInitEmitted)
    return;
  claudeInitEmitted = true;
  process.stdout.write(JSON.stringify(makeClaudeInitEvent({
    sessionId: transcriptSessionId,
    cwd: targetCwd,
    model: model ?? currentModel(),
    version: claudeVersionFromTranscript,
    uuid: randomUUID3(),
    permissionMode: bypassPermissions ? "bypassPermissions" : "default",
    apiKeySource: hasDesplegaFlag("local-auth") ? "ANTHROPIC_API_KEY" : "none",
    memoryPaths: transcriptProjectFolder ? { auto: join6(transcriptProjectFolder, "memory") + "/" } : {}
  })) + `
`);
}
function writeClaudeStreamEvent(event) {
  process.stdout.write(JSON.stringify(event) + `
`);
  claudeStreamEmittedCount++;
}
function emitClaudeStreamRow(row) {
  const event = transcriptRowToClaudeStreamEvent(row, transcriptSessionId);
  if (!event)
    return;
  ensureClaudeInit(row.type === "assistant" ? modelFromAssistantRow(row) : undefined);
  writeClaudeStreamEvent(event);
}
function emitClaudeStreamCatchUp(rows) {
  const events = rows.map((row) => transcriptRowToClaudeStreamEvent(row, transcriptSessionId)).filter((event) => event !== null);
  if (events.length <= claudeStreamEmittedCount)
    return;
  const firstAssistant = rows.find((row) => row.type === "assistant");
  ensureClaudeInit(firstAssistant ? modelFromAssistantRow(firstAssistant) : undefined);
  for (const event of events.slice(claudeStreamEmittedCount))
    writeClaudeStreamEvent(event);
}
async function emitClaudeStreamResult(stopEvent, resultText) {
  const structured = extractStructuredOutput(resultText);
  if (structured && !structured.ok) {
    failPrint(structured.message, { rawResponse: structured.rawResponse });
    return false;
  }
  if (stopEvent.transcript_path) {
    try {
      const allRows = await readTranscript(stopEvent.transcript_path);
      const rows = currentTurnTranscriptRows(allRows);
      if (!rows) {
        failPrint("Fork transcript did not contain the submitted prompt; refusing inherited-history replay.");
        return false;
      }
      emitClaudeStreamCatchUp(rows);
    } catch (err) {
      desplegaDebug("failed to drain transcript for stream-json result", {
        error: err.message
      });
    }
  }
  ensureClaudeInit();
  await new Promise((resolve) => process.stdout.write(JSON.stringify(makeClaudeResultEvent({
    resultText,
    sessionId: transcriptSessionId,
    durationMs: lastTurnDurationMs,
    model: currentModel(),
    usage: lastAssistantUsage(),
    stopReason: lastAssistantRow ? stopReasonFromAssistantRow(lastAssistantRow) : undefined,
    structuredOutput: structured ? { present: true, value: structured.value } : undefined,
    uuid: randomUUID3()
  })) + `
`, () => resolve()));
  shutdown();
  return true;
}
function emitMessageDisplayEvents() {
  if (!existsSync5(messageDisplayPath))
    return;
  let text = "";
  try {
    text = readFileSync7(messageDisplayPath, "utf8");
  } catch {
    return;
  }
  if (text.length <= messageDisplayOffset)
    return;
  const chunk = text.slice(messageDisplayOffset);
  messageDisplayOffset = text.length;
  for (const line of chunk.split(`
`)) {
    const trimmed = line.trim();
    if (!trimmed)
      continue;
    let row;
    try {
      row = JSON.parse(trimmed);
    } catch {
      continue;
    }
    const delta = typeof row.delta === "string" ? row.delta : "";
    if (outputFormat === "stream-json" && !desplegaFormat) {
      continue;
    } else {
      stdoutLine({ type: "message_display", delta });
    }
  }
}
function capturePane() {
  const r = spawnSync(tmuxPath, [...tmuxBase, "capture-pane", "-pt", sessionName, "-S", "-80"], {
    encoding: "utf8"
  });
  return (r.stdout ?? "") + (r.stderr ?? "");
}
function captureFailurePane() {
  paneCaptured = true;
  const path = join6(runDir, "tmux-pane-final.txt");
  const r = spawnSync(tmuxPath, [...tmuxBase, "capture-pane", "-pt", sessionName, "-S", "-"], {
    encoding: "utf8"
  });
  const pane = ((r.stdout ?? "") + (r.stderr ?? "")).trimEnd();
  try {
    writeFileSync3(path, pane ? pane + `
` : "");
  } catch (err) {
    desplegaDebug("failed to write final tmux pane capture", {
      path,
      error: err.message
    });
  }
  let paneTail = tailLines(pane, 40);
  if (printMode && !paneTail) {
    const stderrFile = join6(runDir, "stdout.stderr.log");
    try {
      const stderr = readFileSync7(stderrFile, "utf8").trim();
      if (stderr)
        paneTail = `[stderr.log] ${tailLines(stderr, 40)}`;
    } catch {}
  }
  if (paneTail) {
    desplegaDebug("diagnostics before failure", { path, tail: paneTail });
  }
  return { pane, paneTail, path };
}
function appendPaneTail(message, paneTail) {
  return paneTail ? `${message}
Pane tail:
${paneTail}` : message;
}
function tailLines(text, count) {
  return text.trim().split(`
`).filter(Boolean).slice(-count).join(`
`);
}
var sleep2 = (ms) => new Promise((r) => setTimeout(r, ms));
var THEME_DIALOG_RE = /(Choose the text style that looks best with your terminal|To change this later, run \/theme)/i;
var SECURITY_NOTES_DIALOG_RE = /(Security notes:|Security guide|Press Enter to continue|Enter to confirm)/i;
var CUSTOM_API_KEY_DIALOG_RE = /(Detected a custom API key in your environment|Do you want to use this API key\?)/i;
var TRUST_DIALOG_RE = /(Quick safety check|Is this a project you created|Yes, I trust this folder)/i;
async function autoAcceptStartupPrompts() {
  const t0 = Date.now();
  let lastFire = 0;
  let lastPaneDebug = 0;
  let fires = 0;
  while (Date.now() - t0 < 45000 && !ctrl.signal.aborted) {
    const pane = capturePane();
    if (isClaudePaneReady(pane)) {
      if (fires > 0 && !printMode)
        process.stderr.write(`[claude-bridge] startup prompts cleared
`);
      return;
    }
    if (desplegaVerbose && Date.now() - lastPaneDebug > 2000) {
      const tail = pane.split(`
`).slice(-12).join(`
`);
      desplegaDebug("pane tail", { tail });
      lastPaneDebug = Date.now();
    }
    const prompt = startupPromptFromPane(pane);
    if (prompt && Date.now() - lastFire > 600) {
      if (prompt === "custom-api-key") {
        spawnSync(tmuxPath, [...tmuxBase, "send-keys", "-t", sessionName, "Up", "Enter"]);
      } else if (prompt === "trust") {
        const keys = trustPromptKeys(pane);
        if (keys.length === 0) {
          await sleep2(150);
          continue;
        }
        spawnSync(tmuxPath, [...tmuxBase, "send-keys", "-t", sessionName, ...keys]);
      } else {
        spawnSync(tmuxPath, [...tmuxBase, "send-keys", "-t", sessionName, "Enter"]);
      }
      lastFire = Date.now();
      fires++;
      if (!printMode) {
        process.stderr.write(`[claude-bridge] accepted Claude ${prompt} prompt (#${fires})
`);
      }
    }
    await sleep2(150);
  }
  if (fires === 0 && !printMode) {
    process.stderr.write(`[claude-bridge] startup prompts not detected within 45s; attach to the tmux pane if it's stuck
`);
  }
}
function startupPromptFromPane(pane) {
  if (hasDesplegaFlag("local-auth") && CUSTOM_API_KEY_DIALOG_RE.test(pane)) {
    return "custom-api-key";
  }
  if (TRUST_DIALOG_RE.test(pane))
    return "trust";
  if (THEME_DIALOG_RE.test(pane))
    return "theme";
  if (SECURITY_NOTES_DIALOG_RE.test(pane))
    return "security-notes";
  return null;
}
var MCP_LISTED_TIMEOUT_MS = envDurationMs("CLAUDE_BRIDGE_MCP_LISTED_TIMEOUT_MS", 30000);
async function waitForMcpListed() {
  const listedPath = process.env.CLAUDE_BRIDGE_MCP_LISTED_PATH;
  if (!listedPath)
    return;
  const startedAt = Date.now();
  while (!existsSync5(listedPath)) {
    if (ctrl.signal.aborted)
      return;
    if (Date.now() - startedAt >= MCP_LISTED_TIMEOUT_MS) {
      desplegaDebug("MCP tools not listed before timeout; submitting anyway", { waitedMs: Date.now() - startedAt });
      return;
    }
    await sleep2(100);
  }
  await sleep2(500);
  desplegaDebug("MCP tools listed", { waitedMs: Date.now() - startedAt });
}
async function waitForClaudeReady() {
  const startedAt = Date.now();
  while (Date.now() - startedAt < CLAUDE_READY_TIMEOUT_MS && !ctrl.signal.aborted) {
    const pane = capturePane();
    const exitStatus = readClaudeExitStatus();
    if (exitStatus !== null) {
      failClaudeStartup(exitStatus, pane);
      return;
    }
    if (!tmuxSessionExists()) {
      failClaudeStartup(null, pane);
      return;
    }
    if (isClaudePaneReady(pane)) {
      await waitForMcpListed();
      if (ctrl.signal.aborted)
        return;
      claudeReady = true;
      desplegaDebug("claude pane ready");
      if (!flushPendingMessages())
        promptIfReady();
      return;
    }
    await sleep2(150);
  }
  if (ctrl.signal.aborted)
    return;
  const message = `Timed out after ${CLAUDE_READY_TIMEOUT_MS / 1000}s waiting for Claude to become ready.`;
  if (printMode) {
    failPrint(message);
  } else {
    stdoutLine({ type: "bridge_warning", message });
  }
}
function readClaudeExitStatus() {
  if (!existsSync5(claudeExitStatusPath))
    return null;
  return readFileSync7(claudeExitStatusPath, "utf8").trim() || "unknown";
}
function tmuxSessionExists() {
  const res = spawnSync(tmuxPath, [...tmuxBase, "has-session", "-t", sessionName], {
    stdio: "ignore"
  });
  return res.status === 0;
}
function failClaudeStartup(exitStatus, pane) {
  const statusText = exitStatus === null ? "the tmux session exited" : `Claude exited with status ${exitStatus}`;
  const message = `${statusText} before Claude became ready. Run state: ${runDir}. Check \`${claudePath} -v\` and Claude auth on this host.`;
  if (printMode) {
    failPrint(message);
  } else {
    const failureCapture = captureFailurePane();
    const paneTail = failureCapture.paneTail || tailLines(pane, 40);
    stdoutLine({
      type: "bridge_error",
      where: "claude-startup",
      message,
      run_state: runDir,
      pane_tail: paneTail || undefined
    });
    shutdown(1);
  }
}
function isClaudePaneReady(pane) {
  return claudePaneReady(pane);
}
async function startTranscriptTail() {
  const compatStreamJsonPrint = printMode && outputFormat === "stream-json" && !desplegaFormat;
  let transcriptPath;
  try {
    transcriptPath = (!forkedSession && resumedSessionId ? resumedTranscriptPathBeforeLaunch ?? await findTranscriptBySessionId(resumedSessionId) : null) ?? (compatStreamJsonPrint && !forkedSession ? await waitForRuntimeTranscriptPath(1500) : null) ?? await waitForFreshTranscriptForCwd(targetCwd, transcriptsBefore, ctrl.signal);
  } catch (err) {
    if (printMode) {
      desplegaDebug("live transcript discovery failed; falling back to Stop hook result", {
        error: err.message
      });
      return;
    }
    stdoutLine({
      type: "bridge_error",
      where: "transcript-discovery",
      message: err.message
    });
    return;
  }
  const projectFolder = projectFolderFromTranscript(transcriptPath);
  transcriptProjectFolder = projectFolder;
  stdoutLine({ type: "transcript_folder", path: projectFolder });
  const sessionId = sessionIdFromPath(transcriptPath);
  transcriptSessionId = sessionId;
  stdoutLine({ type: "transcript_open", path: transcriptPath, session_id: sessionId });
  if (compatStreamJsonPrint) {
    await tailTranscriptLines(transcriptPath, (line) => {
      let row;
      try {
        row = JSON.parse(line);
      } catch {
        row = { type: "bridge_parse_error", line };
      }
      handleTranscriptRow(row, line);
      emitClaudeStreamRow(row);
    }, ctrl.signal, Boolean(resumedSessionId));
    return;
  }
  await tailTranscript(transcriptPath, (row, _index, rawLine) => handleTranscriptRow(row, rawLine), ctrl.signal);
}
async function waitForRuntimeTranscriptPath(timeoutMs) {
  const startedAt = Date.now();
  while (!ctrl.signal.aborted && Date.now() - startedAt < timeoutMs) {
    const path = readRuntimeTranscriptPath();
    if (path)
      return path;
    await sleep2(50);
  }
  return null;
}
function readRuntimeTranscriptPath() {
  if (!existsSync5(transcriptEventPath))
    return null;
  try {
    const event = JSON.parse(readFileSync7(transcriptEventPath, "utf8"));
    return typeof event.transcript_path === "string" && event.transcript_path ? event.transcript_path : null;
  } catch {
    return null;
  }
}
function handleTranscriptRow(row, _rawLine) {
  if (printMode && printDone)
    return;
  stdoutLine({ type: "transcript", row });
  if (!printMode)
    return;
  const nativeFailure = nativeApiErrorTurn.observe(row);
  if (nativeFailure)
    pendingNativeApiFailure = nativeFailure;
  if (typeof row.version === "string" && row.version)
    claudeVersionFromTranscript = row.version;
  if (row.type === "system" && row.subtype === "api_error") {
    const status = apiErrorStatusFromRow(row);
    if (status != null)
      lastApiErrorStatus = status;
  }
  const type = String(row.type ?? "");
  if (type === "assistant") {
    const msg = row.message ?? {};
    const text = textFromContent2(msg.content).trim();
    if (text) {
      lastAssistantText = text;
      lastAssistantRow = row;
    }
    return;
  }
  if (isTurnDurationRow(row)) {
    if (typeof row.durationMs === "number")
      lastTurnDurationMs = row.durationMs;
    if (!lastAssistantText) {
      pendingPrintFailure = {
        message: "Claude reached turn end without assistant text.",
        rawResponse: lastAssistantText
      };
      return;
    }
    pendingPrintResultText = lastAssistantText;
    return;
  }
}
function textFromContent2(content) {
  if (typeof content === "string")
    return content;
  if (!Array.isArray(content))
    return "";
  return content.map((part) => {
    if (!part || typeof part !== "object")
      return "";
    const block = part;
    return block.type === "text" ? String(block.text ?? "") : "";
  }).filter(Boolean).join(`
`);
}
function printBanner() {
  process.stderr.write([
    "",
    `   tmux session : ${sessionName}`,
    `   cwd          : ${targetCwd}`,
    `   run state    : ${runDir}`,
    "",
    `   attach to the Claude UI in another terminal:`,
    `     tmux attach -t ${sessionName}`,
    "",
    `   Type a message + Enter on stdin to send it to Claude.`,
    `   Assistant and useful transcript rows print below.`,
    `   Use --desplega-verbose for raw transcript rows and wrapper debug.`,
    `   Ctrl-D to quit (kills the tmux session).`,
    ""
  ].join(`
`));
}
function startRepl() {
  const rl = createInterface({
    input: process.stdin,
    output: process.stdout,
    terminal: isTTY,
    prompt: inputPrompt
  });
  rlRef = rl;
  if (isTTY)
    process.stdout.write(`
`);
  rl.on("line", (line) => {
    const text = line.trim();
    if (!text) {
      promptIfReady();
      return;
    }
    sendUserMessage(text);
  });
  rl.on("close", () => shutdown());
}
var shuttingDown = false;
var paneCaptured = false;
function shutdown(exitCode = 0) {
  if (shuttingDown)
    return;
  shuttingDown = true;
  ctrl.abort();
  if (!paneCaptured && exitCode !== 0) {
    captureFailurePane();
  }
  spawnSync(tmuxPath, [...tmuxBase, tmuxSocket ? "kill-server" : "kill-session", ...tmuxSocket ? [] : ["-t", sessionName]], { stdio: "ignore" });
  tmuxServer?.kill("SIGKILL");
  process.exit(exitCode);
}
desplegaDebug("run state created", { runDir });
process.on("SIGINT", () => shutdown());
process.on("SIGTERM", () => shutdown());
startTmux();
if (!printMode)
  startRepl();
