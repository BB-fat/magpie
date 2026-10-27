package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClaudeTUIRuntimeExactInteractiveToolGrowthRequiresUnchangedOriginalContext(t *testing.T) {
	base := &tuiRuntimeRequest{Request: Request{Model: "claude-opus-5-5", Effort: "high", System: "same system", Tools: []Tool{{Name: "echo", Schema: json.RawMessage(`{"type":"object"}`)}}}}
	run := &tuiRuntimeSubscriptionRun{interactive: true, owner: "owner", tools: map[string]bool{"echo": true}, effort: base.Effort, interactiveContextSnapshot: tuiRuntimeTestContextSnapshot(base)}
	withMore := *base
	withMore.Tools = append(append([]Tool(nil), base.Tools...), Tool{Name: "extra", Schema: json.RawMessage(`{"type":"object"}`)})
	if run.offers(withMore.Tools) {
		t.Fatal("setup must reproduce the old front-door rejection")
	}
	if !run.exactInteractiveToolGrowth("owner", &withMore, tuiRuntimeByExactID) {
		t.Fatal("an exact-ID handoff with only additive tool schema was rejected")
	}
	if run.exactInteractiveToolGrowth("owner", &withMore, tuiRuntimeByRemap) || run.exactInteractiveToolGrowth("other-owner", &withMore, tuiRuntimeByExactID) {
		t.Fatal("unproven ownership accepted")
	}
	changed := withMore
	changed.System = "different system"
	if run.exactInteractiveToolGrowth("owner", &changed, tuiRuntimeByExactID) {
		t.Fatal("changed system accepted")
	}
	changed = withMore
	changed.Effort = "low"
	if run.exactInteractiveToolGrowth("owner", &changed, tuiRuntimeByExactID) {
		t.Fatal("changed effort accepted")
	}
	changed = withMore
	changed.Tools = []Tool{{Name: "echo", Schema: json.RawMessage(`{"type":"string"}`)}, withMore.Tools[1]}
	if run.exactInteractiveToolGrowth("owner", &changed, tuiRuntimeByExactID) {
		t.Fatal("changed original tool schema accepted")
	}
	changed = withMore
	changed.Tools = []Tool{withMore.Tools[1]}
	if run.exactInteractiveToolGrowth("owner", &changed, tuiRuntimeByExactID) {
		t.Fatal("removed original tool accepted")
	}
	changed = withMore
	changed.Tools = append(append([]Tool(nil), withMore.Tools...), withMore.Tools[1])
	if run.exactInteractiveToolGrowth("owner", &changed, tuiRuntimeByExactID) {
		t.Fatal("duplicate tool name accepted")
	}
	run.readOnlyRecovery = true
	if run.exactInteractiveToolGrowth("owner", &withMore, tuiRuntimeByExactID) {
		t.Fatal("read-only recovery must never gain a tool")
	}
}

func TestClaudeTUIRuntimeExactInteractiveToolGrowthPassesHTTPHandoffWithoutReplay(t *testing.T) {
	const body = `{"model":"claude-opus-5-5","messages":[{"role":"user","content":"run echo"},{"role":"assistant","content":[{"type":"tool_use","id":"known-call","name":"echo","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"known-call","content":"completed"}]}],"tools":[{"name":"echo","input_schema":{"type":"object"}},{"name":"new_tool","input_schema":{"type":"object"}}]}`
	req, err := tuiRuntimeParseRequest([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	owner, outer := "owner", "tool-growth"
	initial := *req
	initial.Tools = req.Tools[:1]
	initial.System = "x-anthropic-billing-header: cc_version=1; cch=old;"
	snapshot := tuiRuntimeInteractiveContextDiagnosticFromRequest(&initial)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	waiter := make(chan tuiRuntimeMcpToolResult, 1)
	run := &tuiRuntimeSubscriptionRun{bridge: b, token: "live-tool-growth", owner: owner, outerSession: outer,
		model: req.Model, interactive: true, effort: req.Effort, interactiveContextSnapshot: &snapshot,
		tools: map[string]bool{"echo": true}, shown: []tuiRuntimeShownCall{{id: "known-call", name: "echo", args: "{}"}},
		pending: map[string]chan tuiRuntimeMcpToolResult{"known-call": waiter}, parkedAt: time.Now(),
		timer: time.AfterFunc(time.Minute, func() {}),
	}
	defer run.timer.Stop()
	run.heard(req.Messages[:2])
	b.mu.Lock()
	b.runs[run.token] = run
	b.calls["known-call"] = run
	b.mu.Unlock()
	w := httptest.NewRecorder()
	go func() {
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			run.mu.Lock()
			ready := run.segment != nil
			run.mu.Unlock()
			if ready {
				run.emit(Event{Kind: KStart, Model: req.Model})
				run.emit(Event{Kind: KText, Text: "answer with added tools"})
				run.emit(Event{Kind: KStop, Stop: "stop"})
				run.endSegment()
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()
	startCalls := 0
	start := func(context.Context, *tuiRuntimeRequest) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
		startCalls++
		return nil, nil, errors.New("must not bootstrap a replacement")
	}
	code, detail := (&tuiRuntime{bridge: b}).serveMessagesInner(tuiRuntimeAnthropicResponder{w, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body)), &Usage{}}, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body)), "Claude Code", req.Model, owner, outer, []byte(body), start)
	if code != 200 || startCalls != 0 || !strings.Contains(w.Body.String(), "answer with added tools") {
		t.Fatalf("live exact-ID tool continuation failed: code=%d starts=%d detail=%q body=%s", code, startCalls, detail, w.Body.String())
	}
	run.mu.Lock()
	added := run.tools["new_tool"]
	run.mu.Unlock()
	if !added {
		t.Fatal("added tool was not registered for subsequent native handoffs")
	}
	if !run.interactiveEquivalentContext(req) {
		t.Fatal("accepted tool growth was not committed as the run's current context")
	}
	if !run.offers(req.Tools) && !run.exactInteractiveToolGrowth(owner, req, tuiRuntimeByExactID) {
		t.Fatal("the accepted expanded tool set was not retained for the next handoff")
	}
}
