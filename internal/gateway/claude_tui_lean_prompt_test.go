package gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/provider"
)

// tuiLeanWrapper is a stand-in for the PTY wrapper that keeps, beside it,
// the arguments of each run, the prompt it was given, and the system
// prompt file it was given, copied while it ran, as the file goes with the
// run. Each run's files are numbered from 1.
func tuiLeanWrapper(t *testing.T, rows string) (string, func(n int) (args, prompt, system string, hasSys bool)) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "wrapper")
	script := `#!/bin/sh
d=` + dir + `
n=$(( $(cat "$d/count" 2>/dev/null || echo 0) + 1 ))
echo $n > "$d/count"
printf '%s\n' "$@" > "$d/args.$n"
for a in "$@"; do
  case "$a" in --system-prompt-file=*) cp "${a#--system-prompt-file=}" "$d/system.$n" ;; esac
done
cat > "$d/prompt.$n"
`
	for _, l := range strings.Split(rows, "\n") {
		script += "printf '%s\\n' '" + l + "'\n"
	}
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return path, func(n int) (string, string, string, bool) {
		t.Helper()
		read := func(name string) (string, bool) {
			b, err := os.ReadFile(filepath.Join(dir, name+"."+strconv.Itoa(n)))
			return string(b), err == nil
		}
		args, ok := read("args")
		if !ok {
			t.Fatalf("run %d never started", n)
		}
		prompt, _ := read("prompt")
		system, hasSys := read("system")
		return args, prompt, system, hasSys
	}
}

// With Settings' lean prompt on, a TUI account's run is started with the
// caller's own instructions as its system prompt, in place of Claude
// Code's: in a file, without the client Claude Code's billing line, and
// its first message holds no instructions; its tool choice is still said
// with the turn. Off, the instructions are in the first message, as
// before.
func TestClaudeTUILeanPromptIsTheRunsSystemPrompt(t *testing.T) {
	wrapper, run := tuiLeanWrapper(t, tuiTextRows)
	p := tuiAccount(t, wrapper)
	tuiRuntimeSettings.LeanPrompt = tuiConfig().LeanPrompt
	s := tuiTestServer(t)
	body := `{"model":"claude-sonnet-5","max_tokens":64,` +
		`"system":"x-anthropic-billing-header: cc_version=2.1.300; cch=abc;\nLEAN_RULES stay short.",` +
		`"tools":[{"name":"echo","input_schema":{"type":"object"}}],"tool_choice":{"type":"any"},` +
		`"messages":[{"role":"user","content":"hi"}]}`
	post := func() {
		t.Helper()
		rec := httptest.NewRecorder()
		if code, msg := s.serveClaudeSubscription(rec, httptest.NewRequest("POST", "/", strings.NewReader(body)), provider.Anthropic, p, "claude-sonnet-5", []byte(body), &Usage{}); code != 200 {
			t.Fatalf("%d %s %s", code, msg, rec.Body)
		}
	}

	leanSetting(t, true)
	post()
	args, prompt, system, hasSys := run(1)
	if !hasSys || system != "LEAN_RULES stay short." {
		t.Fatalf("lean run's system prompt = %q (given: %t), args:\n%s", system, hasSys, args)
	}
	if strings.Contains(prompt, "LEAN_RULES") || strings.Contains(prompt, "external_system_instructions") || strings.Contains(prompt, "billing-header") {
		t.Fatalf("lean run's first message holds the instructions:\n%s", prompt)
	}
	if !strings.Contains(prompt, "You must call at least one available tool") || !strings.Contains(prompt, "hi") {
		t.Fatalf("lean run's first message lost its turn:\n%s", prompt)
	}
	// Desktop's suggestion is answered by a run of its own, without the
	// caller's instructions, and lean too: its system prompt is empty
	suggestion := `{"model":"claude-sonnet-5","max_tokens":64,"system":"LEAN_RULES stay short.","messages":[` +
		`{"role":"user","content":"hi"},{"role":"assistant","content":"TUI_OK"},` +
		`{"role":"user","content":"[SUGGESTION MODE: suggest next input]"}]}`
	rec := httptest.NewRecorder()
	if code, msg := s.serveClaudeSubscription(rec, httptest.NewRequest("POST", "/", strings.NewReader(suggestion)), provider.Anthropic, p, "claude-sonnet-5", []byte(suggestion), &Usage{}); code != 200 {
		t.Fatalf("suggestion: %d %s %s", code, msg, rec.Body)
	}
	args, prompt, system, hasSys = run(2)
	if !hasSys || system != "" || strings.Contains(prompt, "LEAN_RULES") || !strings.Contains(prompt, "SUGGESTION MODE") {
		t.Fatalf("lean suggestion run's system prompt = %q (given: %t), args:\n%s\nprompt:\n%s", system, hasSys, args, prompt)
	}

	leanSetting(t, false)
	post()
	args, prompt, _, hasSys = run(3)
	if hasSys || strings.Contains(args, "--system-prompt") {
		t.Fatalf("a run with the setting off was given a system prompt:\n%s", args)
	}
	if !strings.Contains(prompt, "<external_system_instructions>\nLEAN_RULES stay short.") {
		t.Fatalf("a run with the setting off lost its instructions:\n%s", prompt)
	}
}

