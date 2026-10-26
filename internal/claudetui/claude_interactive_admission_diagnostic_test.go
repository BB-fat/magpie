package claudetui

import (
	"strings"
	"testing"
)

func TestRecoveryAdmissionDiagnosticRetrySuffix(t *testing.T) {
	user := func(s string) Message { return Message{Role: "user", Parts: []Part{{Kind: Text, Text: s}}} }
	failed := Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: "API Error: 502 NTT VPS Claude: Claude interactive persisted resume failed: dirty; refusing full history replay"}}}
	base := []Message{user("PRIVATE_PREFIX"), {Role: "assistant", Parts: []Part{{Kind: ToolCall, CallID: "SECRET_ID"}}}, {Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "SECRET_ID", Text: "SECRET_RESULT"}}}, failed, user("PRIVATE_RETRY"), failed, user("PRIVATE_NEW")}
	cases := []struct {
		name  string
		last  Message
		want  string
		optin bool
	}{
		{"ordinary retry", user("PRIVATE_NEW"), "optin_block=missing_sentinel", false},
		{"exact opt in", user("/magpie-recover"), "optin_block=none", true},
		{"embedded opt in", user("PRIVATE /magpie-recover"), "optin_block=missing_sentinel", false},
		{"text after sentinel", Message{Role: "user", Parts: []Part{{Kind: Text, Text: "/magpie-recover"}, {Kind: Text, Text: "PRIVATE_AFTER"}}}, "optin_block=sentinel_not_last", false},
		{"reminder mixed in", Message{Role: "user", Parts: []Part{{Kind: Text, Text: "/magpie-recover"}, {Kind: Text, Text: "<system-reminder>PRIVATE</system-reminder>"}}}, "optin_block=shape_rejected", false},
		{"sentinel inside result", Message{Role: "user", Parts: []Part{{Kind: ToolResult, Text: "/magpie-recover"}, {Kind: Text, Text: "PRIVATE_NEW"}}}, "optin_block=missing_sentinel", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msgs := append([]Message(nil), base...)
			msgs[len(msgs)-1] = tc.last
			req := &Request{Messages: msgs}
			before := interactiveReadOnlyOptInIndex(req)
			got := interactiveRestoreRejectionDiagnostic(interactiveSessionEntry{Dirty: true}, req)
			if !strings.Contains(got, tc.want) || !strings.Contains(got, "transport_502=true") {
				t.Fatalf("missing admission evidence: %s", got)
			}
			if (before >= 0) != tc.optin || interactiveReadOnlyOptInIndex(req) != before {
				t.Fatal("diagnostics changed recovery admission")
			}
			if len(got) > 1024 || strings.Contains(got, "PRIVATE") || strings.Contains(got, "SECRET") || strings.Contains(got, "/magpie-recover") {
				t.Fatalf("unsafe diagnostic: %s", got)
			}
		})
	}
}
