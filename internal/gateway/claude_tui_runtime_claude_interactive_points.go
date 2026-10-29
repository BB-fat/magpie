package gateway

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"
)

func tuiRuntimeInteractiveWorkDirForKey(key string) (string, error) {
	return tuiRuntimeOwnedWorkDir(filepath.Join(tuiRuntimeTempRoot(), "magpie-claude-interactive-"+key))
}

// sessionRun is the live run writing native session sid, if any.
func (b *tuiRuntimeSubscriptionBridge) sessionRun(sid string) *tuiRuntimeSubscriptionRun {
	for _, run := range b.allRuns() {
		run.mu.Lock()
		mine := run.interactive && !run.closed && run.sessionID == sid
		run.mu.Unlock()
		if mine {
			return run
		}
	}
	return nil
}

// claimSession reserves native session sid for one start that goes on in
// it, until release: no live run writes it and no other start holds it.
func (b *tuiRuntimeSubscriptionBridge) claimSession(sid string) (release func(), ok bool) {
	if sid == "" {
		return nil, false
	}
	if run := b.sessionRun(sid); run != nil {
		run.mu.Lock()
		finished := run.terminalDelivered
		run.mu.Unlock()
		if !finished {
			return nil, false // still answering or waiting on tool calls
		}
		// Its reply is whole and recorded; what is left is its teardown.
		run.abort()
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.claimedSessions == nil {
		b.claimedSessions = map[string]bool{}
	}
	if b.claimedSessions[sid] {
		return nil, false
	}
	b.claimedSessions[sid] = true
	return func() {
		b.mu.Lock()
		delete(b.claimedSessions, sid)
		b.mu.Unlock()
	}, true
}

// transcriptEndsAt reports whether session sid's last message row is row:
// nothing was said in it after the point.
func tuiRuntimeTranscriptEndsAt(configDir, work, sid, row string) bool {
	data, _, err := tuiRuntimeReadTranscript(configDir, work, sid)
	if err != nil {
		return false
	}
	lines := bytes.Split(bytes.TrimRight(data, "\n"), []byte("\n"))
	for _, line := range slices.Backward(lines) {
		var r struct{ Type, UUID string }
		if json.Unmarshal(line, &r) != nil || (r.Type != "user" && r.Type != "assistant") {
			continue
		}
		return r.UUID == row
	}
	return false
}

func tuiRuntimeReadTranscript(configDir, work, sid string) ([]byte, string, error) {
	err := os.ErrNotExist
	for _, d := range tuiRuntimeClaudeSessions(configDir, work) {
		var data []byte
		if data, err = os.ReadFile(filepath.Join(d, sid+".jsonl")); err == nil {
			return data, d, nil
		}
	}
	return nil, "", err
}

// forkTranscript copies session sid's rows up to and including row, byte
// for byte but for the session id, into a new session beside it.
func tuiRuntimeForkTranscript(configDir, work, sid, row string) (string, error) {
	data, dir, err := tuiRuntimeReadTranscript(configDir, work, sid)
	if err != nil {
		return "", err
	}
	lines := bytes.SplitAfter(data, []byte("\n"))
	for i, line := range lines {
		var r struct{ UUID string }
		if json.Unmarshal(line, &r) == nil && r.UUID == row {
			id, f, err := tuiRuntimeCopyTranscript(dir, sid, lines[:i+1])
			if err != nil {
				return "", err
			}
			if err := f.Close(); err != nil {
				_ = os.Remove(filepath.Join(dir, id+".jsonl"))
				return "", err
			}
			return id, nil
		}
	}
	return "", fmt.Errorf("session %s has no row %s", sid, row)
}

// copyTranscript writes lines, session sid's rows, as a new session in dir
// and returns it open for more rows.
func tuiRuntimeCopyTranscript(dir, sid string, lines [][]byte) (string, *os.File, error) {
	id := tuiRuntimeRandomToken()[:32]
	id = id[:8] + "-" + id[8:12] + "-4" + id[13:16] + "-a" + id[17:20] + "-" + id[20:]
	path := filepath.Join(dir, id+".jsonl")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", nil, err
	}
	from, to := []byte(`"sessionId":"`+sid+`"`), []byte(`"sessionId":"`+id+`"`)
	for _, line := range lines {
		line = bytes.ReplaceAll(line, from, to)
		if !bytes.HasSuffix(line, []byte("\n")) {
			line = append(line, '\n')
		}
		if _, err := f.Write(line); err != nil {
			f.Close()
			_ = os.Remove(path)
			return "", nil, err
		}
	}
	return id, f, nil
}

// allRuns is a snapshot of the bridge's runs, to look at each under its own
// lock without holding the bridge's.
func (b *tuiRuntimeSubscriptionBridge) allRuns() []*tuiRuntimeSubscriptionRun {
	b.mu.Lock()
	defer b.mu.Unlock()
	runs := make([]*tuiRuntimeSubscriptionRun, 0, len(b.runs))
	for _, run := range b.runs {
		runs = append(runs, run)
	}
	return runs
}

// A request offering no tools, going on from a session that had them
// (Codex's own compaction request: openai/codex#37305), would change the
// head of the prefix Anthropic cached — tool definitions come before the
// system prompt and the messages — and write the whole conversation to
// the cache again. Its run is given the session's previous definitions
// instead, byte for byte, held: a call of one is answered as unavailable
// by mcpCall and never shown to the client, and the turn is told not to
// call them.
const tuiRuntimeHeldToolsConstraint = "<current_turn_constraint>\nNo tools are available for this request. Do not call any tool; answer in text only.\n</current_turn_constraint>\n\n"

const tuiRuntimeHeldToolsRefusal = "No tools are available for this request; the tool was NOT executed. Answer in text only."

const tuiRuntimeInteractiveToolsLongest = 30 * 24 * time.Hour

func tuiRuntimeInteractiveToolsDir() string {
	return filepath.Join(tuiRuntimeMustStateDir(), "claude-interactive-tools")
}

// keepInteractiveTools keeps a run's MCP tool definitions by their hash,
// for a later held turn of its session, and returns the hash ("" when they
// could not be kept). Kept definitions unused for 30 days go.
func tuiRuntimeKeepInteractiveTools(defs []byte) string {
	if len(defs) == 0 || string(defs) == "null" || string(defs) == "[]" {
		return ""
	}
	sum := sha256.Sum256(defs)
	hash := hex.EncodeToString(sum[:])
	dir := tuiRuntimeInteractiveToolsDir()
	path := filepath.Join(dir, hash+".json")
	now := time.Now()
	if os.Chtimes(path, now, now) == nil {
		return hash
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return ""
	}
	tmp, err := os.CreateTemp(dir, ".tools-*")
	if err != nil {
		return ""
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(defs); err != nil {
		tmp.Close()
		return ""
	}
	if tmp.Close() != nil || os.Rename(tmp.Name(), path) != nil {
		return ""
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if fi, err := e.Info(); err == nil && now.Sub(fi.ModTime()) > tuiRuntimeInteractiveToolsLongest {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
	return hash
}

// loadInteractiveTools reads the definitions kept under hash, nil when they
// are gone or no longer match it.
func tuiRuntimeLoadInteractiveTools(hash string) []byte {
	data, err := os.ReadFile(filepath.Join(tuiRuntimeInteractiveToolsDir(), filepath.Base(hash)+".json"))
	if err != nil {
		return nil
	}
	if sum := sha256.Sum256(data); hex.EncodeToString(sum[:]) != hash {
		return nil
	}
	return data
}
