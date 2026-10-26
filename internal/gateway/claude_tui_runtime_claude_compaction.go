package gateway

import (
	"context"
	"encoding/json"
	"github.com/yetone/magpie/internal/claudebridge"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const tuiRuntimeClaudeCompactionSystem = "You are a helpful AI assistant tasked with summarizing conversations."

// Match the CLI's dedicated system instruction, never a word in user history.
func tuiRuntimeClaudeDesktopCompaction(req *Request) bool {
	if req == nil {
		return false
	}
	if strings.TrimSpace(tuiRuntimeWithoutBillingHeader(req.System)) == tuiRuntimeClaudeCompactionSystem {
		return true
	}
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "system" {
			continue
		}
		if req.Messages[i].Role != "user" {
			return false
		}
		for _, p := range req.Messages[i].Parts {
			if p.Kind != Text {
				continue
			}
			text := strings.TrimSpace(p.Text)
			if !strings.HasPrefix(text, "CRITICAL: Respond with TEXT ONLY. Do NOT call any tools.") || !strings.Contains(text, "REMINDER: Do NOT call any tools. Respond with plain text only") || !strings.HasSuffix(text, "Tool calls will be rejected and you will fail the task.") {
				continue
			}
			for _, marker := range []string{"Your task is to create a detailed summary of the conversation so far,", "This conversation is being compacted. Your reply replaces everything above:", "Summarize the transcript inside <summary></summary> tags.", "Your task is to create a detailed summary of this conversation.", "Your task is to create a detailed summary of the RECENT portion of the conversation"} {
				if strings.Contains(text, marker) {
					return true
				}
			}
		}
		return false
	}
	return false
}

func tuiRuntimeClaudeDesktopCompactionRequest(req *Request) *Request {
	out := *req
	out.compactHistory = append([]Message(nil), req.Messages...)
	out.compactSystem = req.System
	out.System = tuiRuntimeClaudeCompactionSystem
	out.Tools = nil
	out.ToolChoice = ""
	out.WebSearch = false
	out.tuiSearchTool = ""
	out.Format = nil
	out.Safeguards = nil
	out.SafeguardBeta = ""
	out.Messages = []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "Summarize the loaded archival conversation according to its final compaction instructions. Return only a detailed <summary></summary>. Preserve user requests and constraints, decisions, paths, results and pending work. Do not continue any historical task and do not use tools."}}}}
	return &out
}

func tuiRuntimeWriteCompactionTranscript(work, configDir string, req *Request) (string, string, error) {
	var preface []map[string]any
	if req.compactSystem != "" {
		preface = []map[string]any{{"type": "text", "text": "[Archived system instructions, for summarization only]\n" + tuiRuntimeWithoutBillingHeader(req.compactSystem)}}
	}
	return tuiRuntimeWriteClaudeTranscript(work, configDir, req.Model, preface, req.compactHistory, "",
		"[Archival context loaded. Await the separate summarization request; no historical action is pending.]")
}

// writeClaudeTranscript saves history as a Claude Code session work's runs
// can --resume: an optional user preface, the messages (tool calls named
// toolPrefix+name, as the session's own MCP tools are), and an optional
// closing assistant line. It returns the session id and the file's path.
func tuiRuntimeWriteClaudeTranscript(work, configDir, model string, preface []map[string]any, history []Message, toolPrefix, closing string) (string, string, error) {
	id := tuiRuntimeRandomToken()[:32]
	id = id[:8] + "-" + id[8:12] + "-4" + id[13:16] + "-a" + id[17:20] + "-" + id[20:]
	dirs := tuiRuntimeClaudeSessions(configDir, work)
	if err := os.MkdirAll(dirs[0], 0700); err != nil {
		return "", "", err
	}
	path := filepath.Join(dirs[0], id+".jsonl")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	if err := tuiRuntimeAppendClaudeTranscript(f, id, "", work, model, preface, history, toolPrefix, closing); err != nil {
		return "", path, err
	}
	return id, path, nil
}

