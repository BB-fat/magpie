package gateway

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
type tuiRuntimeConfig struct {
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

	// TempRoot holds runs' folders and tmux sockets: the system's temp
	// folder when unset (tests keep theirs in their own home).
	TempRoot string

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

	// CanSearch says whether the gateway can answer web_search now. Unset
	// uses SearchHandler's presence, for an embedded bridge with no picker.
	CanSearch func() bool
	// SearchHandler answers the run's web_search tool; nil offers none.
	SearchHandler func(ctx context.Context, query string) (string, []Hit, error)

	// NoteLimits, NoteSignInFailed, MovedOff and LoginVersion are the
	// provider package's for the account a run is on.
	NoteLimits       func(user string, limits []tuiRuntimeClaudeLimit)
	NoteSignInFailed func(user, version, message string)
	MovedOff         func(user string) bool
	LoginVersion     func(user string) string
	// HandedOver is called when a saved account was moved over to Claude
	// Code's own sign-in.
	HandedOver func(user string)

	// LeanPrompt says whether a run starts with the caller's instructions
	// as its system prompt, in place of Claude Code's (ClaudeLeanPrompt).
	LeanPrompt func() bool
}

var tuiRuntimeSettings *tuiRuntimeConfig

func init() { resetTUIRuntime() }

// Init installs cfg's set fields over the defaults. The gateway calls it
// once, before it serves.
func initTUIRuntime(cfg *tuiRuntimeConfig) {
	if cfg == nil {
		return
	}
	c := *tuiRuntimeSettings
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
	if cfg.TempRoot != "" {
		c.TempRoot = cfg.TempRoot
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
	if cfg.CanSearch != nil {
		c.CanSearch = cfg.CanSearch
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
	if cfg.LeanPrompt != nil {
		c.LeanPrompt = cfg.LeanPrompt
	}
	tuiRuntimeSettings = &c
}

// leanPrompt is LeanPrompt's answer, false when the gateway gave none.
func tuiRuntimeLeanPrompt() bool {
	return tuiRuntimeSettings.LeanPrompt != nil && tuiRuntimeSettings.LeanPrompt()
}

// tempRoot is where runs' folders and tmux sockets are made.
func tuiRuntimeTempRoot() string {
	if tuiRuntimeSettings.TempRoot != "" {
		return tuiRuntimeSettings.TempRoot
	}
	return os.TempDir()
}

// mcpToolPrefix is mcp__<server>__, how Claude Code names the helper's
// tools.
func tuiRuntimeMcpToolPrefix() string { return "mcp__" + tuiRuntimeSettings.MCPServer + "__" }

// stateDir is the bridge's private state folder, made 0700.
func tuiRuntimeStateDir() (string, error) {
	d := tuiRuntimeSettings.StateDir
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
func tuiRuntimeMustStateDir() string {
	if d, err := tuiRuntimeStateDir(); err == nil {
		return d
	}
	d := filepath.Join(tuiRuntimeTempRoot(), "magpie-claude-tui-"+strconv.Itoa(os.Getuid()))
	_ = os.MkdirAll(d, 0o700)
	return d
}

// claudeConfigDir is Claude Code's own config folder, for a run on the
// account Claude Code itself is signed in to (configDir "").
func tuiRuntimeClaudeConfigDir() string {
	if d := os.Getenv("CLAUDE_CONFIG_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude")
}

// mcpHelperCommand is how a run starts magpie's MCP helper.
func tuiRuntimeMcpHelperCommand() (string, []string, error) {
	if len(tuiRuntimeSettings.HelperArgv) == 0 {
		exe, err := os.Executable()
		if err != nil {
			return "", nil, err
		}
		return exe, []string{"claude-mcp-helper"}, nil
	}
	return tuiRuntimeSettings.HelperArgv[0], slices.Clone(tuiRuntimeSettings.HelperArgv[1:]), nil
}

// claudeBinary is the Claude Code a run starts in its TUI.
func tuiRuntimeClaudeBinary() (string, error) {
	if tuiRuntimeSettings.ClaudePath == nil {
		return "", errors.New("Claude Code is not installed; install it and run `claude auth login`")
	}
	return tuiRuntimeSettings.ClaudePath()
}

// callbackBaseURL is the gateway address the MCP helper calls back on.
func tuiRuntimeCallbackBaseURL() string { return tuiRuntimeSettings.CallbackBase() }

// transcriptOff says why Claude Code would keep no transcript of a run in
// configDir, "" when it keeps one: the bridge reads every reply from it,
// so a run can't go on without it, and is refused with why rather than
// started with the setting taken away unsaid.
func tuiRuntimeTranscriptOff(env []string, configDir string) string {
	for _, e := range env {
		if k, v, _ := strings.Cut(e, "="); k == "CLAUDE_CODE_SKIP_PROMPT_HISTORY" && v != "" && v != "0" && v != "false" {
			return "CLAUDE_CODE_SKIP_PROMPT_HISTORY is set in magpie's environment, which keeps Claude Code from writing the transcript the Claude TUI reads its replies from; unset it, or set the account to standard"
		}
	}
	if configDir == "" {
		configDir = tuiRuntimeClaudeConfigDir()
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
