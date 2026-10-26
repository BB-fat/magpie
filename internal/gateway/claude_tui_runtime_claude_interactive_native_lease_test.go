//go:build darwin || linux

package gateway

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestClaudeTUIRuntimeNativeLeaseSurvivesRuntimeFDClosureWhileWrapperLives(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "xdg"))
	f, err := tuiRuntimeAcquireInteractiveNativeLease("owner", "session")
	if err != nil {
		t.Fatal(err)
	}
	if other, err := tuiRuntimeAcquireInteractiveNativeLease("owner", "session"); err == nil {
		tuiRuntimeReleaseInteractiveNativeLease(other)
		t.Fatal("second Runtime obtained an active native lease")
	} else if !errors.Is(err, tuiRuntimeErrInteractiveOuterLive) {
		t.Fatal(err)
	}
	// Runtime and wrapper inherit the SAME open file description. Closing
	// Runtime's descriptor must not release the wrapper's active lease.
	child := exec.Command("sh", "-c", "sleep 0.5")
	child.ExtraFiles = []*os.File{f}
	if err := child.Start(); err != nil {
		tuiRuntimeReleaseInteractiveNativeLease(f)
		t.Fatal(err)
	}
	tuiRuntimeReleaseInteractiveNativeLease(f)
	if other, err := tuiRuntimeAcquireInteractiveNativeLease("owner", "session"); err == nil {
		tuiRuntimeReleaseInteractiveNativeLease(other)
		_ = child.Process.Kill()
		_ = child.Wait()
		t.Fatal("original Runtime exiting unlocked a live native wrapper")
	} else if !errors.Is(err, tuiRuntimeErrInteractiveOuterLive) {
		t.Fatal(err)
	}
	if err := child.Wait(); err != nil {
		t.Fatal(err)
	}
	lease, err := tuiRuntimeAcquireInteractiveNativeLease("owner", "session")
	if err != nil {
		t.Fatalf("native lease not released after wrapper exited: %v", err)
	}
	tuiRuntimeReleaseInteractiveNativeLease(lease)
}

func TestClaudeTUIRuntimeNativeLeaseTwoBridgesCannotStartSameInteractiveOuter(t *testing.T) {
	dir := tuiRuntimeArDir(t)
	t.Setenv("MAGPIE_CLAUDE_DURABLE_TOOL_JOURNAL", "1")
	tuiRuntimeArWrapper(t, dir, "native-hold.sh", "#!/bin/sh\ncat >/dev/null\nsleep 1.2\n"+strings.TrimPrefix(tuiRuntimeArAnswerScript(), "#!/bin/sh\ncat >/dev/null\n"))
	b1, b2 := tuiRuntimeNewSubscriptionBridge(), tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b1)
	defer tuiRuntimeBridges.Delete(b2)
	req := &Request{Model: tuiRuntimeAutoTestModel, Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "check"}}}}}
	owner, outer := "owner", "same-outer"
	run, _, err := b1.startInteractiveWithPrompt(context.Background(), req, tuiRuntimeAutoTestModel, "", owner, outer, "test", "", false)
	if err != nil {
		t.Fatal(err)
	}
	defer run.abort()
	if another, _, err := b2.startInteractiveWithPrompt(context.Background(), req, tuiRuntimeAutoTestModel, "", owner, outer, "test", "", false); err == nil {
		another.abort()
		t.Fatal("second Runtime started a competing native session")
	} else if !errors.Is(err, tuiRuntimeErrInteractiveOuterLive) {
		t.Fatalf("unexpected second start: %v", err)
	}
	run.abort()
	lease, err := tuiRuntimeAcquireInteractiveNativeLease(owner, outer)
	if err != nil {
		t.Fatalf("native lease remained held after reaping: %v", err)
	}
	tuiRuntimeReleaseInteractiveNativeLease(lease)
}

