package claudetui

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// Exercise a narrow race that can happen when Claude Desktop receives a
// tool_use, starts a tool, and then switches HTTP streams. An old stream's
// cancellation must never dispose the newer continuation. All tools/results
// are in-memory fakes; no Claude account or shell command is launched.
func TestInteractiveFaultInjectionStaleHTTPStreamsCannotKillNewToolResult(t *testing.T) {
	const rounds = 180
	for i := 0; i < rounds; i++ {
		b := newSubscriptionBridge()
		callID := fmt.Sprintf("synthetic-%d", i)
		waiter := make(chan mcpToolResult, 1)
		r := &subscriptionRun{
			bridge: b, token: callID, interactive: true,
			pending: map[string]chan mcpToolResult{callID: waiter},
			shown:   []shownCall{{id: callID, name: "test_no_side_effect"}},
			timer:   time.AfterFunc(time.Hour, func() {}),
		}
		b.mu.Lock()
		b.runs[r.token] = r
		b.calls[callID] = r
		b.mu.Unlock()
		old := r.attach()
		var wg sync.WaitGroup
		wg.Add(2)
		start := make(chan struct{})
		var newer chan Event
		go func() {
			defer wg.Done()
			<-start
			r.letGoFor(old)
		}()
		go func() {
			defer wg.Done()
			<-start
			newer = r.attach()
		}()
		close(start)
		wg.Wait()
		r.mu.Lock()
		closed, current, waiting := r.closed, r.segment, r.pending[callID]
		r.mu.Unlock()
		if closed || current != newer || waiting != waiter {
			t.Fatalf("round %d lost native handoff during cancel/attach: closed=%t new_stream=%t callback=%t", i, closed, current == newer, waiting == waiter)
		}
		// Its registered MCP callback must stay the same owner; it cannot
		// silently switch to another run or replay the external tool.
		b.mu.Lock()
		owner := b.calls[callID]
		b.mu.Unlock()
		if owner != r {
			t.Fatalf("round %d callback owner changed", i)
		}
		r.finish()
		r.timer.Stop()
	}
}

// Repeated network cancellation and process completion may arrive in any
// order; closing one reply channel twice would crash the entire TUI runtime.
func TestInteractiveFaultInjectionCancelVsFinishNeverDoubleCloses(t *testing.T) {
	const rounds = 120
	for i := 0; i < rounds; i++ {
		b := newSubscriptionBridge()
		r := &subscriptionRun{
			bridge: b, token: fmt.Sprintf("finished-%d", i), interactive: true,
			pending: map[string]chan mcpToolResult{},
			shown:   []shownCall{{id: "fake", name: "test_no_side_effect"}},
			timer:   time.AfterFunc(time.Hour, func() {}),
		}
		b.mu.Lock()
		b.runs[r.token] = r
		b.mu.Unlock()
		old := r.attach()
		var wg sync.WaitGroup
		wg.Add(3)
		for _, fn := range []func(){
			func() { r.letGoFor(old) },
			func() { r.letGoFor(old) },
			func() { r.finish() },
		} {
			go func(action func()) { defer wg.Done(); action() }(fn)
		}
		wg.Wait()
		r.mu.Lock()
		closed, detached := r.closed, r.segment == nil
		r.mu.Unlock()
		if !closed || !detached {
			t.Fatalf("round %d left finished session attached", i)
		}
		r.timer.Stop()
	}
}
