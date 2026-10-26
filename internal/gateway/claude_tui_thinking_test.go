package gateway

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/provider"
)

func TestClaudeTUIThinkingSurvivesMessagesReplay(t *testing.T) {
	rows := []string{
		`{"type":"assistant","uuid":"signed","session_id":"inner","message":{"id":"msg_signed","model":"claude-sonnet-5","content":[{"type":"thinking","thinking":"first","signature":"sig-first"},{"type":"thinking","thinking":"","signature":"sig-empty"},{"type":"thinking","thinking":"last","signature":"sig-last"},{"type":"text","text":"answer"}]}}`,
		`{"type":"result","session_id":"inner","stop_reason":"end_turn"}`,
	}
	p := tuiAccount(t, tuiWrapper(t, rows...))
	s := tuiTestServer(t)
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprintf("stream=%t", stream), func(t *testing.T) {
			body := fmt.Sprintf(`{"model":"claude-sonnet-5","max_tokens":64,"stream":%t,"messages":[{"role":"user","content":"hi"}]}`, stream)
			out := httptest.NewRecorder()
			code, msg := s.serveClaudeSubscription(out, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body)), provider.Anthropic, p, "claude-sonnet-5", []byte(body), &Usage{})
			if code != 200 {
				t.Fatalf("%d %s: %s", code, msg, out.Body)
			}
			var res Result
			if stream {
				var col collector
				for _, line := range strings.Split(out.Body.String(), "\n") {
					if data, ok := strings.CutPrefix(line, "data: "); ok {
						if err := decodeAnthropic(data, col.add); err != nil {
							t.Fatal(err)
						}
					}
				}
				res = col.finish()
			} else {
				var reply struct {
					Content []aBlock `json:"content"`
				}
				if err := json.Unmarshal(out.Body.Bytes(), &reply); err != nil {
					t.Fatal(err)
				}
				for _, b := range reply.Content {
					kind := Text
					if b.Type == "thinking" {
						kind = Thinking
					}
					text := b.Text
					if kind == Thinking {
						text = b.Thinking
					}
					res.Parts = append(res.Parts, Part{Kind: kind, Text: text, Signature: b.Signature})
				}
			}
			want := []Part{{Kind: Thinking, Text: "first", Signature: "sig-first"}, {Kind: Thinking, Signature: "sig-empty"}, {Kind: Thinking, Text: "last", Signature: "sig-last"}, {Kind: Text, Text: "answer"}}
			if len(res.Parts) != len(want) {
				t.Fatalf("got %d blocks: %+v", len(res.Parts), res.Parts)
			}
			for i, p := range want {
				if got := res.Parts[i]; got.Kind != p.Kind || got.Text != p.Text || got.Signature != p.Signature {
					t.Fatalf("block %d = %+v, want %+v", i, got, p)
				}
			}
			replay := &Request{Model: "claude-sonnet-5", Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "hi"}}}, {Role: "assistant", Parts: res.Parts}, {Role: "user", Parts: []Part{{Kind: Text, Text: "continue"}}}}}
			parsed, err := parseAnthropic(buildAnthropic(replay, replay.Model))
			if err != nil || len(parsed.Messages) != 3 || len(parsed.Messages[1].Parts) != 4 {
				t.Fatalf("replay lost signed blocks: %+v, %v", parsed, err)
			}
			for i := range 3 {
				if got := parsed.Messages[1].Parts[i]; got.Signature != want[i].Signature || got.Text != want[i].Text {
					t.Fatalf("replayed block %d changed: %+v", i, got)
				}
			}
		})
	}
}
