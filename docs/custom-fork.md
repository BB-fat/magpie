# BB-fat custom fork (`custom` branch)

This fork tracks [yetone/magpie](https://github.com/yetone/magpie) from the
latest upstream release tag on branch `custom`, with BB-fat-specific changes
merged on top.

## Self-update and official releases

magpie's built-in update checks `https://usemagpie.ai/api/latest`, which
describes **yetone/magpie-releases** builds (`internal/update/update.go`,
`Feed()`). A custom GitHub Release from **BB-fat/magpie** does not replace
that feed unless you point magpie at one.

Custom release binaries are built with a version like
`0.1.1295+custom.N` (build metadata after `+` does not affect semver
ordering). The official `0.1.1296` release is still **newer** than
`0.1.1295+custom.N`, so an unchanged install may download and replace the
custom binary with the official one.

To keep a custom build:

1. Turn off automatic updates: `magpie update auto off` (or Settings ›
   General › Updates), **or**
2. Point the feed elsewhere for testing or a private mirror:
   `MAGPIE_UPDATE_FEED=<url>` (`internal/update/update.go`), **or**
3. Install only from [BB-fat/magpie Releases](https://github.com/BB-fat/magpie/releases)
   and avoid “Check for updates” in the app.

There is no separate “fork channel” in upstream magpie; preventing overwrite
is an install/update policy choice, not a compile-time switch in yetone/magpie
v0.1.1295.

## Claude subscription MCP server name

Enterprise gateways may block the default MCP server name `magpie`. Set
`MAGPIE_CLAUDE_MCP_NAME` to a permitted name before starting magpie; see
[Claude subscription bridge](subsystems/claude-subscription-bridge.md).

## Automation

- `.github/workflows/upstream-sync.yml` — daily (and manual) check for a new
  upstream release tag, merge into `custom` via a PR when tests pass.
- `.github/workflows/custom-release.yml` — build and publish a GitHub Release
  when a `v*` tag is pushed on `custom`.
