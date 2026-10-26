# Claude TUI review fixes: verification

Reviewed base: `8cc75e40b4e22a8c19d6c067e58b5e9b7c3c1b26` (PR #5).
Fixes are on `fix/claude-tui-review`, in a separate worktree. The original
`main` checkout was left unchanged. These results were collected on Linux
on 2026-10-26; no real account credentials or vendor requests were used.

## Observable changes

- A Claude TUI used as the web search provider uses native WebSearch and
  returns its sources. Search availability and search-in-progress context
  cross the gateway/bridge boundary explicitly, preventing recursive search.
- A run waiting for client tool results contributes to `Busy.Tools` after
  its HTTP response ends. An active HTTP continuation is counted as a request
  instead; disconnect/resume and shutdown release the corresponding state.
- Root wrappers honor Claude Code's explicit sandbox capability when choosing
  permission bypass. Fixed stateless directories include the UID; all fixed
  TUI work directories reject symlinks, foreign owners and public write access.
- Signed thinking retains block boundaries and signatures, including empty
  text blocks, in streams, whole replies, Messages replay and rebuilt JSONL.
- Matching live tool results resolves the safeguards beta first, preserving
  the original PTY rather than starting a second run for unchanged safeguards.
- Completed-reply replay uses a complete semantic digest, including tool
  arguments, result status, media, signed thinking and safeguards. JSON key
  order, metadata and response streaming do not change that identity.
- Rebuilding pairs consecutive assistant/user role runs. Split parallel calls
  keep their results; a compacted orphan result becomes labelled historical
  output with its status and images. Visible mismatched call IDs still fail.
- The core script budget's historical comments are shortened; budget values
  are unchanged. The wrapper bundle was rebuilt from its locked dependencies.

The runtime contract and cache limitations are in
[Claude TUI bridge](subsystems/claude-tui-bridge.md).

## Passed checks

Go commands ran under temporary HOME directories with GOPATH, GOMODCACHE and
GOCACHE pinned before changing HOME, and GOPROXY=off. Process tests use local
HTTP services and fake processes unless explicitly named as real-CLI tests.

| Check | Result |
| --- | --- |
| Full `go test -tags nogui ./internal/claudetui -count=1 -timeout=10m` | Passed, 96.678s. The later full-repository run also passed this package, 98.538s. |
| Full TUI package `-race` | Passed, 102.598s. |
| Gateway `go test -race -tags nogui ./internal/gateway -run '^TestClaudeTUI' -count=1` | Passed, 4.035s. |
| Final signed-thinking rebuild, replay, tool conversion and compaction focused `-race` | TUI passed, 4.060s; gateway passed, 4.445s. |
| Search/busy/disconnect race regression, `-count=20` | TUI passed, 2.183s; gateway passed, 39.972s. |
| Claude mode provider, GUI route/settings persistence, CLI mode, script budget, complete MCP bridge package | Passed. |
| Wrapper `bun test` after rebuild | 60 passed, 0 failed, 244 expectations, two test files. |
| Chromium and WebKit browser interactions | 44 passed, 0 failed: mode selection, reload persistence, failed save, restoration to standard, account caps and Claude auth status. Mode tests cover en/zh/zh-TW/ja/de at 1100 and 420px. |
| `GOOS=linux go vet -tags nogui ./...` and `GOOS=windows go vet ./...` | Passed, including a repeat after the final source/test changes. |
| Linux nogui build and Windows build | Passed, outputs placed only in temporary storage. |
| `gofmt -l` for changed/new Go files and `git diff --check` | Passed with no file/error output. |
| Independent read-only final review | No additional actionable findings; 28 selected top-level tests passed. |

The seven real-CLI subtests use the installed Claude Code, Bun 1.4.4, a private
tmux and the built magpie MCP helper, connected to a local Anthropic stand-in:

```sh
MAGPIE_CLAUDE_TUI_REAL=1 \
MAGPIE_TUI_BUN=<absolute Bun> \
MAGPIE_TUI_CLAUDE=<absolute Claude Code> \
MAGPIE_TUI_MAGPIE=<built magpie> \
go test -tags nogui ./internal/claudetui -run '^TestRealClaudeTUI$' -count=1 -timeout=10m -v
```

All seven passed in 19.077s: default/custom MCP names with tool-result replay,
cancel, restart, cache prefix, killed gateway cleanup and memory termination.
The cache-prefix test compares nonempty MCP definitions, system instructions
apart from billing metadata, and the existing message content across a
checkpoint resume. It normalizes equivalent text representations and moving
cache marks. It does not measure the live vendor's cache hit rate.

## Tests fail before the fixes

New regression tests were run against original functions through temporary Go
overlays or isolated wrapper copies; shared source was never rolled back.
Observed failures included:

- Thinking blocks merged to `firstlast` with empty signatures; signed empty
  blocks and rebuilt JSONL thinking vanished.
- Old search code answered without searching or failed when no search API was
  configured; the wrapper lost `tool_use_result` sources.
- Old busy accounting reported `Tools:0` after the tool response ended.
- Foreign-owned fixed directories were accepted, and stateless names lacked
  the UID. Root sandbox wrapper cases omitted permission bypass.
- Changed tool arguments reused the prior completed reply; split parallel
  tool traffic and compacted orphan outputs disappeared from rebuilt JSONL.
- Moving safeguards parsing back after the live match made the four-API tool
  regression start a rebuilt run and return another tool call instead of DONE.

## Full repository test result and existing failures

`go test -tags nogui ./... -timeout=45m` ran to completion with a suite lock:
67 packages passed and six packages failed. This is an unsuccessful full-suite
result. The same failing behaviors were reproduced from a separate archive
of the exact original PR head, without applying these fixes:

| Package | Failing test or error | Original PR reproduction |
| --- | --- | --- |
| `internal/agent` | `TestInstalledVersion`, `TestOmpDetectedWithPiDir`: fake pi/omp on a PATH containing only the temporary bin directory are not detected | Same failures. |
| `internal/cli` | `TestKeyNote`: expects ordinary LAN wording while this host reports the container-specific banner | Same failure and banner. |
| `internal/gateway` | `TestCodexFittedImagesPastTheBudget`: 19 of 30 images with 10 kept; `TestMarshalRawIsEncodingJSON` and `FuzzMarshalRaw`: U+2028/U+2029 escaping differs; `TestServeWSLListensBesideLoopback`: IPv6 loopback still answers | All same failures. |
| `internal/gui` | `TestUpdateNotWritableSaysWhy`: returns `Stuck:container` where the test expects an unwritable binary directory | Same failure. |
| `internal/provider` | `TestUnreadLoginsWriteErrorsKeepCredentials`: `fatal error: close of synctest channel from outside bubble`, through `filememo.Forget` | Same fatal error and path. |
| `internal/usage` | `TestUsageReadAfterAppendHashesNothingOnAFineClock/whole-second_stamp`: zero bytes hashed when the test expects hashing | Reproduced under `-race -count=20` on both the original PR and fixed tree. Its truncated stamp and real `time.Now()+150ms` cross the one-second settle threshold. |

These existing failures are recorded here and remain unresolved; this change
is not a merge-ready full-suite pass. No issue was posted externally.

## Verification limits

- A native Linux GUI `go vet ./...` and `go build` were attempted and failed
  because GTK4, WebKitGTK 6.0, GLib/GIO, libsoup and fontconfig development
  packages are absent. Browser verification used isolated Chromium/WebKit
  runtimes; the desktop's native WebKitGTK was not exercised.
- Windows was cross-built and vetted; no Windows or macOS runtime test ran.
- No real subscription, paid upstream, live cache-hit measurement or native
  root Claude process was used. Root permission/ownership behavior has fake
  CLI and actual-root directory regression coverage.
- No full GUI suite was run: this fix changes no shared UI asset behavior;
  the new mode and neighboring account flows have the browser coverage above.
- The fixes are prepared as a follow-up commit for PR #5. The existing pending
  GitHub review remains available for the maintainer to inspect and submit.
