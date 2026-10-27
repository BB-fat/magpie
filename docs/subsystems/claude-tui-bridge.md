# Claude TUI bridge: a Claude account answered by the interactive Claude Code

A Claude subscription account set to `tui` is answered by the genuine
interactive Claude Code, run in a private tmux on this machine, instead
of the [standard bridge](claude-subscription-bridge.md) to Claude Code's
stream-json. Its provider, model IDs, routing and allowance are the same
either way. Linux only, and off by default: every account is `standard`
until it is set otherwise.

The bridge is a port of
[claude-tui-boost](https://github.com/Guuuu233/claude-tui-boost) v0.2.16
(`f57ee457cf7f41ecc18ed8284519ab565f46ad16`, MIT), itself a fork of this
repository's standard bridge. The Go runtime lives in
`internal/gateway/claude_tui_runtime_*.go`; `internal/claudetui` carries
the embedded wrapper and its licences. Its licences are
kept: `internal/claudetui/LICENSE.claude-tui-boost`, and the wrapper's
`internal/claudetui/wrapper/LICENSE` (desplega.sh, MIT) with
`wrapper/ADAPTATIONS.md`. Not taken: the plugin's precompiled runtime and
helper binaries, its HTTP runtime (bridge-auth-v1 HMAC, nonce cache,
managed key, supervisor), its legacy-migration certificates, its WSL
path, and the `claude -p` start path (`subscriptionBridge.start`): a TUI
account is never answered by `claude -p`.

## Which accounts, and when

- `accountClaudeModes` on the Claude provider in `providers.json`, by
  account name in lower case, holds the accounts set to `tui`
  (`provider.ClaudeModeOf`, `normalAccountClaudeModes`; `standard` is the
  absence of an entry). It is copied, cloned and moved with an account's
  rename as `accountConcurrency` is.
- It is set with `magpie claude-code mode <account> [standard|tui]` or
  `POST /api/providers/claude-mode` `{id, account, mode}`
  (`provider.SetAccountClaudeMode`), and from the account row's mode pill
  in the window. Refused: an unknown account, a mode other than the two,
  a provider that isn't the Claude subscription, `tui` on a platform
  other than Linux (`provider.ClaudeTUIGOOS`), and the account Claude
  Code itself is signed in to in the user's home (`claudeOwnOnly`): run
  in a folder of magpie's, it would share that home's refresh token.
- The gateway reads the modes once, when it starts serving
  (`gateway.tuiModes`, `ClaudeTUILive`); a change takes effect at its next
  start. It writes what it read to `<config>/claude-tui/live.json` with its
  pid (`writeTUILive`), which the CLI and the providers' state
  (`claudeTuiNow`, `claudeTuiLive`) read to show the mode in effect beside
  the one set.

## Runtime path

`serveClaudeSubscription` hands an account set to `tui` to
`serveClaudeTUI` (`internal/gateway/claude_tui.go`), after its
`bridgeTurn` accounting, so a restart to update waits for TUI turns too.

1. `setUpClaudeTUI` (once per process, also run as the gateway starts
   serving when an account is set to `tui`): `tuiRuntimeSupported`
   (Linux), `claudebridge.LoadMCPConfig`, `initTUIRuntime` with the
   gateway's own (below), `tuiRuntimeCheck` for tmux ≥ 3.2, Claude Code ≥
   2.1.268 and Bun, `tuiRuntimeInstallWrapper`, `tuiRuntimeSweep`. What
   is missing is logged, and every TUI request answers 502 with it. Nothing
   is installed, downloaded or run with sudo.
2. A client cache-mark opt-in (`X-Magpie-Cache-Marks: client`) is refused
   400: the run's cache slots are Claude Code's own. The short system
   prompt setting (`claudeLeanPrompt`) does not apply.
3. The account's config folder (`tuiAccountDir`): its saved sign-in's,
   `provider` `claude-accounts/<hash>`, refreshed by magpie as for the
   standard bridge. Claude Code's own home is never run in or copied.
4. The request is parsed in its client's API (`parse`); a Messages
   request goes on as it came, the others as `buildAnthropic` renders them,
   without the stand-in tools it declares for a history's calls when none
   are offered (`tuiMessages`).
5. `tuiRuntimeBridge.Serve` with the owner `tuiRuntimeOwner(id, user,
   MCP server name)` (part of every checkpoint's identity) and the
   conversation from the request's session header (`sessionOf`;
   none: the request runs stateless). Its `serveLite` and
   `serveMessagesInner` answer through a responder: the gateway's
   (`tuiResponder`) writes refusals in the client's protocol and relays the
   events through the gateway's own `relay`, which counts the
   usage and says the account's allowance (`limitHeaders`). Native failures
   are classified before being emitted as shared events;
   an account usage limit is a 429 (`tuiRuntimeErrorStatus`).
6. A failure is answered as a failure: a TUI account is never answered by
   the standard bridge in its place. Routing goes on to a group's next
   member as for any failed try.

The runtime uses the gateway's `Request`, `Message`, `Part`, `Tool`,
`Event`, `Usage`, Claude CLI usage and allowance types, MCP tools and
results, Anthropic codec, bounded body reader and response relay directly.
Protocol changes are owned by the gateway implementation; the TUI keeps only its
native-session, checkpoint, recovery and terminal-process state.

TUI-only request state lives in `tuiRuntimeRequest`
(`internal/gateway/claude_tui_runtime_request.go`), not the gateway's
`Request`. It embeds the shared request by value, preserving the existing
copy-before-transform behavior for compaction, history seeding and resumed
branches. Replay fingerprints and injected search-tool names stay in this
wrapper too. Parsing still delegates to the shared Anthropic codec; shared
gateway helpers receive the embedded request. No protocol fields or codec
implementations are copied into the TUI wrapper.

The UI's additional script allowance is kept in
`internal/gui/claude_tui_budget_test.go`. The upstream budget constants and
their history remain unchanged; only the test's effective thresholds add
the TUI allowance. This keeps recurring upstream budget updates separate
from the fork's per-account mode control.

## What the bridge is given (`tuiRuntimeConfig`)

| Field | From the gateway |
| --- | --- |
| `ClaudePath` | `claudecode.Find` (this machine's, else magpie's download; never WSL) |
| `Wrapper` | the absolute Bun and the installed bundle |
| `Tmux` | the tmux `tuiRuntimeCheck` found |
| `HelperArgv` | magpie's own executable, `claude-mcp-helper` |
| `MCPServer` | `claudebridge.MCPServerName()` |
| `CallbackBase`, `AddCallback`, `DropCallback` | `tuiRuntimeCallbackBaseURL`, and `lanGuard`'s `ownCallback` tokens |
| `Room`, `Leave` | `claudeRoom.take`, `claudeRoom.leave` |
| `Env` | the account's proxy (`netproxy.EnvForBunWith`), `CLAUDE_CONFIG_DIR`, `IS_SANDBOX` as root |
| `MaxRSS` | `claudeTuiMaxRSSMiB`, else `claudeRunBudget` |
| `CanSearch`, `SearchHandler`, `NoteLimits`, `MovedOff` | the gateway's search availability and web search, `provider.NoteClaudeLimits`, `provider.ClaudeCodeMovedOff` |

## The MCP server name and tool names

`MAGPIE_CLAUDE_MCP_NAME` names the server on both bridges: the
`--mcp-config` key (`tuiRuntimeInteractiveMCPConfig`), the helper's
`serverInfo.name`, the tool prefix read back from the transcript and from
callbacks, the image tool the run is told to call, the calls written into
a rebuilt session (`tuiRuntimeWriteClaudeTranscript`), the calls in a text prompt
(`tuiRuntimeRenderParts`), and the hooks' filter (`CLAUDE_BRIDGE_MCP_SERVER`, set by
the bridge for the wrapper). Tool names go through the same encoding as
the standard bridge's, `claudebridge.ToolName`, and come back as declared
(`tuiRuntimeSubscriptionRun.declaredLocked`). The name is part of the owner, so a
conversation's checkpoint is not taken up under another name: its next
turn is rebuilt from the request.

