package claudetui

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"slices"
	"strings"
)

// Preserve JSON numbers and array order; only object-key order and whitespace
// are irrelevant. Invalid JSON remains unequal to a valid representation.
func interactiveCanonicalJSON(raw json.RawMessage) json.RawMessage {
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
func interactiveEquivalentContextKey(req *Request) string {
	normalized := *req
	normalized.System = withoutBillingHeader(req.System)
	normalized.Schema = interactiveCanonicalJSON(req.Schema)
	normalized.Tools = slices.Clone(req.Tools)
	for i := range normalized.Tools {
		normalized.Tools[i].Schema = interactiveCanonicalJSON(normalized.Tools[i].Schema)
	}
	slices.SortFunc(normalized.Tools, func(a, b Tool) int { return strings.Compare(a.Name, b.Name) })
	return turnKey("", &normalized, nil)
}

func (r *subscriptionRun) interactiveEquivalentContext(req *Request) bool {
	r.mu.Lock()
	snapshot := r.interactiveContextSnapshot
	r.mu.Unlock()
	return snapshot != nil && snapshot.equivalentKey == interactiveEquivalentContextKey(req) &&
		snapshot.effort == sha256.Sum256([]byte(req.Effort))
}

func interactiveSameSafeguards(a, b json.RawMessage) bool {
	return bytes.Equal(interactiveCanonicalJSON(a), interactiveCanonicalJSON(b))
}
