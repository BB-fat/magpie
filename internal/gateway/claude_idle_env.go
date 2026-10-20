package gateway

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	claudeIdleMaxEnv = "MAGPIE_CLAUDE_IDLE_MAX"
	claudeIdleTTLEnv = "MAGPIE_CLAUDE_IDLE_TTL"
	defaultIdleMost  = 6
)

// idleLongest and idleMost tune how many Claude Code processes the
// subscription bridge keeps between turns (keepIdle). Defaults match upstream
// when MAGPIE_CLAUDE_IDLE_MAX and MAGPIE_CLAUDE_IDLE_TTL are unset.
var (
	idleLongest     = time.Hour
	idleMost        = defaultIdleMost
	claudeIdleOnce  sync.Once
	claudeIdleError error
)

// LoadClaudeIdleConfig reads MAGPIE_CLAUDE_IDLE_* once per process.
func LoadClaudeIdleConfig() error {
	claudeIdleOnce.Do(func() {
		claudeIdleError = loadClaudeIdleConfig()
	})
	return claudeIdleError
}

// ResetClaudeIdleConfigForTest clears the one-time load for tests.
func ResetClaudeIdleConfigForTest() {
	claudeIdleOnce = sync.Once{}
	claudeIdleError = nil
	idleLongest = time.Hour
	idleMost = defaultIdleMost
}

func loadClaudeIdleConfig() error {
	if v := strings.TrimSpace(os.Getenv(claudeIdleMaxEnv)); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return fmt.Errorf("%s: want a non-negative integer, got %q", claudeIdleMaxEnv, v)
		}
		idleMost = n
	}
	if v := strings.TrimSpace(os.Getenv(claudeIdleTTLEnv)); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			return fmt.Errorf("%s: want a positive duration (e.g. 10m), got %q", claudeIdleTTLEnv, v)
		}
		idleLongest = d
	}
	return nil
}
