package gateway

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

// streamedContent is an Anthropic reply's content as a client that read
// its stream keeps it: every block, whole, in order.
func streamedContent(t *testing.T, body string) []map[string]any {
	t.Helper()
	var content []map[string]any
	partial := map[int]string{}
	for _, ev := range events(body) {
		switch ev["type"] {
		case "content_block_start":
			content = append(content, ev["content_block"].(map[string]any))
		case "content_block_delta":
			i := int(ev["index"].(float64))
			d := ev["delta"].(map[string]any)
			b := content[i]
			switch d["type"] {
			case "text_delta":
				b["text"] = b["text"].(string) + d["text"].(string)
			case "thinking_delta":
				b["thinking"] = b["thinking"].(string) + d["thinking"].(string)
			case "signature_delta":
				s, _ := b["signature"].(string)
				b["signature"] = s + d["signature"].(string)
			case "citations_delta":
				cs, _ := b["citations"].([]any)
				b["citations"] = append(cs, d["citation"])
			case "input_json_delta":
				partial[i] += d["partial_json"].(string)
			}
		case "content_block_stop":
			i := int(ev["index"].(float64))
			if s, ok := partial[i]; ok {
				var in any
				if err := json.Unmarshal([]byte(s), &in); err != nil {
					t.Fatalf("block %d's input %q: %v", i, s, err)
				}
				content[i]["input"] = in
			}
		}
	}
	return content
}

// A Messages client's next request holds the reply as the stream's blocks,
// and the recorder keeps it so: its prefix is the one that request names,
// so the turn after it finds the idle process or the checkpoint. A search
// is two blocks of its own, and after it a line naming a page found is a
// text block, so a reply with a search used to be kept as one text part
// and never matched.
func TestClaudeTUIRuntimeReplyRecorderKeepsMessagesBlocks(t *testing.T) {
	release := []Hit{{Title: "Release History", URL: "https://go.dev/doc/devel/release"}}
	for _, tc := range []struct {
		name   string
		events []Event
		blocks int // the content blocks the client is sent
	}{
		{"text", []Event{
			{Kind: KText, Text: "Go 1.27.2 "}, {Kind: KText, Text: "is the latest."},
		}, 1},
		{"thinking, signature and text", []Event{
			{Kind: KThinkStart}, {Kind: KThink, Text: "the user asks"}, {Kind: KSig, Text: "sig-1"},
			{Kind: KText, Text: "Go 1.27.2."},
		}, 2},
		{"text and a tool call with an id Messages can't take", []Event{
			{Kind: KText, Text: "Reading it."},
			{Kind: KToolStart, ID: "Bash:0#a1", Name: "Read"}, {Kind: KToolArgs, Text: `{"path": "go.mod"}`},
		}, 2},
		{"text, search, text with a cited line", []Event{
			{Kind: KText, Text: "我查一下。"},
			{Kind: KSearch, Text: "Go latest release", Hits: release},
			{Kind: KText, Text: "最新是 Go 1.27.2，\n见 https://go.dev/doc/devel/release\n发布于 10 月。"},
		}, 6},
		{"search first, text with a cited line", []Event{
			{Kind: KSearch, Text: "Go latest release", Hits: release},
			{Kind: KText, Text: "最新是 Go 1.27.2，\n见 https://go.dev/doc/devel/release\n发布于 10 月。"},
		}, 5},
		{"text, search, text citing nothing", []Event{
			{Kind: KText, Text: "我查一下。"},
			{Kind: KSearch, Text: "Go latest release", Hits: release},
			{Kind: KText, Text: "最新是 Go 1.27.2。"},
		}, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			enc := &anthropicEncoder{w: newSSEWriter(rec), model: "m"}
			kept := tuiRuntimeNewReplyRecorder(protoAnthropic)
			for _, ev := range append(tc.events, Event{Kind: KStop, Stop: "stop"}) {
				kept.note(ev)
				enc.event(ev)
			}
			enc.finish()
			content := streamedContent(t, rec.Body.String())
			if len(content) != tc.blocks {
				t.Fatalf("the client was sent %d blocks, want %d: %v", len(content), tc.blocks, content)
			}
			body, err := json.Marshal(map[string]any{"model": "m", "max_tokens": 1, "messages": []any{
				map[string]any{"role": "user", "content": "q1"},
				map[string]any{"role": "assistant", "content": content},
				map[string]any{"role": "user", "content": "q2"},
			}})
			if err != nil {
				t.Fatal(err)
			}
			next, err := tuiRuntimeParseRequest(body)
			if err != nil {
				t.Fatal(err)
			}
			client := next.Messages[:2]
			ours := []Message{next.Messages[0], kept.message()}
			if tuiRuntimePrefixOf("s", client) != tuiRuntimePrefixOf("s", ours) {
				t.Fatalf("the kept reply is not the one the next request holds:\nclient %#v\nkept   %#v", client[1].Parts, ours[1].Parts)
			}
		})
	}
}

// A chat completions client (or a gateway in front of one) sends the reply
// back without its thinking. Its next turn still finds the checkpoint the
// reply made: au-exit 10-29 rebuilt every turn of such a conversation,
// writing its whole history to the cache again each time.
func TestClaudeTUIRuntimeNativeCheckpointFoundWithoutThinking(t *testing.T) {
	tuiRuntimeTestInit(t, &tuiRuntimeConfig{StateDir: t.TempDir()})
	scope := tuiRuntimePrefixScope("owner", "")
	q1 := Message{Role: "user", Parts: []Part{{Kind: Text, Text: "q1"}}}
	reply := Message{Role: "assistant", Parts: []Part{
		{Kind: Thinking, Text: "the user asks", Signature: "sig-1"},
		{Kind: Text, Text: "a1"},
		{Kind: ToolCall, ID: "c1", Name: "Read", Args: json.RawMessage(`{"path":"go.mod"}`)},
	}}
	prefix := tuiRuntimePrefixOf(scope, []Message{q1, reply})
	if err := tuiRuntimeSaveNativeCheckpoint(tuiRuntimeNativeCheckpoint{
		Prefix: tuiRuntimeNativeCheckpointKey(scope, prefix)[len(scope)+1:], Scope: scope,
		SessionID: "sess", Row: "row", Work: "work", Model: "m", InstructionKey: "k",
	}); err != nil {
		t.Fatal(err)
	}
	resent := Message{Role: "assistant", Parts: reply.Parts[1:]}
	q2 := Message{Role: "user", Parts: []Part{{Kind: Text, Text: "q2"}}}
	req := &tuiRuntimeRequest{scope: scope}
	req.Messages = []Message{q1, resent, q2}
	p, since, ok := tuiRuntimeFindNativeCheckpoint(req, "m")
	if !ok || p.SessionID != "sess" || len(since) != 1 {
		t.Fatalf("the checkpoint was not found from a history without thinking: ok=%t %+v since=%d", ok, p, len(since))
	}
	// what was said still counts
	other := Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: "a2"}, reply.Parts[2]}}
	req.Messages = []Message{q1, other, q2}
	if _, _, ok := tuiRuntimeFindNativeCheckpoint(req, "m"); ok {
		t.Fatal("a reply with other text matched the checkpoint")
	}
}
