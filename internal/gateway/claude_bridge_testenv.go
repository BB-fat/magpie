package gateway

import (
	"os"
	"testing"

	"github.com/yetone/magpie/internal/claudebridge"
)

// restoreDefaultClaudeBridgeEnvAfterTest registers a cleanup that puts
// MAGPIE_CLAUDE_* back to unset defaults and reloads MCP and idle config.
// Tests that call Reset* and Load* after t.Setenv leave sync.Once stuck on
// their values until this runs; without it later tests in the same shard
// see the wrong MCP prefix or idle pool size.
func restoreDefaultClaudeBridgeEnvAfterTest(t *testing.T) {
	t.Cleanup(restoreDefaultClaudeBridgeEnv)
}

func restoreDefaultClaudeBridgeEnv() {
	_ = os.Unsetenv(claudebridge.ClaudeMCPEnv)
	_ = os.Unsetenv(claudeIdleMaxEnv)
	_ = os.Unsetenv(claudeIdleTTLEnv)
	claudebridge.ResetMCPConfigForTest()
	ResetClaudeIdleConfigForTest()
	_ = claudebridge.LoadMCPConfig()
	_ = LoadClaudeIdleConfig()
}
