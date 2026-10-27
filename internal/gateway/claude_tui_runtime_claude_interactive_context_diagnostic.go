package gateway

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
)

// Only fixed boolean categories are ever logged. Keep the raw system prompt,
// tool descriptions/schemas and hashes out of diagnostics. This snapshot is
// immutable once published. Canonical field fingerprints compare harmless
// representation changes without a catch-all configuration key.
type tuiRuntimeInteractiveContextDiagnosticSnapshot struct {
	system, tools, unorderedTools, toolChoice, schema [32]byte
	thinkOff, webSearch, effortSet                    bool
	effort                                            [32]byte
	model, instructions                               string
	canonicalTools, canonicalSchema                   [32]byte
}

func tuiRuntimeInteractiveContextDiagnosticFromRequest(req *tuiRuntimeRequest) tuiRuntimeInteractiveContextDiagnosticSnapshot {
	if req == nil {
		return tuiRuntimeInteractiveContextDiagnosticSnapshot{}
	}
	tools, _ := json.Marshal(req.Tools)
	perTool := make([]string, 0, len(req.Tools))
	for _, t := range req.Tools {
		encoded, _ := json.Marshal(t)
		perTool = append(perTool, string(encoded))
	}
	sort.Strings(perTool)
	unordered, _ := json.Marshal(perTool)
	canonicalTools, _ := json.Marshal(tuiRuntimeInteractiveCanonicalTools(req.Tools))
	return tuiRuntimeInteractiveContextDiagnosticSnapshot{
		model: req.Model, instructions: tuiRuntimeInstructionKey(req),
		canonicalTools:  sha256.Sum256(canonicalTools),
		canonicalSchema: sha256.Sum256(tuiRuntimeInteractiveCanonicalJSON(req.Format.schema())),
		effort:          sha256.Sum256([]byte(req.Effort)),
		system:          sha256.Sum256([]byte(req.System)),
		tools:           sha256.Sum256(tools),
		unorderedTools:  sha256.Sum256(unordered),
		toolChoice:      sha256.Sum256([]byte(req.ToolChoice)),
		schema:          sha256.Sum256(req.Format.schema()),
		thinkOff:        req.ThinkOff, webSearch: req.WebSearch, effortSet: req.Effort != "",
	}
}

func (r *tuiRuntimeSubscriptionRun) interactiveContextMismatchDiagnostic(req *tuiRuntimeRequest) string {
	if r == nil || req == nil {
		return "snapshot=unavailable"
	}
	r.mu.Lock()
	old := r.interactiveContextSnapshot
	r.mu.Unlock()
	if old == nil {
		return "snapshot=unavailable"
	}
	now := tuiRuntimeInteractiveContextDiagnosticFromRequest(req)
	toolChanged := old.tools != now.tools
	return fmt.Sprintf("system_changed=%t tools_changed=%t tool_order_only=%t tool_choice_changed=%t schema_changed=%t think_off_changed=%t web_search_changed=%t effort_mode_changed=%t effort_changed=%t",
		old.system != now.system, toolChanged, toolChanged && old.unorderedTools == now.unorderedTools,
		old.toolChoice != now.toolChoice, old.schema != now.schema, old.thinkOff != now.thinkOff,
		old.webSearch != now.webSearch, old.effortSet != now.effortSet, old.effort != now.effort)
}
