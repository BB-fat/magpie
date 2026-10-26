package gateway

import (
	"context"
	"net/http"
	"strings"
)

// runtimeTUI owns the bridge a gateway's TUI requests are answered by.
type tuiRuntime struct {
	bridge *tuiRuntimeSubscriptionBridge
}

// responder answers the client: a refusal before any reply, or a reply's
// events. The gateway's writes the client's own protocol and counts the
// usage; anthropicResponder writes Messages, as the plugin's runtime did.
type tuiRuntimeResponder interface {
	Header() http.Header
	Fail(status int, msg string) (int, string)
	Relay(req *Request, events <-chan Event, abort func(), ended func(said, stop string, ok bool)) (int, string)
}

// Responder is how the gateway answers a TUI request's client.
type tuiRuntimeOutput interface {
	// Header is the reply's headers, set before Fail or Relay.
	Header() http.Header
	// Fail answers status with msg before any reply.
	Fail(status int, msg string) (int, string)
	// Relay answers with events, as the gateway's relay does: abort when
	// there is no answer to give, ended once one was given.
	Relay(events <-chan Event, stream bool, abort func(), ended func(said, stop string, ok bool)) (int, string)
}

type tuiRuntimeGatewayResponder struct{ tuiRuntimeOutput }

func (g tuiRuntimeGatewayResponder) Relay(req *Request, events <-chan Event, abort func(), ended func(said, stop string, ok bool)) (int, string) {
	return g.tuiRuntimeOutput.Relay(events, req.Stream, abort, ended)
}

// Bridge is a gateway's Claude TUI bridge.
type tuiRuntimeBridge struct{ rt *tuiRuntime }

// New makes a bridge. Init has set its Config.
func newTUIBridge() *tuiRuntimeBridge {
	return &tuiRuntimeBridge{rt: &tuiRuntime{bridge: tuiRuntimeNewSubscriptionBridge()}}
}

// Serve answers body, an Anthropic Messages request for model on the
// account owner names, through the interactive Claude Code run in the
// account's config folder configDir: owner is part of every checkpoint's
// identity, outerSession the conversation the client names ("" for a
// request of its own).
func (b *tuiRuntimeBridge) Serve(r *http.Request, out tuiRuntimeOutput, model, owner, outerSession, configDir string, body []byte) (int, string) {
	if b == nil {
		return out.Fail(http.StatusBadGateway, "the Claude TUI bridge is not running in this gateway")
	}
	return b.rt.serveLite(tuiRuntimeGatewayResponder{out}, r, model, owner, outerSession, configDir, body)
}

// WithSearch marks a request made by the gateway's searcher. It uses
// Claude Code's native WebSearch instead of calling the gateway again.
func tuiRuntimeWithSearch(ctx context.Context) context.Context {
	return context.WithValue(ctx, searchingKey{}, true)
}

// ServeMCP answers a run's MCP helper calling back with token, and says
// whether the token was one of this bridge's runs.
func (b *tuiRuntimeBridge) ServeMCP(w http.ResponseWriter, r *http.Request, token string) bool {
	if b == nil {
		return false
	}
	b.rt.bridge.mu.Lock()
	_, ok := b.rt.bridge.runs[token]
	b.rt.bridge.mu.Unlock()
	if !ok {
		return false
	}
	b.rt.bridge.mcpCall(w, r.WithContext(tuiRuntimeWithMCPToken(r.Context(), token)))
	return true
}

// Busy says a run is under way: a turn, or a call parked on the client.
func (b *tuiRuntimeBridge) Busy() bool {
	if b == nil {
		return false
	}
	b.rt.bridge.mu.Lock()
	defer b.rt.bridge.mu.Unlock()
	return len(b.rt.bridge.runs) > 0
}

// Parked counts live runs waiting for client tools between HTTP replies.
func (b *tuiRuntimeBridge) Parked() int {
	if b == nil {
		return 0
	}
	b.rt.bridge.mu.Lock()
	runs := make([]*tuiRuntimeSubscriptionRun, 0, len(b.rt.bridge.runs))
	for _, run := range b.rt.bridge.runs {
		runs = append(runs, run)
	}
	b.rt.bridge.mu.Unlock()
	n := 0
	for _, run := range runs {
		run.mu.Lock()
		if !run.closed && !run.parkedAt.IsZero() && run.httpConsumers == 0 && run.segment == nil && !run.resuming {
			n++
		}
		run.mu.Unlock()
	}
	return n
}

func (r *tuiRuntimeSubscriptionRun) releaseHTTP() {
	r.mu.Lock()
	r.httpConsumers--
	r.mu.Unlock()
}

// Shutdown ends every run, each by its wrapper's own teardown first.
func (b *tuiRuntimeBridge) Shutdown() {
	if b == nil {
		return
	}
	b.rt.bridge.abortAll()
}

// ErrorStatus is the status a failed turn is answered with: Claude Code's
// own when it said one, else 429 when its words say the account hit its
// limit, so clients wait rather than retry at once.
func tuiRuntimeErrorStatus(status int, msg string) int {
	if status < 400 && quotaWords.MatchString(msg) {
		status = http.StatusTooManyRequests
	}
	return tuiRuntimeLiteStatus(status, msg)
}

// Owner is the owner of a TUI run on account user of provider id, under
// MCP server name: a conversation's checkpoint is never taken up by
// another account, nor across a change of the server's name. The account
// is second, where ownerAccount reads it.
func tuiRuntimeOwner(id, user, mcpServer string) string {
	return strings.Join([]string{"claude-tui", user, "mcp=" + mcpServer, id}, "\x00")
}

type tuiRuntimeMcpTokenKey struct{}

func tuiRuntimeWithMCPToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tuiRuntimeMcpTokenKey{}, token)
}

// ResetForTest gives the bridge its defaults again, as a test that called
// Init leaves it.
func resetTUIRuntime() {
	tuiRuntimeSettings = &tuiRuntimeConfig{
		MCPServer:        "magpie",
		NoteLimits:       func(string, []tuiRuntimeClaudeLimit) {},
		NoteSignInFailed: func(string, string, string) {},
		MovedOff:         func(string) bool { return false },
		LoginVersion:     func(string) string { return "" },
		CallbackBase:     func() string { return "http://127.0.0.1" },
		AddCallback:      func(string) {},
		DropCallback:     func(string) {},
		Room:             func(context.Context, string) (func(bool), error) { return func(bool) {}, nil },
		Leave:            func() {},
		Env: func(_ context.Context, env []string, configDir string) ([]string, error) {
			return tuiRuntimeInClaudeDir(env, configDir), nil
		},
	}
}
