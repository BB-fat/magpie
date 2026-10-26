package claudetui

import (
	"testing"
	"time"
)

func TestInteractiveDisconnectKeepsOutstandingToolHandoff(t *testing.T) {
	b := newSubscriptionBridge()
	run := &subscriptionRun{
		bridge: b, token: "disconnected-tool", model: "claude-opus-5-5",
		interactive: true, pending: map[string]chan mcpToolResult{"already-running": make(chan mcpToolResult, 1)},
		shown: []shownCall{{id: "already-running", name: "execute"}},
		timer: time.AfterFunc(time.Hour, func() {}),
	}
	defer run.timer.Stop()
	b.mu.Lock()
	b.runs[run.token] = run
	b.calls["already-running"] = run
	b.mu.Unlock()
	segment := run.attach()
	run.letGo()
	run.mu.Lock()
	stillWaiting := !run.closed && run.segment == nil && !run.parkedAt.IsZero() && len(run.pending) == 1
	run.mu.Unlock()
	if !stillWaiting {
		t.Fatal("client disconnect discarded an outstanding MCP tool result")
	}
	if _, open := <-segment; open {
		t.Fatal("disconnected response stream stayed attached")
	}
	b.mu.Lock()
	owner := b.calls["already-running"]
	b.mu.Unlock()
	if owner != run {
		t.Fatal("tool result lost its original Claude run")
	}
}

func TestInteractiveDisconnectBeforeMCPCallbackRegistration(t *testing.T) {
	b := newSubscriptionBridge()
	run := &subscriptionRun{bridge: b, token: "callback-not-yet-registered", model: "claude-opus-5-5",
		interactive: true, pending: map[string]chan mcpToolResult{},
		shown: []shownCall{{id: "shown-but-not-registered", name: "execute"}},
		timer: time.AfterFunc(time.Hour, func() {}),
	}
	defer run.timer.Stop()
	b.mu.Lock()
	b.runs[run.token] = run
	b.mu.Unlock()
	segment := run.attach()
	run.letGo()
	run.mu.Lock()
	stillWaiting := !run.closed && run.segment == nil && !run.parkedAt.IsZero()
	run.mu.Unlock()
	if !stillWaiting {
		t.Fatal("disconnected run was killed before the MCP callback registered")
	}
	if _, open := <-segment; open {
		t.Fatal("disconnected stream remains attached")
	}
}

func TestInteractiveRepeatedDisconnectCannotKillPendingTool(t *testing.T) {
	b := newSubscriptionBridge()
	r := &subscriptionRun{bridge: b, token: "double-cancel", interactive: true,
		pending: map[string]chan mcpToolResult{"call": make(chan mcpToolResult, 1)},
		shown:   []shownCall{{id: "call", name: "read_file"}}, timer: time.AfterFunc(time.Hour, func() {})}
	defer r.timer.Stop()
	b.mu.Lock()
	b.runs[r.token] = r
	b.calls["call"] = r
	b.mu.Unlock()
	r.attach()
	r.letGo()
	r.letGo()
	r.mu.Lock()
	closed, detached := r.closed, r.segment == nil
	r.mu.Unlock()
	if closed || !detached {
		t.Fatalf("duplicate network cancellation destroyed original tool callback: closed=%v detached=%v", closed, detached)
	}
}

func TestInteractiveOldStreamCancellationCannotDetachNewToolContinuation(t *testing.T) {
	b := newSubscriptionBridge()
	r := &subscriptionRun{bridge: b, token: "late-cancel", interactive: true,
		pending: map[string]chan mcpToolResult{"call": make(chan mcpToolResult, 1)},
		shown:   []shownCall{{id: "call", name: "echo"}}, timer: time.AfterFunc(time.Hour, func() {})}
	defer r.timer.Stop()
	b.mu.Lock()
	b.runs[r.token] = r
	b.calls["call"] = r
	b.mu.Unlock()
	old := r.attach()
	newer := r.attach()
	r.letGoFor(old)
	r.mu.Lock()
	closed, same := r.closed, r.segment == newer
	r.mu.Unlock()
	if closed || !same {
		t.Fatalf("stale HTTP cancel touched the newer stream: closed=%v same=%v", closed, same)
	}
}
