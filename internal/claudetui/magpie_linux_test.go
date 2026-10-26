package claudetui

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

// A run's tree is ended whole: its wrapper's group, and what was seen
// under it in a session of its own, as tmux starts a pane.
func TestProcTreeEndsWhatRanUnderIt(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "pane")
	cmd := commandContext(context.Background(), "/bin/sh", "-c", "setsid sh -c 'echo $$ > "+pidFile+"; exec sleep 60' & sleep 60")
	tree, err := startTree(cmd)
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
		if !same(pane, 0) || zombie(pane) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	_ = syscall.Kill(pane, syscall.SIGKILL)
	t.Fatal("the process in a session of its own outlived the tree")
}

// A tree's memory is its processes', a session of their own among them.
func TestProcTreeMeasuresWhatRunsUnderIt(t *testing.T) {
	cmd := commandContext(context.Background(), "/bin/sh", "-c", "setsid sleep 60 & sleep 60")
	tree, err := startTree(cmd)
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
	alone, _ := pidRSS(cmd.Process.Pid)
	if !ok || n <= alone {
		t.Fatalf("the tree holds %d, the wrapper alone %d: what runs under it isn't counted", n, alone)
	}
}

func zombie(pid int) bool {
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
func TestSweepEndsWhatADeadGatewayLeft(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "tmux.log")
	tmux := filepath.Join(dir, "tmux")
	if err := os.WriteFile(tmux, []byte("#!/bin/sh\necho \"$@\" >> "+log+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	withConfig(t, &Config{Tmux: tmux, TempRoot: dir})
	d, err := tmuxDir()
	if err != nil {
		t.Fatal(err)
	}
	dead := deadPID(t)
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
	Sweep()
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
func deadPID(t *testing.T) int {
	cmd := proc.Command("/bin/true")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	return cmd.Process.Pid
}
