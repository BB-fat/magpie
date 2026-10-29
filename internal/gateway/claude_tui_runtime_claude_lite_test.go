package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const tuiRuntimeScOwner = "claude-tui\x00default"
const tuiRuntimeScOuter = "5b0e2c7a9d4f61e8a3b2c9d0e1f2a3b4c5d6e7f8091a2b3c4d5e6f708192a3b4"

// scWrapper stands in for the native PTY wrapper: it records each start's
// argv, prompt and — when it is told to --resume — the session file it
// would resume, then answers SC_ANSWER in text.
func tuiRuntimeScWrapper(t *testing.T) (string, func() []map[string]string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(dir, "claude"))
	logs := filepath.Join(dir, "log")
	if err := os.MkdirAll(logs, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SC_LOG", logs)
	script := `#!/bin/sh
d="$SC_LOG/run-$$"
mkdir -p "$d"
printf '%s\n' "$@" > "$d/argv"
pwd > "$d/cwd"
cat > "$d/stdin"
sid=inner-sc
prev=""
for a in "$@"; do
  if [ "$prev" = "--resume" ]; then
    sid="$a"; echo "$a" > "$d/resume"
    for f in "$CLAUDE_CONFIG_DIR"/projects/*/"$a".jsonl; do [ -f "$f" ] && cp "$f" "$d/seed.jsonl"; done
  fi
  prev="$a"
done
printf '%s\n' '{"type":"assistant","uuid":"a1","session_id":"'$sid'","message":{"id":"m1","model":"claude-opus-5-5","content":[{"type":"text","text":"SC_ANSWER"}]}}'
printf '%s\n' '{"type":"result","session_id":"'$sid'","stop_reason":"end_turn"}'
`
	p := filepath.Join(dir, "sc-bridge")
	if err := os.WriteFile(p, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, p)
	return logs, func() []map[string]string {
		entries, _ := os.ReadDir(logs)
		var runs []map[string]string
		for _, e := range entries {
			run := map[string]string{}
			for _, name := range []string{"argv", "stdin", "resume", "seed.jsonl", "cwd", "tools.json"} {
				if b, err := os.ReadFile(filepath.Join(logs, e.Name(), name)); err == nil {
					run[name] = strings.TrimSpace(string(b))
				}
			}
			runs = append(runs, run)
		}
		return runs
	}
}

func tuiRuntimeScRuntime(t *testing.T) (*tuiRuntime, *tuiRuntimeSubscriptionBridge) {
	t.Helper()
	b := tuiRuntimeNewSubscriptionBridge()
	t.Cleanup(func() { b.abortAll(); tuiRuntimeBridges.Delete(b) })
	return &tuiRuntime{bridge: b}, b
}

func tuiRuntimeScServe(t *testing.T, rt *tuiRuntime, outer, body string) (int, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
	code, _ := rt.serveMessages(rec, req, "claude-opus-5-5", outer, []byte(body), &Usage{})
	return code, rec.Body.String()
}

// scSaveClean writes a clean checkpoint for the two-message base history.
func tuiRuntimeScSaveClean(t *testing.T, b *tuiRuntimeSubscriptionBridge, session string, dirty bool) {
	t.Helper()
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "先读一下配置"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "配置已读"}}},
	}
	baseReq := &tuiRuntimeRequest{Request: Request{Model: "claude-opus-5-5", Messages: base}}
	keys := tuiRuntimeConversationKeys(tuiRuntimeScOwner, base)
	if err := b.saveInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter, tuiRuntimeInteractiveSessionEntry{
		SessionID: session, Model: "claude-opus-5-5", ThreadKey: firstWords(&baseReq.Request),
		ConvKey: keys[0], ReplyKey: tuiRuntimeAssistantReplyKey(base[1]),
		InstructionKey: tuiRuntimeInstructionKey(baseReq), Dirty: dirty,
	}); err != nil {
		t.Fatal(err)
	}
}

func tuiRuntimeOnlyRun(t *testing.T, runs []map[string]string) map[string]string {
	t.Helper()
	if len(runs) != 1 {
		t.Fatalf("%d native starts, want 1: %v", len(runs), runs)
	}
	return runs[0]
}

// The request that continues its clean checkpoint exactly resumes that
// native session with only the new turn: no rebuild.
func TestClaudeTUIRuntimeLiteExactCheckpointResumes(t *testing.T) {
	_, runs := tuiRuntimeScWrapper(t)
	rt, b := tuiRuntimeScRuntime(t)
	tuiRuntimeScSaveClean(t, b, "inner-saved", false)
	code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, `{"model":"claude-opus-5-5","max_tokens":64,"messages":[
		{"role":"user","content":"先读一下配置"},{"role":"assistant","content":"配置已读"},{"role":"user","content":"再看日志"}]}`)
	if code != 200 || !strings.Contains(out, "SC_ANSWER") {
		t.Fatalf("exact continuation failed: %d %s", code, out)
	}
	run := tuiRuntimeOnlyRun(t, runs())
	if run["resume"] != "inner-saved" || run["seed.jsonl"] != "" {
		t.Fatalf("did not resume the checkpointed session as is: %v", run)
	}
	if !strings.Contains(run["stdin"], "再看日志") || strings.Contains(run["stdin"], "先读一下配置") {
		t.Fatalf("the resumed turn was not just the new message: %q", run["stdin"])
	}
}

// A history Science rewrote (its own compaction, an edit) no longer
// follows the checkpoint. It is answered from a session rebuilt from that
// history — the 502 this used to be is gone.
func TestClaudeTUIRuntimeLiteRewrittenHistoryRebuildsInsteadOfRefusing(t *testing.T) {
	_, runs := tuiRuntimeScWrapper(t)
	rt, b := tuiRuntimeScRuntime(t)
	tuiRuntimeScSaveClean(t, b, "inner-saved", false)
	code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, `{"model":"claude-opus-5-5","max_tokens":64,"messages":[
		{"role":"user","content":"<summary>之前读了配置并修了两处</summary>"},{"role":"assistant","content":"好的，基于摘要继续"},{"role":"user","content":"现在跑测试"}]}`)
	if code != 200 || !strings.Contains(out, "SC_ANSWER") {
		t.Fatalf("rewritten history was refused: %d %s", code, out)
	}
	run := tuiRuntimeOnlyRun(t, runs())
	if run["resume"] == "" || run["resume"] == "inner-saved" {
		t.Fatalf("not a rebuilt session: %v", run)
	}
	seed := run["seed.jsonl"]
	if !strings.Contains(seed, "之前读了配置并修了两处") || !strings.Contains(seed, "好的，基于摘要继续") || strings.Contains(seed, "现在跑测试") {
		t.Fatalf("seed is not the history before the new turn: %s", seed)
	}
	if !strings.Contains(run["stdin"], "现在跑测试") {
		t.Fatalf("the new turn was not the prompt: %q", run["stdin"])
	}
	// The rebuilt session is the conversation's checkpoint now: the next
	// turn resumes it exactly.
	saved, ok, err := b.loadInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter)
	if err != nil || !ok || saved.Dirty || saved.SessionID != run["resume"] {
		t.Fatalf("the rebuilt session did not become the clean checkpoint: %+v ok=%t err=%v", saved, ok, err)
	}
}

// A turn cut off (client gone, quota, crash) leaves a dirty checkpoint;
// the next request is rebuilt from what Science has, not refused.
func TestClaudeTUIRuntimeLiteDirtyCheckpointRebuilds(t *testing.T) {
	_, runs := tuiRuntimeScWrapper(t)
	rt, b := tuiRuntimeScRuntime(t)
	tuiRuntimeScSaveClean(t, b, "inner-saved", true)
	code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, `{"model":"claude-opus-5-5","max_tokens":64,"messages":[
		{"role":"user","content":"先读一下配置"},{"role":"assistant","content":"配置已读"},{"role":"user","content":"继续"}]}`)
	if code != 200 || !strings.Contains(out, "SC_ANSWER") {
		t.Fatalf("dirty checkpoint was refused: %d %s", code, out)
	}
	if run := tuiRuntimeOnlyRun(t, runs()); run["resume"] == "inner-saved" || !strings.Contains(run["seed.jsonl"], "配置已读") {
		t.Fatalf("dirty session was resumed instead of rebuilt: %v", run)
	}
}

