package gateway

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"
	"strings"
)

// Preserve JSON numbers and array order; only object-key order and whitespace
// are irrelevant. Invalid JSON remains unequal to a valid representation.
func tuiRuntimeInteractiveCanonicalJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if decoder.Decode(&value) != nil {
		return raw
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return raw
	}
	return canonical
}

// Hash exactly the external instructions rendered into the native prompt,
// or a lean run's system prompt, set apart from any prompt's text: a
// session holds its instructions in one place or the other. tool_choice is
// a per-turn constraint and settings travel with each launch.
func tuiRuntimeInstructionKey(req *tuiRuntimeRequest) string {
	text := tuiRuntimeRenderInteractiveExternalInstructions(req)
	if req.leanPrompt {
		text = "\x00lean-prompt\x00" + tuiRuntimeLeanSystem(req)
	}
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func tuiRuntimeInteractiveCanonicalTools(tools []Tool) []Tool {
	normalized := slices.Clone(tools)
	for i := range normalized {
		normalized[i].Schema = tuiRuntimeInteractiveCanonicalJSON(normalized[i].Schema)
	}
	slices.SortFunc(normalized, func(a, b Tool) int { return strings.Compare(a.Name, b.Name) })
	return normalized
}

// keepRunMode gives req the lean mode run was started in, when the request
// goes on in that run: its system prompt can't change while it lives, and
// the setting changing under a tool turn is not the caller's context
// changing. The turn's checkpoint and keys then keep the run's mode, and
// the conversation's next turn is rebuilt in the setting's.
func tuiRuntimeKeepRunMode(req *tuiRuntimeRequest, run *tuiRuntimeSubscriptionRun) {
	if run == nil {
		return
	}
	run.mu.Lock()
	snapshot := run.interactiveContextSnapshot
	run.mu.Unlock()
	if snapshot != nil {
		req.leanPrompt = snapshot.lean
	}
}

func (r *tuiRuntimeSubscriptionRun) interactiveEquivalentContext(req *tuiRuntimeRequest) bool {
	r.mu.Lock()
	snapshot := r.interactiveContextSnapshot
	r.mu.Unlock()
	if snapshot == nil {
		return false
	}
	// the setting is not the request's context: a run keeps the mode it
	// was started in (tuiRuntimeKeepRunMode)
	if req.leanPrompt != snapshot.lean {
		copied := *req
		copied.leanPrompt = snapshot.lean
		req = &copied
	}
	now := tuiRuntimeInteractiveContextDiagnosticFromRequest(req)
	return snapshot.model == now.model && snapshot.instructions == now.instructions &&
		snapshot.canonicalTools == now.canonicalTools && snapshot.canonicalSchema == now.canonicalSchema &&
		snapshot.toolChoice == now.toolChoice && snapshot.thinkOff == now.thinkOff &&
		snapshot.webSearch == now.webSearch && snapshot.effort == now.effort
}

func tuiRuntimeInteractiveSameSafeguards(a, b json.RawMessage) bool {
	return bytes.Equal(tuiRuntimeInteractiveCanonicalJSON(a), tuiRuntimeInteractiveCanonicalJSON(b))
}
