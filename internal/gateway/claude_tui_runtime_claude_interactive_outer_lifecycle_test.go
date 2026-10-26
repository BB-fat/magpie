package gateway

// DAV-1878 regression tests for canonical-outer PTY exclusivity.
//
// A canonical interactive run owns its (owner, outerSession) from the
// moment it is published — while the starting request still holds the
// outerTurns claim — until its child is reaped and the run is removed.
// Any non-matching request on the same outer must fail closed (409)
// instead of starting a second PTY; an exact tool-result continuation
// still joins the live run; helpers (outerSession ""), other owners and
// other outer sessions are unaffected.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const tuiRuntimeOlModel = "claude-opus-5-5"

func tuiRuntimeOlDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	return dir
}

func tuiRuntimeOlWrapper(t *testing.T, dir, name, script string) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, p)
}

// olAnswerScript answers one complete text turn; extra (a for-loop gate)
// can delay it until a release file appears.
func tuiRuntimeOlAnswerScript(extra string) string {
	return "#!/bin/sh\n" +
		"cat >/dev/null\n" + extra +
		"printf '%s\\n' '{\"type\":\"assistant\",\"uuid\":\"a1\",\"session_id\":\"inner-ol\",\"message\":{\"id\":\"m1\",\"model\":\"claude-opus-5-5\",\"content\":[{\"type\":\"text\",\"text\":\"OL_ANSWER\"}]}}'\n" +
		"printf '%s\\n' '{\"type\":\"result\",\"session_id\":\"inner-ol\",\"stop_reason\":\"end_turn\"}'\n"
}

// olParkScript emits one client tool_use row, parks until release appears,
// then answers in text. DAV1878_PARK_RELEASE points at the release file.
func tuiRuntimeOlParkScript() string {
	return "#!/bin/sh\n" +
		"cat >/dev/null\n" +
		"printf '%s\\n' '{\"type\":\"assistant\",\"uuid\":\"t1\",\"session_id\":\"inner-ol\",\"message\":{\"id\":\"m1\",\"model\":\"claude-opus-5-5\",\"content\":[{\"type\":\"tool_use\",\"id\":\"call_1\",\"name\":\"mcp__magpie__write_file\",\"input\":{\"path\":\"a\"}}]}}'\n" +
		"for i in $(seq 1 400); do [ -f \"$DAV1878_PARK_RELEASE\" ] && break; sleep 0.05; done\n" +
		"printf '%s\\n' '{\"type\":\"assistant\",\"uuid\":\"a2\",\"session_id\":\"inner-ol\",\"message\":{\"id\":\"m2\",\"model\":\"claude-opus-5-5\",\"content\":[{\"type\":\"text\",\"text\":\"PARKED_ANSWER\"}]}}'\n" +
		"printf '%s\\n' '{\"type\":\"result\",\"session_id\":\"inner-ol\",\"stop_reason\":\"end_turn\"}'\n"
}

func tuiRuntimeOlServe(t *testing.T, rt *tuiRuntime, b *tuiRuntimeSubscriptionBridge, owner, outer string, body []byte) (int, string, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body)))
	code, msg := rt.serveMessagesInner(tuiRuntimeAnthropicResponder{rec, req, &Usage{}}, req, "Claude Code", tuiRuntimeOlModel, owner, outer, body, func(c context.Context, r *Request) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
		return b.startInteractiveRequest(c, r, tuiRuntimeOlModel, "", owner, outer)
	})
	return code, msg, rec.Body.String()
}

func tuiRuntimeOlSave(t *testing.T, b *tuiRuntimeSubscriptionBridge, owner, outer, session string, dirty bool) {
	t.Helper()
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "先读一下配置"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "配置已读"}}},
	}
	baseReq := &Request{Model: tuiRuntimeOlModel, Messages: base}
	keys := tuiRuntimeConversationKeys(owner, base)
	b.saveInteractiveSession(owner, outer, tuiRuntimeInteractiveSessionEntry{
		SessionID: session, Model: tuiRuntimeOlModel, ThreadKey: firstWords(baseReq),
		ConvKey: keys[0], ReplyKey: tuiRuntimeAssistantReplyKey(base[1]),
		ContextKey: tuiRuntimeTurnKey(owner, baseReq, nil), Dirty: dirty,
	})
}

