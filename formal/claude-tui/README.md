# Claude TUI TLA+ Audit

Audit date: 2026-11-02, Asia/Shanghai.

Target: `c01714ae69df64537a96b6c8a4602954ffd28720`, including the activity
summary fix. This is not the old `custom` worktree or an assertion about
what is currently deployed on a reported deployment. No retained production implementation change, commit, push, or deployment
was made during this audit. The temporary observation hooks used by the
witness runner were restored byte-for-byte before the ordinary suite ran.

## Findings First

Four implementation problems have a TLC counterexample and a corresponding
Go safety assertion that fails on the audited source. Each Go witness was
run 20 times with the race detector; the failures are logical safety
violations, not Go data-race warnings. Two witnesses use observation hooks
to make a legal scheduling window deterministic. These hooks are described
below and are not production fixes.

### TUI-F01: Completed replies are not isolated by caller key [P1]

- Source: `internal/gateway/claude_tui_runtime_claude_lite.go`,
  `tuiRuntimeLiteReplayKey` at line 275, `liteCompletedToolTurn` at line 343,
  and the early completed-result lookup at line 146.
- `req.scope` contains the account and caller key, but it is an unexported
  field and is omitted by `json.Marshal(copied)`. The completed-result index
  is `hash(owner, outerSession)`, also without the caller key. Consequently,
  two different caller keys with the same account, outer session header,
  and semantic tool-result request hit the same entry.
- Trace: caller A executes request, stores reply A; caller B supplies the
  same body/session and receives reply A without any native inference.
- Witness: `TestTLAWitnessCompletedReplyCallerIsolation` sets an authenticated
  `access.Identity` for B and observes `REPLY_FOR_KEY_A` in the handler's
  response, with zero native starts. It uses no source observation hook.
- Impact: this violates the documented caller-isolation contract and can
  return another caller's generated reply. This is conditional on a shared
  account and matching session/body; it is not a demonstrated arbitrary
  transcript read, authentication bypass, or model/account ACL bypass.
- Suggested direction, not implemented: include the effective caller scope
  explicitly in both the replay digest and completed-result index, and
  validate all cached replies against that identity.
- Evidence: [TLC trace](results/replay-cross-caller.log),
  [Go witnesses](results/go-witnesses.log).

### TUI-F02: A stale session scan can authorize a second native writer [P1]

- Source: `internal/gateway/claude_tui_runtime_claude_interactive_points.go`,
  `claimSession` at line 34; its caller is `startLiteAtNativeCheckpoint` in
  `claude_tui_runtime_claude_lite.go` at line 655.
- The live-run lookup happens before acquiring `b.mu` for
  `claimedSessions`. After the first launch publishes its live run and
  releases its start reservation, a second launch that previously scanned
  an empty live-run list can acquire that reservation without rechecking
  the now-live writer.
- Trace: A scans; B scans; A reserves, sees the checkpoint row, publishes,
  releases its reservation; B reserves using its stale scan, sees the same
  row before Claude has appended its new turn, and starts a second writer.
- Witness: `TestTLAWitnessSessionReservationRechecksLiveWriter` pauses B
  after its real `sessionRun` lookup, publishes A as a real live-run record,
  releases A's reservation, then observes B's `claimSession` returning true.
- Impact: the reservation no longer proves exclusive native-session access.
  Concurrent resumes can then share one native transcript, mixing replies
  or invalidating checkpoints. The Go witness confirms the invalid grant;
  actual two-CLI transcript corruption was not executed.
- Conditions: distinct new turns/branches can pass the replay gate while
  sharing an earlier checkpoint; the first process has not advanced its
  transcript yet. The
  optional OS lease can block some canonical-outer cases when the durable
  tool journal is enabled. It is not a default universal native-session
  fence, is keyed by outer identity, and is absent for headerless requests.
- Suggested direction, not implemented: make the live-writer check,
  reservation, publication and release one coherent ownership protocol.
  A second scan alone is insufficient unless publication uses the same lock.
- Evidence: [TLC trace](results/session-reservation-observed.log),
  [Go witnesses](results/go-witnesses.log).

### TUI-F03: A finished startup can still launch a child [P1]

- Source: `internal/gateway/claude_tui_runtime_claude_interactive_transport.go`
  at lines 1729, 1747 and 1759; `finish`, `abort` and `removeRun` in
  `claude_tui_runtime_claude_subscription.go` at lines 3225, 3287 and 3346.
- Startup publishes the run before waiting for `Room`. If shutdown/abort
  finishes it in that interval, `tree` and `interactiveDone` are still nil,
  so cleanup closes and removes it. Startup later continues into
  `tuiRuntimeStartTree` without checking the closed run or coordinating
  process publication with abort.
