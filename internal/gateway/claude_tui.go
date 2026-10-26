package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/yetone/magpie/internal/appdir"
	"github.com/yetone/magpie/internal/steady"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/yetone/magpie/internal/claudebridge"
	"github.com/yetone/magpie/internal/claudecode"
	"github.com/yetone/magpie/internal/netproxy"
	"github.com/yetone/magpie/internal/plugin"
	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/settings"
)

// A Claude subscription account set to TUI (provider.ClaudeTUI) is answered
// by the Claude TUI bridge (internal/claudetui): the interactive Claude
// Code in a private tmux, its tools the client's over magpie's MCP helper
// as on the standard bridge. Which accounts are set so is read once, when
// the gateway starts serving (tuiModes): a change takes effect at its next
// start. A TUI account that can't be run here fails its request with why;
// it is never answered by the standard bridge in its place, though routing
// still goes on to the group's next member as for any failure.

// tuiModes are the accounts set to TUI when the gateway started, by
// provider id and account name in lower case.
var tuiModes struct {
	once sync.Once
	m    map[string]map[string]bool
}

// ClaudeTUILive says the Claude account user of provider id runs on the
// Claude TUI in this gateway: as it was set when the gateway started.
func ClaudeTUILive(id, user string) bool {
	tuiModes.once.Do(readTUIModes)
	return tuiModes.m[id][strings.ToLower(strings.TrimSpace(user))]
}

func readTUIModes() {
	tuiModes.m = map[string]map[string]bool{}
	for _, p := range provider.All() {
		if p.Account == nil || p.Account.Agent != "claude" {
			continue
		}
		for u, mode := range p.AccountClaudeModes {
			if mode != provider.ClaudeTUI {
				continue
			}
			if tuiModes.m[p.ID] == nil {
				tuiModes.m[p.ID] = map[string]bool{}
			}
			tuiModes.m[p.ID][strings.ToLower(strings.TrimSpace(u))] = true
		}
	}
}

// TUILive is what the gateway serving runs each Claude account as: the
// accounts set to TUI when it started (Modes, by provider id and account
// name in lower case), by its pid.
type TUILive struct {
	PID   int                        `json:"pid"`
	Modes map[string]map[string]bool `json:"modes"`
}

func tuiLivePath() string { return filepath.Join(appdir.Config(), "claude-tui", "live.json") }

// writeTUILive says, for the CLI and a window of another process, what this
// gateway runs each Claude account as.
func writeTUILive() {
	tuiModes.once.Do(readTUIModes)
	b, _ := json.Marshal(TUILive{PID: os.Getpid(), Modes: tuiModes.m})
	if err := os.MkdirAll(filepath.Dir(tuiLivePath()), 0o700); err == nil {
		err = steady.WriteFile(tuiLivePath(), b, 0o600)
		if err != nil {
			log.Printf("claude-tui: %v", err)
		}
	}
}

// ReadTUILive is what the gateway serving runs each Claude account as;
// false when no gateway is serving, or it can't be read.
func ReadTUILive() (TUILive, bool) {
	var l TUILive
	b, err := steady.ReadFile(tuiLivePath())
	if err != nil || json.Unmarshal(b, &l) != nil || l.PID <= 0 {
		return TUILive{}, false
	}
	if p, err := os.FindProcess(l.PID); err != nil || p.Signal(syscall.Signal(0)) != nil {
		return TUILive{}, false
	}
	return l, true
}

// Runs says the live gateway runs the Claude account user of provider id
// on the TUI.
func (l TUILive) Runs(id, user string) bool {
	return l.Modes[id][strings.ToLower(strings.TrimSpace(user))]
}

// anyTUI says an account was set to TUI when the gateway started.
func anyTUI() bool {
	tuiModes.once.Do(readTUIModes)
	return len(tuiModes.m) > 0
}

// tuiReady is the Claude TUI's setup: its wrapper written and what it
// needs found, or why it can't run, once per process.
var tuiReady struct {
	once sync.Once
	err  error
}

// tuiServer is the gateway serving, whose web search a TUI run's model
// searches with.
var tuiServer atomic.Pointer[Server]

