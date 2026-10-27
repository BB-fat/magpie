package gateway

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeTUIRuntimeLiteInstructionIdentityControlsResume(t *testing.T) {
	for _, tc := range []struct {
		name, system, settings string
		rebuild                bool
	}{
		{"tools_removed", "stable instructions", ``, false},
		{"tools_changed", "stable instructions", `,"tools":[{"name":"new_tool","input_schema":{"type":"object"}}]`, false},
		{"effort_changed", "stable instructions", `,"thinking":{"type":"adaptive"},"output_config":{"effort":"high"}`, false},
		{"schema_changed", "stable instructions", `,"output_config":{"format":{"type":"json_schema","schema":{"type":"string"}}}`, false},
		{"instructions_changed", "replacement instructions", ``, true},
		{"instructions_removed", "", ``, true},
		{"tool_choice_forced", "stable instructions", `,"tool_choice":{"type":"any"},"tools":[{"name":"new_tool","input_schema":{"type":"object"}}]`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logs, runs := tuiRuntimeScWrapper(t)
			// Record the actual MCP definitions and launch options, not just
			// the request that the bridge intended to apply.
			wrapper := filepath.Join(filepath.Dir(logs), "sc-bridge")
			script, err := os.ReadFile(wrapper)
			if err != nil {
				t.Fatal(err)
			}
			script = []byte(strings.Replace(string(script), `prev="$a"`, `case "$a" in
  *tools.json*) tools=$(printf '%s' "$a" | sed 's/.*"\([^"]*\/tools.json\)".*/\1/'); cp "$tools" "$SC_LOG/../tools-captured.json" ;;
  */schema.json) cp "$a" "$SC_LOG/../schema-captured.json" ;;
esac
prev="$a"`, 1))
			script = []byte(strings.ReplaceAll(string(script), `"stop_reason":"end_turn"}`, `"stop_reason":"end_turn","structured_output":"SC_ANSWER"}`))
			if err := os.WriteFile(wrapper, script, 0o700); err != nil {
				t.Fatal(err)
			}
			rt, b := tuiRuntimeScRuntime(t)
			base, err := tuiRuntimeParseRequest([]byte(`{"model":"claude-opus-5-5","system":"stable instructions","tools":[{"name":"old_tool","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"hello"},{"role":"assistant","content":"previous reply"}]}`))
			if err != nil {
				t.Fatal(err)
			}
			if err := b.saveInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter, tuiRuntimeInteractiveSessionEntry{
				SessionID: "original-native", Model: base.Model,
				ConvKey:  tuiRuntimeConversationKeys(tuiRuntimeScOwner, base.Messages)[0],
				ReplyKey: tuiRuntimeAssistantReplyKey(base.Messages[1]), InstructionKey: tuiRuntimeInstructionKey(base),
			}); err != nil {
				t.Fatal(err)
			}
			system, _ := json.Marshal(tc.system)
			body := `{"model":"claude-opus-5-5","system":` + string(system) + tc.settings + `,"messages":[{"role":"user","content":"hello"},{"role":"assistant","content":"previous reply"},{"role":"user","content":"summarize this conversation"}]}`
			if code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, body); code != 200 || !strings.Contains(out, "SC_ANSWER") {
				t.Fatalf("request: %d %s", code, out)
			}
			b.abortAll()
			first := tuiRuntimeOnlyRun(t, runs())
			if (first["resume"] != "original-native") != tc.rebuild {
				t.Fatalf("rebuild=%t, native resume=%q", tc.rebuild, first["resume"])
			}
			// A forced tool choice binds the current request only.
			if strings.Contains(first["stdin"], "<current_turn_constraint>") != strings.Contains(tc.settings, "tool_choice") {
				t.Fatalf("turn constraint: %s", first["stdin"])
			}
			if strings.Contains(first["stdin"], "external_context_refresh") || (!tc.rebuild && strings.Contains(first["stdin"], "stable instructions")) {
				t.Fatalf("redundant instructions: %s", first["stdin"])
			}
			if tc.rebuild && (strings.Contains(first["stdin"], "stable instructions") && tc.system != "stable instructions" || !strings.Contains(first["stdin"], tc.system)) {
				t.Fatalf("rebuild did not use current instructions: %s", first["stdin"])
			}
			tools, err := os.ReadFile(filepath.Join(filepath.Dir(logs), "tools-captured.json"))
			if err != nil || strings.Contains(string(tools), "old_tool") || strings.Contains(tc.settings, "new_tool") && !strings.Contains(string(tools), "new_tool") {
				t.Fatalf("current tools not applied: %s (%v)", tools, err)
			}
			if tc.name == "effort_changed" && !strings.Contains(first["argv"], "--effort\nhigh") {
				t.Fatalf("current effort not applied: %s", first["argv"])
			}
			if tc.name == "schema_changed" {
				schema, err := os.ReadFile(filepath.Join(filepath.Dir(logs), "schema-captured.json"))
				if err != nil || string(schema) != `{"type":"string"}` {
					t.Fatalf("current schema not applied: %s (%v)", schema, err)
				}
			}
			saved, ok, err := b.loadInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter)
			parsed, parseErr := tuiRuntimeParseRequest([]byte(body))
			if parseErr != nil || err != nil || !ok || saved.Dirty || saved.InstructionKey != tuiRuntimeInstructionKey(parsed) {
				t.Fatalf("checkpoint: %+v err=%v parse=%v", saved, err, parseErr)
			}
			// A fresh runtime must use the persisted instruction identity,
			// not refresh again because the old runtime's settings are gone.
			rt, b = tuiRuntimeScRuntime(t)
			reply := "SC_ANSWER"
			if tc.name == "schema_changed" {
				reply = `"SC_ANSWER"`
			}
			encodedReply, _ := json.Marshal(reply)
			body = strings.TrimSuffix(body, `]}`) + `,{"role":"assistant","content":` + string(encodedReply) + `},{"role":"user","content":"continue after restart"}]}`
			if code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, body); code != 200 || !strings.Contains(out, "SC_ANSWER") {
				t.Fatalf("restart: %d %s", code, out)
			}
			b.abortAll()
			for _, run := range runs() {
				if strings.Contains(run["stdin"], "continue after restart") {
					if run["resume"] != saved.SessionID || strings.Contains(run["stdin"], "instructions") {
						t.Fatalf("restart did not resume only the suffix: %v", run)
					}
					return
				}
			}
			t.Fatal("missing continuation")
		})
	}
}

