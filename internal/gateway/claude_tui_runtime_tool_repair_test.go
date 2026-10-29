package gateway

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestClaudeTUIRuntimeBareToolRound(t *testing.T) {
	for _, parallel := range []bool{false, true} {
		t.Run(map[bool]string{false: "one", true: "parallel"}[parallel], func(t *testing.T) {
			work, config := t.TempDir(), t.TempDir()
			const sid = "bare-session"
			prefix := `{"type":"user","uuid":"user-row","sessionId":"bare-session","message":{"role":"user","content":"use the tools"}}` + "\n"
			bare := `{"type":"assistant","uuid":"bare-row","sessionId":"bare-session","message":{"id":"msg_011CgU8TZrBAxkvj66k2W2hm","role":"assistant","content":[{"type":"thinking","thinking":"signed reasoning","signature":"unchanged"},{"type":"tool_use","id":"toolu_011KUQdJfhuto5uwYuBqRbSA","name":"run_terminal_command","input":{"command":"pwd"}}]}}` + "\n"
			native := prefix + bare
			wire := bare
			if parallel {
				other := `{"type":"assistant","uuid":"other-row","sessionId":"bare-session","message":{"id":"msg_011CgU8TZrBAxkvj66k2W2hm","role":"assistant","content":[{"type":"tool_use","id":"call-read","name":"` + tuiRuntimeMcpToolPrefix() + `read_file","input":{"path":"file"}}]}}` + "\n"
				native += other
				wire += other
			}
			failed := `{"type":"user","uuid":"native-error","sessionId":"bare-session","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_011KUQdJfhuto5uwYuBqRbSA","is_error":true,"content":"No such tool available"}]}}` + "\n"
			native += failed
			wire += failed
			dir := tuiRuntimeClaudeSessions(config, work)[0]
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, sid+".jsonl")
			if err := os.WriteFile(path, []byte(native), 0600); err != nil {
				t.Fatal(err)
			}
			b := tuiRuntimeNewSubscriptionBridge()
			run := &tuiRuntimeSubscriptionRun{bridge: b, interactive: true, sessionID: sid, nativeWork: work, nativeConfigDir: config,
				tools: map[string]bool{"run_terminal_command": true, "read_file": true}}
			if parallel {
				b.toolJournal = &tuiRuntimeInteractiveToolJournal{path: filepath.Join(t.TempDir(), "tools.jsonl")}
				run.owner, run.outerSession, run.token = "owner", "conversation", "native-run"
			}
			var released atomic.Int32
			run.releaseWork = func(string) { released.Add(1) }
			ch := run.attach()
			done := make(chan struct{})
			go func() { run.readInteractiveOneTurn(strings.NewReader(wire)); close(done) }()
			var calls []string
			stops := 0
			for ev := range ch {
				switch ev.Kind {
				case KError:
					t.Fatal(ev.Text)
				case KToolStart:
					calls = append(calls, ev.ID)
				case KStop:
					if ev.Stop == "tool" {
						stops++
					}
				}
			}
			<-done
			wantCalls := 1
			if parallel {
				wantCalls++
			}
			if len(calls) != wantCalls || stops != 1 {
				t.Fatalf("calls=%v stops=%d", calls, stops)
			}
			repair := run.toolRepair
			if repair == nil {
				t.Fatal("bare call did not retain a continuation")
			}
			t.Cleanup(repair.discard)
			if released.Load() != 0 {
				t.Fatal("stateless cwd was released before its results")
			}
			if !repair.claim() || repair.claim() {
				t.Fatal("tool round must be claimed exactly once")
			}
			repair.expire()
			if released.Load() != 0 {
				t.Fatal("expiry released a continuation already claimed by the next request")
			}
			result := Message{Role: "user"}
			for _, id := range calls {
				result.Parts = append(result.Parts, Part{Kind: ToolResult, CallID: id, Text: "actual result"})
			}
			if parallel {
				if err := b.toolJournal.verifyResults(run, result.Parts); err != nil {
					t.Fatalf("repaired calls bypassed the existing tool journal: %v", err)
				}
			}
			id, err := repair.write(config, "claude-opus-5-5", []Message{{Role: "assistant"}, result})
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(dir, id+".jsonl"))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(string(data), strings.ReplaceAll(prefix, sid, id)) {
				t.Fatal("changed the prefix before the tool round")
			}
			for _, extra := range []string{"No such tool", tuiRuntimeLiteContinueText} {
				if strings.Contains(string(data), extra) {
					t.Fatalf("repaired transcript contains %q", extra)
				}
			}
			if !strings.Contains(string(data), `"signature":"unchanged"`) || !strings.Contains(string(data), `"name":"`+tuiRuntimeMcpToolPrefix()+`run_terminal_command"`) {
				t.Fatalf("repair lost thinking or kept the bare name: %s", data)
			}
			// the results, then liteResultsReceived: the copy is resumed
			// with a typed prompt, which a session ending in the user's
			// message would have Claude Code answer "No response requested."
			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			if len(lines) != wantCalls+3 || !strings.Contains(lines[len(lines)-1], tuiRuntimeLiteResultsReceived) {
				t.Fatalf("repair added other dialogue rows: %s", data)
			}
			lines = lines[:len(lines)-1]
			var last struct {
				Type    string
				Message struct {
					Content []struct {
						Type      string
						Content   []struct{ Type, Text string }
						ToolUseID string `json:"tool_use_id"`
					}
				}
			}
			if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil || last.Type != "user" {
				t.Fatalf("last row is not the actual results: %s", lines[len(lines)-1])
			}
			if len(last.Message.Content) != wantCalls {
				t.Fatal("result count changed")
			}
			for i, result := range last.Message.Content {
				if result.Type != "tool_result" || result.ToolUseID != calls[i] || len(result.Content) != 1 || result.Content[0].Text != "actual result" {
					t.Fatalf("result changed: %+v", result)
				}
			}
			original, _ := os.ReadFile(path)
			if string(original) != native {
				t.Fatal("repair edited the original native session")
			}
			repair.discard()
			repair.discard()
			if released.Load() != 1 {
				t.Fatal("cwd was not released exactly once")
			}
		})
	}
}

