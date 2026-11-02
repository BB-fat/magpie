package gateway

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/access"
)

func TestClaudeTUIRuntimeCompletedReplyIsCallerScoped(t *testing.T) {
	_, runs := tuiRuntimeScWrapper(t)
	tuiRuntimeWithConfig(t, &tuiRuntimeConfig{StateDir: t.TempDir()})
	rt, b := tuiRuntimeScRuntime(t)
	body := tuiRuntimeReplayBody("")
	a := tuiRuntimeReplayReq(t, body)
	a.leanPrompt = tuiRuntimeLeanPrompt()
	a.scope = tuiRuntimePrefixScope(tuiRuntimeScOwner, "caller-A")
	b.noteLiteCompletedToolTurn(tuiRuntimeScOwner, tuiRuntimeScOuter, a, "PRIVATE_A")
	other := *a
	other.scope = tuiRuntimePrefixScope(tuiRuntimeScOwner, "caller-B")
	if tuiRuntimeLiteReplayKey(tuiRuntimeScOwner, a) == tuiRuntimeLiteReplayKey(tuiRuntimeScOwner, &other) {
		t.Error("different callers have the same replay identity")
	}
	for _, caller := range []string{"caller-A", "caller-B"} {
		r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
		r = r.WithContext(access.WithIdentity(r.Context(), access.Identity{KeyID: caller}))
		rec := httptest.NewRecorder()
		code, _ := rt.serveMessages(rec, r, a.Model, tuiRuntimeScOuter, []byte(body), &Usage{})
		private := strings.Contains(rec.Body.String(), "PRIVATE_A")
		if code != 200 || private != (caller == "caller-A") {
			t.Errorf("%s received code=%d body=%s", caller, code, rec.Body.String())
		}
	}
	b.abortAll()
	if n := len(runs()); n != 1 {
		t.Errorf("expected caller B's own native inference, got %d starts", n)
	}
	// Both callers keep their own completion; neither evicts the other merely
	// by sharing an outer session header.
	b.noteLiteCompletedToolTurn(tuiRuntimeScOwner, tuiRuntimeScOuter, &other, "PRIVATE_B")
	for _, tc := range []struct {
		req  *tuiRuntimeRequest
		want string
	}{{a, "PRIVATE_A"}, {&other, "PRIVATE_B"}} {
		got, ok := b.liteCompletedToolTurn(tuiRuntimeScOwner, tuiRuntimeScOuter, tc.req)
		if !ok || got.reply != tc.want {
			t.Errorf("caller completion = %+v, %t; want %s", got, ok, tc.want)
		}
	}
}
