package gateway

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestClaudeTUIRuntimeRealServe keeps the real Claude Code for its
// conversation's next turn, against the local stand-in for Anthropic: the
// next turn is typed into the same process, nothing is asked of Anthropic
// while it waits, and once it is gone the next turn is still answered from
// the conversation's native session. Run as TestClaudeTUIRuntimeRealClaudeTUI
// is.
func TestClaudeTUIRuntimeRealServe(t *testing.T) {
	if os.Getenv("MAGPIE_CLAUDE_TUI_REAL") != "1" {
		t.Skip("set MAGPIE_CLAUDE_TUI_REAL=1 and the programs to run the real Claude Code")
	}
	t.Run("next turn", tuiRuntimeRealServeNextTurn)
	t.Run("tool round", tuiRuntimeRealServeToolRound)
	t.Run("after the TTL", tuiRuntimeRealServeAfterTTL)
	t.Run("no header after the TTL", tuiRuntimeRealServeNoHeaderAfterTTL)
	t.Run("cancelled", tuiRuntimeRealServeCancelled)
}

// tuiRuntimeRealServeSetup is realSetup with processes kept ttl.
func tuiRuntimeRealServeSetup(t *testing.T, ttl time.Duration) (*tuiRuntimeStandIn, string, *tuiRuntimeBridge) {
	up, profile, bridge := tuiRuntimeRealSetup(t, "magpie")
	initTUIRuntime(&tuiRuntimeConfig{IdleTTL: func() time.Duration { return ttl }})
	return up, profile, bridge()
}

// tuiRuntimeRealSession is the native session Claude Code named in the
// upstream request raw.
func tuiRuntimeRealSession(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var req struct {
		Metadata struct {
			UserID string `json:"user_id"`
		} `json:"metadata"`
	}
	var user struct {
		SessionID string `json:"session_id"`
	}
	if json.Unmarshal(raw, &req) != nil || json.Unmarshal([]byte(req.Metadata.UserID), &user) != nil || user.SessionID == "" {
		t.Fatalf("no session in the upstream request's metadata: %.300s", raw)
	}
	return user.SessionID
}

// tuiRuntimeRealSameProcesses fails unless every process kept is still
// running and the Claude Code processes are the kept ones: a hook started
// for the turn may be running beside them.
func tuiRuntimeRealSameProcesses(t *testing.T, profile string, kept []int) {
	t.Helper()
	after := tuiRuntimeUnderProfile(profile)
	for _, pid := range kept {
		if !slices.Contains(after, pid) {
			t.Fatalf("kept process %d is gone: the turn ran in %v", pid, after)
		}
	}
	claude := func(pids []int) []int {
		var out []int
		for _, pid := range pids {
			cmd, _ := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/cmdline")
			if exe, _, _ := strings.Cut(string(cmd), "\x00"); exe == os.Getenv("MAGPIE_TUI_CLAUDE") {
				out = append(out, pid)
			}
		}
		return out
	}
	if was, now := claude(kept), claude(after); len(was) == 0 || !reflect.DeepEqual(was, now) {
		t.Fatalf("Claude Code processes %v after the turn, kept %v", now, was)
	}
}

// tuiRuntimeRealUnmarked is messages without cache marks, a content of
// one text block as its text.
func tuiRuntimeRealUnmarked(messages []any) []any {
	var visit func(any)
	visit = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			delete(x, "cache_control")
			for _, item := range x {
				visit(item)
			}
			if blocks, ok := x["content"].([]any); ok && len(blocks) == 1 {
				if block, ok := blocks[0].(map[string]any); ok && len(block) == 2 && block["type"] == "text" {
					x["content"] = block["text"]
				}
			}
		case []any:
			for _, item := range x {
				visit(item)
			}
		}
	}
	visit(messages)
	return messages
}

func (s *tuiRuntimeStandIn) all() []json.RawMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.requests)
}

// tuiRuntimeRealMain is the last upstream request holding text: the
// turn's own, not one of Claude Code's side asks.
func tuiRuntimeRealMain(t *testing.T, rows []json.RawMessage, text string) json.RawMessage {
	t.Helper()
	for i := len(rows) - 1; i >= 0; i-- {
		if strings.Contains(string(rows[i]), text) {
			return rows[i]
		}
	}
	t.Fatalf("no upstream request held %q", text)
	return nil
}

