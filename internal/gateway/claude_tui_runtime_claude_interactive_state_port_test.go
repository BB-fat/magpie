package gateway

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Checkpoints persist instruction identity and survive a runtime restart.
func TestClaudeTUIRuntimeInteractiveCheckpointStorePersistsInstructions(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	bridge := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(bridge)
	owner, outer := "claude\x00user", "desktop-session-A"
	original := tuiRuntimeInteractiveSessionEntry{
		SessionID:      "inner-session-A",
		Model:          "claude-opus-5-5",
		ThreadKey:      "thread-A",
		ConvKey:        "checkpoint-A",
		ReplyKey:       "reply-A",
		InstructionKey: "context-A",
	}
	bridge.saveInteractiveSession(owner, outer, original)
	path := tuiRuntimeInteractiveSessionStatePath()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("persisted state mode %o, want 600", info.Mode().Perm())
	}
	state := tuiRuntimeReadInteractiveSessionState()
	if state.Version != 2 || len(state.Sessions) != 1 {
		t.Fatalf("unexpected saved state: %+v", state)
	}
	got, ok := bridge.interactiveSession(owner, outer)
	if !ok || got.SessionID != original.SessionID || got.Model != original.Model ||
		got.ThreadKey != original.ThreadKey || got.ReplyKey != original.ReplyKey ||
		got.InstructionKey != original.InstructionKey || got.Dirty {
		t.Fatalf("checkpoint lost fields: %+v, found=%t", got, ok)
	}
	if _, ok := bridge.interactiveSession(owner, "desktop-session-B"); ok {
		t.Fatal("session B read session A checkpoint")
	}
	if _, ok := bridge.interactiveSession("other-owner", outer); ok {
		t.Fatal("another owner read session A checkpoint")
	}
	bridge.markInteractiveSessionDirty(owner, outer)
	updated, ok := bridge.interactiveSession(owner, outer)
	if !ok || !updated.Dirty || updated.SessionID != original.SessionID || updated.UpdatedAt.IsZero() {
		t.Fatalf("mark dirty did not preserve checkpoint: %+v", updated)
	}
	// Simulate the gateway restarting with a newly created bridge.
	reopened := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(reopened)
	if again, ok := reopened.interactiveSession(owner, outer); !ok || !again.Dirty || again.ConvKey != original.ConvKey {
		t.Fatalf("restart lost dirty checkpoint: %+v, found=%t", again, ok)
	}
}

func TestClaudeTUIRuntimeInteractiveCheckpointStoreRejectsInvalidVersionAndPrunesOld(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path := tuiRuntimeInteractiveSessionStatePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"version":999,"sessions":{"bad":{"session_id":"old"}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := tuiRuntimeReadInteractiveSessionState(); len(got.Sessions) != 0 || got.Version != tuiRuntimeInteractiveSessionStateVersion {
		t.Fatalf("accepted invalid schema: %+v", got)
	}
	state := tuiRuntimeInteractiveSessionState{Version: tuiRuntimeInteractiveSessionStateVersion, Sessions: map[string]tuiRuntimeInteractiveSessionEntry{
		"old": {SessionID: "old", UpdatedAt: time.Now().Add(-31 * 24 * time.Hour)},
		"new": {SessionID: "new", ConvKey: "conv", InstructionKey: "context", UpdatedAt: time.Now()},
	}}
	if err := tuiRuntimeWriteInteractiveSessionState(state); err != nil {
		t.Fatal(err)
	}
	got := tuiRuntimeReadInteractiveSessionState()
	if _, ok := got.Sessions["old"]; ok {
		t.Fatal("old checkpoint was not pruned")
	}
	if got.Sessions["new"].SessionID != "new" {
		t.Fatalf("recent checkpoint lost: %+v", got)
	}
}

func TestClaudeTUIRuntimeInteractiveCheckpointWorkDirIsPerOwnerAndSession(t *testing.T) {
	first, err := tuiRuntimeInteractiveSessionWorkDir("account-A", "desktop-A")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(first) })
	again, err := tuiRuntimeInteractiveSessionWorkDir("account-A", "desktop-A")
	if err != nil || again != first {
		t.Fatalf("work dir changed between turns: %q and %q: %v", first, again, err)
	}
	second, err := tuiRuntimeInteractiveSessionWorkDir("account-A", "desktop-B")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(second) })
	if first == second || strings.Contains(first, "desktop-A") {
		t.Fatalf("work dir collision or session ID disclosure: %s / %s", first, second)
	}
	info, err := os.Stat(first)
	if err != nil || info.Mode().Perm()&0o022 != 0 {
		t.Fatalf("work dir is unsafe: %+v, %v", info, err)
	}
}

