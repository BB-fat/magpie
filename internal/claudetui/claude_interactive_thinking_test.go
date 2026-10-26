package claudetui

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

const interactiveSignedRows = `{"type":"assistant","uuid":"signed","session_id":"inner","message":{"id":"msg_signed","model":"claude-sonnet-5","content":[{"type":"thinking","thinking":"first","signature":"sig-first"},{"type":"thinking","thinking":"","signature":"sig-empty"},{"type":"thinking","thinking":"last","signature":"sig-last"},{"type":"text","text":"answer"}]}}` + "\n" +
	`{"type":"result","session_id":"inner","stop_reason":"end_turn"}`

func TestInteractiveThinkingKeepsSignedBlocks(t *testing.T) {
	run := &subscriptionRun{}
	events := run.attach()
	go run.readInteractiveOneTurn(strings.NewReader(interactiveSignedRows))
	var col collector
	for ev := range events {
		col.add(ev)
	}
	res := col.finish()
	if col.err != "" {
		t.Fatal(col.err)
	}
	if len(res.Parts) != 4 {
		t.Fatalf("got %d blocks: %+v; want three separate signed thinking blocks and text", len(res.Parts), res.Parts)
	}
	for i, want := range []Part{
		{Kind: Thinking, Text: "first", Signature: "sig-first"},
		{Kind: Thinking, Signature: "sig-empty"},
		{Kind: Thinking, Text: "last", Signature: "sig-last"},
		{Kind: Text, Text: "answer"},
	} {
		if got := res.Parts[i]; got.Kind != want.Kind || got.Text != want.Text || got.Signature != want.Signature {
			t.Errorf("block %d = %+v, want %+v", i, got, want)
		}
	}
}

func TestLiteRebuiltSessionKeepsSignedThinking(t *testing.T) {
	req := &Request{Model: "claude-sonnet-5", Messages: []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "hi"}}},
		{Role: "assistant", Parts: []Part{{Kind: Thinking, Signature: "sig-empty"}, {Kind: Thinking, Text: "reason", Signature: "sig-reason"}, {Kind: Text, Text: "answer"}}},
		{Role: "user", Parts: []Part{{Kind: Text, Text: "continue"}}},
	}}
	messages, _ := liteSeedTranscript(t, req)
	if len(messages) != 2 || len(messages[1].Content) != 3 {
		t.Fatalf("rebuilt signed history changed: %+v", messages)
	}
	for i, signature := range []string{"sig-empty", "sig-reason"} {
		if block := messages[1].Content[i]; block.Type != "thinking" || block.Signature != signature {
			t.Fatalf("rebuilt thinking %d = %+v", i, block)
		}
	}
}

func TestInteractiveThinkingStreamRoundTrip(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(map[bool]string{false: "whole", true: "stream"}[stream], func(t *testing.T) {
			run := &subscriptionRun{}
			events := run.attach()
			go run.readInteractiveOneTurn(strings.NewReader(interactiveSignedRows))
			out := httptest.NewRecorder()
			request := httptest.NewRequest("POST", "/v1/messages", nil)
			_, msg := (anthropicResponder{w: out, r: request, usage: &Usage{}}).Relay(&Request{Stream: stream}, events, func() {}, func(string, string, bool) {})
			if msg != "" {
				t.Fatal(msg)
			}
			body := out.Body.Bytes()
			if stream {
				var col collector
				for _, line := range strings.Split(out.Body.String(), "\n") {
					if data, ok := strings.CutPrefix(line, "data: "); ok {
						if err := decodeAnthropic(data, col.add); err != nil {
							t.Fatal(err)
						}
					}
				}
				body = renderAnthropic(col.finish(), "claude-sonnet-5")
			}
			var reply struct {
				Content []aBlock `json:"content"`
			}
			if err := json.Unmarshal(body, &reply); err != nil {
				t.Fatal(err)
			}
			if len(reply.Content) != 4 || reply.Content[1].Signature != "sig-empty" {
				t.Fatalf("signed blocks lost: %s", body)
			}
		})
	}
}
