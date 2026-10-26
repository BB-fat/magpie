package gateway

import (
	"encoding/json"
	"strings"
	"testing"
)

func tuiRuntimeSemanticReplayRequest(t *testing.T) *Request {
	t.Helper()
	req := tuiRuntimeReplayReq(t, tuiRuntimeReplayBody(""))
	req.Messages[0].Parts = append(req.Messages[0].Parts,
		Part{Kind: Image, MediaType: "image/png", Data: "QUFB"},
		Part{Kind: File, MediaType: "text/plain", Data: "QUFB"})
	req.Messages[1].Parts[0].Args = json.RawMessage(`{"path":"a","nested":{"a":1,"b":2},"n":9007199254740992}`)
	req.Messages[1].Parts = append(req.Messages[1].Parts, Part{Kind: Thinking, Text: "reason", Signature: "signed"})
	req.Messages[2].Parts[0].Images = []Part{{Kind: Image, MediaType: "image/png", Data: "QUFB"}}
	req.Safeguards = json.RawMessage(`{"policy":"a","nested":{"a":1,"b":2}}`)
	req.SafeguardBeta = "dangerous-tool-use-2026-09-03"
	return req
}

func TestClaudeTUIRuntimeLiteReplayRequiresAllSemanticContent(t *testing.T) {
	for name, change := range map[string]func(*Request){
		"arguments": func(r *Request) { r.Messages[1].Parts[0].Args = json.RawMessage(`{"path":"b"}`) },
		"large integer": func(r *Request) {
			r.Messages[1].Parts[0].Args = json.RawMessage(`{"path":"a","nested":{"a":1,"b":2},"n":9007199254740993}`)
		},
		"result error": func(r *Request) { r.Messages[2].Parts[0].IsError = true },
		"image bytes":  func(r *Request) { r.Messages[0].Parts[1].Data = "QkJC" },
		"image type":   func(r *Request) { r.Messages[0].Parts[1].MediaType = "image/jpeg" },
		"file bytes":   func(r *Request) { r.Messages[0].Parts[2].Data = "QkJC" },
		"result image": func(r *Request) {
			r.Messages[2].Parts[0].Images[0].Data = "QkJC"
		},
		"call identity": func(r *Request) {
			r.Messages[1].Parts[0].ID = "another-call"
			r.Messages[2].Parts[0].CallID = "another-call"
		},
		"text whitespace": func(r *Request) { r.Messages[0].Parts[0].Text = "run " },
		"thinking":        func(r *Request) { r.Messages[1].Parts[1].Text = "another reason" },
		"signature":       func(r *Request) { r.Messages[1].Parts[1].Signature = "another signature" },
		"safeguards":      func(r *Request) { r.Safeguards = json.RawMessage(`{"policy":"b"}`) },
		"safeguard beta":  func(r *Request) { r.SafeguardBeta = "dangerous-tool-use-another" },
	} {
		t.Run(name, func(t *testing.T) {
			b := &tuiRuntimeSubscriptionBridge{}
			original := tuiRuntimeSemanticReplayRequest(t)
			b.noteLiteCompletedToolTurn(tuiRuntimeScOwner, tuiRuntimeScOuter, original, "ORIGINAL_REPLY")
			if reply, ok := b.liteCompletedToolTurn(tuiRuntimeScOwner, tuiRuntimeScOuter, tuiRuntimeSemanticReplayRequest(t)); !ok || reply.reply != "ORIGINAL_REPLY" {
				t.Fatal("identical request did not replay its completed answer")
			}
			changed := tuiRuntimeSemanticReplayRequest(t)
			change(changed)
			if _, ok := b.liteCompletedToolTurn(tuiRuntimeScOwner, tuiRuntimeScOuter, changed); ok {
				t.Fatal("changed semantic content received the prior completed reply")
			}
		})
	}
}

func TestClaudeTUIRuntimeLiteReplayCanonicalJSONAndMetadata(t *testing.T) {
	original := tuiRuntimeSemanticReplayRequest(t)
	original.Tools[0].Schema = json.RawMessage(`{"type":"object","properties":{"a":{"type":"string"},"b":{"type":"number"}}}`)
	changed := tuiRuntimeSemanticReplayRequest(t)
	changed.Tools[0].Schema = json.RawMessage(`{"properties":{"b":{"type":"number"},"a":{"type":"string"}},"type":"object"}`)
	changed.Messages[1].Parts[0].Args = json.RawMessage(`{ "n":9007199254740992,"nested":{"b":2,"a":1},"path":"a" }`)
	changed.Safeguards = json.RawMessage(`{"nested":{"b":2,"a":1},"policy":"a"}`)
	changed.Metadata = json.RawMessage(`{"user_id":"another"}`)
	changed.ClientMetadata = json.RawMessage(`{"session_id":"another"}`)
	changed.Stream = true
	if tuiRuntimeLiteReplayKey(tuiRuntimeScOwner, original) == "" || tuiRuntimeLiteReplayKey(tuiRuntimeScOwner, original) != tuiRuntimeLiteReplayKey(tuiRuntimeScOwner, changed) {
		t.Fatal("JSON key order, metadata or response transport changed the replay identity")
	}
	for _, args := range []json.RawMessage{nil, json.RawMessage(" \n"), json.RawMessage("null"), json.RawMessage("{}")} {
		a := tuiRuntimeReplayReq(t, tuiRuntimeReplayBody(""))
		b := tuiRuntimeReplayReq(t, tuiRuntimeReplayBody(""))
		a.Messages[1].Parts[0].Args = args
		if tuiRuntimeLiteReplayKey(tuiRuntimeScOwner, a) != tuiRuntimeLiteReplayKey(tuiRuntimeScOwner, b) {
			t.Errorf("empty arguments %q differ from an empty object", args)
		}
	}
}

func TestClaudeTUIRuntimeLiteReplayInvalidJSONCannotProveEquality(t *testing.T) {
	b := &tuiRuntimeSubscriptionBridge{}
	req := tuiRuntimeSemanticReplayRequest(t)
	req.Messages[1].Parts[0].Args = json.RawMessage(`{"broken":`)
	b.noteLiteCompletedToolTurn(tuiRuntimeScOwner, tuiRuntimeScOuter, req, "UNVERIFIABLE")
	if _, ok := b.liteCompletedToolTurn(tuiRuntimeScOwner, tuiRuntimeScOuter, req); ok {
		t.Fatal("invalid argument JSON authorized a completed reply replay")
	}
}

func TestClaudeTUIRuntimeLiteServedChangedArgumentsStartAnotherRun(t *testing.T) {
	_, runs := tuiRuntimeScWrapper(t)
	rt, b := tuiRuntimeScRuntime(t)
	tuiRuntimeScSaveClean(t, b, "inner-saved", false)
	original := tuiRuntimeReplayBody("")
	for _, body := range []string{original, strings.Replace(original, `"input":{}`, `"input":{"path":"different"}`, 1)} {
		if code, out := tuiRuntimeScServe(t, rt, tuiRuntimeScOuter, body); code != 200 || !strings.Contains(out, "SC_ANSWER") {
			t.Fatalf("request failed: %d %s", code, out)
		}
	}
	if got := len(runs()); got != 2 {
		t.Fatalf("changed tool arguments reused a prior answer: %d native starts, want 2", got)
	}
}
