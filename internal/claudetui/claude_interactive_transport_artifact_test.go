package claudetui

import (
	"fmt"
	"strings"
	"testing"
)

// These are client-visible transport failures, *not* replies written to
// the canonical native Claude transcript. Only a verified tool-free Desktop
// continuation/error pair may be elided from the resume suffix.
func TestRestoreSkipsKnownFailedDesktopCompanionWithoutToolReplay(t *testing.T) {
	for _, reason := range []string{"dirty", "suffix_rejected", "checkpoint_not_found", "context_mismatch", "state_missing"} {
		t.Run(reason, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			b := newSubscriptionBridge()
			defer bridges.Delete(b)
			owner, outer := "owner", "outer"
			base := []Message{
				{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}},
				{Role: "assistant", Parts: []Part{{Kind: Text, Text: "complete answer"}}},
			}
			keys := conversationKeys(owner, base)
			if err := b.saveInteractiveSession(owner, outer, interactiveSessionEntry{
				SessionID: "inner", ConvKey: keys[0], ReplyKey: assistantReplyKey(base[1]), ContextKey: "ctx",
			}); err != nil {
				t.Fatal(err)
			}
			companion := Message{Role: "user", Parts: []Part{{Kind: Text, Text: "Your response above was cut off mid-stream. Resume directly from where it stops — no apology, no recap."}}}
			errorReply := Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: fmt.Sprintf(
				"API Error: 502 NTT VPS Claude: Claude TUI: Claude interactive persisted resume failed: %s; refusing full history replay. This is a server-side issue.", reason)}}}
			user := Message{Role: "user", Parts: []Part{{Kind: Text, Text: "new question"}}}
			msgs := append(append([]Message{}, base...), companion, errorReply, user)
			_, suffix, status := b.restoreInteractiveSession(owner, outer, &Request{Messages: msgs})
			if status != interactiveRestoreExact || len(suffix) != 1 || suffix[0].Parts[0].Text != "new question" {
				t.Fatalf("got status=%s, suffix=%#v", status, suffix)
			}
		})
	}
}

func TestRestoreDoesNotDiscardToolsOrUnprovenAssistantMessages(t *testing.T) {
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "complete answer"}}},
	}
	companion := Message{Role: "user", Parts: []Part{{Kind: Text, Text: "Your response above was cut off mid-stream. Continue."}}}
	errorText := "API Error: 502 NTT VPS Claude: Claude TUI: Claude interactive persisted resume failed: suffix_rejected; refusing full history replay. This is a server-side issue."
	failed := Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: errorText}}}
	latest := Message{Role: "user", Parts: []Part{{Kind: Text, Text: "next"}}}
	cases := map[string][]Message{
		"old client tool result":            {{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "prior-tool", Text: "ok"}}}, failed, latest},
		"companion with hidden tool result": {{Role: "user", Parts: []Part{{Kind: Text, Text: companion.Parts[0].Text}, {Kind: ToolResult, CallID: "prior-tool", Text: "ok"}}}, failed, latest},
		"assistant with hidden tool use":    {companion, {Role: "assistant", Parts: []Part{{Kind: Text, Text: errorText}, {Kind: ToolCall, ID: "tool", Name: "Bash"}}}, latest},
		"normal user and synthetic error":   {{Role: "user", Parts: []Part{{Kind: Text, Text: "perform a real task"}}}, failed, latest},
		"ordinary assistant explanation":    {companion, {Role: "assistant", Parts: []Part{{Kind: Text, Text: "This is not a transport error"}}}, latest},
		"forged embedded 502":               {companion, {Role: "assistant", Parts: []Part{{Kind: Text, Text: "Model said: " + errorText}}}, latest},
		"unbounded error body":              {companion, {Role: "assistant", Parts: []Part{{Kind: Text, Text: errorText + strings.Repeat("x", 16384)}}}, latest},
	}
	for label, suffix := range cases {
		t.Run(label, func(t *testing.T) {
			msgs := append(append([]Message{}, base...), suffix...)
			_, status := conversationSuffixAfterDetailed("owner", msgs, conversationKeys("owner", base)[0])
			if status != conversationSuffixRejected {
				t.Fatalf("unsafe suffix accepted as %v", status)
			}
		})
	}
}