// Tool results whose native run is gone (a tool that ran longer than the
// run waited, a runtime restart) continue in a rebuilt session that sees
// its call answered — not the old 409 "no matching live MCP run".
func TestClaudeTUIRuntimeLiteOrphanToolResultRebuildsWithStructuredHistory(t *testing.T) {
	_, runs := tuiRuntimeScWrapper(t)
	rt, _ := tuiRuntimeScRuntime(t)
	code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, `{"model":"claude-opus-5-5","max_tokens":64,
		"tools":[{"name":"Bash","input_schema":{"type":"object"}}],
		"messages":[
		{"role":"user","content":"跑一下分析脚本"},
		{"role":"assistant","content":[{"type":"text","text":"我来运行"},{"type":"tool_use","id":"toolu_run1","name":"Bash","input":{"cmd":"python analyze.py"}}]},
		{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_run1","content":"r2=0.93"}]}]}`)
	if code != 200 || !strings.Contains(out, "SC_ANSWER") {
		t.Fatalf("orphan tool result was refused: %d %s", code, out)
	}
	run := tuiRuntimeOnlyRun(t, runs())
	seed := run["seed.jsonl"]
	for _, want := range []string{`"name":"mcp__magpie__Bash"`, `"tool_use_id":"toolu_run1"`, "r2=0.93", "跑一下分析脚本"} {
		if !strings.Contains(seed, want) {
			t.Fatalf("seed lacks %s: %s", want, seed)
		}
	}
	if !strings.Contains(run["stdin"], tuiRuntimeLiteContinueText) {
		t.Fatalf("a results-only turn was not told to continue: %q", run["stdin"])
	}
	// The seed must not end on the user's message, or Claude Code answers it
	// itself ("No response requested.") into the reply.
	lines := strings.Split(seed, "\n")
	if last := lines[len(lines)-1]; !strings.Contains(last, `"type":"assistant"`) || !strings.Contains(last, tuiRuntimeLiteResultsReceived) {
		t.Fatalf("seed does not end with an assistant row: %s", last)
	}
	if strings.Contains(run["argv"], "python analyze.py") || strings.Contains(run["stdin"], "python analyze.py") {
		t.Fatalf("the old tool call leaked into the new turn: %v", run)
	}
}

// Science's mid-turn message: it reports the tool calls it did not run as
// not_dispatched and sends the user's words beside them. With no live run
// both reach the rebuilt session — results in history, words as the turn.
func TestClaudeTUIRuntimeLiteInterjectionWithoutLiveRunKeepsTheUsersWords(t *testing.T) {
	_, runs := tuiRuntimeScWrapper(t)
	rt, _ := tuiRuntimeScRuntime(t)
	code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, `{"model":"claude-opus-5-5","max_tokens":64,
		"tools":[{"name":"Bash","input_schema":{"type":"object"}}],
		"messages":[
		{"role":"user","content":"训练三个模型"},
		{"role":"assistant","content":[{"type":"tool_use","id":"toolu_a","name":"Bash","input":{"cmd":"train a"}},{"type":"tool_use","id":"toolu_b","name":"Bash","input":{"cmd":"train b"}}]},
		{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_a","content":"a done"},
			{"type":"tool_result","tool_use_id":"toolu_b","is_error":true,"content":"{\"status\":\"not_dispatched\"}"},
			{"type":"text","text":"先别训练 b，改成只比较 a"}]}]}`)
	if code != 200 || !strings.Contains(out, "SC_ANSWER") {
		t.Fatalf("interjection was refused: %d %s", code, out)
	}
	run := tuiRuntimeOnlyRun(t, runs())
	if !strings.Contains(run["stdin"], "先别训练 b，改成只比较 a") || strings.Contains(run["stdin"], tuiRuntimeLiteContinueText) {
		t.Fatalf("the user's words were not the turn: %q", run["stdin"])
	}
	if !strings.Contains(run["seed.jsonl"], "not_dispatched") || !strings.Contains(run["seed.jsonl"], "a done") {
		t.Fatalf("tool results missing from the seed: %s", run["seed.jsonl"])
	}
}

// A run parked on a tool call Science's next request does not answer (the
// user stopped, Science moved on) is stopped and the request answered —
// where Desktop's path refuses with 409 to protect a live tool.
func TestClaudeTUIRuntimeLiteNewTurnBesideAParkedRunInsteadOf409(t *testing.T) {
	for _, tc := range []struct {
		name, reply string
		stopped     bool
	}{
		// The client shows its own words where the call was: another branch
		// from the first message, which may still come back with results.
		{"another_branch", `"已停止"`, false},
		// The client's history holds the call and goes on without its
		// result: the run will never be answered.
		{"went_on_without_the_result", `[{"type":"tool_use","id":"call_1","name":"write_file","input":{"path":"a"}}]`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := tuiRuntimeOlDir(t)
			tuiRuntimeOlWrapper(t, dir, "sc-park-bridge", tuiRuntimeOlParkScript())
			rt, b := tuiRuntimeScRuntime(t)
			t.Setenv("DAV1878_PARK_RELEASE", filepath.Join(dir, "never"))
			first := `{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"write_file","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"把结果写进 notes.txt"}]}`
			if code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, first); code != 200 || !strings.Contains(out, `"type":"tool_use"`) {
				t.Fatalf("first turn did not park: %d %s", code, out)
			}
			parked := b.interactiveOuterOwner(tuiRuntimeScOwner, tuiRuntimeScOuter)
			if parked == nil {
				t.Fatal("no parked run owns the conversation")
			}
			_, runs := tuiRuntimeScWrapper(t)
			code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, `{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"write_file","input_schema":{"type":"object"}}],"messages":[
		{"role":"user","content":"把结果写进 notes.txt"},{"role":"assistant","content":`+tc.reply+`},{"role":"user","content":"算了，直接告诉我结果"}]}`)
			if code != 200 || !strings.Contains(out, "SC_ANSWER") {
				t.Fatalf("new turn on a parked conversation was refused: %d %s", code, out)
			}
			parked.mu.Lock()
			closed := parked.closed
			parked.mu.Unlock()
			if closed != tc.stopped {
				t.Fatalf("parked run stopped=%t, want %t", closed, tc.stopped)
			}
			tuiRuntimeOnlyRun(t, runs())
		})
	}
}

// The very same request while it is still being answered (a client's
// retry) waits for it, and past the wait is told to retry: it is never
// answered by a second native run beside the first. Another request of
// the conversation is not held up by it.
func TestClaudeTUIRuntimeLiteSameRequestInFlightAsksForRetry(t *testing.T) {
	_, runs := tuiRuntimeScWrapper(t)
	rt, b := tuiRuntimeScRuntime(t)
	old := tuiRuntimeLiteTurnWait
	tuiRuntimeLiteTurnWait = 200 * time.Millisecond
	t.Cleanup(func() { tuiRuntimeLiteTurnWait = old })
	body := `{"model":"claude-opus-5-5","max_tokens":64,"messages":[{"role":"user","content":"先读一下配置"},{"role":"assistant","content":"配置已读"},{"role":"user","content":"继续"}]}`
	req, err := tuiRuntimeParseRequest([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Model = "claude-opus-5-5"
	release, ok := b.enterLiteTurn(t.Context(), tuiRuntimeLiteReplayKey(tuiRuntimeScOwner, req))
	if !ok {
		t.Fatal("enter")
	}
	defer release()
	rec := httptest.NewRecorder()
	code, _ := rt.serveMessages(rec, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body)), "claude-opus-5-5", tuiRuntimeScOuter, []byte(body), &Usage{})
	if code != http.StatusServiceUnavailable || rec.Header().Get("Retry-After") == "" {
		t.Fatalf("same request in flight: %d retry-after=%q", code, rec.Header().Get("Retry-After"))
	}
	if n := len(runs()); n != 0 {
		t.Fatalf("%d native runs started for a request already in flight", n)
	}
	other := strings.Replace(body, "继续", "换个问题", 1)
	if code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, other); code != 200 || !strings.Contains(out, "SC_ANSWER") {
		t.Fatalf("another request of the conversation: %d %s", code, out)
	}
}

