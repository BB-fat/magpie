package gateway

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tidwall/gjson"
	"github.com/yetone/magpie/internal/provider"
)

// vendorDeveloper400 is what a chat-role upstream answers a request
// carrying the developer role with (DeepSeek among them, #2955).
const vendorDeveloper400 = `{"error":{"code":"invalid_parameter_error","message":"developer is not one of ['system', 'assistant', 'user', 'tool', 'function']","param":null,"type":"invalid_request_error"}}`

// #2955: Codex on a provider whose Responses endpoint validates chat
// roles — a relay of a chat-only vendor's among them, DeepSeek's preset
// base alike: every compaction failed there with 400 "developer is not
// one of ['system', 'assistant', 'user', 'tool', 'function']" while the
// turns around it passed, and Codex retried each five times. The summary
// request carries a developer message its turns don't. Only after that
// explicit refusal does the summary retry with system messages in place.
func TestCodexCompactDeveloperAsSystemOnAVendorRelay(t *testing.T) {
	f := &fake{t: t, reply: sse(
		`data: {"type":"response.created","response":{"id":"resp_v1","status":"in_progress","output":[]}}`,
		`data: {"type":"response.output_item.done","output_index":0,"item":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"SUMMARY"}]}}`,
		`data: {"type":"response.completed","response":{"id":"resp_v1","status":"completed","output":[]}}`)}
	f.refuse = func(body []byte) (int, string) {
		if strings.Contains(string(body), `"role":"developer"`) {
			return http.StatusBadRequest, vendorDeveloper400
		}
		return 0, ""
	}
	setup(t, provider.Responses, f)
	request := `{"model":"fake/m1","stream":true,"store":false,
	  "input":[{"type":"message","role":"developer","content":[{"type":"input_text","text":"You are Codex, a coding agent."}]},
	  {"type":"message","role":"user","content":[{"type":"input_text","text":"fix the login bug"}]},
	  {"type":"message","role":"assistant","content":[{"type":"output_text","text":"I fixed main.go"}]},
	  {"type":"compaction_trigger"}]}`
	srv := New()
	send := func(body string) (int, string) {
		r := httptest.NewRequest("POST", CodexPath+"/responses", strings.NewReader(body))
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, r)
		return w.Code, w.Body.String()
	}
	code, body := send(request)
	if code != 200 || !strings.Contains(body, `"compaction"`) {
		t.Fatalf("%d %s\nupstream got %s", code, body, f.got)
	}
	if f.calls != 2 {
		t.Fatalf("expected developer refusal then system retry, got %d calls", f.calls)
	}
	if got := compactedText(t, body); got != "SUMMARY" {
		t.Errorf("summary %q", got)
	}
	if strings.Contains(string(f.got), `"role":"developer"`) {
		t.Errorf("developer went over the wire: %s", f.got)
	}
	// the message itself stays, where it stood, as a system one
	if !strings.Contains(string(f.got), `"role":"system"`) || !strings.Contains(string(f.got), "You are Codex, a coding agent.") {
		t.Errorf("system message lost: %s", f.got)
	}
	for _, want := range []string{"fix the login bug", "I fixed main.go", "CONTEXT CHECKPOINT COMPACTION"} {
		if !strings.Contains(string(f.got), want) {
			t.Errorf("summary request lacks %q: %s", want, f.got)
		}
	}
	// The next summary learns the accepted role, but another model and
	// ordinary turns keep the role the client sent.
	if code, body := send(request); code != 200 || f.calls != 3 {
		t.Fatalf("next summary: %d %s, calls %d", code, body, f.calls)
	}
	if !srv.fits("fake", compactDeveloperRefused("m2"), provider.Responses) {
		t.Fatal("role refusal leaked to another model")
	}
	f.refuse = nil
	turn := strings.Replace(request, `{"type":"compaction_trigger"}`, `{"type":"message","role":"user","content":"continue"}`, 1)
	if code, body := send(turn); code != 200 || !strings.Contains(string(f.got), `"role":"developer"`) {
		t.Fatalf("ordinary turn changed after summary retry: %d %s, upstream %s", code, body, f.got)
	}
}

