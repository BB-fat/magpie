package gateway

import (
	"runtime"
	"testing"
	"time"
)

func TestClaudeTUIRuntimeSessionReservationPrecedesWriterInspection(t *testing.T) {
	tuiRuntimeWithConfig(t, &tuiRuntimeConfig{StateDir: t.TempDir()})
	_, b := tuiRuntimeScRuntime(t)
	r := &tuiRuntimeSubscriptionRun{bridge: b, token: "live", interactive: true, sessionID: "native"}
	b.runs[r.token] = r
	// A writer being updated holds its lock. Claiming this session must
	// reserve it before waiting to inspect that writer, not cache an earlier
	// writer-list snapshot and acquire the reservation afterwards.
	r.mu.Lock()
	done := make(chan bool, 1)
	go func() {
		release, ok := b.claimSession("native")
		if ok {
			release()
		}
		done <- ok
	}()
	reserved := false
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		b.mu.Lock()
		reserved = b.claimedSessions["native"]
		b.mu.Unlock()
		if reserved {
			break
		}
		runtime.Gosched()
	}
	r.mu.Unlock()
	if ok := <-done; ok {
		t.Error("a live writer was reserved for another native start")
	}
	if !reserved {
		t.Fatal("native reservation was not held during live-writer inspection; a stale scan can authorize a second writer")
	}
	b.mu.Lock()
	left := b.claimedSessions["native"]
	b.mu.Unlock()
	if left {
		t.Fatal("a refused native reservation leaked")
	}
}

func TestClaudeTUIRuntimeSessionReservationReleaseIsSingleUse(t *testing.T) {
	_, b := tuiRuntimeScRuntime(t)
	first, ok := b.claimSession("native")
	if !ok {
		t.Fatal("first reservation refused")
	}
	if release, ok := b.claimSession("native"); ok {
		release()
		t.Fatal("concurrent reservation granted")
	}
	first()
	second, ok := b.claimSession("native")
	if !ok {
		t.Fatal("released reservation not reusable")
	}
	defer second()
	first()
	if release, ok := b.claimSession("native"); ok {
		release()
		t.Fatal("an old release removed a newer reservation")
	}
}
