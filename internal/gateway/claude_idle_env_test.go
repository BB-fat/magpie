package gateway

import (
	"strconv"
	"testing"
	"time"
)

func TestClaudeIdleConfigDefaults(t *testing.T) {
	restoreDefaultClaudeBridgeEnvAfterTest(t)
	ResetClaudeIdleConfigForTest()
	t.Setenv(claudeIdleMaxEnv, "")
	t.Setenv(claudeIdleTTLEnv, "")
	if err := LoadClaudeIdleConfig(); err != nil {
		t.Fatal(err)
	}
	if idleMost != 6 || idleLongest != time.Hour {
		t.Fatalf("defaults: most=%d longest=%s", idleMost, idleLongest)
	}
}

func TestClaudeIdleConfigCustom(t *testing.T) {
	restoreDefaultClaudeBridgeEnvAfterTest(t)
	ResetClaudeIdleConfigForTest()
	t.Setenv(claudeIdleMaxEnv, "2")
	t.Setenv(claudeIdleTTLEnv, "10m")
	if err := LoadClaudeIdleConfig(); err != nil {
		t.Fatal(err)
	}
	if idleMost != 2 || idleLongest != 10*time.Minute {
		t.Fatalf("custom: most=%d longest=%s", idleMost, idleLongest)
	}
}

func TestClaudeIdleConfigZeroMax(t *testing.T) {
	restoreDefaultClaudeBridgeEnvAfterTest(t)
	ResetClaudeIdleConfigForTest()
	t.Setenv(claudeIdleMaxEnv, "0")
	if err := LoadClaudeIdleConfig(); err != nil {
		t.Fatal(err)
	}
	if idleMost != 0 {
		t.Fatalf("idleMost = %d, want 0", idleMost)
	}
}

func TestClaudeIdleConfigInvalid(t *testing.T) {
	restoreDefaultClaudeBridgeEnvAfterTest(t)
	for _, tc := range []struct {
		env, val string
	}{
		{claudeIdleMaxEnv, "-1"},
		{claudeIdleMaxEnv, "x"},
		{claudeIdleTTLEnv, "0"},
		{claudeIdleTTLEnv, "not-a-duration"},
	} {
		ResetClaudeIdleConfigForTest()
		t.Setenv(claudeIdleMaxEnv, "")
		t.Setenv(claudeIdleTTLEnv, "")
		t.Setenv(tc.env, tc.val)
		if err := LoadClaudeIdleConfig(); err == nil {
			t.Fatalf("%s=%q: want error", tc.env, tc.val)
		}
	}
}

func TestClaudeIdleMostZeroEvictsImmediately(t *testing.T) {
	restoreDefaultClaudeBridgeEnvAfterTest(t)
	ResetClaudeIdleConfigForTest()
	t.Setenv(claudeIdleMaxEnv, "0")
	if err := LoadClaudeIdleConfig(); err != nil {
		t.Fatal(err)
	}
	b := idleTestBridge()
	r := &subscriptionRun{bridge: b}
	if !b.keepIdle(r, "only", "conv") {
		t.Fatal("keepIdle refused a live run")
	}
	if len(b.idle) != 0 {
		t.Fatalf("idleMost=0: %d runs still in idle map, want 0", len(b.idle))
	}
}

func TestClaudeIdleMostCapEviction(t *testing.T) {
	restoreDefaultClaudeBridgeEnvAfterTest(t)
	ResetClaudeIdleConfigForTest()
	t.Setenv(claudeIdleMaxEnv, "2")
	if err := LoadClaudeIdleConfig(); err != nil {
		t.Fatal(err)
	}
	b := idleTestBridge()
	for i := 0; i < 4; i++ {
		b.keepIdle(&subscriptionRun{bridge: b}, fmtKey(i), "")
	}
	if len(b.idle) != 2 {
		t.Fatalf("idle map len = %d, want 2", len(b.idle))
	}
}

func fmtKey(i int) string {
	return strconv.Itoa(i)
}