// A request without a conversation id (helper calls such as memory
// extraction forks) runs on a session rebuilt from its own history and never
// touches any conversation's checkpoint.
func TestClaudeTUIRuntimeLiteStatelessRequestIsSeeded(t *testing.T) {
	_, runs := tuiRuntimeScWrapper(t)
	rt, b := tuiRuntimeScRuntime(t)
	tuiRuntimeScSaveClean(t, b, "inner-saved", false)
	before, _, _ := b.loadInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter)
	code, out := tuiRuntimeScServe(t, rt, "", `{"model":"claude-opus-5-5","max_tokens":64,"tool_choice":{"type":"none"},
		"tools":[{"name":"Bash","input_schema":{"type":"object"}}],
		"messages":[
		{"role":"user","content":"先读一下配置"},
		{"role":"assistant","content":[{"type":"tool_use","id":"toolu_r","name":"Bash","input":{"cmd":"cat cfg"}}]},
		{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_r","content":"k=v"}]},
		{"role":"assistant","content":"配置已读"},
		{"role":"user","content":"Extract durable memories from the conversation above."}]}`)
	if code != 200 || !strings.Contains(out, "SC_ANSWER") {
		t.Fatalf("fork failed: %d %s", code, out)
	}
	run := tuiRuntimeOnlyRun(t, runs())
	if !strings.Contains(run["seed.jsonl"], "k=v") || !strings.Contains(run["stdin"], "Extract durable memories") {
		t.Fatalf("fork not seeded from its history: %v", run)
	}
	after, _, _ := b.loadInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter)
	if after != before {
		t.Fatalf("a helper call changed the conversation checkpoint: %+v -> %+v", before, after)
	}
	if b.interactiveOuterOwner(tuiRuntimeScOwner, tuiRuntimeScOuter) != nil {
		t.Fatal("a helper call took ownership of the conversation")
	}
}

func TestClaudeTUIRuntimeLitePairedToolTraffic(t *testing.T) {
	msgs := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "go"}}},
		{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "a", Name: "Bash"}, {Kind: ToolCall, ID: "lost", Name: "Bash"}}},
		{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "a", Text: "ok"}, {Kind: ToolResult, CallID: "stray", Text: "?"}}},
		{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "unanswered", Name: "Bash"}}},
	}
	paired := tuiRuntimePairedToolTraffic(msgs)
	got, _ := json.Marshal(paired)
	for _, gone := range []string{`"lost"`, `"unanswered"`} {
		if strings.Contains(string(got), gone) {
			t.Fatalf("%s survived: %s", gone, got)
		}
	}
	if !strings.Contains(string(got), `"a"`) {
		t.Fatalf("the answered call was dropped: %s", got)
	}
	// A compacted-away call must not erase the output its client returned.
	if !strings.Contains(string(got), `[tool result id=stray]`) || !strings.Contains(string(got), "?") {
		t.Fatalf("the orphan output was dropped: %s", got)
	}
	for _, m := range paired {
		for _, p := range m.Parts {
			if p.Kind == ToolResult && p.CallID == "stray" {
				t.Fatal("the orphan survived as native tool traffic")
			}
		}
	}
}

// An identified subagent checkpoints and resumes independently of its parent.
func TestClaudeTUIRuntimeLiteSubagentOnTheConversationIdResumesIndependently(t *testing.T) {
	logs, runs := tuiRuntimeScWrapper(t)
	rt, b := tuiRuntimeScRuntime(t)
	tuiRuntimeScSaveClean(t, b, "inner-saved", false)
	b.anchorOuterThread(tuiRuntimeScOwner, tuiRuntimeScOuter, &tuiRuntimeRequest{Request: Request{Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "先读一下配置"}}}}}})
	before, _, _ := b.loadInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter)
	body := `{"model":"claude-opus-5-5","max_tokens":64,"messages":[
		{"role":"user","content":"Explore the repository and list its packages"},{"role":"assistant","content":"Looking"},{"role":"user","content":"go on"}]}`
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
	req.Header.Set("x-claude-code-agent-id", "explorer")
	rec := httptest.NewRecorder()
	code, _ := rt.serveMessages(rec, req, "claude-opus-5-5", tuiRuntimeScOuter, []byte(body), &Usage{})
	out := rec.Body.String()
	if code != 200 || !strings.Contains(out, "SC_ANSWER") {
		t.Fatalf("helper: %d %s", code, out)
	}
	first := tuiRuntimeOnlyRun(t, runs())
	outer := tuiRuntimeLiteOuterSession(tuiRuntimeScOuter, req.Header)
	saved, ok, err := b.loadInteractiveSession(tuiRuntimeScOwner, outer)
	if err != nil || !ok || saved.Dirty || saved.SessionID != first["resume"] {
		t.Fatalf("subagent has no clean checkpoint: %+v ok=%t err=%v", saved, ok, err)
	}
	// A new bridge proves the next turn uses persisted session state.
	rt, b = tuiRuntimeScRuntime(t)
	next := strings.TrimSuffix(body, "]}") + `,{"role":"assistant","content":"SC_ANSWER"},{"role":"user","content":"List the tests too"}]}`
	entries, err := os.ReadDir(logs)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(logs, entry.Name())); err != nil {
			t.Fatal(err)
		}
	}
	req = httptest.NewRequest("POST", "/v1/messages", strings.NewReader(next))
	req.Header.Set("x-claude-code-agent-id", "explorer")
	rec = httptest.NewRecorder()
	code, _ = rt.serveMessages(rec, req, "claude-opus-5-5", tuiRuntimeScOuter, []byte(next), &Usage{})
	if code != 200 {
		t.Fatalf("subagent continuation: %d %s", code, rec.Body.String())
	}
	second := tuiRuntimeOnlyRun(t, runs())
	if second["resume"] != saved.SessionID || strings.Contains(second["stdin"], "Explore the repository") || !strings.Contains(second["stdin"], "List the tests too") {
		t.Fatalf("subagent did not resume with suffix only: %v", second)
	}
	after, _, _ := b.loadInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter)
	if after != before {
		t.Fatalf("a helper thread changed the conversation checkpoint: %+v -> %+v", before, after)
	}
}

// The conversation's own next request that does not carry the parked run's
// tool call (its client never received it: the reply was cut off, and the
// turn is sent again) stops the parked run; exactly one native run is left.
func TestClaudeTUIRuntimeLiteRetriedTurnStopsTheParkedRunItNeverSaw(t *testing.T) {
	dir := tuiRuntimeOlDir(t)
	tuiRuntimeOlWrapper(t, dir, "lite-park-bridge", tuiRuntimeOlParkScript())
	rt, b := tuiRuntimeScRuntime(t)
	t.Setenv("DAV1878_PARK_RELEASE", filepath.Join(dir, "never"))
	first := `{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"write_file","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"把结果写进 notes.txt"}]}`
	if code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, first); code != 200 || !strings.Contains(out, `"type":"tool_use"`) {
		t.Fatalf("first turn did not park: %d %s", code, out)
	}
	parked := b.interactiveOuterOwner(tuiRuntimeScOwner, tuiRuntimeScOuter)
	if parked == nil {
		t.Fatal("no parked run")
	}
	_, runs := tuiRuntimeScWrapper(t)
	code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, first)
	if code != 200 || !strings.Contains(out, "SC_ANSWER") {
		t.Fatalf("retried turn: %d %s", code, out)
	}
	parked.mu.Lock()
	closed := parked.closed
	parked.mu.Unlock()
	if !closed {
		t.Fatal("the parked run was left alive beside the retried turn")
	}
	tuiRuntimeOnlyRun(t, runs())
	if _, same := tuiRuntimeOlLiveOuterRuns(b, tuiRuntimeScOwner, tuiRuntimeScOuter); same > 1 {
		t.Fatalf("%d live runs on one conversation", same)
	}
}

