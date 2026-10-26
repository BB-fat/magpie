//go:build darwin || linux

package claudetui

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestInteractiveDurableToolJournalReopenAndRefuseReplay(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("MAGPIE_CLAUDE_DURABLE_TOOL_JOURNAL", "1")
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	run := &subscriptionRun{interactive: true, owner: "owner", outerSession: "conversation", token: "private-callback-token"}
	registration, err := toolEvidenceFor(run, "Write", "tool-call-1", toolEvidenceHash(`{"path":"test.txt"}`), "", "registered")
	if err != nil {
		t.Fatal(err)
	}
	if err := b.toolJournal.append(registration); err != nil {
		t.Fatal(err)
	}
	result := Part{Kind: ToolResult, CallID: "tool-call-1", Text: "PRIVATE_OUTPUT_VALUE"}
	if err := b.toolJournal.verifyResults(run, []Part{result}); err != nil {
		t.Fatal(err)
	}
	// A new bridge after a crash/restart must see exactly the same evidence.
	restart := newSubscriptionBridge()
	defer bridges.Delete(restart)
	entries, err := restart.toolJournal.read()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Stage != "registered" || entries[1].Stage != "result_verified" {
		t.Fatalf("restarted journal lost ordering: %+v", entries)
	}
	raw, err := os.ReadFile(restart.toolJournal.path)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"private-callback-token", "tool-call-1", "PRIVATE_OUTPUT_VALUE", "test.txt", "conversation"} {
		if strings.Contains(string(raw), secret) {
			t.Errorf("private tool data persisted: %s", secret)
		}
	}
	if err := restart.toolJournal.append(registration); err == nil {
		t.Fatal("duplicate callback was accepted")
	}
	if err := restart.toolJournal.verifyResults(run, []Part{result}); err == nil {
		t.Fatal("duplicate tool result was accepted")
	}
	other := &subscriptionRun{interactive: true, owner: "owner", outerSession: "other", token: run.token}
	if err := restart.toolJournal.verifyResults(other, []Part{result}); err == nil {
		t.Fatal("cross-conversation result accepted")
	}
	other = &subscriptionRun{interactive: true, owner: run.owner, outerSession: run.outerSession, token: "different-run"}
	if err := restart.toolJournal.verifyResults(other, []Part{result}); err == nil {
		t.Fatal("cross-run result accepted")
	}
}

func TestInteractiveDurableToolJournalTruncatedTailFailsClosed(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("MAGPIE_CLAUDE_DURABLE_TOOL_JOURNAL", "1")
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	run := &subscriptionRun{interactive: true, owner: "owner", outerSession: "conversation", token: "token"}
	e, _ := toolEvidenceFor(run, "Bash", "id", toolEvidenceHash(`{"cmd":"echo ok"}`), "", "registered")
	if err := b.toolJournal.append(e); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(b.toolJournal.path, os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`{"version":1,"stage":"result_verified"}`); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := newInteractiveToolJournal().read(); err == nil {
		t.Fatal("truncated journal tail accepted after restart")
	}
	e.Call = toolEvidenceHash("new-id")
	if err := b.toolJournal.append(e); err == nil {
		t.Fatal("appended after corrupt tail")
	}
}

func TestInteractiveDurableToolJournalRefusesTamperingAndUnsafePermissions(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("MAGPIE_CLAUDE_DURABLE_TOOL_JOURNAL", "1")
	j := newInteractiveToolJournal()
	run := &subscriptionRun{interactive: true, owner: "one", outerSession: "conversation", token: "token"}
	e, err := toolEvidenceFor(run, "Edit", "tool-1", toolEvidenceHash("args"), "", "registered")
	if err != nil {
		t.Fatal(err)
	}
	if err := j.append(e); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(j.path, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := j.read(); err == nil {
		t.Fatal("accepted group/world-readable tool evidence")
	}
	if err := os.Chmod(j.path, 0o600); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(j.path)
	if err != nil {
		t.Fatal(err)
	}
	b = []byte(strings.Replace(string(b), `"tool":"Edit"`, `"tool":"Bash"`, 1))
	if err := os.WriteFile(j.path, b, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := j.read(); err == nil {
		t.Fatal("accepted record with mismatched checksum")
	}
}

func TestInteractiveDurableToolJournalOptInOnly(t *testing.T) {
	t.Setenv("MAGPIE_CLAUDE_DURABLE_TOOL_JOURNAL", "0")
	if j := newInteractiveToolJournal(); j != nil {
		t.Fatal("journal unexpectedly enabled before explicit rollout")
	}
}

func TestInteractiveToolArgumentsCanonicalizeWithoutFloatPrecisionLoss(t *testing.T) {
	first, err := toolEvidenceArgumentsHash(json.RawMessage(`{"nested":{"b":2,"a":1},"n":9007199254740992}`))
	if err != nil {
		t.Fatal(err)
	}
	reordered, err := toolEvidenceArgumentsHash(json.RawMessage(`{ "n" : 9007199254740992, "nested" : { "a": 1, "b": 2 } }`))
	if err != nil {
		t.Fatal(err)
	}
	if first != reordered {
		t.Fatal("semantically identical object argument encodings did not match")
	}
	different, err := toolEvidenceArgumentsHash(json.RawMessage(`{"nested":{"a":1,"b":2},"n":9007199254740993}`))
	if err != nil {
		t.Fatal(err)
	}
	if first == different {
		t.Fatal("distinct large JSON integers collided through float64 rounding")
	}
	for _, raw := range []string{`null`, `[]`, `{"ok":true}{"forged":true}`} {
		if _, err := toolEvidenceArgumentsHash(json.RawMessage(raw)); err == nil {
			t.Fatalf("accepted unsafe MCP arguments %q", raw)
		}
	}
}

func TestInteractiveDurableMCPCallbackRegistersBeforeHandoff(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("MAGPIE_CLAUDE_DURABLE_TOOL_JOURNAL", "1")
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	callback := make(chan struct{}, 1)
	run := &subscriptionRun{bridge: b, interactive: true, owner: "owner", outerSession: "outer", token: "callback-token", tools: map[string]bool{"Write": true}}
	run.onCall = func(string, string, json.RawMessage) { callback <- struct{}{} }
	b.mu.Lock()
	b.runs[run.token] = run
	b.mu.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	request := httptest.NewRequest("POST", "/_magpie/claude-mcp/"+run.token, strings.NewReader(`{"tool_call_id":"write-1","name":"Write","arguments":{"file_path":"private.txt"}}`)).WithContext(ctx)
	request.SetPathValue("token", run.token)
	w := httptest.NewRecorder()
	done := make(chan struct{})
	go func() { b.mcpCall(w, request); close(done) }()
	select {
	case <-callback:
	case <-time.After(2 * time.Second):
		t.Fatal("native MCP callback was not reached")
	}
	entries, err := b.toolJournal.read()
	if err != nil || len(entries) != 1 || entries[0].Stage != "registered" {
		t.Fatalf("callback ran before durable evidence was committed: %+v, %v", entries, err)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("cancelled MCP request did not stop")
	}
	// Cancellation must not erase the durable fact that the native tool
	// callback was registered. It still does NOT prove Desktop wrote a file.
	entries, err = b.toolJournal.read()
	if err != nil || len(entries) != 1 {
		t.Fatalf("cancellation erased callback evidence: %v %v", entries, err)
	}
}
