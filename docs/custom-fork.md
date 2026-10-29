# BB-fat custom fork (`custom` branch)

This fork tracks [yetone/magpie](https://github.com/yetone/magpie) from the
latest upstream release tag on branch `custom`, with BB-fat-specific changes
merged on top.

## Self-update (disabled in code)

**Built-in self-update is turned off in this fork.** `internal/update/fork.go`
sets `CustomFork = true`, which makes `update.SelfUpdateDisabled()` true in
production builds:

- No requests to `usemagpie.ai` / **yetone/magpie-releases** for updates
  (`Latest`, `Stage`, `ReplaceBinary`, … return `update.ErrSelfUpdateDisabled`).
- The app's background update timer never runs (`updateDue` is always false).
- `magpie update` and `magpie update check` print `update.DisabledMessage()`
  and do not download.
- `magpie update auto on` is rejected; saving settings forces
  `noAutoUpdate` true so the GUI cannot re-enable official updates.

Newer custom builds are built locally ([Building](#building)).

Tests may set `MAGPIE_UPDATE_ALLOW_OFFICIAL=1` to exercise the upstream
update path against a httptest feed; release builds do not set that.

## Claude subscription MCP server name

Enterprise gateways may block the default MCP server name `magpie`. Set
`MAGPIE_CLAUDE_MCP_NAME` to a permitted name before starting magpie; see
[Claude subscription bridge](subsystems/claude-subscription-bridge.md).

## Claude subscription idle pool (memory)

On small VPS hosts, idle Claude Code processes kept for the next turn can use
a lot of RAM. Optional env vars (read once at gateway startup; invalid values
stop magpie with an error):

| Variable | Default (unset) | Meaning |
| --- | --- | --- |
| `MAGPIE_CLAUDE_IDLE_MAX` | `6` | Max idle Claude Code processes kept in memory (`keepIdle`). **`0` is allowed:** new idle runs are shelved immediately (session on disk, no live process). |
| `MAGPIE_CLAUDE_IDLE_TTL` | `1h` | How long an idle run may stay alive before its timer ends it. Independent of prompt-cache TTL (`claudeCacheTTL`); shortening this does not change cache marks. |

Example for a 1GB box: `MAGPIE_CLAUDE_IDLE_MAX=1` and/or
`MAGPIE_CLAUDE_IDLE_TTL=10m`.

## Claude TUI safeguards

Interactive mode supports serialized safeguards extra-body JSON up to 64 KiB.
Larger bodies are rejected explicitly before launch, without truncation or
switching the account to standard mode. See
[Claude TUI bridge](subsystems/claude-tui-bridge.md#safeguards-size-limit).

## Claude TUI accounts (Linux)

A Claude subscription account can be answered by the interactive Claude
Code in a private tmux instead of the stream-json bridge, ported from
claude-tui-boost v0.2.16 into the gateway's TUI runtime, with its embedded
wrapper under `internal/claudetui/wrapper`; see
[Claude TUI bridge](subsystems/claude-tui-bridge.md). Every account is
`standard` until set otherwise:

```sh
magpie claude-code mode                       # each account: set, and in effect
magpie claude-code mode me@example.com tui    # from the gateway's next start
magpie claude-code mode me@example.com standard
```

or the account row's mode pill in the window
(`POST /api/providers/claude-mode`). Restart the gateway for a change to
take effect, once no tool call is under way. It needs tmux 3.2 or newer,
Claude Code 2.1.268 or newer, and a Bun: `MAGPIE_BUN`, or the one magpie
keeps for plugins (installing any plugin downloads it); nothing is
installed for it. The account must be signed in through magpie (its own
folder), not only be the account Claude Code itself is signed in to.

`MAGPIE_CLAUDE_MCP_NAME` names the MCP server on TUI runs too.
`MAGPIE_CLAUDE_IDLE_MAX` / `MAGPIE_CLAUDE_IDLE_TTL` apply to the standard
bridge alone: a TUI run has its own idle wait, `claudeTuiIdleSeconds`
(default 60s, 0 ends each run with its reply), over the wrapper's
`--desplega-serve` multi-turn protocol. `claudeTuiMaxRSSMiB` in
settings.json (this machine's own) bounds a TUI run's processes together,
by default as large as a standard run may grow. To turn it off: every
account back to `standard`, then restart.

Tested against the real Claude Code, tmux and Bun with a local stand-in for
Anthropic (`TestRealClaudeTUI`); not yet against a real subscription.

## Automation

- `.github/workflows/upstream-sync.yml` — daily (and manual) check for a new
  upstream release tag, merge into `custom` via a PR when tests pass.

No workflow builds or publishes this fork's binaries; a tag starts nothing.
Builds are made locally (below).

## Building

From a clean checkout of the tag `v0.1.<upstream>-custom.<n>` on `custom`,
the static Linux x86_64 CLI (what runs as `/usr/local/bin/magpie`):

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags nogui -trimpath \
  -ldflags="-s -w -X main.version=0.1.<upstream>-custom.<n>" -o magpie-cli-linux-amd64 .
```

Copy it to the host, check its sha256 there, keep the old binary as
`magpie.bak.<version>`, replace it and restart the service.