// Claude Code reports a reached usage limit as a synthetic API error with the
// generic 502; clients must see 429 so they wait instead of retrying at once.
func TestClaudeTUIRuntimeLiteUsageLimitIs429(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(dir, "claude"))
	p := filepath.Join(dir, "limit-bridge")
	os.WriteFile(p, []byte("#!/bin/sh\ncat >/dev/null\nprintf '%s\\n' '{\"type\":\"assistant\",\"isApiErrorMessage\":true,\"session_id\":\"s\",\"message\":{\"model\":\"<synthetic>\",\"content\":[{\"type\":\"text\",\"text\":\"You have hit your session limit \u00b7 resets 5am (UTC)\"}]}}'\nprintf '%s\\n' '{\"type\":\"result\",\"session_id\":\"s\",\"is_error\":true,\"result\":\"You have hit your session limit\"}'\n"), 0o700)
	tuiRuntimeSetWrapper(t, p)
	rt, _ := tuiRuntimeScRuntime(t)
	for _, stream := range []string{"true", "false"} {
		code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, `{"model":"claude-opus-5-5","max_tokens":64,"stream":`+stream+`,"messages":[{"role":"user","content":"hi"}]}`)
		if code != http.StatusTooManyRequests || !strings.Contains(out, "session limit") {
			t.Fatalf("stream=%s: usage limit answered %d %s", stream, code, out)
		}
	}
	// The wrapper's own result for a StopFailure rate_limit, as GTT got it:
	// no status, no text, the reason in `error` above the pane's tail.
	os.WriteFile(p, []byte("#!/bin/sh\ncat >/dev/null\nprintf '%s\\n' '{\"type\":\"result\",\"subtype\":\"error_during_execution\",\"is_error\":true,\"api_error_status\":null,\"result\":null,\"error\":\"Claude StopFailure hook fired: rate_limit\\nPane tail:\\nPRIVATE_PANE\",\"session_id\":\"s\"}'\n"), 0o700)
	for _, stream := range []string{"true", "false"} {
		code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, `{"model":"claude-opus-5-5","max_tokens":64,"stream":`+stream+`,"messages":[{"role":"user","content":"hi"}]}`)
		if code != http.StatusTooManyRequests || strings.Contains(out, "PRIVATE_PANE") {
			t.Fatalf("stream=%s: StopFailure rate limit answered %d %s", stream, code, out)
		}
	}
	if tuiRuntimeLiteStatus(http.StatusBadGateway, "prompt is too long: exceeded context") != http.StatusBadGateway {
		t.Fatal("a context overflow was taken for a usage limit")
	}
	if tuiRuntimeLiteStatus(http.StatusBadRequest, "rate limit") != http.StatusBadRequest {
		t.Fatal("a definite status was overridden")
	}
}

// The plugin's stateless marker reaches serveMessages as no conversation.
func TestClaudeTUIRuntimeLiteSignedStatelessSessionHasNoConversation(t *testing.T) {
	if tuiRuntimeStatelessSession != "stateless" {
		t.Fatal("the plugin and runtime must agree on the stateless marker")
	}
}

func TestClaudeTUIRuntimeLiteVisibleToolHandoffRejectsWrongOrPartialResults(t *testing.T) {
	request := func(t *testing.T, history string) *tuiRuntimeRequest {
		t.Helper()
		req, err := tuiRuntimeParseRequest([]byte(`{"model":"claude-opus-5-5","max_tokens":32,"messages":` + history + `}`))
		if err != nil {
			t.Fatal(err)
		}
		return req
	}
	const prefix = `[{"role":"user","content":"do task"},{"role":"assistant","content":[{"type":"tool_use","id":"call-1","name":"echo","input":{}}]},`
	for name, tc := range map[string]struct {
		history string
		valid   bool
	}{
		"verified":                {prefix + `{"role":"user","content":[{"type":"tool_result","tool_use_id":"call-1","content":"done"}]}]`, true},
		"mismatched":              {prefix + `{"role":"user","content":[{"type":"tool_result","tool_use_id":"forged","content":"done"}]}]`, false},
		"duplicated":              {prefix + `{"role":"user","content":[{"type":"tool_result","tool_use_id":"call-1","content":"done"},{"type":"tool_result","tool_use_id":"call-1","content":"again"}]}]`, false},
		"orphan-after-compaction": {`[{"role":"user","content":"compacted earlier work"},{"role":"assistant","content":"summary"},{"role":"user","content":[{"type":"tool_result","tool_use_id":"earlier-call","content":"done"}]}]`, true},
		"ordinary-text":           {`[{"role":"user","content":"hi"},{"role":"assistant","content":"hello"},{"role":"user","content":"continue"}]`, true},
	} {
		t.Run(name, func(t *testing.T) {
			err := tuiRuntimeLiteVisibleToolHandoff(request(t, tc.history))
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%t, err=%v", tc.valid, err)
			}
		})
	}
}

// Tool results sent again after their reply was delivered (a client retry,
// a side call on the conversation's identity) receive the already completed
// answer: no 409, no second PTY, and no checkpoint change.
func TestClaudeTUIRuntimeLiteRepeatOfDeliveredToolResultsIsAnsweredStateless(t *testing.T) {
	_, runs := tuiRuntimeScWrapper(t)
	rt, b := tuiRuntimeScRuntime(t)
	tuiRuntimeScSaveClean(t, b, "inner-saved", false)
	body := `{"model":"claude-opus-5-5","max_tokens":64,"messages":[{"role":"user","content":"run"},{"role":"assistant","content":[{"type":"tool_use","id":"call-1","name":"echo","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call-1","content":"finished"}]}]}`
	req, err := tuiRuntimeParseRequest([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	b.noteLiteCompletedToolTurn(tuiRuntimeScOwner, tuiRuntimeScOuter, req, "SC_ANSWER")
	before, _, _ := b.loadInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter)
	code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, body)
	if code != 200 || !strings.Contains(out, "SC_ANSWER") {
		t.Fatalf("repeat refused: %d %s", code, out)
	}
	if got := runs(); len(got) != 0 {
		t.Fatalf("a completed duplicate started another native run: %v", got)
	}
	if after, _, _ := b.loadInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter); after != before {
		t.Fatalf("a repeat changed the conversation checkpoint: %+v -> %+v", before, after)
	}
}