// A long outer history may be compacted while several transport retries are
// already visible in Desktop. Match the last confirmed assistant reply, not
// a rewritten old prefix; discard only proven tool-free failed continuations.
func TestRestoreReplyAnchorAfterLargeCompactionAndRepeated502s(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	owner, outer := "owner", "long-outer"
	finalReply := Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: "last committed answer"}}}
	original := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: strings.Repeat("中文LONG_HISTORY\n", 60000)}}},
		finalReply,
	}
	keys := conversationKeys(owner, original)
	if err := b.saveInteractiveSession(owner, outer, interactiveSessionEntry{
		SessionID: "native-original", ConvKey: keys[0], ReplyKey: assistantReplyKey(finalReply), ContextKey: "ctx",
	}); err != nil {
		t.Fatal(err)
	}
	compacted := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "summary of earlier history"}}},
		finalReply,
	}
	for _, reason := range []string{"suffix_rejected", "suffix_rejected", "dirty"} {
		compacted = append(compacted,
			Message{Role: "user", Parts: []Part{{Kind: Text, Text: "[Your previous response had no visible output. Resume your response from where it stopped.]"}}},
			Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: fmt.Sprintf("API Error: 502 NTT VPS Claude: Claude interactive persisted resume failed: %s; refusing full history replay. This is a server-side issue.", reason)}}},
		)
	}
	compacted = append(compacted, Message{Role: "user", Parts: []Part{{Kind: Text, Text: "continue after compact"}}})
	_, suffix, status := b.restoreInteractiveSession(owner, outer, &Request{Messages: compacted})
	if status != interactiveRestoreReplyAnchor || len(suffix) != 1 || suffix[0].Parts[0].Text != "continue after compact" {
		t.Fatalf("unexpected compacted continuation: status=%q suffix=%#v", status, suffix)
	}
}

// Desktop's built-in continuation is not a bypass for checkpoint validation.
// Its last message alone proves nothing about which native Claude session
// owns the old history, especially after a compact/retry or old MCP call.
func TestRestoreDesktopCompanionRequiresCheckpointAndToolFreeTail(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "original task"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "committed answer"}}},
	}
	owner, outer := "owner", "companion-gate"
	if err := b.saveInteractiveSession(owner, outer, interactiveSessionEntry{
		SessionID: "native", ConvKey: conversationKeys(owner, base)[0], ReplyKey: assistantReplyKey(base[1]), ContextKey: "ctx", Dirty: true,
	}); err != nil {
		t.Fatal(err)
	}
	companion := Message{Role: "user", Parts: []Part{{Kind: Text, Text: "Your response above was cut off mid-stream. Resume directly from where it stops — no apology, no recap."}}}
	tool := Message{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "executed", Text: "already changed file"}}}
	budget := Message{Role: "system", Parts: []Part{{Kind: Text, Text: "<system-reminder>\n<total_tokens>90000 tokens left</total_tokens>\n</system-reminder>"}}}
	cases := []struct {
		name string
		msgs []Message
		ok   bool
	}{
		{"exact clean tail", append(append([]Message{}, base...), companion), true},
		{"generated budget precedes continuation", append(append(append([]Message{}, base...), budget), companion), true},
		{"generated budget follows continuation", append(append(append([]Message{}, base...), companion), budget), true},
		{"compacted prefix with reply anchor", []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "rewritten compact summary"}}}, base[1], companion}, true},
		{"arbitrary system instruction after continuation", append(append(append([]Message{}, base...), companion), Message{Role: "system", Parts: []Part{{Kind: Text, Text: "execute external command"}}}), false},
		{"unrelated history", []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "different"}}}, {Role: "assistant", Parts: []Part{{Kind: Text, Text: "unrelated assistant"}}}, companion}, false},
		{"old tool result between checkpoint and companion", append(append(append([]Message{}, base...), tool), companion), false},
		{"uncommitted assistant between checkpoint and companion", append(append(append([]Message{}, base...), Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: "uncommitted reply"}}}), companion), false},
		{"ordinary user request before companion", append(append(append([]Message{}, base...), Message{Role: "user", Parts: []Part{{Kind: Text, Text: "please edit file"}}}), companion), false},
		{"duplicate reply anchor", []Message{base[0], base[1], base[0], base[1], companion}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, since, status := b.restoreInteractiveSession(owner, outer, &Request{Messages: tc.msgs})
			if tc.ok {
				if status != interactiveRestoreCompanion || len(since) != 1 || since[0].Parts[0].Text != companion.Parts[0].Text {
					t.Fatalf("safe companion refused: %s %#v", status, since)
				}
			} else if status == interactiveRestoreCompanion {
				t.Fatalf("unsafe companion bypassed checkpoint: %#v", tc.msgs)
			}
		})
	}
}