func TestClaudeTUIRuntimeLiteRejectsLegacyCheckpoints(t *testing.T) {
	_, runs := tuiRuntimeScWrapper(t)
	rt, b := tuiRuntimeScRuntime(t)
	path := tuiRuntimeInteractiveSessionStatePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	legacy := []byte(`{"version":1,"sessions":{}}`)
	if err := os.WriteFile(path, legacy, 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _ := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, `{"model":"claude-opus-5-5","messages":[{"role":"user","content":"hello"}]}`); code != 500 {
		t.Fatalf("legacy store returned %d, want 500", code)
	}
	if len(runs()) != 0 {
		t.Fatal("legacy store started a native run")
	}
	if err := b.saveInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter, tuiRuntimeInteractiveSessionEntry{SessionID: "new", ConvKey: "history", InstructionKey: "instructions"}); err == nil {
		t.Fatal("legacy store silently overwritten")
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(legacy) {
		t.Fatalf("legacy store modified: %s, %v", got, err)
	}
}

func TestClaudeTUIRuntimeReadOnlyRecoveryKeepsInstructionKey(t *testing.T) {
	req, err := tuiRuntimeParseRequest([]byte(`{"model":"claude-opus-5-5","system":"rules","tool_choice":{"type":"any"},"tools":[{"name":"write_file","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"hello"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if tuiRuntimeInstructionKey(tuiRuntimeInteractiveSanitizeReadOnly(req)) != tuiRuntimeInstructionKey(req) {
		t.Fatal("dropping tool_choice for read-only recovery changed the instruction key")
	}
	prompt, err := tuiRuntimeRenderInteractiveFirstPrompt(req)
	if err != nil {
		t.Fatal(err)
	}
	instructions, constraint := strings.Index(prompt, "</external_system_instructions>"), strings.Index(prompt, "<current_turn_constraint>")
	if instructions < 0 || constraint < instructions || !strings.Contains(prompt, "You must call at least one available tool") {
		t.Fatalf("tool choice not rendered as a turn constraint after the instructions: %q", prompt)
	}
}
