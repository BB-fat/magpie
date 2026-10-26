package claudetui

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/yetone/magpie/internal/proc"
)

// wrapperBundle is the PTY wrapper (wrapper/src) built with its locked
// dependencies into one file: `bun build ./src/cli.ts --target=bun
// --outfile=dist/claude-bridge.js` in wrapper/. magpie carries it; nothing
// is downloaded when it runs.
//
//go:embed wrapper/dist/claude-bridge.js
var wrapperBundle []byte

// InstallWrapper writes the wrapper bundle where runs start it, in a folder
// named for its contents (so a run's hooks, which name the bundle, find
// the same file while a newer magpie writes its own), and gives its path.
// The file is written whole or not at all, and only this user reads it.
func InstallWrapper() (string, error) {
	sum := sha256.Sum256(wrapperBundle)
	dir := filepath.Join(mustStateDir(), "wrapper", hex.EncodeToString(sum[:])[:16])
	path := filepath.Join(dir, "claude-bridge.js")
	if b, err := os.ReadFile(path); err == nil && sha256.Sum256(b) == sum {
		return path, nil
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	f, err := os.CreateTemp(dir, ".claude-bridge-*.js")
	if err != nil {
		return "", err
	}
	tmp := f.Name()
	_, werr := f.Write(wrapperBundle)
	if serr := f.Sync(); werr == nil {
		werr = serr
	}
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	if werr == nil {
		werr = os.Rename(tmp, path)
	}
	if werr != nil {
		_ = os.Remove(tmp)
		return "", werr
	}
	return path, nil
}

// The least versions the bridge runs with: tmux's -D (a server in the
// foreground) is 3.2's; Claude Code's alwaysLoad MCP servers and the hooks
// the wrapper reads are 2.1.268's, the plugin's own floor.
const (
	MinTmux   = "3.2"
	MinClaude = "2.1.268"
	MinBun    = "1.1.0"
)

// Deps are the programs a TUI run needs, as found on this machine.
type Deps struct {
	Tmux, TmuxVersion     string
	Claude, ClaudeVersion string
	Bun                   string
}

// Supported says why the Claude TUI can't run on this platform, nil when
// it can: Linux only.
func Supported() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("the Claude TUI runs on Linux only, not %s", runtime.GOOS)
	}
	return nil
}

// Check finds what a TUI run needs: tmux, Claude Code (claude, found by
// the gateway) and Bun (bun, magpie's own or $MAGPIE_BUN). It installs,
// downloads and asks for nothing: what is missing is said, with how to get
// it. Claude Code is asked its version, a process of the gateway's room
// (config.Room).
func Check(ctx context.Context, claude, bun func() (string, error)) (Deps, error) {
	var d Deps
	if err := Supported(); err != nil {
		return d, err
	}
	var errs []error
	d.Tmux = proc.FindTool("tmux")
	if d.Tmux == "" {
		errs = append(errs, errors.New("tmux is not installed: install it with the system's package manager (apt install tmux, dnf install tmux)"))
	} else if v, err := toolVersion(ctx, d.Tmux, "-V"); err != nil {
		errs = append(errs, fmt.Errorf("tmux (%s) doesn't say its version: %w", d.Tmux, err))
	} else if d.TmuxVersion = v; older(v, MinTmux) {
		errs = append(errs, fmt.Errorf("tmux %s (%s) is older than %s, which the Claude TUI needs", v, d.Tmux, MinTmux))
	}
	if p, err := claude(); err != nil {
		errs = append(errs, err)
	} else {
		d.Claude = p
		free, err := config.Room(ctx, "a Claude Code version check")
		if err != nil {
			errs = append(errs, err)
		} else {
			v, err := toolVersion(ctx, p, "--version")
			free(false)
			if err != nil {
				errs = append(errs, fmt.Errorf("Claude Code (%s) doesn't say its version: %w", p, err))
			} else if d.ClaudeVersion = v; older(v, MinClaude) {
				errs = append(errs, fmt.Errorf("Claude Code %s (%s) is older than %s, which the Claude TUI needs: update it (claude update, or magpie claude-code install)", v, p, MinClaude))
			}
		}
	}
	if p, err := bun(); err != nil {
		errs = append(errs, err)
	} else {
		d.Bun = p
	}
	return d, errors.Join(errs...)
}

var versionRe = regexp.MustCompile(`\d+(?:\.\d+)+`)

// toolVersion is the first version a program prints when asked with flag.
func toolVersion(ctx context.Context, exe, flag string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	out, err := proc.CommandContext(ctx, exe, flag).Output()
	if err != nil {
		return "", err
	}
	v := versionRe.FindString(string(out))
	if v == "" {
		return "", fmt.Errorf("no version in %q", strings.TrimSpace(string(out)))
	}
	return v, nil
}

// older says version a is older than b, each dotted numbers.
func older(a, b string) bool {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < max(len(as), len(bs)); i++ {
		var x, y int
		if i < len(as) {
			x, _ = strconv.Atoi(as[i])
		}
		if i < len(bs) {
			y, _ = strconv.Atoi(bs[i])
		}
		if x != y {
			return x < y
		}
	}
	return false
}