The runtime reads the gateway's shared search-in-progress context.
A TUI run used as the search provider enables native `WebSearch`; an ordinary
run uses the gateway's private MCP search when `CanSearch` says one is
available, and native search otherwise. Native search calls stay inside Claude
Code; the wrapper preserves their `tool_use_result` query and hits for the
client's search events. This keeps a search-provider request from delegating
back to the gateway recursively.

Thinking blocks retain their signatures and boundaries (`KThinkStart`,
`KThink`, `KSig`) through transcript decoding and the shared gateway event
type, including signed blocks with empty text. Rebuilt native transcripts
preserve signed assistant thinking and opaque Anthropic `redacted_thinking`
blocks too (`tuiRuntimeAppendClaudeTranscript`). Messages clients can return
those blocks unchanged to an Anthropic provider.

## Safeguards size limit

Interactive runs pass safeguards through `CLAUDE_CODE_EXTRA_BODY` in the
process environment. The serialized extra-body JSON is limited to 64 KiB,
measured in bytes, not characters. Larger bodies are unsupported and rejected
with the actual size and limit before starting the wrapper or Claude. The
same guard covers first turns, history rebuilds, persisted resumes and live
safeguards updates. Safeguards are never truncated or silently omitted.

There is no Node launcher, preload, file-to-environment workaround or outbound
request rewriting layer. Settings-file and runtime environment injection can
avoid the initial exec limit but still break Claude's child-process launches.
Claude continues to construct and send its own upstream requests.

