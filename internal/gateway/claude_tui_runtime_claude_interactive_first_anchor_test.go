//go:build darwin || linux

package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func tuiRuntimeFirstAnchorFixture(t *testing.T) (*tuiRuntimeSubscriptionBridge, *tuiRuntimeSubscriptionRun) {
	t.Helper()
	tuiRuntimeArDir(t)
	t.Setenv("MAGPIE_CLAUDE_DURABLE_TOOL_JOURNAL", "1")
	t.Setenv("MAGPIE_CLAUDE_FIRST_TURN_ANCHOR", "1")
	b := tuiRuntimeNewSubscriptionBridge()
	t.Cleanup(func() { tuiRuntimeBridges.Delete(b) })
	owner, outer := "owner", "first-conversation"
	first := []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "First user instruction"}}}}
	initial := tuiRuntimeToolEvidenceHash(owner + "\x00" + tuiRuntimeHistoryKey(first))
	r := &tuiRuntimeSubscriptionRun{
		bridge: b, interactive: true, owner: owner, outerSession: outer,
		token: "real-native-token", model: tuiRuntimeAutoTestModel, threadKey: "first-thread",
		firstTurnAnchorKey: initial, interactiveContextKey: "context-hash",
		seenMCPCallIDs: map[string]bool{},
	}
	return b, r
}

func TestClaudeTUIRuntimeInteractiveFirstTurnCallbackBeforeNativeIDNeverDispatchesTool(t *testing.T) {
	b, r := tuiRuntimeFirstAnchorFixture(t)
	r.tools = map[string]bool{"Write": true}
	b.mu.Lock()
	b.runs[r.token] = r
	b.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 85*time.Millisecond)
	defer cancel()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/_magpie/claude-mcp/"+r.token, strings.NewReader(`{"tool_call_id":"tool-1","name":"Write","arguments":{"file_path":"test.txt"}}`)).WithContext(ctx)
	req.SetPathValue("token", r.token)
	started := time.Now()
	b.mcpCall(w, req)
	if time.Since(started) > time.Second {
		t.Fatal("unverified first tool callback did not respect client cancellation")
	}
	if len(r.seenMCPCallIDs) != 0 {
		t.Fatal("callback registered or dispatched without an authentic native session ID")
	}
	if _, ok := b.interactiveSession(r.owner, r.outerSession); ok {
		t.Fatal("phantom first-turn checkpoint persisted")
	}
	entries, err := b.toolJournal.read()
	if err != nil || len(entries) != 0 {
		t.Fatalf("phantom tool registration persisted: %v %+v", err, entries)
	}
}

