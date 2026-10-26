package claudetui

// Test-only compatibility shim: the migrated production regression tests
// were written against the gateway's Server type and call s.serveSubscription
// with a start func. The extracted runtime's handler is serveMessagesInner
// on runtimeTUI, so newTestServer() here wraps that. The signature mirrors the
// deployed one so migrated tests compile without edit.

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// Legacy assertion helpers only. Production callers must handle the error
// returned by loadInteractiveSession rather than treating damage as absence.
func (b *subscriptionBridge) interactiveSession(owner, outer string) (interactiveSessionEntry, bool) {
	entry, ok, _ := b.loadInteractiveSession(owner, outer)
	return entry, ok
}

func readInteractiveSessionState() interactiveSessionState {
	state, _ := loadInteractiveSessionStateStrict()
	return state
}

// Server mirrors the deployed gateway's struct surface the tests reach.
type Server struct {
	subscription *subscriptionBridge
	rt           *runtimeTUI
}

// newTestServer makes a test server. Production tests call it then
// s.serveSubscription(...); the runtime has no provider plumbing, so the
// proto/owner/etc are forwarded straight through.
func newTestServer() *Server {
	rt := NewRuntime([]byte("test-key"))
	return &Server{subscription: rt.bridge, rt: rt}
}

// serveSubscription adapts the production signature:
//
//	s.serveSubscription(w, r, proto, name, model, owner, body, usage, start)
//
// where `proto` was provider.Protocol. In the bridge package it is a
// plain string and only protoAnthropic is ever passed.
func (s *Server) serveSubscription(w http.ResponseWriter, r *http.Request, from, name, model, owner string, body []byte, usage *Usage,
	start func(ctx context.Context, req *Request) (*subscriptionRun, <-chan Event, error)) (int, string) {
	// Tests reseat s.subscription with a bridge of their own; the runtime
	// handler must consult the same one.
	s.rt.bridge = s.subscription
	// Tests drive serveMessagesInner directly with an empty outer session;
	// the persisted-session recovery paths the outer session keys into are
	// exercised through startInteractiveRequest's own outerSession argument.
	return s.rt.serveMessagesInner(anthropicResponder{w, r, usage}, r, name, model, owner, "", body, start)
}

// webSearch is the hook serveMessagesInner reads for the synthetic
// web_search tool; production tests stub it on Server.
func (s *Server) webSearch(ctx context.Context, query string) (string, []Hit, error) {
	if config.SearchHandler != nil {
		return config.SearchHandler(ctx, query)
	}
	return "", nil, errNoSearcher
}

var errNoSearcher = errSentinel("no searcher")

type errSentinel string

func (e errSentinel) Error() string { return string(e) }

// testInit resets the package Config for a test; production tests rely on
// each test starting from a clean globals slate.
func testInit(t *testing.T, cfg *Config) {
	t.Helper()
	Init(cfg)
	t.Cleanup(func() { Init(&Config{}) })
}

// fakeClaude is the script Claude Code double used by subscription tests:
// it answers each stdin line with a stream-json reply from its own pid.
// Extracted verbatim from claude_subscription_test.go.
func fakeClaude(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("a shell script stands in for Claude Code")
	}
	dir := t.TempDir()
	script := `#!/bin/sh
n=0
while read -r line; do
  n=$((n+1))
  echo '{"type":"stream_event","event":{"type":"message_start","message":{"id":"m","model":"claude-sonnet-5","usage":{"input_tokens":1}}}}'
  echo '{"type":"stream_event","event":{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"pid '$$' turn '$n'"}}}'
  echo '{"type":"stream_event","event":{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":3}}}'
  echo '{"type":"stream_event","event":{"type":"message_stop"}}'
  echo '{"type":"result","subtype":"success","is_error":false,"result":""}'
done
`
	os.WriteFile(filepath.Join(dir, "claude"), []byte(script), 0o755)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// setWrapper has runs start argv as their PTY wrapper.
func setWrapper(t *testing.T, argv ...string) {
	t.Helper()
	old := wrapperOverride
	wrapperOverride = argv
	t.Cleanup(func() { wrapperOverride = old })
}

// keepFast makes the keepalives of a test's streams come at its pace.
// Extracted from keepalive_test.go.
func keepFast(t *testing.T, gap, every, longest time.Duration) {
	oldGap, oldEvery, oldLongest := keepaliveGap, keepaliveEvery, keepaliveLongest
	keepaliveGap, keepaliveEvery, keepaliveLongest = gap, every, longest
	t.Cleanup(func() { keepaliveGap, keepaliveEvery, keepaliveLongest = oldGap, oldEvery, oldLongest })
}
