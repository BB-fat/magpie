//go:build darwin || linux

package claudetui

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// The helper is a distinct OS process, not a goroutine: SIGKILL cannot run
// deferred unlock / tmux cleanup. It never calls real Claude or Magpie.
func TestNativeHardKillProcessHelper(t *testing.T) {
	if os.Getenv("MAGPIE_HARDKILL_HELPER") != "1" {
		return
	}
	owner, outer := "hardkill-owner", "hardkill-outer"
	lease, err := acquireInteractiveNativeLease(owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	defer releaseInteractiveNativeLease(lease)
	name := "claude-bridge-deadbeef"
	writeTestNativeTmuxMarker(t, owner, outer, name)
	// The tmux command is wrapped with a dedicated -L socket namespace by
	// the parent. It starts one harmless local side effect then sleeps.
	cmd := exec.Command("tmux", "new-session", "-d", "-s", name, "-c", os.Getenv("MAGPIE_HARDKILL_WORK"), "sh -c 'echo executed >> side-effects.log; sleep 60'")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("isolated native launch failed: %v %s", err, out)
	}
	if err := os.WriteFile(os.Getenv("MAGPIE_HARDKILL_READY"), []byte("1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for {
		time.Sleep(time.Second)
	} // Parent SIGKILLs this exact process.
}

func TestNativeRuntimeHardKillProcessHelper(t *testing.T) {
	if os.Getenv("MAGPIE_RUNTIME_KILL_HELPER") != "1" {
		return
	}
	lease, err := acquireInteractiveNativeLease("hardkill-owner", "runtime-owned-outer")
	if err != nil {
		t.Fatal(err)
	}
	child := exec.Command("sleep", "40")
	child.ExtraFiles = []*os.File{lease}
	if err := child.Start(); err != nil {
		releaseInteractiveNativeLease(lease)
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv("MAGPIE_RUNTIME_CHILD_PID"), []byte(strconv.Itoa(child.Process.Pid)), 0o600); err != nil {
		t.Fatal(err)
	}
	// No cleanup: parent kills this stand-in Runtime with SIGKILL. The
	// native Wrapper still owns its inherited open lease description.
	for {
		time.Sleep(time.Second)
	}
}

func TestNativeRuntimeSIGKILLLeavesWrapperOwnedAndBlocksCompetingRuntime(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	pidFile := filepath.Join(dir, "child.pid")
	runtimeChild := exec.Command(os.Args[0], "-test.run=^TestNativeRuntimeHardKillProcessHelper$", "-test.v")
	runtimeChild.Env = append(os.Environ(), "MAGPIE_RUNTIME_KILL_HELPER=1", "MAGPIE_RUNTIME_CHILD_PID="+pidFile)
	if err := runtimeChild.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if runtimeChild.ProcessState == nil {
			_ = runtimeChild.Process.Kill()
			_ = runtimeChild.Wait()
		}
	}()
	var nativePID int
	deadline := time.Now().Add(7 * time.Second)
	for {
		if b, err := os.ReadFile(pidFile); err == nil {
			nativePID, err = strconv.Atoi(string(b))
			if err != nil {
				t.Fatal(err)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("child Runtime never spawned its mock native wrapper")
		}
		time.Sleep(30 * time.Millisecond)
	}
	defer syscall.Kill(nativePID, syscall.SIGKILL) // prevent leaked test child on failure
	if err := runtimeChild.Process.Signal(syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	_ = runtimeChild.Wait()
	if lease, err := acquireInteractiveNativeLease("hardkill-owner", "runtime-owned-outer"); err == nil {
		releaseInteractiveNativeLease(lease)
		t.Fatal("native wrapper lease vanished merely because Runtime was SIGKILLed")
	} else if !errors.Is(err, errInteractiveOuterLive) {
		t.Fatal(err)
	}
	if err := syscall.Kill(nativePID, syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	deadline = time.Now().Add(3 * time.Second)
	for {
		lease, err := acquireInteractiveNativeLease("hardkill-owner", "runtime-owned-outer")
		if err == nil {
			releaseInteractiveNativeLease(lease)
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("wrapper lease stayed locked after both processes died: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestNativeHardKillDetachedTmuxRefusesRecoveryWithoutToolReplay(t *testing.T) {
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux unavailable for isolated fault injection")
	}
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	socket := "magpie-inject-" + toolEvidenceHash(dir)[:15]
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		t.Fatal(err)
	}
	wrapper := filepath.Join(dir, "tmux")
	if err := os.WriteFile(wrapper, []byte("#!/bin/sh\nexec "+tmux+" -f /dev/null -L "+socket+" \"$@\"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Cleanup(func() { _ = exec.Command(wrapper, "kill-server").Run() })
	ready := filepath.Join(dir, "ready")
	child := exec.Command(os.Args[0], "-test.run=^TestNativeHardKillProcessHelper$", "-test.v")
	child.Env = append(os.Environ(), "MAGPIE_HARDKILL_HELPER=1", "MAGPIE_HARDKILL_WORK="+dir, "MAGPIE_HARDKILL_READY="+ready)
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if child.ProcessState == nil {
			_ = child.Process.Kill()
			_ = child.Wait()
		}
	}()
	deadline := time.Now().Add(8 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("child did not launch isolated tmux pane")
		}
		time.Sleep(30 * time.Millisecond)
	}
	// Wait until the detached tmux pane has actually performed its single
	// harmless side effect, then kill the native wrapper with SIGKILL.
	sideEffects := filepath.Join(dir, "side-effects.log")
	deadline = time.Now().Add(8 * time.Second)
	for {
		if b, err := os.ReadFile(sideEffects); err == nil && strings.Count(string(b), "executed") == 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("detached tmux pane never wrote its one test side effect")
		}
		time.Sleep(30 * time.Millisecond)
	}
	if err := child.Process.Signal(syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	_ = child.Wait()
	// A new Runtime can acquire the wrapper lease: the original process
	// has been killed. That MUST NOT authorize any new tool execution while
	// the detached tmux pane is still alive.
	lease, err := acquireInteractiveNativeLease("hardkill-owner", "hardkill-outer")
	if err != nil {
		t.Fatalf("native wrapper lock still held after SIGKILL: %v", err)
	}
	if err := inspectInteractiveTmuxGuard("hardkill-owner", "hardkill-outer", true); !errors.Is(err, errInteractiveOuterLive) {
		releaseInteractiveNativeLease(lease)
		t.Fatalf("live escaped tmux pane was not fenced after SIGKILL: %v", err)
	}
	releaseInteractiveNativeLease(lease)
	if b, err := os.ReadFile(sideEffects); err != nil || strings.Count(string(b), "executed") != 1 {
		t.Fatalf("external side effect duplicated: %q %v", b, err)
	}
	if out, err := exec.Command(wrapper, "kill-session", "-t", "=claude-bridge-deadbeef").CombinedOutput(); err != nil {
		t.Fatalf("isolated tmux cleanup failed: %v %s", err, out)
	}
	if err := inspectInteractiveTmuxGuard("hardkill-owner", "hardkill-outer", true); err != nil {
		t.Fatalf("old detached tmux stopped but guard still refuses: %v", err)
	}
	if b, err := os.ReadFile(sideEffects); err != nil || strings.Count(string(b), "executed") != 1 {
		t.Fatalf("stopped pane side effect unexpectedly changed: %q %v", b, err)
	}
}
