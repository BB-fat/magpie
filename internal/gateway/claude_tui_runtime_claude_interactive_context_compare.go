package gateway

import (
	"bytes"
	"crypto/sha256"
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

// These are the instructions and definitions actually supplied to the native
// run. Billing metadata is already omitted when rendering those instructions.
func tuiRuntimeInteractiveEquivalentContextKey(req *tuiRuntimeRequest) string {
	normalized := *req
	normalized.System = tuiRuntimeWithoutBillingHeader(req.System)
	if req.Format != nil {
		normalizedFormat := *req.Format
		normalizedFormat.Schema = tuiRuntimeInteractiveCanonicalJSON(req.Format.schema())
		normalized.Format = &normalizedFormat
	}
	normalized.Tools = slices.Clone(req.Tools)
	for i := range normalized.Tools {
		normalized.Tools[i].Schema = tuiRuntimeInteractiveCanonicalJSON(normalized.Tools[i].Schema)
	}
	slices.SortFunc(normalized.Tools, func(a, b Tool) int { return strings.Compare(a.Name, b.Name) })
	return tuiRuntimeTurnKey("", &normalized, nil)
}

func (r *tuiRuntimeSubscriptionRun) interactiveEquivalentContext(req *tuiRuntimeRequest) bool {
	r.mu.Lock()
	snapshot := r.interactiveContextSnapshot
	r.mu.Unlock()
	return snapshot != nil && snapshot.equivalentKey == tuiRuntimeInteractiveEquivalentContextKey(req) &&
		snapshot.effort == sha256.Sum256([]byte(req.Effort))
}

func tuiRuntimeInteractiveSameSafeguards(a, b json.RawMessage) bool {
	return bytes.Equal(tuiRuntimeInteractiveCanonicalJSON(a), tuiRuntimeInteractiveCanonicalJSON(b))
}
