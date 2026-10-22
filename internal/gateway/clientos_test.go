package gateway

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientOS(t *testing.T) {
	for _, c := range []struct{ os, arch, ua, want string }{
		{"MacOS", "arm64", "", "macOS arm64"},
		{"Windows", "x64", "", "Windows x64"},
		{"Linux", "x86_64", "x", "Linux x64"},
		{"Unknown", "arm64", "codex_cli_rs/0.40.0 (Mac OS 15.5; arm64) iTerm", "macOS 15.5 arm64"},
		{"", "", "codex_cli_rs/0.40.0 (Mac OS 15.5; arm64) iTerm.app/3.5", "macOS 15.5 arm64"},
		{"", "", "codex_cli_rs/0.40.0 (Windows 10.0.26100; x86_64) WindowsTerminal", "Windows 10.0.26100 x64"},
		{"", "", "codex_cli_rs/0.40.0 (Ubuntu 24.04; x86_64) xterm", "Ubuntu 24.04 x64"},
		{"", "", "codex_cli_rs/0.40.0 (Linux 6.8; aarch64)", "Linux 6.8 arm64"},
		{"", "", "claude-cli/2.1.0 (external, cli)", ""},
		{"", "", "", ""},
	} {
		r := httptest.NewRequest("POST", "/", nil)
		if c.os != "" {
			r.Header.Set("X-Stainless-OS", c.os)
			r.Header.Set("X-Stainless-Arch", c.arch)
		}
		if c.ua != "" {
			r.Header.Set("User-Agent", c.ua)
		}
		if got := clientOS(r); got != c.want {
			t.Errorf("%q/%q/%q: got %q, want %q", c.os, c.arch, c.ua, got, c.want)
		}
	}
}

func TestClientOSPrompt(t *testing.T) {
	text := func(req *Request) string {
		blocks, err := renderClaudePrompt(req)
		if err != nil {
			t.Fatal(err)
		}
		var b strings.Builder
		for _, bl := range blocks {
			s, _ := bl["text"].(string)
			b.WriteString(s)
		}
		return b.String()
	}
	msgs := []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "hi"}}}}
	got := text(&Request{System: "Be brief.", ClientOS: "macOS 15.5 arm64", ToolChoice: "required", Messages: msgs})
	if !strings.Contains(got, "Be brief.\nThe MCP server runs on a remote machine: macOS 15.5 arm64.\nYou must call at least one") {
		t.Errorf("system with OS: %q", got)
	}
	got = text(&Request{ClientOS: "Windows x64", Messages: msgs})
	if !strings.Contains(got, "<external_system_instructions>\n\nThe MCP server runs on a remote machine: Windows x64.\n</external_system_instructions>") {
		t.Errorf("OS alone: %q", got)
	}
	if got = text(&Request{System: "Be brief.", Messages: msgs}); strings.Contains(got, "MCP server runs") {
		t.Errorf("no OS: %q", got)
	}
	if got := leanSystem(&Request{System: "S", ClientOS: "Linux x64"}); got != "S\nThe MCP server runs on a remote machine: Linux x64." {
		t.Errorf("lean: %q", got)
	}
}
