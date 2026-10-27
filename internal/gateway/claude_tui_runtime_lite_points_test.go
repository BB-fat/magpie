package gateway

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// cpWrapper stands in for the native PTY wrapper and keeps a real
// transcript: each start appends its user and assistant rows to the session
// it resumes (a new one otherwise) in Claude Code's project folder for its
// cwd, and answers ANS<n>, counting starts.
func tuiRuntimeCpWrapper(t *testing.T) func() []map[string]string {
	t.Helper()
	logs, runs := tuiRuntimeScWrapper(t)
	script := `#!/bin/sh
d="$SC_LOG/run-$$"
mkdir -p "$d"
printf '%s\n' "$@" > "$d/argv"
pwd > "$d/cwd"
cat > "$d/stdin"
n=$(cat "$SC_LOG/../count" 2>/dev/null || echo 0); n=$((n+1)); echo $n > "$SC_LOG/../count"
proj="$CLAUDE_CONFIG_DIR/projects/$(pwd | sed 's/[^A-Za-z0-9]/-/g')"
mkdir -p "$proj"
sid=$(cat /proc/sys/kernel/random/uuid)
prev=""
for a in "$@"; do
  if [ "$prev" = "--resume" ]; then sid="$a"; echo "$a" > "$d/resume"; cp "$proj/$a.jsonl" "$d/seed.jsonl"; fi
  prev="$a"
done
row='{"type":"assistant","uuid":"a-'$n'","sessionId":"'$sid'","session_id":"'$sid'","message":{"id":"m'$n'","model":"claude-opus-5-5","role":"assistant","content":[{"type":"text","text":"ANS'$n'"}]}}'
printf '%s\n' '{"type":"user","uuid":"u-'$n'","sessionId":"'$sid'","message":{"role":"user","content":"turn '$n'"}}' >> "$proj/$sid.jsonl"
printf '%s\n' "$row" >> "$proj/$sid.jsonl"
printf '%s\n' "$row"
printf '%s\n' '{"type":"result","session_id":"'$sid'","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(filepath.Join(filepath.Dir(logs), "sc-bridge"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return runs
}

func tuiRuntimeCpBody(msgs ...string) string {
	var parts []string
	for i, m := range msgs {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		parts = append(parts, `{"role":"`+role+`","content":"`+m+`"}`)
	}
	return `{"model":"claude-opus-5-5","max_tokens":64,"system":"rules","messages":[` + strings.Join(parts, ",") + `]}`
}

// Branches of one conversation each go on in a native session that holds
// their own history up to the reply they continue: the session itself when
// nothing was said in it since, else a copy of it up to that reply. None of
// them is rebuilt from the request, and none takes another's place.
func TestClaudeTUIRuntimeLiteBranchesGoOnFromTheirOwnReply(t *testing.T) {
	runs := tuiRuntimeCpWrapper(t)
	rt, b := tuiRuntimeScRuntime(t)
	serve := func(outer string, msgs ...string) map[string]string {
		t.Helper()
		before := len(runs())
		code, out := tuiRuntimeScServe(t, rt, outer, tuiRuntimeCpBody(msgs...))
		if code != 200 || !strings.Contains(out, "ANS") {
			t.Fatalf("%v: %d %s", msgs, code, out)
		}
		b.abortAll()
		all := runs()
		if len(all) != before+1 {
			t.Fatalf("%d native starts, want one", len(all)-before)
		}
		for _, run := range all {
			if strings.Contains(run["stdin"], msgs[len(msgs)-1]) {
				return run
			}
		}
		t.Fatalf("no run for %v", msgs)
		return nil
	}
	first := serve(tuiRuntimeScOuter, "q1")
	if first["resume"] != "" {
		t.Fatalf("first turn resumed: %v", first)
	}
	// The main line goes on in its session.
	main := serve(tuiRuntimeScOuter, "q1", "ANS1", "q2")
	s1 := main["resume"]
	if s1 == "" || strings.Contains(main["stdin"], "q1") {
		t.Fatalf("main line did not resume with only its turn: %v", main)
	}
	// A side question from the first reply: that session has gone on since,
	// so it is forked at the first reply.
	side := serve(tuiRuntimeScOuter, "q1", "ANS1", "q3")
	if side["resume"] == "" || side["resume"] == s1 || strings.Contains(side["stdin"], "q1") {
		t.Fatalf("side question did not fork: %v", side)
	}
	if !strings.Contains(side["seed.jsonl"], `"uuid":"a-1"`) || strings.Contains(side["seed.jsonl"], `"uuid":"a-2"`) ||
		!strings.Contains(side["seed.jsonl"], `"sessionId":"`+side["resume"]+`"`) {
		t.Fatalf("fork is not the session up to the first reply:\n%s", side["seed.jsonl"])
	}
	// Each line goes on in its own session, in either order.
	if again := serve(tuiRuntimeScOuter, "q1", "ANS1", "q2", "ANS2", "q4"); again["resume"] != s1 {
		t.Fatalf("main line left its session: %v", again)
	}
	if again := serve(tuiRuntimeScOuter, "q1", "ANS1", "q3", "ANS3", "q5"); again["resume"] != side["resume"] {
		t.Fatalf("side line left its session: %v", again)
	}
	// The same history under another session id (a client's fork) goes on
	// from it too, in a copy: the main line's session is not written.
	other := strings.Repeat("7", len(tuiRuntimeScOuter))
	forked := serve(other, "q1", "ANS1", "q2", "ANS2", "q6")
	if forked["resume"] == "" || forked["resume"] == s1 || forked["cwd"] != main["cwd"] {
		t.Fatalf("another session id did not fork the main line in its cwd: %v", forked)
	}
}

// A reply whose session another start holds is forked, not waited for nor
// rebuilt: branches run side by side.
func TestClaudeTUIRuntimeLiteBusySessionIsForked(t *testing.T) {
	runs := tuiRuntimeCpWrapper(t)
	rt, b := tuiRuntimeScRuntime(t)
	if code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, tuiRuntimeCpBody("q1")); code != 200 || !strings.Contains(out, "ANS1") {
		t.Fatalf("first: %d %s", code, out)
	}
	b.abortAll()
	sid := tuiRuntimeOnlyRun(t, runs())
	saved, ok, err := b.loadInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter)
	if err != nil || !ok {
		t.Fatalf("checkpoint: %v %v (%v)", ok, err, sid)
	}
	release, ok := b.claimSession(saved.SessionID)
	if !ok {
		t.Fatal("claim")
	}
	defer release()
	if code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, tuiRuntimeCpBody("q1", "ANS1", "q2")); code != 200 || !strings.Contains(out, "ANS2") {
		t.Fatalf("second: %d %s", code, out)
	}
	b.abortAll()
	for _, run := range runs() {
		if strings.Contains(run["stdin"], "q2") {
			if run["resume"] == "" || run["resume"] == saved.SessionID || !strings.Contains(run["seed.jsonl"], `"uuid":"a-1"`) {
				t.Fatalf("busy session not forked: %v", run)
			}
			return
		}
	}
	t.Fatal("no second run")
}
