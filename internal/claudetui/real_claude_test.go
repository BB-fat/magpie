package claudetui

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
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
	t.Run("cancel", realCancel)
	t.Run("restart", realRestart)
	t.Run("cache prefix", realCachePrefix)
	t.Run("gateway killed", realGatewayKilled)
	t.Run("memory", realMemory)
}

// A turn whose processes grow past claudeTuiMaxRSSMiB is ended with an
// error saying so, and nothing of it is left running.
func realMemory(t *testing.T) {
	up, profile, bridge := realSetup(t, "magpie")
	old := rssEvery
	rssEvery = 200 * time.Millisecond
	t.Cleanup(func() { rssEvery = old })
	Init(&Config{MaxRSS: 1 << 20}) // the real tree holds far more than 1 MiB
	b := bridge()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	go func() { <-up.slow }()
	out, code, msg := realServe(t, b, ctx, profile, "", `{"model":"`+realModel+`","max_tokens":64,"stream":true,"messages":[{"role":"user","content":"take your time"}]}`)
	if code == 200 && !strings.Contains(msg, "claudeTuiMaxRSSMiB") {
		t.Fatalf("a run past its memory answered %d %q %q", code, msg, out.text())
	}
	if !strings.Contains(msg, "claudeTuiMaxRSSMiB") {
		t.Fatalf("the run ended with %d %q, not its memory", code, msg)
	}
	waitGone(t, b, profile)
}

// A client that goes away mid-reply ends its run: the wrapper, its tmux
// and Claude Code, with nothing of them left running.
func realCancel(t *testing.T) {
	up, profile, bridge := realSetup(t, "magpie")
	b := bridge()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		realServe(t, b, ctx, profile, "", `{"model":"`+realModel+`","max_tokens":64,"stream":true,"messages":[{"role":"user","content":"take your time"}]}`)
	}()
	select {
	case <-up.slow:
	case <-time.After(2 * time.Minute):
		t.Fatal("Claude Code never asked the stand-in")
	}
	if left := underProfile(profile); len(left) == 0 {
		t.Fatal("no process runs with the profile: the check below tests nothing")
	}
	cancel()
	<-done
	waitGone(t, b, profile)
}

// A conversation goes on after the gateway is started again: its next
// turn is answered from its checkpoint, or rebuilt from the request.
func realRestart(t *testing.T) {
	up, profile, bridge := realSetup(t, "magpie")
	b := bridge()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	first := `{"model":"` + realModel + `","max_tokens":64,"stream":true,"messages":[{"role":"user","content":"say hello"}]}`
	if out, code, msg := realServe(t, b, ctx, profile, "conv-restart", first); code != 200 || out.text() != "PONG" {
		t.Fatalf("first turn: %d %s %q", code, msg, out.text())
	}
	waitGone(t, b, profile)
	b = bridge() // the gateway started again
	next := `{"model":"` + realModel + `","max_tokens":64,"stream":true,"messages":[{"role":"user","content":"say hello"},{"role":"assistant","content":"PONG"},{"role":"user","content":"and again"}]}`
	before := up.asked()
	out, code, msg := realServe(t, b, ctx, profile, "conv-restart", next)
	if code != 200 || out.text() != "PONG" {
		t.Fatalf("after a restart: %d %s %q", code, msg, out.text())
	}
	if !strings.Contains(up.lastAsked(), "say hello") || up.asked() == before {
		t.Fatalf("the turn after a restart was asked without its history: %s", up.lastAsked())
	}
	waitGone(t, b, profile)
}

