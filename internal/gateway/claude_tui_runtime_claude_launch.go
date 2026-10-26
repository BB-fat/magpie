package gateway

import (
	"fmt"
	"os"
	"strings"
)

// tuiRuntimeInteractiveSafeguardInlineMax stays under Linux MAX_ARG_STRLEN (~128 KiB).
const tuiRuntimeInteractiveSafeguardInlineMax = 120 << 10

// tuiRuntimeClaudeNodeLauncher reports whether claudePath is a Node launcher script.
func tuiRuntimeClaudeNodeLauncher(claudePath string) bool {
	b, err := os.ReadFile(claudePath)
	if err != nil || len(b) < 2 || b[0] != '#' || b[1] != '!' {
		return false
	}
	if len(b) > 16<<10 {
		b = b[:16<<10]
	}
	line, _, _ := strings.Cut(string(b), "\n")
	return strings.Contains(line, "node")
}

func tuiRuntimeValidateInteractiveSafeguardLaunch(claudePath string, safeguardBodyLen int) error {
	if tuiRuntimeClaudeNodeLauncher(claudePath) {
		return nil
	}
	if safeguardBodyLen > tuiRuntimeInteractiveSafeguardInlineMax {
		return fmt.Errorf(
			"Claude interactive safeguards are too large for native Claude Code (%d bytes; inline limit %d). Shorten classifier context or use a Node-based Claude install",
			safeguardBodyLen, tuiRuntimeInteractiveSafeguardInlineMax)
	}
	return nil
}