func TestClaudeTUIRuntimeLiteAlreadyDeliveredToolResultsNotRebuiltInSameProcess(t *testing.T) {
	b := &tuiRuntimeSubscriptionBridge{}
	req, err := tuiRuntimeParseRequest([]byte(`{"model":"claude-opus-5-5","messages":[{"role":"user","content":"run"},{"role":"assistant","content":[{"type":"tool_use","id":"call-1","name":"echo","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call-1","content":"finished"}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if b.liteToolTurnWasDelivered(tuiRuntimeScOwner, tuiRuntimeScOuter, req) {
		t.Fatal("unseen result must be eligible for first delivery")
	}
	b.noteLiteCompletedToolTurn(tuiRuntimeScOwner, tuiRuntimeScOuter, req, "Original completed answer")
	if !b.liteToolTurnWasDelivered(tuiRuntimeScOwner, tuiRuntimeScOuter, req) {
		t.Fatal("already delivered result was accepted again")
	}
	if completed, ok := b.liteCompletedToolTurn(tuiRuntimeScOwner, tuiRuntimeScOuter, req); !ok || completed.reply != "Original completed answer" {
		t.Fatalf("duplicate must return the original text without a new native run: ok=%v", ok)
	}
	next := *req
	next.Messages = append(append([]Message(nil), req.Messages...), Message{Role: "user", Parts: []Part{{Kind: Text, Text: "continue"}}})
	if b.liteToolTurnWasDelivered(tuiRuntimeScOwner, tuiRuntimeScOuter, &next) {
		t.Fatal("a genuine additional user message was classified as replay")
	}
	if b.liteToolTurnWasDelivered(tuiRuntimeScOwner, "other-session", req) {
		t.Fatal("tool delivery proof crossed conversation identity")
	}
}

// A conversation's run let go while it waited on its tool calls leaves its
// session: their results resume a copy of it, the rows the model saw kept
// byte for byte (a prefix Anthropic has cached), the results appended where
// the run would have had them, and what it wrote as it was stopped left out.
func TestClaudeTUIRuntimeLiteParkedRunLetGoResumesItsOwnSession(t *testing.T) {
	dir := tuiRuntimeOlDir(t)
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(dir, "claude"))
	tuiRuntimeOlWrapper(t, dir, "lite-park-bridge", tuiRuntimeOlParkScript())
	t.Setenv("DAV1878_PARK_RELEASE", filepath.Join(dir, "never"))
	rt, b := tuiRuntimeScRuntime(t)
	tools := `"system":"PARKED_INSTRUCTIONS","tools":[{"name":"write_file","input_schema":{"type":"object"}}]`
	first := `{"model":"claude-opus-5-5","max_tokens":64,` + tools + `,"messages":[{"role":"user","content":"把结果写进 notes.txt"}]}`
	if code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, first); code != 200 || !strings.Contains(out, `"call_1"`) {
		t.Fatalf("first turn did not park: %d %s", code, out)
	}
	parked := b.interactiveOuterOwner(tuiRuntimeScOwner, tuiRuntimeScOuter)
	if parked == nil {
		t.Fatal("no parked run owns the conversation")
	}
	// The native session as Claude Code saved it: the call row (its input's
	// keys in Claude Code's order), then what an interrupted run appends.
	work, err := tuiRuntimeInteractiveSessionWorkDir(tuiRuntimeScOwner, tuiRuntimeScOuter)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(work) })
	sessions := tuiRuntimeClaudeSessions("", work)[0]
	if err := os.MkdirAll(sessions, 0o700); err != nil {
		t.Fatal(err)
	}
	native := `{"type":"user","uuid":"u1","parentUuid":null,"sessionId":"inner-ol","message":{"role":"user","content":"<external_system_instructions>\nPARKED_INSTRUCTIONS\n</external_system_instructions>\n\n把结果写进 notes.txt"}}
{"type":"assistant","uuid":"t0","parentUuid":"u1","sessionId":"inner-ol","message":{"id":"m1","role":"assistant","content":[{"type":"thinking","thinking":"","signature":"SIG"}]}}
{"type":"assistant","uuid":"t1","parentUuid":"t0","sessionId":"inner-ol","message":{"id":"m1","role":"assistant","content":[{"type":"tool_use","id":"call_1","name":"mcp__magpie__write_file","input":{"zeta":1,"path":"a"}}]}}
{"type":"user","uuid":"i1","parentUuid":"t1","sessionId":"inner-ol","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"call_1","content":"[Request interrupted by user for tool use]"}]}}
{"type":"mode","mode":"normal","sessionId":"inner-ol"}
`
	if err := os.WriteFile(filepath.Join(sessions, "inner-ol.jsonl"), []byte(native), 0o600); err != nil {
		t.Fatal(err)
	}
	parked.abort() // its wait (interactiveParkLongest) is over
	tuiRuntimeOlWaitReaped(t, parked)

	_, runs := tuiRuntimeScWrapper(t)
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(dir, "claude")) // scWrapper moved it
	code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, `{"model":"claude-opus-5-5","max_tokens":64,`+tools+`,"messages":[
		{"role":"user","content":"把结果写进 notes.txt"},
		{"role":"assistant","content":[{"type":"tool_use","id":"call_1","name":"write_file","input":{"path":"a"}}]},
		{"role":"user","content":[{"type":"tool_result","tool_use_id":"call_1","content":"WROTE_OK"}]}]}`)
	if code != 200 || !strings.Contains(out, "SC_ANSWER") {
		t.Fatalf("results after the run was let go: %d %s", code, out)
	}
	run := tuiRuntimeOnlyRun(t, runs())
	if run["resume"] == "" || run["resume"] == "inner-ol" {
		t.Fatalf("did not resume a copy of the run's own session: %v", run)
	}
	copied := run["seed.jsonl"]
	for _, kept := range []string{"PARKED_INSTRUCTIONS", `"signature":"SIG"`, `"input":{"zeta":1,"path":"a"}`, `"sessionId":"` + run["resume"] + `"`, "WROTE_OK", tuiRuntimeLiteResultsReceived} {
		if !strings.Contains(copied, kept) {
			t.Fatalf("resumed session lacks %s:\n%s", kept, copied)
		}
	}
	for _, gone := range []string{"interrupted", `"sessionId":"inner-ol"`, `"mode"`} {
		if strings.Contains(copied, gone) {
			t.Fatalf("resumed session keeps %s:\n%s", gone, copied)
		}
	}
	if !strings.Contains(run["stdin"], tuiRuntimeLiteContinueText) || strings.Contains(run["stdin"], "PARKED_INSTRUCTIONS") {
		t.Fatalf("turn prompt: %q", run["stdin"])
	}
	if b, err := os.ReadFile(filepath.Join(sessions, "inner-ol.jsonl")); err != nil || string(b) != native {
		t.Fatalf("the run's own session was changed: %v", err)
	}
}

// Results that are not exactly the let-go run's calls are not resumed into
// its session: the turn is rebuilt from the request.
func TestClaudeTUIRuntimeLiteParkedBranchNeedsExactlyItsCalls(t *testing.T) {
	b := tuiRuntimeNewSubscriptionBridge()
	t.Cleanup(func() { tuiRuntimeBridges.Delete(b) })
	run := &tuiRuntimeSubscriptionRun{interactive: true, owner: tuiRuntimeScOwner, outerSession: tuiRuntimeScOuter, scope: tuiRuntimeErScope, model: "claude-opus-5-5",
		sessionID: "inner-x", workKey: tuiRuntimeInteractiveSessionStateKey(tuiRuntimeScOwner, tuiRuntimeScOuter), closed: true, parkedAt: time.Now(), shown: []tuiRuntimeShownCall{{id: "call_1"}, {id: "call_2"}}}
	req, err := tuiRuntimeParseRequest([]byte(`{"model":"claude-opus-5-5","messages":[{"role":"user","content":"go"},{"role":"assistant","content":[{"type":"tool_use","id":"call_1","name":"echo","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call_1","content":"x"}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	req.scope = tuiRuntimeErScope
	run.interactiveContextSnapshot = tuiRuntimeTestContextSnapshot(req)
	b.noteLiteParked(run)
	if _, _, _, _, ok := b.liteParkedBranch(tuiRuntimeScOwner, "claude-opus-5-5", "", req); ok {
		t.Fatal("one result of two calls resumed the run's session")
	}
	b.noteLiteParked(run)
	if _, _, _, _, ok := b.liteParkedBranch(tuiRuntimeScOwner, "claude-sonnet-5-5", "", req); ok {
		t.Fatal("another model resumed the run's session")
	}
}

// Stateless runs, one after another, run in the same fixed folder (Claude
// Code names its cwd in the system prompt, the head of the cached prefix)
// and leave no session behind.
func TestClaudeTUIRuntimeLiteStatelessRunsShareAFixedCwdAndLeaveNoSession(t *testing.T) {
	_, runs := tuiRuntimeScWrapper(t)
	rt, _ := tuiRuntimeScRuntime(t)
	// A one-off ask (a title, a topic, a classifier): one user message, no
	// history, no tools. Its cwd comes from the conversation's stable root
	// (its scope plus its first message's digest), so the cwd — part of the
	// cached system prompt — is the same on every turn of a conversation,
	// including a first turn that is a single message. The folder is kept,
	// as a conversation with a history keeps its own for the next turn's
	// native-checkpoint resume.
	body := `{"model":"claude-opus-5-5","max_tokens":64,"messages":[{"role":"user","content":"a"}]}`
	var cwd string
	for i := 0; i < 2; i++ {
		if code, out := tuiRuntimeScServe(t, rt, "", body); code != 200 || !strings.Contains(out, "SC_ANSWER") {
			t.Fatalf("one-off %d: %d %s", i, code, out)
		}
		tuiRuntimeScWaitReaped(t, rt)
		got := runs()
		if len(got) != i+1 {
			t.Fatalf("runs: %v", got)
		}
		if got[i]["cwd"] != cwd && cwd != "" {
			t.Fatalf("the same conversation changed cwd: %v then %v", cwd, got[i]["cwd"])
		}
		cwd = got[i]["cwd"]
		if !strings.Contains(cwd, "magpie-claude-interactive-") {
			t.Fatalf("a one-off did not use the conversation's own cwd: %v", cwd)
		}
	}
	// a different first message is a different conversation, another cwd
	other := `{"model":"claude-opus-5-5","max_tokens":64,"messages":[{"role":"user","content":"b"}]}`
	if code, out := tuiRuntimeScServe(t, rt, "", other); code != 200 {
		t.Fatalf("other: %d %s", code, out)
	}
	tuiRuntimeScWaitReaped(t, rt)
	got := runs()
	if got[len(got)-1]["cwd"] == cwd {
		t.Fatalf("a different first message shared the cwd: %v", cwd)
	}
}

// A no-header conversation with a history gets a cwd of its own, kept
// across turns, so its next turn resumes the native session at the
// strict-prefix checkpoint instead of rebuilding the whole history.
// tuiRuntimeCpWrapper writes real transcripts, so the resumed seed can be
// told apart from a rebuilt one by the first reply's row uuid.
func TestClaudeTUIRuntimeNativeCheckpointResume(t *testing.T) {
	runs := tuiRuntimeCpWrapper(t)
	rt, _ := tuiRuntimeScRuntime(t)
	// no session header (outer ""): the strict prefix alone finds the session
	first := `{"model":"claude-opus-5-5","max_tokens":64,"system":"rules","messages":[{"role":"user","content":"q1"}]}`
	if code, out := tuiRuntimeScServe(t, rt, "", first); code != 200 || !strings.Contains(out, "ANS") {
		t.Fatalf("first turn: %d %s", code, out)
	}
	tuiRuntimeScWaitReaped(t, rt)
	// the next request carries the history and the reply, then a new turn
	next := `{"model":"claude-opus-5-5","max_tokens":64,"system":"rules","messages":[{"role":"user","content":"q1"},{"role":"assistant","content":"ANS1"},{"role":"user","content":"q2"}]}`
	if code, out := tuiRuntimeScServe(t, rt, "", next); code != 200 || !strings.Contains(out, "ANS") {
		t.Fatalf("next turn: %d %s", code, out)
	}
	tuiRuntimeScWaitReaped(t, rt)
	got := runs()
	if len(got) != 2 {
		t.Fatalf("runs: %v", got)
	}
	if !strings.Contains(got[0]["cwd"], "magpie-claude-interactive-") || got[0]["cwd"] != got[1]["cwd"] {
		t.Fatalf("the conversation did not keep one cwd: %v / %v", got[0]["cwd"], got[1]["cwd"])
	}
	// resumed from the checkpoint: the seed is the first turn's own
	// transcript, holding its reply's row uuid a-1; a session rebuilt from
	// this request would have new row uuids instead
	if got[1]["resume"] == "" || !strings.Contains(got[1]["seed.jsonl"], `"uuid":"a-1"`) {
		t.Fatalf("the next turn was not resumed from the first's transcript: resume=%q seed=%q", got[1]["resume"], got[1]["seed.jsonl"])
	}
	// a changed history does not resume that transcript: the reply differs,
	// so the prefix differs and the session is rebuilt from this request's
	// own history — its seed has no a-1 row
	edited := `{"model":"claude-opus-5-5","max_tokens":64,"system":"rules","messages":[{"role":"user","content":"q1"},{"role":"assistant","content":"DIFFERENT"},{"role":"user","content":"q2"}]}`
	if code, out := tuiRuntimeScServe(t, rt, "", edited); code != 200 {
		t.Fatalf("edited turn: %d %s", code, out)
	}
	tuiRuntimeScWaitReaped(t, rt)
	got = runs()
	if len(got) != 3 || strings.Contains(got[2]["seed.jsonl"], `"uuid":"a-1"`) {
		t.Fatalf("an edited history resumed a transcript it does not extend: %+v", got[2:])
	}
}

// A no-header compaction (tools=[]) resumes its native checkpoint and holds
// the session's original tool definitions byte for byte, so the cached
// prefix still matches, and is told not to call them. A call of a held tool
// is answered as not executed, never shown to the client.
func TestClaudeTUIRuntimeNativeCheckpointCompactionHoldsTools(t *testing.T) {
	runs := tuiRuntimeCpWrapper(t)
	t.Setenv("SC_PREFIX", tuiRuntimeMcpToolPrefix())
	rt, b := tuiRuntimeScRuntime(t)
	withTools := func(body string) string {
		return strings.Replace(body, `"system":"rules"`, `"system":"rules","tools":[{"name":"exec_command","description":"run","input_schema":{"type":"object","properties":{"cmd":{"type":"string"}}}}]`, 1)
	}
	serve := func(body, last string) map[string]string {
		t.Helper()
		if code, out := tuiRuntimeScServe(t, rt, "", body); code != 200 || !strings.Contains(out, "ANS") {
			t.Fatalf("%s: %d %s", last, code, out)
		}
		tuiRuntimeScWaitReaped(t, rt)
		for _, run := range runs() {
			if strings.Contains(run["stdin"], last) {
				return run
			}
		}
		t.Fatalf("no run for %s", last)
		return nil
	}
	first := serve(withTools(tuiRuntimeCpBody("q1")), "q1")
	if !strings.Contains(first["tools.json"], "exec") {
		t.Fatalf("first run's tools: %q", first["tools.json"])
	}
	// a no-header compaction: no tools, but the same conversation prefix
	compact := serve(tuiRuntimeCpBody("q1", "ANS1", "summarize"), "summarize")
	if compact["resume"] == "" || compact["tools.json"] != first["tools.json"] {
		t.Fatalf("no-header compaction did not hold the session's definitions: resume=%q tools=%q", compact["resume"], compact["tools.json"])
	}
	if !strings.Contains(compact["stdin"], "No tools are available for this request") {
		t.Fatalf("compaction not told it has no tools: %q", compact["stdin"])
	}
	// a call of a held tool is not shown to the client
	code, out := tuiRuntimeScServe(t, rt, "", tuiRuntimeCpBody("q1", "ANS1", "summarize", "ANS2", "again CALLHELD"))
	if code != 200 || !strings.Contains(out, "ANS3") || strings.Contains(out, "tool_use") || strings.Contains(out, "held_1") {
		t.Fatalf("held call reached the client: %d %s", code, out)
	}
	b.abortAll()
}

// A turn that sends the conversation's tools but may call none of them
// (tool_choice none: Grok's session title, au-exit 10-29, ~667k written to
// the cache in three titles) keeps the session's definitions byte for byte
// and is told not to call them; a call of one never reaches the client. Its
// checkpoint keeps the definitions' hash, and the conversation going on with
// tool_choice auto has its tools offered as before. tool_choice naming one
// tool keeps the others defined, held.
func TestClaudeTUIRuntimeNativeCheckpointToolChoiceHoldsTools(t *testing.T) {
	runs := tuiRuntimeCpWrapper(t)
	t.Setenv("SC_PREFIX", tuiRuntimeMcpToolPrefix())
	rt, _ := tuiRuntimeScRuntime(t)
	withTools := func(body, choice string) string {
		tools := `"tools":[{"name":"exec_command","description":"run","input_schema":{"type":"object","properties":{"cmd":{"type":"string"}}}},{"name":"read_file","description":"read","input_schema":{"type":"object","properties":{"path":{"type":"string"}}}}]`
		if choice != "" {
			tools += `,"tool_choice":` + choice
		}
		return strings.Replace(body, `"system":"rules"`, `"system":"rules",`+tools, 1)
	}
	serve := func(body, last string) (map[string]string, string) {
		t.Helper()
		code, out := tuiRuntimeScServe(t, rt, "", body)
		if code != 200 || !strings.Contains(out, "ANS") {
			t.Fatalf("%s: %d %s", last, code, out)
		}
		tuiRuntimeScWaitReaped(t, rt)
		for _, run := range runs() {
			if strings.Contains(run["stdin"], last) {
				return run, out
			}
		}
		t.Fatalf("no run for %s", last)
		return nil, ""
	}
	first, _ := serve(withTools(tuiRuntimeCpBody("q1"), ""), "q1")
	if !strings.Contains(first["tools.json"], "exec_command") || !strings.Contains(first["tools.json"], "read_file") {
		t.Fatalf("first run's tools: %q", first["tools.json"])
	}
	title, _ := serve(withTools(tuiRuntimeCpBody("q1", "ANS1", "TITLE"), `{"type":"none"}`), "TITLE")
	if title["resume"] == "" || title["tools.json"] != first["tools.json"] {
		t.Fatalf("tool_choice none did not keep the session's definitions: resume=%q tools=%q", title["resume"], title["tools.json"])
	}
	if !strings.Contains(title["stdin"], "No tools are available for this request") {
		t.Fatalf("tool_choice none turn not told: %q", title["stdin"])
	}
	// the title's checkpoint names the definitions: a later tool-less turn
	// from it holds them too
	after, _ := serve(tuiRuntimeCpBody("q1", "ANS1", "TITLE", "ANS2", "after title"), "after title")
	if after["resume"] == "" || after["tools.json"] != first["tools.json"] {
		t.Fatalf("the title's checkpoint lost the definitions: resume=%q tools=%q", after["resume"], after["tools.json"])
	}
	// a call the turn makes anyway is answered as not executed
	if _, out := serve(withTools(tuiRuntimeCpBody("q1", "ANS1", "RETITLE CALLHELD"), `{"type":"none"}`), "RETITLE"); strings.Contains(out, "tool_use") || strings.Contains(out, "held_1") {
		t.Fatalf("held call reached the client: %s", out)
	}
	// the conversation itself goes on with its tools offered
	work, _ := serve(withTools(tuiRuntimeCpBody("q1", "ANS1", "work"), `{"type":"auto"}`), "work")
	if work["resume"] == "" || work["tools.json"] != first["tools.json"] || strings.Contains(work["stdin"], "No tools are available") {
		t.Fatalf("auto turn after the title: %v", work)
	}
	// naming one tool keeps the other defined; a call of it is not shown
	named, out := serve(withTools(tuiRuntimeCpBody("q1", "ANS1", "NAMED CALLHELD"), `{"type":"tool","name":"read_file"}`), "NAMED")
	if named["tools.json"] != first["tools.json"] || !strings.Contains(named["stdin"], "You must call the read_file tool") {
		t.Fatalf("tool_choice naming a tool: %v", named)
	}
	if strings.Contains(out, "held_1") {
		t.Fatalf("a call of a tool tool_choice ruled out reached the client: %s", out)
	}
}

// A tool tool_choice rules out is defined but not offered: its MCP call is
// answered as not executed and never registered for the client, and a
// tool_choice none request is not typed into an idle process that offers
// the tools (their launches differ), nor is its own process reused.
func TestClaudeTUIRuntimeToolChoiceRuledOutToolIsHeld(t *testing.T) {
	_, b := tuiRuntimeScRuntime(t)
	tools := []Tool{{Name: "exec_command"}, {Name: "read_file"}}
	none := &tuiRuntimeRequest{Request: Request{Tools: tools, ToolChoice: "none"}}
	auto := &tuiRuntimeRequest{Request: Request{Tools: tools}}
	named := &tuiRuntimeRequest{Request: Request{Tools: tools, ToolChoice: "name:read_file"}}
	run := &tuiRuntimeSubscriptionRun{bridge: b, interactive: true, pending: map[string]chan tuiRuntimeMcpToolResult{}}
	run.offerLocked(named, false)
	if run.tools["exec_command"] || !run.tools["read_file"] || !run.heldTools["exec_command"] || !run.offers(tools) {
		t.Fatalf("name:read_file: offered %v held %v", run.tools, run.heldTools)
	}
	run.offerLocked(none, false)
	if len(run.tools) != 0 || !run.heldTools["read_file"] {
		t.Fatalf("none: offered %v held %v", run.tools, run.heldTools)
	}
	b.mu.Lock()
	b.runs["choice-token"] = run
	b.mu.Unlock()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/_magpie/claude-mcp/choice-token", strings.NewReader(`{"tool_call_id":"c1","name":"exec_command","arguments":{}}`))
	req.SetPathValue("token", "choice-token")
	b.mcpCall(rec, req)
	if !strings.Contains(rec.Body.String(), "NOT executed") || !strings.Contains(rec.Body.String(), `"is_error":true`) {
		t.Fatalf("ruled-out tool call: %s", rec.Body.String())
	}
	b.mu.Lock()
	_, registered := b.calls["c1"]
	b.mu.Unlock()
	if registered || len(run.pending) != 0 {
		t.Fatal("ruled-out tool call was registered for the client")
	}

	if a, n := tuiRuntimeHotConfigOf(auto, "m", "", "", false), tuiRuntimeHotConfigOf(none, "m", "", "", false); a.tools == n.tools {
		t.Fatal("a tool_choice none launch matches one offering the tools")
	}
	if tuiRuntimeToolChoiceHeld(auto) != nil || len(tuiRuntimeToolChoiceHeld(none)) != 2 {
		t.Fatal("held by tool_choice")
	}
	ad, _ := json.Marshal(tuiRuntimeInteractiveDefinedTools(auto, false))
	nd, _ := json.Marshal(tuiRuntimeInteractiveDefinedTools(none, false))
	if string(ad) != string(nd) {
		t.Fatalf("definitions differ by tool_choice: %s vs %s", ad, nd)
	}
}

// After a compaction replaces the history H with a summary S, S is a new
// prefix: it must not resume the old H session (the client dropped H), and
// it builds its own new checkpoint. A later turn on S resumes S's session.
func TestClaudeTUIRuntimeNativeCheckpointCompactedHistoryIsNewPrefix(t *testing.T) {
	runs := tuiRuntimeCpWrapper(t)
	rt, _ := tuiRuntimeScRuntime(t)
	serve := func(body, last string) map[string]string {
		t.Helper()
		if code, out := tuiRuntimeScServe(t, rt, "", body); code != 200 || !strings.Contains(out, "ANS") {
			t.Fatalf("%s: %d %s", last, code, out)
		}
		tuiRuntimeScWaitReaped(t, rt)
		for _, run := range runs() {
			if strings.Contains(run["stdin"], last) {
				return run
			}
		}
		t.Fatalf("no run for %s", last)
		return nil
	}
	// the original conversation: q1 -> ANS1
	serve(tuiRuntimeCpBody("q1"), "q1")
	// a later turn on the original history: q1,ANS1,q2
	onOriginal := serve(tuiRuntimeCpBody("q1", "ANS1", "q2"), "q2")
	if !strings.Contains(onOriginal["seed.jsonl"], `"uuid":"a-1"`) {
		t.Fatalf("the original history did not resume its checkpoint: %q", onOriginal["seed.jsonl"])
	}
	// the client compacts: H (q1,ANS1,...) is replaced by a summary S. The
	// first message is now the summary, a new prefix — it must not resume
	// the old session (which would reintroduce H). The summary turn carries
	// the summary and its reply, then a new question.
	summary := serve(`{"model":"claude-opus-5-5","max_tokens":64,"system":"rules","messages":[{"role":"user","content":"SUMMARY of q1 and q2"},{"role":"assistant","content":"ANS9"},{"role":"user","content":"continue"}]}`, "continue")
	if strings.Contains(summary["seed.jsonl"], `"uuid":"a-1"`) {
		t.Fatalf("a compacted history resumed the dropped old session: %q", summary["seed.jsonl"])
	}
	// a later turn on the summary resumes the summary's own new session
	onSummary := serve(`{"model":"claude-opus-5-5","max_tokens":64,"system":"rules","messages":[{"role":"user","content":"SUMMARY of q1 and q2"},{"role":"assistant","content":"ANS9"},{"role":"user","content":"continue"},{"role":"assistant","content":"ANS10"},{"role":"user","content":"next"}]}`, "next")
	if onSummary["resume"] == "" || strings.Contains(onSummary["seed.jsonl"], `"uuid":"a-1"`) {
		t.Fatalf("the summary's next turn did not resume its own session: %q", onSummary["seed.jsonl"])
	}
}

// A checkpoint is resumed only at the request's last assistant reply, where
// what follows is a fresh turn: a deeper reply the client has since branched
// away from is not its continuation, a different model is not resumed, and
// a suffix answering a tool call is not a fresh turn.
func TestClaudeTUIRuntimeNativeCheckpointMatchRules(t *testing.T) {
	dir := t.TempDir()
	old := tuiRuntimeSettings
	t.Cleanup(func() { tuiRuntimeSettings = old })
	initTUIRuntime(&tuiRuntimeConfig{StateDir: dir})
	scope := "scope"
	// the conversation as it was when the checkpoint was written
	full := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "u1"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "a1"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "u2"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "a2"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "u3"}}},
	}
	digests := tuiRuntimePrefixDigests(scope, full)
	save := func(at int, model string) {
		t.Helper()
		if err := tuiRuntimeSaveNativeCheckpoint(tuiRuntimeNativeCheckpoint{
			Prefix: tuiRuntimeNativeCheckpointKey(scope, digests[at])[len(scope)+1:], Scope: scope,
			SessionID: "sid", Row: "r", Work: "w", Model: model, InstructionKey: "i",
		}); err != nil {
			t.Fatal(err)
		}
	}
	req := func(msgs []Message) *tuiRuntimeRequest {
		r := &tuiRuntimeRequest{}
		r.scope = scope
		r.Messages = msgs
		return r
	}
	// checkpoints at a1 and a2
	save(1, "claude-opus-5-5")
	save(3, "claude-opus-5-5")
	// the client branched after a1: its history is u1,a1,u3 (no a2). The
	// last assistant is a1, whose suffix u3 is a fresh turn: resumed there,
	// not rebuilt, even though a deeper a2 checkpoint exists.
	branch := []Message{full[0], full[1], {Role: "user", Parts: []Part{{Kind: Text, Text: "u3"}}}}
	point, since, ok := tuiRuntimeFindNativeCheckpoint(req(branch), "claude-opus-5-5")
	if !ok || len(since) != 1 || since[0].Parts[0].Text != "u3" {
		t.Fatalf("the branch after a1 was not resumed at a1: ok=%v since=%v", ok, since)
	}
	_ = point
	// the same prefix but a different model is not resumed
	if _, _, ok := tuiRuntimeFindNativeCheckpoint(req(branch), "claude-haiku-5-5"); ok {
		t.Fatal("resumed a checkpoint of another model")
	}
	// a suffix answering a tool call is not a fresh turn
	withResult := []Message{full[0], full[1], {Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "c1", Text: "out"}}}}
	if _, _, ok := tuiRuntimeFindNativeCheckpoint(req(withResult), "claude-opus-5-5"); ok {
		t.Fatal("resumed into a suffix that answers a tool call")
	}
	// a history whose last assistant has no checkpoint is not resumed from a
	// shallower one either: u1,a1,u2,a2,u3 ends at a2, and a2's suffix u3 is
	// a fresh turn, but a history edited so a2 changed matches no checkpoint
	edited := []Message{full[0], full[1], full[2], {Role: "assistant", Parts: []Part{{Kind: Text, Text: "a2-edited"}}}, full[4]}
	if _, _, ok := tuiRuntimeFindNativeCheckpoint(req(edited), "claude-opus-5-5"); ok {
		t.Fatal("resumed a checkpoint whose reply the client rewrote")
	}
}

// A /btw side question while the main task has gone past the reply it asks
// about forks the native session at that reply, in a copy — the main task's
// own transcript is not truncated or stopped. Both go on side by side, with
// no session header.
func TestClaudeTUIRuntimeNativeCheckpointSideQuestionForks(t *testing.T) {
	runs := tuiRuntimeCpWrapper(t)
	rt, _ := tuiRuntimeScRuntime(t)
	serve := func(body, last string) map[string]string {
		t.Helper()
		if code, out := tuiRuntimeScServe(t, rt, "", body); code != 200 || !strings.Contains(out, "ANS") {
			t.Fatalf("%s: %d %s", last, code, out)
		}
		tuiRuntimeScWaitReaped(t, rt)
		for _, run := range runs() {
			if strings.Contains(run["stdin"], last) {
				return run
			}
		}
		t.Fatalf("no run for %s", last)
		return nil
	}
	// the main line: q1 -> ANS1, then q1,ANS1,q2 -> ANS2 (the session goes on)
	serve(tuiRuntimeCpBody("q1"), "q1")
	main := serve(tuiRuntimeCpBody("q1", "ANS1", "q2"), "q2")
	mainSession := main["resume"]
	// a /btw side question from the first reply: q1,ANS1,q3. The session has
	// gone on to ANS2 since, so it is forked at ANS1, in a copy — a different
	// session from the main line, still holding a-1 and not a-2.
	side := serve(tuiRuntimeCpBody("q1", "ANS1", "q3"), "q3")
	if side["resume"] == "" || side["resume"] == mainSession {
		t.Fatalf("the side question did not fork into its own session: main=%q side=%q", mainSession, side["resume"])
	}
	if !strings.Contains(side["seed.jsonl"], `"uuid":"a-1"`) || strings.Contains(side["seed.jsonl"], `"uuid":"a-2"`) {
		t.Fatalf("the fork is not the session up to the first reply:\n%s", side["seed.jsonl"])
	}
	// the main line goes on in its own session, unchanged
	again := serve(tuiRuntimeCpBody("q1", "ANS1", "q2", "ANS2", "q4"), "q4")
	if again["resume"] != mainSession {
		t.Fatalf("the main line left its session after the fork: %q want %q", again["resume"], mainSession)
	}
	// and the side line goes on in its fork
	sideAgain := serve(tuiRuntimeCpBody("q1", "ANS1", "q3", "ANS3", "q5"), "q5")
	if sideAgain["resume"] != side["resume"] {
		t.Fatalf("the side line left its fork: %q want %q", sideAgain["resume"], side["resume"])
	}
}

// A native-checkpoint store that doesn't read is unknown, not empty:
// nothing resumes from it, and saving never writes over it.
func TestClaudeTUIRuntimeNativeCheckpointCorruptStore(t *testing.T) {
	dir := t.TempDir()
	old := tuiRuntimeSettings
	t.Cleanup(func() { tuiRuntimeSettings = old })
	initTUIRuntime(&tuiRuntimeConfig{StateDir: dir})
	if err := os.WriteFile(tuiRuntimeNativeCheckpointPath(), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := tuiRuntimeLoadNativeCheckpoints(); err == nil {
		t.Fatal("a corrupt store was read as valid")
	}
	req := &tuiRuntimeRequest{}
	req.scope = "scope"
	req.Messages = []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "a"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "b"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "c"}}},
	}
	if _, _, ok := tuiRuntimeFindNativeCheckpoint(req, "m"); ok {
		t.Fatal("resumed from a corrupt store")
	}
	if err := tuiRuntimeSaveNativeCheckpoint(tuiRuntimeNativeCheckpoint{Prefix: "p", Scope: "s", SessionID: "sid", Row: "r", Work: "w", InstructionKey: "i"}); err == nil {
		t.Fatal("saved over a corrupt store")
	}
	data, _ := os.ReadFile(tuiRuntimeNativeCheckpointPath())
	if string(data) != "{not json" {
		t.Fatalf("corrupt store was overwritten: %q", data)
	}
}

