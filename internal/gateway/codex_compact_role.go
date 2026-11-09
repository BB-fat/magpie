package gateway

import (
	"strings"

	"github.com/tidwall/gjson"
)

func compactDeveloperRefused(model string) string { return model + ":compact-developer" }

// The chat-role validator behind a Responses relay (#2955). Only its
// explicit rejection of developer triggers a retry, not an auth failure
// or a refusal of system by a Codex relay.
func compactDeveloperRoleRefusal(body []byte) bool {
	msg := strings.ToLower(gjson.GetBytes(body, "error.message").String())
	return strings.Contains(msg, "developer is not one of ['system', 'assistant', 'user', 'tool', 'function']")
}
