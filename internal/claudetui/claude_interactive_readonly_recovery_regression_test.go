package claudetui

// DAV-1868/DAV-1871 regression tests: the independent-review counterexamples
// for the read-only recovery candidate, kept in the product tree so each
// defect stays covered. They were proven RED on the pre-rework candidate and
// must stay GREEN here.

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// D1: a client that goes away mid-reply must not permanently hold the
// per-conversation read-only recovery claim.
func TestD1ClaimLeaksOnClientCancel(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "slow-bridge")
	script := `#!/bin/sh
cat >/dev/null
sleep 2
printf '%s\n' '{"type":"assistant","uuid":"r","session_id":"inner-d1","message":{"id":"m","model":"claude-opus-5-5","content":[{"type":"text","text":"LATE"}]}}'
printf '%s\n' '{"type":"result","session_id":"inner-d1","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	setWrapper(t, wrapper)
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	rt := &runtimeTUI{bridge: b}
	owner, outer, model := "owner", "outer-d1", "claude-opus-5-5"
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "original request"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "answer"}}},
	}
	baseReq := &Request{Model: model, Messages: base}
	keys := conversationKeys(owner, base)
	b.saveInteractiveSession(owner, outer, interactiveSessionEntry{
		SessionID: "inner-d1", Model: model, ThreadKey: firstWords(baseReq),
		ConvKey: keys[0], ReplyKey: assistantReplyKey(base[1]), ContextKey: turnKey(owner, baseReq, nil), Dirty: true,
	})
	body := mergedOrphanBody()
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body)))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	start := func(c context.Context, r *Request) (*subscriptionRun, <-chan Event, error) {
		return b.startInteractiveRequest(c, r, model, "", owner, outer)
	}
	go func() { time.Sleep(50 * time.Millisecond); cancel() }()
	code, _ := rt.serveMessagesInner(anthropicResponder{w, req, &Usage{}}, req, "Claude Code", model, owner, outer, body, start)
	t.Logf("first request code=%d", code)
	// No run is in flight any more; the claim must be free.
	if !b.claimInteractiveOuterTurn(owner, outer) {
		t.Fatal("D1: recovery claim permanently leaked after client cancel; the dirty conversation can never be recovered")
	}
	b.releaseInteractiveOuterTurn(owner, outer)
	b.abortAll()
}

// D2: mcpCall reads run.readOnlyRecovery without holding run.mu while
// serveMessagesInner writes it under run.mu — a data race under -race.
func TestD2ReadOnlyFlagDataRace(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "d2-bridge")
	script := `#!/bin/sh
cat >/dev/null
sleep 1
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	setWrapper(t, wrapper)
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	owner, outer, model := "owner", "outer-d2", "claude-opus-5-5"
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "original request"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "answer"}}},
	}
	baseReq := &Request{Model: model, Messages: base}
	keys := conversationKeys(owner, base)
	b.saveInteractiveSession(owner, outer, interactiveSessionEntry{
		SessionID: "inner-d2", Model: model, ThreadKey: firstWords(baseReq),
		ConvKey: keys[0], ReplyKey: assistantReplyKey(base[1]), ContextKey: turnKey(owner, baseReq, nil), Dirty: true,
	})
	req, err := parseAnthropic(mergedOrphanBody())
	if err != nil {
		t.Fatal(err)
	}
	run, events, err := b.startInteractiveRequest(context.Background(), req, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	// The run is published in b.runs and the subprocess is already running:
	// exactly the window in which serveMessagesInner still has
	// run.readOnlyRecovery == false.
	var stop int32
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { // what serveMessagesInner does right after start() returns
		defer wg.Done()
		deadline := time.Now().Add(300 * time.Millisecond)
		for time.Now().Before(deadline) {
			run.mu.Lock()
			run.readOnlyRecovery = true
			run.mu.Unlock()
			time.Sleep(time.Millisecond)
		}
		atomic.StoreInt32(&stop, 1)
	}()
	go func() { // what the MCP callback goroutine does on every call
		defer wg.Done()
		for atomic.LoadInt32(&stop) == 0 {
			w := httptest.NewRecorder()
			mr := httptest.NewRequest("POST", "/_magpie/claude-mcp/"+run.token,
				strings.NewReader(`{"tool_call_id":"x","name":"write_file","arguments":{}}`))
			mr = mr.WithContext(context.WithValue(mr.Context(), mcpTokenKey{}, run.token))
			b.mcpCall(w, mr)
		}
	}()
	wg.Wait()
	run.abort()
	for range events {
	}
}