func TestCompactDeveloperRoleRefusal(t *testing.T) {
	for _, tt := range []struct {
		body string
		want bool
	}{
		{vendorDeveloper400, true},
		{`{"error":{"message":"Relay: System messages are not allowed"}}`, false},
		{`{"error":{"message":"developer account not allowed"}}`, false},
		{`{"error":{"message":"invalid developer message content"}}`, false},
		{`{"error":{"message":"native codex clients only"}}`, false},
	} {
		if got := compactDeveloperRoleRefusal([]byte(tt.body)); got != tt.want {
			t.Errorf("refusal(%s) = %v, want %v", tt.body, got, tt.want)
		}
	}
}

// A Responses relay refused a real compact on 11-09 with "System messages are not
// allowed". Its incoming history contained developer messages only;
// forVendor introduced system roles solely because this was a summary.
func TestCodexCompactKeepsDeveloperOnCodexRelay(t *testing.T) {
	for _, path := range []string{CodexPath + "/responses", "/v1/responses"} {
		t.Run(path, func(t *testing.T) {
			f := &fake{t: t, reply: sse(
				`data: {"type":"response.output_item.done","output_index":0,"item":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"SUMMARY"}]}}`,
				`data: {"type":"response.completed","response":{"id":"r","status":"completed","output":[]}}`)}
			f.refuse = func(body []byte) (int, string) {
				for _, item := range gjson.GetBytes(body, "input").Array() {
					if item.Get("role").String() == "system" {
						return 400, `{"error":{"message":"Relay: System messages are not allowed","type":"invalid_request_error"}}`
					}
				}
				return 0, ""
			}
			setup(t, provider.Responses, f)
			body := `{"model":"fake/m1","stream":true,"store":false,"input":[{"type":"message","role":"developer","content":[{"type":"input_text","text":"You are Codex, a coding agent."}]},{"type":"message","role":"user","content":"fix the bug"},{"type":"message","role":"developer","content":[{"type":"input_text","text":"Keep the existing tests."}]},{"type":"compaction_trigger"}]}`
			r := httptest.NewRequest("POST", path, strings.NewReader(body))
			r.Header.Set("User-Agent", "codex-tui/0.167.0")
			r.Header.Set("Originator", "Codex Desktop")
			w := httptest.NewRecorder()
			New().Handler().ServeHTTP(w, r)
			if w.Code != 200 || f.calls != 1 {
				t.Fatalf("compact: %d %s (upstream calls %d)", w.Code, w.Body.String(), f.calls)
			}
			if got := compactedText(t, w.Body.String()); got != "SUMMARY" {
				t.Fatalf("summary = %q", got)
			}
			input := gjson.GetBytes(f.got, "input").Array()
			for _, i := range []int{0, 2} {
				want := gjson.Get(body, "input").Array()[i]
				if !sameJSONText(input[i].Raw, want.Raw) {
					t.Errorf("instruction %d changed: %s", i, input[i].Raw)
				}
			}
		})
	}
}

// A turn on the same provider goes as it came, developer message and all:
// only magpie's own summary ask is fitted, never Codex's own turn, whose
// shape a relay may be checking (#350, #1758).
func TestVendorRelayTurnKeepsItsDeveloperMessage(t *testing.T) {
	f := &fake{t: t, reply: sse(
		`data: {"type":"response.completed","response":{"id":"resp_v2","status":"completed","output":[]}}`)}
	setup(t, provider.Responses, f)
	code, body := codexPost(t, `{"model":"fake/m1","stream":true,"store":false,
	  "input":[{"type":"message","role":"developer","content":[{"type":"input_text","text":"You are Codex, a coding agent."}]},
	  {"type":"message","role":"user","content":[{"type":"input_text","text":"hello"}]}]}`)
	if code != 200 {
		t.Fatalf("%d %s", code, body)
	}
	if !strings.Contains(string(f.got), "hello") || !strings.Contains(string(f.got), `"role":"developer"`) {
		t.Errorf("turn's developer message changed: %s", f.got)
	}
}
