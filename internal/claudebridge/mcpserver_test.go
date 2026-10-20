package claudebridge

import (
	"strings"
	"testing"
)

func TestMCPServerNameDefault(t *testing.T) {
	ResetMCPConfigForTest()
	t.Setenv(ClaudeMCPEnv, "")
	if err := LoadMCPConfig(); err != nil {
		t.Fatal(err)
	}
	if got, want := MCPServerName(), DefaultMCPServerName; got != want {
		t.Fatalf("name = %q, want %q", got, want)
	}
	if got, want := MCPToolPrefix(), "mcp__magpie__"; got != want {
		t.Fatalf("prefix = %q, want %q", got, want)
	}
	if got, want := MaxOfferedToolNameLen(), 51; got != want {
		t.Fatalf("max tool = %d, want 51", got)
	}
}

func TestMCPServerNameCustom(t *testing.T) {
	ResetMCPConfigForTest()
	t.Setenv(ClaudeMCPEnv, "corp_gateway")
	if err := LoadMCPConfig(); err != nil {
		t.Fatal(err)
	}
	if got, want := MCPServerName(), "corp_gateway"; got != want {
		t.Fatalf("name = %q, want %q", got, want)
	}
	prefix := MCPToolPrefix()
	if got, want := prefix, "mcp__corp_gateway__"; got != want {
		t.Fatalf("prefix = %q, want %q", got, want)
	}
	if len(prefix)+MaxOfferedToolNameLen() != 64 {
		t.Fatalf("prefix %d + max %d != 64", len(prefix), MaxOfferedToolNameLen())
	}
}

func TestMCPServerNameInvalid(t *testing.T) {
	bad := []string{
		"1bad",
		"mag pie",
		"mag/pie",
		strings.Repeat("a", 33),
	}
	for _, name := range bad {
		if err := validateMCPServerName(name); err == nil {
			t.Errorf("validateMCPServerName(%q) = nil, want error", name)
		}
	}
}

func TestMCPServerNameEnvInvalid(t *testing.T) {
	ResetMCPConfigForTest()
	t.Setenv(ClaudeMCPEnv, "bad name")
	if err := LoadMCPConfig(); err == nil {
		t.Fatal("LoadMCPConfig with bad env = nil, want error")
	}
}

func TestMCPServerNameBoundaryLength(t *testing.T) {
	ResetMCPConfigForTest()
	// longest valid name that still leaves room for a hashed tool name
	name := "a" + strings.Repeat("b", 30)
	t.Setenv(ClaudeMCPEnv, name)
	if err := LoadMCPConfig(); err != nil {
		t.Fatal(err)
	}
	if len(MCPToolPrefix())+MaxOfferedToolNameLen() != 64 {
		t.Fatalf("budget: prefix=%d max=%d", len(MCPToolPrefix()), MaxOfferedToolNameLen())
	}
}
