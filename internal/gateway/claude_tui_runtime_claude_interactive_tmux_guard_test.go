//go:build darwin || linux

package gateway

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeTUIRuntimeNativeTmuxGuardRejectsActiveAndUnknownDetachedSessions(t *testing.T) {
	b, rt, owner, outer, runLog := tuiRuntimeAutoRecoveryFixture(t)
	tuiRuntimeAutoRegister(t, b, owner, outer, "Write", `{"file_path":"note.txt"}`)
	if err := tuiRuntimeInspectInteractiveTmuxGuard(owner, outer, true); err != nil {
		t.Fatalf("inactive fake pane rejected: %v", err)
	}
	t.Setenv("MAGPIE_TEST_TMUX_ALIVE", "1")
	if err := tuiRuntimeInspectInteractiveTmuxGuard(owner, outer, true); !errors.Is(err, tuiRuntimeErrInteractiveOuterLive) {
		t.Fatalf("active detached pane not detected: %v", err)
	}
	status, _, _ := tuiRuntimeArServe(t, rt, b, owner, outer, tuiRuntimeAutoRecoveryBody("Write", `{"file_path":"note.txt"}`, `"wrote"`))
	if status != 409 {
		t.Fatalf("recovery interrupted a still-live tmux pane: %d", status)
	}
	rows, err := b.toolJournal.read()
	if err != nil || len(rows) != 1 {
		t.Fatalf("active tmux probe consumed journal claim: %v %+v", err, rows)
	}
	if _, err := os.Stat(runLog); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("automatic recovery launched while tmux was live")
	}
	t.Setenv("MAGPIE_TEST_TMUX_ALIVE", "0")
	status, msg, out := tuiRuntimeArServe(t, rt, b, owner, outer, tuiRuntimeAutoRecoveryBody("Write", `{"file_path":"note.txt"}`, `"wrote"`))
	if status != 200 || !strings.Contains(out, "OL_ANSWER") {
		t.Fatalf("released detached pane did not allow safe recovery: %d %s %s", status, msg, out)
	}
}

func TestClaudeTUIRuntimeNativeTmuxGuardRefusesLegacyResumeWithoutOriginalMarker(t *testing.T) {
	dir := tuiRuntimeArDir(t)
	t.Setenv("MAGPIE_CLAUDE_DURABLE_TOOL_JOURNAL", "1")
	t.Setenv("MAGPIE_CLAUDE_NATIVE_TMUX_GUARD", "1")
	tuiRuntimeArWrapper(t, dir, "mock-wrapper", tuiRuntimeArAnswerScript())
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer := "owner", "legacy-with-checkpoint"
	tuiRuntimeArSave(t, b, owner, outer, "legacy-inner", false)
	req := &tuiRuntimeRequest{Request: Request{Model: tuiRuntimeAutoTestModel, Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "continue"}}}}}}
	_, _, err := b.startInteractiveWithPromptImages(context.Background(), req, tuiRuntimeAutoTestModel, "", owner, outer, "continue", "legacy-inner", nil, false)
	if err == nil {
		t.Fatal("older uninstrumented Claude/tmux run was resumed without an ownership marker")
	}
	after, ok := b.interactiveSession(owner, outer)
	if !ok || after.Dirty {
		t.Fatalf("unsafe resume changed old checkpoint: %+v", after)
	}
}

func TestClaudeTUIRuntimeNativeTmuxGuardDetectsRealDetachedPaneAfterWrapperDeath(t *testing.T) {
	// This uses a dedicated tmux -L socket namespace so no production/user tmux
	// server or Claude process can ever be affected by the test.
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		t.Skip("tmux is not installed in isolated test environment")
	}
	dir := t.TempDir()
	// AF_UNIX paths on macOS are too short for Go's long test-directory path.
	// This randomized short namespace never touches the user's default tmux.
	socket := "magpie-test-" + tuiRuntimeToolEvidenceHash(dir)[:16]
	wrapper := filepath.Join(dir, "tmux")
	script := fmt.Sprintf("#!/bin/sh\nexec %q -f /dev/null -L %q \"$@\"\n", tmux, socket)
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	const name = "claude-bridge-deadbeef"
	cmd := exec.Command(wrapper, "new-session", "-d", "-s", name, "-c", dir, "sleep 20")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("isolated tmux cannot launch here: %v (%s)", err, output)
	}
	defer exec.Command(wrapper, "kill-server").Run()
	owner, outer := "owner", "orphan-synthetic"
	tuiRuntimeWriteTestNativeTmuxMarker(t, owner, outer, name)
	// At this point the short-lived tmux *launcher* has already exited,
	// but the real detached pane is still alive: this is the gap the native
	// wrapper file-descriptor lock by itself cannot cover.
	if err := tuiRuntimeInspectInteractiveTmuxGuard(owner, outer, true); !errors.Is(err, tuiRuntimeErrInteractiveOuterLive) {
		t.Fatalf("detached pane falsely considered dead after launcher exit: %v", err)
	}
	if output, err := exec.Command(wrapper, "kill-session", "-t", "="+name).CombinedOutput(); err != nil {
		t.Fatalf("test tmux cleanup failed: %v %s", err, output)
	}
	if err := tuiRuntimeInspectInteractiveTmuxGuard(owner, outer, true); err != nil {
		t.Fatalf("terminated tmux pane still blocked read-only recovery: %v", err)
	}
}

