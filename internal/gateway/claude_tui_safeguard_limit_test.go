package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestClaudeTUIInteractiveSafeguardLimit(t *testing.T) {
	empty := &tuiRuntimeRequest{Request: Request{Safeguards: json.RawMessage(`[""]`)}}
	overhead := len(tuiRuntimeSafeguardBody(empty.Safeguards))
	for _, size := range []int{tuiRuntimeInteractiveSafeguardBodyMax - 1, tuiRuntimeInteractiveSafeguardBodyMax, tuiRuntimeInteractiveSafeguardBodyMax + 1} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			req := &tuiRuntimeRequest{Request: Request{Safeguards: json.RawMessage(`["` + strings.Repeat("x", size-overhead) + `"]`)}}
			err := tuiRuntimeValidateInteractiveSafeguards(req)
			if size <= tuiRuntimeInteractiveSafeguardBodyMax {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("serialized body: %d bytes", size)) {
				t.Fatalf("expected explicit size error, got %v", err)
			}
		})
	}
	if err := tuiRuntimeValidateInteractiveSafeguards(&tuiRuntimeRequest{}); err != nil {
		t.Fatal(err)
	}
}

func TestClaudeTUIInteractiveOversizedSafeguardsNeverLaunch(t *testing.T) {
	for _, text := range []string{strings.Repeat("x", 200<<10), strings.Repeat("\u754c", 24<<10)} {
		raw, err := json.Marshal([]map[string]string{{"type": "dangerous_tool_use", "classifier_context": text}})
		if err != nil {
			t.Fatal(err)
		}
		req := &tuiRuntimeRequest{Request: Request{Safeguards: raw, Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}}}}}
		b := &tuiRuntimeSubscriptionBridge{}
		check := func(err error) {
			t.Helper()
			if err == nil || !strings.Contains(err.Error(), "request rejected without truncating safeguards") {
				t.Fatalf("expected unsupported safeguards error before launch/control access, got %v", err)
			}
		}
		check(tuiRuntimeValidateInteractiveFirstTurn(req))
		// No wrapper or Claude configuration: rejection must precede dependency lookup.
		_, _, err = b.startInteractiveWithPromptImages(context.Background(), req, "", "", "", "", "hello", "", nil, false)
		check(err)
		_, _, err = b.startInteractiveResumeTurn(context.Background(), req, "", "", "", "outer", tuiRuntimeInteractiveSessionEntry{SessionID: "inner"}, req.Messages, false)
		check(err)
		run := &tuiRuntimeSubscriptionRun{interactive: true}
		check(run.setSafeguards(req))
		if string(req.Safeguards) != string(raw) {
			t.Fatal("rejected safeguards were changed")
		}
	}
}
