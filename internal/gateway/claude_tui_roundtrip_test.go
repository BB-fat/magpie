package gateway

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/claudebridge"
	"github.com/yetone/magpie/internal/provider"
)

func tuiRequestBody(t *testing.T, from provider.Protocol, req *Request) []byte {
	t.Helper()
	switch from {
	case provider.Anthropic:
		return buildAnthropic(req, req.Model)
	case provider.Chat:
		return buildChat(req, req.Model, "", false)
	case provider.Responses:
		return buildResponses(req, req.Model, "", false)
	case provider.Gemini:
		b, err := buildGemini(req, req.Model)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	t.Fatalf("unknown protocol %s", from)
	return nil
}

func TestClaudeTUIToolRoundTripEveryAPI(t *testing.T) {
	for _, from := range []provider.Protocol{provider.Anthropic, provider.Chat, provider.Responses, provider.Gemini} {
		t.Run(string(from), func(t *testing.T) {
			name := "spawn.agent"
			offered := claudebridge.MCPToolPrefix() + claudebridge.ToolName(name)
			callRow := `{"type":"assistant","uuid":"tool-row","session_id":"inner","message":{"id":"msg_tool","model":"claude-sonnet-5","usage":{"input_tokens":11,"cache_creation_input_tokens":29},"content":[{"type":"tool_use","id":"toolu_roundtrip","name":"` + offered + `","input":{"task":"hello"}}]}}`
			wrapper := tuiWrapper(t, callRow)
			gate := filepath.Join(filepath.Dir(wrapper), "answered")
			script, err := os.ReadFile(wrapper)
			if err != nil {
				t.Fatal(err)
			}
			script = append(script, []byte("while [ ! -f "+gate+" ]; do sleep 0.01; done\nprintf '%s\\n' '"+`{"type":"assistant","uuid":"answer-row","session_id":"inner","message":{"id":"msg_answer","model":"claude-sonnet-5","usage":{"input_tokens":3,"cache_read_input_tokens":67},"content":[{"type":"text","text":"DONE"}]}}`+"' '"+`{"type":"result","session_id":"inner","stop_reason":"end_turn"}`+"'\n")...)
			if err := os.WriteFile(wrapper, script, 0o700); err != nil {
				t.Fatal(err)
			}
			p := tuiAccount(t, wrapper)
			s := tuiTestServer(t)
			t.Cleanup(func() { s.tui.Shutdown() })
			req := &Request{Model: "claude-sonnet-5", Safeguards: json.RawMessage(`{"id":"same"}`), Tools: []Tool{{Name: name, Schema: json.RawMessage(`{"type":"object","properties":{"task":{"type":"string"}},"required":["task"]}`)}}, Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "do it"}}}}}
			body := tuiRequestBody(t, from, req)
			out := httptest.NewRecorder()
			var usage Usage
			code, msg := s.serveClaudeSubscription(out, httptest.NewRequest("POST", "/", strings.NewReader(string(body))), from, p, req.Model, body, &usage)
			if code != 200 {
				t.Fatalf("tool call: %d %s %s", code, msg, out.Body)
			}
			if !strings.Contains(out.Body.String(), name) || !strings.Contains(out.Body.String(), "hello") || strings.Contains(out.Body.String(), offered) {
				t.Fatalf("tool name or args changed: %s", out.Body)
			}
			if usage.Input != 11 || usage.CacheWrite != 29 {
				t.Fatalf("tool usage changed: %+v", usage)
			}
			args, err := os.ReadFile(filepath.Join(filepath.Dir(wrapper), "args"))
			if err != nil {
				t.Fatal(err)
			}
			var config struct {
				Servers map[string]struct {
					Args []string `json:"args"`
				} `json:"mcpServers"`
			}
			for _, arg := range strings.Split(string(args), "\n") {
				if strings.HasPrefix(arg, `{"mcpServers"`) {
					if err := json.Unmarshal([]byte(arg), &config); err != nil {
						t.Fatal(err)
					}
				}
			}
			cfg := config.Servers[claudebridge.MCPServerName()]
			if len(cfg.Args) < 2 {
				t.Fatalf("missing MCP callback: %s", args)
			}
			callback, err := os.ReadFile(cfg.Args[len(cfg.Args)-2])
			if err != nil {
				t.Fatal(err)
			}
			token := string(callback[strings.LastIndex(string(callback), "/")+1:])
			result := make(chan string, 1)
			go func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", string(callback), strings.NewReader(`{"tool_call_id":"toolu_roundtrip","name":"`+claudebridge.ToolName(name)+`","arguments":{"task":"hello"}}`))
				if !s.tui.ServeMCP(w, r, token) || w.Code != 200 {
					result <- fmt.Sprintf("callback failed: %d %s", w.Code, w.Body)
					return
				}
				if err := os.WriteFile(gate, nil, 0o600); err != nil {
					result <- "release failed: " + err.Error()
					return
				}
				result <- w.Body.String()
			}()
			req.Messages = append(req.Messages,
				Message{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "toolu_roundtrip", Name: name, Args: json.RawMessage(`{"task":"hello"}`)}}},
				Message{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "toolu_roundtrip", Name: name, Text: "RESULT_BYTES", IsError: true}}})
			body = tuiRequestBody(t, from, req)
			out = httptest.NewRecorder()
			usage = Usage{}
			code, msg = s.serveClaudeSubscription(out, httptest.NewRequest("POST", "/", strings.NewReader(string(body))), from, p, req.Model, body, &usage)
			if code != 200 || !strings.Contains(out.Body.String(), "DONE") {
				t.Fatalf("valid converted result failed: %d %s %s", code, msg, out.Body)
			}
			if usage.CacheRead != 67 {
				t.Fatalf("cache-read usage lost after tool result: %+v", usage)
			}
			got := <-result
			if !strings.Contains(got, "RESULT_BYTES") {
				t.Fatalf("tool result changed: %s; answer %d %s", got, code, out.Body)
			}
			if (from == provider.Anthropic || from == provider.Gemini) && !strings.Contains(got, `"is_error":true`) {
				t.Fatalf("native tool error flag lost: %s", got)
			}
		})
	}
}

