# Claude TUI: gateway reuse and review verification

## Request-state isolation follow-up (2026-10-26)

The six TUI-only request fields now belong to `tuiRuntimeRequest`, which
embeds the shared `Request` by value. The public gateway IR file is restored
byte-for-byte to the PR base. TUI transformations still copy both protocol
values and runtime state before changing either; JSON encoding remains the
shared request's encoding, without the private state. The new boundary
regression failed before the change, naming all six fields on the shared
request. Copy independence and encoding have separate regressions.

TUI's 2,393-byte script allowance now lives in its own test file rather
than rewriting upstream budget constants and their historical comments.
Effective app/core script limits are unchanged. No UI asset, wrapper source,
dependency, sync workflow, or upstream branch is changed by this follow-up.

| Follow-up check | Result |
| --- | --- |
| Shared-request boundary, value-copy and encoding regressions, plus the GUI script budget | Passed. The boundary test failed on the previous code, naming all six TUI fields. |
| Complete TUI gateway tests with `-race -count=1` | Passed: 299 top-level tests, 6 conditional skips, 94.844s. |
| Boundary, copy, encoding and native compaction tests with `-race -count=20` | Passed: four tests repeated 20 times, 1.125s. |
| Linux nogui and Windows vet/build | Passed. |
| Full repository `go test -tags nogui -p 2 ./... -count=1 -timeout=25m` | 71 packages passed, including gateway; provider has the previously reproduced `filememo.Forget` synctest panic. Not a full-suite pass. |
| Public IR versus PR base `40892e79` | No diff; existing fork customizations predating this PR are preserved. |

An object-only merge preview against upstream release `v0.1.1610`
(`9809a8ac9a489152f521112136f9de5695da4b57`) now conflicts only in
`internal/update/update.go`, which already conflicted without the TUI PR.
The TUI-induced `internal/gui/core_budget_test.go` conflict is gone. This
preview neither updates branches nor establishes a merged build/runtime pass.

## Gateway reuse (PR #6, 2026-10-26)

The Go runtime and its regression tests now live in `internal/gateway`.
They use the existing gateway request, event, usage and MCP models,
Anthropic codec, bounded request reader, response relay, Claude CLI usage
decoder and provider allowance types. The copied protocol stack and
gateway/TUI event conversion are removed. `internal/claudetui` supplies
only the embedded wrapper and its sources and licences. Wrapper JavaScript,
dependencies, account modes and the standard bridge remain unchanged.

The shared parser preserves opaque Anthropic `redacted_thinking` blocks and
the `defer_loading` tool flag in the gateway model. Native transcript rebuilds
also preserve opaque thinking. The new sealed-thinking round-trip test failed
against the original TUI codec because the block disappeared. The native
transcript test also failed before its serializer handled the shared sealed
part. Both now pass. These are behavioral failures, not compile failures.

Native quota classification still happens before events enter the shared
relay, preserving the TUI's 429 response for synthetic quota errors. Tests
share the gateway's isolated `TestMain` rather than resetting configuration
after each test while a run is still finishing. Native keepalive and RSS
monitors stop before terminal cleanup, and lifecycle tests wait for cleanup;
their exclusivity assertion is scoped to both owner and outer session.

Checks use Go 1.26.8, temporary HOME directories, pinned Go caches, umask 022,
and a child subreaper for fake process tests. `TMPDIR=/var/tmp` avoids the
cloud platform's protected `.git` mounts in ordinary temporary directories.
The Pi fixture used by repository tests is pinned to the repository's version.

| Check | Result |
| --- | --- |
| `go test -race -tags nogui ./internal/gateway -run '^TestClaudeTUI' -count=1` | Passed: 296 top-level tests, 6 conditional skips, 97.854s. |
| Account/room and owner/lifecycle regression, `-race -count=20` | Passed: four top-level tests repeated 20 times, 15.845s. |
| `go test -tags nogui -p 2 ./... -count=1 -timeout=25m` | 71 packages passed; `internal/provider` failed with the existing synctest panic described below. The complete gateway package passed. |
| Linux `go vet -tags nogui -p 2 ./...` and Windows `go vet -p 2 ./...` | Passed. |
| Linux nogui and Windows builds | Passed. |
| Changed Go files' `gofmt -l` and `git diff --check` | Passed, no output. |

The provider package panics in `TestUnreadLoginsWriteErrorsKeepCredentials`:
`fatal error: close of synctest channel from outside bubble`, through
`internal/filememo/filememo.go:91`. Running the complete provider package
from a separate archive of the original PR head
`1818a6ec610fda2e97dd46cc643cd8d5abd0182b` reproduced the same fatal error.
The full-suite command is therefore unsuccessful; this PR does not fix that
unrelated failure or claim a merge-ready full-suite pass.

Not rerun: browser and wrapper tests, since their sources and bundle are
unchanged. Real Claude account requests and real native CLI scenarios were
not run: this environment has no Claude Code or tmux. The native desktop
build/vet still lack the GTK/WebKitGTK development libraries. Windows was
cross-built and vetted, not run on Windows.

## Original review verification

The remainder records the earlier review, before the gateway-reuse refactor.
Its old package paths and command results are historical, not current checks.

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
