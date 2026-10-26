package claudetui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestInteractiveHistoricalBootstrapDesktopSystemTail(t *testing.T) {
	// Claude Desktop 2.1.286 sends environment/model notices as message-role
	// system rows, including after the newly submitted user turn.
	body := []byte(`{"model":"claude-opus-5-5","system":"TOP_SYSTEM","messages":[{"role":"user","content":"OLD_USER"},{"role":"system","content":"OLD_ENVIRONMENT"},{"role":"assistant","content":"OLD_REPLY"},{"role":"user","content":"CURRENT_PREFIX"},{"role":"system","content":"CURRENT_ENVIRONMENT"},{"role":"user","content":"CURRENT_SUFFIX"},{"role":"system","content":[{"type":"text","text":"You are powered by the model named Opus 5.5. The exact model ID is claude-opus-5-5. Assistant knowledge cutoff is June 2026."}]}]}`)
	req, err := parseAnthropic(body)
	if err != nil {
		t.Fatal(err)
	}
	prompt, err := renderInteractiveHistoricalBootstrap(req)
	if err != nil {
		t.Fatal(err)
	}
	at := strings.Index(prompt, "<current_user_request>\n")
	if at < 0 || !strings.Contains(prompt[at:], "CURRENT_PREFIX\nCURRENT_SUFFIX") {
		t.Fatalf("current user fragments lost: %s", prompt)
	}
	for _, want := range []string{"TOP_SYSTEM", "CURRENT_ENVIRONMENT", "You are powered by the model named Opus 5.5.", "OLD_REPLY"} {
		if !strings.Contains(prompt[:at], want) {
			t.Fatalf("context lost %q", want)
		}
	}
	if strings.Contains(prompt[at:], "CURRENT_ENVIRONMENT") {
		t.Fatal("system notice became user text")
	}
	if len(req.Messages) != 7 || req.Messages[6].Role != "system" || req.System != "TOP_SYSTEM" {
		t.Fatal("wire request mutated, invalidating live/checkpoint hashes")
	}
}

func TestInteractiveToolContinuationDesktopSystemRows(t *testing.T) {
	user := Message{Role: "user", Parts: []Part{{Kind: Text, Text: "run tool"}}}
	sys := Message{Role: "system", Parts: []Part{{Kind: Text, Text: "<total_tokens>15000000 tokens left</total_tokens>"}}}
	call := Message{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "c1", Name: "echo_test"}}}
	result := Message{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "c1", Text: "DONE"}}}
	req := &Request{Messages: []Message{user, sys, call, sys, result, sys}}
	if err := validateInteractiveToolContinuation(req); err != nil {
		t.Fatal(err)
	}
	if _, err := renderInteractiveHistoricalBootstrap(req); err == nil {
		t.Fatal("orphan tool result was allowed to bootstrap")
	}
	// An exact live ID can still prove ownership when Desktop omits the call row.
	req.Messages = []Message{user, sys, result, sys}
	run := &subscriptionRun{shown: []shownCall{{id: "c1", name: "echo_test"}}}
	run.heard(req.Messages[:2])
	if err := validateInteractiveExactIDContinuation(req, run, []Part{result.Parts[0]}); err != nil {
		t.Fatal(err)
	}
	req.Messages = []Message{user, sys, {Role: "user", Parts: []Part{{Kind: Text, Text: "new un-heard request"}}}, result, sys}
	if err := validateInteractiveExactIDContinuation(req, run, []Part{result.Parts[0]}); err == nil {
		t.Fatal("un-heard text before result accepted")
	}
}

func TestInteractiveHistoricalBootstrapLiteralDesktopCapture(t *testing.T) {
	path := os.Getenv("MAGPIE_DESKTOP_CAPTURE_TEST")
	if path == "" {
		t.Skip("one-off private Desktop request capture")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	req, err := parseAnthropic(body)
	if err != nil {
		t.Fatal(err)
	}
	prompt, err := renderInteractiveHistoricalBootstrap(req)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prompt, "<current_user_request>\n") || !strings.Contains(prompt, "刚才会话运行出错了") {
		t.Fatal("literal Desktop recovery request missing")
	}
}