// D3: a WebSearch-bearing opt-in recovery request still installs the
// synthetic search handler on the tool-free run (search is computed before
// sanitize and re-installed after start()).
func TestD3SearchHandlerReinstalledOnReadOnlyRun(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "d3-bridge")
	script := `#!/bin/sh
cat >/dev/null
printf '%s\n' '{"type":"assistant","uuid":"r","session_id":"inner-d3","message":{"id":"m","model":"claude-opus-5-5","content":[{"type":"text","text":"D3"}]}}'
printf '%s\n' '{"type":"result","session_id":"inner-d3","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	setWrapper(t, wrapper)
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	rt := &runtimeTUI{bridge: b}
	oldHandler := config.SearchHandler
	config.SearchHandler = func(ctx context.Context, query string) (string, []Hit, error) {
		return "results", nil, nil
	}
	defer func() { config.SearchHandler = oldHandler }()
	owner, outer, model := "owner", "outer-d3", "claude-opus-5-5"
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "original request"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "answer"}}},
	}
	baseReq := &Request{Model: model, Messages: base}
	keys := conversationKeys(owner, base)
	b.saveInteractiveSession(owner, outer, interactiveSessionEntry{
		SessionID: "inner-d3", Model: model, ThreadKey: firstWords(baseReq),
		ConvKey: keys[0], ReplyKey: assistantReplyKey(base[1]), ContextKey: turnKey(owner, baseReq, nil), Dirty: true,
	})
	body := []byte(`{"model":"claude-opus-5-5","tools":[{"type":"web_search_20250305"}],"messages":[` +
		`{"role":"user","content":"original request"},` +
		`{"role":"assistant","content":[{"type":"tool_use","id":"c","name":"write_file","input":{}}]},` +
		`{"role":"user","content":[` +
		`{"type":"tool_result","tool_use_id":"c","content":"x"},` +
		`{"type":"text","text":"/magpie-recover"}]}]}`)
	w := httptest.NewRecorder()
	var started *subscriptionRun
	var startedEvents <-chan Event
	start := func(c context.Context, r *Request) (*subscriptionRun, <-chan Event, error) {
		run, ev, err := b.startInteractiveRequest(c, r, model, "", owner, outer)
		if err == nil {
			started = run
			startedEvents = ev
		}
		return run, ev, err
	}
	code, _ := rt.serveMessagesInner(anthropicResponder{w, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body))), &Usage{}}, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body))), "Claude Code", model, owner, outer, body, start)
	t.Logf("code=%d body=%s", code, w.Body.String())
	run := started
	if run == nil {
		t.Fatal("no run started")
	}
	for range startedEvents { // let the turn finish so the run is idle
	}
	run.mu.Lock()
	searchName := run.searchName
	run.mu.Unlock()
	if searchName != "" {
		t.Fatalf("D3: read-only recovery run carries searchName=%q (search handler reinstalled after sanitize)", searchName)
	}
	run.abort()
}

// D4: the pure-text opt-in shape falls back to the generic bootstrap, which
// treats the whole trailing user run — including older recovery text — as
// the current request instead of only the sentinel.
func TestD4PureTextOptInFallbackLosesSentinelOnly(t *testing.T) {
	req, err := parseAnthropic([]byte(`{"model":"m","messages":[` +
		`{"role":"user","content":"original request"},` +
		`{"role":"assistant","content":"answer"},` +
		`{"role":"user","content":"继续之前没做完的操作：把文件写进去"},` +
		`{"role":"user","content":"/magpie-recover"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if i := interactiveOptInRecoveryIndex(req); i < 0 {
		t.Fatal("pure-text opt-in shape not recognized")
	}
	prompt, err := renderInteractiveReadOnlyBootstrap(req)
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(prompt, "<current_user_request>")
	end := strings.Index(prompt, "</current_user_request>")
	if start < 0 || end < 0 {
		t.Fatalf("no current request block: %q", prompt)
	}
	current := prompt[start+len("<current_user_request>") : end]
	t.Logf("current_user_request=%q", current)
	if strings.Contains(current, "把文件写进去") {
		t.Fatalf("D4: stale user recovery text became the current request: %q", current)
	}
	if strings.TrimSpace(current) != "/magpie-recover" {
		t.Fatalf("D4: current request is not the sentinel alone: %q", current)
	}
}

// Control: a COMPLETED recovery turn must release the claim so the next
// request for the same conversation is not blocked.
func TestD1ControlCompletedTurnReleasesClaim(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "ok-bridge")
	script := `#!/bin/sh
cat >/dev/null
printf '%s\n' '{"type":"assistant","uuid":"r","session_id":"inner-d1c","message":{"id":"m","model":"claude-opus-5-5","content":[{"type":"text","text":"OK"}]}}'
printf '%s\n' '{"type":"result","session_id":"inner-d1c","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	setWrapper(t, wrapper)
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	rt := &runtimeTUI{bridge: b}
	owner, outer, model := "owner", "outer-d1c", "claude-opus-5-5"
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "original request"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "answer"}}},
	}
	baseReq := &Request{Model: model, Messages: base}
	keys := conversationKeys(owner, base)
	b.saveInteractiveSession(owner, outer, interactiveSessionEntry{
		SessionID: "inner-d1c", Model: model, ThreadKey: firstWords(baseReq),
		ConvKey: keys[0], ReplyKey: assistantReplyKey(base[1]), ContextKey: turnKey(owner, baseReq, nil), Dirty: true,
	})
	body := mergedOrphanBody()
	w := httptest.NewRecorder()
	start := func(c context.Context, r *Request) (*subscriptionRun, <-chan Event, error) {
		return b.startInteractiveRequest(c, r, model, "", owner, outer)
	}
	code, _ := rt.serveMessagesInner(anthropicResponder{w, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body))), &Usage{}}, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body))), "Claude Code", model, owner, outer, body, start)
	if code != 200 {
		t.Fatalf("completed recovery code=%d", code)
	}
	if !b.claimInteractiveOuterTurn(owner, outer) {
		t.Fatal("claim not released after a completed turn")
	}
	b.releaseInteractiveOuterTurn(owner, outer)
}
