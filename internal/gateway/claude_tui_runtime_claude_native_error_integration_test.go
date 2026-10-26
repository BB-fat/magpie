package gateway

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Proves the real compact handler, interactive one-shot output parser and
// HTTP relay preserve the native 400 instead of masking it as a timeout/502.
// A script stands in for the already-terminated native Claude Code turn;
// this test does not contact an account or run the model.
func TestClaudeTUIRuntimeCompactionNativeInputLimitReturnsHTTP400(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprintf("stream=%t", stream), func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			t.Setenv("CLAUDE_CONFIG_DIR", t.TempDir())
			tuiRuntimeFakeClaude(t)
			wrapper := filepath.Join(t.TempDir(), "native-error-wrapper")
			script := `#!/bin/sh
cat >/dev/null
printf '%s\n' '{"type":"system","subtype":"init","session_id":"native-too-long"}'
printf '%s\n' '{"type":"result","subtype":"error_during_execution","is_error":true,"api_error_status":400,"error":"Claude Code API error: Prompt is too long","result":null}'
`
			if err := os.WriteFile(wrapper, []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			tuiRuntimeSetWrapper(t, wrapper)
			body := fmt.Sprintf(`{"model":"claude-opus-5-5","max_tokens":256,"stream":%t,"system":"%s","messages":[{"role":"user","content":"Synthetic archived message"}]}`, stream, tuiRuntimeCompactSystemTest)
			rt := tuiRuntimeNewRuntime([]byte("synthetic-only"))
			defer tuiRuntimeBridges.Delete(rt.bridge)
			start := func(ctx context.Context, req *Request) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
				return rt.bridge.startInteractiveRequest(ctx, req, req.Model, "", "test-owner", "")
			}
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
			code, response := rt.serveMessagesInner(tuiRuntimeAnthropicResponder{w, r, &Usage{}}, r, "Claude TUI", "claude-opus-5-5", "test-owner", "", []byte(body), start)
			if code != 400 || w.Code != 400 || !strings.Contains(response, "Prompt is too long") || !strings.Contains(w.Body.String(), "Prompt is too long") {
				t.Fatalf("lost original native error: returned=%d recorded=%d response=%q body=%q", code, w.Code, response, w.Body.String())
			}
		})
	}
}