func tuiRuntimeRealServeNextTurn(t *testing.T) {
	up, profile, b := tuiRuntimeRealServeSetup(t, time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	first := `{"model":"` + tuiRuntimeRealModel + `","max_tokens":64,"stream":true,"messages":[{"role":"user","content":"say hello"}]}`
	if out, code, msg := tuiRuntimeRealServe(t, b, ctx, profile, "", first); code != 200 || out.text() != "PONG" {
		t.Fatalf("first turn: %d %s %q", code, msg, out.text())
	}
	tuiRuntimeServeIdle(t, b, 1)
	pids := tuiRuntimeUnderProfile(profile)
	if len(pids) == 0 {
		t.Fatal("no process kept for the next turn")
	}
	// nothing is asked of Anthropic while the process waits
	asked := len(up.all())
	time.Sleep(3 * time.Second)
	if n := len(up.all()); n != asked {
		t.Fatalf("the idle process asked Anthropic %d more times", n-asked)
	}
	before := up.all()
	next := `{"model":"` + tuiRuntimeRealModel + `","max_tokens":64,"stream":true,"messages":[{"role":"user","content":"say hello"},{"role":"assistant","content":"PONG"},{"role":"user","content":"and again"}]}`
	if out, code, msg := tuiRuntimeRealServe(t, b, ctx, profile, "", next); code != 200 || out.text() != "PONG" {
		t.Fatalf("next turn: %d %s %q", code, msg, out.text())
	}
	tuiRuntimeRealSameProcesses(t, profile, pids)
	old := tuiRuntimeRealMain(t, before, "say hello")
	cur := tuiRuntimeRealMain(t, up.all()[len(before):], "and again")
	if tuiRuntimeRealSession(t, old) != tuiRuntimeRealSession(t, cur) {
		t.Fatal("the next turn was asked in another native session")
	}
	// cache marks move to the newest message, and one text block is its
	// text, as in the cache prefix test
	var o, c struct{ Messages []any }
	_ = json.Unmarshal(old, &o)
	_ = json.Unmarshal(cur, &c)
	was, now := tuiRuntimeRealUnmarked(o.Messages), tuiRuntimeRealUnmarked(c.Messages)
	if len(now) <= len(was) || !reflect.DeepEqual(was, now[:len(was)]) {
		t.Fatalf("the next turn's upstream messages don't go on from the first's:\nbefore=%v\nafter=%v", was, now)
	}
	b.Shutdown()
	tuiRuntimeWaitGone(t, b, profile)
}

// A turn that went through a tool round ends kept too, at the reply after
// its results, and the next turn goes on there.
func tuiRuntimeRealServeToolRound(t *testing.T) {
	_, profile, b := tuiRuntimeRealServeSetup(t, time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	first := `{"model":"` + tuiRuntimeRealModel + `","max_tokens":256,"stream":true,"tools":` + tuiRuntimeRealTools + `,"messages":[{"role":"user","content":"call echo with hi"}]}`
	out, code, msg := tuiRuntimeRealServe(t, b, ctx, profile, "", first)
	call := out.tool()
	if code != 200 || call.name != "echo" {
		t.Fatalf("first turn: %d %s %+v", code, msg, call)
	}
	pids := tuiRuntimeUnderProfile(profile)
	history := `{"role":"user","content":"call echo with hi"},` +
		`{"role":"assistant","content":[{"type":"tool_use","id":` + tuiRuntimeJsonString(call.id) + `,"name":"echo","input":` + call.args + `}]},` +
		`{"role":"user","content":[{"type":"tool_result","tool_use_id":` + tuiRuntimeJsonString(call.id) + `,"content":"ECHOED"}]}`
	results := `{"model":"` + tuiRuntimeRealModel + `","max_tokens":256,"stream":true,"tools":` + tuiRuntimeRealTools + `,"messages":[` + history + `]}`
	out, code, msg = tuiRuntimeRealServe(t, b, ctx, profile, "", results)
	if code != 200 || out.text() != "DONE: ECHOED" {
		t.Fatalf("tool results: %d %s %q", code, msg, out.text())
	}
	tuiRuntimeServeIdle(t, b, 1)
	next := `{"model":"` + tuiRuntimeRealModel + `","max_tokens":256,"stream":true,"tools":` + tuiRuntimeRealTools + `,"messages":[` + history +
		`,{"role":"assistant","content":"DONE: ECHOED"},{"role":"user","content":"and again"}]}`
	if out, code, msg := tuiRuntimeRealServe(t, b, ctx, profile, "", next); code != 200 || out.text() == "" {
		t.Fatalf("next turn: %d %s %q", code, msg, out.text())
	}
	tuiRuntimeRealSameProcesses(t, profile, pids)
	b.Shutdown()
	tuiRuntimeWaitGone(t, b, profile)
}

// Once the TTL ended the process, the conversation's next turn resumes its
// native session from the cut point of its reply, in a new process.
func tuiRuntimeRealServeAfterTTL(t *testing.T) {
	up, profile, b := tuiRuntimeRealServeSetup(t, 2*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	first := `{"model":"` + tuiRuntimeRealModel + `","max_tokens":64,"stream":true,"messages":[{"role":"user","content":"say hello"}]}`
	if out, code, msg := tuiRuntimeRealServe(t, b, ctx, profile, "conv-ttl", first); code != 200 || out.text() != "PONG" {
		t.Fatalf("first turn: %d %s %q", code, msg, out.text())
	}
	pids := tuiRuntimeUnderProfile(profile)
	tuiRuntimeWaitGone(t, b, profile) // the TTL ends it
	before := up.all()
	next := `{"model":"` + tuiRuntimeRealModel + `","max_tokens":64,"stream":true,"messages":[{"role":"user","content":"say hello"},{"role":"assistant","content":"PONG"},{"role":"user","content":"and again"}]}`
	if out, code, msg := tuiRuntimeRealServe(t, b, ctx, profile, "conv-ttl", next); code != 200 || out.text() != "PONG" {
		t.Fatalf("after the TTL: %d %s %q", code, msg, out.text())
	}
	if after := tuiRuntimeUnderProfile(profile); slices.ContainsFunc(after, func(pid int) bool { return slices.Contains(pids, pid) }) {
		t.Fatalf("a process of the first turn %v ran the turn after its TTL %v", pids, after)
	}
	old := tuiRuntimeRealMain(t, before, "say hello")
	cur := tuiRuntimeRealMain(t, up.all()[len(before):], "and again")
	if tuiRuntimeRealSession(t, old) != tuiRuntimeRealSession(t, cur) {
		t.Fatal("the turn after the TTL was rebuilt in a new native session, not resumed")
	}
	b.Shutdown()
	tuiRuntimeWaitGone(t, b, profile)
}

// A conversation with no session header (outer ""): once the TTL ends its
// process, the next turn still resumes its native session from the
// strict-prefix checkpoint — the same upstream session id and message
// prefix, not a rebuild of the whole history. This is the au-exit case:
// consecutive requests without a session header were each rebuilt.
func tuiRuntimeRealServeNoHeaderAfterTTL(t *testing.T) {
	up, profile, b := tuiRuntimeRealServeSetup(t, 2*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	first := `{"model":"` + tuiRuntimeRealModel + `","max_tokens":64,"stream":true,"messages":[{"role":"user","content":"say hello"}]}`
	if out, code, msg := tuiRuntimeRealServe(t, b, ctx, profile, "", first); code != 200 || out.text() != "PONG" {
		t.Fatalf("first turn: %d %s %q", code, msg, out.text())
	}
	tuiRuntimeWaitGone(t, b, profile) // the TTL ends the process
	before := up.all()
	next := `{"model":"` + tuiRuntimeRealModel + `","max_tokens":64,"stream":true,"messages":[{"role":"user","content":"say hello"},{"role":"assistant","content":"PONG"},{"role":"user","content":"and again"}]}`
	if out, code, msg := tuiRuntimeRealServe(t, b, ctx, profile, "", next); code != 200 || out.text() != "PONG" {
		t.Fatalf("after the TTL: %d %s %q", code, msg, out.text())
	}
	old := tuiRuntimeRealMain(t, before, "say hello")
	cur := tuiRuntimeRealMain(t, up.all()[len(before):], "and again")
	if tuiRuntimeRealSession(t, old) != tuiRuntimeRealSession(t, cur) {
		t.Fatal("a no-header conversation was rebuilt in a new native session after the TTL, not resumed from its checkpoint")
	}
	// the resumed request goes on from the first's messages
	var o, c struct{ Messages []any }
	_ = json.Unmarshal(old, &o)
	_ = json.Unmarshal(cur, &c)
	was, now := tuiRuntimeRealUnmarked(o.Messages), tuiRuntimeRealUnmarked(c.Messages)
	if len(now) <= len(was) || !reflect.DeepEqual(was, now[:len(was)]) {
		t.Fatalf("the no-header turn after the TTL did not go on from the first's prefix:\nbefore=%v\nafter=%v", was, now)
	}
	b.Shutdown()
	tuiRuntimeWaitGone(t, b, profile)
}

// A turn whose client goes is never kept: its process ends with it.
func tuiRuntimeRealServeCancelled(t *testing.T) {
	up, profile, b := tuiRuntimeRealServeSetup(t, time.Minute)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		tuiRuntimeRealServe(t, b, ctx, profile, "", `{"model":"`+tuiRuntimeRealModel+`","max_tokens":64,"stream":true,"messages":[{"role":"user","content":"take your time"}]}`)
	}()
	select {
	case <-up.slow:
	case <-time.After(2 * time.Minute):
		t.Fatal("Claude Code never asked the stand-in")
	}
	cancel()
	<-done
	tuiRuntimeWaitGone(t, b, profile)
	if n := tuiRuntimeServeIdleCount(b); n != 0 {
		t.Fatalf("%d processes kept after a cancelled turn", n)
	}
}
