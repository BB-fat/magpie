package gateway

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestClaudeTUIRuntimeClosedStartupCannotLaunch(t *testing.T) {
	tuiRuntimeScWrapper(t)
	tuiRuntimeWithConfig(t, &tuiRuntimeConfig{StateDir: t.TempDir()})
	_, b := tuiRuntimeScRuntime(t)
	atRoom, permit := make(chan struct{}), make(chan struct{})
	var freed atomic.Int32
	tuiRuntimeWithConfig(t, &tuiRuntimeConfig{Room: func(context.Context, string) (func(bool), error) {
		close(atRoom)
		<-permit
		return func(bool) { freed.Add(1) }, nil
	}})
	req, err := tuiRuntimeParseRequest([]byte(`{"model":"claude-opus-5-5","max_tokens":16,"messages":[{"role":"user","content":"hello"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	type result struct {
		run *tuiRuntimeSubscriptionRun
		err error
	}
	done := make(chan result, 1)
	go func() {
		run, _, err := b.startInteractiveOneTurn(context.Background(), req, req.Model, "", tuiRuntimeScOwner, tuiRuntimeScOuter)
		done <- result{run, err}
	}()
	select {
	case <-atRoom:
	case <-time.After(10 * time.Second):
		t.Fatal("startup never published")
	}
	b.abortAll()
	close(permit)
	var got result
	select {
	case got = <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("startup did not finish")
	}
	if got.run == nil {
		if got.err == nil || freed.Load() != 1 {
			t.Fatalf("cancelled startup: err=%v room releases=%d", got.err, freed.Load())
		}
		return
	}
	defer got.run.abort()
	got.run.mu.Lock()
	child, closed, nativeDone := got.run.tree != nil, got.run.closed, got.run.interactiveDone
	got.run.mu.Unlock()
	if nativeDone != nil {
		select {
		case <-nativeDone:
		case <-time.After(10 * time.Second):
			got.run.abort()
			t.Fatal("child not reaped")
		}
	}
	b.mu.Lock()
	tracked := b.runs[got.run.token] != nil
	b.mu.Unlock()
	if got.err == nil && child && closed && !tracked {
		t.Fatal("NoStartAfterFinish violated: a closed, removed run subsequently launched its native wrapper")
	}
}

func TestClaudeTUIRuntimeNativeStartFenceWaitsForPublication(t *testing.T) {
	r := &tuiRuntimeSubscriptionRun{}
	if !r.beginNativeStart() {
		t.Fatal("fresh start was not admitted")
	}
	fenced := make(chan struct{})
	go func() { r.fenceNativeStart(); close(fenced) }()
	waitFor(t, "teardown fences the admitted launch", func() bool {
		r.mu.Lock()
		defer r.mu.Unlock()
		return r.stopNativeStart
	})
	select {
	case <-fenced:
		t.Error("teardown passed an unresolved native start")
	default:
	}
	tree, done := &tuiRuntimeProcTree{}, make(chan struct{})
	r.resolveNativeStart(tree, done)
	<-fenced
	if r.tree != tree || r.interactiveDone != done {
		t.Fatal("teardown cannot observe the native tree/completion it must reap")
	}
	if r.beginNativeStart() {
		t.Fatal("a fenced run admitted another native start")
	}
}