func TestClaudeTUIRuntimeNativeTmuxGuardFailsClosedOnMissingCorruptAndAmbiguousMetadata(t *testing.T) {
	b, rt, owner, outer, runLog := tuiRuntimeAutoRecoveryFixture(t)
	tuiRuntimeAutoRegister(t, b, owner, outer, "Write", `{"file_path":"note.txt"}`)
	path := tuiRuntimeInteractiveTmuxGuardPath(owner, outer)
	for _, tc := range []struct {
		name string
		data []byte
		mode os.FileMode
	}{
		{"missing", nil, 0},
		{"truncated", []byte(`{"version":1`), 0o600},
		{"unsafe permissions", []byte(`{"version":1,"session":"claude-bridge-deadbeef"}`), 0o644},
		{"bad session identifier", []byte(`{"version":1,"session":"other-user-session"}`), 0o600},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_ = os.Remove(path)
			if tc.data != nil {
				if err := os.WriteFile(path, tc.data, tc.mode); err != nil {
					t.Fatal(err)
				}
			}
			if err := tuiRuntimeInspectInteractiveTmuxGuard(owner, outer, true); err == nil {
				t.Fatal("unverifiable pane identity accepted")
			}
			status, _, _ := tuiRuntimeArServe(t, rt, b, owner, outer, tuiRuntimeAutoRecoveryBody("Write", `{"file_path":"note.txt"}`, `"wrote"`))
			if status != 409 {
				t.Fatalf("uncertain detached pane led to recovery: %d", status)
			}
			rows, err := b.toolJournal.read()
			if err != nil || len(rows) != 1 {
				t.Fatalf("unverifiable marker consumed claim: %v %+v", err, rows)
			}
			if _, err := os.Stat(runLog); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("unexpected native process")
			}
		})
	}
}

func TestClaudeTUIRuntimeNativeTmuxGuardRequiresExplicitEnablementForOrphanAutoRecovery(t *testing.T) {
	b, rt, owner, outer, runLog := tuiRuntimeAutoRecoveryFixture(t)
	tuiRuntimeAutoRegister(t, b, owner, outer, "Write", `{"file_path":"note.txt"}`)
	t.Setenv("MAGPIE_CLAUDE_NATIVE_TMUX_GUARD", "0")
	status, _, _ := tuiRuntimeArServe(t, rt, b, owner, outer, tuiRuntimeAutoRecoveryBody("Write", `{"file_path":"note.txt"}`, `"wrote"`))
	if status != 409 {
		t.Fatalf("orphan recovered without native pane fencing: %d", status)
	}
	if _, err := os.Stat(runLog); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("unsafe recovery process started")
	}
}

func TestClaudeTUIRuntimeNativeTmuxGuardRefusesChangedSocketNamespace(t *testing.T) {
	b, _, owner, outer, _ := tuiRuntimeAutoRecoveryFixture(t)
	tuiRuntimeAutoRegister(t, b, owner, outer, "Write", `{"file_path":"note.txt"}`)
	if err := tuiRuntimeInspectInteractiveTmuxGuard(owner, outer, true); err != nil {
		t.Fatalf("correct namespace rejected: %v", err)
	}
	// The old detached Claude may still exist under a different socket.
	// Never treat a query of a *new* socket as proof the old one is dead.
	t.Setenv("TMUX_TMPDIR", filepath.Join(t.TempDir(), "different-tmux-root"))
	if err := tuiRuntimeInspectInteractiveTmuxGuard(owner, outer, true); err == nil {
		t.Fatal("changed tmux socket namespace was accepted")
	}
	// A change in TMUX itself is similarly security-significant.
	t.Setenv("TMUX_TMPDIR", "")
	t.Setenv("TMUX", "/tmp/another-socket,900,1")
	if err := tuiRuntimeInspectInteractiveTmuxGuard(owner, outer, true); err == nil {
		t.Fatal("changed nested tmux environment was accepted")
	}
}

func TestClaudeTUIRuntimeNativeTmuxGuardRejectsUnknownTmuxExitStatus(t *testing.T) {
	b, _, owner, outer, _ := tuiRuntimeAutoRecoveryFixture(t)
	tuiRuntimeAutoRegister(t, b, owner, outer, "Write", `{"file_path":"note.txt"}`)
	dir := filepath.Dir(os.Getenv("MAGPIE_AUTO_TEST_RUNS"))
	if err := os.WriteFile(filepath.Join(dir, "tmux"), []byte("#!/bin/sh\necho 'permission denied' >&2\nexit 1\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := tuiRuntimeInspectInteractiveTmuxGuard(owner, outer, true); err == nil {
		t.Fatal("unknown tmux process state misread as absence")
	}
}
