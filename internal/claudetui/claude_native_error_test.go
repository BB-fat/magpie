package claudetui

import (
	"encoding/json"
	"testing"
)

func TestNativeErrorResultText(t *testing.T) {
	for _, tc := range []struct {
		name, result string
		status       int
		errorField   json.RawMessage
		want         string
	}{
		{"prompt too long propagated", "", 400, json.RawMessage(`"Claude Code API error: Prompt is too long"`), "Claude Code API error: Prompt is too long"},
		{"preserve ordinary result", "existing result", 400, json.RawMessage(`"ignored error"`), "existing result"},
		{"do not trust unknown error without status", "", 0, json.RawMessage(`"unproven diagnostic"`), ""},
		{"a StopFailure rate limit without status", "", 0, json.RawMessage(`"Claude StopFailure hook fired: rate_limit\nPane tail:\nprivate pane text"`), "Claude Code reached its usage or rate limit"},
		{"limit words only in the pane tail", "", 0, json.RawMessage(`"Claude StopFailure hook fired: server_error\nPane tail:\nwe discussed the rate limit"`), ""},
		{"reject non-string error field", "", 400, json.RawMessage(`{"message":"untrusted"}`), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := nativeErrorResultText(tc.result, tc.status, tc.errorField); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
