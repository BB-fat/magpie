package claudetui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/yetone/magpie/internal/proc"
)

// Each run's wrapper starts a tmux server of its own, in the foreground
// (tmux -D) on a socket in tmuxDir, with no config (-f /dev/null): never
// the user's tmux, and one magpie can end by its socket alone. A socket is
// named for the gateway that made it, so a gateway coming up ends what one
// that died left (Sweep), and leaves a live one's (a restart to update
// hands the gateway over while the old one's runs finish).

// tmuxDir is where runs' tmux sockets are, this user's alone. A socket's
// path must fit a sun_path (108 bytes), so it is short and under the temp
// folder.
func tmuxDir() (string, error) {
	d := filepath.Join(tempRoot(), "magpie-claude-tui-"+strconv.Itoa(os.Getuid()))
	return ownedWorkDir(d)
}

// tmuxSocket is a new run's socket path.
func tmuxSocket(token string) (string, error) {
	d, err := tmuxDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, fmt.Sprintf("%d-%s.sock", os.Getpid(), token[:12])), nil
}

// tmuxWait bounds a kill-server.
var tmuxWait = 5 * time.Second

// killTmux ends the tmux server on sock, its panes (Claude Code) with it,
// and removes the socket.
func killTmux(sock string) {
	if sock == "" {
		return
	}
	if _, err := os.Lstat(sock); err != nil {
		return
	}
	if config.Tmux != "" {
		ctx, cancel := context.WithTimeout(context.Background(), tmuxWait)
		_ = proc.CommandContext(ctx, config.Tmux, "-S", sock, "kill-server").Run()
		cancel()
	}
	_ = os.Remove(sock)
}

// runDirPrefix starts the name of a run's own temp folder (its callback
// and tools files), then the gateway's pid.
const runDirPrefix = "magpie-claude-pty-"

// Sweep removes what runs of a gateway no longer running left: their tmux
// servers, ended by their sockets in tmuxDir, and their temp folders.
// Nothing is found by a process's name.
func Sweep() {
	dirs, _ := filepath.Glob(filepath.Join(tempRoot(), runDirPrefix+"*-*"))
	for _, d := range dirs {
		pid, err := strconv.Atoi(strings.SplitN(strings.TrimPrefix(filepath.Base(d), runDirPrefix), "-", 2)[0])
		if err != nil || pid == os.Getpid() || gatewayAlive(pid) {
			continue
		}
		if info, err := os.Lstat(d); err == nil && info.IsDir() {
			_ = os.RemoveAll(d)
		}
	}
	sweepTmux()
}

// sweepTmux ends the tmux servers runs of a gateway no longer running left.
func sweepTmux() {
	d, err := tmuxDir()
	if err != nil {
		return
	}
	entries, err := os.ReadDir(d)
	if err != nil {
		return
	}
	for _, e := range entries {
		name, ok := strings.CutSuffix(e.Name(), ".sock")
		if !ok {
			continue
		}
		pid, err := strconv.Atoi(strings.SplitN(name, "-", 2)[0])
		if err != nil || pid == os.Getpid() || gatewayAlive(pid) {
			continue
		}
		killTmux(filepath.Join(d, e.Name()))
	}
}

// gatewayAlive says pid still runs. A pid taken again by another process
// keeps a dead gateway's socket until a later sweep: nothing is ended on a
// guess.
func gatewayAlive(pid int) bool {
	_, err := os.Stat("/proc/" + strconv.Itoa(pid))
	return err == nil
}