// claudeTUIBun is the Bun TUI runs use; tests set their own.
var claudeTUIBun = func() (string, error) { return plugin.BunPath(tuiRuntimeMinBun) }

// setUpClaudeTUI gives the bridge the gateway's own: its room for Claude
// Code processes, its callbacks, the account's proxy and config folder,
// its MCP server name, and writes the wrapper, once; what is missing is
// kept, and said on every TUI request.
func setUpClaudeTUI(ctx context.Context) error {
	tuiReady.once.Do(func() {
		tuiReady.err = func() error {
			if err := tuiRuntimeSupported(); err != nil {
				return err
			}
			if err := claudebridge.LoadMCPConfig(); err != nil {
				return err
			}
			initTUIRuntime(tuiConfig())
			deps, err := tuiRuntimeCheck(ctx, claudeTUIPath, claudeTUIBun)
			if err != nil {
				return err
			}
			bundle, err := tuiRuntimeInstallWrapper()
			if err != nil {
				return err
			}
			exe, err := os.Executable()
			if err != nil {
				return err
			}
			initTUIRuntime(&tuiRuntimeConfig{
				Tmux:       deps.Tmux,
				Wrapper:    []string{deps.Bun, bundle},
				HelperArgv: []string{exe, "claude-mcp-helper"},
			})
			tuiRuntimeSweep()
			log.Printf("claude-tui: ready: tmux %s, Claude Code %s, Bun %s, MCP server %q", deps.TmuxVersion, deps.ClaudeVersion, deps.Bun, claudebridge.MCPServerName())
			return nil
		}()
		if tuiReady.err != nil {
			log.Printf("claude-tui: can't run here: %v", tuiReady.err)
		}
	})
	return tuiReady.err
}

// claudeTUIPath is the Claude Code TUI runs start: this machine's, else
// the one magpie downloaded. Never one in WSL.
func claudeTUIPath() (string, error) {
	if p := claudecode.Find(); p != "" {
		return p, nil
	}
	return "", errors.New("Claude Code is not installed: install it, " + claudecode.InstallHint)
}

func tuiConfig() *tuiRuntimeConfig {
	budget, _ := claudeRunBudget()
	if mib := settings.Load().ClaudeTuiMaxRSSMiB; mib > 0 {
		budget = int64(mib) << 20
	}
	return &tuiRuntimeConfig{
		ClaudePath: claudeTUIPath,
		MCPServer:  claudebridge.MCPServerName(),
		CallbackBase: func() string {
			return callbackBaseURL()
		},
		AddCallback:  addCallback,
		DropCallback: dropCallback,
		Room: func(ctx context.Context, what string) (func(bool), error) {
			return claudeRoom.take(ctx, what)
		},
		Leave: claudeRoom.leave,
		Env: func(ctx context.Context, env []string, configDir string) ([]string, error) {
			env, err := netproxy.EnvForBunWith(claudeProxy(ctx), env, true)
			if err != nil {
				return nil, err
			}
			return claudeCLI{path: "claude"}.asRoot(inClaudeDir(env, configDir)), nil
		},
		MaxRSS:    budget,
		CanSearch: canSearch,
		SearchHandler: func(ctx context.Context, query string) (string, []Hit, error) {
			s := tuiServer.Load()
			if s == nil {
				return "", nil, errNoSearcher
			}
			return s.webSearch(ctx, query)
		},
		NoteLimits: provider.NoteClaudeLimits,
		MovedOff:   provider.ClaudeCodeMovedOff,
	}
}

