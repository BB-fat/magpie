package gateway

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"
)

// These summaries locate a mismatch without logging prompts, tool arguments,
// results, IDs or even their hashes. They never participate in acceptance.
type tuiRuntimeHistoryMessageFingerprint struct {
	role  string
	shape string // kinds only, never prompt text, tool IDs, names or hashes
	words [32]byte
}

func tuiRuntimeHistoryPartShape(m Message) string {
	var kinds []string
	for _, p := range m.Parts {
		switch p.Kind {
		case Text:
			kinds = append(kinds, "text")
		case ToolCall:
			kinds = append(kinds, "call")
		case ToolResult:
			kinds = append(kinds, "result")
		case Thinking:
			kinds = append(kinds, "thinking")
		case Image:
			kinds = append(kinds, "image")
		case File:
			kinds = append(kinds, "file")
		default:
			kinds = append(kinds, "other")
		}
	}
	return strings.Join(kinds, ",")
}

// Bound the suffix summary and emit only structural counts. This explains why
// an explicit recovery turn was not recognized, without recording its text.
func tuiRuntimeInteractiveRecoveryDiagnostic(req *tuiRuntimeRequest) string {
	var rows []string
	for i := max(0, len(req.Messages)-6); i < len(req.Messages); i++ {
		m := req.Messages[i]
		role := m.Role
		switch role {
		case "user", "assistant", "system":
		default:
			role = "other"
		}
		text, calls, results, reminders := 0, 0, 0, 0
		for _, p := range m.Parts {
			switch p.Kind {
			case Text:
				text++
				if strings.Contains(p.Text, "<system-reminder>") {
					reminders++
				}
			case ToolCall:
				calls++
			case ToolResult:
				results++
			}
		}
		rows = append(rows, fmt.Sprintf("row=%d role=%s text=%d calls=%d results=%d reminders=%d recovery_user=%t synthetic_error=%t transport_502=%t", i, role, text, calls, results, reminders, tuiRuntimeInteractiveRecoveryUserMessage(m), tuiRuntimeInteractiveRecoverySyntheticAssistant(m), tuiRuntimeClaudeDesktopFailedTurnCompanionReply(m)))
	}
	return strings.Join(rows, "; ")
}

// Diagnostic only: a transport_502 match describes a known text format,
// never authenticated provenance or permission to discard/replay a turn.
// Reuse the admission classifiers; do not broaden their acceptance rules.
func tuiRuntimeInteractiveRecoveryAdmissionDiagnostic(req *tuiRuntimeRequest) string {
	clean, _, err := tuiRuntimeInteractiveRequestWithoutSystem(req)
	if err != nil {
		return "optin_block=invalid_system_shape"
	}
	end := len(clean.Messages) - 1
	for end >= 0 && tuiRuntimeInteractiveIgnorableRecoveryTail(clean.Messages[end]) {
		end--
	}
	optin := tuiRuntimeInteractiveReadOnlyOptInIndex(req) >= 0
	legacy := tuiRuntimeInteractiveAbandonedToolRecoveryIndex(req) >= 0
	block, count, last, sentinel := "last_not_user", 0, -1, -1
	if end >= 0 && clean.Messages[end].Role == "user" {
		m := clean.Messages[end]
		shapeRejected := tuiRuntimeClaudeDesktopTurnCompanionMessage(m)
		for i, p := range m.Parts {
			if p.Kind == Thinking || p.Kind == ToolResult || p.Kind == Image {
				continue
			}
			if p.Kind != Text {
				shapeRejected = true
				continue
			}
			s := strings.TrimSpace(p.Text)
			if s == "" || strings.HasPrefix(s, "[Request interrupted by user]") {
				continue
			}
			if strings.Contains(s, "<system-reminder>") || strings.HasPrefix(s, "[SUGGESTION MODE:") {
				shapeRejected = true
			}
			last = i
			if tuiRuntimeInteractiveOptInRecoveryTexts[s] {
				count++
				sentinel = i
			}
		}
		switch {
		case optin:
			block = "none"
		case count == 0:
			block = "missing_sentinel"
		case shapeRejected:
			block = "shape_rejected"
		case count > 1:
			block = "multiple_sentinels"
		case sentinel != last:
			block = "sentinel_not_last"
		default:
			block = "classifier_rejected"
		}
	}
	return fmt.Sprintf("optin=%t legacy=%t tool_handoff=%t sentinel_parts=%d sentinel_last=%t optin_block=%s",
		optin, legacy, tuiRuntimeInteractiveToolContinuation(req), count, count > 0 && sentinel == last, block)
}

