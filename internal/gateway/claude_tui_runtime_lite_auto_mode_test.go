package gateway

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Claude Code 2.1.308's auto mode classifier, as its request is built:
// the policy as a marked system block, the transcript in text blocks
// between <transcript> and </transcript>, its last block and the action
// marked for the cache, no tools.
const tuiRuntimeAutoModeClassifierBody = `{"model":"claude-opus-5-5","max_tokens":64,"temperature":0,"thinking":{"type":"disabled"},
	"system":[{"type":"text","text":"You are a security classifier for an autonomous coding agent.","cache_control":{"type":"ephemeral","ttl":"1h"}}],
	"messages":[{"role":"user","content":[
		{"type":"text","text":"<transcript>\n"},
		{"type":"text","text":"User: fix the failing test\n"},
		{"type":"text","text":"Bash: go test ./...\n","cache_control":{"type":"ephemeral","ttl":"1h"}},
		{"type":"text","text":"Bash: rm -rf build\n","cache_control":{"type":"ephemeral","ttl":"1h"}},
		{"type":"text","text":"</transcript>\n"},
		{"type":"text","text":"Err on the side of blocking. <block> immediately."}]}]}`

// On a TUI account the classifier is refused before a native run starts,
// as a request Claude Code won't retry: each one typed into the PTY as one
// prompt block wrote the whole transcript to the cache again.
func TestClaudeTUIRuntimeLiteRefusesAutoModeClassifier(t *testing.T) {
	_, runs := tuiRuntimeScWrapper(t)
	rt, _ := tuiRuntimeScRuntime(t)
	for _, body := range []string{
		tuiRuntimeAutoModeClassifierBody,
		// a segmented transcript: each block a user message of its own
		`{"model":"claude-opus-5-5","max_tokens":64,"messages":[
			{"role":"user","content":[{"type":"text","text":"<transcript>\n"}]},
			{"role":"user","content":[{"type":"text","text":"Bash: rm -rf build\n"}]},
			{"role":"user","content":[{"type":"text","text":"</transcript>\n"}]}]}`,
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
		code, _ := rt.serveMessages(rec, req, "claude-opus-5-5", tuiRuntimeScOuter, []byte(body), &Usage{})
		if code != http.StatusBadRequest || rec.Header().Get("x-should-retry") != "false" || !strings.Contains(rec.Body.String(), "auto mode") {
			t.Fatalf("classifier not refused as final: %d retry=%q %s", code, rec.Header().Get("x-should-retry"), rec.Body.String())
		}
		if retryable(code, rec.Body.Bytes()) {
			t.Fatalf("the refusal sends the classifier on to the next account: %s", rec.Body.String())
		}
	}
	if got := runs(); len(got) != 0 {
		t.Fatalf("a native run started for the classifier: %v", got)
	}
}

// Only the classifier is: a conversation that quotes a transcript, or a
// request with tools, is answered as ever.
func TestClaudeTUIRuntimeLiteAnswersWhatOnlyLooksLikeAutoMode(t *testing.T) {
	_, runs := tuiRuntimeScWrapper(t)
	rt, _ := tuiRuntimeScRuntime(t)
	for _, body := range []string{
		`{"model":"claude-opus-5-5","max_tokens":64,"messages":[{"role":"user","content":"look at this <transcript>\nfoo</transcript>"}]}`,
		`{"model":"claude-opus-5-5","max_tokens":64,"tools":[{"name":"echo","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":[{"type":"text","text":"<transcript>\n"},{"type":"text","text":"</transcript>\n"}]}]}`,
		`{"model":"claude-opus-5-5","max_tokens":64,"messages":[{"role":"user","content":"<transcript>\nsummarize</transcript>"},{"role":"assistant","content":"ok"},{"role":"user","content":"go on"}]}`,
	} {
		if code, out := tuiRuntimeScServe(t, rt, "", body); code != 200 || !strings.Contains(out, "SC_ANSWER") {
			t.Fatalf("an ordinary request was refused: %d %s\n%s", code, out, body)
		}
	}
	if got := runs(); len(got) != 3 {
		t.Fatalf("%d native runs, want 3", len(got))
	}
}
