//go:build darwin || linux

package gateway

import (
	"bytes"
	"log"
	"reflect"
	"strings"
	"testing"
)

func TestClaudeTUIRuntimeOrphanRecoveryRefusalDiagnosticsDoNotMutateOrLeak(t *testing.T) {
	cases := []struct{ name, checkpoint, model, reason string }{
		{"disabled", "dirty", tuiRuntimeAutoTestModel, "disabled"},
		{"new fork without state", "missing", tuiRuntimeAutoTestModel, "checkpoint_missing"},
		{"clean state", "clean", tuiRuntimeAutoTestModel, "checkpoint_clean"},
		{"wrong model", "dirty", "SECRET_MODEL", "model_mismatch"},
		{"unchanged prefix with unmatched result", "dirty", tuiRuntimeAutoTestModel, "checkpoint_suffix_rejected"},
		{"unchanged prefix with valid suffix", "dirty", tuiRuntimeAutoTestModel, "checkpoint_prefix_present"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tuiRuntimeArDir(t)
			t.Setenv(tuiRuntimeAutoOrphanRecoveryEnv, "1")
			if tc.name == "disabled" {
				t.Setenv(tuiRuntimeAutoOrphanRecoveryEnv, "0")
			}
			b := tuiRuntimeNewSubscriptionBridge()
			defer tuiRuntimeBridges.Delete(b)
			const owner, outer = "SECRET_OWNER", "SECRET_SESSION"
			if tc.checkpoint != "missing" {
				tuiRuntimeArSave(t, b, owner, outer, "SECRET_NATIVE", tc.checkpoint == "dirty")
			}
			req := &Request{Messages: []Message{
				{Role: "user", Parts: []Part{{Kind: Text, Text: "先读一下配置"}}},
				{Role: "assistant", Parts: []Part{{Kind: Text, Text: "配置已读"}}},
				{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "SECRET_CALL", Text: "SECRET_RESULT"}}},
			}}
			before := append([]Message(nil), req.Messages...)
			if tc.reason == "checkpoint_prefix_present" {
				req.Messages[len(req.Messages)-1] = Message{Role: "user", Parts: []Part{{Kind: Text, Text: "SECRET_NEW_USER_TEXT"}}}
				before = append([]Message(nil), req.Messages...)
			}
			var out bytes.Buffer
			previous := log.Writer()
			log.SetOutput(&out)
			defer log.SetOutput(previous)
			if b.beginOrphanReadOnlyRecovery(owner, outer, tc.model, req) {
				t.Fatal("refused fixture was admitted")
			}
			if !reflect.DeepEqual(before, req.Messages) {
				t.Fatal("diagnostics mutated failed request")
			}
			got := out.String()
			if !strings.Contains(got, "reason="+tc.reason) {
				t.Fatalf("missing reason: %s", got)
			}
			if strings.Contains(got, "SECRET") || len(got) > 200 {
				t.Fatalf("unsafe diagnostic: %s", got)
			}
		})
	}
}