- Trace: publish -> shutdown/finish/remove -> room granted -> child starts.
- Witness: `TestTLAWitnessNoLaunchAfterFinish` holds the real startup in the
  configurable Room callback, calls `abortAll`, releases Room, and observes
  a newly started tree on a closed run absent from `b.runs`. No source hook
  is inserted for this test; the Room dependency is a deterministic fake.
- Impact: a run already reported as ended can start native inference outside
  the bridge's live-run/callback bookkeeping. Additional spend, failed tool
  callbacks and incomplete shutdown accounting are possible. A permanent
  orphan or a real upstream tool side effect was not demonstrated.
- Suggested direction, not implemented: synchronize the pre-start,
  starting and started states with abort and retain ownership until a
  started child is actually reaped. Merely adding an unlocked `closed`
  check moves the race rather than eliminating it.
- Evidence: [TLC trace](results/launch-observed.log),
  [Go witnesses](results/go-witnesses.log).

### TUI-F04: A queued identical retry performs inference again [P2]

- Source: the completed-result check at
  `internal/gateway/claude_tui_runtime_claude_lite.go:146`, the gate at
  line 199 and `enterLiteTurn` at line 441.
- A retry checks completed results while the first request is still
  running, misses, and waits. When the first request saves its reply and
  releases the gate, the retry acquires the gate and starts/resumes native
  inference without rechecking the completed-result cache.
- Trace: lookup A, start A, lookup B misses, complete A, enter/start B.
- Witness: `TestTLAWitnessQueuedRetryDoesNotExecuteAgain` sends two identical
  tool-result HTTP-handler requests. It pauses the first in Room, observes
  the second at the actual retry wait, then releases the first. Both return
  200 but two native wrappers start, where the documented retry path should
  replay one completed reply.
- Impact: duplicate inference and cache writes, and an opportunity to
  propose the next tool again. No duplicate external tool execution is
  claimed: client-side tool effects are not made exactly-once by this bridge.
- Suggested direction, not implemented: after acquiring the request gate,
  revalidate completion before any native start, or publish one shared
  completion object for all waiters on that request.
- Evidence: [TLC trace](results/replay-queued-retry.log),
  [Go witnesses](results/go-witnesses.log).

## Rejected And Qualified Counterexamples

1. An initial `ServePool` abstraction allowed an old cancellation to inspect
   turn 1, let turn 1 complete and turn 2 start, then abort turn 2. A direct
   call to the internal cancellation helper can reproduce that sequence,
   but that is not enough to establish a production bug. Both real HTTP
   completion paths first call the stop function returned by
   `context.AfterFunc` (`serveLite:260`, `serveMessagesInner:3892`). If the
   cancellation callback has started, completion does not checkpoint or
   idle the process. If completion wins, that callback cannot later start.
   After adding this ownership fence, the observed model passes. The
   missing-fence case is retained as a negative control, not a finding.
2. `replay-no-header` shows that headerless requests have no identical-request
   admission gate. Without an unambiguous retry/session identity, identical
   content can also mean two independent first turns. This is a contract
   boundary, not a fifth confirmed defect or an argument to deduplicate all
   identical headerless requests.
3. Requester edits to historical text/media intentionally change the strict
   prefix. They are not diagnosed as a magpie defect and no accommodation,
   history normalization or weakened matching has been added.

## Coverage And Source Correspondence

This is a suite of bounded state-machine abstractions of the whole TUI
control surface, not one composed proof of every implementation path.
Atomic actions follow the named source operations, except where the model
deliberately splits a check from its later use to explore interleavings.

