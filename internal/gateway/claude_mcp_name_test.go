package gateway

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/claudebridge"
	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/testenv"
)

func TestClaudeMCPNameUnsetMatchesUpstream(t *testing.T) {
	restoreDefaultClaudeBridgeEnvAfterTest(t)
	claudebridge.ResetMCPConfigForTest()
	t.Setenv(claudebridge.ClaudeMCPEnv, "")
	if err := claudebridge.LoadMCPConfig(); err != nil {
		t.Fatal(err)
	}
	if claudebridge.MCPServerName() != "magpie" || claudebridge.MCPToolPrefix() != "mcp__magpie__" {
		t.Fatalf("default name/prefix = %q / %q", claudebridge.MCPServerName(), claudebridge.MCPToolPrefix())
	}
}

func TestClaudeMCPNameInvalidFailsAtBridgeInit(t *testing.T) {
	restoreDefaultClaudeBridgeEnvAfterTest(t)
	claudebridge.ResetMCPConfigForTest()
	t.Setenv(claudebridge.ClaudeMCPEnv, "no spaces")
	if err := claudebridge.LoadMCPConfig(); err == nil {
		t.Fatal("LoadMCPConfig = nil, want error for invalid name")
	}
}

func TestClaudeMCPNameCustomAcrossBridge(t *testing.T) {
	restoreDefaultClaudeBridgeEnvAfterTest(t)
	if runtime.GOOS == "windows" {
		t.Skip("a shell script stands in for Claude Code")
	}
	const server = "corp_bridge"
	claudebridge.ResetMCPConfigForTest()
	t.Setenv(claudebridge.ClaudeMCPEnv, server)
	if err := claudebridge.LoadMCPConfig(); err != nil {
		t.Fatal(err)
	}
	prefix := claudebridge.MCPToolPrefix()

	dir := t.TempDir()
	offered := filepath.Join(dir, "offered")
	script := `#!/bin/sh
for a in "$@"; do
  case "$a" in *mcpServers*) tp=$(printf '%s' "$a" | sed 's/.*"\([^"]*tools\.json\)".*/\1/');; esac
done
while read -r line; do
  grep -o '"name":"[^"]*"' "$tp" | sed 's/"name":"//; s/"$//' > ` + offered + `
  names=$(tr -c 'A-Za-z0-9_\n-' '_' < ` + offered + `)
  case "$line" in
  *wholeplease*)
    content="" i=0
    for n in $names; do
      i=$((i+1))
      [ -n "$content" ] && content="$content,"
      content="$content"'{"type":"tool_use","id":"toolu_'$i'","name":"` + prefix + `'$n'","input":{"n":'$i'}}'
    done
    echo '{"type":"assistant","message":{"id":"m","model":"claude-sonnet-5","content":['"$content"'],"stop_reason":"tool_use","usage":{"output_tokens":3}}}' ;;
  *)
    echo '{"type":"stream_event","event":{"type":"message_start","message":{"id":"m","model":"claude-sonnet-5","usage":{"input_tokens":1}}}}'
    i=0
    for n in $names; do
      echo '{"type":"stream_event","event":{"type":"content_block_start","index":'$i',"content_block":{"type":"tool_use","id":"toolu_'$i'","name":"` + prefix + `'$n'"}}}'
      echo '{"type":"stream_event","event":{"type":"content_block_delta","index":'$i',"delta":{"type":"input_json_delta","partial_json":"{\"n\":'$i'}"}}}'
      echo '{"type":"stream_event","event":{"type":"content_block_stop","index":'$i'}}'
      i=$((i+1))
    done
    echo '{"type":"stream_event","event":{"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"output_tokens":3}}}'
    echo '{"type":"stream_event","event":{"type":"message_stop"}}' ;;
  esac
done
`
	testenv.Program(t, filepath.Join(dir, "claude"), script)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	runsEndWithTest(t, dir)

	declared := []string{"spawn.agent", "a.b", "a_b"}
	toolsJSON := `[{"name":"spawn.agent","input_schema":{"type":"object"}},{"name":"a.b","input_schema":{"type":"object"}},{"name":"a_b","input_schema":{"type":"object"}}]`
	for _, stream := range []bool{false, true} {
		say := "call them"
		if !stream {
			say = "call them wholeplease"
		}
		body := fmt.Sprintf(`{"model":"claude-sonnet-5","max_tokens":100,"stream":%t,"messages":[{"role":"user","content":%q}],"tools":%s}`, stream, say, toolsJSON)
		s := New()
		rec := httptest.NewRecorder()
		var u Usage
		p := provider.Provider{ID: "claude", Account: &provider.Account{Agent: "claude", User: "u"}}
		if code, msg := s.serveClaudeSubscription(rec, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body)), provider.Anthropic, p, "claude-sonnet-5", []byte(body), &u); code != 200 {
			s.subscription.abortAll()
			t.Fatalf("stream=%t: %d %s", stream, code, msg)
		}
		s.subscription.abortAll()
		for _, want := range declared {
			if !strings.Contains(rec.Body.String(), `"name":"`+want+`"`) {
				t.Errorf("stream=%t: missing declared name %q in %s", stream, want, rec.Body.String())
			}
		}
	}

	req := &Request{Tools: []Tool{{Name: "a.b"}, {Name: "a_b"}}, ToolChoice: "name:a.b"}
	offer := bridgeTools(req)
	if got, want := toolChoiceAsk(req), fmt.Sprintf("You must call the %s%s tool.", prefix, offer[0].Name); !strings.Contains(got, want) {
		t.Errorf("tool choice %q, want %q", got, want)
	}
	var hist strings.Builder
	renderParts(nil, &hist, []Part{{Kind: ToolCall, ID: "c1", Name: "a.b", Args: json.RawMessage(`{}`)}}, offeredTools(req.Tools))
	if !strings.Contains(hist.String(), "[tool call "+prefix+offer[0].Name+" id=c1 ") {
		t.Errorf("history %q", hist.String())
	}
}

func TestBridgeToolNameCollisionWithCustomPrefix(t *testing.T) {
	restoreDefaultClaudeBridgeEnvAfterTest(t)
	claudebridge.ResetMCPConfigForTest()
	t.Setenv(claudebridge.ClaudeMCPEnv, "")
	if err := claudebridge.LoadMCPConfig(); err != nil {
		t.Fatal(err)
	}
	a := bridgeToolName("a.b")
	b := bridgeToolName("a_b")
	if a == b {
		t.Fatalf("a.b and a_b hashed to the same offered name %q", a)
	}
	if len(claudebridge.MCPToolPrefix()+a) > 64 || len(claudebridge.MCPToolPrefix()+b) > 64 {
		t.Fatalf("full names exceed 64 chars")
	}
}