func tuiRuntimeOlLiveOuterRuns(b *tuiRuntimeSubscriptionBridge, owner, outer string) (live, same int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, r := range b.runs {
		r.mu.Lock()
		if !r.closed {
			live++
			if r.owner == owner && r.outerSession == outer {
				same++
			}
		}
		r.mu.Unlock()
	}
	return live, same
}

func tuiRuntimeOlWaitReaped(t *testing.T, run *tuiRuntimeSubscriptionRun) {
	t.Helper()
	select {
	case <-run.interactiveDone:
	case <-time.After(15 * time.Second):
		t.Fatal("the interactive child was not reaped")
	}
}

// The parked tool turn the whole file builds on: first request parks on
// call_1 with its HTTP reply complete.
func tuiRuntimeOlParkedFirstTurn(t *testing.T, dir string, rt *tuiRuntime, b *tuiRuntimeSubscriptionBridge, owner, outer string) *tuiRuntimeSubscriptionRun {
	t.Helper()
	release := filepath.Join(dir, "release-"+outer)
	t.Setenv("DAV1878_PARK_RELEASE", release)
	first := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"write_file","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"把结果写进 notes.txt"}]}`)
	code, msg, out := tuiRuntimeOlServe(t, rt, b, owner, outer, first)
	if code != 200 || !strings.Contains(out, `"type":"tool_use"`) {
		t.Fatalf("first turn did not park on a tool call: %d %s %s", code, msg, out)
	}
	var parked *tuiRuntimeSubscriptionRun
	b.mu.Lock()
	for _, r := range b.runs {
		parked = r
	}
	b.mu.Unlock()
	if parked == nil {
		t.Fatal("no live run after the first turn parked")
	}
	parked.mu.Lock()
	closed, at := parked.closed, parked.parkedAt
	parked.mu.Unlock()
	if closed || at.IsZero() {
		t.Fatalf("the first-turn run is not parked alive (closed=%v parkedAt=%v)", closed, at)
	}
	if _, same := tuiRuntimeOlLiveOuterRuns(b, owner, outer); same != 1 {
		t.Fatalf("%d live runs on the outer after park (want 1)", same)
	}
	// The HTTP claim is free again — this is exactly the window DAV-1872's
	// I9b exploited.
	if !b.claimInteractiveOuterTurn(owner, outer) {
		t.Fatal("the per-request claim is still held after the reply ended")
	}
	b.releaseInteractiveOuterTurn(owner, outer)
	return parked
}

func tuiRuntimeOlRelease(parked *tuiRuntimeSubscriptionRun, dir, outer string) {
	_ = parked
	_ = os.WriteFile(filepath.Join(dir, "release-"+outer), []byte("ok"), 0o600)
}

