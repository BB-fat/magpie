package claudebridge

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"sync"
)

const (
	// DefaultMCPServerName is the MCP server key and tool prefix Claude Code
	// sees when MAGPIE_CLAUDE_MCP_NAME is unset.
	DefaultMCPServerName = "magpie"
	// ClaudeMCPEnv names the Claude subscription bridge MCP server (v1 only).
	ClaudeMCPEnv = "MAGPIE_CLAUDE_MCP_NAME"
	// claudeMCPToolTotal is Claude Code's limit on a full MCP tool name,
	// including the mcp__<server>__ prefix.
	claudeMCPToolTotal = 64
)

var (
	mcpOnce      sync.Once
	mcpLoadErr   error
	mcpName      string
	mcpToolPref  string
	mcpToolMax   int
	mcpShortMax  int
	mcpNameValid = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,31}$`)
)

// LoadMCPConfig reads MAGPIE_CLAUDE_MCP_NAME once per process. Unset keeps
// DefaultMCPServerName. Invalid values return an error and leave the prior
// load, if any, unchanged on repeat calls.
func LoadMCPConfig() error {
	mcpOnce.Do(func() {
		mcpLoadErr = loadMCPConfig()
	})
	return mcpLoadErr
}

// ResetMCPConfigForTest clears the one-time load so tests can change the env.
func ResetMCPConfigForTest() {
	mcpOnce = sync.Once{}
	mcpLoadErr = nil
	mcpName = ""
	mcpToolPref = ""
	mcpToolMax = 0
	mcpShortMax = 0
}

func loadMCPConfig() error {
	name := os.Getenv(ClaudeMCPEnv)
	if name == "" {
		name = DefaultMCPServerName
	} else if err := validateMCPServerName(name); err != nil {
		return fmt.Errorf("%s: %w", ClaudeMCPEnv, err)
	}
	prefix := "mcp__" + name + "__"
	if len(prefix) >= claudeMCPToolTotal {
		return fmt.Errorf("%s: %q leaves no room for tool names under Claude Code's %d-character limit", ClaudeMCPEnv, name, claudeMCPToolTotal)
	}
	maxOffered := claudeMCPToolTotal - len(prefix)
	if maxOffered < 1 {
		return fmt.Errorf("%s: %q leaves no room for tool names", ClaudeMCPEnv, name)
	}
	shortMax := maxOffered - 13 // "_" + 12 hex digits of the hash suffix
	if shortMax < 1 {
		return fmt.Errorf("%s: %q leaves no room for hashed tool names", ClaudeMCPEnv, name)
	}
	mcpName = name
	mcpToolPref = prefix
	mcpToolMax = maxOffered
	mcpShortMax = shortMax
	return nil
}

func validateMCPServerName(name string) error {
	if !mcpNameValid.MatchString(name) {
		return fmt.Errorf("invalid MCP server name %q (want [A-Za-z][A-Za-z0-9_-]{0,31})", name)
	}
	return nil
}

func ensureMCP() {
	_ = LoadMCPConfig()
}

// MCPServerName is the mcpServers key and initialize serverInfo.name.
func MCPServerName() string {
	ensureMCP()
	return mcpName
}

// MCPToolPrefix is mcp__<server>__ for tool names Claude Code exposes.
func MCPToolPrefix() string {
	ensureMCP()
	return mcpToolPref
}

// MaxOfferedToolNameLen is how long a bridge tool name may be before hashing.
func MaxOfferedToolNameLen() int {
	ensureMCP()
	return mcpToolMax
}

// MaxHashedToolStemLen is the stem length before _<hash> when a name is hashed.
func MaxHashedToolStemLen() int {
	ensureMCP()
	return mcpShortMax
}

// ToolName is the name the MCP helper offers a tool of the caller's under.
// Claude Code takes an MCP tool's name as [A-Za-z0-9_-] only, with the MCP
// server prefix before it in at most 64: it made "spawn.agent" its
// spawn_agent, which the client never offered (#2649). A name it would
// change, or one ending as a changed name ends, is told apart by a hash of
// it, the same in every run, so a resumed session and a conversation's
// history name the tool as before; the bridges map it back.
func ToolName(name string) string {
	if ToolNameOK(name) && !toolNameHashed.MatchString(name) {
		return name
	}
	sum := sha256.Sum256([]byte(name))
	short := toolNameBad.ReplaceAllString(name, "_")
	if max := MaxHashedToolStemLen(); len(short) > max {
		short = short[:max]
	}
	return short + "_" + hex.EncodeToString(sum[:6])
}

// ToolNameOK says Claude Code keeps name as it is, under the server's
// prefix.
func ToolNameOK(name string) bool {
	max := MaxOfferedToolNameLen()
	if len(name) < 1 || len(name) > max {
		return false
	}
	return toolNameChars.MatchString(name)
}

var (
	toolNameChars  = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	toolNameBad    = regexp.MustCompile(`[^A-Za-z0-9_-]`)
	toolNameHashed = regexp.MustCompile(`_[0-9a-f]{12}$`)
)
