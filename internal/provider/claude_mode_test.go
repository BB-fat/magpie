package provider

import (
	"strings"
	"testing"
)

// A saved Claude account is set to tui and back; what can't be run on the
// Claude TUI is refused, saying why.
func TestSetAccountClaudeMode(t *testing.T) {
	claudeAuthFixture(t, "")
	old := ClaudeTUIGOOS
	ClaudeTUIGOOS = "linux"
	t.Cleanup(func() { ClaudeTUIGOOS = old })

	if err := SetAccountClaudeMode("claude", "Side@Example.com", "TUI"); err != nil {
		t.Fatal(err)
	}
	p, err := Find("claude")
	if err != nil {
		t.Fatal(err)
	}
	if p.ClaudeModeOf("side@example.com") != ClaudeTUI || p.ClaudeModeOf("own@example.com") != ClaudeStandard {
		t.Fatalf("modes %v", p.AccountClaudeModes)
	}
	if p.AccountClaudeModes["side@example.com"] != ClaudeTUI || len(p.AccountClaudeModes) != 1 {
		t.Fatalf("kept as %v, want one lower-case key", p.AccountClaudeModes)
	}
	// a copy is its own
	c := p.clone()
	c.AccountClaudeModes["x"] = ClaudeTUI
	if _, shared := p.AccountClaudeModes["x"]; shared {
		t.Fatal("a clone shares its modes with the provider")
	}

	for _, c := range []struct{ id, account, mode, says string }{
		{"claude", "nobody@example.com", "tui", "has no account"},
		{"claude", "side@example.com", "fast", "standard or tui"},
		{"claude", "own@example.com", "tui", "Claude Code itself is signed in to"},
	} {
		if err := SetAccountClaudeMode(c.id, c.account, c.mode); err == nil || !strings.Contains(err.Error(), c.says) {
			t.Errorf("%s %s %s: %v, want it refused (%s)", c.id, c.account, c.mode, err, c.says)
		}
	}
	ClaudeTUIGOOS = "darwin"
	if err := SetAccountClaudeMode("claude", "side@example.com", "tui"); err == nil || !strings.Contains(err.Error(), "Linux only") {
		t.Errorf("tui on darwin: %v", err)
	}
	// standard is always taken, and clears the account's entry
	if err := SetAccountClaudeMode("claude", "side@example.com", "standard"); err != nil {
		t.Fatal(err)
	}
	if p, _ := Find("claude"); p.AccountClaudeModes != nil {
		t.Fatalf("standard kept %v", p.AccountClaudeModes)
	}
}

// Only a Claude subscription's accounts have a mode.
func TestSetAccountClaudeModeOnlyClaude(t *testing.T) {
	claudeAuthFixture(t, "")
	if err := Save(Provider{ID: "keyed", Name: "Keyed", Key: "k", Chat: "http://127.0.0.1:1/v1"}); err != nil {
		t.Fatal(err)
	}
	if err := SetAccountClaudeMode("keyed", "k", "tui"); err == nil || !strings.Contains(err.Error(), "not a Claude subscription") {
		t.Fatalf("a key provider's mode: %v", err)
	}
}

// What providers.json holds is kept as it is read: unknown modes and
// standard ones are left out.
func TestNormalAccountClaudeModes(t *testing.T) {
	got := normalAccountClaudeModes(map[string]string{" A@x ": "tui", "b@x": "standard", "c@x": "weird", "": "tui"})
	if len(got) != 1 || got["a@x"] != ClaudeTUI {
		t.Fatalf("%v", got)
	}
	if normalAccountClaudeModes(map[string]string{"b@x": "standard"}) != nil {
		t.Fatal("no tui account keeps an empty map")
	}
}