// A conversation's native session goes on lean while the setting stays on,
// given its system prompt again as it resumes. Once the setting changes it
// is not resumed: its instructions are where the setting no longer puts
// them, and the next turn is rebuilt from the request.
func TestClaudeTUILeanPromptRunNotResumedAcrossTheSetting(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	rows := `{"type":"assistant","uuid":"row","session_id":"inner-main","message":{"id":"m","model":"claude-opus-5-5","usage":{"input_tokens":3,"output_tokens":2},"content":[{"type":"text","text":"REPLY"}]}}` + "\n" +
		`{"type":"result","session_id":"inner-main","stop_reason":"end_turn"}`
	wrapper, run := tuiLeanWrapper(t, rows)
	tuiRuntimeSetWrapper(t, wrapper)
	owner, outer, model := "owner", "outer", "claude-opus-5-5"
	msgs := []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}}}
	turn := func(lean bool) {
		t.Helper()
		// a bridge each turn, as after a gateway restart: the checkpoint
		// alone decides whether the session is resumed
		b := tuiRuntimeNewSubscriptionBridge()
		defer tuiRuntimeBridges.Delete(b)
		req := &tuiRuntimeRequest{Request: Request{Model: model, System: "RULES", Messages: msgs}}
		req.leanPrompt = lean
		r, events, err := b.startInteractiveRequest(context.Background(), req, model, "", owner, outer)
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
				t.Fatalf("turn failed: %s", ev.Text)
			}
		}
		select {
		case <-r.interactiveDone:
		case <-time.After(2 * time.Second):
			t.Fatal("wrapper did not exit")
		}
		r.ended(req, text, stop, true)
		msgs = append(msgs, Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: text}}},
			Message{Role: "user", Parts: []Part{{Kind: Text, Text: "next"}}})
	}

	turn(true)
	turn(true)
	args, prompt, system, hasSys := run(2)
	if !strings.Contains(args, "--resume\ninner-main\n") {
		t.Fatalf("lean session not resumed with the setting unchanged:\n%s", args)
	}
	if !hasSys || system != "RULES" || strings.Contains(prompt, "RULES") {
		t.Fatalf("resumed lean run's system prompt = %q (given: %t), prompt:\n%s", system, hasSys, prompt)
	}

	turn(false)
	args, prompt, _, hasSys = run(3)
	if strings.Contains(args, "inner-main") || hasSys {
		t.Fatalf("lean session resumed with the setting off:\n%s", args)
	}
	if !strings.Contains(prompt, "<external_system_instructions>\nRULES") {
		t.Fatalf("rebuilt run lost its instructions:\n%s", prompt)
	}
}