// A resume finds its checkpoint while another turn's save holds the store's
// lock (as it does across that save's disk sync): readers never wait on it.
func TestClaudeTUIRuntimeNativeCheckpointFindSkipsSaveLock(t *testing.T) {
	dir := t.TempDir()
	old := tuiRuntimeSettings
	t.Cleanup(func() { tuiRuntimeSettings = old })
	initTUIRuntime(&tuiRuntimeConfig{StateDir: dir})
	scope := "scope"
	msgs := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "u1"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "a1"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "u2"}}},
	}
	digests := tuiRuntimePrefixDigests(scope, msgs)
	if err := tuiRuntimeSaveNativeCheckpoint(tuiRuntimeNativeCheckpoint{
		Prefix: tuiRuntimeNativeCheckpointKey(scope, digests[1])[len(scope)+1:], Scope: scope,
		SessionID: "sid", Row: "r", Work: "w", InstructionKey: "i",
	}); err != nil {
		t.Fatal(err)
	}
	req := &tuiRuntimeRequest{}
	req.scope = scope
	req.Messages = msgs
	tuiRuntimeNativeCheckpointMu.Lock()
	defer tuiRuntimeNativeCheckpointMu.Unlock()
	found := make(chan bool, 1)
	go func() {
		_, _, ok := tuiRuntimeFindNativeCheckpoint(req, "m")
		found <- ok
	}()
	select {
	case ok := <-found:
		if !ok {
			t.Fatal("the saved checkpoint was not found")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("finding a checkpoint waited on a save's lock")
	}
}

