package gateway

import "github.com/yetone/magpie/internal/provider"

const protoAnthropic = "anthropic"

type tuiRuntimeBridgeTool = bridgeTool
type tuiRuntimeMcpToolResult = mcpToolResult
type tuiRuntimeCliUsage = cliUsage
type tuiRuntimeClaudeLimit = provider.ClaudeLimit

func tuiRuntimeCanSearch() bool {
	if tuiRuntimeSettings.CanSearch != nil {
		return tuiRuntimeSettings.CanSearch()
	}
	return tuiRuntimeSettings.SearchHandler != nil
}
