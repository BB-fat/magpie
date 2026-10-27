package gateway

// Only a live, exact-ID-owned MCP handoff may advertise additional tools
// between calls. Preserve the original tool schemas, order, context and
// effort; accepting arbitrary changes here could bind a result to a
// different execution context. A remapped or unmatched handoff is not proof.
func (r *tuiRuntimeSubscriptionRun) exactInteractiveToolGrowth(owner string, req *tuiRuntimeRequest, how string) bool {
	if r == nil || req == nil || how != tuiRuntimeByExactID {
		return false
	}
	r.mu.Lock()
	if !r.interactive || r.closed || r.readOnlyRecovery || r.owner != owner || r.interactiveContextSnapshot == nil {
		r.mu.Unlock()
		return false
	}
	known := make(map[string]bool, len(r.tools))
	for name, offered := range r.tools {
		if offered {
			known[name] = true
		}
	}
	r.mu.Unlock()
	old, added, seen := make([]Tool, 0, len(req.Tools)), 0, make(map[string]bool, len(req.Tools))
	for _, tool := range req.Tools {
		if seen[tool.Name] {
			return false
		}
		seen[tool.Name] = true
		if known[tool.Name] {
			old = append(old, tool)
		} else {
			added++
		}
	}
	if added == 0 {
		return false
	}
	prior := *req
	prior.Tools = old
	return r.interactiveEquivalentContext(&prior)
}
