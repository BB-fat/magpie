package claudetui

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestInteractiveHistoryMismatchDiagnostic(t *testing.T) {
	user := func(s string) Message { return Message{Role: "user", Parts: []Part{{Kind: Text, Text: s}}} }
	system := func(s string) Message { return Message{Role: "system", Parts: []Part{{Kind: Text, Text: s}}} }
	baseline := []Message{user("PRIVATE_PROMPT"), system("PRIVATE_ENVIRONMENT"), user("CURRENT_REQUEST")}
	for _, tc := range []struct {
		name     string
		messages []Message
		want     string
	}{
		{"changed environment", []Message{baseline[0], system("SECRET_CHANGED_ENVIRONMENT"), baseline[2]}, "normalized_message=1 old_role=system new_role=system"},
		{"changed user", []Message{user("SECRET_CHANGED_USER"), baseline[1], baseline[2]}, "normalized_message=0 old_role=user new_role=user"},
		{"shorter history", baseline[:2], "normalized_message=2 old_role=user new_role=missing"},
		{"new role", []Message{system("PRIVATE_PROMPT"), baseline[1], baseline[2]}, "normalized_message=0 old_role=user new_role=system"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &subscriptionRun{interactive: true}
			r.heard(baseline)
			if r.follows(tc.messages) {
				t.Fatal("rewritten history accepted")
			}
			got := r.historyMismatchDiagnostic(tc.messages)
			if !strings.Contains(got, tc.want) {
				t.Fatalf("diagnostic = %s, want %s", got, tc.want)
			}
			for _, secret := range []string{"PRIVATE", "SECRET", "CURRENT_REQUEST"} {
				if strings.Contains(got, secret) {
					t.Fatalf("diagnostic leaked text: %s", got)
				}
			}
			if !r.follows(baseline) {
				t.Fatal("diagnostic changed stored continuity baseline")
			}
		})
	}
}

func TestInteractiveHistoryDiagnosticIgnoresOnlyExistingBudgetRule(t *testing.T) {
	u := Message{Role: "user", Parts: []Part{{Kind: Text, Text: "unchanged"}}}
	budget := func(s string) Message { return Message{Role: "system", Parts: []Part{{Kind: Text, Text: s}}} }
	r := &subscriptionRun{interactive: true}
	tail := Message{Role: "user", Parts: []Part{{Kind: Text, Text: "current request"}}}
	r.heard([]Message{u, budget("<total_tokens>100 tokens left</total_tokens>"), tail})
	changed := []Message{u, budget("<total_tokens>90 tokens left</total_tokens>"), tail}
	if !r.follows(changed) {
		t.Fatal("existing budget exception regressed")
	}
	if got := r.historyMismatchDiagnostic(changed); got != "normalized_history_unchanged" {
		t.Fatal(got)
	}
	changed[1] = budget("<total_tokens>90 tokens left</total_tokens> extra instruction")
	if r.follows(changed) {
		t.Fatal("broadened budget exception")
	}
}

func TestInteractiveChangedHistoryExplains409WithoutReplay(t *testing.T) {
	s := newTestServer()
	defer bridges.Delete(s.subscription)
	body := []byte(`{"model":"claude-opus-5-5","messages":[{"role":"user","content":"changed prompt"},{"role":"assistant","content":[{"type":"tool_use","id":"call-one","name":"echo","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call-one","content":"already executed"}]}]}`)
	req, err := parseAnthropic(body)
	if err != nil {
		t.Fatal(err)
	}
	r := &subscriptionRun{bridge: s.subscription, token: "test-history", interactive: true, owner: "owner", model: req.Model,
		pending:               map[string]chan mcpToolResult{"call-one": make(chan mcpToolResult)},
		shown:                 []shownCall{{id: "call-one", name: "echo"}},
		interactiveContextKey: turnKey("owner", req, nil),
	}
	r.heard([]Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "original prompt"}}}})
	s.subscription.runs[r.token] = r
	s.subscription.calls["call-one"] = r
	defer r.abort()
	started := false
	start := func(context.Context, *Request) (*subscriptionRun, <-chan Event, error) {
		started = true
		return nil, nil, errors.New("unexpected bootstrap")
	}
	w := httptest.NewRecorder()
	code, msg := s.serveSubscription(w, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body))), protoAnthropic, "Claude Code", req.Model, "owner", body, &Usage{}, start)
	if code != http.StatusConflict || started {
		t.Fatalf("status=%d replay=%t", code, started)
	}
	if !strings.Contains(msg, "history changed during tool handoff") || !strings.Contains(msg, "new user message") {
		t.Fatalf("unhelpful history rejection: %s", msg)
	}
	r.mu.Lock()
	closed := r.closed
	r.mu.Unlock()
	if !closed {
		t.Fatal("uncertain run left active")
	}
}

