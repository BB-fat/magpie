package gateway

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeTUIRuntimeInteractiveNativeAutoCompactionEnabled(t *testing.T) {
	tuiRuntimeFakeClaude(t)
	script := `#!/bin/sh
printf '%s\n' "${DISABLE_AUTO_COMPACT-unset}" > "$COMPACTION_ENV_CAPTURE"
printf '%s\n' '{"type":"system","subtype":"init","session_id":"compact-test"}' '{"type":"assistant","message":{"id":"m1","model":"claude-opus-5-5","content":[{"type":"text","text":"OK"}],"stop_reason":"end_turn"}}' '{"type":"result","subtype":"success","session_id":"compact-test","result":"OK"}'`
	wrapper := filepath.Join(t.TempDir(), "wrapper")
	if err := os.WriteFile(wrapper, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	path := filepath.Join(t.TempDir(), "env")
	t.Setenv("COMPACTION_ENV_CAPTURE", path)
	t.Setenv("DISABLE_AUTO_COMPACT", "1")
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	req := &tuiRuntimeRequest{Request: Request{Model: "claude-opus-5-5", Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}}}}}
	run, events, err := b.startInteractiveOneTurn(context.Background(), req, req.Model, "", "owner", "")
	if err != nil {
		t.Fatal(err)
	}
	defer run.abort()
	for range events {
	}
	value, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(value)) != "unset" {
		t.Fatalf("native auto-compaction disabled: %q", value)
	}
}

func TestClaudeTUIRuntimeInteractiveDesktopIsSingleAutomaticCompactionOwner(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	tuiRuntimeFakeClaude(t)
	script := `#!/bin/sh
printf '%s\n' "${DISABLE_AUTO_COMPACT-unset}" > "$COMPACTION_ENV_CAPTURE"
printf '%s\n' '{"type":"system","subtype":"init","session_id":"desktop-compaction-test"}' '{"type":"assistant","message":{"id":"m1","model":"claude-opus-5-5","content":[{"type":"text","text":"OK"}],"stop_reason":"end_turn"}}' '{"type":"result","subtype":"success","session_id":"desktop-compaction-test","result":"OK"}'`
	wrapper := filepath.Join(t.TempDir(), "wrapper")
	if err := os.WriteFile(wrapper, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	path := filepath.Join(t.TempDir(), "env")
	t.Setenv("COMPACTION_ENV_CAPTURE", path)
	t.Setenv("DISABLE_AUTO_COMPACT", "0")
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	req := &tuiRuntimeRequest{Request: Request{Model: "claude-opus-5-5", Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "hello"}}}}}}
	run, events, err := b.startInteractiveOneTurn(context.Background(), req, req.Model, "", "owner", "synthetic-desktop-outer")
	if err != nil {
		t.Fatal(err)
	}
	defer run.abort()
	for range events {
	}
	value, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(value)) != "1" {
		t.Fatalf("private Claude independently compacted a Desktop-owned session: %q", value)
	}
}

func TestClaudeTUIRuntimeInteractiveClientCompactionEnvIsolation(t *testing.T) {
	parent := []string{"PATH=/usr/bin", "DISABLE_AUTO_COMPACT=0", "CUSTOM_TOOL=keep"}
	for _, tc := range []struct {
		name     string
		client   bool
		expected string
	}{
		{"external client owns session", true, "DISABLE_AUTO_COMPACT=1"},
		{"one-shot native session", false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			child := tuiRuntimeInteractiveClaudeEnv(parent, tc.client)
			found := []string{}
			for _, e := range child {
				if strings.HasPrefix(e, "DISABLE_AUTO_COMPACT=") {
					found = append(found, e)
				}
			}
			if tc.expected == "" && len(found) != 0 || tc.expected != "" && (len(found) != 1 || found[0] != tc.expected) {
				t.Fatalf("invalid isolated auto-compaction configuration %q", found)
			}
			if !strings.Contains(strings.Join(child, "\n"), "CUSTOM_TOOL=keep") || parent[1] != "DISABLE_AUTO_COMPACT=0" {
				t.Fatal("modified unrelated child environment or parent process")
			}
		})
	}
}
