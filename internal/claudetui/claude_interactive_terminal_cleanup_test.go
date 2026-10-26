package claudetui

// DAV-1888 regression tests for the terminal-cleanup window mis-409.
//
// The Linux DAV-1887 e2e proved a real regression: a canonical PTY run
// whose complete reply was delivered and whose clean checkpoint was saved
// still owns liveOuter until its child is reaped — the wrapper's tmux
// EXIT_HOLD_MS (~2s) teardown window. A normal next-turn request that
// arrives inside that window was refused errInteractiveOuterLive as if a
// second PTY were still running. These tests pin the fixed semantics:
//
//   - a proven terminal-delivered run (complete non-tool reply delivered,
//     clean canonical checkpoint saved, no read-only tool signal) lets the
//     next canonical request wait, outside the lock, for the owner to be
//     reaped and removed, then start normally;
//   - a still-executing, parked, failed, read-only-signaled, cancelled or
//     unknown-state live run keeps refusing a second canonical PTY;
//   - waiting never overlaps the old and new child, never forces an early
//     liveOuter release, and fails closed on timeout/context cancel.

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// tcWrapper installs a fake claude-bridge whose behaviour is selected per
// launch from the MAGPIE_TC_SCRIPT env var, so each HTTP request can pick
// its own script.
func tcWrapper(t *testing.T, dir string) {
	t.Helper()
	p := filepath.Join(dir, "tc-bridge")
	script := "#!/bin/sh\nexec /bin/sh \"$MAGPIE_TC_SCRIPT\"\n"
	if err := os.WriteFile(p, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	setWrapper(t, p)
}

func tcScript(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	return p
}

// tcAnswerScript emits one complete assistant text row and a terminal
// result, then waits for GATEFILE (or times out) before exiting —
// modelling the wrapper's tmux EXIT_HOLD teardown.
func tcAnswerScript() string {
	return "#!/bin/sh\n" +
		"cat >/dev/null\n" +
		"printf '%s\\n' '{\"type\":\"assistant\",\"uuid\":\"a1\",\"session_id\":\"inner-tc\",\"message\":{\"id\":\"m1\",\"model\":\"claude-opus-5-5\",\"content\":[{\"type\":\"text\",\"text\":\"TC_ANSWER\"}]}}'\n" +
		"printf '%s\\n' '{\"type\":\"result\",\"session_id\":\"inner-tc\",\"stop_reason\":\"end_turn\"}'\n" +
		"for i in $(seq 1 300); do [ -n \"$MAGPIE_TC_GATE\" ] && [ -f \"$MAGPIE_TC_GATE\" ] && break; sleep 0.1; done\n"
}

// tcMarkScript emits a complete reply then sleeps holding the reaper,
// exactly like the real wrapper's EXIT_HOLD window; the marker file
// proves the wrapper is still inside its hold.
func tcMarkScript(markEnv string) string {
	return "#!/bin/sh\n" +
		"cat >/dev/null\n" +
		"printf '%s\\n' '{\"type\":\"assistant\",\"uuid\":\"a1\",\"session_id\":\"inner-tc\",\"message\":{\"id\":\"m1\",\"model\":\"claude-opus-5-5\",\"content\":[{\"type\":\"text\",\"text\":\"TC_ANSWER\"}]}}'\n" +
		"printf '%s\\n' '{\"type\":\"result\",\"session_id\":\"inner-tc\",\"stop_reason\":\"end_turn\"}'\n" +
		"printf held > \"${" + markEnv + "}\"\n" +
		"for i in $(seq 1 300); do [ -n \"$MAGPIE_TC_GATE\" ] && [ -f \"$MAGPIE_TC_GATE\" ] && break; sleep 0.1; done\n"
}

// tcParkScript emits one client tool_use row and stays alive parked: the
// run is genuinely waiting on tool results and must keep the outer.
func tcParkScript() string {
	return "#!/bin/sh\n" +
		"cat >/dev/null\n" +
		"printf '%s\\n' '{\"type\":\"assistant\",\"uuid\":\"t1\",\"session_id\":\"inner-tc\",\"message\":{\"id\":\"m1\",\"model\":\"claude-opus-5-5\",\"content\":[{\"type\":\"tool_use\",\"id\":\"call_1\",\"name\":\"mcp__magpie__write_file\",\"input\":{\"path\":\"a\"}}]}}'\n" +
		"printf '%s\\n' '{\"type\":\"result\",\"session_id\":\"inner-tc\",\"stop_reason\":\"tool_use\"}'\n" +
		"sleep 30\n"
}

// tcFailScript ends the turn with an error result row.
func tcFailScript() string {
	return "#!/bin/sh\n" +
		"cat >/dev/null\n" +
		"printf '%s\\n' '{\"type\":\"result\",\"session_id\":\"inner-tc\",\"is_error\":true,\"result\":\"wrapper exploded\",\"stop_reason\":\"error\"}'\n" +
		"sleep 30\n"
}

// tcROToolScript is a read-only-recovery turn that emits a forbidden
// tool_use row before its terminal result: the run must end failed and
// the checkpoint must stay dirty.
func tcROToolScript() string {
	return "#!/bin/sh\n" +
		"cat >/dev/null\n" +
		"printf '%s\\n' '{\"type\":\"assistant\",\"uuid\":\"r1\",\"session_id\":\"inner-tc\",\"message\":{\"id\":\"m\",\"model\":\"claude-opus-5-5\",\"content\":[{\"type\":\"tool_use\",\"id\":\"forged\",\"name\":\"mcp__magpie__write_file\",\"input\":{}}]}}'\n" +
		"printf '%s\\n' '{\"type\":\"result\",\"session_id\":\"inner-tc\",\"stop_reason\":\"end_turn\"}'\n" +
		"sleep 30\n"
}

func tcDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	return dir
}

