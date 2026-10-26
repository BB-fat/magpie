package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestClaudeTUIRuntimeInteractiveTransportRejectsUnsupportedRequests(t *testing.T) {
	first := &Request{Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "hi"}}}}}
	if err := tuiRuntimeValidateInteractiveFirstTurn(first); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		req  Request
	}{
		{"replay", Request{Messages: []Message{{Role: "user"}, {Role: "assistant"}, {Role: "user"}}}},
		{"tool result", Request{Messages: []Message{{Role: "user", Parts: []Part{{Kind: ToolResult}}}}}},
		{"file", Request{Messages: []Message{{Role: "user", Parts: []Part{{Kind: File, MediaType: "application/pdf", Data: "AA=="}}}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tuiRuntimeValidateInteractiveFirstTurn(&tc.req); err == nil {
				t.Fatal("unsupported request accepted by unfinished PTY transport")
			}
		})
	}
	withTools := *first
	withTools.Tools = []Tool{{Name: "echo_test"}}
	if err := tuiRuntimeValidateInteractiveFirstTurn(&withTools); err != nil {
		t.Fatalf("initial MCP tool offer was rejected: %v", err)
	}
	withSearch := *first
	withSearch.WebSearch = true
	if err := tuiRuntimeValidateInteractiveFirstTurn(&withSearch); err != nil {
		t.Fatalf("web search was rejected before PTY start: %v", err)
	}
	withSchema := *first
	withSchema.Format = &Format{Type: "json_schema", Schema: json.RawMessage(`{"type":"object","properties":{"color":{"type":"string"}}}`)}
	if err := tuiRuntimeValidateInteractiveFirstTurn(&withSchema); err != nil {
		t.Fatalf("structured schema was rejected before PTY start: %v", err)
	}
	withoutThinking := *first
	withoutThinking.ThinkOff = true
	if err := tuiRuntimeValidateInteractiveFirstTurn(&withoutThinking); err != nil {
		t.Fatalf("thinking-off was rejected before PTY start: %v", err)
	}
	withSafeguards := *first
	withSafeguards.Safeguards = json.RawMessage(`[{"type":"dangerous_tool_use","classifier_context":"project edits allowed"}]`)
	if err := tuiRuntimeValidateInteractiveFirstTurn(&withSafeguards); err != nil {
		t.Fatalf("normal safeguards were rejected before PTY start: %v", err)
	}
	tooLarge := withSafeguards
	tooLarge.Safeguards = json.RawMessage(`["` + strings.Repeat("x", tuiRuntimeInteractiveSafeguardBodyMax) + `"]`)
	if err := tuiRuntimeValidateInteractiveFirstTurn(&tooLarge); err == nil {
		t.Fatal("oversized safeguards were accepted into PTY process environment")
	}
}

func TestClaudeTUIRuntimeInteractiveTransportToolContinuationShape(t *testing.T) {
	user := Message{Role: "user", Parts: []Part{{Kind: Text, Text: "run tool"}}}
	call := Message{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "c1", Name: "echo_test"}}}
	result := Message{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "c1", Text: "HELLO"}, {Kind: Text, Text: "<system-reminder>new user message</system-reminder>"}}}
	req := Request{Messages: []Message{user, call, result}, Tools: []Tool{{Name: "echo_test"}}}
	if err := tuiRuntimeValidateInteractiveToolContinuation(&req); err != nil {
		t.Fatal(err)
	}
	call2 := Message{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "c2", Name: "echo_test"}}}
	result2 := Message{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "c2", Text: "WORLD"}}}
	chain := req
	chain.Messages = []Message{user, call, result, call2, result2}
	if err := tuiRuntimeValidateInteractiveToolContinuation(&chain); err != nil {
		t.Fatalf("second live tool-result round rejected: %v", err)
	}
	resumed := req
	resumed.Messages = []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "older question"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "older answer"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "now run the tool"}}},
		call,
		result,
	}
	if err := tuiRuntimeValidateInteractiveToolContinuation(&resumed); err != nil {
		t.Fatalf("live tool result after a completed persisted turn rejected: %v", err)
	}
	sameRoleHistory := req
	sameRoleHistory.Messages = []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "older question"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "older answer"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "queued user message"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "now run the tool"}}},
		call,
		result,
	}
	if err := tuiRuntimeValidateInteractiveToolContinuation(&sameRoleHistory); err != nil {
		t.Fatalf("continuation after consecutive user history rejected: %v", err)
	}
	splitAssistant := req
	splitAssistant.Messages = []Message{
		user,
		{Role: "assistant", Parts: []Part{{Kind: Thinking, Text: "checking"}}},
		call,
		result,
	}
	if err := tuiRuntimeValidateInteractiveToolContinuation(&splitAssistant); err != nil {
		t.Fatalf("continuation after split assistant rows rejected: %v", err)
	}
	for _, tc := range []struct {
		name string
		msgs []Message
	}{
		{"unrelated call", []Message{user, call, {Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "c2"}}}}},
		{"stale prior call", []Message{user, call, result, {Role: "assistant", Parts: []Part{{Kind: Text, Text: "done"}}}, {Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "c1"}}}}},
		{"no call", []Message{user, {Role: "assistant", Parts: []Part{{Kind: Text, Text: "old"}}}, result}},
		{"no result", []Message{user, call, {Role: "user", Parts: []Part{{Kind: Text, Text: "hi"}}}}},
		{"text before result", []Message{user, call, {Role: "user", Parts: []Part{{Kind: Text, Text: "do something else"}, {Kind: ToolResult, CallID: "c1"}}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := req
			r.Messages = tc.msgs
			if err := tuiRuntimeValidateInteractiveToolContinuation(&r); err == nil {
				t.Fatal("unsafe continuation was accepted")
			}
		})
	}
}

func TestClaudeTUIRuntimeInteractiveExactIDContinuationCanOmitAssistantWireRow(t *testing.T) {
	body := []byte("{\"model\":\"claude-opus-5-5\",\"max_tokens\":64,\"tools\":[{\"name\":\"echo_test\",\"input_schema\":{\"type\":\"object\"}}],\"messages\":[{\"role\":\"user\",\"content\":\"run tool\"},{\"role\":\"user\",\"content\":[{\"type\":\"tool_result\",\"tool_use_id\":\"c1\",\"content\":\"DONE\"}]}]}")
	req, err := parseAnthropic(body)
	if err != nil {
		t.Fatal(err)
	}
	req.Model = "claude-opus-5-5"
	if err := tuiRuntimeValidateInteractiveRequest(req); err != nil {
		t.Fatalf("pre-match validation rejected compacted live continuation: %v", err)
	}

	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	waiter := make(chan tuiRuntimeMcpToolResult, 1)
	run := &tuiRuntimeSubscriptionRun{
		bridge: b, token: "tok", interactive: true, owner: "owner", model: req.Model,
		pending:               map[string]chan tuiRuntimeMcpToolResult{"c1": waiter},
		tools:                 map[string]bool{"echo_test": true},
		shown:                 []tuiRuntimeShownCall{{id: "c1", name: "echo_test"}},
		interactiveContextKey: tuiRuntimeTurnKey("owner", req, nil),
		parkedAt:              time.Now(),
	}
	run.heard(req.Messages[:1])
	b.runs[run.token] = run
	b.calls["c1"] = run

	found, results, how := b.match(req)
	if found != run || how != tuiRuntimeByExactID || len(results) != 1 {
		t.Fatalf("exact compacted continuation match: found=%t how=%q results=%d", found == run, how, len(results))
	}
	if err := tuiRuntimeValidateInteractiveMatchedContinuation(req, run, results, how); err != nil {
		t.Fatalf("exact live continuation with omitted assistant row rejected: %v", err)
	}
	if !run.follows(req.Messages) {
		t.Fatal("compacted live continuation did not preserve the run's heard history")
	}
	events, err := run.continueWith(results, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := <-waiter
	if len(got.Content) == 0 || got.Content[0]["text"] != "DONE" {
		t.Fatalf("tool result not delivered to live waiter: %#v", got.Content)
	}
	run.endSegment()
	for range events {
	}

	parallel := &tuiRuntimeSubscriptionRun{shown: []tuiRuntimeShownCall{{id: "c1", name: "echo_test"}, {id: "c2", name: "echo_test"}}}
	if err := tuiRuntimeValidateInteractiveExactIDContinuation(req, parallel, results); err == nil {
		t.Fatal("compacted continuation omitted a parallel live tool result")
	}

	unsafeText := *req
	unsafeText.Messages = []Message{
		req.Messages[0],
		{Role: "user", Parts: []Part{{Kind: Text, Text: "new instruction before result"}}},
		req.Messages[1],
	}
	if err := tuiRuntimeValidateInteractiveExactIDContinuation(&unsafeText, run, results); err == nil {
		t.Fatal("new un-heard text before an exact live tool result was accepted")
	}

	extraResults := append(slices.Clone(results), Part{Kind: ToolResult, CallID: "other-call", Text: "OTHER"})
	if err := tuiRuntimeValidateInteractiveExactIDContinuation(req, run, extraResults); err == nil {
		t.Fatal("unowned extra tool result was accepted into exact live continuation")
	}
}

func TestClaudeTUIRuntimeInteractiveTransportMissingWrapperFailsClosed(t *testing.T) {
	tuiRuntimeSetWrapper(t, filepath.Join(t.TempDir(), "no-wrapper"))
	if _, err := tuiRuntimeInteractiveWrapper(); err == nil {
		t.Fatal("unavailable PTY wrapper silently accepted")
	}
	if _, _, err := tuiRuntimeNewSubscriptionBridge().startInteractiveOneTurn(context.Background(),
		&Request{Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "hi"}}}}},
		"claude-haiku-4-5", "", "account", ""); err == nil {
		t.Fatal("unavailable PTY wrapper silently fell back to -p")
	}
}

func TestClaudeTUIRuntimeInteractiveTransportFirstTurnUsesWrapperAndFiniteTextStdin(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	flagsPath := filepath.Join(dir, "args.txt")
	promptPath := filepath.Join(dir, "prompt.txt")
	wrapper := filepath.Join(dir, "fake-bridge")
	// Fake the wrapper, never the real Claude account or user session.
	data := `#!/bin/sh
printf '%s\n' "$@" > "$INTERACTIVE_TEST_ARGS"
cat > "$INTERACTIVE_TEST_PROMPT"
printf '%s\n' '{"type":"assistant","uuid":"row-1","session_id":"fake-inner","message":{"id":"m1","model":"claude-haiku-4-5","usage":{"input_tokens":7,"output_tokens":1,"cache_creation_input_tokens":11},"content":[{"type":"thinking","thinking":"reason"}]}}'
printf '%s\n' '{"type":"assistant","uuid":"row-2","session_id":"fake-inner","message":{"id":"m1","model":"claude-haiku-4-5","usage":{"input_tokens":7,"output_tokens":3,"cache_creation_input_tokens":11},"content":[{"type":"text","text":"PTY_FIRST_OK"}]}}'
printf '%s\n' '{"type":"assistant","uuid":"row-2","session_id":"fake-inner","message":{"id":"m1","model":"claude-haiku-4-5","content":[{"type":"text","text":"DO_NOT_REPEAT"}]}}'
printf '%s\n' '{"type":"result","stop_reason":"end_turn","session_id":"fake-inner"}'
`
	if err := os.WriteFile(wrapper, []byte(data), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_TEST_ARGS", flagsPath)
	t.Setenv("INTERACTIVE_TEST_PROMPT", promptPath)
	req := &Request{Model: "claude-haiku-4-5", System: "custom system", Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "hello from outer client"}}}}}
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	run, events, err := b.startInteractiveOneTurn(context.Background(), req, req.Model, "", "test-owner", "")
	if err != nil {
		t.Fatal(err)
	}
	defer run.abort()
	deadline := time.After(10 * time.Second)
	var text, thought string
	start, stop := 0, 0
	var usage Usage
	for events != nil {
		select {
		case ev, ok := <-events:
			if !ok {
				events = nil
				continue
			}
			switch ev.Kind {
			case KStart:
				start++
				usage.add(ev.Usage)
			case KUsage:
				usage.add(ev.Usage)
			case KText:
				text += ev.Text
			case KThink:
				thought += ev.Text
			case KStop:
				stop++
			case KError:
				t.Fatalf("PTY parse failed: %s", ev.Text)
			}
		case <-deadline:
			t.Fatal("wrapper stream did not end")
		}
	}
	if start != 1 || stop != 1 || text != "PTY_FIRST_OK" || thought != "reason" {
		t.Fatalf("wrong compat events: start=%d stop=%d text=%q thinking=%q", start, stop, text, thought)
	}
	if usage.Input != 7 || usage.Output != 3 || usage.CacheWrite != 11 {
		t.Fatalf("split transcript rows double-counted or dropped usage: %+v", usage)
	}
	args, err := os.ReadFile(flagsPath)
	if err != nil {
		t.Fatal(err)
	}
	flags := strings.Split(strings.TrimSpace(string(args)), "\n")
	if flags[0] != "-p" || strings.Contains(string(args), "--input-format") {
		t.Fatalf("wrong wrapper flags: %q", flags)
	}
	prompt, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(prompt), "hello from outer client") || !strings.Contains(string(prompt), "custom system") || strings.Contains(string(prompt), `"type":"user"`) {
		t.Fatalf("wrapper did not receive finite text: %q", prompt)
	}
	// No interactive checkpoint is written until the continuity stage.
	if _, err := os.Stat(tuiRuntimeInteractiveSessionStatePath()); !os.IsNotExist(err) {
		t.Fatalf("unfinished one-turn transport wrote a persisted checkpoint: %v", err)
	}
}

