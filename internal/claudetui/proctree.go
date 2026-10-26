package claudetui

import (
	"context"
	"os/exec"
	"sync"
	"time"

	"github.com/yetone/magpie/internal/proc"
)

// commandContext is proc.CommandContext with a bounded WaitDelay after ctx
// ends, so a detached child can't hold the run's pipes forever.
func commandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := proc.CommandContext(ctx, name, args...)
	cmd.WaitDelay = 2 * time.Second
	return cmd
}

// procTree is a run's wrapper and every process it starts. tmux gives a
// pane a session of its own, so Claude Code and its MCP helper are not in
// the wrapper's group: the tree also keeps the processes it has seen under
// the wrapper (seen), each by its start time, and ends those still running
// as the same process once the wrapper has gone. It ends nothing by name.
type procTree struct {
	tree *proc.Tree
	pid  int

	mu   sync.Mutex
	seen map[int]uint64 // pid → start time, of the wrapper's descendants
}

func startTree(cmd *exec.Cmd) (*procTree, error) {
	endWithParent(cmd)
	t, err := proc.StartTree(cmd)
	if err != nil {
		return nil, err
	}
	return &procTree{tree: t, pid: cmd.Process.Pid, seen: map[int]uint64{}}, nil
}

// look notes the wrapper's descendants as they are now.
func (t *procTree) look() {
	found := descendants(t.pid)
	t.mu.Lock()
	for pid, start := range found {
		t.seen[pid] = start
	}
	t.mu.Unlock()
}

// rss is what the wrapper and every process seen under it hold resident,
// and false when it could not be read.
func (t *procTree) rss() (int64, bool) {
	t.look()
	t.mu.Lock()
	pids := map[int]uint64{t.pid: 0}
	for pid, start := range t.seen {
		pids[pid] = start
	}
	t.mu.Unlock()
	return groupRSS(pids)
}

// Wait waits for the wrapper, then ends what it and its panes left. It is
// called once the wrapper's output has ended, when what it started has
// been adopted by init already: what was seen under it before (the
// readings, abort) is what is ended.
func (t *procTree) Wait() error {
	err := t.tree.Wait()
	t.sweep()
	return err
}

// Kill ends the wrapper, its group and every process seen under it.
func (t *procTree) Kill() {
	t.look()
	t.tree.Kill()
	t.sweep()
}

// sweep kills the processes seen under the wrapper that still run as the
// process seen.
func (t *procTree) sweep() {
	t.mu.Lock()
	seen := t.seen
	t.seen = map[int]uint64{}
	t.mu.Unlock()
	killSeen(seen)
}