func tcServe(t *testing.T, rt *runtimeTUI, b *subscriptionBridge, owner, outer string, body []byte) (int, string, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body)))
	code, msg := rt.serveMessagesInner(anthropicResponder{rec, req, &Usage{}}, req, "Claude Code", "claude-opus-5-5", owner, outer, body, func(c context.Context, r *Request) (*subscriptionRun, <-chan Event, error) {
		return b.startInteractiveRequest(c, r, "claude-opus-5-5", "", owner, outer)
	})
	return code, msg, rec.Body.String()
}

func tcLaunchCount(t *testing.T, dir string) int {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "launches"))
	if err != nil {
		return 0
	}
	return strings.Count(string(b), "\n")
}

// tcScriptEnv points the dispatcher at one script per turn; gate names a
// file that releases the wrapper's hold.
func tcScriptEnv(t *testing.T, dir, script string, gateEnv string) {
	t.Helper()
	t.Setenv("MAGPIE_TC_SCRIPT", script)
	if gateEnv != "" {
		t.Setenv("MAGPIE_TC_GATE", filepath.Join(dir, gateEnv))
	} else {
		os.Unsetenv("MAGPIE_TC_GATE")
	}
}

// DAV-1888 core regression: a complete terminal reply whose clean
// checkpoint was saved, but whose wrapper is still inside its tmux
// EXIT_HOLD teardown, must not refuse the next canonical turn. The next
// request waits for the owner to be reaped, then starts and resumes.
func TestTerminalCleanupWaitedRunLetsNextTurnResume(t *testing.T) {
	dir := tcDir(t)
	tcWrapper(t, dir)
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	rt := &runtimeTUI{bridge: b}
	owner, outer := "owner-tc", "outer-tc"

	// Turn 1: first canonical turn, complete reply, wrapper still holding
	// the child in teardown (a real PTY wrapper's EXIT_HOLD window).
	mark := filepath.Join(dir, "turn1-held")
	t.Setenv("MAGPIE_TC_MARK1", mark)
	first := tcScript(t, dir, "turn1", "#!/bin/sh\n"+
		"cat >/dev/null\n"+
		"printf '%s\\n' '{\"type\":\"assistant\",\"uuid\":\"a1\",\"session_id\":\"inner-tc1\",\"message\":{\"id\":\"m1\",\"model\":\"claude-opus-5-5\",\"content\":[{\"type\":\"text\",\"text\":\"TC_ONE\"}]}}'\n"+
		"printf '%s\\n' '{\"type\":\"result\",\"session_id\":\"inner-tc1\",\"stop_reason\":\"end_turn\"}'\n"+
		"printf held > \"$MAGPIE_TC_MARK1\"\n"+
		"for i in $(seq 1 300); do [ -f \"$MAGPIE_TC_GATE\" ] && break; sleep 0.1; done\n")
	t.Setenv("MAGPIE_TC_GATE", filepath.Join(dir, "gate1"))
	tcScriptEnv(t, dir, first, "gate1")

	body1 := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"messages":[{"role":"user","content":"turn one"}]}`)
	code, msg, out := tcServe(t, rt, b, owner, outer, body1)
	if code != 200 || !strings.Contains(out, "TC_ONE") {
		t.Fatalf("turn 1 failed: %d %s %s", code, msg, out)
	}
	// The reply was delivered and the clean checkpoint saved, but the
	// wrapper is provably still inside its teardown hold.
	if _, err := os.Stat(mark); err != nil {
		t.Fatalf("wrapper did not reach its hold window: %v", err)
	}
	if _, same := olLiveOuterRuns(b, outer); same != 1 {
		t.Fatalf("expected turn1 run still live in teardown (same=%d)", same)
	}
	saved, ok := b.interactiveSession(owner, outer)
	if !ok || saved.Dirty || saved.SessionID != "inner-tc1" {
		t.Fatalf("clean checkpoint was not saved: %+v ok=%v", saved, ok)
	}

	// Turn 2 arrives immediately — inside the hold window — with the same
	// conversation plus the new user message. It must wait for the owner
	// slot to free, then start a fresh run; never 409, never overlap.
	second := tcScript(t, dir, "turn2", "#!/bin/sh\n"+
		"cat >/dev/null\n"+
		"printf launch2 >> \""+filepath.Join(dir, "launches")+"\"; printf '\\n' >> \""+filepath.Join(dir, "launches")+"\"\n"+
		"printf '%s\\n' '{\"type\":\"assistant\",\"uuid\":\"a2\",\"session_id\":\"inner-tc2\",\"message\":{\"id\":\"m2\",\"model\":\"claude-opus-5-5\",\"content\":[{\"type\":\"text\",\"text\":\"TC_TWO\"}]}}'\n"+
		"printf '%s\\n' '{\"type\":\"result\",\"session_id\":\"inner-tc2\",\"stop_reason\":\"end_turn\"}'\n"+
		"for i in $(seq 1 100); do [ -n \"$MAGPIE_TC_GATE2\" ] && [ -f \"$MAGPIE_TC_GATE2\" ] && break; sleep 0.1; done\n")
	t.Setenv("MAGPIE_TC_GATE2", filepath.Join(dir, "gate2"))
	tcScriptEnv(t, dir, second, "")

	type result struct {
		code int
		msg  string
		out  string
	}
	done := make(chan result, 1)
	go func() {
		c, m, o := tcServe(t, rt, b, owner, outer, []byte(`{"model":"claude-opus-5-5","max_tokens":64,"messages":[{"role":"user","content":"turn one"},{"role":"assistant","content":"TC_ONE"},{"role":"user","content":"turn two"}]}`))
		done <- result{c, m, o}
	}()
	// While turn 2 is queued waiting, release turn 1's teardown.
	time.Sleep(200 * time.Millisecond)
	if err := os.WriteFile(filepath.Join(dir, "gate1"), []byte("go"), 0o600); err != nil {
		t.Fatal(err)
	}
	var r2 result
	select {
	case r2 = <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("turn 2 did not complete after the owner slot was released")
	}
	if r2.code != 200 || !strings.Contains(r2.out, "TC_TWO") {
		t.Fatalf("turn 2 was refused instead of waiting for cleanup: %d %s %s", r2.code, r2.msg, r2.out)
	}
	// The new run only launched after the old owner slot was freed, and it
	// launched exactly one new wrapper (no second PTY overlap).
	if n := tcLaunchCount(t, dir); n != 1 {
		t.Fatalf("expected exactly one new wrapper launch, saw %d", n)
	}
	saved, _ = b.interactiveSession(owner, outer)
	if saved.SessionID != "inner-tc2" || saved.Dirty {
		t.Fatalf("turn2 did not replace the checkpoint: %+v", saved)
	}
	_ = os.WriteFile(filepath.Join(dir, "gate2"), []byte("go"), 0o600)
	b.abortAll()
}

// The same next-turn request must not spawn a second PTY while the owner
// is still held: point the next launch at a script that can never emit a
// result. If the fix ever let it start without waiting, the 15s window
// would lapse with no TC_TWO — here it must wait, then fail closed once
// the owner stays held past the bounded cleanup wait.
func TestTerminalCleanupWaitTimesOutUnreachedOwner(t *testing.T) {
	dir := tcDir(t)
	tcWrapper(t, dir)
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	rt := &runtimeTUI{bridge: b}
	owner, outer := "owner-tc", "outer-tc"

	// Turn 1 holds its owner slot ~30s (longer than the bounded cleanup
	// wait), reply already delivered, clean checkpoint saved.
	mark := filepath.Join(dir, "turn1-held")
	t.Setenv("MAGPIE_TC_MARK1", mark)
	t.Setenv("MAGPIE_TC_GATE", filepath.Join(dir, "gate-never"))
	first := tcScript(t, dir, "turn1", tcMarkScript("MAGPIE_TC_MARK1"))
	tcScriptEnv(t, dir, first, "")
	// keep GATE pointing at a file that never appears so turn1 holds
	t.Setenv("MAGPIE_TC_GATE", filepath.Join(dir, "gate-never"))
	body1 := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"messages":[{"role":"user","content":"turn one"}]}`)
	code, _, out := tcServe(t, rt, b, owner, outer, body1)
	if code != 200 || !strings.Contains(out, "TC_ANSWER") {
		t.Fatalf("turn 1 failed: %d %s", code, out)
	}
	if _, err := os.Stat(mark); err != nil {
		t.Fatal(err)
	}

	// Turn 2: the owner stays held longer than the wait bound. The request
	// must fail closed, not 200, and must not leave a second live run.
	second := tcScript(t, dir, "turn2", "#!/bin/sh\ncat >/dev/null\nprintf launch2 >> \""+filepath.Join(dir, "launches")+"\"\nsleep 5\n")
	tcScriptEnv(t, dir, second, "")
	code2, _, _ := tcServe(t, rt, b, owner, outer, []byte(`{"model":"claude-opus-5-5","max_tokens":64,"messages":[{"role":"user","content":"turn one"},{"role":"assistant","content":"TC_ANSWER"},{"role":"user","content":"turn two"}]}`))
	if code2 != http.StatusConflict {
		t.Fatalf("terminal-cleanup busy timeout must fail closed with 409, got %d", code2)
	}
	if _, same := olLiveOuterRuns(b, outer); same != 1 {
		t.Fatalf("expected only the still-held first run (same=%d)", same)
	}
	if n := tcLaunchCount(t, dir); n != 0 {
		t.Fatalf("a second wrapper was launched during the held window: %d", n)
	}
	b.abortAll()
}

