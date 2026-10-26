package claudetui

import (
	"context"
	"net/http"
	"strings"
)

// runtimeTUI owns the bridge a gateway's TUI requests are answered by.
type runtimeTUI struct {
	bridge *subscriptionBridge
}

// responder answers the client: a refusal before any reply, or a reply's
// events. The gateway's writes the client's own protocol and counts the
// usage; anthropicResponder writes Messages, as the plugin's runtime did.
type responder interface {
	Header() http.Header
	Fail(status int, msg string) (int, string)
	Relay(req *Request, events <-chan Event, abort func(), ended func(said, stop string, ok bool)) (int, string)
}

// Responder is how the gateway answers a TUI request's client.
type Responder interface {
	// Header is the reply's headers, set before Fail or Relay.
	Header() http.Header
	// Fail answers status with msg before any reply.
	Fail(status int, msg string) (int, string)
	// Relay answers with events, as the gateway's relay does: abort when
	// there is no answer to give, ended once one was given.
	Relay(events <-chan Event, stream bool, abort func(), ended func(said, stop string, ok bool)) (int, string)
}

type gatewayResponder struct{ Responder }

func (g gatewayResponder) Relay(req *Request, events <-chan Event, abort func(), ended func(said, stop string, ok bool)) (int, string) {
	return g.Responder.Relay(events, req.Stream, abort, ended)
}

// Bridge is a gateway's Claude TUI bridge.
type Bridge struct{ rt *runtimeTUI }

// New makes a bridge. Init has set its Config.
func New() *Bridge {
	return &Bridge{rt: &runtimeTUI{bridge: newSubscriptionBridge()}}
}

// Serve answers body, an Anthropic Messages request for model on the
// account owner names, through the interactive Claude Code run in the
// account's config folder configDir: owner is part of every checkpoint's
// identity, outerSession the conversation the client names ("" for a
// request of its own).
func (b *Bridge) Serve(r *http.Request, out Responder, model, owner, outerSession, configDir string, body []byte) (int, string) {
	if b == nil {
		return out.Fail(http.StatusBadGateway, "the Claude TUI bridge is not running in this gateway")
	}
	return b.rt.serveLite(gatewayResponder{out}, r, model, owner, outerSession, configDir, body)
}

// ServeMCP answers a run's MCP helper calling back with token, and says
// whether the token was one of this bridge's runs.
func (b *Bridge) ServeMCP(w http.ResponseWriter, r *http.Request, token string) bool {
	if b == nil {
		return false
	}
	b.rt.bridge.mu.Lock()
	_, ok := b.rt.bridge.runs[token]
	b.rt.bridge.mu.Unlock()
	if !ok {
		return false
	}
	b.rt.bridge.mcpCall(w, r.WithContext(withMCPToken(r.Context(), token)))
	return true
}

// Busy says a run is under way: a turn, or a call parked on the client.
func (b *Bridge) Busy() bool {
	if b == nil {
		return false
	}
	b.rt.bridge.mu.Lock()
	defer b.rt.bridge.mu.Unlock()
	return len(b.rt.bridge.runs) > 0
}

// Shutdown ends every run, each by its wrapper's own teardown first.
func (b *Bridge) Shutdown() {
	if b == nil {
		return
	}
	b.rt.bridge.abortAll()
}

// ErrorStatus is the status a failed turn is answered with: Claude Code's
// own when it said one, else 429 when its words say the account hit its
// limit, so clients wait rather than retry at once.
func ErrorStatus(status int, msg string) int {
	if status < 400 && quotaWords.MatchString(msg) {
		status = http.StatusTooManyRequests
	}
	return liteStatus(status, msg)
}

// Owner is the owner of a TUI run on account user of provider id, under
// MCP server name: a conversation's checkpoint is never taken up by
// another account, nor across a change of the server's name. The account
// is second, where ownerAccount reads it.
func Owner(id, user, mcpServer string) string {
	return strings.Join([]string{"claude-tui", user, "mcp=" + mcpServer, id}, "\x00")
}

type mcpTokenKey struct{}

func withMCPToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, mcpTokenKey{}, token)
}
