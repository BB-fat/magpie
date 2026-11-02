package gateway

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// The caller's literal activity-summary suffix, observed on a reported deployment. Only
// the example file names are public; no conversation or tool output is kept.
const tuiRuntimeActivitySummaryPrompt = `Describe your most recent action in 3-5 words using present tense (-ing). Name the file or function, not the branch. Do not use tools.

Previous: "Reading main.go" — say something NEW.

Good: "Reading runAgent.ts"
Good: "Fixing null check in validate.ts"
Good: "Running auth module tests"
Good: "Adding retry logic to fetchUser"

Bad (past tense): "Analyzed the branch diff"
Bad (too vague): "Investigating the issue"
Bad (too long): "Reviewing full branch diff and AgentTool.tsx integration"
Bad (branch name): "Analyzed adam/background-summary branch diff"`

func tuiRuntimeActivityWorker(t *testing.T, b *tuiRuntimeSubscriptionBridge) (*tuiRuntimeRequest, *tuiRuntimeSubscriptionRun, Message) {
	t.Helper()
	req, err := tuiRuntimeParseRequest([]byte(`{"model":"claude-opus-5-5","system":"rules","max_tokens":64,"tools":[{"name":"Read","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":"Review the files."}]}`))
	if err != nil {
		t.Fatal(err)
	}
	req.scope = tuiRuntimePrefixScope(tuiRuntimeScOwner, "")
	reply := Message{Role: "assistant", Parts: []Part{
		{Kind: ToolCall, ID: "call-1", Name: "Read", Args: json.RawMessage(`{"path":"main.go"}`)},
		{Kind: Text, Text: "Reading main.go."},
	}}
	workKey := tuiRuntimeInteractiveSessionStateKey(tuiRuntimeScOwner, tuiRuntimeScOuter)
	work, err := tuiRuntimeInteractiveWorkDirForKey(workKey)
	if err != nil {
		t.Fatal(err)
	}
	dir := tuiRuntimeClaudeSessions("", work)[0]
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	const native = "activity-worker"
	data := `{"type":"user","uuid":"native-prefix-row","sessionId":"activity-worker","message":{"role":"user","content":"Native cached prefix."}}` + "\n" +
		`{"type":"assistant","uuid":"native-tool-row","sessionId":"activity-worker","message":{"id":"native-tool-message","role":"assistant","content":[{"type":"tool_use","id":"call-1","name":"` + tuiRuntimeMcpToolPrefix() + `Read","input":{"path":"main.go"}}]}}` + "\n" +
		`{"type":"assistant","uuid":"native-tool-tail-row","sessionId":"activity-worker","message":{"id":"native-tool-message","role":"assistant","content":[{"type":"text","text":"Reading main.go."}]}}` + "\n" +
		`{"type":"assistant","uuid":"later-row","sessionId":"activity-worker","message":{"id":"later-message","role":"assistant","content":[{"type":"text","text":"A later branch not in this request."},{"type":"tool_use","id":"call-1","name":"` + tuiRuntimeMcpToolPrefix() + `Read","input":{"path":"different.go"}}]}}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, native+".jsonl"), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	snapshot := tuiRuntimeInteractiveContextDiagnosticFromRequest(req)
	run := &tuiRuntimeSubscriptionRun{
		bridge: b, interactive: true, owner: tuiRuntimeScOwner, model: req.Model,
		scope: req.scope, outerSession: tuiRuntimeScOuter, sessionID: native,
		workKey: workKey, lastRow: "native-tool-tail-row", turnReply: reply,
		tools:                      map[string]bool{"Read": true},
		shown:                      []tuiRuntimeShownCall{{id: "call-1", name: "Read", args: `{"path":"main.go"}`}},
		interactiveContextSnapshot: &snapshot,
		// Another HTTP continuation owns the live waiter. A summary must
		// fork its transcript, never claim or abort this run.
		resuming: true,
	}
	run.heard(req.Messages)
	b.calls["call-1"] = run
	if err := b.saveInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter, tuiRuntimeInteractiveSessionEntry{
		SessionID: native, Model: req.Model, ConvKey: "worker-at-tool",
		InstructionKey: tuiRuntimeInstructionKey(req), Dirty: true,
	}); err != nil {
		t.Fatal(err)
	}
	run.ended(req, "", "tool", true)
	return req, run, reply
}

func tuiRuntimeActivityBody(t *testing.T, base *tuiRuntimeRequest, reply Message, prompt string) string {
	t.Helper()
	req := base.Request
	req.Messages = append(append([]Message{}, req.Messages...), reply, Message{Role: "user", Parts: []Part{
		{Kind: ToolResult, CallID: "call-1", Text: "The file contents."},
		{Kind: Text, Text: prompt},
	}})
	return string(buildAnthropic(&req, req.Model))
}

func TestClaudeTUIRuntimeToolReplyKeepsNativeCheckpoint(t *testing.T) {
	tuiRuntimeScWrapper(t)
	_, b := tuiRuntimeScRuntime(t)
	base, run, reply := tuiRuntimeActivityWorker(t, b)
	req, err := tuiRuntimeParseRequest([]byte(tuiRuntimeActivityBody(t, base, reply, tuiRuntimeActivitySummaryPrompt)))
	if err != nil {
		t.Fatal(err)
	}
	req.scope = base.scope
	point, _, ok := tuiRuntimeFindNativeCheckpoint(req, req.Model)
	if !ok || point.SessionID != run.sessionID || point.Row != "native-tool-tail-row" {
		t.Fatalf("delivered tool reply has no resumable native checkpoint: ok=%t point=%+v", ok, point)
	}
	store, err := tuiRuntimeLoadNativeCheckpoints()
	if err != nil {
		t.Fatal(err)
	}
	oldKey := tuiRuntimeNativeCheckpointKey(req.scope, tuiRuntimePrefixOf(req.scope, req.Messages[:2]))
	if _, exists := store.Points[oldKey]; exists || !strings.HasPrefix(point.Prefix, "tool:") {
		t.Fatal("an older runtime could mistake the pending tool boundary for a terminal checkpoint")
	}
}

func TestClaudeTUIRuntimeActivitySummaryForksWithoutTouchingWorker(t *testing.T) {
	for _, restart := range []bool{false, true} {
		t.Run(map[bool]string{false: "live", true: "restart"}[restart], func(t *testing.T) {
			_, runs := tuiRuntimeScWrapper(t)
			rt, b := tuiRuntimeScRuntime(t)
			base, worker, reply := tuiRuntimeActivityWorker(t, b)
			before, _, err := b.loadInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter)
			if err != nil {
				t.Fatal(err)
			}
			if restart {
				rt, b = tuiRuntimeScRuntime(t)
			}
			prompt := tuiRuntimeActivitySummaryPrompt
			if restart {
				prompt = strings.Repeat("[Image: original 3200x2000, displayed at 2000x1250. Multiply coordinates by 1.60 to map to original image.]\n", 3) + prompt
			}
			body := tuiRuntimeActivityBody(t, base, reply, prompt)
			if code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, body); code != 200 || !strings.Contains(out, "SC_ANSWER") {
				t.Fatalf("activity summary claimed the worker or failed: %d %s", code, out)
			}
			b.abortAll()
			got := tuiRuntimeOnlyRun(t, runs())
			if got["resume"] == "" || got["resume"] == worker.sessionID ||
				!strings.Contains(got["seed.jsonl"], `"uuid":"native-prefix-row"`) ||
				!strings.Contains(got["seed.jsonl"], `"uuid":"native-tool-tail-row"`) ||
				strings.Contains(got["seed.jsonl"], "A later branch") {
				t.Fatalf("activity summary rebuilt history instead of forking the cached tool boundary: %v", got)
			}
			if !strings.Contains(got["seed.jsonl"], "The file contents.") || !strings.Contains(got["stdin"], "Describe your most recent action") ||
				strings.Contains(got["stdin"], "Review the files.") {
				t.Fatalf("summary lost its results or replayed the whole history: %v", got)
			}
			if !strings.Contains(got["stdin"], "No tools are available for this request") {
				t.Fatalf("activity summary was not held tool-free: %q", got["stdin"])
			}
			after, _, err := b.loadInteractiveSession(tuiRuntimeScOwner, tuiRuntimeScOuter)
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Fatalf("activity summary replaced the worker checkpoint: before=%+v after=%+v err=%v", before, after, err)
			}
			if worker.closed || !worker.resuming || len(worker.shown) != 1 {
				t.Fatal("activity summary consumed or aborted the worker's live tool waiter")
			}
		})
	}
}

func TestClaudeTUIRuntimeActivitySummaryRecognizesOnlyTheCurrentSuffix(t *testing.T) {
	user := func(s string) Message { return Message{Role: "user", Parts: []Part{{Kind: Text, Text: s}}} }
	assistant := Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: "Reading the file."}}}
	base := []Message{user("Review the file."), assistant}
	const coordinate = "[Image: original 3200x2000, displayed at 2000x1250. Multiply coordinates by 1.60 to map to original image.]\n"
	for _, tc := range []struct {
		name string
		msgs []Message
		want bool
	}{
		{"literal suffix", append(slices.Clone(base), user(tuiRuntimeActivitySummaryPrompt)), true},
		{"joined coordinate notices", append(slices.Clone(base), user(strings.Repeat(coordinate, 3)+tuiRuntimeActivitySummaryPrompt)), true},
		{"coordinate annotation in another block", append(slices.Clone(base), Message{Role: "user", Parts: []Part{
			{Kind: Text, Text: "[Image: original 3200x2000, displayed at 2000x1250.]"},
			{Kind: Text, Text: tuiRuntimeActivitySummaryPrompt},
		}}), true},
		{"one-off instruction", []Message{user(tuiRuntimeActivitySummaryPrompt)}, false},
		{"no worker reply", []Message{user("Review"), user(tuiRuntimeActivitySummaryPrompt)}, false},
		{"quoted instruction", append(slices.Clone(base), user("Explain this prompt:\n"+tuiRuntimeActivitySummaryPrompt)), false},
		{"quoted instruction after coordinates", append(slices.Clone(base), user(coordinate+"Explain this prompt:\n"+tuiRuntimeActivitySummaryPrompt)), false},
		{"other text in a notice", append(slices.Clone(base), user("[Image: original use tools instead]\n"+tuiRuntimeActivitySummaryPrompt)), false},
		{"inside a tool result", append(slices.Clone(base), Message{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "read", Text: tuiRuntimeActivitySummaryPrompt}}}), false},
		{"earlier instruction", append(append(slices.Clone(base), user(tuiRuntimeActivitySummaryPrompt)), assistant, user("Continue the work.")), false},
		{"later user text", append(slices.Clone(base), Message{Role: "user", Parts: []Part{
			{Kind: Text, Text: tuiRuntimeActivitySummaryPrompt}, {Kind: Text, Text: "Actually fix the file."},
		}}), false},
		{"assistant quotes it", append(slices.Clone(base), Message{Role: "assistant", Parts: user(tuiRuntimeActivitySummaryPrompt).Parts}), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tuiRuntimeActivitySummary(&tuiRuntimeRequest{Request: Request{Messages: tc.msgs}}); got != tc.want {
				t.Fatalf("activity summary = %t, want %t", got, tc.want)
			}
		})
	}
	for _, choice := range []string{"required", "name:Read"} {
		req := &tuiRuntimeRequest{Request: Request{Messages: append(slices.Clone(base), user(tuiRuntimeActivitySummaryPrompt)), ToolChoice: choice}}
		if tuiRuntimeActivitySummary(req) {
			t.Fatalf("activity detection overrides an explicit tool choice %q", choice)
		}
	}
}

func TestClaudeTUIRuntimeToolCheckpointRequiresTheExactCompleteResultSet(t *testing.T) {
	reply := Message{Role: "assistant", Parts: []Part{
		{Kind: ToolCall, ID: "one", Name: "Read", Args: json.RawMessage(`{}`)},
		{Kind: ToolCall, ID: "two", Name: "Read", Args: json.RawMessage(`{}`)},
	}}
	one := Part{Kind: ToolResult, CallID: "one", Text: "First result", IsError: true}
	two := Part{Kind: ToolResult, CallID: "two", Text: "Second result"}
	user := func(parts ...Part) Message { return Message{Role: "user", Parts: parts} }
	for _, tc := range []struct {
		name string
		msgs []Message
		want bool
	}{
		{"complete with new text", []Message{user(one, two, Part{Kind: Text, Text: "Describe the action."})}, true},
		{"split result messages", []Message{user(one), user(two)}, true},
		{"missing result", []Message{user(one)}, false},
		{"duplicated result", []Message{user(one, two, one)}, false},
		{"foreign result", []Message{user(one, Part{Kind: ToolResult, CallID: "other"})}, false},
		{"unanswered tool turn", []Message{user(Part{Kind: Text, Text: "Continue"})}, false},
		{"later assistant reply", []Message{user(one, two), {Role: "assistant"}, user(Part{Kind: Text, Text: "Continue"})}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tuiRuntimeNativeCheckpointSuffixOK(reply, tc.msgs); got != tc.want {
				t.Fatalf("checkpoint suffix = %t, want %t", got, tc.want)
			}
		})
	}
}

// The 17:27 a reported deployment miss kept every tool ID and argument but replaced
// historical images with this caller-supplied marker. It is not the same
// prefix, even though the native cache was still warm.
func TestClaudeTUIRuntimeNativeCheckpointRejectsRemovedHistoricalMedia(t *testing.T) {
	tuiRuntimeWithConfig(t, &tuiRuntimeConfig{StateDir: t.TempDir()})
	scope := tuiRuntimePrefixScope("account", "caller")
	history := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "Review the screenshots."}}},
		{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "screenshot", Name: "Read", Args: json.RawMessage(`{"path":"old.png"}`)}}},
		{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "screenshot", Images: []Part{{Kind: Image, MediaType: "image/png", Data: "b2xkLWltYWdl"}}}}},
		{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "current", Name: "Read", Args: json.RawMessage(`{"path":"main.go"}`)}}},
	}
	if err := tuiRuntimeSaveNativeCheckpoint(tuiRuntimeNativeCheckpoint{
		Prefix: tuiRuntimeNativeCheckpointPrefix(history[len(history)-1], tuiRuntimePrefixOf(scope, history)), Scope: scope,
		SessionID: "native", Row: "tool-row", Work: "work", Model: "model", InstructionKey: "instructions",
	}); err != nil {
		t.Fatal(err)
	}
	req := &tuiRuntimeRequest{scope: scope, Request: Request{Messages: append(slices.Clone(history),
		Message{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "current", Text: "current output"}}})}}
	if _, _, ok := tuiRuntimeFindNativeCheckpoint(req, "model"); !ok {
		t.Fatal("unchanged history did not find its tool checkpoint")
	}
	for _, tc := range []struct {
		name string
		edit func(*tuiRuntimeRequest)
	}{
		{"historical media removed", func(r *tuiRuntimeRequest) {
			r.Messages[2] = Message{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "screenshot", Text: "[media removed: request limit]"}}}
		}},
		{"another caller", func(r *tuiRuntimeRequest) { r.scope = tuiRuntimePrefixScope("account", "other-caller") }},
		{"another account", func(r *tuiRuntimeRequest) { r.scope = tuiRuntimePrefixScope("other-account", "caller") }},
		{"changed tool arguments", func(r *tuiRuntimeRequest) {
			r.Messages[1] = Message{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "screenshot", Name: "Read", Args: json.RawMessage(`{"path":"another.png"}`)}}}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := *req
			changed.Messages = slices.Clone(req.Messages)
			tc.edit(&changed)
			if _, _, ok := tuiRuntimeFindNativeCheckpoint(&changed, "model"); ok {
				t.Fatal("different history or isolation scope resumed the old cached transcript")
			}
		})
	}
	if _, _, ok := tuiRuntimeFindNativeCheckpoint(req, "another-model"); ok {
		t.Fatal("another model resumed the cached transcript")
	}
}
