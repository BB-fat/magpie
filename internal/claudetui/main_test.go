package claudetui

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/yetone/magpie/internal/testenv"
)

// The tests run in a home of their own, never the user's real files, and
// keep their runs' folders and sockets in it too (TempRoot), removed with
// it.
// A run finds Claude Code on the test's PATH, where a test puts its
// stand-in.
func TestMain(m *testing.M) {
	os.Exit(testenv.RunIn(m, func(home string) {
		tmp := filepath.Join(home, "tmp")
		_ = os.MkdirAll(tmp, 0o700)
		Init(&Config{ClaudePath: func() (string, error) { return exec.LookPath("claude") }, TempRoot: tmp})
	}))
}