// OL1 is the DAV-1872 I9b payload, now asserting the fixed behaviour: a
// same-outer request without a tool result while the first turn is parked
// on an unanswered tool_use is refused 409 and starts no second PTY.
func TestClaudeTUIRuntimeOuterLiveParkedFirstTurnRefusesSecondStart(t *testing.T) {
	dir := tuiRuntimeOlDir(t)
	tuiRuntimeOlWrapper(t, dir, "ol1-bridge", tuiRuntimeOlParkScript())
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	rt := &tuiRuntime{bridge: b}
	owner, outer := "owner-ol1", "outer-ol1"
	parked := tuiRuntimeOlParkedFirstTurn(t, dir, rt, b, owner, outer)

	// The I9b duplicate: same owner, same outer, same opening words, no
	// tool_result — nothing a live run matches.
	second := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"write_file","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"把结果写进 notes.txt"}]}`)
	code, msg, _ := tuiRuntimeOlServe(t, rt, b, owner, outer, second)
	live, same := tuiRuntimeOlLiveOuterRuns(b, owner, outer)
	t.Logf("OL1: second request = %d %q; live=%d same-outer=%d", code, msg, live, same)
	if code != http.StatusConflict {
		t.Fatalf("OL1: a duplicate on an outer with a live parked run was not refused: %d %s", code, msg)
	}
	if same != 1 {
		t.Fatalf("OL1: %d live PTY runs share outer %q (want the parked one only)", same, outer)
	}

	// Releasing the parked child ends the run; only after its child is
	// reaped may a new turn start on the outer.
	tuiRuntimeOlRelease(parked, dir, outer)
	tuiRuntimeOlWaitReaped(t, parked)
	deadline := time.Now().Add(5 * time.Second)
	for {
		_, same = tuiRuntimeOlLiveOuterRuns(b, owner, outer)
		if same == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("OL1: the reaped run still owns the outer")
		}
		time.Sleep(5 * time.Millisecond)
	}
	tuiRuntimeOlWrapper(t, dir, "ol1b-bridge", tuiRuntimeOlAnswerScript(""))
	code, msg, out := tuiRuntimeOlServe(t, rt, b, owner, outer, second)
	b.abortAll()
	if code != 200 {
		t.Fatalf("OL1: the outer stayed locked after the parked child ended: %d %s", code, msg)
	}
	if !strings.Contains(out, "OL_ANSWER") {
		t.Fatalf("OL1: the post-release turn did not answer: %s", out)
	}
}