| Model | Responsibility and source operations | Finite scope |
| --- | --- | --- |
| `ModeRouting` | Mode configuration, startup snapshot, prerequisites, no silent standard fallback; `provider.SetAccountClaudeMode`, `readTUIModes`, `serveClaudeTUI`, auto-classifier and safeguards guards | 2 accounts; Boolean eligibility/dependency/mark/size inputs |
| `LaunchLifetime` | Publish, room wait, start, shutdown, reaping and room release; `startInteractiveWithPromptImages`, `abort`, `finish`, `removeRun` | 2 startups, one shutdown |
| `ReplayIdentity` | Completed replies, caller identity, request gate and wait/retry; `liteReplayKey`, `liteCompletedToolTurn`, `enterLiteTurn` | 2 requests, 1 semantic body and outer/account, same/different callers |
| `SessionReservation` | In-place native writer reservation vs branch fork; `sessionRun`, `claimSession`, `transcriptEndsAt` | 2 startups, 1 native session, an unchanged/advanced transcript boundary |
| `CheckpointAuthority` | Strict scoped prefix, model/instructions, complete tool results, summary fork isolation, dirty cancel/rewind and downgrade namespace | 2 scopes, 2 prefixes, 2 models, 2 instruction identities, at most 1 stored point, 2 parallel results |
| `ToolHandoff` | Transcript/callback order, duplicate presentation, early results, held/read-only tools, repair ownership, bounded recovery | 2 calls, callback/transcript/results independently ordered |
| `DurableRecovery` | WAL lock, valid/corrupt data, matched result digest, fsynced claim abstraction, one retry, process fencing and read-only launch | 3 contenders, 2 symbolic result digests, 0-2 accepted claims |
| `ServePool` | Wrapper ready, result draining, complete reply, idle admission, claim, generation-based TTL, HTTP cancellation ownership | 1 process, 2 native turns, bounded timer generations |
| `ProcessCleanup` | Private directory gate, room retention, process identity, RSS-known/unknown policy, reaping and PID-safe cleanup | 2 descendant PIDs and 2 start-time identities |
| `UsageAccounting` | Repeated/updated rows of one message, componentwise maxima, cumulative `Usage.add`, 1-hour-write subset | 2 message IDs and 3 repeated/updated rows, within one native accounting segment |

Important source files also read: `claude_tui_runtime_prefix.go`,
`claude_tui_runtime_native_checkpoint.go`,
`claude_tui_runtime_claude_interactive_tool_journal.go`,
`claude_tui_runtime_claude_interactive_orphan_recovery.go`,
`claude_tui_runtime_tool_repair.go`, `claude_tui_runtime_proctree*.go`,
`claude_tui_runtime_native_serve.go`, `ir.go`, and
`internal/claudetui/wrapper/src/{cli,serve-protocol}.ts`.

The following are assumptions or verification limits, not proven facts:

- Cryptographic identities are collision-free symbolic values. No proof of
  the JSON canonicalizer, schema parser, Unicode handling or hash functions.
- Directory ownership and native OS PID/start-time observations are modeled
  as reliable input predicates. No proof of symlink/TOCTOU resistance,
  Linux process enumeration, detached-tmux death, native CLI or shell code.
- A WAL append under its OS lock is atomic or unreadable/corrupt; byte-level
  torn writes, filesystem durability, power loss and cross-process checkpoint
  store updates are not exhaustively modeled. Local journal/file mutexes are
  not assumed to be cross-process native-session fences.
- A cancellation and a successful HTTP completion compete for the
  `context.AfterFunc` ownership fence. The refined pool model includes it.
- Liveness uses explicit weak fairness for enabled modeled completion,
  expiry or reaping actions. It says nothing about an unavailable upstream,
  a frozen OS, a pipe whose peer never reads, or real timeout durations.
- `ToolHandoff` abstracts the codec and callback validation predicates; the
  actual native MCP and caller-side external effects are not modeled.
- `ProcessCleanup` verifies known, sampled descendant identities only; it
  does not assert that every short-lived/reparented descendant was sampled.
- Usage checks are per accounting segment, not a proof that delayed rows,
  route attribution or whole-server billing across segments are accurate.
- Native search, images, thinking signatures, streaming codecs, error-word
  classification, the GUI/CLI editor and OAuth refresh are inspected as
  boundary interfaces, not modeled byte-for-byte.
- Modules are checked separately. No TLAPS refinement proof, unbounded
  proof, composition proof, vendor cache-hit claim or production load test
  is provided. Passing cases must not be called "the entire TUI is correct."

## Execution And Interpretation

`cases.json` records observed models, counterexamples, hypothesis-only
guards and negative controls. Hypothesis cases change abstract guards to
check whether a proposed direction blocks a trace; they do not implement
or verify a Go fix. Mutants intentionally remove duplicate/read-only guards
or add cumulative usage snapshots; their expected failures check that the
assertions can detect those mistakes.

Expected counterexample cases stop at the first violating state. Passing
cases exhaust their finite reachable graph. Quiescent terminal states are
legal, so `CHECK_DEADLOCK FALSE` is explicit: this audit does not claim a
general deadlock proof. State counts are per configuration, not distinct
production executions, and must not be added up as a coverage percentage.

Go witnesses are stored as `.go.txt` outside the normal test package so the
known-red assertions do not silently become regressions in the ordinary
suite. The runner temporarily installs them and inserts two observation
hooks, runs under temporary HOME with pinned Go caches, and restores the
original source bytes in `finally`. Run it only in this audit worktree, not
in a worktree another session is building or editing. These are deterministic
source-correspondence witnesses, not a newly passing regression suite.

