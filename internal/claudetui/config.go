package claudetui

// Package claudetui is the Claude TUI bridge: a Claude subscription account
// set to TUI mode is answered by the genuine interactive Claude Code, run
// in a private tmux by the PTY wrapper (wrapper/), with the caller's tools
// parked over magpie's MCP helper. It is ported from claude-tui-boost
// v0.2.16 (f57ee457cf7f41ecc18ed8284519ab565f46ad16, MIT; see NOTICE.md),
// itself a fork of magpie's own subscription bridge. The gateway embeds it
// in-process: there is no runtime of its own, no HTTP surface and no key.

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/yetone/magpie/internal/appdir"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// Config is what the gateway hands the bridge.
type Config struct {
	// StateDir holds the interactive checkpoints and tool journal (0700):
	// <magpie's config>/claude-tui when unset.
	StateDir string

	// ClaudePath finds the Claude Code binary to run in the TUI.
	ClaudePath func() (string, error)

	// Wrapper is the PTY wrapper's argv: the absolute Bun and the bundle
	// it runs. Empty disables the interactive path.
	Wrapper []string

	// Tmux is the tmux the wrapper runs its private server with.
	Tmux string

	// HelperArgv runs magpie's MCP helper (magpie claude-mcp-helper); the
	// callback file and tools file are appended.
	HelperArgv []string

	// MCPServer is the MCP server name (MAGPIE_CLAUDE_MCP_NAME, already
	// checked), the tool prefix mcp__<MCPServer>__ with it.
	MCPServer string

	// CallbackBase is the gateway's own address the helper calls back on.
	CallbackBase func() string
	// AddCallback and DropCallback make a run's token one the gateway lets
	// in without a key, while the run lives.
	AddCallback, DropCallback func(token string)

	// Room takes a place in the gateway's room for Claude Code processes
	// (claudeRoom) before a run's wrapper starts; the func it gives frees
	// it once the whole process group has ended, left saying magpie ended
	// it. Leave says a run magpie asked to end is ending.
	Room  func(ctx context.Context, what string) (func(left bool), error)
	Leave func()

	// Env completes a run's environment: the account's proxy, its Claude
	// config folder (configDir, "" for Claude Code's own) and IS_SANDBOX
	// as root, as the gateway gives its own runs.
	Env func(ctx context.Context, env []string, configDir string) ([]string, error)

	// MaxRSS is the most a run's process group may hold resident, in
	// bytes, 0 for no bound; RSS reads a group's resident memory.
	MaxRSS int64
	RSS    func(pid int) (int64, bool)

	// SearchHandler answers the run's web_search tool; nil offers none.
	SearchHandler func(ctx context.Context, query string) (string, []Hit, error)

	// NoteLimits, NoteSignInFailed, MovedOff and LoginVersion are the
	// provider package's for the account a run is on.
	NoteLimits       func(user string, limits []ClaudeLimit)
	NoteSignInFailed func(user, version, message string)
	MovedOff         func(user string) bool
	LoginVersion     func(user string) string
	// HandedOver is called when a saved account was moved over to Claude
	// Code's own sign-in.
	HandedOver func(user string)
}

// ClaudeLimit is one allowance window as Claude Code tells it, in the
// shape of provider.ClaudeLimit.
type ClaudeLimit struct {
	Kind     string
	Used     float64
	ResetsAt int64
}

var config *Config

func init() { ResetForTest() }

