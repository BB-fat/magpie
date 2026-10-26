package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/settings"
)

// The wrapper exercises native search or a real MCP callback, according to
// the tools the bridge actually offers. All HTTP goes to the local gateway.
func tuiReviewWrapper(t *testing.T, mode string) string {
	t.Helper()
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("the wrapper stand-in uses Python's JSON and HTTP libraries")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "wrapper")
	script := `#!/usr/bin/env python3
import json, os, sys, urllib.request, time
root = os.path.dirname(__file__)
mode = ` + string(mustJSON(mode)) + `
args = sys.argv[1:]
with open(root + '/starts', 'a') as f: f.write(str(os.getpid()) + '\n')
with open(root + '/args', 'w') as f: json.dump(args, f)
config = json.loads(args[args.index('--mcp-config') + 1])
server = next(iter(config['mcpServers'].values()))
callback, toolsfile = server['args'][-2:]
with open(callback) as f: url = f.read()
with open(toolsfile) as f: tools = json.load(f)
with open(root + '/tools', 'w') as f: json.dump(tools, f)
sys.stdin.read()
def emit(row): print(json.dumps(row), flush=True)
def assistant(content, uuid):
    emit({'type':'assistant','uuid':uuid,'session_id':'review-session','message':{'id':uuid,'model':'claude-sonnet-5','content':content}})
def call(name, ident):
    data = json.dumps({'name':name,'tool_call_id':ident,'arguments':{'query':'latest go'}}).encode()
    with urllib.request.urlopen(urllib.request.Request(url, data=data, headers={'Content-Type':'application/json'}), timeout=15) as r:
        return json.load(r)
if mode == 'tool':
    assistant([{'type':'tool_use','id':'review-call','name':'mcp__magpie__echo','input':{'query':'latest go'}}], 'tool-message')
    result = call('echo', 'review-call')
    with open(root + '/result', 'w') as f: json.dump(result, f)
    while not os.path.exists(root + '/release'): time.sleep(0.01)
    assistant([{'type':'text','text':'TOOL_DONE'}], 'done-message')
else:
    search = next((t['name'] for t in tools if 'web_search' in t['name']), '')
    if search:
        if os.path.exists(root + '/searching'):
            emit({'type':'assistant','isApiErrorMessage':True,'api_error_status':502,'message':{'model':'<synthetic>','content':[{'type':'text','text':'recursive search'}]}})
            sys.exit(0)
        with open(root + '/searching', 'w') as f: f.write('active')
        result = call(search, 'search-' + str(os.getpid()))
        os.remove(root + '/searching')
        text = result['content'][0]['text']
    elif args[args.index('--tools') + 1] == 'WebSearch':
        assistant([{'type':'tool_use','id':'native-search','name':'WebSearch','input':{'query':'latest go'}}], 'search-message')
        emit({'type':'user','session_id':'review-session','message':{'role':'user','content':[{'type':'tool_result','tool_use_id':'native-search','content':'Go downloads'}]},'tool_use_result':{'query':'latest go','results':[{'content':[{'title':'Go downloads','url':'https://go.dev/dl/'}]}]}})
        text = 'NATIVE_SEARCH_OK'
    else:
        text = 'SEARCH_UNAVAILABLE'
    assistant([{'type':'text','text':text}], 'answer-message')
emit({'type':'result','session_id':'review-session','stop_reason':'end_turn'})
`
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func tuiReviewSearch(t *testing.T, s *Server) {
	t.Helper()
	old := tuiServer.Swap(s)
	t.Cleanup(func() { tuiServer.Store(old) })
	initTUIRuntime(tuiConfig())
	callbacks := httptest.NewServer(s.Handler())
	t.Cleanup(func() {
		s.tui.Shutdown()
		callbacks.Close()
	})
	initTUIRuntime(&tuiRuntimeConfig{CallbackBase: func() string { return callbacks.URL }})
}

func tuiReviewStarts(t *testing.T, wrapper string) int {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(filepath.Dir(wrapper), "starts"))
	if err != nil {
		t.Fatal(err)
	}
	return len(strings.Fields(string(b)))
}