func TestClaudeTUIRuntimeInteractiveTransportSchemaUsesTerminalStructuredOutput(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	argsPath := filepath.Join(dir, "schema-args.txt")
	wrapper := filepath.Join(dir, "schema-bridge")
	// the schema comes in a file (one argument holds 128 KiB at most on
	// Linux): the stand-in keeps what the file held
	script := `#!/bin/sh
printf '%s\n' "$@" > "$INTERACTIVE_SCHEMA_ARGS"
while [ $# -gt 0 ]; do [ "$1" = --json-schema ] && cat "$2" > "$INTERACTIVE_SCHEMA_ARGS.schema"; shift; done
cat >/dev/null
printf '%s\n' '{"type":"assistant","uuid":"schema-row-1","session_id":"inner-schema","message":{"id":"m-schema","model":"claude-haiku-4-5","content":[{"type":"text","text":"THIS_MUST_NOT_ESCAPE"}]}}'
printf '%s\n' '{"type":"assistant","uuid":"schema-row-2","session_id":"inner-schema","message":{"id":"m-schema","model":"claude-haiku-4-5","content":[{"type":"tool_use","id":"structured-1","name":"StructuredOutput","input":{"color":"blue"}}]}}'
printf '%s\n' '{"type":"result","subtype":"success","is_error":false,"session_id":"inner-schema","structured_output":{"color":"blue"}}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_SCHEMA_ARGS", argsPath)
	schema := json.RawMessage(`{"type":"object","properties":{"color":{"type":"string"}},"required":["color"]}`)
	req := &Request{Model: "claude-haiku-4-5", Format: &Format{Type: "json_schema", Schema: schema}, Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "sky?"}}}}}
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	run, events, err := b.startInteractiveOneTurn(context.Background(), req, req.Model, "", "owner", "outer-schema")
	if err != nil {
		t.Fatal(err)
	}
	defer run.abort()
	var text, stop string
	for ev := range events {
		switch ev.Kind {
		case KText:
			text += ev.Text
		case KStop:
			stop = ev.Stop
		case KError:
			t.Fatalf("schema PTY failed: %s", ev.Text)
		}
	}
	if text != `{"color":"blue"}` || stop != "stop" || strings.Contains(text, "THIS_MUST_NOT_ESCAPE") {
		t.Fatalf("schema result leaked or changed: text=%q stop=%q", text, stop)
	}
	args, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(args)), "\n")
	i := -1
	for n, arg := range lines {
		if arg == "--json-schema" {
			i = n
			break
		}
	}
	if i < 0 || i+1 >= len(lines) || !filepath.IsAbs(lines[i+1]) {
		t.Fatalf("wrapper schema args = %q", lines)
	}
	if got, err := os.ReadFile(argsPath + ".schema"); err != nil || string(got) != string(schema) {
		t.Fatalf("wrapper schema file = %q, %v", got, err)
	}
}

func TestClaudeTUIRuntimeInteractiveTransportThinkingOffUsesSingleClaudeArg(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	argsPath := filepath.Join(dir, "thinking-args.txt")
	promptPath := filepath.Join(dir, "thinking-prompt.txt")
	wrapper := filepath.Join(dir, "thinking-bridge")
	script := `#!/bin/sh
printf '%s\n' "$@" > "$INTERACTIVE_THINK_ARGS"
cat > "$INTERACTIVE_THINK_PROMPT"
printf '%s\n' '{"type":"assistant","uuid":"think-row","session_id":"inner-think","message":{"id":"m-think","model":"claude-haiku-4-5","content":[{"type":"text","text":"THINK_OFF_OK"}]}}'
printf '%s\n' '{"type":"result","session_id":"inner-think","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_THINK_ARGS", argsPath)
	t.Setenv("INTERACTIVE_THINK_PROMPT", promptPath)
	req := &Request{Model: "claude-haiku-4-5", ThinkOff: true, Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}}}}
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	run, events, err := b.startInteractiveOneTurn(context.Background(), req, req.Model, "", "owner", "outer-think")
	if err != nil {
		t.Fatal(err)
	}
	defer run.abort()
	var text string
	for ev := range events {
		switch ev.Kind {
		case KText:
			text += ev.Text
		case KError:
			t.Fatalf("thinking-off PTY failed: %s", ev.Text)
		}
	}
	if text != "THINK_OFF_OK" {
		t.Fatalf("thinking-off reply = %q", text)
	}
	args, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(args)), "\n")
	found := false
	for _, arg := range lines {
		if arg == "--thinking=disabled" {
			found = true
		}
		if arg == "disabled" {
			t.Fatalf("thinking value became a standalone argv: %q", lines)
		}
	}
	if !found {
		t.Fatalf("wrapper args lack thinking-off: %q", lines)
	}
	prompt, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(prompt), "disabled") || !strings.Contains(string(prompt), "hello") {
		t.Fatalf("thinking option leaked into prompt: %q", prompt)
	}
}

func TestClaudeTUIRuntimeInteractiveTransportSafeguardsReachPTYEnvironment(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	bodyPath := filepath.Join(dir, "extra-body.txt")
	betaPath := filepath.Join(dir, "betas.txt")
	wrapper := filepath.Join(dir, "safeguard-bridge")
	script := `#!/bin/sh
printf '%s' "$CLAUDE_CODE_EXTRA_BODY" > "$INTERACTIVE_SAFEGUARD_BODY"
printf '%s' "$ANTHROPIC_BETAS" > "$INTERACTIVE_SAFEGUARD_BETA"
cat >/dev/null
printf '%s\n' '{"type":"assistant","uuid":"sg-row","session_id":"inner-sg","message":{"id":"m-sg","model":"claude-haiku-4-5","content":[{"type":"text","text":"SAFEGUARD_OK"}]}}'
printf '%s\n' '{"type":"result","session_id":"inner-sg","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_SAFEGUARD_BODY", bodyPath)
	t.Setenv("INTERACTIVE_SAFEGUARD_BETA", betaPath)
	sg := json.RawMessage(`[{"type":"dangerous_tool_use","classifier_context":"project edits allowed"}]`)
	beta := "dangerous-tool-use-2026-09-03"
	req := &Request{Model: "claude-haiku-4-5", Safeguards: sg, SafeguardBeta: beta, Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}}}}
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	run, events, err := b.startInteractiveOneTurn(context.Background(), req, req.Model, "", "owner", "outer-sg")
	if err != nil {
		t.Fatal(err)
	}
	defer run.abort()
	var text string
	for ev := range events {
		if ev.Kind == KText {
			text += ev.Text
		}
		if ev.Kind == KError {
			t.Fatalf("safeguard PTY failed: %s", ev.Text)
		}
	}
	if text != "SAFEGUARD_OK" {
		t.Fatalf("safeguard reply = %q", text)
	}
	body, err := os.ReadFile(bodyPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != tuiRuntimeSafeguardBody(sg) {
		t.Fatalf("extra body = %q, want %q", body, tuiRuntimeSafeguardBody(sg))
	}
	betas, err := os.ReadFile(betaPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(betas) != beta {
		t.Fatalf("betas = %q, want %q", betas, beta)
	}
}

func TestClaudeTUIRuntimeInteractiveTransportRejectsSafeguardChangeDuringMCP(t *testing.T) {
	tuiRuntimeSetWrapper(t, "/unused-for-live-run")
	b := tuiRuntimeNewSubscriptionBridge()
	s := tuiRuntimeNewTestServer()
	s.subscription = b
	defer tuiRuntimeBridges.Delete(b)
	beta := "dangerous-tool-use-2026-09-03"
	body := []byte(`{"model":"claude-haiku-4-5","max_tokens":64,"safeguards":[{"context":"B"}],"tools":[{"name":"echo_test","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"run tool"},{"role":"assistant","content":[{"type":"tool_use","id":"c1","name":"echo_test","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"c1","content":"done"}]}]}`)
	parsed, err := parseAnthropic(body)
	if err != nil {
		t.Fatal(err)
	}
	parsed.Model = "claude-haiku-4-5"
	waiter := make(chan tuiRuntimeMcpToolResult, 1)
	run := &tuiRuntimeSubscriptionRun{
		bridge: b, token: "tok", interactive: true, owner: "owner", model: parsed.Model,
		pending: map[string]chan tuiRuntimeMcpToolResult{"c1": waiter}, tools: map[string]bool{"echo_test": true},
		shown:      []tuiRuntimeShownCall{{id: "c1", name: "echo_test"}},
		safeguards: json.RawMessage(`[{"context":"A"}]`), safeguardBeta: beta,
		interactiveContextKey: tuiRuntimeTurnKey("owner", parsed, nil),
	}
	b.runs[run.token] = run
	b.calls["c1"] = run
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body)))
	req.Header.Set("anthropic-beta", beta)
	rec := httptest.NewRecorder()
	started := false
	start := func(context.Context, *Request) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
		started = true
		return nil, nil, errors.New("must not start a replacement run")
	}
	code, msg := s.serveSubscription(rec, req, protoAnthropic, "Claude Code", parsed.Model, "owner", body, &Usage{}, start)
	if code != http.StatusConflict || !strings.Contains(msg, "safeguards changed") || started || !run.closed {
		t.Fatalf("changed safeguards were not failed closed: code=%d msg=%q started=%t closed=%t", code, msg, started, run.closed)
	}
}

func TestClaudeTUIRuntimeInteractiveTransportOuterSessionUsesStableIsolatedWorkDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "fake-bridge")
	data := `#!/bin/sh
cat >/dev/null
printf '%s\n' '{"type":"assistant","uuid":"row","session_id":"inner","message":{"id":"m","model":"claude-haiku-4-5","content":[{"type":"text","text":"OK"}]}}'
printf '%s\n' '{"type":"result","stop_reason":"end_turn","session_id":"inner"}'
`
	if err := os.WriteFile(wrapper, []byte(data), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	req := &Request{Model: "claude-haiku-4-5", Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "hi"}}}}}
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	start := func(owner, outer string) (string, *tuiRuntimeSubscriptionRun) {
		t.Helper()
		run, events, err := b.startInteractiveOneTurn(context.Background(), req, req.Model, "", owner, outer)
		if err != nil {
			t.Fatal(err)
		}
		for range events {
		}
		select {
		case <-run.interactiveDone:
		case <-time.After(2 * time.Second):
			t.Fatal("wrapper did not exit")
		}
		return run.cmd.Dir, run
	}
	a1, runA1 := start("owner-A", "desktop-session-A")
	a2, _ := start("owner-A", "desktop-session-A")
	b1, _ := start("owner-A", "desktop-session-B")
	c1, _ := start("owner-B", "desktop-session-A")
	if a1 != a2 {
		t.Fatalf("same outer session changed cwd: %q != %q", a1, a2)
	}
	if a1 == b1 || a1 == c1 || b1 == c1 {
		t.Fatalf("outer session/owner cwd collision: A=%q B=%q C=%q", a1, b1, c1)
	}
	if strings.Contains(a1, "desktop-session-A") || strings.Contains(a1, "owner-A") {
		t.Fatalf("stable cwd leaks outer identity: %q", a1)
	}
	if runA1.outerSession != "desktop-session-A" {
		t.Fatalf("run lost outer session: %q", runA1.outerSession)
	}
	for _, p := range []string{a1, b1, c1} {
		t.Cleanup(func() { _ = os.RemoveAll(p) })
	}
}

func TestClaudeTUIRuntimeInteractiveOuterThreadClaimIsolatesHelpers(t *testing.T) {
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	mainReq := &Request{Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "main conversation"}}}}}
	helperReq := &Request{Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "Perform a web search for example"}}}}}
	if got := b.claimInteractiveOuterSession("owner", "outer", mainReq); got != "outer" {
		t.Fatalf("main did not claim outer session: %q", got)
	}
	if got := b.claimInteractiveOuterSession("owner", "outer", mainReq); got != "outer" {
		t.Fatalf("same logical thread lost outer session: %q", got)
	}
	if got := b.claimInteractiveOuterSession("owner", "outer", helperReq); got != "" {
		t.Fatalf("helper inherited main outer session: %q", got)
	}
	if got := b.claimInteractiveOuterSession("other-owner", "outer", helperReq); got != "outer" {
		t.Fatalf("another owner was incorrectly tied to main: %q", got)
	}
	if got := b.claimInteractiveOuterSession("owner", "other-outer", helperReq); got != "other-outer" {
		t.Fatalf("another outer session was incorrectly tied to main: %q", got)
	}
}

func TestClaudeTUIRuntimeInteractiveOuterThreadClaimRaceHasOneWinner(t *testing.T) {
	for n := 0; n < 40; n++ {
		b := tuiRuntimeNewSubscriptionBridge()
		a := &Request{Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "thread A"}}}}}
		c := &Request{Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "thread B"}}}}}
		start := make(chan struct{})
		got := make(chan string, 2)
		for _, req := range []*Request{a, c} {
			go func(req *Request) {
				<-start
				got <- b.claimInteractiveOuterSession("owner", "outer", req)
			}(req)
		}
		close(start)
		x, y := <-got, <-got
		if (x == "outer") == (y == "outer") {
			t.Fatalf("race produced %q and %q; exactly one thread must own outer", x, y)
		}
		tuiRuntimeBridges.Delete(b)
	}
}

func TestClaudeTUIRuntimeClaudeDesktopSuggestionIsRecognizedBeforeOuterClaim(t *testing.T) {
	ordinary := &Request{Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}}}}
	if tuiRuntimeClaudeDesktopSuggestion(ordinary) {
		t.Fatal("ordinary request classified as suggestion")
	}
	suggestion := &Request{Messages: []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "answer"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "  [SUGGESTION MODE: suggest next input]\nshort"}}},
	}}
	if !tuiRuntimeClaudeDesktopSuggestion(suggestion) {
		t.Fatal("Desktop suggestion marker not recognized")
	}
	// Its marker must never be allowed to claim the canonical outer slot.
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	if got := b.claimInteractiveOuterSession("owner", "outer", ordinary); got != "outer" {
		t.Fatal("main did not claim outer")
	}
	if got := b.outerThread("owner", "outer"); got != firstWords(ordinary) {
		t.Fatalf("wrong canonical thread: %q", got)
	}
	side := tuiRuntimeClaudeDesktopSuggestionRequest(suggestion)
	if side.System != "" || len(side.Tools) != 0 || side.ToolChoice != "" || side.Effort != "low" || len(side.Messages) == 0 || len(side.Messages) > 3 {
		t.Fatalf("suggestion was not reduced to an isolated low-effort request: %#v", side)
	}
}

func TestClaudeTUIRuntimeClaudeDesktopSuggestionDoesNotAdvanceInteractiveCheckpoint(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "suggestion-bridge")
	logPath := filepath.Join(dir, "suggestion.log")
	script := `#!/bin/sh
prompt=$(cat)
printf '%s\n' '---' >> "$INTERACTIVE_SUGGEST_LOG"
printf 'args:%s\nprompt:%s\n' "$*" "$prompt" >> "$INTERACTIVE_SUGGEST_LOG"
case "$prompt" in
  *"[SUGGESTION MODE:"*) sid=sess-suggest; text=SUGGEST ;;
  *)
    sid=sess-main
    case " $* " in
      *" --resume sess-main "*) text=MAIN2 ;;
      *) text=MAIN1 ;;
    esac
    ;;
