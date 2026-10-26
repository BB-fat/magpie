package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Automatic retries can contain exactly the same mixed text/tool-result as
// a fresh user turn. Expiring the callback does not establish new provenance.
func TestClaudeTUIRuntimeInteractiveMergedOrphanRetryCannotBecomeNewTurn(t *testing.T) {
	body := []byte(`{"model":"claude-opus-5-5","messages":[{"role":"user","content":"original request"},{"role":"assistant","content":[{"type":"tool_use","id":"failed-call","name":"write_file","input":{"path":"already-written"}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"failed-call","content":"write completed"},{"type":"text","text":"also verify the result"},{"type":"text","text":"[Request interrupted by user]"},{"type":"text","text":"please check read-only"}]},{"role":"system","content":"<system-reminder>\n<total_tokens>90 tokens left</total_tokens>\n</system-reminder>"}]}`)
	for _, scenario := range []string{"automatic retry after abort", "retry after runtime restart"} {
		t.Run(scenario, func(t *testing.T) {
			s := tuiRuntimeNewTestServer()
			defer tuiRuntimeBridges.Delete(s.subscription)
			req, err := tuiRuntimeParseRequest(body)
			if err != nil {
				t.Fatal(err)
			}
			if current := tuiRuntimeInteractiveAbandonedToolRecoveryIndex(req); current >= 0 {
				t.Fatalf("ambiguous mixed retry treated as current user turn: %d", current)
			}
			started := 0
			start := func(context.Context, *tuiRuntimeRequest) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
				started++
				return nil, nil, errors.New("ambiguous history must not start a process")
			}
			for attempt := 0; attempt < 2; attempt++ {
				w := httptest.NewRecorder()
				code, _ := s.serveSubscription(w, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body))), protoAnthropic, "Claude Code", req.Model, "owner", body, &Usage{}, start)
				if code != http.StatusConflict || started != 0 {
					t.Fatalf("retry %d: status=%d starts=%d", attempt, code, started)
				}
			}
		})
	}
}