// Init installs cfg's set fields over the defaults. The gateway calls it
// once, before it serves.
func Init(cfg *Config) {
	if cfg == nil {
		return
	}
	c := *config
	if cfg.StateDir != "" {
		c.StateDir = cfg.StateDir
	}
	if cfg.ClaudePath != nil {
		c.ClaudePath = cfg.ClaudePath
	}
	if len(cfg.Wrapper) > 0 {
		c.Wrapper = slices.Clone(cfg.Wrapper)
	}
	if cfg.Tmux != "" {
		c.Tmux = cfg.Tmux
	}
	if len(cfg.HelperArgv) > 0 {
		c.HelperArgv = slices.Clone(cfg.HelperArgv)
	}
	if cfg.MCPServer != "" {
		c.MCPServer = cfg.MCPServer
	}
	if cfg.CallbackBase != nil {
		c.CallbackBase = cfg.CallbackBase
	}
	if cfg.AddCallback != nil {
		c.AddCallback = cfg.AddCallback
	}
	if cfg.DropCallback != nil {
		c.DropCallback = cfg.DropCallback
	}
	if cfg.Room != nil {
		c.Room = cfg.Room
	}
	if cfg.Leave != nil {
		c.Leave = cfg.Leave
	}
	if cfg.Env != nil {
		c.Env = cfg.Env
	}
	if cfg.MaxRSS != 0 {
		c.MaxRSS = cfg.MaxRSS
	}
	if cfg.RSS != nil {
		c.RSS = cfg.RSS
	}
	if cfg.SearchHandler != nil {
		c.SearchHandler = cfg.SearchHandler
	}
	if cfg.NoteLimits != nil {
		c.NoteLimits = cfg.NoteLimits
	}
	if cfg.NoteSignInFailed != nil {
		c.NoteSignInFailed = cfg.NoteSignInFailed
	}
	if cfg.MovedOff != nil {
		c.MovedOff = cfg.MovedOff
	}
	if cfg.LoginVersion != nil {
		c.LoginVersion = cfg.LoginVersion
	}
	if cfg.HandedOver != nil {
		c.HandedOver = cfg.HandedOver
	}
	config = &c
}

// mcpToolPrefix is mcp__<server>__, how Claude Code names the helper's
// tools.
func mcpToolPrefix() string { return "mcp__" + config.MCPServer + "__" }

// stateDir is the bridge's private state folder, made 0700.
func stateDir() (string, error) {
	d := config.StateDir
	if d == "" {
		d = filepath.Join(appdir.Config(), "claude-tui")
	}
	if err := os.MkdirAll(d, 0o700); err != nil {
		return "", err
	}
	return d, nil
}

// mustStateDir is stateDir that can't fail: without a state folder the
// checkpoint goes to a temp one, so a turn is rebuilt rather than resumed
// from a checkpoint that isn't kept.
func mustStateDir() string {
	if d, err := stateDir(); err == nil {
		return d
	}
	d := filepath.Join(os.TempDir(), "magpie-claude-tui-"+strconv.Itoa(os.Getuid()))
	_ = os.MkdirAll(d, 0o700)
	return d
}

// claudeConfigDir is Claude Code's own config folder, for a run on the
// account Claude Code itself is signed in to (configDir "").
func claudeConfigDir() string {
	if d := os.Getenv("CLAUDE_CONFIG_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude")
}

// mcpHelperCommand is how a run starts magpie's MCP helper.
func mcpHelperCommand() (string, []string, error) {
	if len(config.HelperArgv) == 0 {
		exe, err := os.Executable()
		if err != nil {
			return "", nil, err
		}
		return exe, []string{"claude-mcp-helper"}, nil
	}
	return config.HelperArgv[0], slices.Clone(config.HelperArgv[1:]), nil
}

// claudeBinary is the Claude Code a run starts in its TUI.
func claudeBinary() (string, error) {
	if config.ClaudePath == nil {
		return "", errors.New("Claude Code is not installed; install it and run `claude auth login`")
	}
	return config.ClaudePath()
}

// callbackBaseURL is the gateway address the MCP helper calls back on.
func callbackBaseURL() string { return config.CallbackBase() }

// transcriptOff says why Claude Code would keep no transcript of a run in
// configDir, "" when it keeps one: the bridge reads every reply from it,
// so a run can't go on without it, and is refused with why rather than
// started with the setting taken away unsaid.
func transcriptOff(env []string, configDir string) string {
	for _, e := range env {
		if k, v, _ := strings.Cut(e, "="); k == "CLAUDE_CODE_SKIP_PROMPT_HISTORY" && v != "" && v != "0" && v != "false" {
			return "CLAUDE_CODE_SKIP_PROMPT_HISTORY is set in magpie's environment, which keeps Claude Code from writing the transcript the Claude TUI reads its replies from; unset it, or set the account to standard"
		}
	}
	if configDir == "" {
		configDir = claudeConfigDir()
	}
	b, err := os.ReadFile(filepath.Join(configDir, "settings.json"))
	if err != nil {
		return ""
	}
	var s struct {
		CleanupPeriodDays *int `json:"cleanupPeriodDays"`
	}
	if json.Unmarshal(b, &s) == nil && s.CleanupPeriodDays != nil && *s.CleanupPeriodDays == 0 {
		return "the account's settings.json sets cleanupPeriodDays to 0, which keeps Claude Code from writing the transcript the Claude TUI reads its replies from; set it above 0, or set the account to standard"
	}
	return ""
}
