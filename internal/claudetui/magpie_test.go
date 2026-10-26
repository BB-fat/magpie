package claudetui

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/claudebridge"
)

// What magpie added to the bridge it took from claude-tui-boost.

func withConfig(t *testing.T, c *Config) {
	t.Helper()
	old := config
	Init(c)
	t.Cleanup(func() { config = old })
}

func fixedTUIWorkDirs() []struct {
	name string
	make func() (string, error)
} {
	return []struct {
		name string
		make func() (string, error)
	}{
		{"stateless", func() (string, error) {
			dir, release, err := liteStatelessWorkDir()
			defer release()
			return dir, err
		}},
		{"conversation", func() (string, error) { return interactiveSessionWorkDir("account", "session") }},
		{"tmux", tmuxDir},
		{"work", claudeWorkDir},
	}
}

func TestTUIFixedWorkDirsArePrivate(t *testing.T) {
	for _, work := range fixedTUIWorkDirs() {
		t.Run(work.name, func(t *testing.T) {
			withConfig(t, &Config{TempRoot: t.TempDir()})
			dir, err := work.make()
			if err != nil {
				t.Fatal(err)
			}
			again, err := work.make()
			if err != nil || again != dir {
				t.Fatalf("fixed cwd changed from %q to %q: %v", dir, again, err)
			}
			info, err := os.Lstat(dir)
			if err != nil || !info.IsDir() || !ownedByMe(info) || othersCanWrite(info) {
				t.Fatalf("unsafe cwd %s: %v, %v", dir, info, err)
			}
			if runtime.GOOS != "windows" && info.Mode().Perm() != 0o700 {
				t.Fatalf("cwd permissions %v, want 0700", info.Mode().Perm())
			}
		})
	}
}

func TestTUIFixedWorkDirsRefuseSymlinks(t *testing.T) {
	for _, work := range fixedTUIWorkDirs() {
		t.Run(work.name, func(t *testing.T) {
			withConfig(t, &Config{TempRoot: t.TempDir()})
			dir, err := work.make()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(dir); err != nil {
				t.Fatal(err)
			}
			target := t.TempDir()
			if err := os.Symlink(target, dir); err != nil {
				if runtime.GOOS == "windows" {
					t.Skipf("symlink creation unavailable: %v", err)
				}
				t.Fatal(err)
			}
			if got, err := work.make(); err == nil || got != "" || !strings.Contains(err.Error(), "not a folder") {
				t.Fatalf("symlink cwd accepted: %q, %v", got, err)
			}
			if entries, err := os.ReadDir(target); err != nil || len(entries) != 0 {
				t.Fatalf("symlink target changed: %v, %v", entries, err)
			}
		})
	}
}

func TestLiteStatelessWorkDirUsesUIDAndReleasesSlots(t *testing.T) {
	withConfig(t, &Config{TempRoot: t.TempDir()})
	first, release, err := liteStatelessWorkDir()
	defer release()
	if err != nil {
		t.Fatal(err)
	}
	prefix := fmt.Sprintf("magpie-claude-stateless-%d-", os.Getuid())
	if !strings.HasPrefix(filepath.Base(first), prefix) {
		t.Fatalf("stateless cwd %s lacks user namespace %s", first, prefix)
	}
	second, releaseSecond, err := liteStatelessWorkDir()
	defer releaseSecond()
	if err != nil || second == first {
		t.Fatalf("concurrent stateless runs shared cwd %q: %v", second, err)
	}
	release()
	again, releaseAgain, err := liteStatelessWorkDir()
	defer releaseAgain()
	if err != nil || again != first {
		t.Fatalf("released slot changed from %q to %q: %v", first, again, err)
	}
	release() // A second release must leave the new borrower holding its slot.
	third, releaseThird, err := liteStatelessWorkDir()
	defer releaseThird()
	if err != nil || third == first || third == second {
		t.Fatalf("busy slot reused: %q, %v", third, err)
	}
}