func TestClaudeTUIRuntimeNativeLeaseProbeBlocksOneShotRecoveryClaim(t *testing.T) {
	b, rt, owner, outer, runLog := tuiRuntimeAutoRecoveryFixture(t)
	tuiRuntimeAutoRegister(t, b, owner, outer, "Write", `{"file_path":"note.txt"}`)
	lease, err := tuiRuntimeAcquireInteractiveNativeLease(owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	defer tuiRuntimeReleaseInteractiveNativeLease(lease)
	start := time.Now()
	code, _, _ := tuiRuntimeArServe(t, rt, b, owner, outer, tuiRuntimeAutoRecoveryBody("Write", `{"file_path":"note.txt"}`, `"wrote"`))
	if code != 409 {
		t.Fatalf("auto recovery took over an active external native process, status=%d", code)
	}
	if time.Since(start) > 5*time.Second {
		t.Fatal("active session detection unexpectedly hung")
	}
	rows, err := b.toolJournal.read()
	if err != nil || len(rows) != 1 {
		t.Fatalf("native lease refusal consumed one-shot recovery claim: %v %+v", err, rows)
	}
	if _, err := os.Stat(runLog); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("recovery launched an extra wrapper")
	}
}

func TestClaudeTUIRuntimeNativeLeaseRejectedResumeDoesNotDirtyOtherRuntimeCheckpoint(t *testing.T) {
	b, _, owner, outer, _ := tuiRuntimeAutoRecoveryFixture(t)
	// This was a clean completed checkpoint before an unrelated Runtime
	// attempted to resume while another native process owned the outer.
	tuiRuntimeArSave(t, b, owner, outer, "inner-clean", false)
	lease, err := tuiRuntimeAcquireInteractiveNativeLease(owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	defer tuiRuntimeReleaseInteractiveNativeLease(lease)
	req := &Request{Model: tuiRuntimeAutoTestModel, Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "continue"}}}}}
	saved, ok := b.interactiveSession(owner, outer)
	if !ok || saved.Dirty {
		t.Fatalf("clean checkpoint precondition violated: %+v", saved)
	}
	_, _, err = b.startInteractiveResumeTurn(context.Background(), req, tuiRuntimeAutoTestModel, "", owner, outer, saved, req.Messages, false)
	if !errors.Is(err, tuiRuntimeErrInteractiveOuterLive) {
		t.Fatalf("competing native session was allowed to resume: %v", err)
	}
	after, ok := b.interactiveSession(owner, outer)
	if !ok || after.Dirty || after.SessionID != saved.SessionID {
		t.Fatalf("rejected process dirtied owner's clean checkpoint: %+v", after)
	}
}

func TestClaudeTUIRuntimeNativeLeaseBlocksRetryAfterClaimWithoutConsumingIt(t *testing.T) {
	b, rt, owner, outer, runLog := tuiRuntimeAutoRecoveryFixture(t)
	t.Setenv("MAGPIE_CLAUDE_AUTO_RECOVER_RETRY", "1")
	tuiRuntimeAutoRegister(t, b, owner, outer, "Write", `{"file_path":"note.txt"}`)
	body := tuiRuntimeAutoRecoveryBody("Write", `{"file_path":"note.txt"}`, `"wrote"`)
	req, err := parseAnthropic(body)
	if err != nil {
		t.Fatal(err)
	}
	if err := b.toolJournal.claimOrphanRecovery(owner, outer, req); err != nil {
		t.Fatal(err)
	}
	lease, err := tuiRuntimeAcquireInteractiveNativeLease(owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	status, _, _ := tuiRuntimeArServe(t, rt, b, owner, outer, body)
	if status != 409 {
		tuiRuntimeReleaseInteractiveNativeLease(lease)
		t.Fatalf("claimed recovery hijacked an ongoing native wrapper: %d", status)
	}
	rows, err := b.toolJournal.read()
	if err != nil || len(rows) != 2 {
		t.Fatalf("native lease conflict consumed retry: %v %+v", err, rows)
	}
	if _, err := os.Stat(runLog); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("native retry process was launched while blocked")
	}
	tuiRuntimeReleaseInteractiveNativeLease(lease)
	status, msg, output := tuiRuntimeArServe(t, rt, b, owner, outer, body)
	if status != 200 || !strings.Contains(output, "OL_ANSWER") {
		t.Fatalf("released lease could not recover: %d %s %s", status, msg, output)
	}
	rows, err = b.toolJournal.read()
	if err != nil || len(rows) != 3 || rows[2].Stage != "auto_recovery_retry_claimed" {
		t.Fatalf("verified retry did not produce exact one-shot transition: %v %+v", err, rows)
	}
}
