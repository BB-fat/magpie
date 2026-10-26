package provider

import (
	"fmt"
	"runtime"
	"strings"
)

// A Claude subscription account is run one of two ways: by the bridge to
// Claude Code's stream-json (ClaudeStandard, as always), or by the Claude
// TUI bridge, the interactive Claude Code in a private tmux (ClaudeTUI,
// Linux only). Its provider, models, routing and allowance are the same
// either way. The gateway reads the modes when it starts: a change takes
// effect once it is restarted.
const (
	ClaudeStandard = "standard"
	ClaudeTUI      = "tui"
)

// ClaudeModeOf is how the gateway runs the Claude account user of p:
// ClaudeStandard unless it is set to ClaudeTUI.
func (p Provider) ClaudeModeOf(user string) string {
	m := p.AccountClaudeModes
	if m == nil && p.Account != nil {
		if s, ok := storedPicks(p.ID); ok {
			m = s.AccountClaudeModes
		}
	}
	if m[accountKey(user)] == ClaudeTUI {
		return ClaudeTUI
	}
	return ClaudeStandard
}

// ClaudeTUIGOOS is the platform the Claude TUI bridge runs on; tests set
// another.
var ClaudeTUIGOOS = runtime.GOOS

// SetAccountClaudeMode sets how the gateway runs the Claude account ref of
// the provider id, from its next start: ClaudeStandard or ClaudeTUI.
func SetAccountClaudeMode(id, ref, mode string) error {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode != ClaudeStandard && mode != ClaudeTUI {
		return fmt.Errorf("a Claude account runs as standard or tui, not %q", mode)
	}
	p, err := Find(id)
	if err != nil {
		return err
	}
	if p.Account == nil || p.Account.Agent != "claude" {
		return fmt.Errorf("%s is not a Claude subscription: only a Claude account runs as standard or tui", p.Name)
	}
	r, ok := p.accountRef(ref)
	if !ok {
		return fmt.Errorf("%s has no account %q", p.Name, ref)
	}
	if mode == ClaudeTUI {
		if ClaudeTUIGOOS != "linux" {
			return fmt.Errorf("the Claude TUI runs on Linux only, not %s", ClaudeTUIGOOS)
		}
		if claudeOwnOnly(r) {
			return fmt.Errorf("%s is Claude Code's own sign-in, which the Claude TUI doesn't run: sign it in through magpie (magpie accounts add claude), which gives it a folder of its own, then set it to tui", r)
		}
	}
	m := map[string]string{}
	for k, v := range p.AccountClaudeModes {
		m[k] = v
	}
	if mode == ClaudeStandard {
		if _, had := m[accountKey(r)]; !had {
			return nil
		}
		delete(m, accountKey(r))
	} else {
		m[accountKey(r)] = mode
	}
	if len(m) == 0 {
		m = nil
	}
	p.AccountClaudeModes = m
	return Save(*p)
}

// ClaudeSavedDir is the folder of its own magpie runs the saved Claude
// account user in, its sign-in put there from magpie's saved one: never
// Claude Code's own home.
func ClaudeSavedDir(user string) (string, error) { return claudeSavedDir(user) }

// claudeOwnOnly says user is known only as Claude Code's own sign-in, with
// no saved sign-in magpie runs in a folder of its own.
func claudeOwnOnly(user string) bool {
	saved, own := false, false
	for _, l := range Logins("claude") {
		if accountKey(l.User) != accountKey(user) {
			continue
		}
		if l.Own {
			own = true
		} else {
			saved = true
		}
	}
	return own && !saved
}

// normalAccountClaudeModes is m as it is kept: keys in lower case, only
// accounts set to ClaudeTUI (ClaudeStandard is the default); nil for none.
func normalAccountClaudeModes(m map[string]string) map[string]string {
	var out map[string]string
	for u, mode := range m {
		u = accountKey(u)
		if u == "" || strings.ToLower(strings.TrimSpace(mode)) != ClaudeTUI {
			continue
		}
		if out == nil {
			out = map[string]string{}
		}
		out[u] = ClaudeTUI
	}
	return out
}
