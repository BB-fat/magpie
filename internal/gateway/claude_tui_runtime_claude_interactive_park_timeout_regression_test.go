package gateway

import (
	"testing"
	"time"
)

func TestClaudeTUIRuntimeInteractiveNativeToolWaitDoesNotExpireAfterFiveMinutes(t *testing.T) {
	// A native run waiting on its client's tools is let go after a short
	// wait, not held for an hour: a result that comes later resumes its
	// session (liteParkedBranch) instead of being refused.
	native := &tuiRuntimeSubscriptionRun{interactive: true, asked: []string{"Bash", "Write"}}
	if wait := native.parkedToolWaitTimeout(); wait <= 0 || wait > 10*time.Minute {
		t.Fatalf("native PTY run held %s waiting on its tools", wait)
	}
	if tuiRuntimeInteractiveActiveLongest < 6*time.Hour {
		t.Fatalf("native active-run watchdog too short: %s", tuiRuntimeInteractiveActiveLongest)
	}
	// The short path remains intact for disposable -p executions.
	oneShot := &tuiRuntimeSubscriptionRun{interactive: false}
	if wait := oneShot.parkedToolWaitTimeout(); wait != tuiRuntimeParkLongest {
		t.Fatalf("one-shot tool timeout unexpectedly changed: %s", wait)
	}
	if native.stdin != nil {
		t.Fatal("regression fixture must model actual interactive nil stdin")
	}
}
