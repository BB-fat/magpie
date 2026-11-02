package gateway

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

// The real CLI must send the worker's cached messages, system and tools
// unchanged on the summary fork. A copied file alone does not prove that.
func TestClaudeTUIRuntimeRealActivitySummary(t *testing.T) {
	if os.Getenv("MAGPIE_CLAUDE_TUI_REAL") != "1" {
		t.Skip("set MAGPIE_CLAUDE_TUI_REAL=1 and the real CLI programs")
	}
	up, profile, bridge := tuiRuntimeRealSetup(t, "magpie")
	b := bridge()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	const outer = "worker-with-activity-summary"
	first := `{"model":"` + tuiRuntimeRealModel + `","max_tokens":256,"stream":true,"tools":` + tuiRuntimeRealTools + `,"messages":[{"role":"user","content":"call echo with hi"}]}`
	out, code, msg := tuiRuntimeRealServe(t, b, ctx, profile, outer, first)
	call := out.tool()
	if code != 200 || call.name != "echo" {
		t.Fatalf("worker's tool round: %d %s %+v", code, msg, call)
	}
	before := up.all()
	worker := tuiRuntimeRealMain(t, before, "call echo with hi")
	history := `{"role":"user","content":"call echo with hi"},` +
		`{"role":"assistant","content":[{"type":"tool_use","id":` + tuiRuntimeJsonString(call.id) + `,"name":"echo","input":` + call.args + `}]},` +
		`{"role":"user","content":[{"type":"tool_result","tool_use_id":` + tuiRuntimeJsonString(call.id) + `,"content":"ECHOED"}`
	activity := strings.Repeat("[Image: original 3200x2000, displayed at 2000x1250. Multiply coordinates by 1.60 to map to original image.]\n", 3) + tuiRuntimeActivitySummaryPrompt
	summary := `{"model":"` + tuiRuntimeRealModel + `","max_tokens":256,"stream":true,"tools":` + tuiRuntimeRealTools + `,"messages":[` + history +
		`,{"type":"text","text":` + tuiRuntimeJsonString(activity) + `}]}]}`
	if out, code, msg := tuiRuntimeRealServe(t, b, ctx, profile, outer, summary); code != 200 || out.text() != "DONE: ECHOED" {
		t.Fatalf("summary: %d %s %q", code, msg, out.text())
	}
	side := tuiRuntimeRealMain(t, up.all()[len(before):], "Describe your most recent action")
	if tuiRuntimeRealSession(t, side) == tuiRuntimeRealSession(t, worker) {
		t.Fatal("activity summary ran in the worker's native session")
	}
	var original, fork map[string]any
	if json.Unmarshal(worker, &original) != nil || json.Unmarshal(side, &fork) != nil {
		t.Fatal("invalid captured upstream request")
	}
	clean := func(v any) any {
		return tuiRuntimeRealUnmarked([]any{v})[0]
	}
	if !reflect.DeepEqual(clean(original["tools"]), clean(fork["tools"])) {
		t.Fatal("activity summary changed the worker's upstream tool definitions")
	}
	system := func(v any) any {
		blocks, ok := clean(v).([]any)
		if !ok {
			return v
		}
		for _, item := range blocks {
			if block, ok := item.(map[string]any); ok {
				if text, ok := block["text"].(string); ok {
					block["text"] = tuiRuntimeWithoutBillingHeader(text)
				}
			}
		}
		return blocks
	}
	if !reflect.DeepEqual(system(original["system"]), system(fork["system"])) {
		t.Fatal("activity summary changed the worker's upstream system prefix")
	}
	prior, current := original["messages"].([]any), fork["messages"].([]any)
	if len(current) <= len(prior) || !reflect.DeepEqual(clean(prior), clean(current[:len(prior)])) {
		t.Fatal("activity summary rebuilt the prior upstream messages")
	}
	// The original MCP waiter must still take the actual results afterwards.
	results := `{"model":"` + tuiRuntimeRealModel + `","max_tokens":256,"stream":true,"tools":` + tuiRuntimeRealTools + `,"messages":[` + history + `]}]}`
	mark := len(up.all())
	if out, code, msg := tuiRuntimeRealServe(t, b, ctx, profile, outer, results); code != 200 || !strings.Contains(out.text(), "DONE: ECHOED") {
		t.Fatalf("worker after its activity summary: %d %s %q", code, msg, out.text())
	}
	next := tuiRuntimeRealMain(t, up.all()[mark:], "ECHOED")
	if tuiRuntimeRealSession(t, next) != tuiRuntimeRealSession(t, worker) {
		t.Fatal("activity summary displaced the worker's original MCP continuation")
	}
	tuiRuntimeWaitGone(t, b, profile)
}
