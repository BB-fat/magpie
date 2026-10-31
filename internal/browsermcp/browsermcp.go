// Package browsermcp is `magpie mcp browser`: the browser MCP server
// (gateway.BrowserMCPPath) over stdio, for an agent that starts its MCP
// servers as commands. Each message is passed to the running gateway and
// its answer back; the process is one agent session, which goes on in the
// tab it was working in.
package browsermcp

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/yetone/magpie/internal/gateway"
)

// Run serves stdin and stdout until stdin ends.
func Run(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("magpie mcp browser takes no arguments")
	}
	base := strings.TrimRight(os.Getenv("MAGPIE_GATEWAY"), "/")
	if base == "" {
		base = gateway.URL()
	}
	return serve(os.Stdin, os.Stdout, base)
}

func session() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "stdio-" + hex.EncodeToString(b)
}

type message struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
}

func serve(in io.Reader, out io.Writer, base string) error {
	s := &bridge{base: base, session: session(), client: &http.Client{Timeout: 2 * time.Minute}, out: out}
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 64<<10), 16<<20)
	var wg sync.WaitGroup
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var m message
		if json.Unmarshal(line, &m) != nil {
			continue
		}
		body := append([]byte(nil), line...)
		wg.Add(1)
		// calls run side by side, as an agent may send the next before the
		// last is answered (a ping during a long wait)
		go func() {
			defer wg.Done()
			s.pass(m, body)
		}()
	}
	wg.Wait()
	return sc.Err()
}

type bridge struct {
	base, session string
	client        *http.Client
	mu            sync.Mutex
	out           io.Writer
}

func (s *bridge) write(b []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.out.Write(append(bytes.TrimSpace(b), '\n'))
}

func (s *bridge) fail(id json.RawMessage, msg string) {
	if len(id) == 0 {
		return
	}
	b, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "error": map[string]any{"code": -32603, "message": msg}})
	s.write(b)
}

func (s *bridge) pass(m message, body []byte) {
	r, err := http.NewRequest(http.MethodPost, s.base+gateway.BrowserMCPPath, bytes.NewReader(body))
	if err != nil {
		s.fail(m.ID, err.Error())
		return
	}
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", "Bearer "+gateway.Token)
	r.Header.Set(gateway.BrowserSessionHeader, s.session)
	resp, err := s.client.Do(r)
	if err != nil {
		s.fail(m.ID, "magpie isn't running, so Magpie Browser can't be reached: start magpie ("+err.Error()+")")
		return
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 128<<20))
	if len(m.ID) == 0 || resp.StatusCode == http.StatusAccepted {
		return
	}
	if resp.StatusCode != http.StatusOK {
		s.fail(m.ID, fmt.Sprintf("magpie answered %d: %s", resp.StatusCode, strings.TrimSpace(string(b))))
		return
	}
	s.write(b)
}
