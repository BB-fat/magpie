package claudetui

// Focused tests for mcpHelperCommand path handling — macOS install paths
// with spaces ("~/my magpie tui/...") must not split the executable into
// fragments. RED before the argv fix; GREEN after.

import (
	"path/filepath"
	"testing"
)

func helperConfig(t *testing.T, argv []string) {
	t.Helper()
	prev := config
	config = &Config{}
	t.Cleanup(func() { config = prev })
	if argv != nil {
		config.HelperArgv = argv
	}
}

func TestHelperArgvSpacedExecutableStaysWhole(t *testing.T) {
	spaced := filepath.Join(t.TempDir(), "my magpie tui", "bin", "magpie-tui-mcp-helper")
	helperConfig(t, []string{spaced})
	exe, argv, err := mcpHelperCommand()
	if err != nil {
		t.Fatalf("mcpHelperCommand: %v", err)
	}
	if exe != spaced {
		t.Fatalf("executable split on space: got %q want %q (argv %v)", exe, spaced, argv)
	}
	if len(argv) != 0 {
		t.Fatalf("standalone helper must get no extra args, got %v", argv)
	}
}

func TestHelperArgvReentrySubcommandSurvivesSpaces(t *testing.T) {
	spaced := filepath.Join(t.TempDir(), "my magpie tui", "bin", "magpie-tui-runtime")
	helperConfig(t, []string{spaced, "claude-mcp-helper"})
	exe, argv, err := mcpHelperCommand()
	if err != nil {
		t.Fatalf("mcpHelperCommand: %v", err)
	}
	if exe != spaced {
		t.Fatalf("executable split on space: got %q want %q", exe, spaced)
	}
	if len(argv) != 1 || argv[0] != "claude-mcp-helper" {
		t.Fatalf("subcommand argv wrong: %v", argv)
	}
}

func TestHelperDefaultReentryUnchangedWhenNoOverride(t *testing.T) {
	helperConfig(t, nil)
	exe, argv, err := mcpHelperCommand()
	if err != nil {
		t.Fatalf("mcpHelperCommand: %v", err)
	}
	if exe == "" || len(argv) != 1 || argv[0] != "claude-mcp-helper" {
		t.Fatalf("default re-entry broken: exe=%q argv=%v", exe, argv)
	}
}