// A legal exact-ID tool continuation joins the parked run and starts no
// new PTY, even though the live-outer gate now refuses unmatched starts.
func TestClaudeTUIRuntimeOuterLiveExactIDContinuationJoinsParkedRun(t *testing.T) {
	dir := tuiRuntimeOlDir(t)
	tuiRuntimeOlWrapper(t, dir, "ol2-bridge", tuiRuntimeOlParkScript())
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	rt := &tuiRuntime{bridge: b}
	owner, outer := "owner-ol2", "outer-ol2"
	parked := tuiRuntimeOlParkedFirstTurn(t, dir, rt, b, owner, outer)

	// The real continuation path needs the MCP helper's callback to have
	// registered the live call first (b.calls[callID]). Simulate the
	// genuine Claude MCP helper posting its tool call.
	go func() {
		w := httptest.NewRecorder()
		mr := httptest.NewRequest("POST", "/_magpie/claude-mcp/"+parked.token,
			strings.NewReader(`{"tool_call_id":"call_1","name":"write_file","arguments":{"path":"a"}}`))
		mr = mr.WithContext(context.WithValue(mr.Context(), tuiRuntimeMcpTokenKey{}, parked.token))
		b.mcpCall(w, mr)
	}()
	deadline := time.Now().Add(3 * time.Second)
	for {
		b.mu.Lock()
		waiting := b.calls["call_1"] == parked
		b.mu.Unlock()
		if waiting {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("OL2: the MCP callback never registered")
		}
		time.Sleep(5 * time.Millisecond)
	}

	// The client's tool result for the live call_1, sent on the canonical
	// wire shape: original user message, the assistant tool_use row, then
	// the tool_result. Exact CallID ownership routes it to the parked run.
	cont := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"write_file","input_schema":{"type":"object"}}],"messages":[` +
		`{"role":"user","content":"把结果写进 notes.txt"},` +
		`{"role":"assistant","content":[{"type":"tool_use","id":"call_1","name":"write_file","input":{"path":"a"}}]},` +
		`{"role":"user","content":[{"type":"tool_result","tool_use_id":"call_1","content":"written ok"}]}]}`)
	done := make(chan struct{})
	var code int
	var msg, out string
	go func() {
		defer close(done)
		code, msg, out = tuiRuntimeOlServe(t, rt, b, owner, outer, cont)
	}()
	// The continuation must reach the parked run's MCP callback before its
	// reply arrives. Give the bridge a moment to route it, then release
	// the parked child's pending answer.
	time.Sleep(300 * time.Millisecond)
	tuiRuntimeOlRelease(parked, dir, outer)
	<-done
	b.abortAll()
	t.Logf("OL2: continuation = %d %q body=%s", code, msg, out)
	if code != 200 {
		t.Fatalf("OL2: an exact-ID tool continuation was refused while its run was parked: %d %s", code, msg)
	}
	if !strings.Contains(out, "PARKED_ANSWER") {
		t.Fatalf("OL2: the continuation did not produce the parked run's answer: %s", out)
	}
}

// A first-turn duplicate that arrives while the first turn's reply is
// still in flight is refused by the in-flight claim; once that reply is
// done and the run fully ended, a retry on the same outer succeeds — the
// outer is never locked permanently.
func TestClaudeTUIRuntimeOuterLiveInFlightThenRetryAfterFinish(t *testing.T) {
	dir := tuiRuntimeOlDir(t)
	release := filepath.Join(dir, "release-inflight")
	t.Setenv("DAV1878_PARK_RELEASE", release)
	slow := "#!/bin/sh\ncat >/dev/null\n" +
		"for i in $(seq 1 400); do [ -f \"$DAV1878_PARK_RELEASE\" ] && break; sleep 0.05; done\n" +
		"printf '%s\\n' '{\"type\":\"assistant\",\"uuid\":\"a1\",\"session_id\":\"inner-ol\",\"message\":{\"id\":\"m1\",\"model\":\"claude-opus-5-5\",\"content\":[{\"type\":\"text\",\"text\":\"OL_ANSWER\"}]}}'\n" +
		"printf '%s\\n' '{\"type\":\"result\",\"session_id\":\"inner-ol\",\"stop_reason\":\"end_turn\"}'\n"
	tuiRuntimeOlWrapper(t, dir, "ol3-bridge", slow)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	rt := &tuiRuntime{bridge: b}
	owner, outer := "owner-ol3", "outer-ol3"
	body := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"write_file","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"把结果写进 notes.txt"}]}`)

	var starts atomic.Int64
	start := func(c context.Context, r *Request) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
		run, ev, err := b.startInteractiveRequest(c, r, tuiRuntimeOlModel, "", owner, outer)
		if err == nil && run != nil {
			starts.Add(1)
		}
		return run, ev, err
	}
	firstDone := make(chan struct{})
	var firstCode int
	go func() {
		defer close(firstDone)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body)))
		firstCode, _ = rt.serveMessagesInner(tuiRuntimeAnthropicResponder{rec, req, &Usage{}}, req, "Claude Code", tuiRuntimeOlModel, owner, outer, body, start)
	}()
	deadline := time.Now().Add(5 * time.Second)
	for starts.Load() == 0 {
		if time.Now().After(deadline) {
			t.Fatal("the in-flight turn never started")
		}
		time.Sleep(5 * time.Millisecond)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body)))
	code, msg := rt.serveMessagesInner(tuiRuntimeAnthropicResponder{rec, req, &Usage{}}, req, "Claude Code", tuiRuntimeOlModel, owner, outer, body, start)
	if code != http.StatusConflict {
		t.Fatalf("a concurrent duplicate was not refused: %d %s", code, msg)
	}
	_ = os.WriteFile(release, []byte("ok"), 0o600)
	<-firstDone
	if firstCode != 200 {
		t.Fatalf("the in-flight turn failed: %d", firstCode)
	}
	// The completed run's child is reaped asynchronously; wait for the
	// outer to be free, then a genuine next turn — the saved checkpoint's
	// conversation plus a new user message — must start a fresh PTY.
	deadline = time.Now().Add(5 * time.Second)
	for {
		_, same := tuiRuntimeOlLiveOuterRuns(b, owner, outer)
		if same == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the completed run still owns the outer")
		}
		time.Sleep(5 * time.Millisecond)
	}
	tuiRuntimeOlWrapper(t, dir, "ol3b-bridge", tuiRuntimeOlAnswerScript(""))
	next := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"write_file","input_schema":{"type":"object"}}],"messages":[` +
		`{"role":"user","content":"把结果写进 notes.txt"},` +
		`{"role":"assistant","content":"OL_ANSWER"},` +
		`{"role":"user","content":"再说一句"}]}`)
	code, msg, out := tuiRuntimeOlServe(t, rt, b, owner, outer, next)
	b.abortAll()
	if code != 200 || !strings.Contains(out, "OL_ANSWER") {
		t.Fatalf("the post-completion next turn did not get a fresh PTY: %d %s %s", code, msg, out)
	}
}

// Live-outer exclusivity is scoped to (owner, outerSession): a different
// outer on the same owner, and the same outer under a different owner,
// run in parallel undisturbed.
func TestClaudeTUIRuntimeOuterLiveScopeDifferentOuterAndOwner(t *testing.T) {
	dir := tuiRuntimeOlDir(t)
	tuiRuntimeOlWrapper(t, dir, "ol4-bridge", tuiRuntimeOlParkScript())
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	rt := &tuiRuntime{bridge: b}
	owner, outer := "owner-ol4", "outer-ol4"
	defer tuiRuntimeStopBridge(t, b)
	parked := tuiRuntimeOlParkedFirstTurn(t, dir, rt, b, owner, outer)

	body := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"write_file","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"另一个会话的问题"}]}`)
	// Same owner, different outer: must not be blocked.
	tuiRuntimeOlWrapper(t, dir, "ol4b-bridge", tuiRuntimeOlAnswerScript(""))
	code, msg, out := tuiRuntimeOlServe(t, rt, b, owner, "outer-ol4-other", body)
	if code != 200 || !strings.Contains(out, "OL_ANSWER") {
		t.Fatalf("a different outer on the same owner was blocked: %d %s %s", code, msg, out)
	}
	// Same outer id, different owner: must not be blocked.
	code, msg, out = tuiRuntimeOlServe(t, rt, b, "owner-ol4-other", outer, body)
	if code != 200 || !strings.Contains(out, "OL_ANSWER") {
		t.Fatalf("the same outer under a different owner was blocked: %d %s %s", code, msg, out)
	}
	// And the canonical outer is still owned by the parked run alone; the
	// helper runs carried outerSession "" and never claimed it.
	if _, same := tuiRuntimeOlLiveOuterRuns(b, owner, outer); same != 1 {
		t.Fatalf("%d live runs claim outer %q (want 1; the parallel runs are not canonical)", same, outer)
	}
	tuiRuntimeOlRelease(parked, dir, outer)
	b.abortAll()
}

