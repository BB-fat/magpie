package gateway

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestClaudeTUIRuntimeLiteChangedHistoryKeepsCheckpoint(t *testing.T) {
	for _, tc := range []struct{ name, history string }{
		{"single_message", `[{"role":"user","content":"Configuration checked. Run the tests next."}]`},
		{"multiple_messages", `[{"role":"user","content":"Configuration checked."},{"role":"assistant","content":"Ready to continue."},{"role":"user","content":"Run the tests next."}]`},
	} {
		for _, restart := range []bool{false, true} {
			name := "live"
			if restart {
				name = "restart"
			}
			t.Run(tc.name+"/"+name, func(t *testing.T) {
				_, runs := tuiRuntimeScWrapper(t)
				rt, b := tuiRuntimeScRuntime(t)
				tuiRuntimeScSaveClean(t, b, "inner-before-compaction", false)
				before, _, _ := b.loadInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter)
				// The old test only seeded disk; the failing route depends on
				// the first user message remembered by a running gateway.
				b.seedOuterThread(tuiRuntimeScOwner, tuiRuntimeScOuter, before.ThreadKey)
				body := `{"model":"claude-opus-5-5","max_tokens":64,"messages":` + tc.history + `}`
				if code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, body); code != 200 || !strings.Contains(out, "SC_ANSWER") {
					t.Fatalf("changed history: %d %s", code, out)
				}
				b.abortAll()
				saved, ok, err := b.loadInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter)
				if err != nil || !ok || saved.Dirty || saved.SessionID == before.SessionID || saved.ConvKey == before.ConvKey {
					t.Fatalf("changed history did not replace checkpoint: %+v ok=%t err=%v", saved, ok, err)
				}
				first := tuiRuntimeOnlyRun(t, runs())
				if first["resume"] == before.SessionID {
					t.Fatal("changed history resumed the old context")
				}
				if restart {
					rt, b = tuiRuntimeScRuntime(t)
				}
				var msgs []map[string]any
				if err := json.Unmarshal([]byte(tc.history), &msgs); err != nil {
					t.Fatal(err)
				}
				msgs = append(msgs, map[string]any{"role": "assistant", "content": "SC_ANSWER"}, map[string]any{"role": "user", "content": "Report the outcome."})
				next, err := json.Marshal(map[string]any{"model": "claude-opus-5-5", "max_tokens": 64, "messages": msgs})
				if err != nil {
					t.Fatal(err)
				}
				if code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, string(next)); code != 200 || !strings.Contains(out, "SC_ANSWER") {
					t.Fatalf("continuation: %d %s", code, out)
				}
				b.abortAll()
				got := runs()
				if len(got) != 2 {
					t.Fatalf("native starts = %d, want 2", len(got))
				}
				for _, run := range got {
					if strings.Contains(run["stdin"], "Report the outcome.") {
						if run["resume"] != saved.SessionID || run["cwd"] != first["cwd"] || strings.Contains(run["stdin"], "Configuration checked") {
							t.Fatalf("continuation rebuilt instead of resuming only the suffix: %v", run)
						}
						return
					}
				}
				t.Fatal("no native continuation prompt")
			})
		}
	}
}
