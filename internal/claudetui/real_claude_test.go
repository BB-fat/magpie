package claudetui

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestRealClaudeTUI runs the real Claude Code in a real private tmux, by the
// real wrapper bundle on a real Bun, its tool over a built magpie's MCP
// helper, against a local stand-in for Anthropic's API: everything the
// bridge does but Anthropic's own answers. It runs only when asked:
//
//	MAGPIE_CLAUDE_TUI_REAL=1 MAGPIE_TUI_BUN=<bun> MAGPIE_TUI_CLAUDE=<claude> \
//	MAGPIE_TUI_MAGPIE=<built magpie> go test -run TestRealClaudeTUI ./internal/claudetui/
func TestRealClaudeTUI(t *testing.T) {
	if os.Getenv("MAGPIE_CLAUDE_TUI_REAL") != "1" {
		t.Skip("set MAGPIE_CLAUDE_TUI_REAL=1 and the programs to run the real Claude Code")
	}
	for _, server := range []string{"magpie", "corp-tools"} {
		t.Run(server, func(t *testing.T) { realRun(t, server) })
	}
}

func realRun(t *testing.T, server string) {
	bun, claude, magpie := os.Getenv("MAGPIE_TUI_BUN"), os.Getenv("MAGPIE_TUI_CLAUDE"), os.Getenv("MAGPIE_TUI_MAGPIE")
	tmux, err := exec.LookPath("tmux")
	if bun == "" || claude == "" || magpie == "" || err != nil {
		t.Fatal("MAGPIE_TUI_BUN, MAGPIE_TUI_CLAUDE, MAGPIE_TUI_MAGPIE and tmux are needed")
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "xdg"))
	t.Setenv("PATH", "/usr/bin:/bin") // no bun on PATH: the hooks run the one given
	up := newAnthropicStandIn(t)
	profile := t.TempDir()
	settings := fmt.Sprintf(`{"env":{"ANTHROPIC_BASE_URL":%q,"ANTHROPIC_AUTH_TOKEN":"stand-in"}}`, up.URL)
	if err := os.WriteFile(filepath.Join(profile, "settings.json"), []byte(settings), 0o600); err != nil {
		t.Fatal(err)
	}
	b := New()
	cb := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !b.ServeMCP(w, r, strings.TrimPrefix(r.URL.Path, "/_magpie/claude-mcp/")) {
			http.NotFound(w, r)
		}
	}))
	defer cb.Close()
	bundle, err := InstallWrapper()
	if err != nil {
		t.Fatal(err)
	}
	old := config
	t.Cleanup(func() { config = old })
	Init(&Config{
		ClaudePath:   func() (string, error) { return claude, nil },
		Tmux:         tmux,
		Wrapper:      []string{bun, bundle},
		HelperArgv:   []string{magpie, "claude-mcp-helper"},
		MCPServer:    server,
		CallbackBase: func() string { return cb.URL },
	})
	t.Setenv("MAGPIE_CLAUDE_MCP_NAME", server) // the helper reads the same name
	owner := Owner("claude", "real@example.com", server)
	model := "claude-sonnet-5-5"
	tools := `[{"name":"echo","description":"Echo text back.","input_schema":{"type":"object","properties":{"text":{"type":"string"}},"required":["text"]}}]`
	first := `{"model":"` + model + `","max_tokens":256,"stream":true,"tools":` + tools + `,"messages":[{"role":"user","content":"call echo with hi"}]}`

	out := &recorder{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(first)).WithContext(ctx)
	code, msg := b.Serve(r, out, model, owner, "conv-1", profile, []byte(first))
	if code != 200 {
		t.Fatalf("first turn: %d %s", code, msg)
	}
	call := out.tool()
	if call.name != "echo" || !strings.Contains(call.args, "hi") {
		t.Fatalf("first turn called %+v, want echo; events %v; said %q; stand-in saw tools %s", call, out.kinds(), out.text(), up.toolNames())
	}

	second := `{"model":"` + model + `","max_tokens":256,"stream":true,"tools":` + tools + `,"messages":[` +
		`{"role":"user","content":"call echo with hi"},` +
		`{"role":"assistant","content":[{"type":"tool_use","id":` + jsonString(call.id) + `,"name":"echo","input":` + call.args + `}]},` +
		`{"role":"user","content":[{"type":"tool_result","tool_use_id":` + jsonString(call.id) + `,"content":"ECHOED"}]}]}`
	out2 := &recorder{}
	r2 := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(second)).WithContext(ctx)
	if code, msg := b.Serve(r2, out2, model, owner, "conv-1", profile, []byte(second)); code != 200 {
		t.Fatalf("tool results: %d %s", code, msg)
	}
	if !strings.Contains(out2.text(), "DONE: ECHOED") {
		t.Fatalf("tool results answered %q", out2.text())
	}
	if p := up.toolNames(); !strings.Contains(p, "mcp__"+server+"__echo") {
		t.Fatalf("Claude Code offered %s, not mcp__%s__echo", p, server)
	}
	// the run ends with its turn: its tmux server and every process under
	// the wrapper go, and the room's place with them
	deadline := time.Now().Add(30 * time.Second)
	for b.Busy() {
		if time.Now().After(deadline) {
			t.Fatal("the run's processes outlived its reply")
		}
		time.Sleep(100 * time.Millisecond)
	}
	if left, _ := filepath.Glob(filepath.Join(os.TempDir(), fmt.Sprintf("magpie-claude-tui-%d", os.Getuid()), fmt.Sprintf("%d-*.sock", os.Getpid()))); len(left) != 0 {
		t.Fatalf("tmux sockets left: %v", left)
	}
	// nothing of magpie's went in the user's settings
	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".claude", "settings.json")); err == nil {
		t.Fatal("the run wrote ~/.claude/settings.json")
	}
}

