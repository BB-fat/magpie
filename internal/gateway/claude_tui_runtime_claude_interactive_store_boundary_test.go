package gateway

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeTUIRuntimeInteractiveRestoreRejectsStoreChangedAfterPreflight(t *testing.T) {
	for _, kind := range []string{"corrupt", "public", "oversized", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			b := tuiRuntimeNewSubscriptionBridge()
			defer tuiRuntimeBridges.Delete(b)
			if err := b.saveInteractiveSession("owner", "outer", tuiRuntimeInteractiveSessionEntry{
				SessionID: "native", ConvKey: "checkpoint", InstructionKey: "context",
			}); err != nil {
				t.Fatal(err)
			}
			if err := b.verifyInteractiveSessionStore(); err != nil {
				t.Fatal(err)
			}
			path := tuiRuntimeInteractiveSessionStatePath()
			switch kind {
			case "corrupt":
				if err := os.WriteFile(path, []byte(`{"version":`), 0o600); err != nil {
					t.Fatal(err)
				}
			case "public":
				if err := os.Chmod(path, 0o644); err != nil {
					t.Fatal(err)
				}
			case "oversized":
				if err := os.Truncate(path, tuiRuntimeMaxInteractiveSessionStoreBytes+1); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				target := filepath.Join(t.TempDir(), "store")
				if err := os.Rename(path, target); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, path); err != nil {
					t.Fatal(err)
				}
			}
			_, _, status := b.restoreInteractiveSession("owner", "outer", &tuiRuntimeRequest{Request: Request{}})
			if string(status) != "state_invalid" {
				t.Fatalf("changed store classified as %s; must distinguish damage from a missing conversation", status)
			}
		})
	}
}

func TestClaudeTUIRuntimeInteractiveCheckpointOversizedUpdatePreservesPreviousStore(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	entry := tuiRuntimeInteractiveSessionEntry{SessionID: "native", ConvKey: "checkpoint", InstructionKey: "context"}
	if err := b.saveInteractiveSession("owner", "outer", entry); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(tuiRuntimeInteractiveSessionStatePath())
	if err != nil {
		t.Fatal(err)
	}
	entry.InstructionKey = strings.Repeat("x", tuiRuntimeMaxInteractiveSessionStoreBytes)
	if err := b.saveInteractiveSession("owner", "outer", entry); err == nil {
		t.Fatal("an update may not create a store that its own reader rejects")
	}
	after, err := os.ReadFile(tuiRuntimeInteractiveSessionStatePath())
	if err != nil || !bytes.Equal(after, before) {
		t.Fatalf("oversized update replaced previous checkpoint: %v", err)
	}
}

// Exercise the last launch boundary directly: the request preflight may have
// succeeded before the file was damaged. No child may start until its dirty
// checkpoint has actually been persisted.
func TestClaudeTUIRuntimeInteractiveResumeRequiresPersistedDirtyMarker(t *testing.T) {
	for _, kind := range []string{"corrupt", "missing", "write denied"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
			wrapper := filepath.Join(dir, "bridge")
			sentinel := filepath.Join(dir, "started")
			if err := os.WriteFile(wrapper, []byte("#!/bin/sh\nprintf started > \"$START_SENTINEL\"\ncat >/dev/null\n"), 0o700); err != nil {
				t.Fatal(err)
			}
			tuiRuntimeSetWrapper(t, wrapper)
			t.Setenv("START_SENTINEL", sentinel)
			b := tuiRuntimeNewSubscriptionBridge()
			defer tuiRuntimeBridges.Delete(b)
			path := tuiRuntimeInteractiveSessionStatePath()
			broken := []byte(`{"version":`)
			if kind == "write denied" {
				if os.Geteuid() == 0 {
					t.Skip("root bypasses directory write permissions")
				}
				if err := b.saveInteractiveSession("boundary-owner", "boundary-"+kind, tuiRuntimeInteractiveSessionEntry{SessionID: "native", ConvKey: "checkpoint", InstructionKey: "context"}); err != nil {
					t.Fatal(err)
				}
				parent := filepath.Dir(path)
				if err := os.Chmod(parent, 0o500); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(parent, 0o700) })
			}
			if kind == "corrupt" {
				if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, broken, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			req := &tuiRuntimeRequest{Request: Request{Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "continue"}}}}}}
			run, _, err := b.startInteractiveWithPrompt(context.Background(), req, "claude-opus-5-5", "", "boundary-owner", "boundary-"+kind, "continue", "native", false)
			if run != nil {
				<-run.interactiveDone
				run.abort()
			}
			if err == nil {
				t.Error("native resume admitted without a durable dirty checkpoint")
			}
			if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
				t.Errorf("native child executed before dirty state persisted: %v", err)
			}
			if kind == "corrupt" {
				got, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(got, broken) {
					t.Fatalf("damaged store changed: %q %v", got, err)
				}
			}
		})
	}
}

func TestClaudeTUIRuntimeInteractiveLateStoreFailureReturnsConflict(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	if err := b.saveInteractiveSession("owner", "outer", tuiRuntimeInteractiveSessionEntry{SessionID: "native", ConvKey: "checkpoint", InstructionKey: "context"}); err != nil {
		t.Fatal(err)
	}
	rt := &tuiRuntime{bridge: b}
	body := []byte(`{"model":"claude-opus-5-5","messages":[{"role":"user","content":"continue"}]}`)
	started := 0
	start := func(ctx context.Context, req *tuiRuntimeRequest) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
		started++
		if err := os.WriteFile(tuiRuntimeInteractiveSessionStatePath(), []byte(`{"version":`), 0o600); err != nil {
			t.Fatal(err)
		}
		return b.startInteractiveRequest(ctx, req, "claude-opus-5-5", "", "owner", "outer")
	}
	w := httptest.NewRecorder()
	code, _ := rt.serveMessagesInner(tuiRuntimeAnthropicResponder{w, httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body)), &Usage{}}, httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body)), "Claude", "claude-opus-5-5", "owner", "outer", body, start)
	if started != 1 || code != http.StatusConflict {
		t.Fatalf("late integrity failure: callback=%d HTTP=%d; want one callback and 409", started, code)
	}
}

func TestClaudeTUIRuntimeRestoreUnchangedFailedCompanionNeverReplays(t *testing.T) {
	base := []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "task"}}}, {Role: "assistant", Parts: []Part{{Kind: Text, Text: "answer"}}}}
	companion := Message{Role: "user", Parts: []Part{{Kind: Text, Text: "Your response above was cut off mid-stream. Continue."}}}
	failed := Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: "API Error: 502 Claude interactive persisted resume failed: suffix_rejected; refusing full history replay"}}}
	for _, msgs := range [][]Message{append(append([]Message{}, base...), companion, failed), append(append([]Message{}, base...), companion, failed, companion, failed)} {
		for _, replyAnchor := range []bool{false, true} {
			var since []Message
			var status tuiRuntimeConversationSuffixStatus
			if replyAnchor {
				since, status = tuiRuntimeConversationSuffixAfterReplyDetailed(msgs, tuiRuntimeAssistantReplyKey(base[1]))
			} else {
				since, status = tuiRuntimeConversationSuffixAfterDetailed("owner", msgs, tuiRuntimeConversationKeys("owner", base)[0])
			}
			if status != tuiRuntimeConversationSuffixRejected || len(since) != 0 {
				t.Fatalf("unchanged retry replayed: %v %#v", status, since)
			}
		}
	}
}