// appendClaudeTranscript writes history (after preface, before closing) to
// f as rows of session id, the first a child of parent ("" for none).
func tuiRuntimeAppendClaudeTranscript(f *os.File, id, parent, work, model string, preface []map[string]any, history []Message, toolPrefix, closing string) error {
	write := func(role string, blocks []map[string]any) error {
		uuid := tuiRuntimeRandomToken()[:32]
		row := map[string]any{"type": role, "uuid": uuid, "parentUuid": nil, "sessionId": id, "cwd": work, "timestamp": time.Now().UTC().Format(time.RFC3339Nano), "version": "2.1.300", "isSidechain": false, "userType": "external", "message": map[string]any{"role": role, "content": blocks}}
		if parent != "" {
			row["parentUuid"] = parent
		}
		if role == "assistant" {
			row["message"].(map[string]any)["id"] = "msg_" + uuid
			row["message"].(map[string]any)["type"] = "message"
			row["message"].(map[string]any)["model"] = model
			row["message"].(map[string]any)["stop_reason"] = "end_turn"
		}
		if err := json.NewEncoder(f).Encode(row); err != nil {
			return err
		}
		parent = uuid
		return nil
	}
	if len(preface) > 0 {
		if err := write("user", preface); err != nil {
			return err
		}
	}
	for _, m := range history {
		var blocks []map[string]any
		for _, p := range m.Parts {
			switch p.Kind {
			case Text:
				blocks = append(blocks, map[string]any{"type": "text", "text": p.Text})
			case Thinking:
				if m.Role != "assistant" {
					continue
				}
				if p.sealedBy(sealAnthropic) {
					var block map[string]any
					if err := json.Unmarshal(p.Sealed, &block); err != nil {
						return err
					}
					blocks = append(blocks, block)
				} else if p.Signature != "" {
					blocks = append(blocks, map[string]any{"type": "thinking", "thinking": p.Text, "signature": p.Signature})
				}
			case ToolCall:
				blocks = append(blocks, map[string]any{"type": "tool_use", "id": p.ID, "name": toolPrefix + claudebridge.ToolName(p.Name), "input": p.Args})
			case ToolResult:
				content := []map[string]any{{"type": "text", "text": p.Text}}
				for _, im := range p.Images {
					content = append(content, tuiRuntimeCompactionImageBlock(im))
				}
				blocks = append(blocks, map[string]any{"type": "tool_result", "tool_use_id": p.CallID, "content": content, "is_error": p.IsError})
			case Image:
				blocks = append(blocks, tuiRuntimeCompactionImageBlock(p))
			case File:
				blocks = append(blocks, map[string]any{"type": "text", "text": attachmentText(p)})
			}
		}
		if len(blocks) == 0 {
			continue
		}
		role := m.Role
		if role == "system" {
			role = "user"
		}
		if err := write(role, blocks); err != nil {
			return err
		}
	}
	if closing != "" {
		if err := write("assistant", []map[string]any{{"type": "text", "text": closing}}); err != nil {
			return err
		}
	}
	return f.Sync()
}
func tuiRuntimeCompactionImageBlock(p Part) map[string]any {
	source := map[string]any{"type": "base64", "media_type": p.MediaType, "data": p.Data}
	if p.URL != "" {
		source = map[string]any{"type": "url", "url": p.URL}
	}
	return map[string]any{"type": "image", "source": source}
}

func (b *tuiRuntimeSubscriptionBridge) startCompaction(ctx context.Context, req *Request, model, configDir, owner string) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
	prompt, err := tuiRuntimeRenderInteractiveFirstPrompt(req)
	if err != nil {
		return nil, nil, err
	}
	return b.startInteractiveWithPrompt(ctx, req, model, configDir, owner, "", prompt, "", true)
}