// The terminal-cleanup wait is bound by the request context: cancel the
// caller while it waits and the start fails without forcing the owner
// slot, without a second PTY and without clearing dirty.
func TestTerminalCleanupWaitCancelledKeepsOwner(t *testing.T) {
	dir := tcDir(t)
	tcWrapper(t, dir)
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	rt := &runtimeTUI{bridge: b}
	owner, outer := "owner-tc", "outer-tc"

	mark := filepath.Join(dir, "turn1-held")
	t.Setenv("MAGPIE_TC_MARK1", mark)
	first := tcScript(t, dir, "turn1", tcMarkScript("MAGPIE_TC_MARK1"))
	tcScriptEnv(t, dir, first, "")
	body1 := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"messages":[{"role":"user","content":"turn one"}]}`)
	code, _, out := tcServe(t, rt, b, owner, outer, body1)
	if code != 200 || !strings.Contains(out, "TC_ANSWER") {
		t.Fatalf("turn 1 failed: %d %s", code, out)
	}
	if _, err := os.Stat(mark); err != nil {
		t.Fatalf("wrapper did not reach its hold window: %v", err)
	}

	// Turn 2 on a cancellable request context.
	ctx, cancel := context.WithCancel(context.Background())
	second := tcScript(t, dir, "turn2", tcAnswerScript())
	tcScriptEnv(t, dir, second, "")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude-opus-5-5","max_tokens":64,"messages":[{"role":"user","content":"turn one"},{"role":"assistant","content":"TC_ANSWER"},{"role":"user","content":"turn two"}]}`)).WithContext(ctx)
	done := make(chan int, 1)
	go func() {
		body, _ := io.ReadAll(req.Body)
		req2 := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body))).WithContext(ctx)
		code, _ := rt.serveMessagesInner(anthropicResponder{rec, req2, &Usage{}}, req2, "Claude Code", "claude-opus-5-5", owner, outer, body, func(c context.Context, r *Request) (*subscriptionRun, <-chan Event, error) {
			return b.startInteractiveRequest(c, r, "claude-opus-5-5", "", owner, outer)
		})
		done <- code
	}()
	time.Sleep(200 * time.Millisecond)
	cancel()
	var code2 int
	select {
	case code2 = <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("cancelled wait did not return")
	}
	if code2 == 200 {
		t.Fatalf("a cancelled wait still started a second PTY: %s", rec.Body.String())
	}
	if _, same := olLiveOuterRuns(b, outer); same != 1 {
		t.Fatalf("the old run lost its owner slot to a cancelled wait (same=%d)", same)
	}
	b.mu.Lock()
	liveOuter := b.liveOuter[interactiveSessionStateKey(owner, outer)]
	b.mu.Unlock()
	if liveOuter == nil {
		t.Fatal("cancelled wait force-released liveOuter")
	}
	b.abortAll()
}

