package claudetui

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

func TestInteractiveRestoreRejectsStoreChangedAfterPreflight(t *testing.T) {
	for _, kind := range []string{"corrupt", "public", "oversized", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			b := newSubscriptionBridge()
			defer bridges.Delete(b)
			if err := b.saveInteractiveSession("owner", "outer", interactiveSessionEntry{
				SessionID: "native", ConvKey: "checkpoint", ContextKey: "context",
			}); err != nil {
				t.Fatal(err)
			}
			if err := b.verifyInteractiveSessionStore(); err != nil {
				t.Fatal(err)
			}
			path := interactiveSessionStatePath()
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
				if err := os.Truncate(path, maxInteractiveSessionStoreBytes+1); err != nil {
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
			_, _, status := b.restoreInteractiveSession("owner", "outer", &Request{})
			if string(status) != "state_invalid" {
				t.Fatalf("changed store classified as %s; must distinguish damage from a missing conversation", status)
			}
		})
	}
}

func TestInteractiveCheckpointOversizedUpdatePreservesPreviousStore(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	entry := interactiveSessionEntry{SessionID: "native", ConvKey: "checkpoint", ContextKey: "context"}
	if err := b.saveInteractiveSession("owner", "outer", entry); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(interactiveSessionStatePath())
	if err != nil {
		t.Fatal(err)
	}
	entry.ContextKey = strings.Repeat("x", maxInteractiveSessionStoreBytes)
	if err := b.saveInteractiveSession("owner", "outer", entry); err == nil {
		t.Fatal("an update may not create a store that its own reader rejects")
	}
	after, err := os.ReadFile(interactiveSessionStatePath())
	if err != nil || !bytes.Equal(after, before) {
		t.Fatalf("oversized update replaced previous checkpoint: %v", err)
	}
}

// Exercise the last launch boundary directly: the request preflight may have
// succeeded before the file was damaged. No child may start until its dirty
// checkpoint has actually been persisted.
func TestInteractiveResumeRequiresPersistedDirtyMarker(t *testing.T) {
	for _, kind := range []string{"corrupt", "missing", "write denied"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
			wrapper := filepath.Join(dir, "bridge")
			sentinel := filepath.Join(dir, "started")
			if err := os.WriteFile(wrapper, []byte("#!/bin/sh\nprintf started > \"$START_SENTINEL\"\ncat >/dev/null\n"), 0o700); err != nil {
				t.Fatal(err)
			}
			setWrapper(t, wrapper)
			t.Setenv("START_SENTINEL", sentinel)
			b := newSubscriptionBridge()
			defer bridges.Delete(b)
			path := interactiveSessionStatePath()
			broken := []byte(`{"version":`)
			if kind == "write denied" {
				if os.Geteuid() == 0 {
					t.Skip("root bypasses directory write permissions")
				}
				if err := b.saveInteractiveSession("boundary-owner", "boundary-"+kind, interactiveSessionEntry{SessionID: "native", ConvKey: "checkpoint", ContextKey: "context"}); err != nil {
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
			req := &Request{Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "continue"}}}}}
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

func TestInteractiveLateStoreFailureReturnsConflict(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	b := newSubscriptionBridge()
	defer bridges.Delete(b)
	if err := b.saveInteractiveSession("owner", "outer", interactiveSessionEntry{SessionID: "native", ConvKey: "checkpoint", ContextKey: "context"}); err != nil {
		t.Fatal(err)
	}
	rt := &runtimeTUI{bridge: b}
	body := []byte(`{"model":"claude-opus-5-5","messages":[{"role":"user","content":"continue"}]}`)
	started := 0
	start := func(ctx context.Context, req *Request) (*subscriptionRun, <-chan Event, error) {
		started++
		if err := os.WriteFile(interactiveSessionStatePath(), []byte(`{"version":`), 0o600); err != nil {
			t.Fatal(err)
		}
		return b.startInteractiveRequest(ctx, req, "claude-opus-5-5", "", "owner", "outer")
	}
	w := httptest.NewRecorder()
	code, _ := rt.serveMessagesInner(anthropicResponder{w, httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body)), &Usage{}}, httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body)), "Claude", "claude-opus-5-5", "owner", "outer", body, start)
	if started != 1 || code != http.StatusConflict {
		t.Fatalf("late integrity failure: callback=%d HTTP=%d; want one callback and 409", started, code)
	}
}

func TestRestoreUnchangedFailedCompanionNeverReplays(t *testing.T) {
	base := []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "task"}}}, {Role: "assistant", Parts: []Part{{Kind: Text, Text: "answer"}}}}
	companion := Message{Role: "user", Parts: []Part{{Kind: Text, Text: "Your response above was cut off mid-stream. Continue."}}}
	failed := Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: "API Error: 502 Claude interactive persisted resume failed: suffix_rejected; refusing full history replay"}}}
	for _, msgs := range [][]Message{append(append([]Message{}, base...), companion, failed), append(append([]Message{}, base...), companion, failed, companion, failed)} {
		for _, replyAnchor := range []bool{false, true} {
			var since []Message
			var status conversationSuffixStatus
			if replyAnchor {
				since, status = conversationSuffixAfterReplyDetailed(msgs, assistantReplyKey(base[1]))
			} else {
				since, status = conversationSuffixAfterDetailed("owner", msgs, conversationKeys("owner", base)[0])
			}
			if status != conversationSuffixRejected || len(since) != 0 {
				t.Fatalf("unchanged retry replayed: %v %#v", status, since)
			}
		}
	}
}
