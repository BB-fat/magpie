//go:build darwin || linux

package gateway

import (
	"strings"
	"testing"
	"time"
)

func TestClaudeTUIRuntimeInteractiveExplicitRecoveryReapsAbandonedParkedOwner(t *testing.T) {
	dir := tuiRuntimeArDir(t)
	tuiRuntimeArWrapper(t, dir, "tool-free-recovery.sh", tuiRuntimeArAnswerScript())
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	rt := &tuiRuntime{bridge: b}
	owner, outer := "owner", "abandoned-readonly"
	tuiRuntimeArSave(t, b, owner, outer, "inner-abandoned", true)
	old := &tuiRuntimeSubscriptionRun{bridge: b, token: "old-stalled", owner: owner, model: tuiRuntimeAutoTestModel, outerSession: outer, interactive: true, abandonedByClient: true, parkedAt: time.Now(), pending: map[string]chan tuiRuntimeMcpToolResult{}, shown: []tuiRuntimeShownCall{{id: "write-1", name: "write_file"}}, timer: time.AfterFunc(time.Hour, func() {})}
	defer old.timer.Stop()
	b.mu.Lock()
	b.runs[old.token] = old
	b.liveOuter[tuiRuntimeInteractiveSessionStateKey(owner, outer)] = old
	b.mu.Unlock()
	// A stalled tool_result+new explicit user recovery has no further pending
	// tool callbacks; it must free the old PTY slot and launch tool-free.
	body := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"write_file","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"先读一下配置"},{"role":"assistant","content":"配置已读"},{"role":"user","content":"write it"},{"role":"assistant","content":[{"type":"tool_use","id":"write-1","name":"write_file","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"write-1","content":"written"},{"type":"text","text":"/magpie-recover"}]}]}`)
	code, msg, out := tuiRuntimeArServe(t, rt, b, owner, outer, body)
	if code != 200 || !strings.Contains(out, "OL_ANSWER") {
		t.Fatalf("explicit read-only recovery failed: %d %s %s", code, msg, out)
	}
	old.mu.Lock()
	closed := old.closed
	old.mu.Unlock()
	if !closed {
		t.Fatal("abandoned owner not reaped before recovery")
	}
}

func TestClaudeTUIRuntimeInteractiveExplicitRecoveryRefusesOutstandingTool(t *testing.T) {
	dir := tuiRuntimeArDir(t)
	_ = dir
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	rt := &tuiRuntime{bridge: b}
	owner, outer := "owner", "tool-active"
	tuiRuntimeArSave(t, b, owner, outer, "inner-active", true)
	old := &tuiRuntimeSubscriptionRun{bridge: b, token: "still-working", owner: owner, model: tuiRuntimeAutoTestModel, outerSession: outer, interactive: true, parkedAt: time.Now(), pending: map[string]chan tuiRuntimeMcpToolResult{"write-1": make(chan tuiRuntimeMcpToolResult, 1)}, shown: []tuiRuntimeShownCall{{id: "write-1", name: "write_file"}}, timer: time.AfterFunc(time.Hour, func() {})}
	defer old.timer.Stop()
	b.mu.Lock()
	b.runs[old.token] = old
	b.liveOuter[tuiRuntimeInteractiveSessionStateKey(owner, outer)] = old
	b.mu.Unlock()
	body := []byte(`{"model":"claude-opus-5-5","messages":[{"role":"user","content":"先读一下配置"},{"role":"assistant","content":"配置已读"},{"role":"user","content":"write it"},{"role":"assistant","content":[{"type":"tool_use","id":"write-1","name":"write_file","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"write-1","content":"still working"},{"type":"text","text":"/magpie-recover"}]}]}`)
	code, _, _ := tuiRuntimeArServe(t, rt, b, owner, outer, body)
	if code != 409 {
		t.Fatalf("unsettled callback was interrupted for recovery: %d", code)
	}
	old.mu.Lock()
	closed := old.closed
	old.mu.Unlock()
	if closed {
		t.Fatal("tool-active run aborted")
	}
	old.abort()
}
