package gateway

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const tuiRuntimeErTools = `"system":"STEER_INSTRUCTIONS","tools":[{"name":"write_file","input_schema":{"type":"object"}}]`

// erRound is the run's first tool round as the client keeps it: the user's
// ask, the reply calling write_file, and its result.
const tuiRuntimeErRound = `{"role":"user","content":"把结果写进 notes.txt"},
	{"role":"assistant","content":[{"type":"tool_use","id":"call_1","name":"write_file","input":{"path":"a"}}]}`

const tuiRuntimeErResult = `{"type":"tool_result","tool_use_id":"call_1","content":"WROTE_OK"}`

// erNative is the run's session as Claude Code saved it: the round's call
// (its input's keys in Claude Code's order) and native result, then the
// call it went on to make, which the client dropped.
const tuiRuntimeErNative = `{"type":"user","uuid":"u1","parentUuid":null,"sessionId":"inner-steer","message":{"role":"user","content":"<external_system_instructions>\nSTEER_INSTRUCTIONS\n</external_system_instructions>\n\n把结果写进 notes.txt"}}
{"type":"assistant","uuid":"t0","parentUuid":"u1","sessionId":"inner-steer","message":{"id":"m1","role":"assistant","content":[{"type":"thinking","thinking":"","signature":"SIG"}]}}
{"type":"assistant","uuid":"t1","parentUuid":"t0","sessionId":"inner-steer","message":{"id":"m1","role":"assistant","content":[{"type":"tool_use","id":"call_1","name":"mcp__magpie__write_file","input":{"zeta":1,"path":"a"}}]}}
{"type":"user","uuid":"r1","parentUuid":"t1","sessionId":"inner-steer","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"call_1","content":"WROTE_OK"}]}}
{"type":"attachment","uuid":"x1","parentUuid":"r1","sessionId":"inner-steer"}
{"type":"assistant","uuid":"t2","parentUuid":"x1","sessionId":"inner-steer","message":{"id":"m2","role":"assistant","content":[{"type":"tool_use","id":"call_2","name":"mcp__magpie__write_file","input":{"path":"b"}}]}}
`

