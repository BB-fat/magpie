package claudetui

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/claudebridge"
	"github.com/yetone/magpie/internal/proc"
)

// What magpie added to the bridge it took from claude-tui-boost.

func withConfig(t *testing.T, c *Config) {
	t.Helper()
	old := config
	Init(c)
	t.Cleanup(func() { config = old })
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
		done := make(chan struct{})
		go r.watchRSS(&procTree{pid: 1}, done, time.Millisecond)
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

// A run's tree is ended whole: its wrapper's group, and what was seen
// under it in a session of its own, as tmux starts a pane.
func TestProcTreeEndsWhatRanUnderIt(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the tree's processes are read from /proc")
	}
	pidFile := filepath.Join(t.TempDir(), "pane")
	cmd := commandContext(context.Background(), "/bin/sh", "-c", "setsid sh -c 'echo $$ > "+pidFile+"; exec sleep 60' & sleep 60")
	tree, err := startTree(cmd)
	if err != nil {
		t.Fatal(err)
	}
	var pane int
	for i := 0; i < 200 && pane == 0; i++ {
		b, _ := os.ReadFile(pidFile)
		pane, _ = strconv.Atoi(strings.TrimSpace(string(b)))
		time.Sleep(10 * time.Millisecond)
	}
	if pane == 0 {
		t.Fatal("the pane stand-in never started")
	}
	if sid, _ := syscall.Getpgid(pane); sid == cmd.Process.Pid {
		t.Fatal("the stand-in pane is in the wrapper's group: it tests nothing")
	}
	// no reading has looked under the wrapper: Kill finds the pane itself
	tree.Kill()
	_ = tree.Wait()
	for i := 0; i < 200; i++ {
		if !same(pane, 0) || zombie(pane) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	_ = syscall.Kill(pane, syscall.SIGKILL)
	t.Fatal("the process in a session of its own outlived the tree")
}

// A tree's memory is its processes', a session of their own among them.
func TestProcTreeMeasuresWhatRunsUnderIt(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the tree's processes are read from /proc")
	}
	cmd := commandContext(context.Background(), "/bin/sh", "-c", "setsid sleep 60 & sleep 60")
	tree, err := startTree(cmd)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { tree.Kill(); _ = tree.Wait() }()
	var n int64
	for i := 0; i < 200; i++ {
		tree.look()
		tree.mu.Lock()
		seen := len(tree.seen)
		tree.mu.Unlock()
		if seen >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	n, ok := tree.rss()
	alone, _ := pidRSS(cmd.Process.Pid)
	if !ok || n <= alone {
		t.Fatalf("the tree holds %d, the wrapper alone %d: what runs under it isn't counted", n, alone)
	}
}

func zombie(pid int) bool {
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return true
	}
	s := string(b)
	f := strings.Fields(s[strings.LastIndexByte(s, ')')+1:])
	return len(f) > 0 && f[0] == "Z"
}

// A gateway coming up ends the tmux servers one that died left, by their
// sockets, and leaves a live gateway's and its own.
func TestSweepTmuxEndsWhatADeadGatewayLeft(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux only")
	}
	dir := t.TempDir()
	t.Setenv("TMPDIR", dir)
	log := filepath.Join(dir, "tmux.log")
	tmux := filepath.Join(dir, "tmux")
	if err := os.WriteFile(tmux, []byte("#!/bin/sh\necho \"$@\" >> "+log+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	withConfig(t, &Config{Tmux: tmux})
	d, err := tmuxDir()
	if err != nil {
		t.Fatal(err)
	}
	dead := deadPID(t)
	for _, name := range []string{fmt.Sprintf("%d-aaaaaaaaaaaa.sock", dead), fmt.Sprintf("%d-bbbbbbbbbbbb.sock", os.Getpid()), "1-cccccccccccc.sock", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(d, name), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	SweepTmux()
	b, _ := os.ReadFile(log)
	if got := strings.TrimSpace(string(b)); got != "-S "+filepath.Join(d, fmt.Sprintf("%d-aaaaaaaaaaaa.sock", dead))+" kill-server" {
		t.Fatalf("tmux was asked %q", got)
	}
	for name, kept := range map[string]bool{fmt.Sprintf("%d-aaaaaaaaaaaa.sock", dead): false, fmt.Sprintf("%d-bbbbbbbbbbbb.sock", os.Getpid()): true, "1-cccccccccccc.sock": true, "notes.txt": true} {
		if _, err := os.Stat(filepath.Join(d, name)); (err == nil) != kept {
			t.Errorf("%s kept %v, want %v", name, err == nil, kept)
		}
	}
}

// deadPID is a pid no process has now.
func deadPID(t *testing.T) int {
	cmd := proc.Command("/bin/true")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	return cmd.Process.Pid
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
