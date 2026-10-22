package gateway

import (
	"net/http"
	"regexp"
	"strings"
)

// clientOS is the OS of the machine a client's tools run on, as an official
// client's own request headers tell it: the X-Stainless-OS and
// X-Stainless-Arch of Anthropic's and OpenAI's SDKs, else the platform in
// a User-Agent's parentheses (Codex: "codex_cli_rs/0.x (Mac OS 15.5; arm64)").
// Empty when they tell none (Claude Code: "claude-cli/x (external, cli)").
func clientOS(r *http.Request) string {
	if r == nil {
		return ""
	}
	if os := normalizeOS(r.Header.Get("X-Stainless-OS")); os != "" && !strings.HasPrefix(strings.ToLower(os), "unknown") {
		if arch := normalizeArch(r.Header.Get("X-Stainless-Arch")); arch != "" {
			return os + " " + arch
		}
		return os
	}
	return uaOS(r.Header.Get("User-Agent"))
}

var uaParens = regexp.MustCompile(`\(([^()]*)\)`)

func uaOS(ua string) string {
	for _, m := range uaParens.FindAllStringSubmatch(ua, -1) {
		parts := strings.Split(m[1], ";")
		os := normalizeOS(parts[0])
		if os == "" {
			continue
		}
		if len(parts) > 1 {
			if arch := normalizeArch(parts[1]); arch != "" {
				return os + " " + arch
			}
		}
		return os
	}
	return ""
}

var osNames = []struct{ prefix, name string }{
	{"mac os x", "macOS"}, {"macos", "macOS"}, {"mac os", "macOS"}, {"darwin", "macOS"},
	{"windows", "Windows"}, {"linux", "Linux"}, {"ubuntu", "Ubuntu"}, {"debian", "Debian"},
	{"fedora", "Fedora"}, {"arch linux", "Arch Linux"}, {"freebsd", "FreeBSD"},
}

// normalizeOS names an OS as people write it ("MacOS" → "macOS"), keeping
// its version; empty for what names no OS.
func normalizeOS(s string) string {
	s = strings.TrimSpace(s)
	l := strings.ToLower(s)
	for _, n := range osNames {
		if strings.HasPrefix(l, n.prefix) {
			rest := strings.TrimSpace(s[len(n.prefix):])
			if rest == "" {
				return n.name
			}
			return n.name + " " + rest
		}
	}
	return ""
}

func normalizeArch(s string) string {
	switch l := strings.ToLower(strings.TrimSpace(s)); l {
	case "":
		return ""
	case "arm64", "aarch64":
		return "arm64"
	case "x64", "x86_64", "amd64":
		return "x64"
	case "x32", "x86", "i386", "i686", "ia32":
		return "x86"
	default:
		if strings.HasPrefix(l, "other:") || l == "unknown" {
			return ""
		}
		return l
	}
}

// clientOSLine tells a run where the client's tools run.
func clientOSLine(req *Request) string {
	if req.ClientOS == "" {
		return ""
	}
	return "\nThe MCP server runs on a remote machine: " + req.ClientOS + "."
}