func TestClaudeTUISearcherUsesNativeSearch(t *testing.T) {
	fresh(t)
	wrapper := tuiReviewWrapper(t, "search")
	tuiAccount(t, wrapper)
	if err := os.MkdirAll(filepath.Dir(provider.Path()), 0o700); err != nil {
		t.Fatal(err)
	}
	login := []map[string]any{{"agent": "claude", "user": "tui@example.com", "plan": "max", "on": true, "seen": time.Now(),
		"auth": map[string]any{"claudeAiOauth": map[string]any{"accessToken": "fake-tui", "refreshToken": "fake-refresh", "expiresAt": time.Now().Add(24 * time.Hour).UnixMilli(), "subscriptionType": "max"}}}}
	if err := os.WriteFile(filepath.Join(filepath.Dir(provider.Path()), "logins.json"), mustJSON(login), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(catalog.CachePath()), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(catalog.CachePath(), []byte(`{"anthropic":{"models":{"claude-sonnet-5":{"id":"claude-sonnet-5","name":"Claude Sonnet 5","modalities":{"input":["text"],"output":["text"]}}}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	catalog.Reset()
	t.Cleanup(catalog.Reset)
	provider.ForgetAccounts()
	t.Cleanup(provider.ForgetAccounts)
	st := settings.Load()
	st.Searcher = "claude/claude-sonnet-5"
	if err := settings.Save(st); err != nil {
		t.Fatal(err)
	}
	s := tuiTestServer(t)
	tuiReviewSearch(t, s)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	said, hits, err := s.webSearch(ctx, "latest go")
	if err != nil || !strings.Contains(said, "NATIVE_SEARCH_OK") || len(hits) != 1 || hits[0].URL != "https://go.dev/dl/" {
		t.Fatalf("TUI searcher: said %q, hits %+v, error %v", said, hits, err)
	}
	if n := tuiReviewStarts(t, wrapper); n != 1 {
		t.Fatalf("one search started %d wrappers, want 1", n)
	}
}

func TestClaudeTUISearchEveryAPI(t *testing.T) {
	for _, native := range []bool{false, true} {
		name := "search-api"
		if native {
			name = "no-search-api"
		}
		t.Run(name, func(t *testing.T) {
			fresh(t)
			wrapper := tuiReviewWrapper(t, "search")
			p := tuiAccount(t, wrapper)
			st := settings.Load()
			st.Searcher = "off"
			if err := settings.Save(st); err != nil {
				t.Fatal(err)
			}
			apis := newSearchAPIs(t)
			if !native {
				if err := provider.SetSearchAPI(apis.api("searxng", "")); err != nil {
					t.Fatal(err)
				}
			}
			s := tuiTestServer(t)
			tuiReviewSearch(t, s)
			for _, c := range []struct {
				from provider.Protocol
				body string
			}{
				{provider.Anthropic, `{"model":"claude-sonnet-5","max_tokens":128,"tools":[{"type":"web_search_20250305","name":"web_search"}],"messages":[{"role":"user","content":"latest go"}]}`},
				{provider.Chat, `{"model":"claude-sonnet-5","web_search_options":{},"messages":[{"role":"user","content":"latest go"}]}`},
				{provider.Responses, `{"model":"claude-sonnet-5","tools":[{"type":"web_search_preview"}],"input":"latest go"}`},
				{provider.Gemini, `{"tools":[{"googleSearch":{}}],"contents":[{"role":"user","parts":[{"text":"latest go"}]}]}`},
			} {
				rec := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/", strings.NewReader(c.body))
				code, msg := s.serveClaudeSubscription(rec, r, c.from, p, "claude-sonnet-5", []byte(c.body), &Usage{})
				want := "Go 1.27.1 is out"
				if native {
					want = "NATIVE_SEARCH_OK"
				}
				if code != 200 || !strings.Contains(rec.Body.String(), want) {
					t.Fatalf("%s: %d %s %s, want %q", c.from, code, msg, rec.Body, want)
				}
				if c.from == provider.Anthropic && !strings.Contains(rec.Body.String(), "https://go.dev/dl/") {
					t.Fatalf("Messages lost the search sources: %s", rec.Body)
				}
				waitFor(t, "the search run is reaped", func() bool { return !s.tui.Busy() })
			}
			apis.mu.Lock()
			n := len(apis.asked)
			apis.mu.Unlock()
			want := 4
			if native {
				want = 0
			}
			if n != want || tuiReviewStarts(t, wrapper) != 4 {
				t.Fatalf("API calls %d, wrapper starts %d; want %d and 4", n, tuiReviewStarts(t, wrapper), want)
			}
		})
	}
}

// A tool response stays parked after HTTP has ended, then belongs to the
// active HTTP continuation until it finishes. A shutdown clears the wait.
func TestClaudeTUIBusyTracksParkedTools(t *testing.T) {
	for _, stop := range []bool{false, true} {
		name := "continue"
		if stop {
			name = "shutdown"
		}
		t.Run(name, func(t *testing.T) {
			fresh(t)
			wrapper := tuiReviewWrapper(t, "tool")
			p := tuiAccount(t, wrapper)
			s := tuiTestServer(t)
			tuiReviewSearch(t, s)
			body := `{"model":"claude-sonnet-5","max_tokens":128,"tools":[{"name":"echo","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"latest go"}]}`
			writing, releaseWrite := make(chan struct{}), make(chan struct{})
			var writeGate sync.Once
			var writes int
			defer writeGate.Do(func() { close(releaseWrite) })
			serve := func(b string) *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				h := s.counted(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					s.serveClaudeSubscription(w, r, provider.Anthropic, p, "claude-sonnet-5", []byte(b), &Usage{})
				}))
				var out http.ResponseWriter = rec
				if writes == 0 {
					out = &tuiReviewHeldWriter{ResponseWriter: rec, writing: writing, release: releaseWrite}
				}
				writes++
				h.ServeHTTP(out, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(b)))
				return rec
			}
			firstDone := make(chan *httptest.ResponseRecorder, 1)
			go func() { firstDone <- serve(body) }()
			<-writing
			if b := s.Busy(); b.Requests != 1 || b.Bridge != 1 || b.Tools != 0 {
				t.Errorf("while the tool response is writing: %+v, want Requests=1 Bridge=1 Tools=0", b)
			}
			writeGate.Do(func() { close(releaseWrite) })
			first := <-firstDone
			if first.Code != 200 || !strings.Contains(first.Body.String(), `"tool_use"`) {
				t.Fatalf("tool response: %d %s", first.Code, first.Body)
			}
			if b := s.Busy(); b.Requests != 0 || b.Bridge != 0 || b.Tools != 1 || !b.Any() {
				t.Fatalf("after response, before result: %+v, want only Tools=1", b)
			}
			if stop {
				s.tui.Shutdown()
				if b := s.Busy(); b.Any() || b.Tools != 0 {
					t.Fatalf("after shutdown: %+v", b)
				}
				return
			}
			var reply map[string]json.RawMessage
			if err := json.Unmarshal(first.Body.Bytes(), &reply); err != nil {
				t.Fatal(err)
			}
			var req map[string]any
			if err := json.Unmarshal([]byte(body), &req); err != nil {
				t.Fatal(err)
			}
			var content any
			if err := json.Unmarshal(reply["content"], &content); err != nil {
				t.Fatal(err)
			}
			messages := req["messages"].([]any)
			messages = append(messages, map[string]any{"role": "assistant", "content": content},
				map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": "review-call", "content": "echoed"}}})
			req["messages"] = messages
			var wg sync.WaitGroup
			var final *httptest.ResponseRecorder
			wg.Add(1)
			go func() { defer wg.Done(); final = serve(string(mustJSON(req))) }()
			waitFor(t, "MCP accepts the tool result", func() bool {
				_, err := os.Stat(filepath.Join(filepath.Dir(wrapper), "result"))
				return err == nil
			})
			if b := s.Busy(); b.Requests != 1 || b.Bridge != 1 || b.Tools != 0 {
				t.Errorf("during continuation: %+v, want Requests=1 Bridge=1 Tools=0", b)
			}
			if err := os.WriteFile(filepath.Join(filepath.Dir(wrapper), "release"), nil, 0o600); err != nil {
				t.Fatal(err)
			}
			wg.Wait()
			if final.Code != 200 || !strings.Contains(final.Body.String(), "TOOL_DONE") {
				t.Fatalf("tool continuation: %d %s", final.Code, final.Body)
			}
			waitFor(t, "the tool run is reaped", func() bool { return !s.tui.Busy() })
			if b := s.Busy(); b.Any() || b.Bridge != 0 {
				t.Fatalf("after completion: %+v", b)
			}
		})
	}
}

type tuiReviewHeldWriter struct {
	http.ResponseWriter
	writing chan<- struct{}
	release <-chan struct{}
}

func (w *tuiReviewHeldWriter) Write(p []byte) (int, error) {
	close(w.writing)
	<-w.release
	return w.ResponseWriter.Write(p)
}
