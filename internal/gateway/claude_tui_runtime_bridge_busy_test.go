package gateway

import (
	"sync"
	"testing"
	"time"
)

func TestClaudeTUIRuntimeBridgeParkedConsumerCanDetachAndResume(t *testing.T) {
	b := newTUIBridge()
	defer tuiRuntimeBridges.Delete(b.rt.bridge)
	run := &tuiRuntimeSubscriptionRun{bridge: b.rt.bridge, token: "consumer", interactive: true,
		parkedAt: time.Now(), shown: []tuiRuntimeShownCall{{id: "tool-1", name: "echo"}}}
	b.rt.bridge.runs[run.token] = run
	defer run.abort()
	if got := b.Parked(); got != 1 {
		t.Fatalf("before continuation: %d parked, want 1", got)
	}
	var readers sync.WaitGroup
	stop := make(chan struct{})
	readers.Add(1)
	go func() {
		defer readers.Done()
		for {
			select {
			case <-stop:
				return
			default:
				b.Parked()
			}
		}
	}()
	defer func() { close(stop); readers.Wait() }()
	for i := 0; i < 5; i++ {
		if !run.claimResume() {
			t.Fatalf("continuation %d cannot claim the detached run", i)
		}
		run.mu.Lock()
		run.httpConsumers++
		run.mu.Unlock()
		events := run.attach()
		if got := b.Parked(); got != 0 {
			t.Fatalf("active continuation: %d parked, want 0", got)
		}
		run.letGoFor(events)
		if got := b.Parked(); got != 0 {
			t.Fatalf("detached HTTP still unwinding: %d parked, want 0", got)
		}
		run.releaseHTTP()
		if got := b.Parked(); got != 1 {
			t.Fatalf("after consumer detached: %d parked, want 1", got)
		}
	}
	run.finish()
	if got := b.Parked(); got != 0 {
		t.Fatalf("after run ended: %d parked, want 0", got)
	}
}