esac
printf '{"type":"assistant","uuid":"row-%s","session_id":"%s","message":{"id":"m-%s","model":"claude-opus-5-5","content":[{"type":"text","text":"%s"}]}}\n' "$text" "$sid" "$text" "$text"
printf '{"type":"result","session_id":"%s","stop_reason":"end_turn"}\n' "$sid"
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_SUGGEST_LOG", logPath)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer, model := "owner", "desktop-suggestion", "claude-opus-5-5"
	collect := func(run *tuiRuntimeSubscriptionRun, events <-chan Event) (string, string) {
		t.Helper()
		var text, stop string
		for ev := range events {
			switch ev.Kind {
			case KText:
				text += ev.Text
			case KStop:
				stop = ev.Stop
			case KError:
				t.Fatal(ev.Text)
			}
		}
		return text, stop
	}
	firstReq := &Request{Model: model, Tools: []Tool{{Name: "read"}}, Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}}}}
	main1, events, err := b.startInteractiveRequest(context.Background(), firstReq, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	text1, stop1 := collect(main1, events)
	main1.ended(firstReq, text1, stop1, true)
	before, ok := b.interactiveSession(owner, outer)
	if !ok || before.SessionID != "sess-main" {
		t.Fatalf("main checkpoint missing before suggestion: %#v, %v", before, ok)
	}

	suggestionText := "[SUGGESTION MODE: Suggest what the user might naturally type next into Claude Code.]\nReturn one short suggestion."
	suggestionReq := &Request{Model: model, System: "MAIN_SYSTEM_MUST_NOT_LEAK", Tools: []Tool{{Name: "read"}}, Messages: []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: text1}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: suggestionText}}},
	}}
	side, sideEvents, err := b.startInteractiveRequest(context.Background(), suggestionReq, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	sideText, sideStop := collect(side, sideEvents)
	if sideText != "SUGGEST" || sideStop != "stop" {
		t.Fatalf("suggestion = %q %q", sideText, sideStop)
	}
	side.ended(suggestionReq, sideText, sideStop, true)
	after, ok := b.interactiveSession(owner, outer)
	if !ok || after.SessionID != before.SessionID || after.ConvKey != before.ConvKey || after.ReplyKey != before.ReplyKey {
		t.Fatalf("suggestion advanced main checkpoint:\nbefore=%#v\nafter=%#v", before, after)
	}

	nextReq := &Request{Model: model, Tools: firstReq.Tools, Messages: []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: text1}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "real next turn"}}},
	}}
	main2, events2, err := b.startInteractiveRequest(context.Background(), nextReq, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	text2, stop2 := collect(main2, events2)
	if text2 != "MAIN2" || stop2 != "stop" {
		t.Fatalf("main after suggestion = %q %q", text2, stop2)
	}
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	log := string(logBytes)
	var suggestionIsolated, mainResumed bool
	for _, block := range strings.Split(log, "---\n") {
		switch {
		case strings.Contains(block, "SUGGESTION MODE:"):
			suggestionIsolated = strings.Contains(block, "--effort low") && !strings.Contains(block, "--resume sess-main") && !strings.Contains(block, "MAIN_SYSTEM_MUST_NOT_LEAK")
		case strings.Contains(block, "real next turn"):
			mainResumed = strings.Contains(block, "--resume sess-main")
		}
	}
	if !suggestionIsolated || !mainResumed {
		t.Fatalf("suggestion isolation/main resume failed:\n%s", log)
	}
}

func TestClaudeTUIRuntimeClaudeDesktopAuxiliaryModelDoesNotAdvanceInteractiveCheckpoint(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "aux-model-bridge")
	logPath := filepath.Join(dir, "aux-model.log")
	script := `#!/bin/sh
prompt=$(cat)
printf '%s\n' '---' >> "$INTERACTIVE_AUX_MODEL_LOG"
printf 'args:%s\nprompt:%s\n' "$*" "$prompt" >> "$INTERACTIVE_AUX_MODEL_LOG"
case " $* " in
  *" --model claude-haiku-4-5 "*) sid=sess-side; text=SIDE ;;
  *" --resume sess-main "*) sid=sess-main; text=MAIN2 ;;
  *) sid=sess-main; text=MAIN1 ;;
esac
printf '{"type":"assistant","uuid":"row-%s","session_id":"%s","message":{"id":"m-%s","model":"claude-opus-5-5","content":[{"type":"text","text":"%s"}]}}\n' "$text" "$sid" "$text" "$text"
printf '{"type":"result","session_id":"%s","stop_reason":"end_turn"}\n' "$sid"
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_AUX_MODEL_LOG", logPath)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer := "owner", "desktop-aux-model"
	collect := func(run *tuiRuntimeSubscriptionRun, events <-chan Event) (string, string) {
		t.Helper()
		var text, stop string
		for ev := range events {
			switch ev.Kind {
			case KText:
				text += ev.Text
			case KStop:
				stop = ev.Stop
			case KError:
				t.Fatal(ev.Text)
			}
		}
		return text, stop
	}
	main1Req := &Request{Model: "claude-opus-5-5", Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "main thread"}}}}}
	main1, events1, err := b.startInteractiveRequest(context.Background(), main1Req, main1Req.Model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	text1, stop1 := collect(main1, events1)
	main1.ended(main1Req, text1, stop1, true)
	before, ok := b.interactiveSession(owner, outer)
	if !ok || before.SessionID != "sess-main" {
		t.Fatalf("main checkpoint missing: %#v, %v", before, ok)
	}

	helperReq := &Request{Model: "claude-haiku-4-5", Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "summarize fetched web page"}}}}}
	helper, helperEvents, err := b.startInteractiveRequest(context.Background(), helperReq, helperReq.Model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	helperText, helperStop := collect(helper, helperEvents)
	if helperText != "SIDE" || helperStop != "stop" {
		t.Fatalf("helper = %q %q", helperText, helperStop)
	}
	helper.ended(helperReq, helperText, helperStop, true)
	after, ok := b.interactiveSession(owner, outer)
	if !ok || after.SessionID != before.SessionID || after.ConvKey != before.ConvKey || after.ReplyKey != before.ReplyKey {
		t.Fatalf("auxiliary model advanced main checkpoint:\nbefore=%#v\nafter=%#v", before, after)
	}

	main2Req := &Request{Model: "claude-opus-5-5", Messages: []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "main thread"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: text1}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "main next"}}},
	}}
	main2, events2, err := b.startInteractiveRequest(context.Background(), main2Req, main2Req.Model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	text2, stop2 := collect(main2, events2)
	if text2 != "MAIN2" || stop2 != "stop" {
		t.Fatalf("main after helper = %q %q", text2, stop2)
	}
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	log := string(logBytes)
	var sideIsolated, mainResumed bool
	for _, block := range strings.Split(log, "---\n") {
		switch {
		case strings.Contains(block, "summarize fetched web page"):
			sideIsolated = strings.Contains(block, "--model claude-haiku-4-5") && !strings.Contains(block, "--resume sess-main")
		case strings.Contains(block, "main next"):
			mainResumed = strings.Contains(block, "--resume sess-main")
		}
	}
	if !sideIsolated || !mainResumed {
		t.Fatalf("auxiliary model isolation/main resume failed:\n%s", log)
	}
}

func TestClaudeTUIRuntimeClaudeDesktopSameModelSideThreadDoesNotAdvanceInteractiveCheckpoint(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "side-thread-bridge")
	logPath := filepath.Join(dir, "side-thread.log")
	script := `#!/bin/sh
prompt=$(cat)
printf '%s\n' '---' >> "$INTERACTIVE_SIDE_THREAD_LOG"
printf 'args:%s\nprompt:%s\n' "$*" "$prompt" >> "$INTERACTIVE_SIDE_THREAD_LOG"
case "$prompt" in
  *"Perform a web search"*) sid=sess-side; text=SIDE ;;
  *"main next"*) sid=sess-main; text=MAIN2 ;;
  *) sid=sess-main; text=MAIN1 ;;
esac
printf '{"type":"assistant","uuid":"row-%s","session_id":"%s","message":{"id":"m-%s","model":"claude-opus-5-5","content":[{"type":"text","text":"%s"}]}}\n' "$text" "$sid" "$text" "$text"
printf '{"type":"result","session_id":"%s","stop_reason":"end_turn"}\n' "$sid"
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_SIDE_THREAD_LOG", logPath)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer, model := "owner", "desktop-side-thread", "claude-opus-5-5"
	collect := func(run *tuiRuntimeSubscriptionRun, events <-chan Event) (string, string) {
		t.Helper()
		var text, stop string
		for ev := range events {
			switch ev.Kind {
			case KText:
				text += ev.Text
			case KStop:
				stop = ev.Stop
			case KError:
				t.Fatal(ev.Text)
			}
		}
		return text, stop
	}
	mainReq := &Request{Model: model, Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "main thread"}}}}}
	main, mainEvents, err := b.startInteractiveRequest(context.Background(), mainReq, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	mainText, mainStop := collect(main, mainEvents)
	main.ended(mainReq, mainText, mainStop, true)
	before, ok := b.interactiveSession(owner, outer)
	if !ok || before.SessionID != "sess-main" {
		t.Fatalf("main checkpoint missing: %#v, %v", before, ok)
	}

	sideReq := &Request{Model: model, Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "Perform a web search for example"}}}}}
	side, sideEvents, err := b.startInteractiveRequest(context.Background(), sideReq, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	sideText, sideStop := collect(side, sideEvents)
	if sideText != "SIDE" || sideStop != "stop" {
		t.Fatalf("side thread = %q %q", sideText, sideStop)
	}
	side.ended(sideReq, sideText, sideStop, true)
	after, ok := b.interactiveSession(owner, outer)
	if !ok || after.SessionID != before.SessionID || after.ConvKey != before.ConvKey || after.ReplyKey != before.ReplyKey {
		t.Fatalf("same-model side thread advanced main checkpoint:\nbefore=%#v\nafter=%#v", before, after)
	}

	nextReq := &Request{Model: model, Messages: []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "main thread"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: mainText}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "main next"}}},
	}}
	next, nextEvents, err := b.startInteractiveRequest(context.Background(), nextReq, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	nextText, nextStop := collect(next, nextEvents)
	if nextText != "MAIN2" || nextStop != "stop" {
		t.Fatalf("main after side thread = %q %q", nextText, nextStop)
	}
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	log := string(logBytes)
	var sideIsolated, mainResumed bool
	for _, block := range strings.Split(log, "---\n") {
		switch {
		case strings.Contains(block, "Perform a web search"):
			sideIsolated = !strings.Contains(block, "--resume sess-main")
		case strings.Contains(block, "main next"):
			mainResumed = strings.Contains(block, "--resume sess-main")
		}
	}
	if !sideIsolated || !mainResumed {
		t.Fatalf("same-model side-thread isolation/main resume failed:\n%s", log)
	}
}

func TestClaudeTUIRuntimeClaudeDesktopHistoricalSideThreadIsEphemeralAndStripsToolHistory(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "historical-side-bridge")
	logPath := filepath.Join(dir, "historical-side.log")
	script := `#!/bin/sh
prompt=$(cat)
printf 'args:%s\nprompt:%s\n' "$*" "$prompt" > "$INTERACTIVE_HIST_SIDE_LOG"
printf '%s\n' '{"type":"assistant","uuid":"row-side-hist","session_id":"sess-side-hist","message":{"id":"m-side-hist","model":"claude-opus-5-5","content":[{"type":"text","text":"SIDE_HIST_OK"}]}}'
printf '%s\n' '{"type":"result","session_id":"sess-side-hist","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_HIST_SIDE_LOG", logPath)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer, model := "owner", "desktop-historical-side", "claude-opus-5-5"
	mainReq := &Request{Model: model, Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "main conversation"}}}}}
	if got := b.seedOuterThread(owner, outer, firstWords(mainReq)); got != firstWords(mainReq) {
		t.Fatalf("failed to seed main outer thread: %q", got)
	}
	mainCheckpoint := tuiRuntimeInteractiveSessionEntry{
		SessionID: "sess-main", Model: model, ThreadKey: firstWords(mainReq),
		ConvKey: "main-conv", ReplyKey: "main-reply", ContextKey: "main-context",
	}
	b.saveInteractiveSession(owner, outer, mainCheckpoint)

	sideReq := &Request{Model: model, Messages: []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "HELPER_OLD_USER"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "HELPER_VISIBLE"}, {Kind: ToolCall, ID: "old-side-call", Name: "dangerous_old_tool"}}},
		{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "old-side-call", Text: "STALE_SIDE_RESULT"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "HELPER_AFTER_TOOL"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "CURRENT_HELPER_REQUEST"}}},
	}}
	run, events, err := b.startInteractiveRequest(context.Background(), sideReq, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	var text, stop string
	for ev := range events {
		switch ev.Kind {
		case KText:
			text += ev.Text
		case KStop:
			stop = ev.Stop
		case KError:
			t.Fatal(ev.Text)
		}
	}
	if text != "SIDE_HIST_OK" || stop != "stop" || run.outerSession != "" {
		t.Fatalf("historical helper was not isolated: text=%q stop=%q outer=%q", text, stop, run.outerSession)
	}
	run.ended(sideReq, text, stop, true)
	after, ok := b.interactiveSession(owner, outer)
	if !ok || after.SessionID != mainCheckpoint.SessionID || after.ConvKey != mainCheckpoint.ConvKey ||
		after.ReplyKey != mainCheckpoint.ReplyKey || after.ContextKey != mainCheckpoint.ContextKey {
		t.Fatalf("historical helper changed main checkpoint: %#v", after)
	}
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	log := string(logBytes)
	for _, want := range []string{"<historical_conversation_import>", "HELPER_OLD_USER", "HELPER_VISIBLE", "HELPER_AFTER_TOOL", "CURRENT_HELPER_REQUEST"} {
		if !strings.Contains(log, want) {
			t.Fatalf("historical helper missing %q:\n%s", want, log)
		}
	}
	for _, forbidden := range []string{"--resume sess-main", "dangerous_old_tool", "STALE_SIDE_RESULT"} {
		if strings.Contains(log, forbidden) {
			t.Fatalf("historical helper leaked/replayed %q:\n%s", forbidden, log)
		}
	}
}

func TestClaudeTUIRuntimeInteractiveTransportHistoryBootstrapsButLiveToolWithoutRunFailsClosed(t *testing.T) {
	tuiRuntimeSetWrapper(t, "/not-a-wrapper")
	s := tuiRuntimeNewTestServer()
	called := false
	start := func(ctx context.Context, req *Request) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
		called = true
		return s.subscription.startInteractiveRequest(ctx, req, "claude-sonnet-5", "", "test-owner", "outer")
	}
	body := `{"model":"claude-sonnet-5","max_tokens":64,"messages":[{"role":"user","content":"previous"},{"role":"assistant","content":"done"},{"role":"user","content":"continue"}]}`
	r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
	w := httptest.NewRecorder()
	code, msg := s.serveSubscription(w, r, protoAnthropic, "Claude Code", "claude-sonnet-5", "test-owner", []byte(body), &Usage{}, start)
	if code != http.StatusBadGateway || !called || !strings.Contains(msg, "wrapper unavailable") {
		t.Fatalf("safe historical bootstrap did not reach PTY entry: %d %s, started=%t", code, msg, called)
	}
	// A current tool_result is not historical text. Without the live PTY run
	// that made the call, it must be rejected before start() can bootstrap it.
	liveTool := `{"model":"claude-sonnet-5","max_tokens":64,"tools":[{"name":"echo_test","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"run tool"},{"role":"assistant","content":[{"type":"tool_use","id":"c1","name":"echo_test","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"c1","content":"HELLO"}]}]}`
	called = false
	r = httptest.NewRequest("POST", "/v1/messages", strings.NewReader(liveTool))
	w = httptest.NewRecorder()
	code, msg = s.serveSubscription(w, r, protoAnthropic, "Claude Code", "claude-sonnet-5", "test-owner", []byte(liveTool), &Usage{}, start)
	if code != http.StatusConflict || called || !strings.Contains(msg, "no matching live MCP run") {
		t.Fatalf("orphan tool continuation reached bootstrap: %d %s, started=%t", code, msg, called)
	}
	// Once that orphaned result is historical and the user has explicitly
	// started a newer plain-text turn, it is no longer a live continuation.
	// The old tool traffic is imported as read-only history (and stripped by
	// the bootstrap renderer) instead of trapping the conversation in 409.
	staleThenNew := `{"model":"claude-sonnet-5","max_tokens":64,"tools":[{"name":"echo_test","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"run tool"},{"role":"assistant","content":[{"type":"tool_use","id":"c1","name":"echo_test","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"c1","content":"HELLO"}]},{"role":"user","content":"continue"}]}`
	called = false
	r = httptest.NewRequest("POST", "/v1/messages", strings.NewReader(staleThenNew))
	w = httptest.NewRecorder()
	code, msg = s.serveSubscription(w, r, protoAnthropic, "Claude Code", "claude-sonnet-5", "test-owner", []byte(staleThenNew), &Usage{}, start)
	if code != http.StatusBadGateway || !called || !strings.Contains(msg, "wrapper unavailable") {
		t.Fatalf("stale tool history plus a newer user turn did not reach safe bootstrap: %d %s, started=%t", code, msg, called)
	}
	// Claude Desktop can compact the assistant call out of the wire request,
	// then append its own interruption marker after the user's retry. That
	// marker is not the actionable turn; the explicit user text before it is.
	compactedStaleThenNew := `{"model":"claude-sonnet-5","max_tokens":64,"messages":[{"role":"user","content":"compacted summary"},{"role":"user","content":[{"type":"tool_result","tool_use_id":"old-call","content":"OLD"}]},{"role":"user","content":"continue"},{"role":"user","content":"[Request interrupted by user]"}]}`
	called = false
	r = httptest.NewRequest("POST", "/v1/messages", strings.NewReader(compactedStaleThenNew))
	w = httptest.NewRecorder()
	code, msg = s.serveSubscription(w, r, protoAnthropic, "Claude Code", "claude-sonnet-5", "test-owner", []byte(compactedStaleThenNew), &Usage{}, start)
	if code != http.StatusBadGateway || !called || !strings.Contains(msg, "wrapper unavailable") {
		t.Fatalf("compacted stale tool recovery did not reach safe bootstrap: %d %s, started=%t", code, msg, called)
	}
	// A genuinely fresh plain-text request reaches the opt-in entry too; the
	// unavailable fake wrapper then fails closed instead of falling back to -p.
	first := `{"model":"claude-sonnet-5","max_tokens":64,"messages":[{"role":"user","content":"hello"}]}`
	called = false
	r = httptest.NewRequest("POST", "/v1/messages", strings.NewReader(first))
	w = httptest.NewRecorder()
	code, _ = s.serveSubscription(w, r, protoAnthropic, "Claude Code", "claude-sonnet-5", "test-owner", []byte(first), &Usage{}, start)
	if !called || code != http.StatusBadGateway {
		t.Fatal("fresh text was rejected before the interactive start callback")
	}
	// Disabling the opt-in gate must leave the default headless call path
	// unchanged (the supplied start callback must be invoked).
	tuiRuntimeSetWrapper(t, "")
	called = false
	r = httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
	w = httptest.NewRecorder()
	_, _ = s.serveSubscription(w, r, protoAnthropic, "Claude Code", "claude-sonnet-5", "test-owner", []byte(body), &Usage{}, start)
	if !called {
		t.Fatal("disabling opt-in still intercepted the normal headless route")
	}
}

func TestClaudeTUIRuntimeInteractiveDirtyAbandonedToolTurnBootstrapsFreshSession(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "abandoned-tool-bridge")
	logPath := filepath.Join(dir, "abandoned-tool.log")
	script := `#!/bin/sh
