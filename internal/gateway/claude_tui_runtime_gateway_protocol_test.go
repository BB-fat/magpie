package gateway

import (
	"bytes"
	"os"
	"testing"
)

func TestClaudeTUIRuntimeGatewayProtocolPreservesSealedThinking(t *testing.T) {
	body := []byte(`{"model":"claude-sonnet-5","max_tokens":64,"messages":[{"role":"assistant","content":[{"type":"redacted_thinking","data":"opaque-signed-history"},{"type":"text","text":"previous answer"}]},{"role":"user","content":"continue"}]}`)
	request, err := parseAnthropic(body)
	if err != nil {
		t.Fatal(err)
	}
	encoded := buildAnthropic(request, request.Model)
	if !bytes.Contains(encoded, []byte(`"redacted_thinking"`)) || !bytes.Contains(encoded, []byte(`"opaque-signed-history"`)) {
		t.Fatalf("sealed thinking disappeared from the request: %s", encoded)
	}
}

func TestClaudeTUIRuntimeGatewayProtocolPreservesDeferredTools(t *testing.T) {
	body := []byte(`{"model":"claude-sonnet-5","max_tokens":64,"tools":[{"name":"read_file","description":"Read a file","defer_loading":true,"input_schema":{"type":"object","properties":{"path":{"type":"string"}}}}],"messages":[{"role":"user","content":"read the file"}]}`)
	request, err := parseAnthropic(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(request.Tools) != 1 || !request.Tools[0].Deferred {
		t.Fatalf("deferred tool setting disappeared while parsing: %#v", request.Tools)
	}
}

func TestClaudeTUIRuntimeGatewayProtocolPreservesSealedTranscript(t *testing.T) {
	request, err := parseAnthropic([]byte(`{"model":"claude-sonnet-5","max_tokens":64,"messages":[{"role":"assistant","content":[{"type":"redacted_thinking","data":"opaque-signed-history"},{"type":"text","text":"previous answer"}]},{"role":"user","content":"continue"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	_, path, err := tuiRuntimeWriteClaudeTranscript(t.TempDir(), t.TempDir(), request.Model, nil, request.Messages, "", "")
	if err != nil {
		t.Fatal(err)
	}
	transcript, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(transcript, []byte(`"redacted_thinking"`)) || !bytes.Contains(transcript, []byte(`"opaque-signed-history"`)) {
		t.Fatalf("sealed thinking disappeared from the native transcript: %s", transcript)
	}
}
