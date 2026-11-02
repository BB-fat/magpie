package gateway

// Native launch and teardown must agree on whether a child can still appear.
// Room waits and process startup hold no run lock; teardown waits only for
// an already-admitted StartTree to publish its result, not for room.
func (r *tuiRuntimeSubscriptionRun) beginNativeStart() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed || r.stopNativeStart {
		return false
	}
	r.nativeStarting = make(chan struct{})
	return true
}

func (r *tuiRuntimeSubscriptionRun) resolveNativeStart(tree *tuiRuntimeProcTree, done chan struct{}) {
	r.mu.Lock()
	r.tree, r.interactiveDone = tree, done
	close(r.nativeStarting)
	r.mu.Unlock()
}

func (r *tuiRuntimeSubscriptionRun) fenceNativeStart() {
	r.mu.Lock()
	r.stopNativeStart = true
	starting := r.nativeStarting
	r.mu.Unlock()
	if starting != nil {
		<-starting
	}
}