prompt=$(cat)
printf 'args:%s\nprompt:%s\n' "$*" "$prompt" > "$INTERACTIVE_ABANDONED_TOOL_LOG"
printf '%s\n' '{"type":"assistant","uuid":"row","session_id":"inner-recovered","message":{"id":"m","model":"claude-opus-5-5","content":[{"type":"text","text":"RECOVERED_OK"}]}}'
printf '%s\n' '{"type":"result","session_id":"inner-recovered","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_ABANDONED_TOOL_LOG", logPath)

	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer, model := "owner", "outer-abandoned", "claude-opus-5-5"
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "original question"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "SAVED_REPLY"}}},
	}
	baseReq := &Request{Model: model, Messages: base}
	keys := tuiRuntimeConversationKeys(owner, base)
	b.saveInteractiveSession(owner, outer, tuiRuntimeInteractiveSessionEntry{
		SessionID: "inner-dirty", Model: model, ThreadKey: firstWords(baseReq),
		ConvKey: keys[0], ReplyKey: tuiRuntimeAssistantReplyKey(base[1]), ContextKey: tuiRuntimeTurnKey(owner, baseReq, nil), Dirty: true,
	})

	req := &Request{Model: model, Messages: []Message{
		base[0], base[1],
		{Role: "user", Parts: []Part{{Kind: Text, Text: "run the old tool"}}},
		{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "stale-call", Name: "dangerous_old_tool"}}},
		{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "stale-call", Text: "STALE_RESULT"}}},
		{Role: "system", Parts: []Part{{Kind: Text, Text: "OLD_ENVIRONMENT"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "API Error: 409 US VPS Claude: Claude interactive migration: old validator rejected the result"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "continue now"}}},
		{Role: "system", Parts: []Part{{Kind: Text, Text: "CURRENT_RECOVERY_ENVIRONMENT"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "[Request interrupted by user]"}}},
	}}
	run, events, err := b.startInteractiveRequest(context.Background(), req, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	var text, stop string
	for ev := range events {
		switch ev.Kind {
		case KText:
			text += ev.Text
		case KStop:
			stop = ev.Stop
		case KError:
			t.Fatal(ev.Text)
		}
	}
	if text != "RECOVERED_OK" || stop != "stop" {
		t.Fatalf("abandoned-tool recovery = %q %q", text, stop)
	}
	run.ended(req, text, stop, true)
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	log := string(logBytes)
	for _, want := range []string{"<historical_conversation_import>", "original question", "SAVED_REPLY", "<current_user_request>", "continue now", "CURRENT_RECOVERY_ENVIRONMENT"} {
		if !strings.Contains(log, want) {
			t.Fatalf("abandoned-tool recovery missing %q:\n%s", want, log)
		}
	}
	for _, stale := range []string{"--resume inner-dirty", "dangerous_old_tool", "STALE_RESULT", "API Error: 409", "[Request interrupted by user]"} {
		if strings.Contains(log, stale) {
			t.Fatalf("abandoned-tool recovery replayed stale work %q:\n%s", stale, log)
		}
	}
	saved, ok := b.interactiveSession(owner, outer)
	if !ok || saved.SessionID != "inner-recovered" || saved.Dirty {
		t.Fatalf("abandoned-tool recovery did not replace dirty checkpoint: %#v, %v", saved, ok)
	}
}

func TestClaudeTUIRuntimeInteractiveAbandonedToolRecoveryIsSuffixOnly(t *testing.T) {
	req := &Request{Messages: []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "old question"}}},
		{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "old-call", Name: "read"}}},
		{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "old-call", Text: "old result"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "completed answer"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "new question"}}},
	}}
	if got := tuiRuntimeInteractiveAbandonedToolRecoveryIndex(req); got >= 0 {
		t.Fatalf("completed historical tool exchange was treated as abandoned suffix at %d", got)
	}

	compacted := &Request{Messages: []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "compacted summary"}}},
		{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "stale", Text: "old result"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "API Error: 409 US VPS Claude: Claude interactive migration: no preceding assistant turn"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "继续"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "[Request interrupted by user]"}}},
	}}
	if got := tuiRuntimeInteractiveAbandonedToolRecoveryIndex(compacted); got != 3 {
		t.Fatalf("real stale-result recovery index = %d, want 3", got)
	}
}

// An MCP callback is deliberately independent of the model-wrapper fixture:
// the callback blocks waiting for the client's second HTTP request, and
// that request must wake the SAME wrapper, without starting another one.
func TestClaudeTUIRuntimeInteractiveTransportMCPResultContinuesSamePTYRun(t *testing.T) {
	for _, wrapped := range []bool{false, true} {
		t.Run(fmt.Sprintf("wrapped-budget-%t", wrapped), func(t *testing.T) {
			tuiRuntimeTestInteractiveTransportMCPResultContinuesSamePTYRun(t, wrapped)
		})
	}
}