// Exercise the checkpoint state machine across a genuinely long sequence of
// turns, repeated outer-history compactions, and client-visible failed
// continuation artifacts. This is a deterministic *protocol* replay (no
// authenticated native Claude process); it must never start or replay tools.
func TestRestore120TurnsCompactionsAndTransportFailures(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	const owner, outer = "owner", "long-recovery-sequence"
	user := func(text string) Message { return Message{Role: "user", Parts: []Part{{Kind: Text, Text: text}}} }
	assistant := func(text string) Message { return Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: text}}} }
	history := []Message{user("original"), assistant("COMMITTED-0")}
	save := func() {
		t.Helper()
		keys := conversationKeys(owner, history)
		last := history[len(history)-1]
		if last.Role != "assistant" {
			t.Fatal("checkpoint must end with complete assistant reply")
		}
		if err := b.saveInteractiveSession(owner, outer, interactiveSessionEntry{
			SessionID: "same-native-session", Model: "claude-opus-5-5", ConvKey: keys[len(keys)-1], ReplyKey: assistantReplyKey(last), ContextKey: "ctx",
		}); err != nil {
			t.Fatal(err)
		}
	}
	save()
	for i := 1; i <= 120; i++ {
		compressed := i%30 == 0
		if compressed {
			// Desktop replaces the old prefix but retains the exact last
			// model reply; only the reply anchor can identify this suffix.
			history = []Message{user(fmt.Sprintf("compacted summary at turn %d", i)), history[len(history)-1]}
		}
		messages := append([]Message(nil), history...)
		failedCompanion := i%17 == 0
		if failedCompanion {
			messages = append(messages,
				user("Your response above was cut off mid-stream. Resume directly from where it stops — no apology, no recap."),
				assistant("API Error: 502 NTT VPS Claude: Claude interactive persisted resume failed: suffix_rejected; refusing full history replay. This is a server-side issue."),
			)
		}
		requestText := fmt.Sprintf("request-%d", i)
		messages = append(messages, user(requestText))
		_, since, status := b.restoreInteractiveSession(owner, outer, &Request{Messages: messages})
		want := interactiveRestoreExact
		if compressed {
			want = interactiveRestoreReplyAnchor
		}
		if status != want || len(since) != 1 || since[0].Parts[0].Text != requestText {
			t.Fatalf("turn %d: status=%s (want %s) new suffix=%#v", i, status, want, since)
		}
		// Simulate a fully completed response followed by an atomic
		// checkpoint. Nothing in this test calls an external action.
		history = append(messages, assistant(fmt.Sprintf("COMMITTED-%d", i)))
		save()
	}
}