type standIn struct {
	*httptest.Server
	mu    sync.Mutex
	tools []string
}

// newAnthropicStandIn answers Claude Code as Anthropic's Messages API
// would: a call of the echo tool when asked to, then "DONE: <result>".
func newAnthropicStandIn(t *testing.T) *standIn {
	s := &standIn{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/count_tokens") {
			fmt.Fprint(w, `{"input_tokens":10}`)
			return
		}
		if !strings.HasSuffix(r.URL.Path, "/v1/messages") {
			http.NotFound(w, r)
			return
		}
		var req struct {
			Stream bool `json:"stream"`
			Tools  []struct {
				Name string `json:"name"`
			} `json:"tools"`
			Messages []struct {
				Role    string          `json:"role"`
				Content json.RawMessage `json:"content"`
			} `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		echo := ""
		for _, tl := range req.Tools {
			s.mu.Lock()
			s.tools = append(s.tools, tl.Name)
			s.mu.Unlock()
			if strings.HasSuffix(tl.Name, "__echo") {
				echo = tl.Name
			}
		}
		// what Claude Code adds to a turn (reminders, context) comes after
		// the user's words: the conversation is read whole
		var all strings.Builder
		for _, m := range req.Messages {
			all.Write(m.Content)
		}
		last := all.String()
		var blocks []map[string]any
		stop := "end_turn"
		switch {
		case strings.Contains(last, "ECHOED"):
			blocks = []map[string]any{{"type": "text", "text": "DONE: ECHOED"}}
		case echo != "" && strings.Contains(last, "call echo"):
			blocks = []map[string]any{{"type": "tool_use", "id": "toolu_standin_1", "name": echo, "input": map[string]any{"text": "hi"}}}
			stop = "tool_use"
		default:
			blocks = []map[string]any{{"type": "text", "text": "PONG"}}
		}
		writeStandIn(w, req.Stream, blocks, stop)
	}))
	t.Cleanup(s.Close)
	return s
}

func (s *standIn) toolNames() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return strings.Join(s.tools, ",")
}

func writeStandIn(w http.ResponseWriter, stream bool, blocks []map[string]any, stop string) {
	msg := map[string]any{"id": "msg_standin", "type": "message", "role": "assistant", "model": "claude-sonnet-5-5",
		"stop_reason": stop, "usage": map[string]any{"input_tokens": 10, "output_tokens": 5}}
	if !stream {
		msg["content"] = blocks
		_ = json.NewEncoder(w).Encode(msg)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	bw := bufio.NewWriter(w)
	send := func(ev string, v any) {
		b, _ := json.Marshal(v)
		fmt.Fprintf(bw, "event: %s\ndata: %s\n\n", ev, b)
	}
	start := map[string]any{}
	for k, v := range msg {
		start[k] = v
	}
	start["content"], start["stop_reason"] = []any{}, nil
	send("message_start", map[string]any{"type": "message_start", "message": start})
	for i, bl := range blocks {
		if bl["type"] == "text" {
			send("content_block_start", map[string]any{"type": "content_block_start", "index": i, "content_block": map[string]any{"type": "text", "text": ""}})
			send("content_block_delta", map[string]any{"type": "content_block_delta", "index": i, "delta": map[string]any{"type": "text_delta", "text": bl["text"]}})
		} else {
			args, _ := json.Marshal(bl["input"])
			send("content_block_start", map[string]any{"type": "content_block_start", "index": i, "content_block": map[string]any{"type": "tool_use", "id": bl["id"], "name": bl["name"], "input": map[string]any{}}})
			send("content_block_delta", map[string]any{"type": "content_block_delta", "index": i, "delta": map[string]any{"type": "input_json_delta", "partial_json": string(args)}})
		}
		send("content_block_stop", map[string]any{"type": "content_block_stop", "index": i})
	}
	send("message_delta", map[string]any{"type": "message_delta", "delta": map[string]any{"stop_reason": stop}, "usage": map[string]any{"output_tokens": 5}})
	send("message_stop", map[string]any{"type": "message_stop"})
	_ = bw.Flush()
}

// recorder is a Responder that keeps the events of the reply.
type recorder struct {
	mu     sync.Mutex
	header http.Header
	events []Event
}

func (c *recorder) Header() http.Header {
	if c.header == nil {
		c.header = http.Header{}
	}
	return c.header
}

func (c *recorder) Fail(status int, msg string) (int, string) { return status, msg }

func (c *recorder) Relay(events <-chan Event, _ bool, _ func(), ended func(said, stop string, ok bool)) (int, string) {
	var said, stop, failed string
	for ev := range events {
		c.mu.Lock()
		c.events = append(c.events, ev)
		c.mu.Unlock()
		switch ev.Kind {
		case KText:
			said += ev.Text
		case KStop:
			stop = ev.Stop
		case KError:
			failed = ev.Text
		}
	}
	ended(said, stop, failed == "")
	if failed != "" {
		return 502, failed
	}
	return 200, ""
}

type seenCall struct{ id, name, args string }

func (c *recorder) tool() seenCall {
	var s seenCall
	for _, ev := range c.events {
		switch ev.Kind {
		case KToolStart:
			s.id, s.name = ev.ID, ev.Name
		case KToolArgs:
			s.args += ev.Text
		}
	}
	return s
}

func (c *recorder) text() string {
	var b strings.Builder
	for _, ev := range c.events {
		if ev.Kind == KText {
			b.WriteString(ev.Text)
		}
	}
	return b.String()
}

func (c *recorder) kinds() []EventKind {
	var k []EventKind
	for _, ev := range c.events {
		k = append(k, ev.Kind)
	}
	return k
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