// scWaitReaped waits for every run's process to be reaped.
func tuiRuntimeScWaitReaped(t *testing.T, rt *tuiRuntime) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		rt.bridge.mu.Lock()
		n := len(rt.bridge.runs)
		rt.bridge.mu.Unlock()
		if n == 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("%d runs not reaped", n)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// A long conversation full of screenshots is an ordinary request up to
// Anthropic's 32 MiB; past it the client gets a 413 it can act on (compact).
func TestClaudeTUIRuntimeLiteSignedRequestSizeLimit(t *testing.T) {
	rt, _ := tuiRuntimeScRuntime(t)
	post := func(size int) (int, string) {
		rec := httptest.NewRecorder()
		rt.serveSignedMessages(rec, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(strings.Repeat("x", size))))
		return rec.Code, rec.Body.String()
	}
	if code, out := post(20 << 20); code != http.StatusUnauthorized {
		t.Fatalf("a 20 MiB request was not read through to authentication: %d %.200s", code, out)
	}
	code, out := post(32<<20 + 1)
	if code != http.StatusRequestEntityTooLarge || !strings.Contains(out, `"request_too_large"`) {
		t.Fatalf("an oversized request answered %d %.200s", code, out)
	}
}

func TestClaudeTUIRuntimeLiteSubagentSessionIdentity(t *testing.T) {
	header := http.Header{}
	key := func(outer, agent, parent string) string {
		header.Set("x-claude-code-agent-id", agent)
		header.Set("x-claude-code-parent-agent-id", parent)
		return tuiRuntimeLiteOuterSession(outer, header)
	}
	first := key("conversation", "explorer", "")
	if first == "" || first == "conversation" {
		t.Fatal("subagent shares parent identity")
	}
	if key("conversation", "explorer", "parent-agent") != first {
		t.Fatal("optional parent header changed identity")
	}
	for _, other := range []string{key("conversation", "reviewer", ""), key("other-conversation", "explorer", ""), key("ab", "c", ""), key("a", "bc", "")} {
		if other == first {
			t.Fatal("different subagents share identity")
		}
	}
	if key("ab", "c", "") == key("a", "bc", "") {
		t.Fatal("ambiguous field boundaries")
	}
	if key("", "explorer", "") != "" {
		t.Fatal("missing conversation became stateful")
	}
	if key("conversation", "", "") != "conversation" {
		t.Fatal("lead identity changed")
	}
}
