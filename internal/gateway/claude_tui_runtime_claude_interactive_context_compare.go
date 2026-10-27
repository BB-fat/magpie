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

// Hash exactly the external instructions rendered into the native prompt.
// tool_choice is a per-turn constraint and settings travel with each launch.
func tuiRuntimeInstructionKey(req *tuiRuntimeRequest) string {
	sum := sha256.Sum256([]byte(tuiRuntimeRenderInteractiveExternalInstructions(req)))
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

func (r *tuiRuntimeSubscriptionRun) interactiveEquivalentContext(req *tuiRuntimeRequest) bool {
	r.mu.Lock()
	snapshot := r.interactiveContextSnapshot
	r.mu.Unlock()
	if snapshot == nil {
		return false
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
