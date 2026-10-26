package claudetui

import (
	"context"
	"strings"
	"sync"
	"testing"
)

// A tool-result turn whose reply was delivered is replayed only for the very
// same request. The same history under another system prompt, other tools or
// other generation settings is another question and must not get the old
// answer (v0.2.12 compared the messages alone).

const replayHistory = `"messages":[{"role":"user","content":"run"},{"role":"assistant","content":[{"type":"tool_use","id":"call-1","name":"echo","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call-1","content":"finished"}]}]`

func replayBody(extra string) string {
	return `{"model":"claude-opus-5-5","max_tokens":64,"system":"S1","tools":[{"name":"echo","input_schema":{"type":"object"}}],` + extra + replayHistory + `}`
}

func replayReq(t *testing.T, body string) *Request {
	t.Helper()
	req, err := parseAnthropic([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Model = "claude-opus-5-5"
	return req
}

func TestLiteCompletedReplayNeedsTheSameRequest(t *testing.T) {
	b := &subscriptionBridge{}
	b.noteLiteCompletedToolTurn(scOwner, scOuter, replayReq(t, replayBody("")), "CACHED_REPLY")
	if c, ok := b.liteCompletedToolTurn(scOwner, scOuter, replayReq(t, replayBody(""))); !ok || c.reply != "CACHED_REPLY" {
		t.Fatalf("the same request was not replayed: ok=%v", ok)
	}
	same := strings.Replace(replayBody(""), `"max_tokens":64`, `"max_tokens":64,"metadata":{"user_id":"another"}`, 1)
	if _, ok := b.liteCompletedToolTurn(scOwner, scOuter, replayReq(t, same)); !ok {
		t.Fatal("client metadata alone made a true repeat a new request")
	}
	for name, body := range map[string]string{
		"system":      strings.Replace(replayBody(""), `"system":"S1"`, `"system":"S2"`, 1),
		"tools":       strings.Replace(replayBody(""), `{"name":"echo","input_schema":{"type":"object"}}`, `{"name":"echo","input_schema":{"type":"object"}},{"name":"grep","input_schema":{"type":"object"}}`, 1),
		"tool schema": strings.Replace(replayBody(""), `{"type":"object"}`, `{"type":"object","required":["q"]}`, 1),
		"max_tokens":  strings.Replace(replayBody(""), `"max_tokens":64`, `"max_tokens":4096`, 1),
		"temperature": replayBody(`"temperature":0.2,`),
		"top_p":       replayBody(`"top_p":0.5,`),
		"stop":        replayBody(`"stop_sequences":["END"],`),
		"tool_choice": replayBody(`"tool_choice":{"type":"any"},`),
		"thinking":    replayBody(`"thinking":{"type":"disabled"},`),
		"schema":      replayBody(`"output_config":{"format":{"type":"json_schema","schema":{"type":"object"}}},`),
	} {
		if _, ok := b.liteCompletedToolTurn(scOwner, scOuter, replayReq(t, body)); ok {
			t.Errorf("a request with another %s was answered with the old reply", name)
		}
	}
	// Effort is read only when thinking is asked for.
	high := replayBody(`"thinking":{"type":"adaptive"},"output_config":{"effort":"high"},`)
	b.noteLiteCompletedToolTurn(scOwner, "effort-session", replayReq(t, high), "CACHED_REPLY")
	if _, ok := b.liteCompletedToolTurn(scOwner, "effort-session", replayReq(t, high)); !ok {
		t.Error("the same effort was not replayed")
	}
	if _, ok := b.liteCompletedToolTurn(scOwner, "effort-session", replayReq(t, strings.Replace(high, `"high"`, `"low"`, 1))); ok {
		t.Error("a request with another effort was answered with the old reply")
	}
	other := replayReq(t, replayBody(""))
	other.Model = "claude-sonnet-5-5"
	if _, ok := b.liteCompletedToolTurn(scOwner, scOuter, other); ok {
		t.Error("another model was answered with the old reply")
	}
	if _, ok := b.liteCompletedToolTurn(scOwner, "other-session", replayReq(t, replayBody(""))); ok {
		t.Error("the replay crossed into another conversation")
	}
	if _, ok := b.liteCompletedToolTurn("claude-tui\x00other-account", scOuter, replayReq(t, replayBody(""))); ok {
		t.Error("the replay crossed into another account")
	}
}

// Served end to end: the reply a native run completed is recorded from the
// request as it arrived, so the repeat is answered without a second PTY,
// and a changed system prompt with the same history starts a native run.
func TestLiteServedReplayIsKeyedByTheWholeRequest(t *testing.T) {
	_, runs := scWrapper(t)
	rt, b := scRuntime(t)
	scSaveClean(t, b, "inner-saved", false)
	if code, out := scServe(t, rt, scOuter, replayBody("")); code != 200 || !strings.Contains(out, "SC_ANSWER") {
		t.Fatalf("first delivery: %d %s", code, out)
	}
	if got := runs(); len(got) != 1 {
		t.Fatalf("first delivery: %d native starts, want 1", len(got))
	}
	if _, ok := b.liteCompletedToolTurn(scOwner, scOuter, replayReq(t, replayBody(""))); !ok {
		t.Fatal("the completed reply was not recorded for its own request")
	}
	if code, out := scServe(t, rt, scOuter, replayBody("")); code != 200 || !strings.Contains(out, "SC_ANSWER") {
		t.Fatalf("repeat: %d %s", code, out)
	}
	if got := runs(); len(got) != 1 {
		t.Fatalf("a true repeat started another native run: %d starts", len(got))
	}
	changed := strings.Replace(replayBody(""), `"system":"S1"`, `"system":"S2"`, 1)
	if code, out := scServe(t, rt, scOuter, changed); code != 200 || !strings.Contains(out, "SC_ANSWER") {
		t.Fatalf("changed system: %d %s", code, out)
	}
	if got := runs(); len(got) != 2 {
		t.Fatalf("the same history under another system prompt was replayed: %d starts, want 2", len(got))
	}
}

func TestLiteConcurrentRetriesOfACompletedTurnAreAllReplayed(t *testing.T) {
	_, runs := scWrapper(t)
	rt, b := scRuntime(t)
	scSaveClean(t, b, "inner-saved", false)
	b.noteLiteCompletedToolTurn(scOwner, scOuter, replayReq(t, replayBody("")), "CACHED_REPLY")
	var wg sync.WaitGroup
	outs := make([]string, 6)
	codes := make([]int, 6)
	for i := range outs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			codes[i], outs[i] = scServe(t, rt, scOuter, replayBody(""))
		}(i)
	}
	wg.Wait()
	for i := range outs {
		if codes[i] != 200 || !strings.Contains(outs[i], "CACHED_REPLY") {
			t.Fatalf("retry %d: %d %s", i, codes[i], outs[i])
		}
	}
	if got := runs(); len(got) != 0 {
		t.Fatalf("concurrent retries started %d native runs", len(got))
	}
}

// A client that offers web search gets Magpie's own search tool added to the
// run. The reply is recorded under the request as the client sent it, so
// its repeat is still replayed rather than run again.
func TestLiteReplayOfASearchingRequestIgnoresTheAddedTool(t *testing.T) {
	_, runs := scWrapper(t)
	rt, b := scRuntime(t)
	scSaveClean(t, b, "inner-saved", false)
	old := config.SearchHandler
	config.SearchHandler = func(context.Context, string) (string, []Hit, error) { return "results", nil, nil }
	t.Cleanup(func() { config.SearchHandler = old })
	body := strings.Replace(replayBody(""), `"tools":[`, `"tools":[{"type":"web_search_20250305","name":"web_search"},`, 1)
	if code, out := scServe(t, rt, scOuter, body); code != 200 || !strings.Contains(out, "SC_ANSWER") {
		t.Fatalf("first delivery: %d %s", code, out)
	}
	if code, out := scServe(t, rt, scOuter, body); code != 200 || !strings.Contains(out, "SC_ANSWER") {
		t.Fatalf("repeat: %d %s", code, out)
	}
	if got := runs(); len(got) != 1 {
		t.Fatalf("the repeat of a searching request started another native run: %d starts, want 1", len(got))
	}
}