func tuiRuntimeTestInteractiveTransportMCPResultContinuesSamePTYRun(t *testing.T, wrapped bool) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	release := filepath.Join(dir, "tool-result-received")
	wrapper := filepath.Join(dir, "mock-pty")
	data := `#!/bin/sh
cat >/dev/null
printf '1\n' > "$CLAUDE_BRIDGE_USER_CONTEXT_PATH.ready"
printf '%s\n' '{"type":"assistant","uuid":"tool-row","session_id":"inner-A","message":{"id":"first","model":"claude-haiku-4-5","usage":{"input_tokens":11,"cache_creation_input_tokens":29},"content":[{"type":"tool_use","id":"call_1","name":"mcp__magpie__echo_test","input":{"text":"HELLO"}}]}}'
for i in $(seq 1 200); do
  if [ -f "$INTERACTIVE_TEST_RELEASE" ]; then break; fi
  sleep 0.05
done
if [ ! -f "$INTERACTIVE_TEST_RELEASE" ]; then exit 9; fi
printf '%s\n' '{"type":"assistant","uuid":"answer-row","session_id":"inner-A","message":{"id":"second","model":"claude-haiku-4-5","usage":{"input_tokens":4,"output_tokens":5,"cache_read_input_tokens":67},"content":[{"type":"text","text":"MCP_CONTINUED_OK"}]}}'
printf '%s\n' '{"type":"result","session_id":"inner-A","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(data), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_TEST_RELEASE", release)
	s := tuiRuntimeNewTestServer()
	var run *tuiRuntimeSubscriptionRun
	start := func(ctx context.Context, req *Request) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
		var err error
		var events <-chan Event
		run, events, err = s.subscription.startInteractiveOneTurn(ctx, req, "claude-haiku-4-5", "", "owner-A", "outer-A")
		return run, events, err
	}
	// Desktop may update a message-role system token budget notice while
	// the same native MCP callback is still waiting for its result.
	first := `{"model":"claude-haiku-4-5","max_tokens":100,"tools":[{"name":"echo_test","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"Call echo_test."},{"role":"system","content":"<total_tokens>15000000 tokens left</total_tokens>"}]}`
	if wrapped {
		first = strings.Replace(first, `<total_tokens>15000000 tokens left</total_tokens>`, `<system-reminder>\n<total_tokens>15000000 tokens left</total_tokens>\n</system-reminder>`, 1)
	}
	rec := httptest.NewRecorder()
	code, msg := s.serveSubscription(rec, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(first)),
		protoAnthropic, "Claude Code", "claude-haiku-4-5", "owner-A", []byte(first), &Usage{}, start)
	if code != 200 || !strings.Contains(rec.Body.String(), `"type":"tool_use"`) ||
		!strings.Contains(rec.Body.String(), `"id":"call_1"`) {
		t.Fatalf("first response did not yield the MCP call: %d %s %s", code, msg, rec.Body)
	}
	if run == nil {
		t.Fatal("tool-producing run missing")
	}
	defer run.abort()
	// Simulate the genuine Claude MCP helper posting its tool call.
	mux := http.NewServeMux()
	mux.HandleFunc("POST /cb/{token}", s.subscription.mcpCall)
	callbackResult := make(chan struct {
		code int
		body string
	}, 1)
	go func() {
		body := `{"tool_call_id":"call_1","name":"echo_test","arguments":{"text":"HELLO"}}`
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/cb/"+run.token, strings.NewReader(body)))
		callbackResult <- struct {
			code int
			body string
		}{w.Code, w.Body.String()}
		// A real PostToolUse hook receives user input through its own channel,
		// after the tool's MCP result has returned, rather than in that result.
		contextPath := filepath.Join(run.tmp, "user-context.json")
		contextData, _ := os.ReadFile(contextPath)
		var update tuiRuntimeInteractiveUserContext
		if json.Unmarshal(contextData, &update) == nil && update.ID != "" &&
			strings.Contains(update.Text, "new user message arrived while working") {
			_ = os.WriteFile(contextPath+".ack", []byte(`{"id":"`+update.ID+`"}`), 0o600)
		}
		// The test wrapper cannot finish before MCP has returned its result.
		_ = os.WriteFile(release, []byte("ok"), 0o600)
	}()
	// Wait until the callback has actually registered before continuing.
	deadline := time.Now().Add(3 * time.Second)
	for {
		s.subscription.mu.Lock()
		waiting := s.subscription.calls["call_1"] == run
		s.subscription.mu.Unlock()
		if waiting {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("MCP callback never registered its call")
		}
		time.Sleep(5 * time.Millisecond)
	}
	second := `{"model":"claude-haiku-4-5","max_tokens":100,"tools":[{"name":"echo_test","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"Call echo_test."},{"role":"system","content":"<total_tokens>14998088 tokens left</total_tokens>"},{"role":"assistant","content":[{"type":"tool_use","id":"call_1","name":"echo_test","input":{"text":"HELLO"}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call_1","content":"HELLO"},{"type":"text","text":"<system-reminder>new user message arrived while working</system-reminder>"}]}]}`
	if wrapped {
		second = strings.Replace(second, `<total_tokens>14998088 tokens left</total_tokens>`, `<system-reminder>\n<total_tokens>14998088 tokens left</total_tokens>\n</system-reminder>`, 1)
	}
	for _, tc := range []struct {
		name, owner, model, body string
	}{
		{"wrong owner", "owner-B", "claude-haiku-4-5", second},
		{"wrong model", "owner-A", "claude-opus-5-5", second},
		{"changed tool schema", "owner-A", "claude-haiku-4-5", strings.Replace(second, `"input_schema":{"type":"object"}`, `"input_schema":{"type":"object","required":["text"]}`, 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code, _ := s.serveSubscription(httptest.NewRecorder(), httptest.NewRequest("POST", "/v1/messages", strings.NewReader(tc.body)),
				protoAnthropic, "Claude Code", tc.model, tc.owner, []byte(tc.body), &Usage{}, start)
			if code != http.StatusConflict {
				t.Fatalf("unsafe MCP handoff accepted: status=%d", code)
			}
			run.mu.Lock()
			closed := run.closed
			run.mu.Unlock()
			if closed {
				t.Fatal("bad handoff aborted another owner's valid PTY run")
			}
		})
	}
	rec = httptest.NewRecorder()
	code, msg = s.serveSubscription(rec, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(second)),
		protoAnthropic, "Claude Code", "claude-haiku-4-5", "owner-A", []byte(second), &Usage{}, start)
	if code != 200 || !strings.Contains(rec.Body.String(), "MCP_CONTINUED_OK") {
		t.Fatalf("tool result failed to resume the original PTY: %d %s %s", code, msg, rec.Body)
	}
	select {
	case got := <-callbackResult:
		if got.code != 200 || !strings.Contains(got.body, "HELLO") ||
			strings.Contains(got.body, "new user message arrived while working") {
			t.Fatalf("MCP callback lost tool data or mixed user input into it: %+v", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("MCP callback never received the forwarded tool result")
	}
	select {
	case <-run.interactiveDone:
	case <-time.After(3 * time.Second):
		t.Fatal("tool-result wrapper did not exit after final answer")
	}
	// The old MCP result must never bootstrap another process on replay.
	code, _ = s.serveSubscription(httptest.NewRecorder(), httptest.NewRequest("POST", "/v1/messages", strings.NewReader(second)),
		protoAnthropic, "Claude Code", "claude-haiku-4-5", "owner-A", []byte(second), &Usage{}, start)
	if code != http.StatusConflict {
		t.Fatalf("stale MCP result was replayed in a new run: status=%d", code)
	}
	if saved, ok := s.subscription.interactiveSession("owner-A", "outer-A"); !ok || saved.SessionID != "inner-A" || saved.Dirty {
		t.Fatalf("completed MCP turn did not persist a clean checkpoint: %#v, %v", saved, ok)
	}
}

func TestClaudeTUIRuntimeInteractiveTransportSequentialMCPContinuesSamePTYRun(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	release1 := filepath.Join(dir, "release-1")
	release2 := filepath.Join(dir, "release-2")
	wrapper := filepath.Join(dir, "mock-pty-chain")
	data := `#!/bin/sh
cat >/dev/null
printf '%s\n' '{"type":"assistant","uuid":"tool-row-1","session_id":"inner-chain","message":{"id":"m1","model":"claude-haiku-4-5","content":[{"type":"tool_use","id":"call_1","name":"mcp__magpie__echo_test","input":{"text":"ONE"}}]}}'
for i in $(seq 1 200); do [ -f "$INTERACTIVE_RELEASE_1" ] && break; sleep 0.02; done
[ -f "$INTERACTIVE_RELEASE_1" ] || exit 11
printf '%s\n' '{"type":"assistant","uuid":"tool-row-2","session_id":"inner-chain","message":{"id":"m2","model":"claude-haiku-4-5","content":[{"type":"tool_use","id":"call_2","name":"mcp__magpie__echo_test","input":{"text":"TWO"}}]}}'
for i in $(seq 1 200); do [ -f "$INTERACTIVE_RELEASE_2" ] && break; sleep 0.02; done
[ -f "$INTERACTIVE_RELEASE_2" ] || exit 12
printf '%s\n' '{"type":"assistant","uuid":"answer-row","session_id":"inner-chain","message":{"id":"m3","model":"claude-haiku-4-5","content":[{"type":"text","text":"CHAIN_DONE"}]}}'
printf '%s\n' '{"type":"result","session_id":"inner-chain","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(data), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_RELEASE_1", release1)
	t.Setenv("INTERACTIVE_RELEASE_2", release2)
	s := tuiRuntimeNewTestServer()
	starts := 0
	var run *tuiRuntimeSubscriptionRun
	start := func(ctx context.Context, req *Request) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
		starts++
		var err error
		var events <-chan Event
		run, events, err = s.subscription.startInteractiveOneTurn(ctx, req, "claude-haiku-4-5", "", "owner-chain", "outer-chain")
		return run, events, err
	}
	serve := func(body string) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		code, msg := s.serveSubscription(rec, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body)),
			protoAnthropic, "Claude Code", "claude-haiku-4-5", "owner-chain", []byte(body), &Usage{}, start)
		if code != 200 {
			t.Fatalf("request failed: %d %s %s", code, msg, rec.Body.String())
		}
		return rec
	}
	first := `{"model":"claude-haiku-4-5","max_tokens":100,"tools":[{"name":"echo_test","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"Use the tool twice."}]}`
	if rec := serve(first); !strings.Contains(rec.Body.String(), `"id":"call_1"`) {
		t.Fatalf("first tool call missing: %s", rec.Body)
	}
	if run == nil {
		t.Fatal("interactive run missing")
	}
	defer run.abort()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /cb/{token}", s.subscription.mcpCall)
	callback := func(id, value, release string) <-chan string {
		t.Helper()
		out := make(chan string, 1)
		go func() {
			body := fmt.Sprintf(`{"tool_call_id":%q,"name":"echo_test","arguments":{"text":%q}}`, id, value)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest("POST", "/cb/"+run.token, strings.NewReader(body)))
			out <- w.Body.String()
			_ = os.WriteFile(release, []byte("ok"), 0o600)
		}()
		deadline := time.Now().Add(2 * time.Second)
		for {
			s.subscription.mu.Lock()
			waiting := s.subscription.calls[id] == run
			s.subscription.mu.Unlock()
			if waiting {
				return out
			}
			if time.Now().After(deadline) {
				t.Fatalf("callback %s never registered", id)
			}
			time.Sleep(5 * time.Millisecond)
		}
	}
	cb1 := callback("call_1", "ONE", release1)
	second := `{"model":"claude-haiku-4-5","max_tokens":100,"tools":[{"name":"echo_test","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"Use the tool twice."},{"role":"assistant","content":[{"type":"tool_use","id":"call_1","name":"echo_test","input":{"text":"ONE"}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call_1","content":"RESULT_ONE"}]}]}`
	if rec := serve(second); !strings.Contains(rec.Body.String(), `"id":"call_2"`) {
		t.Fatalf("second tool call missing: %s", rec.Body)
	}
	if got := <-cb1; !strings.Contains(got, "RESULT_ONE") {
		t.Fatalf("first MCP result not forwarded: %s", got)
	}
	cb2 := callback("call_2", "TWO", release2)
	third := `{"model":"claude-haiku-4-5","max_tokens":100,"tools":[{"name":"echo_test","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"Use the tool twice."},{"role":"assistant","content":[{"type":"tool_use","id":"call_1","name":"echo_test","input":{"text":"ONE"}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call_1","content":"RESULT_ONE"}]},{"role":"assistant","content":[{"type":"tool_use","id":"call_2","name":"echo_test","input":{"text":"TWO"}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call_2","content":"RESULT_TWO"}]}]}`
	if rec := serve(third); !strings.Contains(rec.Body.String(), "CHAIN_DONE") {
		t.Fatalf("final answer missing: %s", rec.Body)
	}
	if got := <-cb2; !strings.Contains(got, "RESULT_TWO") {
		t.Fatalf("second MCP result not forwarded: %s", got)
	}
	if starts != 1 {
		t.Fatalf("tool chain started %d PTY runs, want exactly one", starts)
	}
	select {
	case <-run.interactiveDone:
	case <-time.After(2 * time.Second):
		t.Fatal("sequential MCP wrapper did not exit")
	}
}

func TestClaudeTUIRuntimeInteractiveTransportParallelMCPResultsCanArriveTogether(t *testing.T) {
	b := tuiRuntimeNewSubscriptionBridge()
	run := &tuiRuntimeSubscriptionRun{
		bridge: b, token: "parallel", interactive: true, owner: "owner",
		pending: map[string]chan tuiRuntimeMcpToolResult{},
		tools:   map[string]bool{"echo_test": true},
	}
	b.runs[run.token] = run
	call := func(id string) <-chan string {
		out := make(chan string, 1)
		go func() {
			body := fmt.Sprintf(`{"tool_call_id":%q,"name":"echo_test","arguments":{}}`, id)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/cb/"+run.token, strings.NewReader(body))
			req.SetPathValue("token", run.token)
			b.mcpCall(w, req)
			out <- w.Body.String()
		}()
		deadline := time.Now().Add(2 * time.Second)
		for {
			b.mu.Lock()
			waiting := b.calls[id] == run
			b.mu.Unlock()
			if waiting {
				return out
			}
			if time.Now().After(deadline) {
				t.Fatalf("parallel callback %s never registered", id)
			}
			time.Sleep(5 * time.Millisecond)
		}
	}
	one := call("call_1")
	two := call("call_2")
	assistant := Message{Role: "assistant", Parts: []Part{
		{Kind: ToolCall, ID: "call_1", Name: "echo_test"},
		{Kind: ToolCall, ID: "call_2", Name: "echo_test"},
	}}
	user := Message{Role: "user", Parts: []Part{
		{Kind: ToolResult, CallID: "call_1", Text: "ONE"},
		{Kind: ToolResult, CallID: "call_2", Text: "TWO"},
	}}
	req := &Request{
		Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "parallel"}}}, assistant, user},
		Tools:    []Tool{{Name: "echo_test"}},
	}
	if err := tuiRuntimeValidateInteractiveToolContinuation(req); err != nil {
		t.Fatal(err)
	}
	found, results := b.findRun(req)
	if found != run || len(results) != 2 {
		t.Fatalf("parallel results did not find one run: found=%t results=%d", found == run, len(results))
	}
	if _, err := run.continueWith(results, nil); err != nil {
		t.Fatal(err)
	}
	if got := <-one; !strings.Contains(got, "ONE") {
		t.Fatalf("first parallel result missing: %s", got)
	}
	if got := <-two; !strings.Contains(got, "TWO") {
		t.Fatalf("second parallel result missing: %s", got)
	}
}

func TestClaudeTUIRuntimeInteractiveTransportToolResultWaitsForLateCallbackRegistration(t *testing.T) {
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	run := &tuiRuntimeSubscriptionRun{
		bridge: b, token: "late", interactive: true, owner: "owner", model: "claude-haiku-4-5",
		pending: map[string]chan tuiRuntimeMcpToolResult{}, tools: map[string]bool{"echo_test": true},
	}
	b.runs[run.token] = run
	req := &Request{
		Model: "claude-haiku-4-5",
		Messages: []Message{
			{Role: "user", Parts: []Part{{Kind: Text, Text: "call it"}}},
			{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "call_late", Name: "echo_test"}}},
			{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "call_late", Text: "LATE_OK"}}},
		},
		Tools: []Tool{{Name: "echo_test"}},
	}
	waiter := make(chan tuiRuntimeMcpToolResult, 1)
	registered := make(chan struct{})
	go func() {
		time.Sleep(125 * time.Millisecond)
		run.mu.Lock()
		run.pending["call_late"] = waiter
		run.mu.Unlock()
		b.mu.Lock()
		b.calls["call_late"] = run
		b.mu.Unlock()
		close(registered)
	}()
	started := time.Now()
	found, results, how := b.match(req)
	if found != run || how != tuiRuntimeByExactID || len(results) != 1 {
		t.Fatalf("late callback registration not matched: found=%t how=%q results=%d", found == run, how, len(results))
	}
	if elapsed := time.Since(started); elapsed < 100*time.Millisecond || elapsed > 550*time.Millisecond {
		t.Fatalf("callback grace waited %s, want bounded delayed match", elapsed)
	}
	<-registered
	ch, err := run.continueWith(results, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := <-waiter
	if len(got.Content) == 0 || !strings.Contains(got.Content[0]["text"].(string), "LATE_OK") {
		t.Fatalf("late callback got wrong result: %+v", got)
	}
	run.endSegment()
	for range ch {
	}
}

func TestClaudeTUIRuntimeInteractiveTransportAllowsSequentialToolRounds(t *testing.T) {
	run := &tuiRuntimeSubscriptionRun{tools: map[string]bool{"echo_test": true}}
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	first := run.attach()
	done := make(chan struct{})
	go func() {
		run.readInteractiveOneTurn(reader)
		close(done)
	}()
	write := func(line string) {
		t.Helper()
		if _, err := io.WriteString(writer, line+"\n"); err != nil {
			t.Fatal(err)
		}
	}
	collect := func(ch <-chan Event) []Event {
		t.Helper()
		var events []Event
		timer := time.NewTimer(2 * time.Second)
		defer timer.Stop()
		for {
			select {
			case ev, ok := <-ch:
				if !ok {
					return events
				}
				events = append(events, ev)
			case <-timer.C:
				t.Fatal("interactive segment did not close")
			}
		}
	}
	assertTool := func(events []Event, id string) {
		t.Helper()
		var call, stop bool
		for _, ev := range events {
			if ev.Kind == KError {
				t.Fatalf("sequential tool round failed: %s", ev.Text)
			}
			call = call || ev.Kind == KToolStart && ev.ID == id && ev.Name == "echo_test"
			stop = stop || ev.Kind == KStop && ev.Stop == "tool"
		}
		if !call || !stop {
			t.Fatalf("tool round missing events: %+v", events)
		}
	}

	write(`{"type":"assistant","uuid":"row-1","message":{"id":"m1","model":"claude-haiku-4-5","content":[{"type":"tool_use","id":"call_1","name":"mcp__magpie__echo_test","input":{"text":"one"}}]}}`)
	assertTool(collect(first), "call_1")

	run.mu.Lock()
	run.resuming = true // tool_result request claimed this same PTY run
	run.mu.Unlock()
	second := run.attach()
	write(`{"type":"assistant","uuid":"row-2","message":{"id":"m2","model":"claude-haiku-4-5","content":[{"type":"tool_use","id":"call_2","name":"mcp__magpie__echo_test","input":{"text":"two"}}]}}`)
	assertTool(collect(second), "call_2")

	run.mu.Lock()
	run.resuming = true
	run.mu.Unlock()
	third := run.attach()
	write(`{"type":"assistant","uuid":"row-3","message":{"id":"m3","model":"claude-haiku-4-5","content":[{"type":"text","text":"ALL_DONE"}]}}`)
	write(`{"type":"result","stop_reason":"end_turn"}`)
	var text string
	for _, ev := range collect(third) {
		if ev.Kind == KError {
			t.Fatalf("final answer after sequential tools failed: %s", ev.Text)
		}
		if ev.Kind == KText {
			text += ev.Text
		}
	}
	if text != "ALL_DONE" {
		t.Fatalf("final answer = %q", text)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("interactive transcript reader did not finish")
	}
}

