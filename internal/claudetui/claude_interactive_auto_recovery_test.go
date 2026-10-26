//go:build darwin || linux

package claudetui

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// Keep these fixtures self-contained: the community plugin package ships
// the recovery tests but not the much larger upstream lifecycle test file.
const autoTestModel = "claude-opus-5-5"

func arDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	return dir
}

func arWrapper(t *testing.T, dir, name, script string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	setWrapper(t, path)
}

func arAnswerScript() string {
	return "#!/bin/sh\n" +
		"cat >/dev/null\n" +
		"printf '%s\\n' '{\"type\":\"assistant\",\"uuid\":\"a1\",\"session_id\":\"inner-ol\",\"message\":{\"id\":\"m1\",\"model\":\"claude-opus-5-5\",\"content\":[{\"type\":\"text\",\"text\":\"OL_ANSWER\"}]}}'\n" +
		"printf '%s\\n' '{\"type\":\"result\",\"session_id\":\"inner-ol\",\"stop_reason\":\"end_turn\"}'\n"
}

func arSave(t *testing.T, b *subscriptionBridge, owner, outer, session string, dirty bool) {
	t.Helper()
	msgs := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "先读一下配置"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "配置已读"}}},
	}
	req := &Request{Model: autoTestModel, Messages: msgs}
	keys := conversationKeys(owner, msgs)
	if err := b.saveInteractiveSession(owner, outer, interactiveSessionEntry{
		SessionID: session, Model: autoTestModel, ThreadKey: firstWords(req),
		ConvKey: keys[0], ReplyKey: assistantReplyKey(msgs[1]),
		ContextKey: turnKey(owner, req, nil), Dirty: dirty,
	}); err != nil {
		t.Fatal(err)
	}
}

func arServe(t *testing.T, rt *runtimeTUI, b *subscriptionBridge, owner, outer string, body []byte) (int, string, string) {
	t.Helper()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body)))
	code, msg := rt.serveMessagesInner(anthropicResponder{w, r, &Usage{}}, r, "Claude Code", autoTestModel, owner, outer, body, func(ctx context.Context, req *Request) (*subscriptionRun, <-chan Event, error) {
		return b.startInteractiveRequest(ctx, req, autoTestModel, "", owner, outer)
	})
	return code, msg, w.Body.String()
}

