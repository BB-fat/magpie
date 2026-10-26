package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Corrupt or unreadable persisted state must never be reinterpreted as an
// unknown new conversation. That path may otherwise start a second native
// Claude session against a history whose original checkpoint is still owned.
func TestClaudeTUIRuntimeInteractiveInvalidCheckpointStoreFailsBeforeNewNativeRun(t *testing.T) {
	for _, tc := range []struct {
		name, file string
	}{
		{"invalid JSON", `{"version":`},
		{"incompatible version", `{"version":999,"sessions":{}}`},
		{"missing session map", `{"version":1}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			path := tuiRuntimeInteractiveSessionStatePath()
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(tc.file), 0o600); err != nil {
				t.Fatal(err)
			}
			b := tuiRuntimeNewSubscriptionBridge()
			defer tuiRuntimeBridges.Delete(b)
			rt := &tuiRuntime{bridge: b}
			body := []byte(`{"model":"claude-opus-5-5","messages":[{"role":"user","content":"continue old task"}]}`)
			started := 0
			start := func(context.Context, *Request) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
				started++
				return nil, nil, errors.New("would start a new native session")
			}
			w := httptest.NewRecorder()
			code, _ := rt.serveMessagesInner(tuiRuntimeAnthropicResponder{w, httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(string(body))), &Usage{}}, httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(string(body))), "Claude Code", "claude-opus-5-5", "owner", "persisted-conversation", body, start)
			if code != http.StatusConflict || started != 0 {
				t.Fatalf("invalid checkpoint started=%d code=%d; must fail before native run", started, code)
			}
		})
	}
}

func TestClaudeTUIRuntimeInteractiveCheckpointStoreIntegrityAcceptsValidAndAbsentButRejectsSymlink(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	if err := b.verifyInteractiveSessionStore(); err != nil {
		t.Fatalf("fresh account with no checkpoint store refused: %v", err)
	}
	entry := tuiRuntimeInteractiveSessionEntry{SessionID: "native", ConvKey: "known-history-hash", ContextKey: "known-context-hash"}
	if err := b.saveInteractiveSession("owner", "outer", entry); err != nil {
		t.Fatal(err)
	}
	if err := b.verifyInteractiveSessionStore(); err != nil {
		t.Fatalf("valid atomic checkpoint store refused: %v", err)
	}
	path := tuiRuntimeInteractiveSessionStatePath()
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := b.verifyInteractiveSessionStore(); err == nil {
		t.Fatal("public checkpoint store silently accepted")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(t.TempDir(), "other-state")
	if err := os.WriteFile(other, []byte(`{"version":1,"sessions":{}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(other, path); err != nil {
		t.Fatal(err)
	}
	if err := b.verifyInteractiveSessionStore(); err == nil {
		t.Fatal("symlink to a foreign checkpoint store accepted")
	}
}

// The HTTP preflight is not enough: a damaged store can appear between
// preflight and a terminal checkpoint write, and internal callers can save
// without taking the HTTP path. A write must not silently replace the entire
// checkpoint map with a single new session.
func TestClaudeTUIRuntimeInteractiveCheckpointSaveNeverOverwritesCorruptOrInsecureStore(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		mode       os.FileMode
	}{
		{"truncated JSON", `{"version":`, 0o600},
		{"future version", `{"version":999,"sessions":{}}`, 0o600},
		{"null session map", `{"version":1,"sessions":null}`, 0o600},
		{"incomplete session entry", `{"version":1,"sessions":{"a":{"session_id":"native"}}}`, 0o600},
		{"publicly readable store", `{"version":1,"sessions":{}}`, 0o644},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			path := tuiRuntimeInteractiveSessionStatePath()
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(tc.body), tc.mode); err != nil {
				t.Fatal(err)
			}
			b := tuiRuntimeNewSubscriptionBridge()
			defer tuiRuntimeBridges.Delete(b)
			err := b.saveInteractiveSession("owner", "outer", tuiRuntimeInteractiveSessionEntry{
				SessionID: "native", ConvKey: "history", ContextKey: "context",
			})
			if err == nil {
				t.Fatal("corrupted state overwritten instead of rejecting save")
			}
			got, err := os.ReadFile(path)
			if err != nil || string(got) != tc.body {
				t.Fatalf("failed save replaced original checkpoint bytes: got=%q err=%v", got, err)
			}
		})
	}
}

func TestClaudeTUIRuntimeInteractiveCheckpointDirtyUpdateDoesNotReplaceDamagedStore(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer := "owner", "outer"
	if err := b.saveInteractiveSession(owner, outer, tuiRuntimeInteractiveSessionEntry{
		SessionID: "native", ConvKey: "checkpoint", ContextKey: "context",
	}); err != nil {
		t.Fatal(err)
	}
	path := tuiRuntimeInteractiveSessionStatePath()
	broken := []byte(`{"version":1,"sessions":`)
	if err := os.WriteFile(path, broken, 0o600); err != nil {
		t.Fatal(err)
	}
	b.markInteractiveSessionDirty(owner, outer)
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(broken) {
		t.Fatalf("dirty update destroyed damaged checkpoint: got=%q err=%v", got, err)
	}
}

func TestClaudeTUIRuntimeInteractiveCheckpointSaveRejectsSymlinkAndOversizedStore(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path := tuiRuntimeInteractiveSessionStatePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	entry := tuiRuntimeInteractiveSessionEntry{SessionID: "native", ConvKey: "history", ContextKey: "context"}
	target := filepath.Join(t.TempDir(), "real-state")
	data := []byte(`{"version":1,"sessions":{}}`)
	if err := os.WriteFile(target, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	if err := b.saveInteractiveSession("owner", "outer", entry); err == nil {
		t.Fatal("symlinked store accepted")
	}
	got, err := os.ReadFile(target)
	if err != nil || string(got) != string(data) {
		t.Fatalf("symlink target modified: %q %v", got, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(tuiRuntimeMaxInteractiveSessionStoreBytes + 1); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()
	if err := b.saveInteractiveSession("owner", "outer", entry); err == nil {
		t.Fatal("oversized state file accepted")
	}
	info, err := os.Stat(path)
	if err != nil || info.Size() != tuiRuntimeMaxInteractiveSessionStoreBytes+1 {
		t.Fatalf("oversized checkpoint silently replaced: %+v %v", info, err)
	}
}