// A parked run — live, waiting on unanswered client tool calls — is not a
// terminal-delivered run: a same-outer non-matching request stays 409.
func TestTerminalCleanupParkedRunStillRefuses(t *testing.T) {
	dir := tcDir(t)
	tcWrapper(t, dir)
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	rt := &runtimeTUI{bridge: b}
	owner, outer := "owner-tc", "outer-tc"
	tcScriptEnv(t, dir, tcScript(t, dir, "park", tcParkScript()), "")
	body1 := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"write_file","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"把结果写进 notes.txt"}]}`)
	code, _, out := tcServe(t, rt, b, owner, outer, body1)
	if code != 200 || !strings.Contains(out, `"type":"tool_use"`) {
		t.Fatalf("first turn did not park: %d %s", code, out)
	}
	// The park left no checkpoint and the run is still live: refuse.
	tcScriptEnv(t, dir, tcScript(t, dir, "next", tcAnswerScript()), "")
	code, msg, _ := tcServe(t, rt, b, owner, outer, body1)
	if code != http.StatusConflict {
		t.Fatalf("a parked run let a second canonical PTY start: %d %s", code, msg)
	}
	b.abortAll()
}

// A failed turn — error result, no clean reply — is not
// terminal-delivered: the checkpoint stays dirty.
func TestTerminalCleanupFailedRunStaysDirty(t *testing.T) {
	dir := tcDir(t)
	tcWrapper(t, dir)
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	rt := &runtimeTUI{bridge: b}
	owner, outer := "owner-tc", "outer-tc"
	olSave(t, b, owner, outer, "inner-tc", true)
	tcScriptEnv(t, dir, tcScript(t, dir, "fail", tcFailScript()), "")
	body := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"messages":[{"role":"user","content":"先读一下配置"},{"role":"assistant","content":"配置已读"},{"role":"user","content":"/magpie-recover"}]}`)
	code, _, _ := tcServe(t, rt, b, owner, outer, body)
	if code == 200 {
		t.Fatalf("a failed wrapper turn returned 200")
	}
	saved, _ := b.interactiveSession(owner, outer)
	if !saved.Dirty {
		t.Fatal("a failed terminal turn cleared the dirty checkpoint")
	}
	b.abortAll()
}

// A read-only recovery run that saw a tool signal is a failed turn: the
// checkpoint must stay dirty and the run is not a waitable terminal run.
func TestTerminalCleanupReadOnlySignaledStaysDirty(t *testing.T) {
	dir := tcDir(t)
	tcWrapper(t, dir)
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	rt := &runtimeTUI{bridge: b}
	owner, outer := "owner-tc", "outer-tc"
	olSave(t, b, owner, outer, "inner-tc", true)
	tcScriptEnv(t, dir, tcScript(t, dir, "ro-tool", tcROToolScript()), "")
	body := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"write_file","description":"w","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"先读一下配置"},{"role":"assistant","content":[{"type":"tool_use","id":"c1","name":"write_file","input":{"path":"p"}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"c1","content":"x"},{"type":"text","text":"/magpie-recover"}]}]}`)
	code, _, out := tcServe(t, rt, b, owner, outer, body)
	saved, _ := b.interactiveSession(owner, outer)
	if !saved.Dirty {
		t.Fatalf("a read-only turn that signaled a tool cleared dirty: %d %s", code, out)
	}
	b.abortAll()
}