// serveClaudeTUI answers a request on a Claude account set to TUI, in any
// of the APIs the gateway takes.
func (s *Server) serveClaudeTUI(w http.ResponseWriter, r *http.Request, from provider.Protocol, p provider.Provider, model string, body []byte, usage *Usage) (int, string) {
	fail := func(code int, msg string) (int, string) {
		return writeError(w, from, code, "Claude TUI: "+msg), msg
	}
	if err := setUpClaudeTUI(r.Context()); err != nil {
		return fail(http.StatusBadGateway, err.Error())
	}
	if wantsClientCache(r, from) {
		// its runs keep Claude Code's own cache slots: a client's marks
		// can't be placed, and are refused rather than dropped unsaid
		return fail(http.StatusBadRequest, "client cache marks (X-Magpie-Cache-Marks: client) are not taken by the Claude TUI; send the request without them, or set the account to standard")
	}
	user := p.Account.User
	ctx := p.Via(r.Context())
	if searching(ctx) {
		ctx = tuiRuntimeWithSearch(ctx)
	}
	dir, err := tuiAccountDir(ctx, p)
	if err != nil {
		return fail(http.StatusBadGateway, "the account "+user+" can't run on the Claude TUI: "+err.Error()+". Sign it in through magpie (magpie accounts add claude), or set it to standard")
	}
	req, err := parse(from, body)
	if err != nil {
		return writeError(w, from, 400, err.Error()), err.Error()
	}
	req.Model = model
	anthropic := body
	if from != provider.Anthropic {
		anthropic = tuiMessages(req, model)
	}
	owner := p.ID + "\x00" + user
	out := &tuiResponder{w: w, r: r, from: from, req: req, usage: usage, owner: owner}
	return s.tui.Serve(r.WithContext(ctx), out, model, tuiRuntimeOwner(p.ID, user, claudebridge.MCPServerName()), sessionOf(r.Header), dir, anthropic)
}

// tuiAccountDir is the config folder a TUI run on p's account runs in: its
// saved sign-in's, a folder of magpie's, never Claude Code's own home.
// Tests give their own.
var tuiAccountDir = func(ctx context.Context, p provider.Provider) (string, error) {
	if p.Account.AgentsOwn() {
		return "", errors.New("it is the account Claude Code itself is signed in to, in your home, whose sign-in the Claude TUI doesn't share")
	}
	dir, _, err := p.Account.Token(ctx)
	return dir, err
}

// tuiMessages is req as a Messages request, as the bridge reads one. The
// stand-in tools buildAnthropic declares for a history's calls with no
// tools offered are left out: a run would offer them as real ones.
func tuiMessages(req *Request, model string) []byte {
	b := buildAnthropic(req, model)
	if len(req.Tools) > 0 || req.WebSearch {
		return b
	}
	var m map[string]json.RawMessage
	if json.Unmarshal(b, &m) != nil {
		return b
	}
	delete(m, "tools")
	delete(m, "tool_choice")
	if out, err := json.Marshal(m); err == nil {
		return out
	}
	return b
}

// tuiResponder answers a TUI request in its client's protocol, through
// the gateway's own relay: its usage counted, its errors classed, the
// account's allowance said as Claude Code's are.
type tuiResponder struct {
	w     http.ResponseWriter
	r     *http.Request
	from  provider.Protocol
	req   *Request
	usage *Usage
	owner string
}

func (t *tuiResponder) Header() http.Header { return t.w.Header() }

func (t *tuiResponder) Fail(status int, msg string) (int, string) {
	return writeError(t.w, t.from, status, msg), msg
}

func (t *tuiResponder) Relay(events <-chan Event, _ bool, abort func(), ended func(said, stop string, ok bool)) (int, string) {
	limitHeaders(t.w, t.owner)
	return relay(t.w, t.r, t.from, "Claude Code", t.req, events, t.usage, abort, func() { limitHeaders(t.w, t.owner) }, ended)
}

// claudeMCP answers a run's MCP helper: the TUI bridge's runs' and the
// standard bridge's.
func (s *Server) claudeMCP(w http.ResponseWriter, r *http.Request) {
	if s.tui.ServeMCP(w, r, r.PathValue("token")) {
		return
	}
	s.subscription.mcpCall(w, r)
}

// startClaudeTUI readies the Claude TUI as the gateway starts serving,
// when an account is set to it: what it needs is checked now, and said in
// the log, rather than on the first request.
func (s *Server) startClaudeTUI(ctx context.Context) {
	writeTUILive()
	if !anyTUI() {
		return
	}
	_ = setUpClaudeTUI(ctx)
}
