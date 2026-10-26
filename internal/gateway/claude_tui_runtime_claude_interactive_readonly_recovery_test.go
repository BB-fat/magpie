package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// The production merged shape that produced the incident: a single user
// message holding both a tool_result and new user text. The pre-existing
// recovery predicate rejects any non-Text/Thinking part, so this whole
// shape stays 409 unless the request carries an explicit opt-in sentinel.
func tuiRuntimeMergedOrphanBody() []byte {
	return []byte(`{"model":"claude-opus-5-5","messages":[` +
		`{"role":"user","content":"original request"},` +
		`{"role":"assistant","content":[{"type":"tool_use","id":"failed-call","name":"write_file","input":{"path":"already-written"}}]},` +
		`{"role":"user","content":[` +
		`{"type":"tool_result","tool_use_id":"failed-call","content":"write completed"},` +
		`{"type":"text","text":"also verify the result"},` +
		`{"type":"text","text":"[Request interrupted by user]"},` +
		`{"type":"text","text":"/magpie-recover"}]},` +
		`{"role":"system","content":"<system-reminder>\n<total_tokens>90 tokens left</total_tokens>\n</system-reminder>"}]}`)
}

func TestClaudeTUIRuntimeInteractiveOptInRecoveryIndexRecognizesMergedSentinel(t *testing.T) {
	req, err := parseAnthropic(tuiRuntimeMergedOrphanBody())
	if err != nil {
		t.Fatal(err)
	}
	if i := tuiRuntimeInteractiveOptInRecoveryIndex(req); i < 0 {
		t.Fatal("merged opt-in row was not recognized")
	}
	if i := tuiRuntimeInteractiveReadOnlyOptInIndex(req); i < 0 {
		t.Fatal("read-only recovery index did not accept the merged opt-in")
	}
}

