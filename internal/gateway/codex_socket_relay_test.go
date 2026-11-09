package gateway

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/tidwall/gjson"
	"github.com/yetone/magpie/internal/provider"
)

// codexSocketTurn is a real Codex Desktop turn over a websocket: the
// warm-up's handshake headers the turn was sent on with, the turn's own
// header, and the turn's client_metadata as Codex wrote it
// (testdata/codex-ws-turn-metadata.json, captured from codex exec).
type codexSocketTurn struct {
	Prewarm    string `json:"prewarm_turn_metadata"`
	Handshake  string `json:"turn_handshake_turn_metadata"`
	UserAgent  string `json:"user_agent"`
	Originator string `json:"originator"`
	Metadata   string `json:"client_metadata_raw"`
}

func loadCodexSocketTurn(t *testing.T) codexSocketTurn {
	t.Helper()
	b, err := os.ReadFile("testdata/codex-ws-turn-metadata.json")
	if err != nil {
		t.Fatal(err)
	}
	var c codexSocketTurn
	if err := json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	return c
}

// A Codex websocket turn relayed to a keyed Responses relay goes as Codex
// sends the same turn over HTTP: a company gateway that serves only Codex
// answered the warm-up's x-codex-turn-metadata 403 "native codex clients
// only", and the websocket's own client_metadata keys 400
// request_shape_not_codex. The turn's header is its body's, and the keys
// are gone, the rest of client_metadata byte for byte. An HTTP turn, with
// neither, goes as it came.
func TestCodexSocketTurnReachesARelayAsAnHTTPTurn(t *testing.T) {
	c := loadCodexSocketTurn(t)
	if !strings.Contains(c.Prewarm, `"request_kind":"prewarm"`) || !strings.Contains(c.Metadata, "x-codex-ws-stream-request-start-ms") {
		t.Fatal("the fixture is not a warm-up's header and a socket turn's metadata")
	}
	f := &fake{t: t, ctype: "application/json", reply: `{"id":"r","object":"response","status":"completed","output":[]}`}
	setup(t, provider.Responses, f)
	send := func(meta, header string) {
		t.Helper()
		body := `{"model":"fake/m1","input":"hi","stream":false,"client_metadata":` + meta + `}`
		req := httptest.NewRequest("POST", "/v1/responses", strings.NewReader(body))
		req.Header.Set("User-Agent", c.UserAgent)
		req.Header.Set("Originator", c.Originator)
		req.Header.Set("X-Codex-Turn-Metadata", header)
		rec := httptest.NewRecorder()
		New().Handler().ServeHTTP(rec, req)
		if rec.Code != 200 || f.calls == 0 {
			t.Fatalf("%d %s", rec.Code, rec.Body.String())
		}
	}

	send(c.Metadata, c.Prewarm)
	got := f.head.Get("X-Codex-Turn-Metadata")
	if got != c.Handshake {
		t.Errorf("the relay was sent x-codex-turn-metadata\n%s\nnot the turn's own\n%s", got, c.Handshake)
	}
	if len(f.head.Values("X-Codex-Turn-Metadata")) != 1 {
		t.Errorf("x-codex-turn-metadata sent %d times", len(f.head.Values("X-Codex-Turn-Metadata")))
	}
	cm := gjson.GetBytes(f.got, "client_metadata")
	cm.ForEach(func(k, _ gjson.Result) bool {
		if socketOnlyMetadata(k.String()) {
			t.Errorf("client_metadata still has %s", k.String())
		}
		return true
	})
	want, _ := withoutKeys([]byte(c.Metadata), func(k string) bool { return k == "x-codex-ws-stream-request-start-ms" })
	if cm.Raw != string(want) {
		t.Errorf("client_metadata\n%s\nwant the rest as Codex wrote it\n%s", cm.Raw, want)
	}

	// an HTTP turn: no socket keys, its header its body's
	plain := string(want)
	send(plain, c.Handshake)
	if got := f.head.Get("X-Codex-Turn-Metadata"); got != c.Handshake {
		t.Errorf("an HTTP turn's header changed:\n%s", got)
	}
	if got := gjson.GetBytes(f.got, "client_metadata").Raw; got != plain {
		t.Errorf("an HTTP turn's client_metadata changed:\n%s", got)
	}
}

// Every key Codex adds to client_metadata over a websocket alone goes,
// but the Lite marker, which is Lite's shape's (fromLite, liteKept);
// tool_namespaces_info stays in the body and is not in the header.
func TestSocketMetadataTakenOut(t *testing.T) {
	body := []byte(`{"model":"m","client_metadata":{"a":"1","x-codex-ws-stream-request-start-ms":"2","ws_request_header_x_openai_internal_codex_responses_lite":"true","ws_request_header_other":"o","x-codex-turn-state":"s","x-codex-turn-metadata":"{\"turn_id\":\"t\",\"tool_namespaces_info\":{\"n\":1},\"request_kind\":\"turn\"}","b":2},"input":"hi"}`)
	got := string(withoutSocketMetadata(body))
	want := `{"model":"m","client_metadata":{"a":"1","ws_request_header_x_openai_internal_codex_responses_lite":"true","x-codex-turn-metadata":"{\"turn_id\":\"t\",\"tool_namespaces_info\":{\"n\":1},\"request_kind\":\"turn\"}","b":2},"input":"hi"}`
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
	if h, ok := socketTurnMetadata(body); !ok || h != `{"turn_id":"t","request_kind":"turn"}` {
		t.Errorf("header %q %v", h, ok)
	}
	for _, b := range []string{`{"model":"m"}`, `{"client_metadata":{"a":"1"}}`, `{"client_metadata":"x"}`} {
		if got := string(withoutSocketMetadata([]byte(b))); got != b {
			t.Errorf("%s changed to %s", b, got)
		}
	}
}
