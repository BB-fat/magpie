package gateway

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestClaudeTUIRuntimeInteractiveContextMismatchDiagnosticClassifiesWithoutLeaking(t *testing.T) {
	base := &tuiRuntimeRequest{Request: Request{System: "SECRET_BASE_PROMPT", Model: "claude-opus-5-5", Effort: "high",
		ToolChoice: "auto", Tools: []Tool{
			{Name: "private-a", Description: "SECRET_DESCRIPTION_A", Schema: json.RawMessage(`{"type":"object"}`)},
			{Name: "private-b", Description: "SECRET_DESCRIPTION_B", Schema: json.RawMessage(`{"type":"string"}`)},
		}}}
	old := tuiRuntimeInteractiveContextDiagnosticFromRequest(base)
	run := &tuiRuntimeSubscriptionRun{interactiveContextSnapshot: &old}
	check := func(req *tuiRuntimeRequest, want ...string) {
		t.Helper()
		got := run.interactiveContextMismatchDiagnostic(req)
		for _, s := range want {
			if !strings.Contains(got, s) {
				t.Fatalf("missing %s: %s", s, got)
			}
		}
		if strings.Contains(got, "SECRET") || strings.Contains(got, "private") || len(got) > 310 {
			t.Fatalf("sensitive or unbounded diagnostic: %s", got)
		}
	}
	check(base, "system_changed=false", "tools_changed=false", "effort_changed=false", "effort_mode_changed=false")
	for _, effort := range []string{"max", "low", ""} {
		changed := *base
		changed.Effort = effort
		mode := "effort_mode_changed=false"
		if effort == "" {
			mode = "effort_mode_changed=true"
		}
		check(&changed, "effort_changed=true", mode)
		diagnostic := run.interactiveContextMismatchDiagnostic(&changed)
		if strings.Contains(diagnostic, "high") || strings.Contains(diagnostic, "max") || strings.Contains(diagnostic, "low") {
			t.Fatalf("effort value leaked: %s", diagnostic)
		}
	}
	s := *base
	s.System = "SECRET_CHANGED_PROMPT"
	check(&s, "system_changed=true", "tools_changed=false")
	s = *base
	s.Tools = []Tool{base.Tools[1], base.Tools[0]}
	check(&s, "tools_changed=true", "tool_order_only=true")
	s = *base
	s.Tools = append([]Tool(nil), base.Tools...)
	s.Tools[0].Schema = json.RawMessage(`{"type":"array"}`)
	check(&s, "tools_changed=true", "tool_order_only=false")
	s = *base
	s.ToolChoice = "required"
	s.WebSearch = true
	check(&s, "tool_choice_changed=true", "web_search_changed=true")
	check(nil, "snapshot=unavailable")
}