func TestInteractiveMatchedHistoryDriftOptInStartsOnlyReadOnlyReconstruction(t *testing.T) {
	t.Setenv("MAGPIE_CLAUDE_AUTO_READONLY_HISTORY_DRIFT", "1")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	rt := &runtimeTUI{bridge: b}
	const owner, outer = "owner", "existing-outer"
	// A new user message after the orphan-looking tool result makes the
	// original request eligible for abandoned-turn tail trimming. Automatic
	// recovery must not accidentally trim its own tool-free sentinel.
	body := []byte(`{"model":"claude-opus-5-5","tools":[{"name":"echo","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"FIRST_IMMUTABLE_USER_TEXT"},{"role":"user","content":"CHANGED_SECOND_USER_TEXT"},{"role":"assistant","content":[{"type":"tool_use","id":"call-one","name":"echo","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call-one","content":"already executed"}]},{"role":"user","content":"NEW_USER_MESSAGE_NOT_A_TOOL_RESULT"}]}`)
	req, err := parseAnthropic(body)
	if err != nil {
		t.Fatal(err)
	}
	if got := b.seedOuterThread(owner, outer, firstWords(req)); got != firstWords(req) {
		t.Fatal("failed to seed canonical outer thread")
	}
	if err := b.saveInteractiveSession(owner, outer, interactiveSessionEntry{
		SessionID: "inner-existing", Model: req.Model, ThreadKey: firstWords(req),
		ConvKey: "known-persistent-conversation", ContextKey: turnKey(owner, req, nil), Dirty: true,
	}); err != nil {
		t.Fatal(err)
	}
	run := &subscriptionRun{bridge: b, token: "live-run", interactive: true, owner: owner, outerSession: outer,
		model: req.Model, parkedAt: time.Now(),
		pending:               map[string]chan mcpToolResult{"call-one": make(chan mcpToolResult)},
		shown:                 []shownCall{{id: "call-one", name: "echo"}},
		tools:                 map[string]bool{"echo": true},
		interactiveContextKey: turnKey(owner, req, nil),
	}
	run.heard([]Message{
		req.Messages[0],
		{Role: "user", Parts: []Part{{Kind: Text, Text: "ORIGINAL_SECOND_USER_TEXT"}}},
	})
	b.runs[run.token] = run
	b.calls["call-one"] = run
	defer run.abort()
	started := false
	start := func(_ context.Context, got *Request) (*subscriptionRun, <-chan Event, error) {
		started = true
		if len(got.Tools) != 0 || got.WebSearch || got.ToolChoice != "" || len(got.Schema) != 0 {
			t.Fatal("unsafe tool surface remained on history-drift reconstruction")
		}
		last := got.Messages[len(got.Messages)-1]
		if last.Role != "user" || len(last.Parts) != 1 || last.Parts[0].Text != "/magpie-recover" {
			t.Fatal("read-only recovery sentinel was not isolated at end")
		}
		if interactiveReadOnlyOptInIndex(got) < 0 {
			t.Fatal("the next stage would not recognize the tool-free recovery")
		}
		prompt, err := renderInteractiveReadOnlyBootstrap(got)
		if err != nil || prompt == "" || strings.Contains(prompt, "already executed") {
			t.Fatalf("unsafe or invalid recovery bootstrap: err=%v empty=%t exposed_tool_result=%t", err, prompt == "", strings.Contains(prompt, "already executed"))
		}
		return nil, nil, errors.New("mock read-only start intentionally blocked")
	}
	w := httptest.NewRecorder()
	code, _ := rt.serveMessagesInner(anthropicResponder{w, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body))), &Usage{}}, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body))), "Claude Code", req.Model, owner, outer, body, start)
	if code != http.StatusBadGateway || !started {
		t.Fatalf("expected one tool-free start attempt, got HTTP %d started=%t", code, started)
	}
	run.mu.Lock()
	closed := run.closed
	run.mu.Unlock()
	if !closed {
		t.Fatal("original callback was not terminated before the recovery")
	}
}

func TestInteractiveRecoveryDiagnosticIsBoundedAndPrivate(t *testing.T) {
	r := &Request{}
	for i := 0; i < 10; i++ {
		r.Messages = append(r.Messages, Message{Role: "user", Parts: []Part{{Kind: Text, Text: "PRIVATE_PROMPT"}, {Kind: ToolResult, CallID: "SECRET_TOOL_ID", Text: "SECRET_RESULT"}}})
	}
	s := interactiveRecoveryDiagnostic(r)
	if strings.Count(s, "row=") != 6 || !strings.Contains(s, "row=4 role=user text=1 calls=0 results=1") || strings.Contains(s, "PRIVATE") || strings.Contains(s, "SECRET") {
		t.Fatalf("unsafe structural summary: %s", s)
	}
}
