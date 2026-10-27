package gateway

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// The Desktop retries the identical tool-result request after the gateway's
// long response times out. Claude has already received those results; the
// retry must reattach to the pending answer, not deliver them again.
func TestClaudeTUIRuntimeInteractiveAlreadyDeliveredToolResultsReconnect(t *testing.T) {
	body := `{"model":"claude-opus-5-5","messages":[{"role":"user","content":"do two steps"},{"role":"assistant","content":[{"type":"tool_use","id":"first","name":"echo","input":{"n":1}},{"type":"tool_use","id":"second","name":"echo","input":{"n":2}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"first","content":"done one"},{"type":"tool_result","tool_use_id":"second","content":"done two"}]},{"role":"system","content":"session notice"}],"tools":[{"name":"echo","description":"echo","input_schema":{"type":"object"}}]}`
	req, err := tuiRuntimeParseRequest([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer := "owner", "outer-retry"
	run := &tuiRuntimeSubscriptionRun{bridge: b, token: "in-flight", model: req.Model, owner: owner, outerSession: outer, interactive: true, pending: map[string]chan tuiRuntimeMcpToolResult{}, tools: map[string]bool{"echo": true}, shown: []tuiRuntimeShownCall{{id: "first", name: "echo", args: `{"n":1}`}, {id: "second", name: "echo", args: `{"n":2}`}}, parkedAt: time.Now(), timer: time.AfterFunc(1*time.Minute, func() {})}
	defer run.timer.Stop()
	run.heard(req.Messages)
	run.interactiveContextSnapshot = tuiRuntimeTestContextSnapshot(req)
	b.mu.Lock()
	b.runs[run.token] = run
	b.mu.Unlock()
	if !tuiRuntimeInteractiveAlreadyDeliveredHandoff(run, req, []Part{{Kind: ToolResult, CallID: "first"}, {Kind: ToolResult, CallID: "second"}}) {
		t.Fatal("exact already-delivered results must reconnect")
	}
	rt := &tuiRuntime{bridge: b}
	w := httptest.NewRecorder()
	go func() {
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			run.mu.Lock()
			ready := run.segment != nil
			run.mu.Unlock()
			if ready {
				run.emit(Event{Kind: KStart, Model: req.Model})
				run.emit(Event{Kind: KText, Text: "finished after reconnect"})
				run.emit(Event{Kind: KStop, Stop: "stop"})
				run.endSegment()
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()
	started := false
	start := func(context.Context, *tuiRuntimeRequest) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
		started = true
		return nil, nil, nil
	}
	code, detail := rt.serveMessagesInner(tuiRuntimeAnthropicResponder{w, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body)), &Usage{}}, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body)), "Claude Code", req.Model, owner, outer, []byte(body), start)
	if code != 200 || started || !strings.Contains(w.Body.String(), "finished after reconnect") {
		t.Fatalf("retry code=%d started=%t detail=%q body=%s", code, started, detail, w.Body.String())
	}
	if len(run.early) != 0 || len(run.pending) != 0 {
		t.Fatal("retry delivered tool results twice")
	}
	run.abort()
}

func TestClaudeTUIRuntimeInteractiveAlreadyDeliveredRejectsFreshUserText(t *testing.T) {
	msg := []Message{{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "id", Name: "echo"}}}, {Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "id", Text: "done"}}}}
	run := &tuiRuntimeSubscriptionRun{interactive: true, pending: map[string]chan tuiRuntimeMcpToolResult{}, parkedAt: time.Now()}
	run.heard(msg)
	req := &tuiRuntimeRequest{Request: Request{Messages: append(append([]Message{}, msg...), Message{Role: "user", Parts: []Part{{Kind: Text, Text: "a NEW request"}}})}}
	if tuiRuntimeInteractiveAlreadyDeliveredHandoff(run, req, []Part{{Kind: ToolResult, CallID: "id"}}) {
		t.Fatal("reconnection consumed a new user instruction")
	}
}

func TestClaudeTUIRuntimeInteractiveRetryIgnoresOnlyDynamicBudgetNotice(t *testing.T) {
	msg := []Message{{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "id", Name: "echo"}}},
		{Role: "system", Parts: []Part{{Kind: Text, Text: "<total_tokens>30000 tokens left</total_tokens>"}}},
		{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "id", Text: "done"}}}}
	run := &tuiRuntimeSubscriptionRun{interactive: true, pending: map[string]chan tuiRuntimeMcpToolResult{}, parkedAt: time.Now()}
	run.heard(msg)
	req := &tuiRuntimeRequest{Request: Request{Messages: append([]Message(nil), msg...)}}
	req.Messages[1].Parts = []Part{{Kind: Text, Text: "<total_tokens>28000 tokens left</total_tokens>"}}
	req.Messages = append(req.Messages, Message{Role: "system", Parts: []Part{{Kind: Text, Text: "<total_tokens>27000 tokens left</total_tokens>"}}})
	if !tuiRuntimeInteractiveAlreadyDeliveredHandoff(run, req, []Part{{Kind: ToolResult, CallID: "id"}}) {
		t.Fatal("counter-only notice altered a proven same tool-result retry")
	}
	req.Messages[len(req.Messages)-1].Parts[0].Text = "injected new instruction"
	if tuiRuntimeInteractiveAlreadyDeliveredHandoff(run, req, []Part{{Kind: ToolResult, CallID: "id"}}) {
		t.Fatal("new system instruction was silently swallowed")
	}
}

func TestClaudeTUIRuntimeInteractiveReplyWhileDisconnectedInvalidatesStaleShownCall(t *testing.T) {
	run := &tuiRuntimeSubscriptionRun{shown: []tuiRuntimeShownCall{{id: "old-tool"}}}
	run.emit(Event{Kind: KStart})
	if len(run.shown) != 0 {
		t.Fatal("reply progressed without client, but old call still looked reattachable")
	}
}