// The MCP server's name is the configured one wherever the run names it:
// its --mcp-config, the image tool it is told of, and the calls a rebuilt
// session's history holds, each under the name it was offered.
func TestMCPServerNameOnTheTUIPath(t *testing.T) {
	t.Setenv("MAGPIE_CLAUDE_MCP_NAME", "corp-tools")
	claudebridge.ResetMCPConfigForTest()
	t.Cleanup(claudebridge.ResetMCPConfigForTest)
	withConfig(t, &Config{MCPServer: claudebridge.MCPServerName()})

	var cfg struct {
		MCPServers map[string]struct {
			Args []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal([]byte(interactiveMCPConfig("/m", []string{"claude-mcp-helper"}, "/run/callback", "/run/tools.json")), &cfg); err != nil {
		t.Fatal(err)
	}
	s, ok := cfg.MCPServers["corp-tools"]
	if !ok || len(cfg.MCPServers) != 1 {
		t.Fatalf("mcp servers %v, want corp-tools alone", cfg.MCPServers)
	}
	if strings.Join(s.Args, " ") != "claude-mcp-helper /run/callback /run/tools.json" {
		t.Fatalf("helper args %v: the callback is a file, the token on no command line", s.Args)
	}
	prompt, _, err := flattenInteractiveBlocksWithImages([]map[string]any{{"type": "image", "source": map[string]any{"type": "base64", "media_type": "image/png", "data": "AAAA"}}})
	if err != nil || !strings.Contains(prompt, "mcp__corp-tools__"+interactiveImageTool) || strings.Contains(prompt, "mcp__magpie__") {
		t.Fatalf("image prompt %q, %v", prompt, err)
	}

	long := strings.Repeat("very_long_tool_name_", 4)
	tools := bridgeTools(&Request{Tools: []Tool{{Name: "Bash"}, {Name: "spawn.agent"}, {Name: long}}})
	offered := map[string]bool{}
	for _, tl := range tools {
		offered[tl.Name] = true
		if full := "mcp__corp-tools__" + tl.Name; len(full) > 64 {
			t.Errorf("%s is longer than Claude Code's 64", full)
		}
	}
	if !offered["Bash"] || offered["spawn.agent"] || offered[long] {
		t.Fatalf("offered %v: Bash as it is, the others encoded", offered)
	}
	r := &subscriptionRun{tools: map[string]bool{"Bash": true, "spawn.agent": true, long: true}}
	for _, name := range []string{"Bash", "spawn.agent", long} {
		if got := r.declaredLocked(claudebridge.ToolName(name)); got != name {
			t.Errorf("%s came back as %s", claudebridge.ToolName(name), got)
		}
	}

	// a rebuilt session's history names its calls as the run offers them
	work := t.TempDir()
	cfgDir := t.TempDir()
	id, path, err := writeClaudeTranscript(work, cfgDir, "claude-sonnet-5", nil, []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "go"}}},
		{Role: "assistant", Parts: []Part{{Kind: ToolCall, ID: "toolu_1", Name: "spawn.agent", Args: json.RawMessage(`{}`)}}},
		{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "toolu_1", Text: "done"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "ok"}}},
	}, mcpToolPrefix(), "")
	if err != nil || id == "" {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	if !strings.Contains(string(b), `"name":"mcp__corp-tools__`+claudebridge.ToolName("spawn.agent")+`"`) {
		t.Fatalf("history names the call %s", b)
	}
}

