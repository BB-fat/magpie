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

Install newer custom builds from
[BB-fat/magpie Releases](https://github.com/BB-fat/magpie/releases).

Tests may set `MAGPIE_UPDATE_ALLOW_OFFICIAL=1` to exercise the upstream
update path against a httptest feed; release builds do not set that.

## Claude subscription MCP server name

Enterprise gateways may block the default MCP server name `magpie`. Set
`MAGPIE_CLAUDE_MCP_NAME` to a permitted name before starting magpie; see
[Claude subscription bridge](subsystems/claude-subscription-bridge.md).

## Automation

- `.github/workflows/upstream-sync.yml` — daily (and manual) check for a new
  upstream release tag, merge into `custom` via a PR when tests pass.
- `.github/workflows/custom-release.yml` — on `v*` tag, publish two assets:
  `magpie-darwin-arm64.zip` (macOS Apple Silicon `magpie.app`, same as upstream)
  and `magpie-cli-linux-amd64` (static Linux x86_64 CLI).
