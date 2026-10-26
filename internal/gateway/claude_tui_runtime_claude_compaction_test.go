package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const tuiRuntimeCompactSystemTest = "You are a helpful AI assistant tasked with summarizing conversations."

func TestClaudeTUIRuntimeCompactionHistoricalToolResultStartsReadOnly(t *testing.T) {
	body := `{"model":"claude-opus-5-5","max_tokens":100,"system":"` + tuiRuntimeCompactSystemTest + `","tools":[{"name":"Bash","input_schema":{"type":"object"}}],"messages":[{"role":"assistant","content":[{"type":"tool_use","id":"old","name":"Bash","input":{"command":"NEVER_EXECUTE"}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"old","content":"KEEP_RESULT"},{"type":"text","text":"Your task is to create a detailed summary of the conversation so far."}]}]}`
	rt := tuiRuntimeNewRuntime([]byte("test"))
	called := false
	start := func(ctx context.Context, req *tuiRuntimeRequest) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
		called = true
		if len(req.Tools) != 0 || tuiRuntimeInteractiveToolContinuation(req) || req.WebSearch {
			t.Fatal("compaction exposed executable history")
		}
		if len(req.Messages) != 1 || len(req.compactHistory) != 2 || req.compactHistory[1].Parts[0].Text != "KEEP_RESULT" {
			t.Fatal("summary lost historical evidence")
		}
		return nil, nil, errors.New("compaction-start-probe")
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
	_, msg := rt.serveMessagesInner(tuiRuntimeAnthropicResponder{w, r, &Usage{}}, r, "Claude Code", "claude-opus-5-5", "test-owner", "outer", []byte(body), start)
	if !called || !strings.Contains(msg, "compaction-start-probe") {
		t.Fatalf("compaction rejected before isolated start: %s", msg)
	}
}
func TestClaudeTUIRuntimeOrdinarySummaryDoesNotBypassOrphanGuard(t *testing.T) {
	body := `{"model":"claude-opus-5-5","max_tokens":100,"system":"Summarize the results","messages":[{"role":"assistant","content":[{"type":"tool_use","id":"old","name":"Bash","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"old","content":"result"},{"type":"text","text":"compact this"}]}]}`
	rt := tuiRuntimeNewRuntime([]byte("test"))
	called := false
	start := func(context.Context, *tuiRuntimeRequest) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
		called = true
		return nil, nil, errors.New("unexpected")
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
	code, _ := rt.serveMessagesInner(tuiRuntimeAnthropicResponder{w, r, &Usage{}}, r, "Claude Code", "claude-opus-5-5", "test-owner", "outer", []byte(body), start)
	if called || code != 409 {
		t.Fatalf("ordinary orphan escaped guard: %d", code)
	}
}

func TestClaudeTUIRuntimeCompactionReactiveCLIPrompt(t *testing.T) {
	req := &tuiRuntimeRequest{Request: Request{System: "x-anthropic-billing-header: cc_version=2.1.295; cch=test;You are Claude Code.", Tools: []Tool{{Name: "Bash"}}, Messages: []Message{{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "old", Text: "result"}, {Kind: Text, Text: "CRITICAL: Respond with TEXT ONLY. Do NOT call any tools.\nThis conversation is being compacted. Your reply replaces everything above: the next context window opens with what you write inside <summary></summary> tags, and you carry on from there.\nREMINDER: Do NOT call any tools. Respond with plain text only — return a summary. Tool calls will be rejected and you will fail the task."}}}}}}
	if !tuiRuntimeClaudeDesktopCompaction(req) {
		t.Fatal("real reactive CLI compaction not recognized")
	}
	out := tuiRuntimeClaudeDesktopCompactionRequest(req)
	if out.System != tuiRuntimeClaudeCompactionSystem || len(out.Tools) > 0 || tuiRuntimeInteractiveToolContinuation(out) {
		t.Fatal("reactive compaction not isolated")
	}
	req.Messages[0].Parts[1].Text = "Please compact this source file"
	if tuiRuntimeClaudeDesktopCompaction(req) {
		t.Fatal("ordinary user request recognized as CLI compaction")
	}
}

func TestClaudeTUIRuntimeCompactionNativeTranscriptPreservesStructuredContent(t *testing.T) {
	t.Setenv("CLAUDE_CONFIG_DIR", t.TempDir())
	req := &tuiRuntimeRequest{Request: Request{System: tuiRuntimeCompactSystemTest, Model: "claude-opus-5-5", Messages: []Message{{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "old-call", Name: "Bash", Args: json.RawMessage(`{"command":"NEVER_EXECUTE"}`)}}}, {Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "old-call", Text: "KEEP_RESULT"}, {Kind: Image, MediaType: "image/png", Data: "BINARY_NOT_TEXT"}, {Kind: Text, Text: "CURRENT_SUMMARY_DIRECTIVE"}}}}}}
	side := tuiRuntimeClaudeDesktopCompactionRequest(req)
	id, path, err := tuiRuntimeWriteCompactionTranscript(t.TempDir(), "", side)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	rows := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(rows) != 4 {
		t.Fatalf("rows %d", len(rows))
	}
	var row map[string]any
	if err := json.Unmarshal([]byte(rows[2]), &row); err != nil {
		t.Fatal(err)
	}
	blocks := row["message"].(map[string]any)["content"].([]any)
	if blocks[0].(map[string]any)["type"] != "tool_result" || blocks[1].(map[string]any)["type"] != "image" || blocks[1].(map[string]any)["source"].(map[string]any)["data"] != "BINARY_NOT_TEXT" {
		t.Fatal("wire structure lost")
	}
	if len(id) != 36 || side.Messages[0].Parts[0].Text == req.Messages[1].Parts[2].Text {
		t.Fatal("not an isolated summary turn")
	}
	if strings.Contains(side.Messages[0].Parts[0].Text, "BINARY_NOT_TEXT") {
		t.Fatal("binary flattened into text")
	}
}