// The wait must never produce two live runs on the same outer: while one
// request waits on terminal cleanup, a second concurrent canonical start
// still fails closed rather than queueing a second new PTY.
func TestTerminalCleanupConcurrentStartStaysExclusive(t *testing.T) {
	dir := tcDir(t)
	tcWrapper(t, dir)
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	rt := &runtimeTUI{bridge: b}
	owner, outer := "owner-tc", "outer-tc"

	mark := filepath.Join(dir, "turn1-held")
	t.Setenv("MAGPIE_TC_MARK1", mark)
	first := tcScript(t, dir, "turn1", tcMarkScript("MAGPIE_TC_MARK1"))
	tcScriptEnv(t, dir, first, "")
	body1 := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"messages":[{"role":"user","content":"turn one"}]}`)
	code, _, out := tcServe(t, rt, b, owner, outer, body1)
	if code != 200 || !strings.Contains(out, "TC_ANSWER") {
		t.Fatalf("turn 1 failed: %d %s", code, out)
	}
	if _, err := os.Stat(mark); err != nil {
		t.Fatal(err)
	}
	var got atomic.Int32
	var wg sync.WaitGroup
	body2 := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"messages":[{"role":"user","content":"turn one"},{"role":"assistant","content":"TC_ANSWER"},{"role":"user","content":"turn two"}]}`)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, _, _ := tcServe(t, rt, b, owner, outer, body2)
			got.Add(int32(c))
		}()
	}
	wg.Wait()
	if _, same := olLiveOuterRuns(b, outer); same > 1 {
		t.Fatalf("two live canonical runs on one outer (live=%d)", same)
	}
	if got.Load() == 0 {
		t.Fatal("no request ran")
	}
	b.abortAll()
}

var _ = fmt.Sprintf // keep fmt imported for future scripted cases
