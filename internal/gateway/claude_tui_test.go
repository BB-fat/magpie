package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/claudebridge"
	"github.com/yetone/magpie/internal/claudetui"
	"github.com/yetone/magpie/internal/provider"
)

// tuiAccount is a Claude account the gateway started with set to TUI, its
// runs answered by wrapper (a stand-in for the PTY wrapper) in a folder of
// its own.
func tuiAccount(t *testing.T, wrapper string) provider.Provider {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("the Claude TUI runs on Linux only")
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "xdg"))
	dir := t.TempDir()
	oldDir := tuiAccountDir
	tuiAccountDir = func(context.Context, provider.Provider) (string, error) { return dir, nil }
	tuiModes.once.Do(func() {})
	oldModes := tuiModes.m
	tuiModes.m = map[string]map[string]bool{"claude": {"tui@example.com": true}}
	tuiReady.once.Do(func() {})
	oldReady := tuiReady.err
	tuiReady.err = nil
	claudetui.Init(&claudetui.Config{
		ClaudePath: func() (string, error) { return "/bin/true", nil },
		Wrapper:    []string{wrapper},
		HelperArgv: []string{"/bin/false"},
		MCPServer:  claudebridge.MCPServerName(),
		TempRoot:   t.TempDir(),
		Room: func(ctx context.Context, what string) (func(bool), error) {
			return claudeRoom.take(ctx, what)
		},
		Leave: claudeRoom.leave,
	})
	t.Cleanup(func() {
		tuiAccountDir = oldDir
		tuiModes.m = oldModes
		tuiReady.err = oldReady
		claudetui.ResetForTest()
	})
	return provider.Provider{ID: "claude", Account: &provider.Account{Agent: "claude", User: "tui@example.com"}}
}

// tuiServer is a gateway whose TUI runs have all ended, their processes
// reaped, before the test's other cleanups put the bridge back.
func tuiTestServer(t *testing.T) *Server {
	t.Helper()
	s := New()
	t.Cleanup(func() {
		s.tui.Shutdown()
		waitFor(t, "the TUI runs end", func() bool { return !s.tui.Busy() })
	})
	return s
}

// tuiWrapper writes a stand-in for the PTY wrapper that answers each turn
// with lines (stream-json rows, as the wrapper writes them), and keeps
// its arguments and environment in args and env beside it.
func tuiWrapper(t *testing.T, lines ...string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "wrapper")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > " + filepath.Join(dir, "args") + "\nenv > " + filepath.Join(dir, "env") + "\ncat >/dev/null\n"
	for _, l := range lines {
		script += "printf '%s\\n' '" + l + "'\n"
	}
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

const tuiTextRows = `{"type":"assistant","uuid":"a1","session_id":"s1","message":{"id":"m1","model":"claude-sonnet-5","usage":{"input_tokens":3,"output_tokens":2},"content":[{"type":"text","text":"TUI_OK"}]}}` + "\n" +
	`{"type":"result","session_id":"s1","stop_reason":"end_turn"}`

// A TUI account answers in every API the gateway takes, its reply in the
// client's own shape.
func TestClaudeTUIAnswersEveryAPI(t *testing.T) {
	p := tuiAccount(t, tuiWrapper(t, strings.Split(tuiTextRows, "\n")...))
	s := tuiTestServer(t)
	for _, c := range []struct {
		from provider.Protocol
		body string
		want func(string) bool
	}{
		{provider.Anthropic, `{"model":"claude-sonnet-5","max_tokens":64,"messages":[{"role":"user","content":"hi"}]}`,
			func(b string) bool { return strings.Contains(b, `"type":"message"`) && strings.Contains(b, "TUI_OK") }},
		{provider.Chat, `{"model":"claude-sonnet-5","messages":[{"role":"user","content":"hi"}]}`,
			func(b string) bool { return strings.Contains(b, `"choices"`) && strings.Contains(b, "TUI_OK") }},
		{provider.Responses, `{"model":"claude-sonnet-5","input":"hi"}`,
			func(b string) bool { return strings.Contains(b, `"output"`) && strings.Contains(b, "TUI_OK") }},
		{provider.Gemini, `{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`,
			func(b string) bool { return strings.Contains(b, `"candidates"`) && strings.Contains(b, "TUI_OK") }},
	} {
		rec := httptest.NewRecorder()
		var u Usage
		code, msg := s.serveClaudeSubscription(rec, httptest.NewRequest("POST", "/", strings.NewReader(c.body)), c.from, p, "claude-sonnet-5", []byte(c.body), &u)
		if code != 200 || !c.want(rec.Body.String()) {
			t.Errorf("%s: %d %s %s", c.from, code, msg, rec.Body)
		}
		if u.Input != 3 || u.Output != 2 {
			t.Errorf("%s: usage %+v, want the run's 3 in, 2 out", c.from, u)
		}
	}
	// and streamed
	body := `{"model":"claude-sonnet-5","stream":true,"messages":[{"role":"user","content":"hi"}]}`
	rec := httptest.NewRecorder()
	if code, msg := s.serveClaudeSubscription(rec, httptest.NewRequest("POST", "/", strings.NewReader(body)), provider.Chat, p, "claude-sonnet-5", []byte(body), &Usage{}); code != 200 ||
		!strings.Contains(rec.Body.String(), "data: ") || !strings.Contains(rec.Body.String(), "TUI_OK") {
		t.Fatalf("streamed: %d %s %s", code, msg, rec.Body)
	}
}