func TestClaudeTUIRuntimeInteractiveCompletedCanonicalTurnWritesCheckpoint(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	req := &tuiRuntimeRequest{Request: Request{
		Model: "claude-opus-5-5",
		Tools: []Tool{{Name: "echo_test"}},
		Messages: []Message{
			{Role: "user", Parts: []Part{{Kind: Text, Text: "main request"}}},
			{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "call_1", Name: "echo_test"}}},
			{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "call_1", Text: "done"}}},
		},
	}}
	run := &tuiRuntimeSubscriptionRun{
		bridge: b, interactive: true, owner: "owner", model: req.Model,
		outerSession: "outer-main", threadKey: firstWords(&req.Request), sessionID: "inner-main",
	}
	run.ended(req, "FINAL", "stop", true)
	got, ok := b.interactiveSession("owner", "outer-main")
	if !ok {
		t.Fatal("completed canonical PTY turn did not write a checkpoint")
	}
	if got.SessionID != "inner-main" || got.Model != req.Model || got.ThreadKey != firstWords(&req.Request) ||
		got.ConvKey == "" || got.ReplyKey == "" || got.InstructionKey != tuiRuntimeInstructionKey(req) || got.Dirty {
		t.Fatalf("incomplete checkpoint: %+v", got)
	}
	reply := Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: "FINAL"}}}
	if got.ReplyKey != tuiRuntimeAssistantReplyKey(reply) {
		t.Fatalf("reply anchor mismatch: %q", got.ReplyKey)
	}
}

func TestClaudeTUIRuntimeInteractiveCompletedHelperCannotWriteCheckpoint(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	req := &tuiRuntimeRequest{Request: Request{Model: "claude-haiku-4-5", Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "helper request"}}}}}}
	run := &tuiRuntimeSubscriptionRun{
		bridge: b, interactive: true, owner: "owner", model: req.Model,
		outerSession: "", threadKey: firstWords(&req.Request), sessionID: "inner-helper",
	}
	run.ended(req, "SIDE", "stop", true)
	state := tuiRuntimeReadInteractiveSessionState()
	if len(state.Sessions) != 0 {
		t.Fatalf("ephemeral helper wrote persisted state: %+v", state.Sessions)
	}
}

func TestClaudeTUIRuntimeInteractiveRestoreDecisionExactReplyAnchorDirtyAndCompanion(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer := "owner", "outer"
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "answer"}}},
	}
	req0 := &tuiRuntimeRequest{Request: Request{Model: "claude-opus-5-5", Messages: base}}
	keys := tuiRuntimeConversationKeys(owner, base)
	if len(keys) != 1 {
		t.Fatalf("conversation key count = %d", len(keys))
	}
	entry := tuiRuntimeInteractiveSessionEntry{
		SessionID: "inner", Model: req0.Model, ThreadKey: firstWords(&req0.Request),
		ConvKey: keys[0], ReplyKey: tuiRuntimeAssistantReplyKey(base[1]),
		InstructionKey: tuiRuntimeInstructionKey(req0),
	}
	b.saveInteractiveSession(owner, outer, entry)

	exact := *req0
	exact.Messages = append(append([]Message{}, base...), Message{Role: "user", Parts: []Part{{Kind: Text, Text: "next"}}})
	got, since, status := b.restoreInteractiveSession(owner, outer, &exact)
	if status != tuiRuntimeInteractiveRestoreExact || got.SessionID != "inner" || len(since) != 1 || since[0].Parts[0].Text != "next" {
		t.Fatalf("exact restore = %q %#v %#v", status, got, since)
	}

	rewritten := *req0
	rewritten.Messages = []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "rewritten earlier context"}}},
		base[1],
		{Role: "user", Parts: []Part{{Kind: Text, Text: "after rewrite"}}},
	}
	_, since, status = b.restoreInteractiveSession(owner, outer, &rewritten)
	if status != tuiRuntimeInteractiveRestoreReplyAnchor || len(since) != 1 || since[0].Parts[0].Text != "after rewrite" {
		t.Fatalf("reply-anchor restore = %q %#v", status, since)
	}

	b.markInteractiveSessionDirty(owner, outer)
	_, _, status = b.restoreInteractiveSession(owner, outer, &exact)
	if status != tuiRuntimeInteractiveRestoreDirty {
		t.Fatalf("dirty restore = %q", status)
	}
	companion := exact
	companion.Messages = append(append([]Message{}, base...), Message{Role: "user", Parts: []Part{{Kind: Text, Text: "Your response above was cut off mid-stream. Resume directly from where it stops — no apology, no recap."}}})
	_, since, status = b.restoreInteractiveSession(owner, outer, &companion)
	if status != tuiRuntimeInteractiveRestoreCompanion || len(since) != 1 {
		t.Fatalf("turn companion did not override dirty checkpoint: %q %#v", status, since)
	}
}

