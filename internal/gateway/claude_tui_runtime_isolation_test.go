package gateway

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func initTUIRuntimeForTests(home string) error {
	temporary := filepath.Join(home, "tui-tmp")
	if err := os.MkdirAll(temporary, 0o700); err != nil {
		return err
	}
	initTUIRuntime(&tuiRuntimeConfig{
		StateDir: os.Getenv("MAGPIE_TUI_TEST_STATE_DIR"),
		TempRoot: temporary,
		ClaudePath: func() (string, error) {
			return exec.LookPath("claude")
		},
	})
	return nil
}

func tuiRuntimeWaitStopped(test *testing.T, bridge *tuiRuntimeSubscriptionBridge) {
	test.Helper()
	waitFor(test, "the TUI runtime finishes cleanup", func() bool {
		bridge.mu.Lock()
		defer bridge.mu.Unlock()
		return len(bridge.runs) == 0
	})
}

func tuiRuntimeStopBridge(test *testing.T, bridge *tuiRuntimeSubscriptionBridge) {
	test.Helper()
	bridge.abortAll()
	tuiRuntimeWaitStopped(test, bridge)
}
