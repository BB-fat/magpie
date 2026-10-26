package gui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/provider"
)

// A Claude account is set to the Claude TUI with POST
// /api/providers/claude-mode; the providers' state shows it, the editor's
// save keeps it, and what can't be set is refused with why.
func TestClaudeModeRoute(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	old := provider.ClaudeTUIGOOS
	provider.ClaudeTUIGOOS = "linux"
	t.Cleanup(func() { provider.ClaudeTUIGOOS = old })
	logins := []map[string]any{{"agent": "claude", "user": "side@example.com", "on": true, "plan": "max", "seen": time.Now().Add(-time.Hour),
		"auth": map[string]any{"claudeAiOauth": map[string]any{"accessToken": "a", "refreshToken": "r", "expiresAt": time.Now().Add(time.Hour).UnixMilli(), "subscriptionType": "max"}}}}
	b, _ := json.Marshal(logins)
	if err := os.MkdirAll(filepath.Dir(provider.Path()), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(filepath.Dir(provider.Path()), "logins.json"), b, 0o600); err != nil {
		t.Fatal(err)
	}
	var id string
	for _, p := range provider.All() {
		if p.Account != nil && p.Account.Agent == "claude" {
			id = p.ID
		}
	}
	if id == "" {
		t.Fatal("no Claude subscription from the saved account")
	}
	mux := http.NewServeMux()
	providerRoutes(mux, nil)
	post := func(path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader(body)))
		return w
	}
	if w := post("/api/providers/claude-mode", `{"id":"`+id+`","account":"side@example.com","mode":"tui"}`); w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	shown := func() map[string]string {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/providers", nil))
		var state providersJSON
		if err := json.Unmarshal(w.Body.Bytes(), &state); err != nil {
			t.Fatal(err)
		}
		for _, q := range state.Providers {
			if q.ID == id {
				return q.AccountClaudeModes
			}
		}
		return nil
	}
	if m := shown(); m["side@example.com"] != "tui" {
		t.Fatalf("the providers' state shows %v", m)
	}
	// the editor's save, which doesn't send them, keeps them
	if w := post("/api/provider/save", `{"id":"`+id+`","models":[]}`); w.Code != 200 {
		t.Fatalf("save: %d %s", w.Code, w.Body)
	}
	if m := shown(); m["side@example.com"] != "tui" {
		t.Fatalf("a save of the editor lost the mode: %v", m)
	}
	for _, body := range []string{
		`{"id":"` + id + `","account":"nobody@example.com","mode":"tui"}`,
		`{"id":"` + id + `","account":"side@example.com","mode":"fast"}`,
	} {
		if w := post("/api/providers/claude-mode", body); w.Code != 400 || !strings.Contains(w.Body.String(), "error") {
			t.Errorf("%s: %d %s", body, w.Code, w.Body)
		}
	}
	if w := post("/api/providers/claude-mode", `{"id":"`+id+`","account":"side@example.com","mode":"standard"}`); w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	if m := shown(); m != nil {
		t.Fatalf("standard left %v", m)
	}
}