// A run whose processes hold more than the budget is ended with an error
// saying so; one whose memory can't be read is not.
func TestRunPastItsMemoryIsEnded(t *testing.T) {
	withConfig(t, &Config{MaxRSS: 100 << 20})
	old := readRSS
	t.Cleanup(func() { readRSS = old })
	for _, c := range []struct {
		rss   int64
		ok    bool
		ended bool
	}{{50 << 20, true, false}, {0, false, false}, {150 << 20, true, true}} {
		readRSS = func(*procTree) (int64, bool) { return c.rss, c.ok }
		r := &subscriptionRun{model: "m", pending: map[string]chan mcpToolResult{}, tools: map[string]bool{}}
		events := r.attach()
		done, watched := make(chan struct{}), make(chan struct{})
		go func() { r.watchRSS(&procTree{pid: 1}, done, time.Millisecond); close(watched) }()
		var got []Event
		timeout := time.After(200 * time.Millisecond)
	read:
		for {
			select {
			case ev, open := <-events:
				if !open {
					break read
				}
				got = append(got, ev)
			case <-timeout:
				break read
			}
		}
		close(done)
		<-watched // read no more before the next reading is put in
		r.mu.Lock()
		closed := r.closed
		r.mu.Unlock()
		if closed != c.ended {
			t.Errorf("rss %d MiB (read %v): ended %v", c.rss>>20, c.ok, closed)
		}
		if c.ended && (len(got) == 0 || got[0].Kind != KError || !strings.Contains(got[0].Text, "claudeTuiMaxRSSMiB")) {
			t.Errorf("an ended run said %+v", got)
		}
		if !c.ended {
			r.abort()
		}
	}
}

// A run whose Claude Code would keep no transcript is refused, saying why.
func TestTranscriptOffIsSaid(t *testing.T) {
	dir := t.TempDir()
	if why := transcriptOff([]string{"PATH=/bin"}, dir); why != "" {
		t.Fatalf("%q", why)
	}
	if why := transcriptOff([]string{"CLAUDE_CODE_SKIP_PROMPT_HISTORY=1"}, dir); !strings.Contains(why, "CLAUDE_CODE_SKIP_PROMPT_HISTORY") {
		t.Fatalf("%q", why)
	}
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(`{"cleanupPeriodDays":0}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if why := transcriptOff(nil, dir); !strings.Contains(why, "cleanupPeriodDays") {
		t.Fatalf("%q", why)
	}
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(`{"cleanupPeriodDays":30}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if why := transcriptOff(nil, dir); why != "" {
		t.Fatalf("%q", why)
	}
}

// The wrapper is written once, whole, private, in a folder named for it.
func TestInstallWrapper(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	p, err := InstallWrapper()
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil || sha256.Sum256(b) != sha256.Sum256(wrapperBundle) {
		t.Fatalf("%s holds other bytes: %v", p, err)
	}
	if info, _ := os.Stat(p); info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("%s is %v", p, info.Mode())
	}
	if again, err := InstallWrapper(); err != nil || again != p {
		t.Fatalf("again: %s %v", again, err)
	}
	if !strings.Contains(string(wrapperBundle), "CLAUDE_BRIDGE_TMUX_SOCKET") {
		t.Fatal("the embedded wrapper is not the one magpie builds from wrapper/src: build it again (see wrapperBundle)")
	}
}

// Versions are told apart as dotted numbers.
func TestOlder(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{{"3.1", "3.2", true}, {"3.7", "3.2", false}, {"2.1.267", "2.1.268", true}, {"2.1.308", "2.1.268", false}, {"3.2", "3.2", false}, {"10.0", "9.9", false}} {
		if older(c.a, c.b) != c.want {
			t.Errorf("older(%s, %s) != %v", c.a, c.b, c.want)
		}
	}
}

// A conversation's checkpoint and folder are its account's under one MCP
// server name: after a rename, or on another account, its next turn is
// rebuilt from the request, never resumed from the other's session.
func TestOwnerKeepsNamesAndAccountsApart(t *testing.T) {
	a := Owner("claude", "me@example.com", "magpie")
	for _, other := range []string{Owner("claude", "me@example.com", "corp-tools"), Owner("claude", "you@example.com", "magpie")} {
		if interactiveSessionStateKey(a, "conv") == interactiveSessionStateKey(other, "conv") {
			t.Fatalf("%q and %q share a checkpoint", a, other)
		}
	}
	if user, own := ownerAccount(a); user != "me@example.com" || own {
		t.Fatalf("the owner names %q (own %v), not its account", user, own)
	}
}