// Explain a failed persisted resume through bounded structural metadata
// only. Never emit the checkpoint hash, native/outer session IDs, prompt,
// tool names/IDs, args, results or file paths. Crucially, do not rehash the
// whole historical transcript: the actual verifier has already done that,
// and a second pass on a 500k context can allocate tens of MiB during a
// 502 retry storm on a 1 GiB VPS. Diagnostic work never authorizes a run.
func tuiRuntimeInteractiveRestoreRejectionDiagnostic(saved tuiRuntimeInteractiveSessionEntry, req *tuiRuntimeRequest) string {
	if req == nil {
		return "history=missing"
	}
	replies, replyIndex := 0, -1
	if saved.ReplyKey != "" {
		for i, m := range req.Messages {
			if m.Role == "assistant" && tuiRuntimeAssistantReplyKey(m) == saved.ReplyKey {
				replies++
				replyIndex = i
			}
		}
	}
	anchor := "none"
	switch {
	case replies > 1:
		anchor = "ambiguous"
	case replies == 1:
		anchor = "unique"
	}
	// An exact cumulative-key match may exist even when replyKey is absent.
	// Leave that unknown instead of guessing or rereading the full history.
	match := -1
	if replies == 1 {
		match = replyIndex
	}
	suffix := "no_checkpoint"
	if match >= 0 {
		suffix = "empty"
		if match < len(req.Messages)-1 {
			suffix = "unrecognized_suffix"
			for _, m := range req.Messages[match+1:] {
				for _, p := range m.Parts {
					if p.Kind == ToolResult {
						suffix = "unmatched_tool_result"
						break
					}
				}
				if suffix == "unmatched_tool_result" {
					break
				}
				if m.Role == "assistant" {
					suffix = "intervening_assistant"
				}
			}
		}
	}
	summary := fmt.Sprintf("checkpoint_saved=%t anchor=%s suffix=%s dirty=%t rows=%d; %s; %s",
		saved.ConvKey != "", anchor, suffix, saved.Dirty, len(req.Messages), tuiRuntimeInteractiveRecoveryAdmissionDiagnostic(req), tuiRuntimeInteractiveRecoveryDiagnostic(req))
	if len(summary) > 1024 {
		return summary[:1024]
	}
	return summary
}

func tuiRuntimeInteractiveHistoryFingerprints(msgs []Message) []tuiRuntimeHistoryMessageFingerprint {
	msgs = tuiRuntimeWithoutDynamicTokenBudgetNotices(msgs)
	words := tuiRuntimeMessageWords(msgs)
	out := make([]tuiRuntimeHistoryMessageFingerprint, 0, len(msgs))
	for i, word := range words {
		if word == "" {
			continue
		}
		role := msgs[i].Role
		switch role {
		case "user", "assistant", "system":
		default:
			role = "other"
		}
		out = append(out, tuiRuntimeHistoryMessageFingerprint{role: role, shape: tuiRuntimeHistoryPartShape(msgs[i]), words: sha256.Sum256([]byte(word))})
	}
	return out
}

func (r *tuiRuntimeSubscriptionRun) historyMismatchDiagnostic(msgs []Message) string {
	r.mu.Lock()
	old := slices.Clone(r.heardHistory)
	r.mu.Unlock()
	if old == nil {
		return "history_baseline_unavailable"
	}
	current := tuiRuntimeInteractiveHistoryFingerprints(msgs)
	for i, before := range old {
		if i < len(current) && before == current[i] {
			continue
		}
		afterRole, afterShape := "missing", "missing"
		if i < len(current) {
			afterRole = current[i].role
			afterShape = current[i].shape
		}
		return fmt.Sprintf("normalized_message=%d old_role=%s new_role=%s old_parts=%s new_parts=%s old_messages=%d current_messages=%d", i, before.role, afterRole, before.shape, afterShape, len(old), len(current))
	}
	return "normalized_history_unchanged"
}
