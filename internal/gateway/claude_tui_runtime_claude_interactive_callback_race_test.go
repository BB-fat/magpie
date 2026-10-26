package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClaudeTUIRuntimeInteractiveShownToolResultPrecedesSlowCallback(t *testing.T) {
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	req := &Request{Model: "claude-opus-5-5", Messages: []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "call echo"}}},
		{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "shown-call", Name: "echo", Args: json.RawMessage(`{"value":"HELLO"}`)}}},
		{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "shown-call", Text: "HELLO"}}},
	}}
	run := &tuiRuntimeSubscriptionRun{bridge: b, token: "slow-callback", interactive: true, owner: "owner", model: req.Model,
		pending: map[string]chan tuiRuntimeMcpToolResult{}, tools: map[string]bool{"echo": true},
		shown: []tuiRuntimeShownCall{{id: "shown-call", name: "echo", args: `{"value":"HELLO"}`}}, parkedAt: time.Now()}
	run.heard(req.Messages[:1])
	b.runs[run.token] = run
	defer run.abort()
	// A TUI transcript can reach the gateway before Claude schedules its MCP
	// callback. Ownership is already known from the exact call shown to the client.
	callback := make(chan *httptest.ResponseRecorder, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /cb/{token}", b.mcpCall)
	go func() {
		time.Sleep(800 * time.Millisecond)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/cb/slow-callback", strings.NewReader(`{"tool_call_id":"shown-call","name":"echo","arguments":{"value":"HELLO"}}`)))
		callback <- w
	}()
	found, results, how := b.match(req)
	if found != run || how != tuiRuntimeByExactID || len(results) != 1 {
		t.Fatalf("shown live call lost before callback registration: found=%t how=%q results=%d", found == run, how, len(results))
	}
	if !run.claimResume() {
		t.Fatal("shown call was not claimable")
	}
	events, err := run.continueWith(results, nil)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case w := <-callback:
		if w.Code != 200 || !strings.Contains(w.Body.String(), "HELLO") {
			t.Fatalf("delayed callback did not receive queued result: %d %s", w.Code, w.Body)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("delayed callback did not complete")
	}
	run.endSegment()
	for range events {
	}
}

func TestClaudeTUIRuntimeInteractiveShownMatchRejectsUnownedAndStaleResults(t *testing.T) {
	for _, name := range []string{"unknown id", "duplicate results", "partial results", "changed history", "wrong model", "closed", "already claimed", "not parked", "headless", "ambiguous", "duplicate shown ids"} {
		t.Run(name, func(t *testing.T) {
			b := tuiRuntimeNewSubscriptionBridge()
			defer tuiRuntimeBridges.Delete(b)
			run := &tuiRuntimeSubscriptionRun{interactive: true, model: "model", told: "heard", parkedAt: time.Now(), shown: []tuiRuntimeShownCall{{id: "one"}, {id: "two"}}}
			b.runs["test"] = run
			req := &Request{Model: "model"}
			results := []Part{{Kind: ToolResult, CallID: "one"}, {Kind: ToolResult, CallID: "two"}}
			before := map[string]bool{"heard": true}
			switch name {
			case "unknown id":
				results[1].CallID = "unknown"
			case "duplicate results":
				results[1].CallID = "one"
			case "partial results":
				results = results[:1]
			case "changed history":
				before = nil
			case "wrong model":
				req.Model = "other"
			case "closed":
				run.closed = true
			case "already claimed":
				run.resuming = true
			case "not parked":
				run.parkedAt = time.Time{}
			case "headless":
				run.interactive = false
			case "ambiguous":
				b.runs["other"] = &tuiRuntimeSubscriptionRun{interactive: true, model: run.model, told: run.told, parkedAt: run.parkedAt, shown: run.shown}
			case "duplicate shown ids":
				run.shown[1].id = "one"
			}
			if b.matchInteractiveShown(req, results, before) != nil {
				t.Fatal("unsafe result used the delayed-callback recovery path")
			}
		})
	}
}

func TestClaudeTUIRuntimeInteractiveMatchedContinuationRequiresAllNativeShownCalls(t *testing.T) {
	for _, kind := range []string{"partial", "forged sibling", "complete remapped"} {
		t.Run(kind, func(t *testing.T) {
			req := &Request{Messages: []Message{
				{Role: "user", Parts: []Part{{Kind: Text, Text: "call echo"}}},
				{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "client-one", Name: "echo"}}},
				{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "client-one", Text: "FIRST"}}},
			}}
			run := &tuiRuntimeSubscriptionRun{shown: []tuiRuntimeShownCall{{id: "native-one"}, {id: "native-two"}}}
			results := []Part{{Kind: ToolResult, CallID: "native-one", Text: "FIRST"}}
			if kind != "partial" {
				req.Messages[1].Parts = append(req.Messages[1].Parts, Part{Kind: ToolCall, ID: "client-two", Name: "echo"})
				req.Messages[2].Parts = append(req.Messages[2].Parts, Part{Kind: ToolResult, CallID: "client-two", Text: "SECOND"})
				id := "native-two"
				if kind == "forged sibling" {
					id = "unrelated"
				}
				results = append(results, Part{Kind: ToolResult, CallID: id, Text: "SECOND"})
			}
			err := tuiRuntimeValidateInteractiveMatchedContinuation(req, run, results, tuiRuntimeByRemap)
			if kind == "complete remapped" && err != nil {
				t.Fatalf("complete native result set rejected: %v", err)
			}
			if kind != "complete remapped" && err == nil {
				t.Fatalf("unsafe %s result set accepted", kind)
			}
		})
	}
}
