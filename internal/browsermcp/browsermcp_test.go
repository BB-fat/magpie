package browsermcp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/yetone/magpie/internal/gateway"
)

// Each line goes to the gateway's browser server with this process's
// session, and its answer comes back a line; a notification is answered
// nothing, and a gateway that can't be reached is said so, by id.
func TestBridgePassesLinesWithItsSession(t *testing.T) {
	var mu sync.Mutex
	var sessions []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != gateway.BrowserMCPPath || r.Header.Get("Authorization") != "Bearer "+gateway.Token {
			http.Error(w, "wrong", 400)
			return
		}
		mu.Lock()
		sessions = append(sessions, r.Header.Get(gateway.BrowserSessionHeader))
		mu.Unlock()
		b, _ := io.ReadAll(r.Body)
		var m struct {
			ID     json.RawMessage
			Method string
		}
		json.Unmarshal(b, &m)
		if len(m.ID) == 0 {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": m.ID, "result": map[string]any{"method": m.Method}})
	}))
	defer srv.Close()
	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n" +
		`{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n\n" +
		`{"jsonrpc":"2.0","id":"b","method":"tools/list"}` + "\n")
	var out bytes.Buffer
	if err := serve(in, &out, srv.URL); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 || !strings.Contains(out.String(), `"method":"initialize"`) || !strings.Contains(out.String(), `"id":"b"`) {
		t.Fatalf("out: %q", out.String())
	}
	if len(sessions) != 3 || sessions[0] == "" || sessions[0] != sessions[1] || sessions[1] != sessions[2] {
		t.Fatalf("sessions: %q", sessions)
	}

	out.Reset()
	srv.Close()
	if err := serve(strings.NewReader(`{"jsonrpc":"2.0","id":7,"method":"tools/list"}`+"\n"), &out, srv.URL); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"id":7`) || !strings.Contains(out.String(), "magpie isn't running") {
		t.Fatalf("gateway gone: %q", out.String())
	}
}
