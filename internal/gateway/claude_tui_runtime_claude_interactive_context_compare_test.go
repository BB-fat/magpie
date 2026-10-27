package gateway

import (
	"encoding/json"
	"testing"
)

func TestClaudeTUIRuntimeInteractiveEquivalentContextPreservesEffectiveSettings(t *testing.T) {
	base := &tuiRuntimeRequest{Request: Request{Model: "claude-opus-5-5", Effort: "high", System: "x-anthropic-billing-header: cc_version=1; cch=old;\nInstructions", Tools: []Tool{
		{Name: "a", Description: "original", Schema: json.RawMessage(`{"type":"object","properties":{"n":{"minimum":9007199254740993}}}`)},
		{Name: "b", Schema: json.RawMessage(`{"type":"string"}`)},
	}}}
	snapshot := tuiRuntimeInteractiveContextDiagnosticFromRequest(base)
	run := &tuiRuntimeSubscriptionRun{interactiveContextSnapshot: &snapshot}
	equivalent := *base
	equivalent.System = "x-anthropic-billing-header: cc_version=2; cch=new;\nInstructions"
	equivalent.Tools = []Tool{base.Tools[1], base.Tools[0]}
	equivalent.Tools[1].Schema = json.RawMessage(`{ "properties": {"n": {"minimum": 9007199254740993}}, "type": "object" }`)
	if !run.interactiveEquivalentContext(&equivalent) {
		t.Fatal("equivalent request refused")
	}
	for _, mutate := range []func(*tuiRuntimeRequest){
		func(r *tuiRuntimeRequest) { r.System = "changed instructions" },
		func(r *tuiRuntimeRequest) { r.Model = "claude-sonnet-5-5" },
		func(r *tuiRuntimeRequest) {
			r.Format = &Format{Type: "json_schema", Schema: json.RawMessage(`{"type":"string"}`)}
		},
		func(r *tuiRuntimeRequest) { r.Effort = "max" },
		func(r *tuiRuntimeRequest) { r.ToolChoice = "required" },
		func(r *tuiRuntimeRequest) { r.WebSearch = true },
		func(r *tuiRuntimeRequest) { r.ThinkOff = true },
		func(r *tuiRuntimeRequest) { r.Tools = []Tool{base.Tools[0]} },
		func(r *tuiRuntimeRequest) {
			r.Tools = append([]Tool(nil), base.Tools...)
			r.Tools[0].Description = "changed"
		},
		func(r *tuiRuntimeRequest) {
			r.Tools = append([]Tool(nil), base.Tools...)
			r.Tools[0].Schema = json.RawMessage(`{"type":"object","properties":{"n":{"minimum":9007199254740992}}}`)
		},
	} {
		changed := equivalent
		mutate(&changed)
		if run.interactiveEquivalentContext(&changed) {
			t.Fatal("effective configuration change accepted")
		}
	}
}

func TestClaudeTUIRuntimeInteractiveSafeguardsIgnoreRepresentationOnly(t *testing.T) {
	if !tuiRuntimeInteractiveSameSafeguards(json.RawMessage(`{"a":1,"b":[2,3]}`), json.RawMessage(`{ "b": [2,3], "a":1 }`)) {
		t.Fatal("equivalent safeguards differ")
	}
	if tuiRuntimeInteractiveSameSafeguards(json.RawMessage(`{"a":1,"b":[2,3]}`), json.RawMessage(`{"a":1,"b":[3,2]}`)) {
		t.Fatal("policy array order ignored")
	}
}