// Negative shapes: substring, tool_result-embedded, history-position and
// sentinel-followed-by-text must all stay rejected.
func TestClaudeTUIRuntimeInteractiveOptInRecoveryIndexRejectsWrongPositions(t *testing.T) {
	cases := map[string]string{
		"substring":           `{"model":"m","messages":[{"role":"user","content":[{"type":"tool_result","tool_use_id":"c","content":"x"},{"type":"text","text":"please /magpie-recover now"}]}]}`,
		"in tool result body": `{"model":"m","messages":[{"role":"user","content":[{"type":"tool_result","tool_use_id":"c","content":"/magpie-recover"},{"type":"text","text":"verify"}]}]}`,
		"older message":       `{"model":"m","messages":[{"role":"user","content":"/magpie-recover"},{"role":"assistant","content":"ok"},{"role":"user","content":[{"type":"tool_result","tool_use_id":"c","content":"x"},{"type":"text","text":"verify"}]}]}`,
		"text after sentinel": `{"model":"m","messages":[{"role":"user","content":[{"type":"tool_result","tool_use_id":"c","content":"x"},{"type":"text","text":"/magpie-recover"},{"type":"text","text":"and also do this"}]}]}`,
		"system only":         `{"model":"m","messages":[{"role":"user","content":[{"type":"tool_result","tool_use_id":"c","content":"x"}]},{"role":"system","content":"/magpie-recover"}]}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			req, err := parseAnthropic([]byte(body))
			if err != nil {
				t.Fatal(err)
			}
			if i := tuiRuntimeInteractiveOptInRecoveryIndex(req); i >= 0 {
				t.Fatalf("wrong-position sentinel accepted at %d", i)
			}
			if i := tuiRuntimeInteractiveReadOnlyOptInIndex(req); i >= 0 {
				t.Fatalf("read-only index accepted %s", name)
			}
		})
	}
}

// A merged retry that does NOT carry the sentinel must still be refused at
// the HTTP gate with no run started — the existing conservative guarantee.
func TestClaudeTUIRuntimeInteractiveMergedRetryWithoutSentinelStaysRejected(t *testing.T) {
	body := []byte(`{"model":"claude-opus-5-5","messages":[` +
		`{"role":"user","content":"original request"},` +
		`{"role":"assistant","content":[{"type":"tool_use","id":"failed-call","name":"write_file","input":{"path":"already-written"}}]},` +
		`{"role":"user","content":[` +
		`{"type":"tool_result","tool_use_id":"failed-call","content":"write completed"},` +
		`{"type":"text","text":"also verify the result"}]}]}`)
	s := tuiRuntimeNewTestServer()
	defer tuiRuntimeBridges.Delete(s.subscription)
	started := 0
	start := func(context.Context, *Request) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
		started++
		return nil, nil, errors.New("must not start")
	}
	for attempt := 0; attempt < 2; attempt++ {
		w := httptest.NewRecorder()
		code, _ := s.serveSubscription(w, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body))), protoAnthropic, "Claude Code", "claude-opus-5-5", "owner", body, &Usage{}, start)
		if code != http.StatusConflict || started != 0 {
			t.Fatalf("attempt %d: status=%d starts=%d", attempt, code, started)
		}
	}
}

// startInteractiveRequest on a dirty session must re-check the opt-in and
// the persisted owner/model itself; a non-dirty checkpoint must never take
// this path.
func TestClaudeTUIRuntimeInteractiveReadOnlyStartRechecksDirtyAndModel(t *testing.T) {
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer, model := "owner", "outer-ro", "claude-opus-5-5"
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "original request"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "answer"}}},
	}
	baseReq := &Request{Model: model, Messages: base}
	keys := tuiRuntimeConversationKeys(owner, base)

	// Not dirty: the opt-in request must not reach the recovery branch.
	b.saveInteractiveSession(owner, outer, tuiRuntimeInteractiveSessionEntry{
		SessionID: "inner", Model: model, ThreadKey: firstWords(baseReq),
		ConvKey: keys[0], ReplyKey: tuiRuntimeAssistantReplyKey(base[1]), ContextKey: tuiRuntimeTurnKey(owner, baseReq, nil),
	})
	req, _ := parseAnthropic(tuiRuntimeMergedOrphanBody())
	if _, _, err := b.startInteractiveRequest(context.Background(), req, model, "", owner, outer); err == nil ||
		!strings.Contains(err.Error(), "refusing full history replay") && !strings.Contains(err.Error(), "checkpoint") {
		t.Fatalf("non-dirty checkpoint took the opt-in path: %v", err)
	}

	// Dirty + wrong model: refused before any wrapper could start.
	b.markInteractiveSessionDirty(owner, outer)
	if _, _, err := b.startInteractiveRequest(context.Background(), req, "claude-haiku-4-5", "", owner, outer); err == nil ||
		!strings.Contains(err.Error(), "model_mismatch") {
		t.Fatalf("model-mismatched dirty recovery was not refused: %v", err)
	}
}

// The recovery turn must be structurally tool-free: tools.json empty, no
// image helper, --tools empty, and any tool signal fails the turn.
func TestClaudeTUIRuntimeInteractiveReadOnlyRecoveryTurnIsToolFree(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	var gotArgs, gotToolsJSON, gotPrompt string
	argsPath := filepath.Join(dir, "args")
	toolsPath := filepath.Join(dir, "tools")
	promptPath := filepath.Join(dir, "prompt")
	wrapper := filepath.Join(dir, "fake-bridge")
	script := `#!/bin/sh
printf '%s\n' "$@" > "$INTERACTIVE_RO_ARGS"
while [ $# -gt 0 ]; do
  if [ "$1" = "--mcp-config" ]; then shift; printf '%s' "$1" > "$INTERACTIVE_RO_MCP"; fi
  shift
done
# Record the tools.json the helper would have been given before the bridge
# reaps run.tmp on exit: on Linux the child is reaped fast enough that the
# file is gone by the time the test asserts, so snapshot it inside the
# wrapper's own lifetime. The bridge writes it next to user-context.json,
# whose path is in our env.
snap=$(dirname "$CLAUDE_BRIDGE_USER_CONTEXT_PATH")/tools.json
[ -f "$snap" ] && cp "$snap" "$INTERACTIVE_RO_TOOLSNAP"
cat > "$INTERACTIVE_RO_PROMPT"
printf '%s\n' '{"type":"assistant","uuid":"r1","session_id":"inner-ro","message":{"id":"m","model":"claude-opus-5-5","content":[{"type":"text","text":"READ_ONLY_OK"}]}}'
printf '%s\n' '{"type":"result","session_id":"inner-ro","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_RO_ARGS", argsPath)
	t.Setenv("INTERACTIVE_RO_PROMPT", promptPath)
	t.Setenv("INTERACTIVE_RO_MCP", toolsPath)
	t.Setenv("INTERACTIVE_RO_TOOLSNAP", filepath.Join(dir, "tools-snap"))
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer, model := "owner", "outer-ro2", "claude-opus-5-5"
	defer tuiRuntimeStopBridge(t, b)
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "original request"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "answer"}}},
	}
	baseReq := &Request{Model: model, Tools: []Tool{{Name: "write_file"}}, Messages: base}
	keys := tuiRuntimeConversationKeys(owner, base)
	b.saveInteractiveSession(owner, outer, tuiRuntimeInteractiveSessionEntry{
		SessionID: "inner-ro", Model: model, ThreadKey: firstWords(baseReq),
		ConvKey: keys[0], ReplyKey: tuiRuntimeAssistantReplyKey(base[1]), ContextKey: tuiRuntimeTurnKey(owner, baseReq, nil), Dirty: true,
	})

	// Request carries client tools + a tool result + opt-in sentinel.
	req, err := parseAnthropic([]byte(`{"model":"claude-opus-5-5","tools":[{"name":"write_file","description":"w","input_schema":{"type":"object"}}],"messages":[` +
		`{"role":"user","content":"original request"},` +
		`{"role":"assistant","content":[{"type":"tool_use","id":"failed-call","name":"write_file","input":{"path":"p"}}]},` +
		`{"role":"user","content":[` +
		`{"type":"tool_result","tool_use_id":"failed-call","content":"write completed"},` +
		`{"type":"text","text":"/magpie-recover"}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	run, events, err := b.startInteractiveRequest(context.Background(), req, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	// The constructor marked this as a read-only recovery run before
	// publishing it; nothing else needs to set the flag.
	if !run.readOnlyRecovery {
		t.Fatal("recovery run was not marked read-only at construction")
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
	if text != "READ_ONLY_OK" || stop != "stop" {
		t.Fatalf("recovery turn = %q %q", text, stop)
	}
	tuiRuntimeWaitStopped(t, b)
	raw, _ := os.ReadFile(argsPath)
	gotArgs = string(raw)
	if !strings.Contains(gotArgs, "--tools") {
		t.Fatalf("wrapper argv lacks --tools: %q", gotArgs)
	}
	raw, _ = os.ReadFile(toolsPath)
	gotToolsJSON = string(raw)
	// The mcp-config still registers the magpie helper, but the tools.json
	// it hands it must be empty for a tool-free turn.
	if !strings.Contains(gotToolsJSON, `"mcpServers"`) {
		t.Fatalf("mcp config missing helper registration: %q", gotToolsJSON)
	}
	// tools.json is snapshotted by the wrapper itself (inside its own tmp
	// lifetime) because removeRun reaps run.tmp too fast on Linux for the
	// test to read it afterwards.
	raw, err = os.ReadFile(filepath.Join(dir, "tools-snap"))
	if err != nil {
		t.Fatalf("no tools.json snapshot captured by the wrapper: %v", err)
	}
	gotToolsJSON = string(raw)
	if strings.TrimSpace(gotToolsJSON) != "[]" && strings.TrimSpace(gotToolsJSON) != "null" {
		t.Fatalf("tools.json was not empty: %q", gotToolsJSON)
	}
	raw, _ = os.ReadFile(promptPath)
	gotPrompt = string(raw)
	if !strings.Contains(gotPrompt, "/magpie-recover") {
		t.Fatalf("sentinel did not become the current request: %q", gotPrompt)
	}
	if strings.Contains(gotPrompt, "write completed") {
		t.Fatalf("tool result body leaked into the prompt: %q", gotPrompt)
	}
	if strings.Contains(gotPrompt, "You must call at least one available tool") {
		t.Fatalf("required-tool instruction leaked into a tool-free prompt")
	}
	// run.tools must be empty even though req carried write_file.
	run.mu.Lock()
	tools := len(run.tools)
	run.mu.Unlock()
	if tools != 0 {
		t.Fatalf("read-only run registered %d tools", tools)
	}
	// An MCP callback on a read-only run is refused and registers nothing.
	// Use a wrapper that stays alive so the run is still in b.runs when the
	// callback arrives.
	wrapper2 := filepath.Join(dir, "fake-bridge-live")
	script2 := `#!/bin/sh
cat >/dev/null
printf '%s\n' '{"type":"assistant","uuid":"r","session_id":"inner-live","message":{"id":"m","model":"claude-opus-5-5","content":[{"type":"text","text":"LIVE"}]}}'
sleep 5
`
	if err := os.WriteFile(wrapper2, []byte(script2), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper2)
	req2, err := parseAnthropic([]byte(`{"model":"claude-opus-5-5","messages":[` +
		`{"role":"user","content":"original request"},` +
		`{"role":"assistant","content":[{"type":"tool_use","id":"c","name":"write_file","input":{}}]},` +
		`{"role":"user","content":[{"type":"tool_result","tool_use_id":"c","content":"x"},{"type":"text","text":"/magpie-recover"}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	run2, ev2, err := b.startInteractiveRequest(context.Background(), req2, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	if !run2.readOnlyRecovery {
		t.Fatal("live recovery run was not marked read-only at construction")
	}
	w := httptest.NewRecorder()
	mr := httptest.NewRequest("POST", "/_magpie/claude-mcp/"+run2.token, strings.NewReader(`{"tool_call_id":"x","name":"write_file","arguments":{}}`))
	mr = mr.WithContext(context.WithValue(mr.Context(), tuiRuntimeMcpTokenKey{}, run2.token))
	b.mcpCall(w, mr)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "read-only recovery") {
		t.Fatalf("tool callback was not refused: %d %q", w.Code, w.Body.String())
	}
	b.mu.Lock()
	registered := len(b.calls)
	b.mu.Unlock()
	if registered != 0 {
		t.Fatalf("refused callback still registered %d calls", registered)
	}
	if !run2.readOnlyToolSignaled() {
		t.Fatal("refused callback was not recorded as a tool signal")
	}
	run2.abort()
	for range ev2 {
	}
}

// Two concurrent opt-in requests on the same dirty outer session must
// produce exactly one started run; the second gets 409.
func TestClaudeTUIRuntimeInteractiveReadOnlyRecoveryConcurrentOneRun(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "concurrent-bridge")
	script := `#!/bin/sh
cat >/dev/null
printf '%s\n' '{"type":"assistant","uuid":"r","session_id":"inner-cc","message":{"id":"m","model":"claude-opus-5-5","content":[{"type":"text","text":"ONE"}]}}'
sleep 3
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	rt := &tuiRuntime{bridge: b}
	owner, outer, model := "owner", "outer-cc", "claude-opus-5-5"
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "original request"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "answer"}}},
	}
	baseReq := &Request{Model: model, Messages: base}
	keys := tuiRuntimeConversationKeys(owner, base)
	b.saveInteractiveSession(owner, outer, tuiRuntimeInteractiveSessionEntry{
		SessionID: "inner-cc", Model: model, ThreadKey: firstWords(baseReq),
		ConvKey: keys[0], ReplyKey: tuiRuntimeAssistantReplyKey(base[1]), ContextKey: tuiRuntimeTurnKey(owner, baseReq, nil), Dirty: true,
	})
	body := tuiRuntimeMergedOrphanBody()
	var started int32
	var conflicts int32
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			start := func(ctx context.Context, req *Request) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
				atomic.AddInt32(&started, 1)
				return b.startInteractiveRequest(ctx, req, model, "", owner, outer)
			}
			code, _ := rt.serveMessagesInner(tuiRuntimeAnthropicResponder{w, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body))), &Usage{}}, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body))), "Claude Code", model, owner, outer, body, start)
			if code == http.StatusConflict {
				atomic.AddInt32(&conflicts, 1)
			}
		}()
	}
	wg.Wait()
	if started != 1 || conflicts != 1 {
		t.Fatalf("concurrent opt-ins: starts=%d conflicts=%d, want 1/1", started, conflicts)
	}
	b.abortAll()
}

// A tool signal on the recovery transcript turns the whole recovery into a
// failure: the checkpoint stays dirty and no fresh checkpoint is saved.
func TestClaudeTUIRuntimeInteractiveReadOnlyToolSignalFailsTurnKeepsDirty(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "signal-bridge")
	script := `#!/bin/sh
cat >/dev/null
printf '%s\n' '{"type":"assistant","uuid":"r1","session_id":"inner-sig","message":{"id":"m","model":"claude-opus-5-5","content":[{"type":"tool_use","id":"t1","name":"mcp__magpie__write_file","input":{"path":"p"}}]}}'
printf '%s\n' '{"type":"result","session_id":"inner-sig","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer, model := "owner", "outer-sig", "claude-opus-5-5"
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "original request"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "answer"}}},
	}
	baseReq := &Request{Model: model, Messages: base}
	keys := tuiRuntimeConversationKeys(owner, base)
	b.saveInteractiveSession(owner, outer, tuiRuntimeInteractiveSessionEntry{
		SessionID: "inner-sig", Model: model, ThreadKey: firstWords(baseReq),
		ConvKey: keys[0], ReplyKey: tuiRuntimeAssistantReplyKey(base[1]), ContextKey: tuiRuntimeTurnKey(owner, baseReq, nil), Dirty: true,
	})
	req, err := parseAnthropic(tuiRuntimeMergedOrphanBody())
	if err != nil {
		t.Fatal(err)
	}
	run, events, err := b.startInteractiveRequest(context.Background(), req, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	var stop, failed string
	for ev := range events {
		switch ev.Kind {
		case KStop:
			stop = ev.Stop
		case KError:
			failed = ev.Text
		}
	}
	// The tool_use row was noted as a signal, not delivered as a tool call.
	if !run.readOnlyToolSignaled() {
		t.Fatal("transcript tool_use row was not recorded as a signal")
	}
	// ended() with a "stop" that carried a tool signal must abort, keeping
	// the checkpoint dirty.
	run.ended(req, "x", stop, failed == "")
	saved, ok := b.interactiveSession(owner, outer)
	if !ok || !saved.Dirty {
		t.Fatalf("tool signal turn wrote a clean checkpoint: %#v %v", saved, ok)
	}
}

// GREEN counterpart to the frozen-baseline RED: the same opt-in merged
// request on a dirty session now drives a real tool-free recovery turn —
// the request reaches start(), the wrapper launches with --tools empty and
// an empty tools.json, and the reply completes.
func TestClaudeTUIRuntimeGREENOptInMergedDrivesToolFreeTurn(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	argsPath := filepath.Join(dir, "args")
	promptPath := filepath.Join(dir, "prompt")
	wrapper := filepath.Join(dir, "green-bridge")
	script := `#!/bin/sh
printf '%s\n' "$@" > "$GREEN_ARGS"
cat > "$GREEN_PROMPT"
printf '%s\n' '{"type":"assistant","uuid":"r","session_id":"inner-green","message":{"id":"m","model":"claude-opus-5-5","content":[{"type":"text","text":"GREEN_OK"}]}}'
printf '%s\n' '{"type":"result","session_id":"inner-green","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("GREEN_ARGS", argsPath)
	t.Setenv("GREEN_PROMPT", promptPath)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	rt := &tuiRuntime{bridge: b}
	owner, outer, model := "owner", "outer-green", "claude-opus-5-5"
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "original request"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "answer"}}},
	}
	baseReq := &Request{Model: model, Messages: base}
	keys := tuiRuntimeConversationKeys(owner, base)
	b.saveInteractiveSession(owner, outer, tuiRuntimeInteractiveSessionEntry{
		SessionID: "inner-green", Model: model, ThreadKey: firstWords(baseReq),
		ConvKey: keys[0], ReplyKey: tuiRuntimeAssistantReplyKey(base[1]), ContextKey: tuiRuntimeTurnKey(owner, baseReq, nil), Dirty: true,
	})
	var started int
	start := func(ctx context.Context, req *Request) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
		started++
		return b.startInteractiveRequest(ctx, req, model, "", owner, outer)
	}
	w := httptest.NewRecorder()
	code, _ := rt.serveMessagesInner(tuiRuntimeAnthropicResponder{w, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(tuiRuntimeMergedOrphanBody()))), &Usage{}}, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(tuiRuntimeMergedOrphanBody()))), "Claude Code", model, owner, outer, tuiRuntimeMergedOrphanBody(), start)
	if code != 200 || started != 1 {
		t.Fatalf("opt-in merged did not drive one recovery turn: code=%d starts=%d", code, started)
	}
	args, _ := os.ReadFile(argsPath)
	if !strings.Contains(string(args), "--tools") {
		t.Fatalf("wrapper argv lacks --tools: %q", args)
	}
	prompt, _ := os.ReadFile(promptPath)
	if !strings.Contains(string(prompt), "/magpie-recover") {
		t.Fatalf("sentinel not the current request: %q", prompt)
	}
	if strings.Contains(string(prompt), "write completed") {
		t.Fatalf("tool result body leaked into prompt: %q", prompt)
	}
}