// A helper request — an independent logical thread that reuses the main
// conversation's outer id (here: a different model, so it takes the
// auxiliary path) — must not be blocked by the live canonical run, and
// must not take over the outer either.
func TestClaudeTUIRuntimeOuterLiveHelperNotBlocked(t *testing.T) {
	dir := tuiRuntimeOlDir(t)
	tuiRuntimeOlWrapper(t, dir, "ol5-bridge", tuiRuntimeOlParkScript())
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	rt := &tuiRuntime{bridge: b}
	owner, outer := "owner-ol5", "outer-ol5"
	// Park the canonical first turn before any checkpoint exists; the
	// helper branch below then needs no saved state.
	parked := tuiRuntimeOlParkedFirstTurn(t, dir, rt, b, owner, outer)

	// A Desktop helper on the same outer id but a different (cheaper)
	// model takes the auxiliary path with outerSession "".
	tuiRuntimeOlWrapper(t, dir, "ol5b-bridge", tuiRuntimeOlAnswerScript(""))
	helper := []byte(`{"model":"claude-haiku-4-5","max_tokens":64,"messages":[{"role":"user","content":"快速查个小事"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(helper)))
	code, msg := rt.serveMessagesInner(tuiRuntimeAnthropicResponder{rec, req, &Usage{}}, req, "Claude Code", "claude-haiku-4-5", owner, outer, helper, func(c context.Context, r *Request) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
		return b.startInteractiveRequest(c, r, "claude-haiku-4-5", "", owner, outer)
	})
	if code != 200 {
		t.Fatalf("the helper request was blocked by the live canonical run: %d %s", code, msg)
	}
	// The canonical outer still belongs to the parked run; the helper kept
	// outerSession "" and never registered against it.
	key := tuiRuntimeInteractiveSessionStateKey(owner, outer)
	b.mu.Lock()
	live := b.liveOuter[key]
	b.mu.Unlock()
	if live == nil {
		t.Fatal("the parked run lost canonical ownership of its outer")
	}
	tuiRuntimeOlRelease(parked, dir, outer)
	b.abortAll()
}

// When a start fails before its child is published — the wrapper is
// missing — the outer keeps no live marker, and the next request may
// start again.
func TestClaudeTUIRuntimeOuterLiveStartFailureLeavesNoOwner(t *testing.T) {
	dir := tuiRuntimeOlDir(t)
	// Point the bridge env at a nonexistent wrapper: startInteractiveRequest
	// fails closed before any run exists.
	tuiRuntimeSetWrapper(t, filepath.Join(dir, "missing-bridge"))
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	rt := &tuiRuntime{bridge: b}
	owner, outer := "owner-ol6", "outer-ol6"
	body := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"messages":[{"role":"user","content":"把结果写进 notes.txt"}]}`)
	code, _, _ := tuiRuntimeOlServe(t, rt, b, owner, outer, body)
	if code != http.StatusBadGateway {
		t.Fatalf("a missing wrapper must fail closed 502, got %d", code)
	}
	key := tuiRuntimeInteractiveSessionStateKey(owner, outer)
	b.mu.Lock()
	live := b.liveOuter[key]
	b.mu.Unlock()
	if live != nil {
		t.Fatal("a failed start left a live-outer owner behind")
	}
	// Now with a working wrapper the same outer starts normally.
	tuiRuntimeOlWrapper(t, dir, "ol6-bridge", tuiRuntimeOlAnswerScript(""))
	code, msg, out := tuiRuntimeOlServe(t, rt, b, owner, outer, body)
	b.abortAll()
	if code != 200 || !strings.Contains(out, "OL_ANSWER") {
		t.Fatalf("the outer stayed locked after a failed start: %d %s %s", code, msg, out)
	}
}

