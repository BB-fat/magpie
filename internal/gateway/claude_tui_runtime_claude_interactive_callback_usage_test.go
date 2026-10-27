package gateway

import (
	"encoding/json"
	"io"
	"testing"
	"time"
)

// The first of a message's parallel calls reaches magpie by its MCP
// callback, and Claude Code writes the message's rows only after the
// callback was shown. The rows come past unstreamedQuiet: they still
// belong to the request that showed the callback — the message's usage
// and its second call with it, not the next request's (16:43 on au-exit,
// out=1592 counted on a 196ms request that asked Anthropic nothing).
func TestClaudeTUIRuntimeInteractiveCallbackFirstKeepsItsMessageUsage(t *testing.T) {
	was := tuiRuntimeUnstreamedQuiet
	tuiRuntimeUnstreamedQuiet = 40 * time.Millisecond
	t.Cleanup(func() { tuiRuntimeUnstreamedQuiet = was })
	rd, wr := io.Pipe()
	defer rd.Close()
	defer wr.Close()
	run := &tuiRuntimeSubscriptionRun{interactive: true, model: "claude-opus-5-5", tools: map[string]bool{"apply_patch": true, "exec_command": true},
		pending: map[string]chan tuiRuntimeMcpToolResult{"call-patch": make(chan tuiRuntimeMcpToolResult, 1)}}
	ch := run.attach()
	done := make(chan struct{})
	go func() { run.readInteractiveOneTurn(rd); close(done) }()
	var callback func(string, string, json.RawMessage)
	for until := time.Now().Add(2 * time.Second); callback == nil && time.Now().Before(until); time.Sleep(2 * time.Millisecond) {
		run.mu.Lock()
		callback = run.onCall
		run.mu.Unlock()
	}
	if callback == nil {
		t.Fatal("native callback listener never initialized")
	}
	callback("call-patch", "apply_patch", json.RawMessage(`{"input":"*** Begin Patch"}`))
	var calls []string
	var usage Usage
	stops := 0
	limit := time.NewTimer(6 * time.Second)
	defer limit.Stop()
	wrote := false
loop:
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				break loop
			}
			switch ev.Kind {
			case KStart, KUsage:
				usage = ev.Usage
			case KToolStart:
				calls = append(calls, ev.ID)
				if !wrote {
					wrote = true
					// well past unstreamedQuiet, well inside callbackRowWait
					go func() {
						time.Sleep(10 * tuiRuntimeUnstreamedQuiet)
						_, _ = io.WriteString(wr, `{"type":"assistant","uuid":"row-patch","message":{"id":"msg-parallel","model":"claude-opus-5-5","usage":{"input_tokens":2,"output_tokens":1592,"cache_read_input_tokens":63519},"content":[{"type":"tool_use","id":"call-patch","name":"mcp__magpie__apply_patch","input":{"input":"*** Begin Patch"}}]}}`+"\n"+
							`{"type":"assistant","uuid":"row-exec","message":{"id":"msg-parallel","model":"claude-opus-5-5","usage":{"input_tokens":2,"output_tokens":1592,"cache_read_input_tokens":63519},"content":[{"type":"tool_use","id":"call-exec","name":"mcp__magpie__exec_command","input":{"cmd":"git diff"}}]}}`+"\n")
					}()
				}
			case KStop:
				if ev.Stop == "tool" {
					stops++
				}
			case KError:
				t.Fatalf("error event: %s", ev.Text)
			}
		case <-limit.C:
			t.Fatal("tool turn never closed")
		}
	}
	if len(calls) != 2 || calls[0] != "call-patch" || calls[1] != "call-exec" || stops != 1 {
		t.Fatalf("the message's calls were split: calls=%v stops=%d", calls, stops)
	}
	if usage.Output != 1592 || usage.CacheRead != 63519 {
		t.Fatalf("the request that showed the callback lost its message's usage: %+v", usage)
	}
	_ = wr.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("reader was not reaped")
	}
}
