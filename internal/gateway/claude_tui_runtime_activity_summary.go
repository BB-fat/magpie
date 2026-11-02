package gateway

import (
	"regexp"
	"strings"
)

var tuiRuntimeActivityImageNotice = regexp.MustCompile(`^\[Image: original [0-9]+x[0-9]+, displayed at [0-9]+x[0-9]+\. Multiply coordinates by [0-9]+(?:\.[0-9]+)? to map to original image\.\]$`)

// Claude Code appends this literal instruction to a copy of its worker's
// history. The short reply is UI activity text, not the worker's next turn.
func tuiRuntimeActivitySummary(req *tuiRuntimeRequest) bool {
	if len(req.Messages) < 2 || req.ToolChoice != "" && req.ToolChoice != "auto" && req.ToolChoice != "none" {
		return false
	}
	last := req.Messages[len(req.Messages)-1]
	if last.Role != "user" {
		return false
	}
	for i := len(last.Parts) - 1; i >= 0; i-- {
		p := last.Parts[i]
		if p.Kind == Text && strings.TrimSpace(p.Text) != "" {
			text := p.Text
			// Claude Code may join its coordinate notices and the activity
			// instruction into one text block. Inspect without changing it.
			for strings.HasPrefix(text, "[Image: original ") {
				line, rest, ok := strings.Cut(text, "\n")
				if !ok || !tuiRuntimeActivityImageNotice.MatchString(line) {
					return false
				}
				text = rest
			}
			if !strings.HasPrefix(text, "Describe your most recent action in 3-5 words using present tense (-ing). Name the file or function, not the branch. Do not use tools.\n") {
				return false
			}
			for _, m := range req.Messages[:len(req.Messages)-1] {
				if m.Role == "assistant" {
					return true
				}
			}
			return false
		}
	}
	return false
}
