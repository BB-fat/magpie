package gateway

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A resumed turn its client cancels before Claude Code answered (16:46 on
// au-exit: two 499s, then a 46-message rebuild that wrote 44k tokens to
// the cache) is taken back to the checkpoint it resumed: the session's
// transcript as it was, the checkpoint clean. The resend resumes the same
// session with the message once, not a rebuilt one.
func TestClaudeTUIRuntimeLiteCancelledResumeRewindsToCheckpoint(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(dir, "claude"))
	logs := filepath.Join(dir, "log")
	if err := os.MkdirAll(logs, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SC_LOG", logs)
	// The first start appends its turn to the resumed session, as Claude
	// Code does, and hangs; later starts note what they resume and answer.
	script := `#!/bin/sh
cat >/dev/null
prev=""; sid=""
for a in "$@"; do [ "$prev" = "--resume" ] && sid="$a"; prev="$a"; done
if mkdir "$SC_LOG/first" 2>/dev/null; then
  for f in "$CLAUDE_CONFIG_DIR"/projects/*/"$sid".jsonl; do printf '%s\n' '{"type":"user","message":{"role":"user","content":"再看日志"}}' >> "$f"; done
  touch "$SC_LOG/appended"
  exec sleep 30
fi
echo "$sid" > "$SC_LOG/resumed"
printf '%s\n' '{"type":"assistant","uuid":"a1","session_id":"'$sid'","message":{"id":"m1","model":"claude-opus-5-5","content":[{"type":"text","text":"SC_ANSWER"}]}}'
printf '%s\n' '{"type":"result","session_id":"'$sid'","stop_reason":"end_turn"}'
`
	wrapper := filepath.Join(dir, "sc-bridge")
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	rt, b := tuiRuntimeScRuntime(t)
	tuiRuntimeScSaveClean(t, b, "inner-saved", false)
	work, err := tuiRuntimeInteractiveSessionWorkDir(tuiRuntimeScOwner, tuiRuntimeScOuter)
	if err != nil {
		t.Fatal(err)
	}
	sessions := tuiRuntimeClaudeSessions("", work)[0]
	if err := os.MkdirAll(sessions, 0o700); err != nil {
		t.Fatal(err)
	}
	transcript := filepath.Join(sessions, "inner-saved.jsonl")
	const before = `{"type":"user","message":{"role":"user","content":"先读一下配置"}}` + "\n" +
		`{"type":"assistant","message":{"id":"m0","role":"assistant","content":[{"type":"text","text":"配置已读"}]}}` + "\n"
	if err := os.WriteFile(transcript, []byte(before), 0o600); err != nil {
		t.Fatal(err)
	}
	body := `{"model":"claude-opus-5-5","max_tokens":64,"messages":[
		{"role":"user","content":"先读一下配置"},{"role":"assistant","content":"配置已读"},{"role":"user","content":"再看日志"}]}`

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		for until := time.Now().Add(10 * time.Second); time.Now().Before(until); time.Sleep(10 * time.Millisecond) {
			if _, err := os.Stat(filepath.Join(logs, "appended")); err == nil {
				break
			}
		}
		cancel()
	}()
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body)).WithContext(ctx)
	rt.serveMessages(httptest.NewRecorder(), req, "claude-opus-5-5", tuiRuntimeScOuter, []byte(body), &Usage{})
	if _, err := os.Stat(filepath.Join(logs, "appended")); err != nil {
		t.Fatal("the cancelled turn never resumed the checkpointed session")
	}

	var saved tuiRuntimeInteractiveSessionEntry
	for until := time.Now().Add(15 * time.Second); time.Now().Before(until); time.Sleep(20 * time.Millisecond) {
		if saved, _, _ = b.loadInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter); !saved.Dirty {
			break
		}
	}
	if saved.Dirty || saved.SessionID != "inner-saved" {
		t.Fatalf("the cancelled turn left its checkpoint dirty: %+v", saved)
	}
	if got, _ := os.ReadFile(transcript); string(got) != before {
		t.Fatalf("the session keeps the cancelled turn:\n%s", got)
	}

	code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, body)
	if code != 200 || !strings.Contains(out, "SC_ANSWER") {
		t.Fatalf("resend failed: %d %s", code, out)
	}
	if got, _ := os.ReadFile(filepath.Join(logs, "resumed")); strings.TrimSpace(string(got)) != "inner-saved" {
		t.Fatalf("the resend did not resume the checkpointed session: %q", got)
	}
}

// Only a turn that did nothing is taken back. One that called a tool, or
// whose session no longer begins with the checkpoint's, stays dirty and
// is rebuilt: its side effects or its history can't be undone by a cut.
func TestClaudeTUIRuntimeLiteCancelledResumeKeepsDirtyAfterAnything(t *testing.T) {
	for _, tc := range []string{"tool_called", "callback_seen", "answered", "prefix_changed"} {
		t.Run(tc, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			b := tuiRuntimeNewSubscriptionBridge()
			defer tuiRuntimeBridges.Delete(b)
			tuiRuntimeScSaveClean(t, b, "inner-saved", false)
			transcript := filepath.Join(t.TempDir(), "inner-saved.jsonl")
			if err := os.WriteFile(transcript, []byte("BASE\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			base := b.interactiveResumeBase(tuiRuntimeScOwner, tuiRuntimeScOuter, "inner-saved", []string{filepath.Dir(transcript)})
			if base == nil {
				t.Fatal("no base for a clean checkpoint")
			}
			if err := b.markInteractiveSessionDirty(tuiRuntimeScOwner, tuiRuntimeScOuter); err != nil {
				t.Fatal(err)
			}
			after := "BASE\nTURN\n"
			run := &tuiRuntimeSubscriptionRun{bridge: b, interactive: true, owner: tuiRuntimeScOwner, outerSession: tuiRuntimeScOuter, resumeBase: base}
			switch tc {
			case "tool_called":
				run.asked = []string{"exec_command"}
			case "callback_seen":
				run.seenMCPCallIDs = map[string]bool{"call": true}
			case "answered":
				run.replied = true
			case "prefix_changed":
				after = "REWRITTEN\nTURN\n"
			}
			if err := os.WriteFile(transcript, []byte(after), 0o600); err != nil {
				t.Fatal(err)
			}
			run.rewindCancelledResume()
			saved, _, _ := b.loadInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter)
			got, _ := os.ReadFile(transcript)
			if !saved.Dirty || string(got) != after {
				t.Fatalf("taken back after it did something: dirty=%t transcript=%q", saved.Dirty, got)
			}
		})
	}
}