// A bootstrap prompt of a lean run (the history imported into a new native
// session, and a read-only recovery) says the current notices but neither
// the instructions, which are its system prompt, nor that there are none.
func TestClaudeTUILeanPromptBootstrapsWithoutInstructions(t *testing.T) {
	history := &tuiRuntimeRequest{Request: Request{System: "RULES", Messages: []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "OLD_USER"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "OLD_ASSISTANT"}}},
		{Role: "system", Parts: []Part{{Kind: Text, Text: "CURRENT_NOTICE"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "CURRENT"}}},
	}}}
	quiet := *history
	quiet.Messages = slices.DeleteFunc(slices.Clone(history.Messages), func(m Message) bool { return m.Role == "system" })
	orphan, err := tuiRuntimeParseRequest(tuiRuntimeMergedOrphanBody())
	if err != nil {
		t.Fatal(err)
	}
	orphan.System = "RULES"
	for _, c := range []struct {
		name   string
		req    *tuiRuntimeRequest
		render func(*tuiRuntimeRequest) (string, error)
		notice bool
	}{
		{"historical", history, tuiRuntimeRenderInteractiveHistoricalBootstrap, true},
		{"historical without notices", &quiet, tuiRuntimeRenderInteractiveHistoricalBootstrap, false},
		{"read-only", orphan, tuiRuntimeRenderInteractiveReadOnlyBootstrap, false},
	} {
		for _, lean := range []bool{false, true} {
			req := *c.req
			req.leanPrompt = lean
			prompt, err := c.render(&req)
			if err != nil {
				t.Fatalf("%s lean=%t: %v", c.name, lean, err)
			}
			if got := strings.Contains(prompt, "RULES"); got == lean {
				t.Errorf("%s lean=%t: instructions in the prompt: %t\n%s", c.name, lean, got, prompt)
			}
			if lean && strings.Contains(prompt, "There are no current external system instructions") {
				t.Errorf("%s: a lean prompt says it has no instructions:\n%s", c.name, prompt)
			}
			if c.notice && !strings.Contains(prompt, "<external_system_instructions>") || c.notice && !strings.Contains(prompt, "CURRENT_NOTICE") {
				t.Errorf("%s lean=%t: the current notice is lost:\n%s", c.name, lean, prompt)
			}
		}
	}
}

// A run started lean is told apart from one started with Claude Code's
// prompt under the same instructions, and the keys of runs not lean are the
// ones they were before.
func TestClaudeTUILeanPromptKeys(t *testing.T) {
	req := &tuiRuntimeRequest{Request: Request{Model: "m", System: "RULES", Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "hi"}}}}}}
	lean := *req
	lean.leanPrompt = true
	if tuiRuntimeInstructionKey(req) == tuiRuntimeInstructionKey(&lean) {
		t.Error("instruction key ignores the lean prompt")
	}
	if tuiRuntimeTurnKey("o", req, req.Messages) == tuiRuntimeTurnKey("o", &lean, lean.Messages) {
		t.Error("turn key ignores the lean prompt")
	}
	// as b8d1090b computed them, which checkpoints saved before hold
	if got := tuiRuntimeInstructionKey(req); got != "c006fdaadc23331422cb860ac163da782ffd30bcb72c7df38cb2d108735af4d5" {
		t.Errorf("instruction key of a run not lean changed: %s", got)
	}
	if got := tuiRuntimeTurnKey("o", req, req.Messages); got != "bf8848d2aedac6f41acbe5c5c2d42e4b17b582189e3070c225e878a4ce8b8f0d" {
		t.Errorf("turn key of a run not lean changed: %s", got)
	}
}

// A run waiting on a tool's result when the setting changes finishes its
// turn in the mode it was started in: the result is its own, by its call's
// id, and the setting is not "settings changed while a PTY tool was running"
// (409, the turn lost), nor a reason to let the run go and branch its
// session. Its checkpoint keeps that mode, so the conversation's next turn
// is rebuilt in the new one. Both the gateway's path (serveLite) and the
// handler under it (serveMessagesInner) are held to it.
func TestClaudeTUILeanPromptChangeKeepsARunningToolTurn(t *testing.T) {
	for _, path := range []string{"lite", "inner"} {
		for _, from := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s-from-lean-%t", path, from), func(t *testing.T) {
				tuiLeanToolTurnAcrossTheSetting(t, path, from)
			})
		}
	}
}