func TestInteractiveSystemTailDoesNotReviveOrphanToolResult(t *testing.T) {
	setWrapper(t, "/not-a-wrapper")
	s := newTestServer()
	body := []byte(`{"model":"claude-opus-5-5","messages":[{"role":"user","content":"run tool"},{"role":"assistant","content":[{"type":"tool_use","id":"orphan","name":"echo_test","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"orphan","content":"DONE"}]},{"role":"system","content":"<total_tokens>15000000 tokens left</total_tokens>"}]}`)
	called := false
	start := func(context.Context, *Request) (*subscriptionRun, <-chan Event, error) {
		called = true
		return nil, nil, nil
	}
	w := httptest.NewRecorder()
	code, msg := s.serveSubscription(w, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body))), protoAnthropic, "Claude Code", "claude-opus-5-5", "owner", body, &Usage{}, start)
	if code != http.StatusConflict || called || !strings.Contains(msg, "no matching live MCP run") {
		t.Fatalf("orphan reached bootstrap: %d %s started=%t", code, msg, called)
	}
}

func TestInteractiveSystemMessageCannotHideToolTraffic(t *testing.T) {
	for _, kind := range []Kind{ToolCall, ToolResult, Image} {
		req := &Request{Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "continue"}}}, {Role: "system", Parts: []Part{{Kind: kind, ID: "orphan", CallID: "orphan"}}}}}
		if err := validateInteractiveRequest(req); err == nil {
			t.Fatalf("non-text system kind %v accepted", kind)
		}
		if _, err := renderInteractiveHistoricalBootstrap(req); err == nil {
			t.Fatalf("non-text system kind %v bootstrapped", kind)
		}
	}
}

func TestInteractiveAbandonedToolRecoveryDesktopSystemRows(t *testing.T) {
	user := func(text string) Message { return Message{Role: "user", Parts: []Part{{Kind: Text, Text: text}}} }
	sys := Message{Role: "system", Parts: []Part{{Kind: Text, Text: "<total_tokens>15000000 tokens left</total_tokens>"}}}
	result := Message{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "stale", Text: "old result"}}}
	errorRow := Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: "API Error: 409 US VPS Claude: Claude interactive migration: no matching live MCP run"}}}
	for _, tc := range []struct {
		name     string
		messages []Message
		want     int
	}{
		{"tail", []Message{user("summary"), result, errorRow, user("continue"), sys}, 3},
		{"between", []Message{user("summary"), result, sys, errorRow, sys, user("continue"), sys}, 5},
		{"interrupted", []Message{user("summary"), result, sys, errorRow, user("continue"), sys, user("[Request interrupted by user]"), sys}, 4},
		{"completed", []Message{user("summary"), result, sys, {Role: "assistant", Parts: []Part{{Kind: Text, Text: "completed answer"}}}, user("new question"), sys}, -1},
		{"unsafe system", []Message{user("summary"), result, errorRow, user("continue"), {Role: "system", Parts: []Part{{Kind: ToolResult, CallID: "hidden"}}}}, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := interactiveAbandonedToolRecoveryIndex(&Request{Messages: tc.messages}); got != tc.want {
				t.Fatalf("Desktop recovery index = %d, want original wire index %d", got, tc.want)
			}
		})
	}
}

func TestInteractiveFirstTurnDesktopSystemNotices(t *testing.T) {
	req := &Request{System: "TOP_SYSTEM", Messages: []Message{
		{Role: "system", Parts: []Part{{Kind: Text, Text: "BEFORE_USER"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "DESCRIBE_IMAGE"}, {Kind: Image, Data: "AA==", MediaType: "image/png"}}},
		{Role: "system", Parts: []Part{{Kind: Text, Text: "AFTER_USER"}}},
	}}
	if err := validateInteractiveFirstTurn(req); err != nil {
		t.Fatal(err)
	}
	prompt, images, err := renderInteractiveFirstPromptWithImages(req)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"TOP_SYSTEM", "BEFORE_USER", "AFTER_USER", "DESCRIBE_IMAGE"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("single-turn context missing %q: %s", want, prompt)
		}
	}
	if len(images) != 1 || len(req.Messages) != 3 || req.System != "TOP_SYSTEM" {
		t.Fatalf("image or original wire context lost: %d %#v", len(images), req)
	}
}
