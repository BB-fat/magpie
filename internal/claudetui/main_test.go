package claudetui

import (
	"os"
	"os/exec"
	"testing"

	"github.com/yetone/magpie/internal/testenv"
)

// The tests run in a home of their own, never the user's real files. A run
// finds Claude Code on the test's PATH, where a test puts its stand-in.
func TestMain(m *testing.M) {
	os.Exit(testenv.RunIn(m, func(home string) {
		Init(&Config{ClaudePath: func() (string, error) { return exec.LookPath("claude") }})
	}))
}
