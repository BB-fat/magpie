package gateway

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func TestClaudeTUIRuntimeRealBareTool(t *testing.T) {
	if os.Getenv("MAGPIE_CLAUDE_TUI_REAL") != "1" {
		t.Skip("set MAGPIE_CLAUDE_TUI_REAL=1 and the real programs")
	}
	for _, outer := range []string{"", "repair-conversation"} {
		t.Run("session="+outer, func(t *testing.T) {
			up, profile, bridge := tuiRuntimeRealSetup(t, "magpie")
			b := bridge()
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			prefix := `{"model":"` + tuiRuntimeRealModel + `","tools":` + tuiRuntimeRealTools + `,"stream":true,"messages":[{"role":"user","content":"call bare echo with hi"}`
			out, code, why := tuiRuntimeRealServe(t, b, ctx, profile, outer, prefix+`]}`)
			if code != 200 || out.tool().id == "" {
				t.Fatalf("bare tool: %d %s call=%v", code, why, out.tool())
			}
			call := out.tool()
			if call.id != "toolu_standin_bare" || call.name != "echo" || call.args != `{"text":"hi"}` {
				t.Fatalf("rewrote the client's call: %+v", call)
			}
			next := prefix + `,{"role":"assistant","content":[{"type":"tool_use","id":"toolu_standin_bare","name":"echo","input":{"text":"hi"}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_standin_bare","content":"ECHOED"}]}]}`
			b.rt.bridge.mu.Lock()
			parked := b.rt.bridge.liteParked[call.id]
			b.rt.bridge.mu.Unlock()
			if parked == nil || parked.toolRepair == nil {
				t.Fatal("bare call has no transcript continuation")
			}
			out, code, why = tuiRuntimeRealServe(t, b, ctx, profile, outer, next)
			if code != 200 || out.text() != "DONE: ECHOED" {
				t.Fatalf("tool results: %d %s %q events=%v asked=%s", code, why, out.text(), out.kinds(), up.lastAsked())
			}
			up.mu.Lock()
			requests := append([]json.RawMessage(nil), up.requests...)
			up.mu.Unlock()
			found := false
			for _, raw := range requests {
				if !strings.Contains(string(raw), "ECHOED") {
					continue
				}
				found = true
				for _, extra := range []string{tuiRuntimeLiteResultsReceived, tuiRuntimeLiteContinueText, "No response requested.", "No such tool"} {
					if strings.Contains(string(raw), extra) {
						t.Fatalf("repaired history added %q: %s", extra, raw)
					}
				}
				if !strings.Contains(string(raw), `"name":"mcp__magpie__echo"`) {
					t.Fatalf("native history did not normalize the tool: %s", raw)
				}
			}
			if !found {
				t.Fatal("Claude did not receive the real tool result")
			}
			tuiRuntimeWaitGone(t, b, profile)
			third := strings.TrimSuffix(next, `]}`) + `,{"role":"assistant","content":"DONE: ECHOED"},{"role":"user","content":"What did the tool return?"}]}`
			out, code, why = tuiRuntimeRealServe(t, b, ctx, profile, outer, third)
			if code != 200 || out.text() != "DONE: ECHOED" {
				t.Fatalf("following message: %d %s %q", code, why, out.text())
			}
			tuiRuntimeWaitGone(t, b, profile)
		})
	}
}
