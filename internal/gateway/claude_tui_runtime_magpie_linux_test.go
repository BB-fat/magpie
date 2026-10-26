package gateway

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/proc"
)

// The process tree and tmux sweep read /proc: Linux, where the Claude TUI
// runs.

func TestClaudeTUIRuntimeTUIFixedWorkDirsRefuseOtherOwners(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root to create a writable cwd owned by another uid")
	}
	for _, work := range tuiRuntimeFixedTUIWorkDirs() {
		t.Run(work.name, func(t *testing.T) {
			tuiRuntimeWithConfig(t, &tuiRuntimeConfig{TempRoot: t.TempDir()})
			dir, err := work.make()
			if err != nil {
				t.Fatal(err)
			}
			marker := filepath.Join(dir, "CLAUDE.md")
			if err := os.WriteFile(marker, []byte("another user's instructions"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.Chown(dir, os.Getuid()+1, -1); err != nil {
				t.Fatal(err)
			}
			probe, err := os.CreateTemp(dir, ".writable-by-root-*")
			if err != nil {
				t.Fatalf("fixture must be writable by root: %v", err)
			}
			probe.Close()
			os.Remove(probe.Name())
			if got, err := work.make(); err == nil || got != "" || !strings.Contains(err.Error(), "belongs to another user") {
				t.Fatalf("another user's writable cwd accepted: %q, %v", got, err)
			}
			if data, err := os.ReadFile(marker); err != nil || string(data) != "another user's instructions" {
				t.Fatalf("another user's file changed: %q, %v", data, err)
			}
		})
	}
}

func TestClaudeTUIRuntimeTUIFixedWorkDirsRefuseOtherWriters(t *testing.T) {
	for _, work := range tuiRuntimeFixedTUIWorkDirs() {
		t.Run(work.name, func(t *testing.T) {
			tuiRuntimeWithConfig(t, &tuiRuntimeConfig{TempRoot: t.TempDir()})
			dir, err := work.make()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(dir, 0o770); err != nil {
				t.Fatal(err)
			}
			if got, err := work.make(); err == nil || got != "" || !strings.Contains(err.Error(), "can be written by others") {
				t.Fatalf("shared writable cwd accepted: %q, %v", got, err)
			}
		})
	}
}

// A run's tree is ended whole: its wrapper's group, and what was seen
// under it in a session of its own, as tmux starts a pane.
func TestClaudeTUIRuntimeProcTreeEndsWhatRanUnderIt(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "pane")
	cmd := tuiRuntimeCommandContext(context.Background(), "/bin/sh", "-c", "setsid sh -c 'echo $$ > "+pidFile+"; exec sleep 60' & sleep 60")
	tree, err := tuiRuntimeStartTree(cmd)
	if err != nil {
		t.Fatal(err)
	}
	var pane int
	for i := 0; i < 200 && pane == 0; i++ {
		b, _ := os.ReadFile(pidFile)
		pane, _ = strconv.Atoi(strings.TrimSpace(string(b)))
		time.Sleep(10 * time.Millisecond)
	}
	if pane == 0 {
		t.Fatal("the pane stand-in never started")
	}
	if sid, _ := syscall.Getpgid(pane); sid == cmd.Process.Pid {
		t.Fatal("the stand-in pane is in the wrapper's group: it tests nothing")
	}
	// no reading has looked under the wrapper: Kill finds the pane itself
	tree.Kill()
	_ = tree.Wait()
	for i := 0; i < 200; i++ {
		if !tuiRuntimeSame(pane, 0) || tuiRuntimeZombie(pane) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	_ = syscall.Kill(pane, syscall.SIGKILL)
	t.Fatal("the process in a session of its own outlived the tree")
}

// A tree's memory is its processes', a session of their own among them.
func TestClaudeTUIRuntimeProcTreeMeasuresWhatRunsUnderIt(t *testing.T) {
	cmd := tuiRuntimeCommandContext(context.Background(), "/bin/sh", "-c", "setsid sleep 60 & sleep 60")
	tree, err := tuiRuntimeStartTree(cmd)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { tree.Kill(); _ = tree.Wait() }()
	var n int64
	for i := 0; i < 200; i++ {
		tree.look()
		tree.mu.Lock()
		seen := len(tree.seen)
		tree.mu.Unlock()
		if seen >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	n, ok := tree.rss()
	alone, _ := tuiRuntimePidRSS(cmd.Process.Pid)
	if !ok || n <= alone {
		t.Fatalf("the tree holds %d, the wrapper alone %d: what runs under it isn't counted", n, alone)
	}
}

func tuiRuntimeZombie(pid int) bool {
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return true
	}
	s := string(b)
	f := strings.Fields(s[strings.LastIndexByte(s, ')')+1:])
	return len(f) > 0 && f[0] == "Z"
}

// A gateway coming up ends the tmux servers one that died left, by their
// sockets, and leaves a live gateway's and its own.
func TestClaudeTUIRuntimeSweepEndsWhatADeadGatewayLeft(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "tmux.log")
	tmux := filepath.Join(dir, "tmux")
	if err := os.WriteFile(tmux, []byte("#!/bin/sh\necho \"$@\" >> "+log+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeWithConfig(t, &tuiRuntimeConfig{Tmux: tmux, TempRoot: dir})
	d, err := tuiRuntimeTmuxDir()
	if err != nil {
		t.Fatal(err)
	}
	dead := tuiRuntimeDeadPID(t)
	for _, name := range []string{fmt.Sprintf("%d-aaaaaaaaaaaa.sock", dead), fmt.Sprintf("%d-bbbbbbbbbbbb.sock", os.Getpid()), "1-cccccccccccc.sock", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(d, name), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	// and the runs' temp folders, named for their gateway
	runDirs := map[string]bool{
		fmt.Sprintf("magpie-claude-pty-%d-111", dead):        false,
		fmt.Sprintf("magpie-claude-pty-%d-222", os.Getpid()): true,
		"magpie-claude-pty-1-333":                            true,
		"magpie-claude-pty-444":                              true, // no gateway named: left alone
	}
	for name := range runDirs {
		if err := os.MkdirAll(filepath.Join(dir, name), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name, "callback"), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	tuiRuntimeSweep()
	for name, kept := range runDirs {
		if _, err := os.Stat(filepath.Join(dir, name)); (err == nil) != kept {
			t.Errorf("run folder %s kept %v, want %v", name, err == nil, kept)
		}
	}
	b, _ := os.ReadFile(log)
	if got := strings.TrimSpace(string(b)); got != "-S "+filepath.Join(d, fmt.Sprintf("%d-aaaaaaaaaaaa.sock", dead))+" kill-server" {
		t.Fatalf("tmux was asked %q", got)
	}
	for name, kept := range map[string]bool{fmt.Sprintf("%d-aaaaaaaaaaaa.sock", dead): false, fmt.Sprintf("%d-bbbbbbbbbbbb.sock", os.Getpid()): true, "1-cccccccccccc.sock": true, "notes.txt": true} {
		if _, err := os.Stat(filepath.Join(d, name)); (err == nil) != kept {
			t.Errorf("%s kept %v, want %v", name, err == nil, kept)
		}
	}
}

// deadPID is a pid no process has now.
func tuiRuntimeDeadPID(t *testing.T) int {
	cmd := proc.Command("/bin/true")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	return cmd.Process.Pid
}