See [results/summary.json](results/summary.json),
[results/go-witnesses-summary.json](results/go-witnesses-summary.json), and
[results/existing-tui-suite.log](results/existing-tui-suite.log) for actual
counts, logs and checker outcomes. The existing TUI Go suite was run
separately on unmodified production source with `-race`, Go 1.27.1, and
passed in 138.373s. Its passing tests do not invalidate the four missing guarantees.

## Recorded Run

TLC 2.19 / official tools 1.7.4: 10 modules, 23 configurations,
727,862 generated states and 481,164 distinct states summed across
configurations. These are overlapping bounded state spaces, not a
production coverage percentage. All 23 runs completed with their
declared outcomes; no timeout or unexpected checker error remains.

| Case | Distinct States | Actual Outcome |
| --- | ---: | --- |
| `launch-observed` | 17 | Counterexample: `NoStartAfterFinish` |
| `launch-fence-hypothesis` | 61 | No modeled violation |
| `replay-cross-caller` | 11 | Counterexample: `CallerIsolation` |
| `replay-queued-retry` | 15 | Counterexample: `OneExecutionPerRetry` |
| `replay-no-header` | 13 | Counterexample: `OneConcurrentRetry` |
| `replay-gate-safety` | 17 | No modeled violation |
| `replay-scoped-recheck-hypothesis-true` | 14 | No modeled violation |
| `replay-scoped-recheck-hypothesis-false` | 15 | No modeled violation |
| `pool-observed-cancel` | 39 | No modeled violation |
| `pool-idle-safety` | 39 | No modeled violation |
| `pool-cancel-fence-hypothesis` | 39 | No modeled violation |
| `checkpoint-safety` | 475,458 | No modeled violation |
| `tools-safety` | 1,260 | No modeled violation |
| `tools-duplicate-mutant` | 48 | Counterexample: `DispatchAtMostOnce` |
| `tools-readonly-mutant` | 37 | Counterexample: `ReadOnlyNeverDispatches` |
| `mode-safety` | 2,336 | No modeled violation |
| `session-reservation-observed` | 48 | Counterexample: `SingleNativeWriter` |
| `session-reservation-fence-hypothesis` | 52 | No modeled violation |
| `durable-recovery-safety` | 1,212 | No modeled violation |
| `process-cleanup-safety` | 360 | No modeled violation |
| `pool-missing-http-fence-negative-control` | 60 | Counterexample: `OldCancellationCannotKillNewTurn` |
| `usage-snapshot-safety` | 8 | No modeled violation |
| `usage-additive-mutant` | 5 | Counterexample: `NoRepeatedSnapshotBilling` |

Go witness run: `go test -race -count=20`; each of four intended
safety assertions fails 20/20 times, with no race-detector warning.
Unmodified-source existing TUI suite: `go test -race -tags nogui
./internal/gateway -run '^TestClaudeTUI' -count=1 -timeout=15m` passed
in 138.373 seconds. Go 1.27.1 was selected, matching the preceding
TUI verification rather than the Go 1.26.8 hard-coded-key mismatch.

The whole repository suite, cross-platform builds, GUI tests and real
Anthropic calls were not rerun: this audit adds formal artifacts and
isolated intentionally failing witnesses, not a runtime fix.

## Reproduce

Dependencies: Python 3, Java, Go, and the official `tla2tools.jar` from the
[TLA+ 1.7.4 release](https://github.com/tlaplus/tlaplus/releases/tag/v1.7.4).
The JAR reports TLC 2.19, revision `5a47802`; it is not stored in this repo.
The runner checks SHA-256:

```text
936a262061c914694dfd669a543be24573c45d5aa0ff20a8b96b23d01e050e88
```

This checksum pins the downloaded official release asset, not an independent
cryptographic signature of its publisher.

```sh
python3 formal/claude-tui/run_models.py --jar /path/to/tla2tools-1.7.4.jar
python3 formal/claude-tui/run_models.py --jar /path/to/tla2tools-1.7.4.jar \
  --case session-reservation-observed
python3 formal/claude-tui/run_witnesses.py
```

The runners return success only when every executed case matches its
declared outcome. A counterexample's success means the expected defect was
reproduced, not that its safety invariant passed. Unexpected errors, timeouts,
missing witness failures and race-detector warnings fail the runners.

The per-account `tui` bridge is the audited feature. The upstream standard
subscription bridge may itself launch an interactive CLI through a different
implementation; that path, general gateway routing, credentials, and the
rest of the application are not claimed as formally verified.

No runtime fix is included. Recommended order: caller isolation and native
writer exclusivity, startup/abort coordination, then queued retry completion.