// Claude Code can serialize one assistant message as several whole-message
// rows, one tool_use block per row. The PTY transport must wait for the same
// quiet boundary as the official headless parser before ending the client's
// tool turn, or the second parallel call is lost/rejected.
func TestClaudeTUIRuntimeInteractiveTransportCollectsSplitParallelToolRows(t *testing.T) {
	was := tuiRuntimeUnstreamedQuiet
	tuiRuntimeUnstreamedQuiet = 40 * time.Millisecond
	t.Cleanup(func() { tuiRuntimeUnstreamedQuiet = was })
	run := &tuiRuntimeSubscriptionRun{tools: map[string]bool{"read": true}}
	reader, writer := io.Pipe()
	t.Cleanup(func() { _ = reader.Close(); _ = writer.Close() })
	events := run.attach()
	done := make(chan struct{})
	go func() {
		run.readInteractiveOneTurn(reader)
		close(done)
	}()
	write := func(line string) {
		t.Helper()
		if _, err := io.WriteString(writer, line+"\n"); err != nil {
			t.Fatal(err)
		}
	}
	write(`{"type":"assistant","uuid":"row-a","message":{"id":"msg_whole","model":"claude-opus-5-5","stop_reason":"tool_use","content":[{"type":"tool_use","id":"toolu_a","name":"mcp__magpie__read","input":{"p":1}}]}}`)
	write(`{"type":"assistant","uuid":"row-b","message":{"id":"msg_whole","model":"claude-opus-5-5","stop_reason":"tool_use","content":[{"type":"tool_use","id":"toolu_b","name":"mcp__magpie__read","input":{"p":2}}]}}`)
	var calls []string
	stops := 0
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	for {
		select {
		case ev, ok := <-events:
			if !ok {
				if strings.Join(calls, ",") != "toolu_a,toolu_b" || stops != 1 {
					t.Fatalf("split parallel calls=%v stops=%d", calls, stops)
				}
				_ = writer.Close()
				select {
				case <-done:
				case <-time.After(time.Second):
					t.Fatal("interactive parser did not stop after pipe close")
				}
				return
			}
			switch ev.Kind {
			case KToolStart:
				calls = append(calls, ev.ID)
			case KStop:
				if ev.Stop == "tool" {
					stops++
				}
			case KError:
				t.Fatalf("split parallel tool rows failed: %s", ev.Text)
			}
		case <-deadline.C:
			t.Fatalf("split parallel tool turn did not settle: calls=%v stops=%d", calls, stops)
		}
	}
}

func TestClaudeTUIRuntimeInteractiveTransportRejectsMissingTerminalResult(t *testing.T) {
	run := &tuiRuntimeSubscriptionRun{}
	ch := run.attach()
	go run.readInteractiveOneTurn(strings.NewReader(`{"type":"assistant","uuid":"1","message":{"id":"m","content":[{"type":"text","text":"partial"}]}}`))
	var sawError bool
	for ev := range ch {
		if ev.Kind == KError && strings.Contains(ev.Text, "without a terminal result") {
			sawError = true
		}
	}
	if !sawError {
		t.Fatal("incomplete wrapper output was silently accepted")
	}
}

func TestClaudeTUIRuntimeInteractiveTransportKeepsStreamAliveBeforeFirstModelEvent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "quiet-bridge")
	script := `#!/bin/sh
cat >/dev/null
sleep 0.12
printf '%s\n' '{"type":"assistant","uuid":"quiet-row","session_id":"quiet-inner","message":{"id":"quiet-msg","model":"claude-haiku-4-5","content":[{"type":"text","text":"KEEPALIVE_OK"}]}}'
printf '%s\n' '{"type":"result","session_id":"quiet-inner","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	oldAlive := tuiRuntimeInteractiveAliveEvery
	tuiRuntimeInteractiveAliveEvery = 20 * time.Millisecond
	t.Cleanup(func() { tuiRuntimeInteractiveAliveEvery = oldAlive })
	tuiRuntimeKeepFast(t, 5*time.Millisecond, 20*time.Millisecond, 300*time.Millisecond)

	s := tuiRuntimeNewTestServer()
	t.Cleanup(s.subscription.abortAll)
	body := `{"model":"claude-haiku-4-5","max_tokens":64,"stream":true,"messages":[{"role":"user","content":"hello"}]}`
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
	rec := httptest.NewRecorder()
	var usage Usage
	start := func(ctx context.Context, parsed *Request) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
		return s.subscription.startInteractiveOneTurn(ctx, parsed, parsed.Model, "", "keepalive-owner", "")
	}
	code, failed := s.serveSubscription(rec, req, protoAnthropic, "Claude Code", "claude-haiku-4-5", "keepalive-owner", []byte(body), &usage, start)
	if code != 200 || failed != "" {
		t.Fatalf("quiet PTY stream = %d %q: %s", code, failed, rec.Body.String())
	}
	out := rec.Body.String()
	if !strings.Contains(out, "KEEPALIVE_OK") || !strings.Contains(out, `"type":"message_stop"`) {
		t.Fatalf("quiet PTY reply incomplete:\n%s", out)
	}
	if pings := strings.Count(out, `"type":"ping"`); pings < 2 {
		t.Fatalf("quiet PTY emitted only %d pre-output/downstream pings:\n%s", pings, out)
	}
	firstPing, answer := strings.Index(out, `"type":"ping"`), strings.Index(out, "KEEPALIVE_OK")
	if firstPing < 0 || answer < 0 || firstPing >= answer {
		t.Fatalf("first keepalive did not precede the model answer:\n%s", out)
	}
}

func TestClaudeTUIRuntimeInteractivePersistedExactResumeUsesInnerSessionAndSuffixOnly(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "resume-bridge")
	logPath := filepath.Join(dir, "resume.log")
	script := `#!/bin/sh
prompt=$(cat)
printf '%s\n' '---' >> "$INTERACTIVE_RESUME_LOG"
printf 'args:%s\n' "$*" >> "$INTERACTIVE_RESUME_LOG"
printf 'prompt:%s\n' "$prompt" >> "$INTERACTIVE_RESUME_LOG"
case " $* " in
  *" --resume inner-main "*) text=SECOND ;;
  *) text=FIRST ;;
esac
printf '{"type":"assistant","uuid":"row-%s","session_id":"inner-main","message":{"id":"m-%s","model":"claude-opus-5-5","usage":{"input_tokens":3,"output_tokens":2},"content":[{"type":"text","text":"%s"}]}}\n' "$text" "$text" "$text"
printf '{"type":"result","session_id":"inner-main","stop_reason":"end_turn"}\n'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_RESUME_LOG", logPath)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer, model := "owner", "outer", "claude-opus-5-5"
	collect := func(run *tuiRuntimeSubscriptionRun, events <-chan Event) (string, string) {
		t.Helper()
		var text, stop string
		for ev := range events {
			switch ev.Kind {
			case KText:
				text += ev.Text
			case KStop:
				stop = ev.Stop
			case KError:
				t.Fatalf("interactive turn failed: %s", ev.Text)
			}
		}
		select {
		case <-run.interactiveDone:
		case <-time.After(2 * time.Second):
			t.Fatal("wrapper did not exit")
		}
		return text, stop
	}
	req1 := &Request{Model: model, Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}}}}
	run1, events1, err := b.startInteractiveRequest(context.Background(), req1, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	text1, stop1 := collect(run1, events1)
	if text1 != "FIRST" || stop1 != "stop" {
		t.Fatalf("first turn = %q %q", text1, stop1)
	}
	run1.ended(req1, text1, stop1, true)
	saved, ok := b.interactiveSession(owner, outer)
	if !ok || saved.SessionID != "inner-main" || saved.Dirty {
		t.Fatalf("first turn checkpoint = %#v, %v", saved, ok)
	}
	firstDir := run1.cmd.Dir
	tuiRuntimeBridges.Delete(b)
	b2 := tuiRuntimeNewSubscriptionBridge() // simulate a Magpie gateway restart
	defer tuiRuntimeBridges.Delete(b2)

	req2 := &Request{Model: model, Messages: []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: text1}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "and?"}}},
	}}
	if err := tuiRuntimeValidateInteractiveRequest(req2); err != nil {
		t.Fatalf("ordinary next turn rejected before restore: %v", err)
	}
	run2, events2, err := b2.startInteractiveRequest(context.Background(), req2, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	if run2.cmd.Dir != firstDir {
		t.Fatalf("resumed turn changed cwd: %q != %q", run2.cmd.Dir, firstDir)
	}
	dirty, ok := b2.interactiveSession(owner, outer)
	if !ok || !dirty.Dirty || dirty.SessionID != "inner-main" {
		t.Fatalf("resume was not marked dirty before completion: %#v, %v", dirty, ok)
	}
	text2, stop2 := collect(run2, events2)
	if text2 != "SECOND" || stop2 != "stop" {
		t.Fatalf("second turn = %q %q", text2, stop2)
	}
	run2.ended(req2, text2, stop2, true)
	clean, ok := b2.interactiveSession(owner, outer)
	if !ok || clean.Dirty || clean.SessionID != "inner-main" {
		t.Fatalf("completed resume did not clean checkpoint: %#v, %v", clean, ok)
	}
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(string(logBytes), "---\n")
	if len(parts) < 3 {
		t.Fatalf("wrapper log missing turns:\n%s", logBytes)
	}
	second := parts[2]
	if !strings.Contains(second, "--resume inner-main") {
		t.Fatalf("second wrapper did not resume inner session:\n%s", second)
	}
	if !strings.Contains(second, "prompt:and?") || strings.Contains(second, "hello") || strings.Contains(second, "FIRST") {
		t.Fatalf("resumed wrapper got replayed history instead of suffix:\n%s", second)
	}
}

func TestClaudeTUIRuntimeInteractivePersistedResumeChangesEffortWithoutChangingInnerSession(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "effort-bridge")
	logPath := filepath.Join(dir, "effort.log")
	script := `#!/bin/sh
prompt=$(cat)
printf 'args:%s\nprompt:%s\n---\n' "$*" "$prompt" >> "$INTERACTIVE_EFFORT_LOG"
printf '%s\n' '{"type":"assistant","uuid":"row","session_id":"inner-effort","message":{"id":"m","model":"claude-opus-5-5","content":[{"type":"text","text":"OK"}]}}'
printf '%s\n' '{"type":"result","session_id":"inner-effort","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_EFFORT_LOG", logPath)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer, model := "owner", "outer-effort", "claude-opus-5-5"
	collect := func(run *tuiRuntimeSubscriptionRun, events <-chan Event, req *Request) string {
		t.Helper()
		var text, stop string
		for ev := range events {
			switch ev.Kind {
			case KText:
				text += ev.Text
			case KStop:
				stop = ev.Stop
			case KError:
				t.Fatal(ev.Text)
			}
		}
		if stop != "stop" {
			t.Fatalf("turn stop = %q", stop)
		}
		run.ended(req, text, stop, true)
		select {
		case <-run.interactiveDone:
		case <-time.After(2 * time.Second):
			t.Fatal("effort wrapper did not exit")
		}
		return text
	}
	req1 := &Request{Model: model, Effort: "high", Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}}}}
	run1, ev1, err := b.startInteractiveRequest(context.Background(), req1, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	first := collect(run1, ev1, req1)
	req2 := &Request{Model: model, Effort: "max", Messages: []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: first}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "more"}}},
	}}
	run2, ev2, err := b.startInteractiveRequest(context.Background(), req2, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	_ = collect(run2, ev2, req2)
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	log := string(logBytes)
	if !strings.Contains(log, "--effort high") || !strings.Contains(log, "--effort max") ||
		!strings.Contains(log, "--resume inner-effort") {
		t.Fatalf("effort change did not resume same inner session:\n%s", log)
	}
	if strings.Count(log, "--resume inner-effort") != 1 {
		t.Fatalf("first turn unexpectedly resumed or second did not resume exactly once:\n%s", log)
	}
}

func TestClaudeTUIRuntimeInteractivePersistedReplyAnchorResumesAfterOuterRewrite(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "reply-anchor-bridge")
	logPath := filepath.Join(dir, "reply-anchor.log")
	script := `#!/bin/sh
prompt=$(cat)
printf 'args:%s\nprompt:%s\n' "$*" "$prompt" >> "$INTERACTIVE_REPLY_ANCHOR_LOG"
printf '%s\n' '{"type":"assistant","uuid":"row","session_id":"inner-anchor","message":{"id":"m","model":"claude-opus-5-5","content":[{"type":"text","text":"ANCHOR_OK"}]}}'
printf '%s\n' '{"type":"result","session_id":"inner-anchor","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_REPLY_ANCHOR_LOG", logPath)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer, model := "owner", "outer-anchor", "claude-opus-5-5"
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "original history"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "SAVED_REPLY"}}},
	}
	baseReq := &Request{Model: model, Messages: base}
	keys := tuiRuntimeConversationKeys(owner, base)
	b.saveInteractiveSession(owner, outer, tuiRuntimeInteractiveSessionEntry{
		SessionID: "inner-anchor", Model: model, ThreadKey: firstWords(baseReq),
		ConvKey: keys[0], ReplyKey: tuiRuntimeAssistantReplyKey(base[1]), ContextKey: tuiRuntimeTurnKey(owner, baseReq, nil),
	})
	// The same Runtime normally still remembers the original outer prefix.
	// A compacted prefix must not be mistaken for a helper before checking
	// the unique saved assistant anchor.
	b.seedOuterThread(owner, outer, firstWords(baseReq))
	rewritten := &Request{Model: model, Messages: []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "compacted earlier history"}}},
		base[1],
		{Role: "user", Parts: []Part{{Kind: Text, Text: "new request"}}},
	}}
	run, events, err := b.startInteractiveRequest(context.Background(), rewritten, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	var text, stop string
	for ev := range events {
		switch ev.Kind {
		case KText:
			text += ev.Text
		case KStop:
			stop = ev.Stop
		case KError:
			t.Fatal(ev.Text)
		}
	}
	if text != "ANCHOR_OK" || stop != "stop" {
		t.Fatalf("reply-anchor resume = %q %q", text, stop)
	}
	run.ended(rewritten, text, stop, true)
	if run.outerSession != outer || b.outerThread(owner, outer) != firstWords(rewritten) {
		t.Fatal("verified compacted continuation was not retained as the canonical outer thread")
	}
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	log := string(logBytes)
	if !strings.Contains(log, "--resume inner-anchor") || !strings.Contains(log, "prompt:new request") ||
		strings.Contains(log, "compacted earlier history") || strings.Contains(log, "SAVED_REPLY") {
		t.Fatalf("reply-anchor replayed outer history or missed resume:\n%s", log)
	}
}

