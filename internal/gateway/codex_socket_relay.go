package gateway

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/tidwall/gjson"
)

// A Codex turn that came over a websocket (to this magpie, which sends it
// on as an HTTP request in responses_ws.go, or to another magpie that sent
// it here) goes to a keyed Responses relay as Codex sends the same turn
// over HTTP. A relay that serves only Codex (a company gateway answering
// 403 "native codex clients only", then 400 request_shape_not_codex)
// checks both:
//
//   - client_metadata: over a websocket Codex adds keys it never sends
//     over HTTP (x-codex-ws-stream-request-start-ms, the
//     ws_request_header_* copies of its headers, x-codex-turn-state, which
//     HTTP carries as a header). withoutSocketMetadata takes them out, all
//     but the Lite marker, which goes with the rest of Lite's shape or
//     stays with it (fromLite, liteKept).
//   - x-codex-turn-metadata: a socket's handshake headers are its warm-up
//     frame's (request_kind "prewarm", no turn_id), and each turn on the
//     socket was sent on with them. The turn's own is in the body's
//     client_metadata; socketTurnMetadata is it as Codex puts it in the
//     header, without tool_namespaces_info.
//   - input[].metadata: a turn that continued a reply on a socket
//     (previous_response_id) is given that reply's output items as the
//     relay streamed them (responses_ws.go, here or on the magpie that
//     sent it on), and the relay puts metadata {turn_id} on each. Codex
//     reads that output over HTTP too, and never sends metadata back: the
//     relay answered such a turn 400 request_shape_not_codex
//     ("input[].metadata, input[].metadata.turn_id"), or 403 "native codex
//     clients only". withoutItemMetadata takes it out.
//
// An HTTP turn has none of them, and its header is already its body's.

// socketOnlyMetadata says a client_metadata key is one Codex sends over a
// websocket alone, other than the Lite marker (withoutLiteMetadata).
func socketOnlyMetadata(k string) bool {
	k = strings.ToLower(k)
	if k == "ws_request_header_x_openai_internal_codex_responses_lite" {
		return false
	}
	return strings.HasPrefix(k, "x-codex-ws-") || strings.HasPrefix(k, "ws_request_header_") || k == "x-codex-turn-state"
}

// withoutSocketMetadata is body without the client_metadata keys Codex
// sends over a websocket alone, every other member and key byte for byte
// and in order. A body without any is returned as it was.
func withoutSocketMetadata(body []byte) []byte {
	r, found, ok := topMember(body, "client_metadata")
	if !found || !ok || len(r) == 0 || r[0] != '{' {
		return body
	}
	cm, changed := withoutKeys(r, socketOnlyMetadata)
	if !changed {
		return body
	}
	return withMembers(body, map[string][]byte{"client_metadata": cm})
}

// withoutItemMetadata is body without the metadata member of each input
// item, every other member and item byte for byte and in order. A body
// with no item that has one is returned as it was.
func withoutItemMetadata(body []byte) []byte {
	r, found, ok := topMember(body, "input")
	if !found || !ok || len(r) == 0 || r[0] != '[' || !bytes.Contains(r, []byte(`"metadata"`)) {
		return body
	}
	var out bytes.Buffer
	out.WriteByte('[')
	changed := false
	n := 0
	gjson.ParseBytes(r).ForEach(func(_, it gjson.Result) bool {
		item := []byte(it.Raw)
		if it.IsObject() {
			if kept, dropped := withoutKeys(item, func(k string) bool { return k == "metadata" }); dropped {
				item, changed = kept, true
			}
		}
		if n > 0 {
			out.WriteByte(',')
		}
		out.Write(item)
		n++
		return true
	})
	if !changed {
		return body
	}
	out.WriteByte(']')
	return withMembers(body, map[string][]byte{"input": out.Bytes()})
}

// socketTurnMetadata is the x-codex-turn-metadata header Codex sends with
// the turn in body: its client_metadata's, without tool_namespaces_info,
// which Codex leaves out of the header. False when the body has none.
func socketTurnMetadata(body []byte) (string, bool) {
	v := gjson.GetBytes(body, "client_metadata.x-codex-turn-metadata")
	if v.Type != gjson.String || !gjson.Valid(v.Str) || !gjson.Parse(v.Str).IsObject() {
		return "", false
	}
	h, _ := withoutKeys([]byte(v.Str), func(k string) bool { return k == "tool_namespaces_info" })
	return string(h), true
}

// withoutKeys is obj, a JSON object, without the keys drop names, the rest
// as they were and in order; false when it drops none.
func withoutKeys(obj []byte, drop func(string) bool) ([]byte, bool) {
	o := gjson.ParseBytes(obj)
	if !o.IsObject() {
		return obj, false
	}
	var out bytes.Buffer
	out.WriteByte('{')
	changed := false
	o.ForEach(func(k, v gjson.Result) bool {
		if drop(k.String()) {
			changed = true
			return true
		}
		if out.Len() > 1 {
			out.WriteByte(',')
		}
		out.WriteString(k.Raw)
		out.WriteByte(':')
		out.WriteString(v.Raw)
		return true
	})
	if !changed {
		return obj, false
	}
	out.WriteByte('}')
	return out.Bytes(), true
}

// sameJSONText says a and b are the same JSON value, however each is written.
func sameJSONText(a, b string) bool {
	var x, y any
	if json.Unmarshal([]byte(a), &x) != nil || json.Unmarshal([]byte(b), &y) != nil {
		return false
	}
	return reflect.DeepEqual(x, y)
}