// A live canonical run that dies without its client — abort() on the
// SIGINT-ignoring wrapper path — frees the outer only once the child is
// reaped, never while abort is still killing it.
func TestClaudeTUIRuntimeOuterLiveAbortReleasesOnlyAfterReap(t *testing.T) {
	dir := tuiRuntimeOlDir(t)
	// The wrapper parks on a tool call and ignores SIGINT, so abort must go
	// through the full interactiveAbortGrace + Kill path.
	script := "#!/bin/sh\ntrap '' INT\n" +
		"cat >/dev/null\n" +
		"printf '%s\\n' '{\"type\":\"assistant\",\"uuid\":\"t1\",\"session_id\":\"inner-ol\",\"message\":{\"id\":\"m1\",\"model\":\"claude-opus-5-5\",\"content\":[{\"type\":\"tool_use\",\"id\":\"call_1\",\"name\":\"mcp__magpie__write_file\",\"input\":{\"path\":\"a\"}}]}}'\n" +
		"sleep 30\n"
	tuiRuntimeOlWrapper(t, dir, "ol7-bridge", script)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	rt := &tuiRuntime{bridge: b}
	owner, outer := "owner-ol7", "outer-ol7"
	release := filepath.Join(dir, "release-outer-ol7")
	t.Setenv("DAV1878_PARK_RELEASE", release)
	first := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"write_file","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"把结果写进 notes.txt"}]}`)
	code, msg, out := tuiRuntimeOlServe(t, rt, b, owner, outer, first)
	if code != 200 || !strings.Contains(out, `"type":"tool_use"`) {
		t.Fatalf("first turn did not park on a tool call: %d %s %s", code, msg, out)
	}
	var parked *tuiRuntimeSubscriptionRun
	b.mu.Lock()
	for _, r := range b.runs {
		parked = r
	}
	b.mu.Unlock()
	if parked == nil {
		t.Fatal("no parked run")
	}
	// While abort is in flight (SIGINT ignored → grace → Kill → reap), the
	// outer must still be owned: the child is not provably gone yet.
	aborted := make(chan struct{})
	go func() { parked.abort(); close(aborted) }()
	time.Sleep(50 * time.Millisecond)
	key := tuiRuntimeInteractiveSessionStateKey(owner, outer)
	b.mu.Lock()
	stillOwned := b.liveOuter[key]
	b.mu.Unlock()
	if stillOwned == nil {
		t.Fatal("the outer was freed while abort was still killing the child")
	}
	<-aborted
	tuiRuntimeOlWaitReaped(t, parked)
	deadline := time.Now().Add(5 * time.Second)
	for {
		b.mu.Lock()
		live := b.liveOuter[key]
		_, inRuns := b.runs[parked.token]
		b.mu.Unlock()
		if live == nil && !inRuns {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the aborted run still owns the outer after its child was reaped")
		}
		time.Sleep(5 * time.Millisecond)
	}
	// And a new turn may start on the outer again.
	tuiRuntimeOlWrapper(t, dir, "ol7b-bridge", tuiRuntimeOlAnswerScript(""))
	code, msg, out = tuiRuntimeOlServe(t, rt, b, owner, outer, first)
	b.abortAll()
	if code != 200 || !strings.Contains(out, "OL_ANSWER") {
		t.Fatalf("the outer stayed locked after abort+reap: %d %s %s", code, msg, out)
	}
}