func TestClaudeTUIRuntimeInteractiveRestoreRejectsUnsafeSuffixesAndAmbiguousReplyAnchor(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner := "owner"
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "answer"}}},
	}
	keys := tuiRuntimeConversationKeys(owner, base)
	entry := tuiRuntimeInteractiveSessionEntry{
		SessionID: "inner", ConvKey: keys[0], ReplyKey: tuiRuntimeAssistantReplyKey(base[1]), InstructionKey: "ctx",
	}
	b.saveInteractiveSession(owner, "tool", entry)
	toolReq := &tuiRuntimeRequest{Request: Request{Messages: append(append([]Message{}, base...), Message{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "old", Text: "stale"}}})}}
	_, _, status := b.restoreInteractiveSession(owner, "tool", toolReq)
	if status != tuiRuntimeInteractiveRestoreSuffixRejected {
		t.Fatalf("tool-result suffix accepted: %q", status)
	}

	b.saveInteractiveSession(owner, "ambiguous", entry)
	ambiguous := &tuiRuntimeRequest{Request: Request{Messages: []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "rewritten"}}},
		base[1],
		{Role: "user", Parts: []Part{{Kind: Text, Text: "middle"}}},
		base[1],
		{Role: "user", Parts: []Part{{Kind: Text, Text: "next"}}},
	}}}
	_, _, status = b.restoreInteractiveSession(owner, "ambiguous", ambiguous)
	if status != tuiRuntimeInteractiveRestoreSuffixRejected {
		t.Fatalf("ambiguous repeated reply anchor accepted: %q", status)
	}
}

func TestClaudeTUIRuntimeInteractiveRestoreSkipsOnlyFailedCompanionTransportArtifact(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer := "owner", "outer"
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "answer"}}},
	}
	keys := tuiRuntimeConversationKeys(owner, base)
	b.saveInteractiveSession(owner, outer, tuiRuntimeInteractiveSessionEntry{
		SessionID: "inner", ConvKey: keys[0], ReplyKey: tuiRuntimeAssistantReplyKey(base[1]), InstructionKey: "ctx",
	})
	req := &tuiRuntimeRequest{Request: Request{Messages: append(append([]Message{}, base...),
		Message{Role: "user", Parts: []Part{{Kind: Text, Text: "Your response above was cut off mid-stream. Resume directly from where it stops — no apology, no recap."}}},
		Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: "API Error: 502 US VPS Claude: Claude Code: Claude interactive persisted resume failed: dirty; refusing full history replay. This is a server-side issue."}}},
		Message{Role: "user", Parts: []Part{{Kind: Text, Text: "real next turn"}}},
	)}}
	_, since, status := b.restoreInteractiveSession(owner, outer, req)
	if status != tuiRuntimeInteractiveRestoreExact || len(since) != 1 || since[0].Parts[0].Text != "real next turn" {
		t.Fatalf("failed companion artifact not skipped safely: %q %#v", status, since)
	}
}
