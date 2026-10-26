package gateway

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/claudebridge"
)

type tuiRuntimeLiteTranscriptMessage struct {
	Role    string   `json:"role"`
	Content []aBlock `json:"content"`
}

func tuiRuntimeLiteSeedTranscript(t *testing.T, req *Request) ([]tuiRuntimeLiteTranscriptMessage, string) {
	t.Helper()
	seeded, prompt, _, err := tuiRuntimeLiteSeed(req)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	_, path, err := tuiRuntimeWriteClaudeTranscript(dir, filepath.Join(dir, "config"), req.Model, nil, seeded.seedHistory, tuiRuntimeMcpToolPrefix(), "")
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var messages []tuiRuntimeLiteTranscriptMessage
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var row struct {
			Message tuiRuntimeLiteTranscriptMessage `json:"message"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		messages = append(messages, row.Message)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return messages, prompt
}

func TestClaudeTUIRuntimeLiteRebuildJSONLKeepsSplitRoleRuns(t *testing.T) {
	for _, trailing := range []bool{false, true} {
		name := "current tool round"
		if trailing {
			name = "historical tool round"
		}
		t.Run(name, func(t *testing.T) {
			req := &Request{Model: "claude-opus-5-5", Messages: []Message{
				{Role: "user", Parts: []Part{{Kind: Text, Text: "go"}}},
				{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "a", Name: "spawn.agent", Args: json.RawMessage(`{"path":"a"}`)}}},
				{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "b", Name: "echo"}}},
				{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "a", Text: "RESULT_A"}}},
				{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "b", Text: "RESULT_B", IsError: true, Images: []Part{{Kind: Image, MediaType: "image/png", Data: "QUFB"}}}, {Kind: Text, Text: "USER_NEXT"}}},
			}}
			if err := tuiRuntimeLiteVisibleToolHandoff(req); err != nil {
				t.Fatalf("valid split role-run was rejected: %v", err)
			}
			if trailing {
				req.Messages = append(req.Messages, Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: "done"}}}, Message{Role: "user", Parts: []Part{{Kind: Text, Text: "FINAL_USER"}}})
			}
			messages, prompt := tuiRuntimeLiteSeedTranscript(t, req)
			calls, results := map[string]aBlock{}, map[string]aBlock{}
			var historicalText strings.Builder
			for _, m := range messages {
				for _, block := range m.Content {
					switch block.Type {
					case "tool_use":
						calls[block.ID] = block
					case "tool_result":
						results[block.ToolUseID] = block
					case "text":
						historicalText.WriteString(block.Text)
					}
				}
			}
			if len(calls) != 2 || len(results) != 2 {
				t.Fatalf("split parallel exchange lost or unmatched call retained: calls=%+v results=%+v", calls, results)
			}
			if calls["a"].Name != tuiRuntimeMcpToolPrefix()+claudebridge.ToolName("spawn.agent") {
				t.Fatalf("call is not a native MCP tool: %+v", calls["a"])
			}
			if string(calls["a"].Input) != `{"path":"a"}` || string(calls["b"].Input) != `{}` {
				t.Fatalf("call arguments changed or empty input is not an object: %+v", calls)
			}
			for _, id := range []string{"a", "b"} {
				text, images := toolOutput(results[id].Content)
				if text != "RESULT_"+strings.ToUpper(id) {
					t.Fatalf("result %s lost its content: %s", id, results[id].Content)
				}
				if id == "b" && (!results[id].IsError || len(images) != 1 || images[0].Data != "QUFB" || images[0].MediaType != "image/png") {
					t.Fatalf("result error/image lost: %+v %s", results[id], results[id].Content)
				}
			}
			if trailing {
				if !strings.Contains(historicalText.String(), "USER_NEXT") || !strings.Contains(prompt, "FINAL_USER") {
					t.Fatalf("historical/current user text lost: history=%q prompt=%q", historicalText.String(), prompt)
				}
			} else if !strings.Contains(prompt, "USER_NEXT") || strings.Contains(prompt, "RESULT_A") || strings.Contains(prompt, "RESULT_B") {
				t.Fatalf("current user text mixed with results: %q", prompt)
			}
			if req.Messages[2].Parts[0].Args != nil {
				t.Fatal("rebuild mutated the caller's empty arguments")
			}
		})
	}
}

func TestClaudeTUIRuntimeLiteRebuildJSONLKeepsOrphanOutput(t *testing.T) {
	for _, failed := range []bool{false, true} {
		name := "success"
		if failed {
			name = "error"
		}
		t.Run(name, func(t *testing.T) {
			req := &Request{Model: "claude-opus-5-5", Messages: []Message{
				{Role: "user", Parts: []Part{{Kind: Text, Text: "compacted context"}}},
				{Role: "assistant", Parts: []Part{{Kind: Text, Text: "summary"}}},
				{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "earlier-call", Text: "IMPORTANT_OUTPUT", IsError: failed, Images: []Part{{Kind: Image, MediaType: "image/png", Data: "QUFB"}, {Kind: Image, URL: "https://example.invalid/image"}}}, {Kind: Text, Text: "USER_NEXT"}}},
			}}
			if err := tuiRuntimeLiteVisibleToolHandoff(req); err != nil {
				t.Fatal(err)
			}
			messages, prompt := tuiRuntimeLiteSeedTranscript(t, req)
			var text strings.Builder
			var images []aBlock
			for _, m := range messages {
				for _, block := range m.Content {
					switch block.Type {
					case "tool_use", "tool_result":
						t.Fatalf("orphan became native tool traffic: %+v", block)
					case "text":
						text.WriteString(block.Text)
					case "image":
						images = append(images, block)
					}
				}
			}
			marker := "[tool result id=earlier-call"
			if failed {
				marker += " error"
			}
			if !strings.Contains(text.String(), marker+"]\nIMPORTANT_OUTPUT") || len(images) != 2 || images[0].Source == nil || images[0].Source.Data != "QUFB" || images[1].Source == nil || images[1].Source.URL != "https://example.invalid/image" {
				t.Fatalf("orphan content/error/media lost: text=%q images=%+v", text.String(), images)
			}
			if prompt != "USER_NEXT" {
				t.Fatalf("user text is not its own current turn: %q", prompt)
			}
		})
	}
}

func TestClaudeTUIRuntimeLiteRebuildJSONLNormalizesEmptyArguments(t *testing.T) {
	for name, args := range map[string]json.RawMessage{"missing": nil, "empty": json.RawMessage(" \n"), "null": json.RawMessage("null"), "object": json.RawMessage("{}")} {
		t.Run(name, func(t *testing.T) {
			req := tuiRuntimeReplayReq(t, tuiRuntimeReplayBody(""))
			req.Messages[1].Parts[0].Args = args
			messages, _ := tuiRuntimeLiteSeedTranscript(t, req)
			for _, m := range messages {
				for _, block := range m.Content {
					if block.Type == "tool_use" {
						if string(block.Input) != "{}" {
							t.Fatalf("empty arguments became native input %s", block.Input)
						}
						return
					}
				}
			}
			t.Fatal("the answered tool call disappeared")
		})
	}
}

func TestClaudeTUIRuntimeLiteRebuildRejectsVisibleWrongResultID(t *testing.T) {
	req := tuiRuntimeReplayReq(t, strings.Replace(tuiRuntimeReplayBody(""), `"tool_use_id":"call-1"`, `"tool_use_id":"wrong"`, 1))
	if err := tuiRuntimeLiteVisibleToolHandoff(req); err == nil {
		t.Fatal("visible wrong CallID became an allowed orphan")
	}
}