// A dirty-checkpoint opt-in recovery start is refused while an ordinary
// run is live on the outer, and an ordinary start is refused while a
// recovery run is live — live-outer exclusion covers every canonical
// start kind, symmetrically with the per-request claim.
func TestClaudeTUIRuntimeOuterLiveCoversRecoveryAndOrdinaryStarts(t *testing.T) {
	dir := tuiRuntimeOlDir(t)
	tuiRuntimeOlWrapper(t, dir, "ol8-bridge", tuiRuntimeOlParkScript())
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	rt := &tuiRuntime{bridge: b}
	owner, outer := "owner-ol8", "outer-ol8"
	parked := tuiRuntimeOlParkedFirstTurn(t, dir, rt, b, owner, outer)

	// Mark a checkpoint dirty; the opt-in request that follows would
	// otherwise start a read-only recovery turn. The checkpoint is saved
	// after the park so the parked run could start as a first turn, and
	// its first user message is the parked run's own thread words so the
	// request is canonical (a different first thread would be auxiliary
	// and fail closed for its own reasons, not via live-outer).
	tuiRuntimeOlSave(t, b, owner, outer, "inner-ol8", true)
	optIn := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"messages":[` +
		`{"role":"user","content":"把结果写进 notes.txt"},` +
		`{"role":"assistant","content":[{"type":"tool_use","id":"abandoned","name":"write_file","input":{"path":"notes.txt"}}]},` +
		`{"role":"user","content":[` +
		`{"type":"tool_result","tool_use_id":"abandoned","content":"written ok"},` +
		`{"type":"text","text":"/magpie-recover"}]}]}`)
	code, msg, _ := tuiRuntimeOlServe(t, rt, b, owner, outer, optIn)
	if code != http.StatusConflict {
		t.Fatalf("a recovery start on an outer with a live run was not refused: %d %s", code, msg)
	}
	if _, same := tuiRuntimeOlLiveOuterRuns(b, owner, outer); same != 1 {
		t.Fatalf("%d live runs on the outer after the refused recovery (want 1)", same)
	}
	tuiRuntimeOlRelease(parked, dir, outer)
	b.abortAll()

	// Symmetric direction: while a recovery turn is live, an ordinary
	// request on the same outer is refused.
	dir2 := tuiRuntimeOlDir(t)
	release2 := filepath.Join(dir2, "slow-release")
	t.Setenv("DAV1878_SLOW", release2)
	slow := "#!/bin/sh\ncat >/dev/null\n" +
		"for i in $(seq 1 400); do [ -f \"$DAV1878_SLOW\" ] && break; sleep 0.05; done\n" +
		"printf '%s\\n' '{\"type\":\"assistant\",\"uuid\":\"a1\",\"session_id\":\"inner-ol\",\"message\":{\"id\":\"m1\",\"model\":\"claude-opus-5-5\",\"content\":[{\"type\":\"text\",\"text\":\"OL_ANSWER\"}]}}'\n" +
		"printf '%s\\n' '{\"type\":\"result\",\"session_id\":\"inner-ol\",\"stop_reason\":\"end_turn\"}'\n"
	tuiRuntimeOlWrapper(t, dir2, "ol8b-bridge", slow)
	b2 := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b2)
	rt2 := &tuiRuntime{bridge: b2}
	owner2, outer2 := "owner-ol8b", "outer-ol8b"
	tuiRuntimeOlSave(t, b2, owner2, outer2, "inner-ol8b", true)
	var starts atomic.Int64
	start2 := func(c context.Context, r *Request) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
		run, ev, err := b2.startInteractiveRequest(c, r, tuiRuntimeOlModel, "", owner2, outer2)
		if err == nil && run != nil {
			starts.Add(1)
		}
		return run, ev, err
	}
	firstDone := make(chan struct{})
	go func() {
		defer close(firstDone)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(optIn)))
		rt2.serveMessagesInner(tuiRuntimeAnthropicResponder{rec, req, &Usage{}}, req, "Claude Code", tuiRuntimeOlModel, owner2, outer2, optIn, start2)
	}()
	deadline := time.Now().Add(5 * time.Second)
	for starts.Load() == 0 {
		if time.Now().After(deadline) {
			t.Fatal("the recovery turn never started")
		}
		time.Sleep(5 * time.Millisecond)
	}
	ordinary := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"messages":[` +
		`{"role":"user","content":"先读一下配置"},{"role":"assistant","content":"配置已读"},{"role":"user","content":"新一轮"}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(ordinary)))
	code, msg = rt2.serveMessagesInner(tuiRuntimeAnthropicResponder{rec, req, &Usage{}}, req, "Claude Code", tuiRuntimeOlModel, owner2, outer2, ordinary, start2)
	if code != http.StatusConflict {
		t.Fatalf("an ordinary start on an outer with a live recovery run was not refused: %d %s", code, msg)
	}
	if n := starts.Load(); n != 1 {
		t.Fatalf("%d PTY runs started on one outer (want the recovery only)", n)
	}
	_ = os.WriteFile(release2, []byte("ok"), 0o600)
	<-firstDone
	b2.abortAll()
}

// Counterexample guard: the refusal is about a live canonical run, not
// about the request shape. The exact same second request that is refused
// while the run is parked is accepted once the run is gone — so the test
// above proves liveness, not a shape filter.
func TestClaudeTUIRuntimeOuterLiveSameRequestAcceptedAfterRunEnds(t *testing.T) {
	dir := tuiRuntimeOlDir(t)
	tuiRuntimeOlWrapper(t, dir, "ol9-bridge", tuiRuntimeOlParkScript())
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	rt := &tuiRuntime{bridge: b}
	owner, outer := "owner-ol9", "outer-ol9"
	parked := tuiRuntimeOlParkedFirstTurn(t, dir, rt, b, owner, outer)
	second := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"write_file","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"把结果写进 notes.txt"}]}`)
	code, _, _ := tuiRuntimeOlServe(t, rt, b, owner, outer, second)
	if code != http.StatusConflict {
		t.Fatalf("expected the parked duplicate to be refused, got %d", code)
	}
	tuiRuntimeOlRelease(parked, dir, outer)
	tuiRuntimeOlWaitReaped(t, parked)
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, same := tuiRuntimeOlLiveOuterRuns(b, owner, outer); same == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the ended run still owns the outer")
		}
		time.Sleep(5 * time.Millisecond)
	}
	tuiRuntimeOlWrapper(t, dir, "ol9b-bridge", tuiRuntimeOlAnswerScript(""))
	code, msg, out := tuiRuntimeOlServe(t, rt, b, owner, outer, second)
	b.abortAll()
	if code != 200 {
		t.Fatalf("the identical request stayed refused after the run ended: %d %s", code, msg)
	}
	if !strings.Contains(out, "OL_ANSWER") {
		t.Fatalf("the identical request did not start a fresh turn: %s", out)
	}
}