func TestClaudeTUIRuntimeInteractivePersistedResumeFailsClosedWhenDirtyOrModelChanged(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer, model := "owner", "outer", "claude-opus-5-5"
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "answer"}}},
	}
	baseReq := &Request{Model: model, Messages: base}
	keys := tuiRuntimeConversationKeys(owner, base)
	entry := tuiRuntimeInteractiveSessionEntry{
		SessionID: "inner", Model: model, ThreadKey: firstWords(baseReq),
		ConvKey: keys[0], ReplyKey: tuiRuntimeAssistantReplyKey(base[1]), ContextKey: tuiRuntimeTurnKey(owner, baseReq, nil),
	}
	b.saveInteractiveSession(owner, outer, entry)
	next := &Request{Model: model, Messages: append(append([]Message{}, base...), Message{Role: "user", Parts: []Part{{Kind: Text, Text: "next"}}})}

	b.markInteractiveSessionDirty(owner, outer)
	if _, _, err := b.startInteractiveRequest(context.Background(), next, model, "", owner, outer); err == nil || !strings.Contains(err.Error(), "dirty") {
		t.Fatalf("dirty checkpoint did not fail closed: %v", err)
	}

	entry.Dirty = false
	b.saveInteractiveSession(owner, outer, entry)
	if _, _, err := b.startInteractiveRequest(context.Background(), next, "claude-haiku-4-5", "", owner, outer); err == nil || !strings.Contains(err.Error(), "model_mismatch") {
		t.Fatalf("changed model resumed old checkpoint: %v", err)
	}
	if got, _ := b.interactiveSession(owner, outer); got.Dirty {
		t.Fatal("model mismatch dirtied a checkpoint that was never launched")
	}
}

func TestClaudeTUIRuntimeInteractivePersistedContextChangeRefreshesWithoutHistoryReplay(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "context-refresh-bridge")
	logPath := filepath.Join(dir, "context-refresh.log")
	script := `#!/bin/sh
prompt=$(cat)
printf 'args:%s\nprompt:%s\n' "$*" "$prompt" > "$INTERACTIVE_CONTEXT_LOG"
printf '%s\n' '{"type":"assistant","uuid":"row","session_id":"inner-context","message":{"id":"m","model":"claude-opus-5-5","content":[{"type":"text","text":"CTX_OK"}]}}'
printf '%s\n' '{"type":"result","session_id":"inner-context","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_CONTEXT_LOG", logPath)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer, model := "owner", "outer-context", "claude-opus-5-5"
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "original hello"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "SAVED_REPLY"}}},
	}
	baseReq := &Request{Model: model, System: "old system", Tools: []Tool{{Name: "old_tool"}}, Messages: base}
	keys := tuiRuntimeConversationKeys(owner, base)
	b.saveInteractiveSession(owner, outer, tuiRuntimeInteractiveSessionEntry{
		SessionID: "inner-context", Model: model, ThreadKey: firstWords(baseReq),
		ConvKey: keys[0], ReplyKey: tuiRuntimeAssistantReplyKey(base[1]), ContextKey: tuiRuntimeTurnKey(owner, baseReq, nil),
	})
	changed := &Request{
		Model: model, System: "new system", ToolChoice: "required", Tools: []Tool{{Name: "new_tool"}},
		Messages: append(append([]Message{}, base...), Message{Role: "user", Parts: []Part{{Kind: Text, Text: "current request"}}}),
	}
	run, events, err := b.startInteractiveRequest(context.Background(), changed, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	var text, stop string
	for ev := range events {
		switch ev.Kind {
		case KText:
			text += ev.Text
		case KStop:
			stop = ev.Stop
		case KError:
			t.Fatal(ev.Text)
		}
	}
	if text != "CTX_OK" || stop != "stop" {
		t.Fatalf("context refresh resume = %q %q", text, stop)
	}
	run.ended(changed, text, stop, true)
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	log := string(logBytes)
	for _, want := range []string{"--resume inner-context", "<external_context_refresh>", "new system", "current request", "You must call at least one available tool"} {
		if !strings.Contains(log, want) {
			t.Fatalf("context refresh missing %q:\n%s", want, log)
		}
	}
	for _, old := range []string{"original hello", "SAVED_REPLY", "old system"} {
		if strings.Contains(log, old) {
			t.Fatalf("context refresh replayed stale context %q:\n%s", old, log)
		}
	}
}

func TestClaudeTUIRuntimeInteractivePersistedDirtyCompanionCanResumeSameInnerSession(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "companion-bridge")
	logPath := filepath.Join(dir, "companion.log")
	script := `#!/bin/sh
prompt=$(cat)
printf 'args:%s\nprompt:%s\n' "$*" "$prompt" > "$INTERACTIVE_COMPANION_LOG"
printf '%s\n' '{"type":"assistant","uuid":"row","session_id":"inner-companion","message":{"id":"m","model":"claude-opus-5-5","content":[{"type":"text","text":"COMPANION_OK"}]}}'
printf '%s\n' '{"type":"result","session_id":"inner-companion","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_COMPANION_LOG", logPath)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer, model := "owner", "outer-companion", "claude-opus-5-5"
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "SAVED_REPLY"}}},
	}
	baseReq := &Request{Model: model, Messages: base}
	keys := tuiRuntimeConversationKeys(owner, base)
	b.saveInteractiveSession(owner, outer, tuiRuntimeInteractiveSessionEntry{
		SessionID: "inner-companion", Model: model, ThreadKey: firstWords(baseReq),
		ConvKey: keys[0], ReplyKey: tuiRuntimeAssistantReplyKey(base[1]), ContextKey: tuiRuntimeTurnKey(owner, baseReq, nil), Dirty: true,
	})
	companionText := "Your response above was cut off mid-stream. Resume directly from where it stops — no apology, no recap."
	req := &Request{Model: model, Messages: append(append([]Message{}, base...), Message{Role: "user", Parts: []Part{{Kind: Text, Text: companionText}}})}
	run, events, err := b.startInteractiveRequest(context.Background(), req, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	var text, stop string
	for ev := range events {
		switch ev.Kind {
		case KText:
			text += ev.Text
		case KStop:
			stop = ev.Stop
		case KError:
			t.Fatal(ev.Text)
		}
	}
	if text != "COMPANION_OK" || stop != "stop" {
		t.Fatalf("companion resume = %q %q", text, stop)
	}
	run.ended(req, text, stop, true)
	clean, ok := b.interactiveSession(owner, outer)
	if !ok || clean.Dirty || clean.SessionID != "inner-companion" {
		t.Fatalf("companion did not clean checkpoint: %#v, %v", clean, ok)
	}
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	log := string(logBytes)
	if !strings.Contains(log, "--resume inner-companion") || !strings.Contains(log, companionText) || strings.Contains(log, "SAVED_REPLY") {
		t.Fatalf("companion replayed history or missed resume:\n%s", log)
	}
}

func TestClaudeTUIRuntimeInteractiveResumeCrashLeavesDirtyUntilCompanionRecovers(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "crash-bridge")
	countPath := filepath.Join(dir, "count")
	logPath := filepath.Join(dir, "crash.log")
	script := `#!/bin/sh
prompt=$(cat)
n=0
[ -f "$INTERACTIVE_CRASH_COUNT" ] && n=$(cat "$INTERACTIVE_CRASH_COUNT")
n=$((n+1)); printf '%s' "$n" > "$INTERACTIVE_CRASH_COUNT"
printf '%s\n' '---' >> "$INTERACTIVE_CRASH_LOG"
printf 'args:%s\n' "$*" >> "$INTERACTIVE_CRASH_LOG"
printf 'prompt:%s\n' "$prompt" >> "$INTERACTIVE_CRASH_LOG"
case "$n" in
  1)
    printf '%s\n' '{"type":"assistant","uuid":"row-first","session_id":"inner-crash","message":{"id":"m-first","model":"claude-opus-5-5","content":[{"type":"text","text":"FIRST_OK"}]}}'
    printf '%s\n' '{"type":"result","session_id":"inner-crash","stop_reason":"end_turn"}'
    ;;
  2)
    printf '%s\n' '{"type":"assistant","uuid":"row-partial","session_id":"inner-crash","message":{"id":"m-partial","model":"claude-opus-5-5","content":[{"type":"text","text":"PARTIAL_BEFORE_CRASH"}]}}'
    exit 17
    ;;
  *)
    printf '%s\n' '{"type":"assistant","uuid":"row-recover","session_id":"inner-crash","message":{"id":"m-recover","model":"claude-opus-5-5","content":[{"type":"text","text":"COMPANION_RECOVERED"}]}}'
    printf '%s\n' '{"type":"result","session_id":"inner-crash","stop_reason":"end_turn"}'
    ;;
esac
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_CRASH_COUNT", countPath)
	t.Setenv("INTERACTIVE_CRASH_LOG", logPath)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer, model := "owner", "outer-crash", "claude-opus-5-5"
	collect := func(run *tuiRuntimeSubscriptionRun, events <-chan Event) (text, stop, failed string) {
		t.Helper()
		for ev := range events {
			switch ev.Kind {
			case KText:
				text += ev.Text
			case KStop:
				stop = ev.Stop
			case KError:
				failed = ev.Text
			}
		}
		select {
		case <-run.interactiveDone:
		case <-time.After(2 * time.Second):
			t.Fatal("crash wrapper did not exit")
		}
		return
	}

	req1 := &Request{Model: model, Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}}}}
	run1, ev1, err := b.startInteractiveRequest(context.Background(), req1, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	text1, stop1, failed1 := collect(run1, ev1)
	if text1 != "FIRST_OK" || stop1 != "stop" || failed1 != "" {
		t.Fatalf("first turn = text %q stop %q error %q", text1, stop1, failed1)
	}
	run1.ended(req1, text1, stop1, true)
	saved, ok := b.interactiveSession(owner, outer)
	if !ok || saved.Dirty || saved.SessionID != "inner-crash" {
		t.Fatalf("initial checkpoint = %#v, %v", saved, ok)
	}

	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: text1}}},
	}
	req2 := &Request{Model: model, Messages: append(append([]Message{}, base...), Message{Role: "user", Parts: []Part{{Kind: Text, Text: "trigger crash"}}})}
	run2, ev2, err := b.startInteractiveRequest(context.Background(), req2, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	partial, stop2, failed2 := collect(run2, ev2)
	if partial != "PARTIAL_BEFORE_CRASH" || stop2 != "" || !strings.Contains(failed2, "without a terminal result") {
		t.Fatalf("crashed turn = text %q stop %q error %q", partial, stop2, failed2)
	}
	// relay calls ended(..., ok=false) when a reply has visible partial text
	// plus an error. That must not clean the checkpoint.
	run2.ended(req2, partial, stop2, false)
	dirty, ok := b.interactiveSession(owner, outer)
	if !ok || !dirty.Dirty || dirty.SessionID != "inner-crash" {
		t.Fatalf("crash did not leave dirty checkpoint: %#v, %v", dirty, ok)
	}

	ordinary := &Request{Model: model, Messages: append(append([]Message{}, base...), Message{Role: "user", Parts: []Part{{Kind: Text, Text: "ordinary after crash"}}})}
	if _, _, err := b.startInteractiveRequest(context.Background(), ordinary, model, "", owner, outer); err == nil || !strings.Contains(err.Error(), "dirty") {
		t.Fatalf("ordinary turn did not fail closed on dirty checkpoint: %v", err)
	}
	if raw, err := os.ReadFile(countPath); err != nil || strings.TrimSpace(string(raw)) != "2" {
		t.Fatalf("fail-closed path unexpectedly launched another wrapper: %q %v", raw, err)
	}

	companionText := "Your response above was cut off mid-stream. Resume directly from where it stops — no apology, no recap."
	companion := &Request{Model: model, Messages: append(append([]Message{}, base...), Message{Role: "user", Parts: []Part{{Kind: Text, Text: companionText}}})}
	run3, ev3, err := b.startInteractiveRequest(context.Background(), companion, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	recovered, stop3, failed3 := collect(run3, ev3)
	if recovered != "COMPANION_RECOVERED" || stop3 != "stop" || failed3 != "" {
		t.Fatalf("companion recovery = text %q stop %q error %q", recovered, stop3, failed3)
	}
	run3.ended(companion, recovered, stop3, true)
	clean, ok := b.interactiveSession(owner, outer)
	if !ok || clean.Dirty || clean.SessionID != "inner-crash" {
		t.Fatalf("companion did not clean crash checkpoint: %#v, %v", clean, ok)
	}

	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	turns := strings.Split(string(logBytes), "---\n")
	if len(turns) < 4 {
		t.Fatalf("crash wrapper log missing turns:\n%s", logBytes)
	}
	if !strings.Contains(turns[2], "--resume inner-crash") || !strings.Contains(turns[2], "prompt:trigger crash") ||
		strings.Contains(turns[2], "FIRST_OK") || strings.Contains(turns[2], "prompt:hello") {
		t.Fatalf("crashing resume replayed old history or missed resume:\n%s", turns[2])
	}
	if !strings.Contains(turns[3], "--resume inner-crash") || !strings.Contains(turns[3], companionText) ||
		strings.Contains(turns[3], "FIRST_OK") || strings.Contains(turns[3], "prompt:hello") {
		t.Fatalf("companion recovery replayed old history or missed resume:\n%s", turns[3])
	}
}

func TestClaudeTUIRuntimeInteractiveHistoricalBootstrapImportsTextButNeverReplaysTools(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "bootstrap-bridge")
	logPath := filepath.Join(dir, "bootstrap.log")
	script := `#!/bin/sh
prompt=$(cat)
printf 'args:%s\nprompt:%s\n' "$*" "$prompt" > "$INTERACTIVE_BOOTSTRAP_LOG"
printf '%s\n' '{"type":"assistant","uuid":"row","session_id":"inner-bootstrap","message":{"id":"m","model":"claude-opus-5-5","content":[{"type":"text","text":"BOOT_OK"}]}}'
printf '%s\n' '{"type":"result","session_id":"inner-bootstrap","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_BOOTSTRAP_LOG", logPath)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer, model := "owner", "old-desktop-session", "claude-opus-5-5"
	req := &Request{
		Model: model, System: "CURRENT_SYSTEM", Tools: []Tool{{Name: "current_tool"}},
		Messages: []Message{
			{Role: "user", Parts: []Part{{Kind: Text, Text: "OLD_USER_TEXT"}}},
			{Role: "assistant", Parts: []Part{{Kind: Text, Text: "OLD_ASSISTANT_TEXT"}, {Kind: ToolCall, ID: "old-call", Name: "dangerous_old_tool"}}},
			{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "old-call", Text: "STALE_TOOL_RESULT"}}},
			{Role: "assistant", Parts: []Part{{Kind: Text, Text: "POST_TOOL_VISIBLE_TEXT"}}},
			{Role: "user", Parts: []Part{{Kind: Text, Text: "CURRENT_ONLY"}}},
		},
	}
	run, events, err := b.startInteractiveRequest(context.Background(), req, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	var text, stop string
	for ev := range events {
		switch ev.Kind {
		case KText:
			text += ev.Text
		case KStop:
			stop = ev.Stop
		case KError:
			t.Fatal(ev.Text)
		}
	}
	if text != "BOOT_OK" || stop != "stop" {
		t.Fatalf("historical bootstrap = %q %q", text, stop)
	}
	run.ended(req, text, stop, true)
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	log := string(logBytes)
	for _, want := range []string{
		"<historical_conversation_import>", "OLD_USER_TEXT", "OLD_ASSISTANT_TEXT",
		"POST_TOOL_VISIBLE_TEXT", "<current_user_request>", "CURRENT_ONLY", "CURRENT_SYSTEM",
	} {
		if !strings.Contains(log, want) {
			t.Fatalf("historical bootstrap missing %q:\n%s", want, log)
		}
	}
	for _, stale := range []string{"dangerous_old_tool", "STALE_TOOL_RESULT", "--resume"} {
		if strings.Contains(log, stale) {
			t.Fatalf("historical bootstrap replayed stale work %q:\n%s", stale, log)
		}
	}
	saved, ok := b.interactiveSession(owner, outer)
	if !ok || saved.SessionID != "inner-bootstrap" || saved.Dirty {
		t.Fatalf("historical bootstrap did not establish a clean checkpoint: %#v, %v", saved, ok)
	}
}

