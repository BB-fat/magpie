package gateway

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestClaudeTUIRuntimeLiteSearchHandoffKeepsLiveRun(t *testing.T) {
	tuiRuntimeWithConfig(t, &tuiRuntimeConfig{CanSearch: func() bool { return true }})
	rt, b := tuiRuntimeScRuntime(t)
	body := `{"model":"claude-opus-5-5","tools":[{"name":"echo","input_schema":{"type":"object"}},{"type":"web_search_20250305","name":"web_search"}],"messages":[{"role":"user","content":"hello"},{"role":"assistant","content":[{"type":"tool_use","id":"call-1","name":"echo","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call-1","content":"done"}]}]}`
	req, err := tuiRuntimeParseRequest([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	initial := *req
	initial.Tools = append(append([]Tool(nil), req.Tools...), searchTool(req.Tools))
	snapshot := tuiRuntimeInteractiveContextDiagnosticFromRequest(&initial)
	run := &tuiRuntimeSubscriptionRun{
		bridge: b, interactive: true, owner: tuiRuntimeScOwner, model: req.Model,
		tools:                      map[string]bool{"echo": true, "web_search": true},
		shown:                      []tuiRuntimeShownCall{{id: "call-1", name: "echo", args: string(json.RawMessage(`{}`))}},
		interactiveContextSnapshot: &snapshot,
		// A concurrent continuation owns this run. The handler must reach
		// the live-run claim guard, not start a replacement native process.
		resuming: true,
	}
	run.heard(req.Messages[:1])
	b.calls["call-1"] = run
	code, out := tuiRuntimeScServe(t, rt, "", body)
	if code != http.StatusConflict || !strings.Contains(out, "already being resumed") {
		t.Fatalf("live search-enabled handoff fell through to a new run: %d %s", code, out)
	}
	if run.closed {
		t.Fatal("live run was destroyed")
	}
}