func tuiLeanToolTurnAcrossTheSetting(t *testing.T, path string, from bool) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	release := filepath.Join(dir, "tool-result-received")
	wrapper := filepath.Join(dir, "mock-pty")
	data := `#!/bin/sh
cat >/dev/null
printf '%s\n' '{"type":"assistant","uuid":"tool-row","session_id":"inner-L","message":{"id":"first","model":"claude-haiku-4-5","usage":{"input_tokens":11},"content":[{"type":"tool_use","id":"call_1","name":"mcp__magpie__echo_test","input":{"text":"HELLO"}}]}}'
for i in $(seq 1 200); do
  if [ -f "$LEAN_TEST_RELEASE" ]; then break; fi
  sleep 0.05
done
if [ ! -f "$LEAN_TEST_RELEASE" ]; then exit 9; fi
printf '%s\n' '{"type":"assistant","uuid":"answer-row","session_id":"inner-L","message":{"id":"second","model":"claude-haiku-4-5","usage":{"input_tokens":4,"output_tokens":5},"content":[{"type":"text","text":"LEAN_TOOL_TURN_OK"}]}}'
printf '%s\n' '{"type":"result","session_id":"inner-L","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(data), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("LEAN_TEST_RELEASE", release)
	lean := from
	old := tuiRuntimeSettings
	cfg := *old
	cfg.LeanPrompt = func() bool { return lean }
	tuiRuntimeSettings = &cfg
	t.Cleanup(func() { tuiRuntimeSettings = old })

	s := tuiRuntimeNewTestServer()
	owner, outer, model := "owner-L", "outer-L", "claude-haiku-4-5"
	if path == "lite" {
		owner = tuiRuntimeTestOwner
	}
	start := func(ctx context.Context, req *tuiRuntimeRequest) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
		return s.subscription.startInteractiveOneTurn(ctx, req, model, "", owner, outer)
	}
	post := func(body string) (int, string, string) {
		t.Helper()
		rec := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
		var code int
		var msg string
		if path == "lite" {
			code, msg = s.rt.serveMessages(rec, r, model, outer, []byte(body), &Usage{})
		} else {
			code, msg = s.serveSubscription(rec, r, protoAnthropic, "Claude Code", model, owner, []byte(body), &Usage{}, start)
		}
		return code, msg, rec.Body.String()
	}
	first := `{"model":"claude-haiku-4-5","max_tokens":100,"system":"RULES","tools":[{"name":"echo_test","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"Call echo_test."}]}`
	if code, msg, body := post(first); code != 200 || !strings.Contains(body, `"id":"call_1"`) {
		t.Fatalf("first turn did not yield the call: %d %s %s", code, msg, body)
	}
	var run *tuiRuntimeSubscriptionRun
	s.subscription.mu.Lock()
	for _, r := range s.subscription.runs {
		run = r
	}
	s.subscription.mu.Unlock()
	if run == nil {
		t.Fatal("no run")
	}
	defer run.abort()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /cb/{token}", s.subscription.mcpCall)
	go func() {
		body := `{"tool_call_id":"call_1","name":"echo_test","arguments":{"text":"HELLO"}}`
		mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", "/cb/"+run.token, strings.NewReader(body)))
		_ = os.WriteFile(release, []byte("ok"), 0o600)
	}()
	waitFor(t, "the MCP callback registers its call", func() bool {
		s.subscription.mu.Lock()
		defer s.subscription.mu.Unlock()
		return s.subscription.calls["call_1"] == run
	})

	lean = !from // the user turns the setting over while the tool runs
	second := `{"model":"claude-haiku-4-5","max_tokens":100,"system":"RULES","tools":[{"name":"echo_test","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"Call echo_test."},{"role":"assistant","content":[{"type":"tool_use","id":"call_1","name":"echo_test","input":{"text":"HELLO"}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call_1","content":"HELLO"}]}]}`
	if code, msg, body := post(second); code != 200 || !strings.Contains(body, "LEAN_TOOL_TURN_OK") {
		t.Fatalf("the running tool turn was lost when the setting changed: %d %s %s", code, msg, body)
	}
	select {
	case <-run.interactiveDone:
	case <-time.After(3 * time.Second):
		t.Fatal("wrapper did not exit")
	}
	saved, ok := s.subscription.interactiveSession(owner, outer)
	asStarted := &tuiRuntimeRequest{Request: Request{System: "RULES"}}
	asStarted.leanPrompt = from
	if !ok || saved.Dirty || saved.SessionID != "inner-L" || saved.InstructionKey != tuiRuntimeInstructionKey(asStarted) {
		t.Fatalf("checkpoint not kept in the run's own mode: %#v, %v", saved, ok)
	}
}