// A tool whose name Claude Code would change is offered under the encoded
// name, with the MCP server's prefix, and handed to the client under its
// own.
func TestClaudeTUIToolNamesComeBackAsDeclared(t *testing.T) {
	declared := "spawn.agent"
	offered := claudebridge.MCPToolPrefix() + claudebridge.ToolName(declared)
	rows := []string{
		`{"type":"assistant","uuid":"t1","session_id":"s1","message":{"id":"m1","model":"claude-sonnet-5","content":[{"type":"tool_use","id":"toolu_1","name":"` + offered + `","input":{"task":"x"}}]}}`,
	}
	p := tuiAccount(t, tuiWrapper(t, rows...))
	s := tuiTestServer(t)
	body := `{"model":"claude-sonnet-5","tools":[{"type":"function","function":{"name":"` + declared + `","parameters":{"type":"object"}}}],"messages":[{"role":"user","content":"go"}]}`
	rec := httptest.NewRecorder()
	code, msg := s.serveClaudeSubscription(rec, httptest.NewRequest("POST", "/", strings.NewReader(body)), provider.Chat, p, "claude-sonnet-5", []byte(body), &Usage{})
	if code != 200 {
		t.Fatalf("%d %s %s", code, msg, rec.Body)
	}
	var out struct {
		Choices []struct {
			Message struct {
				ToolCalls []struct {
					Function struct{ Name string } `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}
	if json.Unmarshal(rec.Body.Bytes(), &out) != nil || len(out.Choices) == 0 || len(out.Choices[0].Message.ToolCalls) != 1 ||
		out.Choices[0].Message.ToolCalls[0].Function.Name != declared {
		t.Fatalf("the client got %s, want a call of %q", rec.Body, declared)
	}
}

// A TUI account that can't run here is refused with why, and is not
// answered by the standard bridge in its place.
func TestClaudeTUIFailureIsNotTheStandardBridge(t *testing.T) {
	p := tuiAccount(t, filepath.Join(t.TempDir(), "no-wrapper"))
	fakeClaude(t) // the standard bridge's Claude Code, which must not answer
	s := tuiTestServer(t)
	body := `{"model":"claude-sonnet-5","max_tokens":64,"messages":[{"role":"user","content":"hi"}]}`
	rec := httptest.NewRecorder()
	code, _ := s.serveClaudeSubscription(rec, httptest.NewRequest("POST", "/", strings.NewReader(body)), provider.Anthropic, p, "claude-sonnet-5", []byte(body), &Usage{})
	if code < 500 || strings.Contains(rec.Body.String(), "pid ") || !strings.Contains(rec.Body.String(), "wrapper") {
		t.Fatalf("a TUI account that can't run answered %d %s", code, rec.Body)
	}
	// and a setup that failed is said, not taken for standard
	tuiReady.err = errTestTUI
	rec = httptest.NewRecorder()
	if code, _ := s.serveClaudeSubscription(rec, httptest.NewRequest("POST", "/", strings.NewReader(body)), provider.Anthropic, p, "claude-sonnet-5", []byte(body), &Usage{}); code != 502 || !strings.Contains(rec.Body.String(), "tmux is not installed") {
		t.Fatalf("a TUI that can't run here answered %d %s", code, rec.Body)
	}
}

var errTestTUI = testErr("tmux is not installed")

type testErr string

func (e testErr) Error() string { return string(e) }

// The standard account beside a TUI one stays on the standard bridge.
func TestClaudeTUIOnlyTheAccountsSet(t *testing.T) {
	tuiAccount(t, tuiWrapper(t, strings.Split(tuiTextRows, "\n")...))
	fakeClaude(t)
	s := tuiTestServer(t)
	defer s.subscription.abortAll()
	p := provider.Provider{ID: "claude", Account: &provider.Account{Agent: "claude", User: "standard@example.com"}}
	body := `{"model":"claude-sonnet-5","max_tokens":64,"messages":[{"role":"user","content":"hi"}]}`
	rec := httptest.NewRecorder()
	code, msg := s.serveClaudeSubscription(rec, httptest.NewRequest("POST", "/", strings.NewReader(body)), provider.Anthropic, p, "claude-sonnet-5", []byte(body), &Usage{})
	if code != 200 || strings.Contains(rec.Body.String(), "TUI_OK") || !strings.Contains(rec.Body.String(), "pid ") {
		t.Fatalf("a standard account answered %d %s %s", code, msg, rec.Body)
	}
}

// Client cache marks can't be placed on a TUI run: they are refused,
// not dropped unsaid.
func TestClaudeTUIRefusesClientCacheMarks(t *testing.T) {
	p := tuiAccount(t, tuiWrapper(t, strings.Split(tuiTextRows, "\n")...))
	s := tuiTestServer(t)
	body := `{"model":"claude-sonnet-5","max_tokens":64,"messages":[{"role":"user","content":[{"type":"text","text":"hi","cache_control":{"type":"ephemeral"}}]}]}`
	r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
	r.Header.Set("X-Magpie-Cache-Marks", "client")
	rec := httptest.NewRecorder()
	if code, _ := s.serveClaudeSubscription(rec, r, provider.Anthropic, p, "claude-sonnet-5", []byte(body), &Usage{}); code != 400 || !strings.Contains(rec.Body.String(), "cache marks") {
		t.Fatalf("client cache marks answered %d %s", code, rec.Body)
	}
}

// A TUI run holds one place in the room every Claude Code process takes,
// from before its wrapper starts until the wrapper's whole tree is gone,
// and frees it once.
func TestClaudeTUIRunTakesARoomPlace(t *testing.T) {
	gate := filepath.Join(t.TempDir(), "go")
	rows := append([]string{}, strings.Split(tuiTextRows, "\n")...)
	wrapper := tuiWrapper(t, rows...)
	// the wrapper waits for gate before it answers
	b, _ := os.ReadFile(wrapper)
	script := strings.Replace(string(b), "cat >/dev/null\n", "cat >/dev/null\nwhile [ ! -f "+gate+" ]; do sleep 0.02; done\n", 1)
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	p := tuiAccount(t, wrapper)
	before := claudeRoomLive()
	s := tuiTestServer(t)
	body := `{"model":"claude-sonnet-5","max_tokens":64,"messages":[{"role":"user","content":"hi"}]}`
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		rec := httptest.NewRecorder()
		s.serveClaudeSubscription(rec, httptest.NewRequest("POST", "/", strings.NewReader(body)), provider.Anthropic, p, "claude-sonnet-5", []byte(body), &Usage{})
	}()
	waitFor(t, "the run takes a place", func() bool { return claudeRoomLive() == before+1 })
	if err := os.WriteFile(gate, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	waitFor(t, "the run frees its place", func() bool { return claudeRoomLive() == before })
}

func claudeRoomLive() int {
	claudeRoom.mu.Lock()
	defer claudeRoom.mu.Unlock()
	return claudeRoom.live
}

// The run's MCP helper reads its callback from a file of the run's own:
// no token on any command line.
func TestClaudeTUIKeepsTheTokenOffTheCommandLine(t *testing.T) {
	wrapper := tuiWrapper(t, strings.Split(tuiTextRows, "\n")...)
	p := tuiAccount(t, wrapper)
	s := tuiTestServer(t)
	body := `{"model":"claude-sonnet-5","max_tokens":64,"tools":[{"name":"echo","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"hi"}]}`
	rec := httptest.NewRecorder()
	if code, msg := s.serveClaudeSubscription(rec, httptest.NewRequest("POST", "/", strings.NewReader(body)), provider.Anthropic, p, "claude-sonnet-5", []byte(body), &Usage{}); code != 200 {
		t.Fatalf("%d %s", code, msg)
	}
	args, err := os.ReadFile(filepath.Join(filepath.Dir(wrapper), "args"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(args), claudeMCPPath) || strings.Contains(string(args), "http://") {
		t.Fatalf("a callback URL is on the wrapper's command line: %s", args)
	}
	if !strings.Contains(string(args), `"mcpServers":{"`+claudebridge.MCPServerName()+`"`) || !strings.Contains(string(args), "/callback") {
		t.Fatalf("the MCP helper isn't given its callback file: %s", args)
	}
}

func waitFor(t *testing.T, what string, ok func() bool) {
	t.Helper()
	for i := 0; i < 500; i++ {
		if ok() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("%s: never", what)
}

// A Chat client's history of calls with no tools offered reaches the TUI
// with no tools: the stand-ins buildAnthropic declares for Anthropic would
// be offered to the run as real ones.
func TestClaudeTUIMessagesDeclareNoStandInTools(t *testing.T) {
	body := `{"model":"claude-sonnet-5","messages":[{"role":"user","content":"read it"},` +
		`{"role":"assistant","content":null,"tool_calls":[{"id":"c1","type":"function","function":{"name":"read","arguments":"{}"}}]},` +
		`{"role":"tool","tool_call_id":"c1","content":"text"},{"role":"user","content":"thanks"}]}`
	req, err := parse(provider.Chat, []byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(buildAnthropic(req, "claude-sonnet-5")), `"tools"`) {
		t.Fatal("buildAnthropic declares no stand-ins here any more: this test tests nothing")
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(tuiMessages(req, "claude-sonnet-5"), &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["tools"]; ok {
		t.Fatalf("the TUI is offered the stand-in tools: %s", m["tools"])
	}
	if !strings.Contains(string(m["messages"]), `"tool_use"`) {
		t.Fatalf("the history's calls are gone: %s", m["messages"])
	}
}

// A TUI run's callback token is let in without a gateway key while the run
// lives, and not after; a token no run has is 404.
func TestClaudeTUICallbackTokenLivesWithTheRun(t *testing.T) {
	gate := filepath.Join(t.TempDir(), "go")
	wrapper := tuiWrapper(t, strings.Split(tuiTextRows, "\n")...)
	b, _ := os.ReadFile(wrapper)
	script := strings.Replace(string(b), "cat >/dev/null\n", "cat >/dev/null\nwhile [ ! -f "+gate+" ]; do sleep 0.02; done\n", 1)
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	p := tuiAccount(t, wrapper)
	claudetui.Init(&claudetui.Config{AddCallback: addCallback, DropCallback: dropCallback})
	s := tuiTestServer(t)
	body := `{"model":"claude-sonnet-5","max_tokens":64,"messages":[{"role":"user","content":"hi"}]}`
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.serveClaudeSubscription(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(body)), provider.Anthropic, p, "claude-sonnet-5", []byte(body), &Usage{})
	}()
	// the run's callback file, as its wrapper was given it
	var token string
	argsFile := filepath.Join(filepath.Dir(wrapper), "args")
	waitFor(t, "the run's callback file", func() bool {
		args, err := os.ReadFile(argsFile)
		if err != nil {
			return false
		}
		for _, a := range strings.Split(string(args), "\n") {
			var cfg struct {
				MCPServers map[string]struct{ Args []string } `json:"mcpServers"`
			}
			if json.Unmarshal([]byte(a), &cfg) != nil {
				continue
			}
			for _, srv := range cfg.MCPServers {
				if n := len(srv.Args); n >= 2 {
					if b, err := os.ReadFile(srv.Args[n-2]); err == nil {
						token = strings.TrimPrefix(string(b)[strings.Index(string(b), claudeMCPPath):], claudeMCPPath)
						return true
					}
				}
			}
		}
		return false
	})
	call := func(tok string) *http.Request {
		return httptest.NewRequest("POST", claudeMCPPath+tok, strings.NewReader(`{}`))
	}
	if !ownCallback(call(token)) {
		t.Fatal("a live TUI run's token isn't let in")
	}
	if ownCallback(call(randomToken())) {
		t.Fatal("a token no run has is let in")
	}
	rec := httptest.NewRecorder()
	r := call(randomToken())
	r.SetPathValue("token", strings.TrimPrefix(r.URL.Path, claudeMCPPath))
	s.claudeMCP(rec, r)
	if rec.Code != 404 {
		t.Fatalf("an unknown token answered %d", rec.Code)
	}
	if err := os.WriteFile(gate, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	<-done
	waitFor(t, "the run ends", func() bool { return !s.tui.Busy() })
	if ownCallback(call(token)) {
		t.Fatal("an ended TUI run's token is still let in")
	}
}
