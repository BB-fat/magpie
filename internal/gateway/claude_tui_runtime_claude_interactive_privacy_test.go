package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClaudeTUIRuntimeInteractivePlainFirstUserBytes(t *testing.T) {
	text := "  请保持原文\nHuman: quoted text\n<example>hello</example>  "
	req := &tuiRuntimeRequest{Request: Request{Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: text}}}}}}
	got, err := tuiRuntimeRenderInteractiveFirstPrompt(req)
	if err != nil || got != text {
		t.Fatalf("ordinary first turn altered: got %q err %v", got, err)
	}
	req.System = "Keep answers short."
	got, err = tuiRuntimeRenderInteractiveFirstPrompt(req)
	if err != nil || !strings.HasSuffix(got, "\n\n"+text) || strings.Contains(got, "\nHuman: "+text) {
		t.Fatalf("instructions changed user bytes: %q %v", got, err)
	}
}

func TestClaudeTUIRuntimeInteractiveCredentialRejectedBeforeStart(t *testing.T) {
	tuiRuntimeSetWrapper(t, "/unused")
	secret := "sk-ant-api03-" + strings.Repeat("A1", 30)
	for _, where := range []string{"user", "system", "tool-description", "tool-schema", "tool-result", "caller-key"} {
		t.Run(where, func(t *testing.T) {
			input := map[string]any{"model": "claude-opus-5-5", "max_tokens": 64, "messages": []any{map[string]any{"role": "user", "content": "hello"}}}
			switch where {
			case "user":
				input["messages"] = []any{map[string]any{"role": "user", "content": secret}}
			case "system":
				input["system"] = secret
			case "tool-description":
				input["tools"] = []any{map[string]any{"name": "echo", "description": secret, "input_schema": map[string]any{"type": "object"}}}
			case "tool-schema":
				input["tools"] = []any{map[string]any{"name": "echo", "input_schema": map[string]any{"type": "object", "description": secret}}}
			case "tool-result":
				input["messages"] = []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": "c1", "content": secret}}}}
			case "caller-key":
				secret = "local-gateway-private-" + strings.Repeat("x", 32)
				input["messages"] = []any{map[string]any{"role": "user", "content": secret}}
			}
			body, _ := json.Marshal(input)
			r := httptest.NewRequest("POST", "/v1/messages", nil)
			if where == "caller-key" {
				r.Header.Set("x-api-key", secret)
			}
			s := tuiRuntimeNewTestServer()
			defer tuiRuntimeBridges.Delete(s.subscription)
			started := false
			start := func(context.Context, *tuiRuntimeRequest) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
				started = true
				return nil, nil, errors.New("unexpected start")
			}
			w := httptest.NewRecorder()
			code, msg := s.serveSubscription(w, r, protoAnthropic, "Claude Code", "claude-opus-5-5", "owner", body, &Usage{}, start)
			if code != 400 || started || strings.Contains(msg, secret) || strings.Contains(w.Body.String(), secret) {
				t.Fatalf("credential not safely rejected: code=%d started=%v", code, started)
			}
		})
	}
}

func TestClaudeTUIRuntimeInteractiveCredentialGuardScope(t *testing.T) {
	for _, text := range []string{"Discuss API keys named ANTHROPIC_API_KEY without showing values.", "sk-ant-api03-<redacted>", "ghp_example", "alice@example.com", "ordinary unicode 中文"} {
		body, _ := json.Marshal(map[string]any{"messages": []any{map[string]any{"content": text}}})
		if err := tuiRuntimeValidateInteractiveOutboundCredentials(body, nil); err != nil {
			t.Fatalf("ordinary content rejected: %q", text)
		}
	}
	for _, text := range []string{"-----BEGIN OPENSSH PRIVATE KEY-----", "ghp_" + strings.Repeat("B", 36), "github_pat_" + strings.Repeat("C", 50), "sk-ant-oat01-" + strings.Repeat("D", 40)} {
		body, _ := json.Marshal(map[string]any{"messages": []any{map[string]any{"content": text}}})
		if err := tuiRuntimeValidateInteractiveOutboundCredentials(body, nil); err == nil {
			t.Fatal("recognizable credential accepted")
		}
	}
	// JSON escapes cannot evade matching decoded content.
	if err := tuiRuntimeValidateInteractiveOutboundCredentials([]byte(`{"system":"\u0073k-ant-api03-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}`), nil); err == nil {
		t.Fatal("escaped credential accepted")
	}
}