func TestClaudeTUIRuntimeBareToolMustBeOffered(t *testing.T) {
	for _, name := range []string{"missing", "mcp__another__run_terminal_command"} {
		t.Run(name, func(t *testing.T) {
			run := &tuiRuntimeSubscriptionRun{interactive: true, tools: map[string]bool{"run_terminal_command": true}}
			ch := run.attach()
			go run.readInteractiveOneTurn(strings.NewReader(`{"type":"assistant","message":{"id":"m","content":[{"type":"tool_use","id":"c","name":"` + name + `","input":{}}]}}` + "\n"))
			failed := false
			for ev := range ch {
				if ev.Kind == KToolStart {
					t.Fatal("unoffered tool was delivered")
				}
				failed = failed || ev.Kind == KError
			}
			if !failed || run.toolRepair != nil {
				t.Fatal("unoffered tool was treated as a name repair")
			}
		})
	}
}

// A repaired round's copy is resumed with the turn typed in. Claude Code's
// --reply-on-resume sends its first request before the MCP tools are
// listed, with none (au-exit 10-29: 28 tools became 0, and the turn ended
// on one line); a typed prompt waits for them. The copy holds the client's
// results and liteResultsReceived; the prompt is what the user said with
// the results, or liteContinueText.
func TestClaudeTUIRuntimeBareToolResumesWithTheTurnTyped(t *testing.T) {
	const call = "toolu_011KUQdJfhuto5uwYuBqRbSA"
	tools := `"system":"REPAIR_INSTRUCTIONS","tools":[{"name":"run_terminal_command","input_schema":{"type":"object"}}]`
	ask := `{"role":"user","content":"use the tools"}`
	for _, c := range []struct{ name, said, want string }{
		{"results alone", "", tuiRuntimeLiteContinueText},
		{"results and what the user said", `,{"type":"text","text":"只看 au-exit"}`, "只看 au-exit"},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, runs := tuiRuntimeScWrapper(t)
			rt, b := tuiRuntimeScRuntime(t)
			work := t.TempDir()
			const sid = "bare-session"
			native := `{"type":"user","uuid":"user-row","sessionId":"bare-session","message":{"role":"user","content":"use the tools"}}` + "\n"
			bare := `{"type":"assistant","uuid":"bare-row","sessionId":"bare-session","message":{"id":"msg_bare","role":"assistant","content":[{"type":"tool_use","id":"` + call + `","name":"run_terminal_command","input":{"command":"pwd"}}]}}` + "\n"
			failed := `{"type":"user","uuid":"native-error","sessionId":"bare-session","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"` + call + `","is_error":true,"content":"No such tool available"}]}}` + "\n"
			dir := tuiRuntimeClaudeSessions("", work)[0]
			if err := os.MkdirAll(dir, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, sid+".jsonl"), []byte(native+bare+failed), 0o600); err != nil {
				t.Fatal(err)
			}
			first := tuiRuntimeErParse(t, ask)
			first.System, first.Tools = "REPAIR_INSTRUCTIONS", []Tool{{Name: "run_terminal_command", Schema: json.RawMessage(`{"type":"object"}`)}}
			run := &tuiRuntimeSubscriptionRun{bridge: b, interactive: true, sessionID: sid, nativeWork: work, owner: tuiRuntimeScOwner,
				model: "claude-opus-5-5", scope: tuiRuntimeErScope, tools: map[string]bool{"run_terminal_command": true}}
			run.interactiveContextSnapshot = tuiRuntimeTestContextSnapshot(first)
			run.heard(first.Messages)
			ch := run.attach()
			done := make(chan struct{})
			go func() { run.readInteractiveOneTurn(strings.NewReader(bare + failed)); close(done) }()
			for range ch {
			}
			<-done
			if run.toolRepair == nil {
				t.Fatal("bare call did not retain a continuation")
			}
			b.noteLiteParked(run)
			code, out := tuiRuntimeScServe(t, rt, "", `{"model":"claude-opus-5-5","max_tokens":64,`+tools+`,"messages":[`+ask+
				`,{"role":"assistant","content":[{"type":"tool_use","id":"`+call+`","name":"run_terminal_command","input":{"command":"pwd"}}]}`+
				`,{"role":"user","content":[{"type":"tool_result","tool_use_id":"`+call+`","content":"ACTUAL_RESULT"}`+c.said+`]}]}`)
			if code != 200 || !strings.Contains(out, "SC_ANSWER") {
				t.Fatalf("repaired round: %d %s", code, out)
			}
			got := tuiRuntimeOnlyRun(t, runs())
			if got["resume"] == "" || got["resume"] == sid || strings.Contains(got["argv"], "--reply-on-resume") {
				t.Fatalf("did not resume a copy with the turn typed: %v", got)
			}
			if !strings.Contains(got["stdin"], c.want) || strings.Contains(got["stdin"], "ACTUAL_RESULT") {
				t.Fatalf("turn prompt: %q", got["stdin"])
			}
			seed := got["seed.jsonl"]
			for _, kept := range []string{`"name":"` + tuiRuntimeMcpToolPrefix() + `run_terminal_command"`, "ACTUAL_RESULT", tuiRuntimeLiteResultsReceived} {
				if !strings.Contains(seed, kept) {
					t.Fatalf("resumed copy lacks %s:\n%s", kept, seed)
				}
			}
			for _, gone := range []string{"No such tool", "只看 au-exit", tuiRuntimeLiteContinueText} {
				if strings.Contains(seed, gone) {
					t.Fatalf("resumed copy holds %s:\n%s", gone, seed)
				}
			}
		})
	}
}