// The real CLI must preserve the tools and prior message prefix across a
// checkpoint resume; changing them would prevent upstream cache reuse.
func realCachePrefix(t *testing.T) {
	up, profile, bridge := realSetup(t, "magpie")
	b := bridge()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	first := `{"model":"` + realModel + `","tools":` + realTools + `,"max_tokens":64,"stream":true,"messages":[{"role":"user","content":"say hello"}]}`
	if out, code, msg := realServe(t, b, ctx, profile, "cache-prefix", first); code != 200 || out.text() != "PONG" {
		t.Fatalf("first: %d %s %q", code, msg, out.text())
	}
	waitGone(t, b, profile)
	up.mu.Lock()
	before := append([]json.RawMessage(nil), up.requests...)
	up.mu.Unlock()
	b = bridge()
	next := `{"model":"` + realModel + `","tools":` + realTools + `,"max_tokens":64,"stream":true,"messages":[{"role":"user","content":"say hello"},{"role":"assistant","content":"PONG"},{"role":"user","content":"and again"}]}`
	if out, code, msg := realServe(t, b, ctx, profile, "cache-prefix", next); code != 200 || out.text() != "PONG" {
		t.Fatalf("resumed: %d %s %q", code, msg, out.text())
	}
	waitGone(t, b, profile)
	up.mu.Lock()
	after := append([]json.RawMessage(nil), up.requests[len(before):]...)
	up.mu.Unlock()
	main := func(rows []json.RawMessage) map[string]any {
		t.Helper()
		for _, raw := range rows {
			var value map[string]any
			if json.Unmarshal(raw, &value) == nil && strings.Contains(string(raw), "say hello") {
				return value
			}
		}
		t.Fatalf("no main request in %d upstream requests", len(rows))
		return nil
	}
	old, resumed := main(before), main(after)
	if !reflect.DeepEqual(old["tools"], resumed["tools"]) {
		t.Fatal("tool definitions changed on a checkpoint resume")
	}
	clean := func(value any) any {
		raw, _ := json.Marshal(value)
		var result any
		_ = json.Unmarshal(raw, &result)
		var visit func(any)
		visit = func(v any) {
			switch x := v.(type) {
			case map[string]any:
				delete(x, "cache_control")
				for _, item := range x {
					visit(item)
				}
				if blocks, ok := x["content"].([]any); ok && len(blocks) == 1 {
					if block, ok := blocks[0].(map[string]any); ok && len(block) == 2 && block["type"] == "text" {
						x["content"] = block["text"]
					}
				}
			case []any:
				for _, item := range x {
					visit(item)
				}
			}
		}
		visit(result)
		return result
	}
	if tools, ok := old["tools"].([]any); !ok || len(tools) == 0 {
		t.Fatal("the prefix test offered no MCP tools")
	}
	instructions := func(value any) any {
		value = clean(value)
		if blocks, ok := value.([]any); ok {
			for _, item := range blocks {
				if block, ok := item.(map[string]any); ok {
					if text, ok := block["text"].(string); ok {
						block["text"] = withoutBillingHeader(text)
					}
				}
			}
		}
		return value
	}
	if !reflect.DeepEqual(instructions(old["system"]), instructions(resumed["system"])) {
		t.Fatal("the resumed request changed its system prefix beyond billing metadata")
	}
	prior := old["messages"].([]any)
	current := resumed["messages"].([]any)
	if len(current) <= len(prior) || !reflect.DeepEqual(clean(prior), clean(current[:len(prior)])) {
		a, _ := json.Marshal(clean(prior))
		z, _ := json.Marshal(clean(current))
		t.Fatalf("the resumed upstream request changed its prior message prefix:\nbefore=%s\nafter=%s", a, z)
	}
}