func tuiRuntimeErParse(t *testing.T, messages string) *tuiRuntimeRequest {
	t.Helper()
	req, err := tuiRuntimeParseRequest([]byte(`{"model":"claude-opus-5-5","max_tokens":64,` + tuiRuntimeErTools + `,"messages":[` + messages + `]}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Model = "claude-opus-5-5"
	return req
}

// erParkedPast stands in for a conversation's run that parked on call_1,
// was told its result, and went on to park on call_2: its session saved
// as native says, under the conversation's cwd.
func tuiRuntimeErParkedPast(t *testing.T, b *tuiRuntimeSubscriptionBridge, native string) (*tuiRuntimeSubscriptionRun, string) {
	t.Helper()
	work := tuiRuntimeInteractiveSessionStateKey(tuiRuntimeScOwner, tuiRuntimeScOuter)
	dir, err := tuiRuntimeInteractiveSessionWorkDir(tuiRuntimeScOwner, tuiRuntimeScOuter)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	sessions := tuiRuntimeClaudeSessions("", dir)[0]
	if err := os.MkdirAll(sessions, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(sessions, "inner-steer.jsonl")
	if err := os.WriteFile(path, []byte(native), 0o600); err != nil {
		t.Fatal(err)
	}
	told := tuiRuntimeErParse(t, tuiRuntimeErRound+`,{"role":"user","content":[`+tuiRuntimeErResult+`]}`)
	run := &tuiRuntimeSubscriptionRun{interactive: true, owner: tuiRuntimeScOwner, outerSession: tuiRuntimeScOuter, model: "claude-opus-5-5",
		sessionID: "inner-steer", workKey: work, closed: true}
	run.interactiveContextSnapshot = tuiRuntimeTestContextSnapshot(told)
	run.shown, run.parkedAt = []tuiRuntimeShownCall{{id: "call_1"}}, time.Now()
	b.noteLiteParked(run)
	run.heard(told.Messages)
	// It goes on (continueWith clears its park) while another
	// conversation's run parks: the round is kept all the same.
	run.shown, run.parkedAt = nil, time.Time{}
	b.noteLiteParked(&tuiRuntimeSubscriptionRun{interactive: true, owner: tuiRuntimeScOwner, outerSession: "another",
		shown: []tuiRuntimeShownCall{{id: "call_x"}}, parkedAt: time.Now()})
	run.shown, run.parkedAt = []tuiRuntimeShownCall{{id: "call_2"}}, time.Now()
	b.noteLiteParked(run)
	return run, path
}

// A user who steers mid-turn: the client kept the run's first tool round,
// its call and result, and sent what the user said, dropping the call the
// run went on to make. The turn resumes a copy of the run's session up to
// that round — the rows the model saw byte for byte, a prefix Anthropic
// has cached — where it was rebuilt from the request, the history written
// to the cache again (au-exit 10-28). The run is left waiting on its call.
func TestClaudeTUIRuntimeLiteSteerForksAtTheRoundTheClientKept(t *testing.T) {
	_, runs := tuiRuntimeScWrapper(t)
	rt, b := tuiRuntimeScRuntime(t)
	run, native := tuiRuntimeErParkedPast(t, b, tuiRuntimeErNative)
	code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, `{"model":"claude-opus-5-5","max_tokens":64,`+tuiRuntimeErTools+`,"messages":[`+
		tuiRuntimeErRound+`,{"role":"user","content":[`+tuiRuntimeErResult+`,{"type":"text","text":"换成 b.txt 再写"}]}]}`)
	if code != 200 || !strings.Contains(out, "SC_ANSWER") {
		t.Fatalf("steered turn failed: %d %s", code, out)
	}
	got := tuiRuntimeOnlyRun(t, runs())
	if got["resume"] == "" || got["resume"] == "inner-steer" {
		t.Fatalf("did not resume a copy of the run's session: %v", got)
	}
	copied := got["seed.jsonl"]
	for _, kept := range []string{"STEER_INSTRUCTIONS", `"signature":"SIG"`, `"input":{"zeta":1,"path":"a"}`, `"sessionId":"` + got["resume"] + `"`, "WROTE_OK", tuiRuntimeLiteResultsReceived} {
		if !strings.Contains(copied, kept) {
			t.Fatalf("resumed session lacks %s:\n%s", kept, copied)
		}
	}
	for _, gone := range []string{"call_2", `"uuid":"r1"`, `"uuid":"x1"`, `"sessionId":"inner-steer"`} {
		if strings.Contains(copied, gone) {
			t.Fatalf("resumed session keeps %s:\n%s", gone, copied)
		}
	}
	if !strings.Contains(got["stdin"], "换成 b.txt 再写") || strings.Contains(got["stdin"], tuiRuntimeLiteContinueText) || strings.Contains(got["stdin"], "STEER_INSTRUCTIONS") {
		t.Fatalf("turn prompt: %q", got["stdin"])
	}
	if data, err := os.ReadFile(native); err != nil || string(data) != tuiRuntimeErNative {
		t.Fatalf("the run's own session was changed: %v", err)
	}
	b.mu.Lock()
	still := b.liteParked["call_2"] == run
	b.mu.Unlock()
	if !still {
		t.Fatal("the run's own call was let go")
	}
}

// A copy of the run's session is taken up only for exactly the history it
// was told, with its context and mode, a round whose calls are all the
// request's last reply made, and a reply that made no other call.
func TestClaudeTUIRuntimeLiteEarlierRoundNeedsTheRunsOwnRound(t *testing.T) {
	tuiRuntimeScWrapper(t)
	steer := func(round, result string) *tuiRuntimeRequest {
		return tuiRuntimeErParse(t, round+`,{"role":"user","content":[`+result+`,{"type":"text","text":"换成 b.txt 再写"}]}`)
	}
	held := strings.Replace(tuiRuntimeErNative, `"input":{"zeta":1,"path":"a"}}]`,
		`"input":{"zeta":1,"path":"a"}},{"type":"tool_use","id":"call_held","name":"mcp__magpie__write_file","input":{}}]`, 1)
	for _, c := range []struct {
		name, native string
		model        string
		req          func() *tuiRuntimeRequest
		ok           bool
	}{
		{"the round it was told", tuiRuntimeErNative, "claude-opus-5-5", func() *tuiRuntimeRequest { return steer(tuiRuntimeErRound, tuiRuntimeErResult) }, true},
		{"another result than it was told", tuiRuntimeErNative, "claude-opus-5-5", func() *tuiRuntimeRequest {
			return steer(tuiRuntimeErRound, strings.Replace(tuiRuntimeErResult, "WROTE_OK", "WROTE_ELSEWHERE", 1))
		}, false},
		{"a reply with a call the round lacks", tuiRuntimeErNative, "claude-opus-5-5", func() *tuiRuntimeRequest {
			round := strings.Replace(tuiRuntimeErRound, `"input":{"path":"a"}}]`, `"input":{"path":"a"}},{"type":"tool_use","id":"call_1b","name":"write_file","input":{}}]`, 1)
			return steer(round, tuiRuntimeErResult+`,{"type":"tool_result","tool_use_id":"call_1b","content":"x"}`)
		}, false},
		{"another model", tuiRuntimeErNative, "claude-sonnet-5-5", func() *tuiRuntimeRequest { return steer(tuiRuntimeErRound, tuiRuntimeErResult) }, false},
		{"the other mode", tuiRuntimeErNative, "claude-opus-5-5", func() *tuiRuntimeRequest {
			req := steer(tuiRuntimeErRound, tuiRuntimeErResult)
			req.leanPrompt = true
			return req
		}, false},
		{"a native reply with a call the client never saw", held, "claude-opus-5-5", func() *tuiRuntimeRequest { return steer(tuiRuntimeErRound, tuiRuntimeErResult) }, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := tuiRuntimeNewSubscriptionBridge()
			t.Cleanup(func() { tuiRuntimeBridges.Delete(b) })
			tuiRuntimeErParkedPast(t, b, c.native)
			if _, _, _, _, ok := b.liteEarlierRound(tuiRuntimeScOwner, c.model, "", c.req()); ok != c.ok {
				t.Fatalf("forked: %t, want %t", ok, c.ok)
			}
		})
	}
	// still parked on the round's own calls: liteParkedBranch's to take
	b := tuiRuntimeNewSubscriptionBridge()
	t.Cleanup(func() { tuiRuntimeBridges.Delete(b) })
	run, _ := tuiRuntimeErParkedPast(t, b, tuiRuntimeErNative)
	run.shown = []tuiRuntimeShownCall{{id: "call_1"}}
	if _, _, _, _, ok := b.liteEarlierRound(tuiRuntimeScOwner, "claude-opus-5-5", "", steer(tuiRuntimeErRound, tuiRuntimeErResult)); ok {
		t.Fatal("forked a run still waiting on the round's calls")
	}
}

// A round of two calls whose results came one at a time: the run was told
// the history with the first result alone, but a copy of its session with
// the second call unanswered is not resumed.
func TestClaudeTUIRuntimeLiteEarlierRoundNeedsAllTheRoundsResults(t *testing.T) {
	tuiRuntimeScWrapper(t)
	b := tuiRuntimeNewSubscriptionBridge()
	t.Cleanup(func() { tuiRuntimeBridges.Delete(b) })
	two := strings.Replace(tuiRuntimeErRound, `"input":{"path":"a"}}]`, `"input":{"path":"a"}},{"type":"tool_use","id":"call_1b","name":"write_file","input":{}}]`, 1)
	native := strings.Replace(tuiRuntimeErNative, `"input":{"zeta":1,"path":"a"}}]`,
		`"input":{"zeta":1,"path":"a"}},{"type":"tool_use","id":"call_1b","name":"mcp__magpie__write_file","input":{}}]`, 1)
	run, _ := tuiRuntimeErParkedPast(t, b, native)
	told := tuiRuntimeErParse(t, two+`,{"role":"user","content":[`+tuiRuntimeErResult+`]}`)
	run.interactiveContextSnapshot = tuiRuntimeTestContextSnapshot(told)
	run.shown, run.parkedAt = []tuiRuntimeShownCall{{id: "call_1"}, {id: "call_1b"}}, time.Now()
	b.noteLiteParked(run)
	run.heard(told.Messages)
	run.shown = []tuiRuntimeShownCall{{id: "call_2"}}
	steer := tuiRuntimeErParse(t, two+`,{"role":"user","content":[`+tuiRuntimeErResult+`,{"type":"text","text":"换成 b.txt 再写"}]}`)
	if _, _, _, _, ok := b.liteEarlierRound(tuiRuntimeScOwner, "claude-opus-5-5", "", steer); ok {
		t.Fatal("forked a round with one of its two calls unanswered")
	}
}