func TestClaudeTUICacheUsageEveryAPI(t *testing.T) {
	rows := []string{
		`{"type":"assistant","uuid":"r1","session_id":"s1","message":{"id":"m1","model":"claude-sonnet-5","usage":{"input_tokens":3,"output_tokens":1,"cache_read_input_tokens":11,"cache_creation_input_tokens":5,"cache_creation":{"ephemeral_5m_input_tokens":2,"ephemeral_1h_input_tokens":3}},"content":[{"type":"text","text":"first "}]}}`,
		`{"type":"assistant","uuid":"r2","session_id":"s1","message":{"id":"m1","model":"claude-sonnet-5","usage":{"input_tokens":3,"output_tokens":2,"cache_read_input_tokens":11,"cache_creation_input_tokens":5,"cache_creation":{"ephemeral_5m_input_tokens":2,"ephemeral_1h_input_tokens":3}},"content":[{"type":"text","text":"part "}]}}`,
		`{"type":"assistant","uuid":"r3","session_id":"s1","message":{"id":"m2","model":"claude-sonnet-5","usage":{"input_tokens":7,"output_tokens":4,"cache_read_input_tokens":13,"cache_creation_input_tokens":35,"cache_creation":{"ephemeral_5m_input_tokens":26,"ephemeral_1h_input_tokens":9}},"content":[{"type":"text","text":"done"}]}}`,
		`{"type":"result","session_id":"s1","stop_reason":"end_turn"}`,
	}
	p := tuiAccount(t, tuiWrapper(t, rows...))
	s := tuiTestServer(t)
	cases := []struct {
		from provider.Protocol
		body string
	}{
		{provider.Anthropic, `{"messages":[{"role":"user","content":"go"}]}`},
		{provider.Chat, `{"messages":[{"role":"user","content":"go"}]}`},
		{provider.Responses, `{"input":"go"}`},
		{provider.Gemini, `{"contents":[{"role":"user","parts":[{"text":"go"}]}]}`},
	}
	for _, c := range cases {
		for _, stream := range []bool{false, true} {
			t.Run(string(c.from)+map[bool]string{false: "/json", true: "/stream"}[stream], func(t *testing.T) {
				var m map[string]any
				_ = json.Unmarshal([]byte(c.body), &m)
				m["stream"] = stream
				b, _ := json.Marshal(m)
				var u Usage
				rec := httptest.NewRecorder()
				code, msg := s.serveClaudeSubscription(rec, httptest.NewRequest("POST", "/", strings.NewReader(string(b))), c.from, p, "claude-sonnet-5", b, &u)
				if code != 200 {
					t.Fatalf("%d %s %s", code, msg, rec.Body)
				}
				if u.Input != 10 || u.Output != 6 || u.CacheRead != 24 || u.CacheWrite != 40 || u.CacheWrite1h != 12 {
					t.Fatalf("usage mismatch: %+v", u)
				}
			})
		}
	}
}
