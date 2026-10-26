package gateway

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestClaudeTUIRuntimeGatewayRequestHasNoTUIState(test *testing.T) {
	requestType := reflect.TypeFor[Request]()
	for _, name := range []string{"compactHistory", "compactSystem", "seedHistory", "replayKey", "resumedCopy", "tuiSearchTool"} {
		if _, found := requestType.FieldByName(name); found {
			test.Errorf("shared gateway Request contains TUI-only state %s", name)
		}
	}
}

func TestClaudeTUIRuntimeRequestCopyKeepsProtocolAndStateIndependent(test *testing.T) {
	original := tuiRuntimeRequest{
		Request:       Request{Model: "claude-sonnet-5", MaxTokens: 64},
		compactSystem: "archived instructions",
		replayKey:     "original replay",
		resumedCopy:   true,
		tuiSearchTool: "gateway_search",
	}
	copied := original
	copied.Model = "claude-opus-5-5"
	copied.MaxTokens = 128
	copied.compactSystem = "other instructions"
	copied.replayKey = "other replay"
	copied.resumedCopy = false
	copied.tuiSearchTool = "other_search"
	if original.Model != "claude-sonnet-5" || original.MaxTokens != 64 {
		test.Fatal("copying a TUI request aliases its shared protocol fields")
	}
	if original.compactSystem != "archived instructions" || original.replayKey != "original replay" || !original.resumedCopy || original.tuiSearchTool != "gateway_search" {
		test.Fatal("copying a TUI request aliases its runtime state")
	}
}

func TestClaudeTUIRuntimeRequestEncodingMatchesGateway(test *testing.T) {
	request := tuiRuntimeRequest{
		Request:        Request{Model: "claude-sonnet-5", MaxTokens: 64, Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "continue"}}}}},
		compactHistory: []Message{{Role: "assistant", Parts: []Part{{Kind: Text, Text: "archived history"}}}},
		compactSystem:  "archived instructions",
		seedHistory:    []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "seed history"}}}},
		replayKey:      "replay fingerprint",
		resumedCopy:    true,
		tuiSearchTool:  "gateway_search",
	}
	shared, err := json.Marshal(request.Request)
	if err != nil {
		test.Fatal(err)
	}
	encoded, err := json.Marshal(request)
	if err != nil {
		test.Fatal(err)
	}
	if !bytes.Equal(shared, encoded) {
		test.Fatalf("TUI runtime state changes shared request encoding: %s", encoded)
	}
}
