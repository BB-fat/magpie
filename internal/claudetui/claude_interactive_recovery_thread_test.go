package claudetui

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A running Runtime remembers the original firstWords in outerThreads.
// Compaction changes them; a cold fixture misses this production state.
func TestRecoveryWithLiveThreadHeuristicCommitsClean(t *testing.T) {
	dir := arDir(t)
	arWrapper(t, dir, "answer", arAnswerScript())
	t.Setenv(autoOrphanRecoveryEnv, "1")
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	rt := &runtimeTUI{bridge: b}
	owner, outer := "owner", "compacted-live-thread"
	arSave(t, b, owner, outer, "inner-old", true)
	before, _, _ := b.loadInteractiveSession(owner, outer)
	b.seedOuterThread(owner, outer, before.ThreadKey)
	body := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"echo","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"[compacted summary] earlier work"},{"role":"assistant","content":"ack"},{"role":"user","content":[{"type":"tool_result","tool_use_id":"orphan-1","content":"DONE"}]}]}`)
	code, msg, out := arServe(t, rt, b, owner, outer, body)
	if code != 200 || !strings.Contains(out, "OL_ANSWER") {
		t.Fatalf("recovery: %d %s %s", code, msg, out)
	}
	after, ok, err := b.loadInteractiveSession(owner, outer)
	if err != nil || !ok || after.Dirty || after.SessionID == before.SessionID {
		t.Fatalf("HTTP 200 without canonical clean checkpoint: dirty=%v same_native=%v err=%v", after.Dirty, after.SessionID == before.SessionID, err)
	}
	// The next normal request must keep both tools and deliver a fresh MCP call.
	deadline := time.Now().Add(5 * time.Second)
	for {
		_, same := olLiveOuterRuns(b, outer)
		if same == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("recovery owner not released")
		}
		time.Sleep(5 * time.Millisecond)
	}
	release := filepath.Join(dir, "release")
	t.Setenv("RECOVERY_TEST_RELEASE", release)
	arWrapper(t, dir, "next", "#!/bin/sh\ncat >/dev/null\nfor i in $(seq 1 500); do [ -f \"$RECOVERY_TEST_RELEASE\" ] && break; sleep 0.01; done\n"+strings.TrimPrefix(arAnswerScript(), "#!/bin/sh\ncat >/dev/null\n"))
	next := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"echo","input_schema":{"type":"object"}},{"name":"status","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"[compacted summary] earlier work"},{"role":"assistant","content":"OL_ANSWER"},{"role":"user","content":"echo fresh probe"}]}`)
	type response struct {
		code     int
		msg, out string
	}
	done := make(chan response, 1)
	go func() { c, m, o := arServe(t, rt, b, owner, outer, next); done <- response{c, m, o} }()
	var run *subscriptionRun
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		b.mu.Lock()
		for _, r := range b.runs {
			if r.outerSession == outer {
				run = r
				break
			}
		}
		b.mu.Unlock()
		if run != nil {
			run.mu.Lock()
			ready := run.onCall != nil && len(run.tools) == 2
			run.mu.Unlock()
			if ready {
				break
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	if run == nil {
		t.Fatal("next normal request lost canonical outer")
	}
	run.mu.Lock()
	readonly := run.readOnlyRecovery
	ntools := len(run.tools)
	run.mu.Unlock()
	if readonly || ntools != 2 {
		t.Fatalf("next tool surface: readonly=%v tools=%d", readonly, ntools)
	}
	raw, err := os.ReadFile(filepath.Join(run.tmp, "tools.json"))
	if err != nil || !strings.Contains(string(raw), "echo") || !strings.Contains(string(raw), "status") {
		t.Fatalf("MCP tool registration missing: %v", err)
	}
	callbackDone := make(chan string, 1)
	go func() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/_magpie/claude-mcp/"+run.token, strings.NewReader(`{"tool_call_id":"fresh-probe","name":"echo","arguments":{"text":"fresh"}}`))
		r.SetPathValue("token", run.token)
		b.mcpCall(w, r)
		callbackDone <- w.Body.String()
	}()
	var first response
	select {
	case first = <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("fresh tool call not delivered to client")
	}
	if first.code != 200 || !strings.Contains(first.out, "tool_use") || !strings.Contains(first.out, "fresh-probe") {
		t.Fatalf("new tool not delivered: %+v", first)
	}
	follow := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"echo","input_schema":{"type":"object"}},{"name":"status","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"[compacted summary] earlier work"},{"role":"assistant","content":"OL_ANSWER"},{"role":"user","content":"echo fresh probe"},{"role":"assistant","content":[{"type":"tool_use","id":"fresh-probe","name":"echo","input":{"text":"fresh"}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"fresh-probe","content":"fresh"}]}]}`)
	go func() { c, m, o := arServe(t, rt, b, owner, outer, follow); done <- response{c, m, o} }()
	select {
	case got := <-callbackDone:
		if !strings.Contains(got, "fresh") {
			t.Fatalf("result not returned to native MCP callback: %s", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("fresh MCP result stuck")
	}
	os.WriteFile(release, []byte("ok"), 0600)
	select {
	case final := <-done:
		if final.code != 200 || !strings.Contains(final.out, "OL_ANSWER") {
			t.Fatalf("post-tool response: %+v", final)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("post-tool final reply stuck")
	}
	b.abortAll()
}
