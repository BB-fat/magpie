//go:build darwin || linux

package claudetui

import (
	"strings"
	"testing"
)

// A client-side compaction can rewrite the history while a PTY tool call is
// still out. The live run is let go, the checkpoint stays dirty, and the
// trailing tool result the client keeps sending matches no run. With the opt-in
// seam on, that exact shape becomes one bounded, tool-free read-only recovery
// instead of an endless orphan-result refusal.
func TestOrphanToolHandoffAfterCompactionRecoversReadOnly(t *testing.T) {
	dir := arDir(t)
	arWrapper(t, dir, "orphan-recovery.sh", arAnswerScript())
	t.Setenv(autoOrphanRecoveryEnv, "1")
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	rt := &runtimeTUI{bridge: b}
	owner, outer := "owner", "compacted-orphan"
	arSave(t, b, owner, outer, "inner-old", true)
	body := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"bash","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"[compacted summary] earlier work"},{"role":"assistant","content":"ack"},{"role":"user","content":[{"type":"tool_result","tool_use_id":"orphan-1","content":"DONE"}]}]}`)
	code, msg, out := arServe(t, rt, b, owner, outer, body)
	if code != 200 || !strings.Contains(out, "OL_ANSWER") {
		t.Fatalf("orphan handoff not recovered: %d %s %s", code, msg, out)
	}
}

// Without the opt-in the same request keeps the historical fail-closed answer.
func TestOrphanToolHandoffWithoutOptInStillRefuses(t *testing.T) {
	dir := arDir(t)
	arWrapper(t, dir, "orphan-refuse.sh", arAnswerScript())
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	rt := &runtimeTUI{bridge: b}
	owner, outer := "owner", "compacted-orphan-off"
	arSave(t, b, owner, outer, "inner-old", true)
	body := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"bash","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"[compacted summary] earlier work"},{"role":"assistant","content":"ack"},{"role":"user","content":[{"type":"tool_result","tool_use_id":"orphan-1","content":"DONE"}]}]}`)
	code, msg, _ := arServe(t, rt, b, owner, outer, body)
	if code != 409 || !strings.Contains(msg, "no matching live MCP run") {
		t.Fatalf("opt-in off still recovered: %d %s", code, msg)
	}
}

// A clean checkpoint is not a compaction: recovery must not fire.
func TestOrphanToolHandoffRecoveryNeedsDirtyCompactedCheckpoint(t *testing.T) {
	dir := arDir(t)
	arWrapper(t, dir, "orphan-clean.sh", arAnswerScript())
	t.Setenv(autoOrphanRecoveryEnv, "1")
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	rt := &runtimeTUI{bridge: b}
	owner, outer := "owner", "compacted-orphan-clean"
	arSave(t, b, owner, outer, "inner-old", false)
	body := []byte(`{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"bash","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"[compacted summary] earlier work"},{"role":"assistant","content":"ack"},{"role":"user","content":[{"type":"tool_result","tool_use_id":"orphan-1","content":"DONE"}]}]}`)
	code, msg, _ := arServe(t, rt, b, owner, outer, body)
	if code != 409 {
		t.Fatalf("clean checkpoint recovered as compaction: %d %s", code, msg)
	}
}