func TestClaudeTUIRuntimeInteractiveFirstTurnDirtyAnchorRecoversToolFreeAndCommitsClean(t *testing.T) {
	dir := tuiRuntimeArDir(t)
	t.Setenv("MAGPIE_CLAUDE_DURABLE_TOOL_JOURNAL", "1")
	t.Setenv("MAGPIE_CLAUDE_FIRST_TURN_ANCHOR", "1")
	t.Setenv("MAGPIE_CLAUDE_AUTO_RECOVER_ORPHANS", "1")
	t.Setenv("MAGPIE_CLAUDE_NATIVE_TMUX_GUARD", "1")
	if err := os.WriteFile(filepath.Join(dir, "tmux"), []byte("#!/bin/sh\necho \"can't find session\" >&2\nexit 1\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	runLog := filepath.Join(dir, "wrapper-launches")
	t.Setenv("MAGPIE_AUTO_TEST_RUNS", runLog)
	tuiRuntimeArWrapper(t, dir, "first-turn-fake-wrapper", "#!/bin/sh\necho LAUNCHED >> \"$MAGPIE_AUTO_TEST_RUNS\"\n"+strings.TrimPrefix(tuiRuntimeArAnswerScript(), "#!/bin/sh\n"))
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer := "owner", "first-turn-outer"
	body := tuiRuntimeAutoRecoveryBody("Write", `{"file_path":"note.txt"}`, `"wrote"`)
	request, err := parseAnthropic(body)
	if err != nil {
		t.Fatal(err)
	}
	first := request.Messages[:1]
	r := &tuiRuntimeSubscriptionRun{
		bridge: b, interactive: true, owner: owner, outerSession: outer,
		token: "true-native-callback", model: tuiRuntimeAutoTestModel,
		sessionID: "real-native-init-session", threadKey: firstWords(request),
		firstTurnAnchorKey:    tuiRuntimeToolEvidenceHash(owner + "\x00" + tuiRuntimeHistoryKey(first)),
		interactiveContextKey: "initial-context", seenMCPCallIDs: map[string]bool{},
	}
	argsHash, err := tuiRuntimeToolEvidenceArgumentsHash(json.RawMessage(`{"file_path":"note.txt"}`))
	if err != nil {
		t.Fatal(err)
	}
	e, err := tuiRuntimeToolEvidenceFor(r, "Write", "tool-A", argsHash, "", "registered")
	if err != nil {
		t.Fatal(err)
	}
	if err := b.toolJournal.append(e); err != nil {
		t.Fatal(err)
	}
	r.seenMCPCallIDs["tool-A"] = true
	if err := r.persistFirstToolAnchorLocked(); err != nil {
		t.Fatal(err)
	}
	before, ok := b.interactiveSession(owner, outer)
	if !ok || !before.Dirty || before.ReplyKey != "" || before.SessionID != "real-native-init-session" {
		t.Fatalf("first-turn anchor wasn't dirty: %+v", before)
	}
	tuiRuntimeWriteTestNativeTmuxMarker(t, owner, outer, "claude-bridge-deadbeef")
	// The prior process is now gone. This is a *fresh, tool-free* context
	// reconstruction, never a re-issue of the original Write tool.
	rt := &tuiRuntime{bridge: b}
	status, msg, out := tuiRuntimeArServe(t, rt, b, owner, outer, body)
	if status != 200 || !strings.Contains(out, "OL_ANSWER") {
		t.Fatalf("first-turn recovery refused: %d %s %s", status, msg, out)
	}
	after, ok := b.interactiveSession(owner, outer)
	if !ok || after.Dirty || after.SessionID == before.SessionID || after.ReplyKey == "" {
		t.Fatalf("first-turn recovery did not commit clean checkpoint: %+v", after)
	}
	launches, err := os.ReadFile(runLog)
	if err != nil || strings.Count(string(launches), "LAUNCHED") != 1 {
		t.Fatalf("recovery unexpectedly launched extra work: %q %v", launches, err)
	}
}

func TestClaudeTUIRuntimeInteractiveFirstTurnAnchorsOnlyRealNativeSessionAndRegisteredCallback(t *testing.T) {
	b, r := tuiRuntimeFirstAnchorFixture(t)
	if err := r.persistFirstToolAnchorLocked(); err != nil {
		t.Fatal(err)
	}
	if _, ok := b.interactiveSession(r.owner, r.outerSession); ok {
		t.Fatal("made checkpoint before a real native session existed")
	}
	r.sessionID = "real-session-from-wrapper"
	if err := r.persistFirstToolAnchorLocked(); err != nil {
		t.Fatal(err)
	}
	if _, ok := b.interactiveSession(r.owner, r.outerSession); ok {
		t.Fatal("anchored without any tool registration")
	}
	digest := tuiRuntimeToolEvidenceHash(`{"file_path":"test.txt"}`)
	e, err := tuiRuntimeToolEvidenceFor(r, "Write", "tool-1", digest, "", "registered")
	if err != nil {
		t.Fatal(err)
	}
	if err := b.toolJournal.append(e); err != nil {
		t.Fatal(err)
	}
	r.seenMCPCallIDs["tool-1"] = true
	if err := r.persistFirstToolAnchorLocked(); err != nil {
		t.Fatal(err)
	}
	saved, ok := b.interactiveSession(r.owner, r.outerSession)
	if !ok || !saved.Dirty || saved.ReplyKey != "" || saved.SessionID != "real-session-from-wrapper" || saved.ConvKey != r.firstTurnAnchorKey {
		t.Fatalf("first-turn anchor is not a real dirty checkpoint: %+v,ok=%v", saved, ok)
	}
	if err := r.persistFirstToolAnchorLocked(); err != nil {
		t.Fatalf("idempotent repeat callback refused: %v", err)
	}
	r.sessionID = "different-native-id"
	if err := r.persistFirstToolAnchorLocked(); err == nil {
		t.Fatal("overwrote first session with a different native session ID")
	}
}

func TestClaudeTUIRuntimeInteractiveFirstTurnAnchorDoesNotOverwriteCompletedCheckpoint(t *testing.T) {
	b, r := tuiRuntimeFirstAnchorFixture(t)
	entry := tuiRuntimeInteractiveSessionEntry{SessionID: "previous-session", ConvKey: "completed", ReplyKey: "reply", ContextKey: "context"}
	if err := b.saveInteractiveSession(r.owner, r.outerSession, entry); err != nil {
		t.Fatal(err)
	}
	r.sessionID = "new-session"
	r.seenMCPCallIDs["tool-1"] = true
	if err := r.persistFirstToolAnchorLocked(); err == nil {
		t.Fatal("overwrote a completed checkpoint")
	}
	after, ok := b.interactiveSession(r.owner, r.outerSession)
	if !ok || after.SessionID != "previous-session" || after.Dirty {
		t.Fatalf("existing checkpoint changed: %+v", after)
	}
}

func TestClaudeTUIRuntimeInteractiveFirstTurnAnchorRefusesCorruptStore(t *testing.T) {
	_, r := tuiRuntimeFirstAnchorFixture(t)
	r.sessionID = "real-session"
	r.seenMCPCallIDs["tool-1"] = true
	path := tuiRuntimeInteractiveSessionStatePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"version":1,"sessions":broken}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := r.persistFirstToolAnchorLocked(); err == nil {
		t.Fatal("corrupt checkpoint store silently replaced")
	}
	b, err := os.ReadFile(path)
	if err != nil || string(b) != `{"version":1,"sessions":broken}` {
		t.Fatalf("corrupt store mutated: %q, %v", b, err)
	}
}

func TestClaudeTUIRuntimeInteractiveFirstTurnAnchorNotEnabledWithoutFullFeature(t *testing.T) {
	_, r := tuiRuntimeFirstAnchorFixture(t)
	r.firstTurnAnchorKey = ""
	r.sessionID = "real-session"
	r.seenMCPCallIDs["tool-1"] = true
	if err := r.persistFirstToolAnchorLocked(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(tuiRuntimeInteractiveSessionStatePath()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("disabled first-turn anchor wrote state: %v", err)
	}
}
