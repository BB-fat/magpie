package gateway

// End-to-end test of the extracted runtime's own HTTP surface: a mock
// claude-bridge wrapper is invoked by the runtime's interactive transport,
// and the plugin's bridge-auth-v1 signature carries a signed request to
// it. Covers the same ground as test/host-integration.mjs does over the
// plugin JS, but at the runtime's own boundary — the session the runtime
// claims, the MCP callback the helper answers, and the stream the caller
// sees.

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func tuiRuntimeSign(key, method, path, ts, nonce, session string, body []byte) string {
	canonical := strings.Join([]string{
		"bridge-auth-v1", method, path, ts, nonce, session, tuiRuntimeSha256Hex(body),
	}, "\n")
	m := hmac.New(sha256.New, []byte(key))
	m.Write([]byte(canonical))
	return hex.EncodeToString(m.Sum(nil))
}

func tuiRuntimeAuthHeaders(key, session string, body []byte) map[string]string {
	ts := fmt.Sprintf("%d", time.Now().Unix())
	nonce := "t" + newID()
	return map[string]string{
		"content-type":             "application/json",
		"x-bridge-auth-version":    "1",
		"x-bridge-timestamp":       ts,
		"x-bridge-nonce":           nonce,
		"x-bridge-session":         session,
		"x-bridge-mac":             tuiRuntimeSign(key, "POST", "/v1/messages", ts, nonce, session, body),
		"x-bridge-auth-version-ok": "1",
	}
}

func tuiRuntimePostJSON(t *testing.T, base, key, session string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	req, _ := http.NewRequest("POST", base+"/v1/messages", strings.NewReader(string(body)))
	for k, v := range tuiRuntimeAuthHeaders(key, session, body) {
		if strings.HasSuffix(k, "-ok") {
			continue
		}
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	return rec
}

func TestClaudeTUIRuntimeMessagesRequiresSignature(t *testing.T) {
	rt := tuiRuntimeNewRuntime([]byte("test-key"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader("{}"))
	rt.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unsigned request: status %d", rec.Code)
	}
}

func TestClaudeTUIRuntimeMessagesRejectsBadMAC(t *testing.T) {
	rt := tuiRuntimeNewRuntime([]byte("test-key"))
	body := []byte(`{"model":"claude-tui","messages":[{"role":"user","content":[{"type":"text","text":"hi"}]}]}`)
	rec := tuiRuntimePostJSON(t, "http://x", "test-key", "sess", body)
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(string(body)))
	for k, v := range tuiRuntimeAuthHeaders("WRONG", "sess", body) {
		if !strings.HasSuffix(k, "-ok") {
			req.Header.Set(k, v)
		}
	}
	rt.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("forged signature: status %d", rec.Code)
	}
}

// fakeBridge writes one assistant row and a terminal result row over a
// transcript stream that readInteractiveOneTurn reads, then exits. It
// stands in for the real claude-bridge wrapper the runtime launches.
func tuiRuntimeFakeBridge(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell-script bridge stub is unix-only")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "fake-bridge.sh")
	body := `#!/bin/sh
# Emit one assistant text row, then a result, then exit.
cat >/dev/null
printf '%s\n' '{"type":"assistant","uuid":"u1","session_id":"sess1","message":{"id":"m1","model":"claude-tui","usage":{"input_tokens":1,"output_tokens":2},"content":[{"type":"text","text":"hello"}]}}'
printf '%s\n' '{"type":"result","uuid":"u2","session_id":"sess1","stop_reason":"end_turn","result":"done","is_error":false}'
`
	if err := os.WriteFile(path, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestClaudeTUIRuntimeInteractiveRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	tuiRuntimeTestInit(t, &tuiRuntimeConfig{
		StateDir:   tmp,
		HelperArgv: []string{"/bin/false"}, // unused: the fake bridge never calls back
	})
	tuiRuntimeSetWrapper(t, tuiRuntimeFakeBridge(t))
	key := "test-key"
	rt := tuiRuntimeNewRuntime([]byte(key))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	base := "http://" + ln.Addr().String()
	tuiRuntimeTestInit(t, &tuiRuntimeConfig{CallbackBase: func() string { return base }})
	srv := &http.Server{Handler: rt}
	go srv.Serve(ln)
	defer srv.Close()

	body, _ := json.Marshal(map[string]any{
		"model": "claude-tui",
		"messages": []any{map[string]any{
			"role":    "user",
			"content": []any{map[string]any{"type": "text", "text": "hi"}},
		}},
	})
	req, _ := http.NewRequest("POST", base+"/v1/messages", strings.NewReader(string(body)))
	for k, v := range tuiRuntimeAuthHeaders(key, "session-a", body) {
		if !strings.HasSuffix(k, "-ok") {
			req.Header.Set(k, v)
		}
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	out, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		t.Fatalf("status %d body %s", res.StatusCode, out)
	}
	if !strings.Contains(string(out), "hello") {
		t.Fatalf("reply missing text: %s", out)
	}
}