## The wrapper

`internal/claudetui/wrapper/src` is claude-bridge as claude-tui-boost
adapted it, with magpie's changes (see `ADAPTATIONS.md`). It is built with
its locked dependency (`zod` 4.6.6, `bun.lock`) into
`wrapper/dist/claude-bridge.js`, embedded (`tuiRuntimeWrapperBundle`) and written by
`tuiRuntimeInstallWrapper` to `<config>/claude-tui/wrapper/<content hash>/`, 0600,
whole or not at all. Rebuild it after a change to `wrapper/src`:

```sh
cd internal/claudetui/wrapper
$BUN install --ignore-scripts --frozen-lockfile
$BUN build ./src/cli.ts --target=bun --outfile=dist/claude-bridge.js
$BUN test
```

It runs on the Bun `plugin.BunPath` gives: `$MAGPIE_BUN` (1.1.0 or newer)
or the one magpie keeps for plugins, under its `magpie-bun` name or the
name it was downloaded under before, which is not renamed. It never
downloads one. Hooks run that Bun by its absolute path (`process.execPath`
in `buildInternalHookCommand`), so no `bun` on PATH is needed. Every hook
is written to the run's own `.claude/settings.local.json`; the user's
`~/.claude/settings.json` is never written (the wrapper's install and
uninstall flags are left out). The trust record goes to the account's
`.claude.json` (`CLAUDE_CONFIG_DIR`); under magpie the wrapper refuses to
start without one. A JSON Schema is passed in a file (one argument holds
128 KiB on Linux) and checked within bounds: schema 256 KiB, reply 4 MiB,
depth 64, 10,000 items in one list or object, 100,000 values, 8 issues
said (`JSON_LIMITS`). A root run uses permission bypass only when its
own environment supplies `IS_SANDBOX=1` or `CLAUDE_CODE_BUBBLEWRAP`; the
gateway supplies the former for root, as on the standard bridge. The same
choice controls both the launch flag and work-directory settings.

## Processes, room and memory

- Each run's wrapper starts its own tmux server in the foreground
  (`tmux -S <socket> -f /dev/null -D`), on a socket in
  `<tmp>/magpie-claude-tui-<uid>/<gateway pid>-<token>.sock`: never the
  user's tmux. The wrapper ends it with `kill-server`; the gateway ends it
  again by its socket once the wrapper is reaped (`tuiRuntimeKillTmux`).
- A run takes one place in `claudeRoom` before its wrapper starts and frees
  it once, after the whole tree has ended (the goroutine after
  `tree.Wait` in `startInteractiveWithPromptImages`). A place is held while
  the run waits on tool calls. A run that finds no room is answered 503
  (`tuiRuntimeRoomError`), as the standard bridge's is.