func TestClaudeTUIRuntimeInteractiveHistoricalBootstrapRejectsLiveToolContinuation(t *testing.T) {
	req := &Request{Messages: []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "old"}}},
		{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "live", Name: "tool"}}},
		{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "live", Text: "result"}, {Kind: Text, Text: "new text"}}},
	}}
	if _, err := tuiRuntimeRenderInteractiveHistoricalBootstrap(req); err == nil || !strings.Contains(err.Error(), "live tool continuation") {
		t.Fatalf("live tool continuation was imported as history: %v", err)
	}
}

func TestClaudeTUIRuntimeInteractiveHistoricalBootstrapCapsTotalUTF8Bytes(t *testing.T) {
	old := "OLDEST_START" + strings.Repeat("中", 220_000) + "OLDEST_END"
	newest := "RECENT_HISTORY_KEEP_ME"
	current := "当前请求必须完整保留：继续核验，不要丢前半段。"
	req := &Request{Messages: []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: old}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: newest}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: current}}},
	}}
	prompt, err := tuiRuntimeRenderInteractiveHistoricalBootstrap(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(prompt) > tuiRuntimeInteractiveHistoricalBootstrapMaxBytes {
		t.Fatalf("bootstrap is %d bytes, limit %d", len(prompt), tuiRuntimeInteractiveHistoricalBootstrapMaxBytes)
	}
	if !utf8.ValidString(prompt) {
		t.Fatal("byte trimming broke UTF-8")
	}
	if !strings.Contains(prompt, "<current_user_request>\n"+current+"\n</current_user_request>") {
		t.Fatal("current user request was truncated or rewritten")
	}
	if !strings.Contains(prompt, newest) {
		t.Fatal("newest historical turn was not preserved")
	}
	if !strings.Contains(prompt, "[Earlier historical text was omitted to keep the bootstrap within a safe transport budget.]") {
		t.Fatal("trim marker missing")
	}
	if strings.Contains(prompt, "OLDEST_START") {
		t.Fatal("oldest history was not trimmed from the front")
	}
	if !strings.Contains(prompt, "OLDEST_END") {
		t.Fatal("newest tail of an oversized historical turn was not preserved")
	}
}

func TestClaudeTUIRuntimeInteractiveHistoricalBootstrapKeepsFinalUserRoleRunAsCurrentRequest(t *testing.T) {
	req := &Request{Messages: []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "OLD_USER"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "OLD_ASSISTANT"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "CURRENT_PREFIX"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "CURRENT_SUFFIX"}}},
	}}
	prompt, err := tuiRuntimeRenderInteractiveHistoricalBootstrap(req)
	if err != nil {
		t.Fatal(err)
	}
	currentStart := strings.Index(prompt, "<current_user_request>\n")
	currentEnd := strings.Index(prompt, "\n</current_user_request>")
	if currentStart < 0 || currentEnd <= currentStart {
		t.Fatalf("current request block missing:\n%s", prompt)
	}
	current := prompt[currentStart:currentEnd]
	if !strings.Contains(current, "CURRENT_PREFIX") || !strings.Contains(current, "CURRENT_SUFFIX") {
		t.Fatalf("final user role-run was split across history/current:\n%s", prompt)
	}
	history := prompt[:currentStart]
	if strings.Contains(history, "CURRENT_PREFIX") || strings.Contains(history, "CURRENT_SUFFIX") {
		t.Fatalf("current user role-run leaked into historical import:\n%s", prompt)
	}
}

func TestClaudeTUIRuntimeInteractiveImagePromptUsesPrivateMCPTool(t *testing.T) {
	im := Part{Kind: Image, MediaType: "image/png", Data: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB"}
	req := &Request{
		Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "describe image"}, im}}},
		Tools:    []Tool{{Name: "read"}, {Name: tuiRuntimeInteractiveImageTool}},
	}
	if err := tuiRuntimeValidateInteractiveFirstTurn(req); err != nil {
		t.Fatalf("inline image rejected before render: %v", err)
	}
	prompt, images, err := tuiRuntimeRenderInteractiveFirstPromptWithImages(req)
	if err != nil {
		t.Fatal(err)
	}
	id := tuiRuntimeInteractiveImageID(im)
	if !strings.Contains(prompt, id) || !strings.Contains(prompt, "mcp__magpie__"+tuiRuntimeInteractiveImageTool) {
		t.Fatalf("prompt did not name private image tool/id: %q", prompt)
	}
	if strings.Contains(prompt, im.Data) {
		t.Fatal("raw image bytes leaked into the text prompt")
	}
	if got, ok := images[id]; !ok || got.Data != im.Data || got.MediaType != im.MediaType {
		t.Fatalf("image map lost attachment: %#v, %v", got, ok)
	}
	tools := tuiRuntimeInteractiveTools(req, true)
	private, read := 0, 0
	for _, tool := range tools {
		switch tool.Name {
		case tuiRuntimeInteractiveImageTool:
			private++
		case "read":
			read++
		}
	}
	if private != 1 || read != 1 {
		t.Fatalf("private image tool was duplicated or caller tools lost: %#v", tools)
	}
	urlReq := &Request{Messages: []Message{{Role: "user", Parts: []Part{{Kind: Image, URL: "https://example.com/a.png"}}}}}
	if _, _, err := tuiRuntimeRenderInteractiveFirstPromptWithImages(urlReq); err == nil || !strings.Contains(err.Error(), "URL image") {
		t.Fatalf("URL image was silently accepted: %v", err)
	}
}

func TestClaudeTUIRuntimeInteractiveImageMCPReturnsInlineBytes(t *testing.T) {
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	im := Part{Kind: Image, MediaType: "image/png", Data: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB"}
	id := tuiRuntimeInteractiveImageID(im)
	run := &tuiRuntimeSubscriptionRun{
		bridge: b, token: "tok", interactive: true,
		images: map[string]Part{id: im}, pending: map[string]chan tuiRuntimeMcpToolResult{},
	}
	b.runs["tok"] = run
	call := func(imageID string) tuiRuntimeMcpToolResult {
		t.Helper()
		body := fmt.Sprintf(`{"tool_call_id":"call-img","name":%q,"arguments":{"id":%q}}`, tuiRuntimeInteractiveImageTool, imageID)
		req := httptest.NewRequest(http.MethodPost, "/cb/tok", strings.NewReader(body))
		req.SetPathValue("token", "tok")
		rec := httptest.NewRecorder()
		b.mcpCall(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("image MCP status %d: %s", rec.Code, rec.Body.String())
		}
		var got tuiRuntimeMcpToolResult
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		return got
	}
	got := call(id)
	if got.IsError || len(got.Content) != 2 || got.Content[1]["type"] != "image" ||
		got.Content[1]["data"] != im.Data || got.Content[1]["mimeType"] != im.MediaType {
		t.Fatalf("private image MCP returned wrong content: %#v", got)
	}
	missing := call("img_missing")
	if !missing.IsError || len(missing.Content) == 0 {
		t.Fatalf("unknown image id was not rejected: %#v", missing)
	}
	if len(b.calls) != 0 {
		t.Fatalf("private image MCP was registered as a client tool call: %#v", b.calls)
	}
}

func TestClaudeTUIRuntimeInteractiveImageToolIsNotSurfacedToOuterClient(t *testing.T) {
	run := &tuiRuntimeSubscriptionRun{tools: map[string]bool{tuiRuntimeInteractiveImageTool: true}}
	events := run.attach()
	input := strings.Join([]string{
		`{"type":"assistant","uuid":"img-row","session_id":"inner","message":{"id":"m1","model":"claude-opus-5-5","content":[{"type":"tool_use","id":"img-call","name":"mcp__magpie__magpie_view_attached_image","input":{"id":"img_1"}}]}}`,
		`{"type":"assistant","uuid":"text-row","session_id":"inner","message":{"id":"m2","model":"claude-opus-5-5","content":[{"type":"text","text":"IMAGE_SEEN"}]}}`,
		`{"type":"result","session_id":"inner","stop_reason":"end_turn"}`,
	}, "\n")
	go run.readInteractiveOneTurn(strings.NewReader(input))
	var text, stop string
	for ev := range events {
		switch ev.Kind {
		case KToolStart, KToolArgs:
			t.Fatalf("private image tool leaked to outer client: %#v", ev)
		case KText:
			text += ev.Text
		case KStop:
			stop = ev.Stop
		case KError:
			t.Fatal(ev.Text)
		}
	}
	if text != "IMAGE_SEEN" || stop != "stop" {
		t.Fatalf("image-internal turn did not finish normally: %q %q", text, stop)
	}
}

// serveSubscription rewrites run.tools after start returns. The private image
// tool must survive that whole-path handoff; testing startInteractive directly
// would miss the overwrite that broke the real canary.
func TestClaudeTUIRuntimeInteractiveImagePrivateToolSurvivesServeSubscriptionStart(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "image-wrapper")
	script := `#!/bin/sh
cat >/dev/null
# Let serveSubscription finish its post-start bookkeeping before the tool row.
sleep 0.2
printf '%s\n' '{"type":"assistant","uuid":"img-row","session_id":"inner-img","message":{"id":"m1","model":"claude-haiku-4-5","content":[{"type":"tool_use","id":"img-call","name":"mcp__magpie__magpie_view_attached_image","input":{"id":"img_fake"}}]}}'
printf '%s\n' '{"type":"assistant","uuid":"text-row","session_id":"inner-img","message":{"id":"m2","model":"claude-haiku-4-5","content":[{"type":"text","text":"IMAGE_PATH_OK"}]}}'
printf '%s\n' '{"type":"result","session_id":"inner-img","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	s := tuiRuntimeNewTestServer()
	var run *tuiRuntimeSubscriptionRun
	start := func(ctx context.Context, req *Request) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
		var events <-chan Event
		var err error
		run, events, err = s.subscription.startInteractiveOneTurn(ctx, req, "claude-haiku-4-5", "", "owner-img", "")
		return run, events, err
	}
	body := `{"model":"claude-haiku-4-5","max_tokens":64,"messages":[{"role":"user","content":[{"type":"image","source":{"type":"base64","media_type":"image/png","data":"aGVsbG8="}},{"type":"text","text":"describe"}]}]}`
	rec := httptest.NewRecorder()
	code, msg := s.serveSubscription(rec, httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(body)),
		protoAnthropic, "Claude Code", "claude-haiku-4-5", "owner-img", []byte(body), &Usage{}, start)
	if run != nil {
		defer run.abort()
	}
	if code != http.StatusOK || !strings.Contains(rec.Body.String(), "IMAGE_PATH_OK") || strings.Contains(rec.Body.String(), `"type":"tool_use"`) {
		t.Fatalf("private image tool was lost after start: status=%d msg=%q body=%s", code, msg, rec.Body.String())
	}
}

func TestClaudeTUIRuntimeInteractiveWebSearchMCPWaitsForHandlerAndStaysInternal(t *testing.T) {
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	run := &tuiRuntimeSubscriptionRun{
		bridge: b, token: "tok-search", interactive: true,
		searchName: "web_search", tools: map[string]bool{"web_search": true},
		pending: map[string]chan tuiRuntimeMcpToolResult{},
	}
	b.runs[run.token] = run
	events := run.attach()
	body := `{"tool_call_id":"call-search","name":"web_search","arguments":{"query":"latest release"}}`
	req := httptest.NewRequest(http.MethodPost, "/cb/tok-search", strings.NewReader(body))
	req.SetPathValue("token", run.token)
	rec := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		b.mcpCall(rec, req)
		close(done)
	}()
	// Let the callback arrive before serveSubscription's post-start hook.
	time.Sleep(20 * time.Millisecond)
	run.mu.Lock()
	run.search = func(ctx context.Context, query string) (string, []Hit, error) {
		if query != "latest release" {
			t.Fatalf("wrong search query: %q", query)
		}
		return "FOUND", []Hit{{Title: "Release", URL: "https://example.com/release"}}, nil
	}
	run.mu.Unlock()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("search callback did not finish after handler became ready")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("search MCP status %d: %s", rec.Code, rec.Body.String())
	}
	var got tuiRuntimeMcpToolResult
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.IsError || len(got.Content) == 0 || got.Content[0]["text"] != "FOUND" {
		t.Fatalf("internal search returned wrong result: %#v", got)
	}
	select {
	case ev := <-events:
		if ev.Kind != KSearch || ev.Text != "latest release" || len(ev.Hits) != 1 {
			t.Fatalf("search event = %#v", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("internal search did not emit KSearch")
	}
	if len(b.calls) != 0 {
		t.Fatalf("internal search was registered as outer client tool: %#v", b.calls)
	}
}

func TestClaudeTUIRuntimeInteractiveWebSearchToolIsNotSurfacedToOuterClient(t *testing.T) {
	run := &tuiRuntimeSubscriptionRun{
		searchName: "web_search",
		tools:      map[string]bool{"web_search": true},
	}
	events := run.attach()
	input := strings.Join([]string{
		`{"type":"assistant","uuid":"search-row","session_id":"inner","message":{"id":"m1","model":"claude-opus-5-5","content":[{"type":"tool_use","id":"search-call","name":"mcp__magpie__web_search","input":{"query":"x"}}]}}`,
		`{"type":"assistant","uuid":"text-row","session_id":"inner","message":{"id":"m2","model":"claude-opus-5-5","content":[{"type":"text","text":"SEARCH_DONE"}]}}`,
		`{"type":"result","session_id":"inner","stop_reason":"end_turn"}`,
	}, "\n")
	go run.readInteractiveOneTurn(strings.NewReader(input))
	var text, stop string
	for ev := range events {
		switch ev.Kind {
		case KToolStart, KToolArgs:
			t.Fatalf("internal search leaked to outer client: %#v", ev)
		case KText:
			text += ev.Text
		case KStop:
			stop = ev.Stop
		case KError:
			t.Fatal(ev.Text)
		}
	}
	if text != "SEARCH_DONE" || stop != "stop" {
		t.Fatalf("internal search turn did not finish normally: %q %q", text, stop)
	}
}