func autoRecoveryFixture(t *testing.T) (*subscriptionBridge, *runtimeTUI, string, string, string) {
	t.Helper()
	dir := arDir(t)
	t.Setenv("MAGPIE_CLAUDE_DURABLE_TOOL_JOURNAL", "1")
	t.Setenv("MAGPIE_CLAUDE_AUTO_RECOVER_ORPHANS", "1")
	t.Setenv("MAGPIE_CLAUDE_NATIVE_TMUX_GUARD", "1")
	// Never query or modify the user's real tmux server. Fake only the
	// has-session status of the exact synthetic prior native run.
	if err := os.WriteFile(filepath.Join(dir, "tmux"), []byte("#!/bin/sh\nif [ \"$MAGPIE_TEST_TMUX_ALIVE\" = 1 ]; then exit 0; fi\necho \"can't find session\" >&2\nexit 1\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	runLog := filepath.Join(dir, "wrapper-launches")
	t.Setenv("MAGPIE_AUTO_TEST_RUNS", runLog)
	arWrapper(t, dir, "auto-recovery-fake-wrapper", "#!/bin/sh\necho LAUNCHED >> \"$MAGPIE_AUTO_TEST_RUNS\"\n"+strings.TrimPrefix(arAnswerScript(), "#!/bin/sh\n"))
	owner, outer := "owner", "desktop-session"
	b := newSubscriptionBridge()
	t.Cleanup(func() { bridges.Delete(b) })
	arSave(t, b, owner, outer, "inner-original", true)
	return b, &runtimeTUI{bridge: b}, owner, outer, runLog
}

func autoRecoveryBody(toolName, args, result string) []byte {
	// `result` is a JSON-escaped wire text value (e.g. `"wrote"`).
	return []byte(`{"model":"claude-opus-5-5","max_tokens":200,"tools":[{"name":"Write","input_schema":{"type":"object"}}],"messages":[` +
		`{"role":"user","content":"先读一下配置"},` +
		`{"role":"assistant","content":[{"type":"tool_use","id":"tool-A","name":"` + toolName + `","input":` + args + `}]},` +
		`{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-A","content":` + result + `}]}` +
		`]}`)
}

func autoRegister(t *testing.T, b *subscriptionBridge, owner, outer, name, args string) {
	t.Helper()
	run := &subscriptionRun{interactive: true, owner: owner, outerSession: outer, token: "previous-native-run"}
	digest, err := toolEvidenceArgumentsHash(json.RawMessage(args))
	if err != nil {
		t.Fatal(err)
	}
	e, err := toolEvidenceFor(run, name, "tool-A", digest, "", "registered")
	if err != nil {
		t.Fatal(err)
	}
	if err := b.toolJournal.append(e); err != nil {
		t.Fatal(err)
	}
	// The mock represents a previous wrapper that wrote its ownership
	// marker before creating a tmux pane, then died after dispatch.
	writeTestNativeTmuxMarker(t, owner, outer, "claude-bridge-deadbeef")
}

func writeTestNativeTmuxMarker(t *testing.T, owner, outer, name string) {
	t.Helper()
	path := interactiveTmuxGuardPath(owner, outer)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(interactiveTmuxOwner{Version: 2, Session: name, SocketEnvSHA256: interactiveTmuxSocketEnvironmentHash()})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestInteractiveAutoRecoveryRestartsToolFreeAfterNativeExit(t *testing.T) {
	b, rt, owner, outer, runLog := autoRecoveryFixture(t)
	// Registration is persisted, then the native process dies before it
	// sees the Desktop result. Simulate a completely new Runtime instance.
	autoRegister(t, b, owner, outer, "Write", `{"file_path":"note.txt"}`)
	restarted := newSubscriptionBridge()
	defer bridges.Delete(restarted)
	rt.bridge = restarted
	body := autoRecoveryBody("Write", `{"file_path":"note.txt"}`, `"wrote"`)
	status, msg, out := arServe(t, rt, restarted, owner, outer, body)
	if status != 200 {
		t.Fatalf("automatic tool-free recovery returned %d %s %s", status, msg, out)
	}
	if !strings.Contains(out, "OL_ANSWER") || strings.Contains(out, `"type":"tool_use"`) {
		t.Fatalf("recovery did not return tool-free text: %s", out)
	}
	entries, err := restarted.toolJournal.read()
	if err != nil || len(entries) != 2 || entries[1].Stage != "auto_recovery_claimed" {
		t.Fatalf("recovery claim missing: %v %+v", err, entries)
	}
	checkpoint, ok := restarted.interactiveSession(owner, outer)
	if !ok || checkpoint.Dirty {
		t.Fatalf("tool-free recovery did not commit clean checkpoint: %+v, found=%v", checkpoint, ok)
	}
	launches, _ := os.ReadFile(runLog)
	if strings.Count(string(launches), "LAUNCHED") != 1 {
		t.Fatalf("recovery launched %q native workers", launches)
	}
	status, _, _ = arServe(t, rt, restarted, owner, outer, body)
	if status != 409 {
		t.Fatalf("stale duplicate orphan results were replayed: status=%d", status)
	}
	launches, _ = os.ReadFile(runLog)
	if strings.Count(string(launches), "LAUNCHED") != 1 {
		t.Fatal("duplicate started a second Claude process")
	}
	// A subsequent genuinely new user message must be able to continue
	// through the canonical clean checkpoint (no further orphan result).
	follow := []byte(`{"model":"claude-opus-5-5","max_tokens":200,"tools":[{"name":"Write","input_schema":{"type":"object"}}],"messages":[` +
		`{"role":"user","content":"先读一下配置"},` +
		`{"role":"assistant","content":[{"type":"tool_use","id":"tool-A","name":"Write","input":{"file_path":"note.txt"}}]},` +
		`{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-A","content":"wrote"}]},` +
		`{"role":"assistant","content":"OL_ANSWER"},` +
		`{"role":"user","content":"继续核查"}]}`)
	status, msg, _ = arServe(t, rt, restarted, owner, outer, follow)
	if status != 200 {
		t.Fatalf("new user request after recovery could not resume: %d %s", status, msg)
	}
}

func TestInteractiveAutoRecoveryParallelClaimsAreAtomicAndSingleUse(t *testing.T) {
	b, _, owner, outer, _ := autoRecoveryFixture(t)
	run := &subscriptionRun{interactive: true, owner: owner, outerSession: outer, token: "old-parallel-native"}
	for _, call := range []struct{ id, name, args string }{
		{"id-1", "Write", `{"file_path":"note.txt"}`},
		{"id-2", "Bash", `{"command":"echo ok"}`},
	} {
		digest, err := toolEvidenceArgumentsHash(json.RawMessage(call.args))
		if err != nil {
			t.Fatal(err)
		}
		e, err := toolEvidenceFor(run, call.name, call.id, digest, "", "registered")
		if err != nil {
			t.Fatal(err)
		}
		if err := b.toolJournal.append(e); err != nil {
			t.Fatal(err)
		}
	}
	partial := `{"model":"claude-opus-5-5","messages":[{"role":"user","content":"先读一下配置"},` +
		`{"role":"assistant","content":[{"type":"tool_use","id":"id-1","name":"Write","input":{"file_path":"note.txt"}},{"type":"tool_use","id":"id-2","name":"Bash","input":{"command":"echo ok"}}]},` +
		`{"role":"user","content":[{"type":"tool_result","tool_use_id":"id-1","content":"wrote"}`
	request, err := parseAnthropic([]byte(partial + `]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if err := b.toolJournal.claimOrphanRecovery(owner, outer, request); err == nil {
		t.Fatal("partial parallel result was accepted")
	}
	complete, err := parseAnthropic([]byte(partial + `,{"type":"tool_result","tool_use_id":"id-2","content":"ok"}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if err := b.toolJournal.claimOrphanRecovery(owner, outer, complete); err != nil {
		t.Fatalf("complete parallel results were refused: %v", err)
	}
	if err := b.toolJournal.claimOrphanRecovery(owner, outer, complete); err == nil {
		t.Fatal("parallel result replay started another recovery")
	}
	records, err := b.toolJournal.read()
	if err != nil || len(records) != 4 || records[2].Stage != "auto_recovery_claimed" || records[3].Stage != "auto_recovery_claimed" {
		t.Fatalf("parallel claims did not persist atomically: %v %+v", err, records)
	}
}

func TestInteractiveAutoRecoveryCrashAfterClaimAllowsOneVerifiedReadOnlyRetry(t *testing.T) {
	b, rt, owner, outer, runLog := autoRecoveryFixture(t)
	autoRegister(t, b, owner, outer, "Write", `{"file_path":"note.txt"}`)
	t.Setenv("MAGPIE_CLAUDE_AUTO_RECOVER_RETRY", "1")
	body := autoRecoveryBody("Write", `{"file_path":"note.txt"}`, `"wrote"`)
	req, err := parseAnthropic(body)
	if err != nil {
		t.Fatal(err)
	}
	// Crash injection: Runtime fsyncs the one-shot claim but dies before
	// starting a recovery process. The checkpoint remains dirty.
	if err := b.toolJournal.claimOrphanRecovery(owner, outer, req); err != nil {
		t.Fatal(err)
	}
	newRuntime := newSubscriptionBridge()
	defer bridges.Delete(newRuntime)
	rt.bridge = newRuntime
	status, msg, out := arServe(t, rt, newRuntime, owner, outer, body)
	if status != 200 || !strings.Contains(out, "OL_ANSWER") {
		t.Fatalf("crash-after-claim could not reconstruct read-only answer: %d %s %s", status, msg, out)
	}
	records, err := newRuntime.toolJournal.read()
	if err != nil || len(records) != 3 || records[2].Stage != "auto_recovery_retry_claimed" {
		t.Fatalf("one-shot retry not committed: %v %+v", err, records)
	}
	launches, err := os.ReadFile(runLog)
	if err != nil || strings.Count(string(launches), "LAUNCHED") != 1 {
		t.Fatalf("unexpected native process count %q, %v", launches, err)
	}
	if err := newRuntime.toolJournal.claimOrphanRecovery(owner, outer, req); err == nil {
		t.Fatal("third read-only retry admitted after bounded claim")
	}
}

func TestInteractiveAutoRecoveryReadOnlyRetryRefusesChangedResult(t *testing.T) {
	b, _, owner, outer, _ := autoRecoveryFixture(t)
	autoRegister(t, b, owner, outer, "Write", `{"file_path":"note.txt"}`)
	t.Setenv("MAGPIE_CLAUDE_AUTO_RECOVER_RETRY", "1")
	correct, err := parseAnthropic(autoRecoveryBody("Write", `{"file_path":"note.txt"}`, `"wrote"`))
	if err != nil {
		t.Fatal(err)
	}
	if err := b.toolJournal.claimOrphanRecovery(owner, outer, correct); err != nil {
		t.Fatal(err)
	}
	forged, err := parseAnthropic(autoRecoveryBody("Write", `{"file_path":"note.txt"}`, `"forged"`))
	if err != nil {
		t.Fatal(err)
	}
	if err := b.toolJournal.claimOrphanRecovery(owner, outer, forged); err == nil {
		t.Fatal("accepted changed tool result on recovery retry")
	}
	entries, err := b.toolJournal.read()
	if err != nil || len(entries) != 2 {
		t.Fatalf("forgery consumed retry claim: %v %+v", err, entries)
	}
}

func TestInteractiveAutoRecoveryCrossInstanceClaimsSingleWinner(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("cross-process journal requires advisory OS locks")
	}
	b, _, owner, outer, _ := autoRecoveryFixture(t)
	autoRegister(t, b, owner, outer, "Write", `{"file_path":"note.txt"}`)
	req, err := parseAnthropic(autoRecoveryBody("Write", `{"file_path":"note.txt"}`, `"wrote"`))
	if err != nil {
		t.Fatal(err)
	}
	other := newInteractiveToolJournal()
	var wg sync.WaitGroup
	var winners int
	var mu sync.Mutex
	for _, j := range []*interactiveToolJournal{b.toolJournal, other} {
		wg.Add(1)
		go func(j *interactiveToolJournal) {
			defer wg.Done()
			if err := j.claimOrphanRecovery(owner, outer, req); err == nil {
				mu.Lock()
				winners++
				mu.Unlock()
			}
		}(j)
	}
	wg.Wait()
	if winners != 1 {
		t.Fatalf("concurrent independent Runtime journals launched %d recovery claims", winners)
	}
}

func TestInteractiveAutoRecoveryJournalLocksSerializeDifferentInstances(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("cross-process journal requires advisory OS locks")
	}
	b, _, owner, outer, _ := autoRecoveryFixture(t)
	release, err := lockInteractiveToolJournal(b.toolJournal.path)
	if err != nil {
		t.Fatal(err)
	}
	run := &subscriptionRun{interactive: true, owner: owner, outerSession: outer, token: "old"}
	e, err := toolEvidenceFor(run, "Write", "tool-1", toolEvidenceHash("args"), "", "registered")
	if err != nil {
		t.Fatal(err)
	}
	other := newInteractiveToolJournal()
	done := make(chan error, 1)
	go func() { done <- other.append(e) }()
	select {
	case err := <-done:
		release()
		t.Fatalf("second journal bypassed exclusive file lock: %v", err)
	case <-time.After(75 * time.Millisecond):
	}
	release()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("journal remained blocked after lock release")
	}
}

func TestInteractiveAutoRecoveryRefusesChangedPreviouslyVerifiedResult(t *testing.T) {
	b, _, owner, outer, _ := autoRecoveryFixture(t)
	autoRegister(t, b, owner, outer, "Write", `{"file_path":"note.txt"}`)
	run := &subscriptionRun{interactive: true, owner: owner, outerSession: outer, token: "previous-native-run"}
	if err := b.toolJournal.verifyResults(run, []Part{{Kind: ToolResult, CallID: "tool-A", Text: "wrote"}}); err != nil {
		t.Fatal(err)
	}
	wrong, err := parseAnthropic(autoRecoveryBody("Write", `{"file_path":"note.txt"}`, `"forged"`))
	if err != nil {
		t.Fatal(err)
	}
	if err := b.toolJournal.claimOrphanRecovery(owner, outer, wrong); err == nil {
		t.Fatal("accepted changed result from already-verified client")
	}
	correct, err := parseAnthropic(autoRecoveryBody("Write", `{"file_path":"note.txt"}`, `"wrote"`))
	if err != nil {
		t.Fatal(err)
	}
	if err := b.toolJournal.claimOrphanRecovery(owner, outer, correct); err != nil {
		t.Fatalf("refused identical verified result: %v", err)
	}
}

func TestInteractiveAutoRecoveryNeverReplacesActiveNativeOwner(t *testing.T) {
	b, rt, owner, outer, runLog := autoRecoveryFixture(t)
	autoRegister(t, b, owner, outer, "Write", `{"file_path":"note.txt"}`)
	b.mu.Lock()
	b.liveOuter[interactiveSessionStateKey(owner, outer)] = &subscriptionRun{interactive: true, owner: owner, outerSession: outer}
	b.mu.Unlock()
	status, _, _ := arServe(t, rt, b, owner, outer, autoRecoveryBody("Write", `{"file_path":"note.txt"}`, `"wrote"`))
	if status != 409 {
		t.Fatalf("replaced active owner: HTTP %d", status)
	}
	if _, err := os.Stat(runLog); !os.IsNotExist(err) {
		t.Fatal("launched a second process while original still owned session")
	}
	records, err := b.toolJournal.read()
	if err != nil || len(records) != 1 {
		t.Fatalf("claimed result before active-owner check: %v %+v", err, records)
	}
}

func TestInteractiveAutoRecoveryRefusesUnverifiedAndForgedOrphans(t *testing.T) {
	for _, tc := range []struct {
		name, tool, args, result string
		register                 bool
	}{
		{"no-journal", "Write", `{"file_path":"note.txt"}`, `"wrote"`, false},
		{"forged-tool-name", "Edit", `{"file_path":"note.txt"}`, `"wrote"`, true},
		{"forged-arguments", "Write", `{"file_path":"other.txt"}`, `"wrote"`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b, rt, owner, outer, runLog := autoRecoveryFixture(t)
			if tc.register {
				autoRegister(t, b, owner, outer, "Write", `{"file_path":"note.txt"}`)
			}
			status, _, _ := arServe(t, rt, b, owner, outer, autoRecoveryBody(tc.tool, tc.args, tc.result))
			if status != 409 {
				t.Fatalf("uncertain orphan recovered: HTTP %d", status)
			}
			if _, err := os.Stat(runLog); !os.IsNotExist(err) {
				t.Fatalf("unexpected execution: %v", err)
			}
		})
	}
}

func TestInteractiveAutoRecoveryRequiresExplicitOptInConfiguration(t *testing.T) {
	b, rt, owner, outer, runLog := autoRecoveryFixture(t)
	autoRegister(t, b, owner, outer, "Write", `{"file_path":"note.txt"}`)
	t.Setenv("MAGPIE_CLAUDE_AUTO_RECOVER_ORPHANS", "0")
	status, _, _ := arServe(t, rt, b, owner, outer, autoRecoveryBody("Write", `{"file_path":"note.txt"}`, `"wrote"`))
	if status != 409 {
		t.Fatalf("disabled automatic recovery returned %d", status)
	}
	if _, err := os.Stat(runLog); !os.IsNotExist(err) {
		t.Fatalf("unexpected execution: %v", err)
	}
}
