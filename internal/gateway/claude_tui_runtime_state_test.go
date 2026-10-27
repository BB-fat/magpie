package gateway

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClaudeTUIRuntimeInteractiveSessionStateWritten(t *testing.T) {
	tmp := t.TempDir()
	tuiRuntimeTestInit(t, &tuiRuntimeConfig{StateDir: tmp})
	b := tuiRuntimeNewSubscriptionBridge()
	e := tuiRuntimeInteractiveSessionEntry{SessionID: "s1", ConvKey: "k", InstructionKey: "c", ThreadKey: "t", Model: "m"}
	b.saveInteractiveSession("owner", "outer", e)
	p := filepath.Join(tmp, "claude-interactive-sessions.json")
	b2, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(b2) == 0 {
		t.Fatal("empty state file")
	}
	t.Log(string(b2))
}