// The gateway killed outright: its wrapper is told to end (its parent's
// death signal) and takes its tmux and Claude Code with it; the next
// gateway's sweep removes the tmux socket left.
func realGatewayKilled(t *testing.T) {
	up, profile, _ := realSetup(t, "magpie")
	cmd := exec.Command(os.Args[0], "-test.run=^TestRealClaudeTUIChild$")
	cmd.Env = append(os.Environ(), "MAGPIE_TUI_CHILD_PROFILE="+profile, "MAGPIE_TUI_CHILD_UP="+up.URL, "XDG_CONFIG_HOME="+os.Getenv("XDG_CONFIG_HOME"))
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-up.slow:
	case <-time.After(2 * time.Minute):
		_ = cmd.Process.Kill()
		t.Fatal("the child gateway's Claude Code never asked the stand-in")
	}
	if len(underProfile(profile)) == 0 {
		t.Fatal("no process runs with the profile: the check below tests nothing")
	}
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
	deadline := time.Now().Add(30 * time.Second)
	for len(underProfile(profile)) > 0 {
		if time.Now().After(deadline) {
			t.Fatalf("processes outlived the killed gateway: %v", underProfile(profile))
		}
		time.Sleep(100 * time.Millisecond)
	}
	Sweep()
	if left, _ := filepath.Glob(filepath.Join(tempRoot(), fmt.Sprintf("magpie-claude-tui-%d", os.Getuid()), fmt.Sprintf("%d-*.sock", cmd.Process.Pid))); len(left) != 0 {
		t.Fatalf("the killed gateway's tmux sockets are left: %v", left)
	}
	if left, _ := filepath.Glob(filepath.Join(tempRoot(), fmt.Sprintf("magpie-claude-pty-%d-*", cmd.Process.Pid))); len(left) != 0 {
		t.Fatalf("the killed gateway's run folders are left: %v", left)
	}
}

// TestRealClaudeTUIChild is the gateway realGatewayKilled kills: it starts
// a turn the stand-in answers slowly, and waits to be killed.
func TestRealClaudeTUIChild(t *testing.T) {
	profile := os.Getenv("MAGPIE_TUI_CHILD_PROFILE")
	if profile == "" {
		t.Skip("run by realGatewayKilled")
	}
	b := New()
	cb := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b.ServeMCP(w, r, strings.TrimPrefix(r.URL.Path, "/_magpie/claude-mcp/"))
	}))
	defer cb.Close()
	bundle, err := InstallWrapper()
	if err != nil {
		t.Fatal(err)
	}
	tmux, _ := exec.LookPath("tmux")
	Init(&Config{
		ClaudePath: func() (string, error) { return os.Getenv("MAGPIE_TUI_CLAUDE"), nil },
		Tmux:       tmux, Wrapper: []string{os.Getenv("MAGPIE_TUI_BUN"), bundle},
		HelperArgv: []string{os.Getenv("MAGPIE_TUI_MAGPIE"), "claude-mcp-helper"}, CallbackBase: func() string { return cb.URL },
	})
	realServe(t, b, context.Background(), profile, "", `{"model":"`+realModel+`","max_tokens":64,"stream":true,"messages":[{"role":"user","content":"take your time"}]}`)
	select {}
}

// underProfile is the processes running with profile as their Claude
// config folder.
func underProfile(profile string) []int {
	var out []int
	entries, _ := os.ReadDir("/proc")
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil || pid == os.Getpid() {
			continue
		}
		b, err := os.ReadFile("/proc/" + e.Name() + "/environ")
		if err == nil && strings.Contains(string(b), "CLAUDE_CONFIG_DIR="+profile+"\x00") {
			out = append(out, pid)
		}
	}
	return out
}

// waitGone waits for b's runs to end and every process under profile with
// them, and for their tmux sockets to go.
func waitGone(t *testing.T, b *Bridge, profile string) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for b.Busy() || len(underProfile(profile)) > 0 {
		if time.Now().After(deadline) {
			t.Fatalf("the run outlived its reply: busy %v, processes %v", b.Busy(), underProfile(profile))
		}
		time.Sleep(100 * time.Millisecond)
	}
	if left, _ := filepath.Glob(filepath.Join(tempRoot(), fmt.Sprintf("magpie-claude-tui-%d", os.Getuid()), fmt.Sprintf("%d-*.sock", os.Getpid()))); len(left) != 0 {
		t.Fatalf("tmux sockets left: %v", left)
	}
}

