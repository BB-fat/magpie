package gateway

import (
	"strings"
	"testing"
)

// Diagnostic data must explain why resume was refused without exposing
// conversation content, tool names/results, authenticated IDs or their hashes.
func TestClaudeTUIRuntimeInteractiveRestoreRejectionDiagnosticIsActionablePrivateAndBounded(t *testing.T) {
	const owner = "OWNER_SECRET_ID"
	user := func(s string) Message { return Message{Role: "user", Parts: []Part{{Kind: Text, Text: s}}} }
	answer := Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: "SECRET_COMMITTED_REPLY"}}}
	base := []Message{user("SECRET_ORIGINAL_PROMPT"), answer}
	saved := tuiRuntimeInteractiveSessionEntry{ConvKey: tuiRuntimeConversationKeys(owner, base)[0], ReplyKey: tuiRuntimeAssistantReplyKey(answer)}
	cases := []struct {
		name string
		msgs []Message
		want string
	}{
		{"unmatched historical prefix", []Message{user("OTHER_PRIVATE_PREFIX"), {Role: "assistant", Parts: []Part{{Kind: Text, Text: "SECRET_OTHER_REPLY"}}}, user("SECRET_NEW")}, "anchor=none"},
		{"duplicate reply anchor", []Message{user("CHANGED_PREFIX"), answer, user("PRIVATE_MIDDLE"), answer, user("SECRET_NEW")}, "anchor=ambiguous"},
		{"uncommitted assistant after checkpoint", append(append([]Message{}, base...), Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: "SECRET_UNCOMMITTED_REPLY"}}}, user("SECRET_NEW")), "suffix=intervening_assistant"},
		{"historical tool result after checkpoint", append(append([]Message{}, base...), Message{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "SECRET_TOOL_ID", Text: "SECRET_TOOL_OUTPUT"}}}, user("SECRET_NEW")), "suffix=unmatched_tool_result"},
		{"no new user after checkpoint", append([]Message{}, base...), "suffix=empty"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &tuiRuntimeRequest{Request: Request{Messages: tc.msgs}}
			got := tuiRuntimeInteractiveRestoreRejectionDiagnostic(saved, req)
			if !strings.Contains(got, tc.want) {
				t.Fatalf("missing reason %q: %s", tc.want, got)
			}
			if len(got) > 1024 {
				t.Fatalf("unbounded diagnostic: %d bytes", len(got))
			}
			for _, secret := range []string{"SECRET", "OWNER", "PRIVATE", "CHANGED_PREFIX", "OTHER_PRIVATE", "COMMITTED_REPLY"} {
				if strings.Contains(got, secret) {
					t.Fatalf("diagnostic leaked content: %s", got)
				}
			}
		})
	}
}

func tuiRuntimeBenchmarkInteractiveRestoreRejectionDiagnosticLargeHistory(b *testing.B) {
	owner := "bench-owner"
	last := Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: "committed"}}}
	old := []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: strings.Repeat("multi turn archive 中文 ABCDEF\n", 65000)}}}, last}
	saved := tuiRuntimeInteractiveSessionEntry{ConvKey: tuiRuntimeConversationKeys(owner, old)[0], ReplyKey: tuiRuntimeAssistantReplyKey(last)}
	req := &tuiRuntimeRequest{Request: Request{Messages: append(append([]Message{}, old...), Message{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "old-tool", Text: "already executed"}}})}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if got := tuiRuntimeInteractiveRestoreRejectionDiagnostic(saved, req); !strings.Contains(got, "unmatched_tool_result") {
			b.Fatal(got)
		}
	}
}