- `tuiRuntimeProcTree` notes the processes under the wrapper, each by its start
  time (tmux gives a pane a session of its own, so they aren't in the
  wrapper's group), every `tuiRuntimeRssEvery` (2s) and as the run is ended. Their
  resident memory together past `MaxRSS` ends the run with an error naming
  `claudeTuiMaxRSSMiB`; a reading that fails ends nothing. It is a
  reading, not a limit: systemd's `MemoryHigh`/`MemoryMax` are the hard
  ones, if wanted.
- Ending a run (`abort`): SIGINT to the wrapper, whose teardown ends its
  tmux; after `tuiRuntimeInteractiveAbortGrace` (5s), the group and every process
  noted under it are killed, by pid and start time, never by name. The
  wrapper gets SIGTERM if the gateway dies (`Pdeathsig`). A gateway coming
  up ends the tmux servers of gateways no longer running, by their
  sockets, and removes those runs' temp folders
  (`magpie-claude-pty-<gateway pid>-*`, their callback and tools files)
  (`tuiRuntimeSweep`).
- `MAGPIE_CLAUDE_IDLE_MAX` and `MAGPIE_CLAUDE_IDLE_TTL` are the standard
  bridge's idle pool alone: a TUI run ends with its reply, and its
  conversation's next turn resumes the checkpointed session.
- Between a tool-use reply and its results, `tuiRuntimeBridge.Parked` contributes to
  `tuiRuntimeServer.Busy().Tools`. An attached HTTP response stays in request/bridge
  accounting instead; keep-awake and idle restart consumers see the waiting
  tool chain without counting it twice.
- Stateless cwd slots are `<tmp>/magpie-claude-stateless-<uid>-<slot>`.
  `tuiRuntimeOwnedWorkDir` rejects symlinks, directories owned by another user, and
  directories writable by others before its write probe. The same check
  protects conversation and tmux directories. Older stateless paths are left
  in place and are no longer selected.

## Turns, recovery and its limits

Kept from the plugin (`claude_tui_runtime_claude_lite.go`): a conversation's turns wait for
each other (60s, then 503 with `Retry-After: 5`); tool results go on in
the live run, else the parked session copied up to the calls
(`liteParkedBranch`), else the conversation's checkpoint
(`restoreInteractiveSession`), else a session rebuilt from the request's
history (`tuiRuntimeLiteSeed`), with its tools. Tool results that don't answer the
visible reply's calls are refused 409. The very same tool-result request
after its reply is answered with that reply, with no new run
(`liteCompletedToolTurn`). A run parked on tool calls is let go after 5
minutes (`tuiRuntimeInteractiveParkLongest`), a run ends after 12 hours
(`tuiRuntimeInteractiveActiveLongest`). Checkpoints live in
`<config>/claude-tui/claude-interactive-sessions.json`; one that doesn't
read is refused, never taken for none, and those not used for 30 days go.
The safeguards beta is resolved before matching a live run, so unchanged
safeguards continue the same tool call.

A resume marks its checkpoint dirty before it starts. When it ends with no
answer and no tool call (its client went away, its start failed), the run,
once reaped, takes it back (`rewindCancelledResume`): the session's
transcript is cut to its length at the start, if it still begins with the
same bytes, and the checkpoint is clean again. The resend resumes the same
session, with its message in it once, instead of rebuilding the history.
A turn that answered, called a tool or had any MCP callback stays dirty.

A tool shown from its MCP callback before its transcript row waits up to
`tuiRuntimeCallbackRowWait` (2s) for that row, then `tuiRuntimeUnstreamedQuiet`
once a row of its message comes (at once when its row was read already).
The rows carry the message's usage and its parallel calls, so both stay
with the request that showed the call. A row that comes after that wait
is counted with the next request.

Claude Code's auto mode classifier is refused 400 with
`x-should-retry: false`, before any native run
(`tuiRuntimeAutoModeClassifier`: no tools, only user messages, the first
beginning `<transcript>\n`). Auto mode asks the API to review each action
and reads `safeguard_results`; a TUI run's native Claude Code keeps that
verdict (its transcript has none), so the client classifies each action
locally: its whole transcript in a request the PTY types as one prompt
block, written to the cache again every time (about 300k tokens a call on
a long session). Refused, Claude Code asks the user about the action, or
stops the turn after several in a row; nothing runs unreviewed. Auto mode
wants the account set to `standard`.

A completed reply is reused only for the same request semantics: canonical
tool arguments, result error flags, media contents, safeguards and generation
settings participate in its replay key. This strict replay identity is separate
from checkpoint matching, whose existing reply anchor permits a client to
rewrite or compact an earlier prefix.

Rebuilding pairs tool traffic across consecutive assistant and user messages,
so split parallel calls keep all their matched results. A result whose call was
removed by compaction is preserved as labelled historical output with its
error status and images, rather than emitted as an invalid orphan tool result.
Visible mismatched tool handoffs still fail with 409.

## Prompt-cache continuity

A stable supported session header lets a completed conversation resume its
native session and send only the new suffix. Without such a header, later
text turns run stateless and rebuild the supplied history; live tool results
can still find their run by call ID. Clients that need checkpoint continuity
should send `X-Magpie-Session` or an agent session header recognized by
`sessionOf`. Body metadata and `prompt_cache_key` are not checkpoint IDs.

The cwd remains stable across resumes and stateless slot reuse, auto memory
is disabled, and the bridge keeps Claude Code's 1-hour cache setting. A model,
tool definition, system-context or conversation change can still alter the
upstream cache prefix. Rebuilding preserves history for correctness and can
require a fresh cache write; changing between standard and TUI also rebuilds.
The gateway preserves reported cache reads, writes and 1-hour writes, and
counts repeated transcript rows of one upstream message once. These counters
and stable-prefix tests do not establish Anthropic's live cache hit rate.

A rebuilt session can ask for a tool call again: a write or a command may
be proposed a second time. Nothing here makes a side effect happen exactly
once; the tool journal and the reasons the log gives are what there is.
Claude Code's transcript is what every reply is read from: a run is
refused, saying why, when `CLAUDE_CODE_SKIP_PROMPT_HISTORY` is set or the
account's `settings.json` has `cleanupPeriodDays: 0` (`tuiRuntimeTranscriptOff`).

## Turning it off

Set every account back to `standard` (`magpie claude-code mode <account>
standard`) and restart the gateway. A conversation does not go on across
the two modes: its next turn on the standard bridge is told the whole
conversation.

## Verification

The review-fix checks and remaining full-suite failures are recorded in
[Claude TUI review verification](../claude-tui-review-verification.md).

```sh
go test -tags nogui ./internal/gateway/ -run '^TestClaudeTUI' -count=1
go test -race -tags nogui ./internal/gateway/ -run '^TestClaudeTUI' -count=1
go test -tags nogui ./internal/provider/ -run 'ClaudeMode' -count=1
go test -tags nogui ./internal/gui/ -run 'TestClaudeModeRoute|TestSettingsSaveKeepsWhatItDoesNotSend' -count=1
go test -tags nogui ./internal/cli/ -run 'TestClaudeCodeMode' -count=1
go test -tags nogui ./internal/plugin/ -run 'TestBunPath' -count=1
go test -tags nogui ./internal/claudebridge/ -count=1
(cd internal/claudetui/wrapper && $BUN test)
# the real Claude Code, tmux and Bun against a local stand-in for Anthropic
MAGPIE_CLAUDE_TUI_REAL=1 MAGPIE_TUI_BUN=<bun> MAGPIE_TUI_CLAUDE=<claude> \
  MAGPIE_TUI_MAGPIE=<a built magpie> go test -run '^TestClaudeTUIRuntimeRealClaudeTUI$' ./internal/gateway/
```

`TestClaudeTUIRuntimeRealClaudeTUI/cache_prefix` compares the real CLI's outgoing tools,
system instructions (apart from its billing metadata), and previous message
content before and after a checkpoint resume. Cache marks can move to a later
boundary; equivalent plain-text string and block-array forms are normalized.
The stand-in supplies no evidence of Anthropic's server-side cache hits.

`claude-mode.test.cjs` exercises the Providers editor's mode selection,
persistence across reload, failed saves and restoration to standard in
Chromium and WebKit, all five languages, at 1100 and 420 pixels. The neighboring
account-cap and Claude auth-status flows share its verification run.

Not verified: a real Claude subscription (the real test answers from a
local stand-in), live-vendor cache hit rates, or the desktop's native webview.