// realSetup readies the real programs for server's runs: the stand-in for
// Anthropic, a profile pointed at it, and the callback the helper calls.
func realSetup(t *testing.T, server string) (*standIn, string, func() *Bridge) {
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
	// a gateway's bridge; a new one stands for the gateway started again
	var mu sync.Mutex
	var b *Bridge
	cb := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		cur := b
		mu.Unlock()
		if cur == nil || !cur.ServeMCP(w, r, strings.TrimPrefix(r.URL.Path, "/_magpie/claude-mcp/")) {
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(cb.Close)
	bridge := func() *Bridge {
		mu.Lock()
		defer mu.Unlock()
		if b != nil {
			b.Shutdown()
		}
		b = New()
		nb := b
		t.Cleanup(func() {
			nb.Shutdown()
			for deadline := time.Now().Add(30 * time.Second); nb.Busy() && time.Now().Before(deadline); {
				time.Sleep(50 * time.Millisecond)
			}
		})
		return b
	}
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
	return up, profile, bridge
}

const realModel = "claude-sonnet-5-5"
const realTools = `[{"name":"echo","description":"Echo text back.","input_schema":{"type":"object","properties":{"text":{"type":"string"}},"required":["text"]}}]`

func realServe(t *testing.T, b *Bridge, ctx context.Context, profile, outer, body string) (*recorder, int, string) {
	t.Helper()
	out := &recorder{}
	r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body)).WithContext(ctx)
	code, msg := b.Serve(r, out, realModel, Owner("claude", "real@example.com", config.MCPServer), outer, profile, []byte(body))
	return out, code, msg
}

func realRun(t *testing.T, server string) {
	up, profile, bridge := realSetup(t, server)
	b := bridge()
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
	waitGone(t, b, profile)
	// the very same results again are answered as before, with no new run
	asked := up.asked()
	out3 := &recorder{}
	r3 := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(second)).WithContext(ctx)
	if code, msg := b.Serve(r3, out3, model, owner, "conv-1", profile, []byte(second)); code != 200 || !strings.Contains(out3.text(), "DONE: ECHOED") {
		t.Fatalf("the same results again: %d %s %q", code, msg, out3.text())
	}
	if up.asked() != asked {
		t.Fatal("the same results again asked Anthropic again")
	}
	// results for a call the visible reply didn't make are refused
	wrong := strings.ReplaceAll(second, `"tool_use_id":`+jsonString(call.id), `"tool_use_id":"toolu_other"`)
	if code, _ := b.Serve(httptest.NewRequest("POST", "/v1/messages", strings.NewReader(wrong)), &recorder{}, model, owner, "conv-1", profile, []byte(wrong)); code != 409 {
		t.Fatalf("a result for another call: %d, want 409", code)
	}
	// nothing of magpie's went in the user's settings
	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".claude", "settings.json")); err == nil {
		t.Fatal("the run wrote ~/.claude/settings.json")
	}
}

type standIn struct {
	*httptest.Server
	mu       sync.Mutex
	tools    []string
	n        int
	last     string
	requests []json.RawMessage
	// slow is told once a turn asked to take its time has come
	slow     chan struct{}
	slowOnce sync.Once
}

func (s *standIn) asked() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.n
}

func (s *standIn) lastAsked() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.last
}

// newAnthropicStandIn answers Claude Code as Anthropic's Messages API
// would: a call of the echo tool when asked to, then "DONE: <result>".
func newAnthropicStandIn(t *testing.T) *standIn {
	s := &standIn{slow: make(chan struct{})}
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
		raw, err := io.ReadAll(r.Body)
		if err != nil || json.Unmarshal(raw, &req) != nil {
			http.Error(w, "invalid stand-in request", 400)
			return
		}
		s.mu.Lock()
		s.requests = append(s.requests, json.RawMessage(raw))
		s.mu.Unlock()
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
		if echo != "" || len(req.Tools) == 0 && strings.Contains(last, "say hello") || strings.Contains(last, "take your time") {
			s.mu.Lock()
			s.n++
			s.last = last
			s.mu.Unlock()
		}
		if strings.Contains(last, "take your time") {
			s.slowOnce.Do(func() { close(s.slow) })
			select {
			case <-r.Context().Done():
			case <-time.After(5 * time.Minute):
			}
			return
		}
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
