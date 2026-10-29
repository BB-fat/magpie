package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

// A native checkpoint is where a native transcript holds a complete reply:
// the strict prefix the client had sent up to and including that reply, and
// the native session and transcript row that end it. Unlike the loose cut
// point (tuiRuntimeInteractivePoint), it is keyed by the strict content
// prefix (tuiRuntimePrefixOf) within an isolation scope (account + caller
// key), never by an external session header: a client that sends no session
// header still resumes its native session, and two callers of one account
// never resume each other's.
//
// The prefix is matched exactly — a history whose digest differs anywhere
// is a different history. A request goes on from the deepest checkpoint its
// history extends: resumed in place when no run writes that session and the
// transcript still ends at the checkpoint's row, else forked there. Only
// then is Anthropic sent the prefix it has cached; a rebuild from the
// request writes the whole history again.
//
// The store lives in a new file (claude-native-checkpoints.json), not the
// loose cut-point store: the two are not interchangeable, and the loose
// store stays for the header-keyed paths that still read it.
type tuiRuntimeNativeCheckpoint struct {
	Prefix         string    `json:"prefix"`     // hex of tuiRuntimePrefixDigest of msgs up to the reply
	Scope          string    `json:"scope"`      // tuiRuntimePrefixScope(owner, caller key)
	SessionID      string    `json:"session_id"` // native session holding the reply
	Row            string    `json:"row"`        // uuid of the reply's last transcript row
	Work           string    `json:"work"`       // state key naming the session's cwd
	Model          string    `json:"model,omitempty"`
	InstructionKey string    `json:"instruction_key"`
	ToolsHash      string    `json:"tools,omitempty"` // definitions the run served (keepInteractiveTools)
	UpdatedAt      time.Time `json:"updated_at"`
}

type tuiRuntimeNativeCheckpointStore struct {
	Version int                                   `json:"version"`
	Points  map[string]tuiRuntimeNativeCheckpoint `json:"points"` // scope+"\x00"+prefix → checkpoint
}

const (
	tuiRuntimeNativeCheckpointVersion  = 1
	tuiRuntimeNativeCheckpointMost     = 4096
	tuiRuntimeNativeCheckpointPerScope = 256
	tuiRuntimeNativeCheckpointLongest  = 30 * 24 * time.Hour
)

// tuiRuntimeNativeCheckpointMu serializes saves, each a load, change and
// write of the whole store. Readers never take it: a save replaces the file
// by rename, so a read sees the whole store before or after it, and a resume
// never waits on another turn's disk sync.
var tuiRuntimeNativeCheckpointMu sync.Mutex

func tuiRuntimeNativeCheckpointPath() string {
	return filepath.Join(tuiRuntimeMustStateDir(), "claude-native-checkpoints.json")
}

// loadNativeCheckpoints reads the store. A store that doesn't read is an
// error, never taken for an empty one, so it is not written over either.
func tuiRuntimeLoadNativeCheckpoints() (tuiRuntimeNativeCheckpointStore, error) {
	empty := tuiRuntimeNativeCheckpointStore{Version: tuiRuntimeNativeCheckpointVersion, Points: map[string]tuiRuntimeNativeCheckpoint{}}
	data, err := os.ReadFile(tuiRuntimeNativeCheckpointPath())
	if errors.Is(err, os.ErrNotExist) {
		return empty, nil
	}
	if err != nil {
		return empty, err
	}
	var store tuiRuntimeNativeCheckpointStore
	if err := json.Unmarshal(data, &store); err != nil {
		return empty, fmt.Errorf("native checkpoint store: %w", err)
	}
	if store.Version != tuiRuntimeNativeCheckpointVersion {
		return empty, fmt.Errorf("native checkpoint store version %d", store.Version)
	}
	if store.Points == nil {
		store.Points = map[string]tuiRuntimeNativeCheckpoint{}
	}
	return store, nil
}

func tuiRuntimeWriteNativeCheckpoints(store tuiRuntimeNativeCheckpointStore) error {
	cutoff := time.Now().Add(-tuiRuntimeNativeCheckpointLongest)
	perScope := map[string][]string{}
	for k, p := range store.Points {
		if p.UpdatedAt.Before(cutoff) {
			delete(store.Points, k)
			continue
		}
		perScope[p.Scope] = append(perScope[p.Scope], k)
	}
	older := func(a, b string) int { return store.Points[b].UpdatedAt.Compare(store.Points[a].UpdatedAt) }
	for _, keys := range perScope {
		if len(keys) > tuiRuntimeNativeCheckpointPerScope {
			slices.SortFunc(keys, older)
			for _, k := range keys[tuiRuntimeNativeCheckpointPerScope:] {
				delete(store.Points, k)
			}
		}
	}
	if len(store.Points) > tuiRuntimeNativeCheckpointMost {
		keys := make([]string, 0, len(store.Points))
		for k := range store.Points {
			keys = append(keys, k)
		}
		slices.SortFunc(keys, older)
		for _, k := range keys[tuiRuntimeNativeCheckpointMost:] {
			delete(store.Points, k)
		}
	}
	data, err := json.Marshal(store)
	if err != nil {
		return err
	}
	path := tuiRuntimeNativeCheckpointPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".claude-native-checkpoints-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

func tuiRuntimeNativeCheckpointKey(scope string, prefix tuiRuntimePrefixDigest) string {
	return scope + "\x00" + fmt.Sprintf("%x", prefix[:])
}

// saveNativeCheckpoint records the boundary this run's complete reply made:
// the strict prefix of the request's messages plus the reply exactly as the
// client was told it (r.turnReply, kept by noteTurnReply), in the request's
// isolation scope, pointing at the native session and transcript row that
// hold it. A later request whose history extends that exact prefix resumes
// here. Only runs with a known native session, row and cwd are recorded.
func (r *tuiRuntimeSubscriptionRun) saveNativeCheckpoint(req *tuiRuntimeRequest) {
	r.mu.Lock()
	reply := r.turnReply
	sessionID, row, work, toolsHash := r.sessionID, r.lastRow, r.workKey, r.toolsHash
	model := r.model
	r.mu.Unlock()
	if req.scope == "" || sessionID == "" || row == "" || work == "" || len(reply.Parts) == 0 {
		return
	}
	msgs := append(req.Messages[:len(req.Messages):len(req.Messages)], reply)
	prefix := tuiRuntimePrefixOf(req.scope, msgs)
	if err := tuiRuntimeSaveNativeCheckpoint(tuiRuntimeNativeCheckpoint{
		Prefix: fmt.Sprintf("%x", prefix[:]), Scope: req.scope,
		SessionID: sessionID, Row: row, Work: work,
		Model: model, InstructionKey: tuiRuntimeInstructionKey(req), ToolsHash: toolsHash,
	}); err != nil {
		log.Printf("Claude interactive native checkpoint not saved: %v", err)
	}
}

// nativeRootKey is a no-header conversation's stable work key: its
// isolation scope plus the digest of its first message. The first message
// does not change as the conversation grows, so every later turn derives
// the same cwd; a compacted or edited history has a new first message and
// so a new root. Returns "" with no scope or no messages (a truly stateless
// request keeps the shared slot).
func tuiRuntimeNativeRootKey(req *tuiRuntimeRequest) string {
	if req.scope == "" || len(req.Messages) == 0 {
		return ""
	}
	first := tuiRuntimePrefixOf(req.scope, req.Messages[:1])
	sum := sha256.Sum256([]byte(req.scope + "\x00root\x00" + fmt.Sprintf("%x", first[:])))
	return hex.EncodeToString(sum[:])
}

// saveNativeCheckpoint records the boundary at (scope, prefix).
func tuiRuntimeSaveNativeCheckpoint(p tuiRuntimeNativeCheckpoint) error {
	if p.Prefix == "" || p.Scope == "" || p.SessionID == "" || p.Row == "" || p.Work == "" || p.InstructionKey == "" {
		return errors.New("incomplete native checkpoint")
	}
	tuiRuntimeNativeCheckpointMu.Lock()
	defer tuiRuntimeNativeCheckpointMu.Unlock()
	store, err := tuiRuntimeLoadNativeCheckpoints()
	if err != nil {
		return err
	}
	p.UpdatedAt = time.Now().UTC()
	store.Points[p.Scope+"\x00"+p.Prefix] = p
	return tuiRuntimeWriteNativeCheckpoints(store)
}

// findNativeCheckpoint finds the deepest checkpoint req's history goes on
// from: for each prefix of the request's messages that ends at an assistant
// reply, deepest first, the checkpoint recorded under that exact prefix in
// req's scope. It returns the checkpoint and the messages after it.
func tuiRuntimeFindNativeCheckpoint(req *tuiRuntimeRequest, model string) (tuiRuntimeNativeCheckpoint, []Message, bool) {
	if req.scope == "" {
		return tuiRuntimeNativeCheckpoint{}, nil, false
	}
	store, err := tuiRuntimeLoadNativeCheckpoints()
	if err != nil {
		// a store that doesn't read is unknown, not empty: resume from
		// nothing (the request is rebuilt), and never write over it
		log.Printf("Claude interactive native checkpoints unreadable, not resuming from them: %v", err)
		return tuiRuntimeNativeCheckpoint{}, nil, false
	}
	if len(store.Points) == 0 {
		return tuiRuntimeNativeCheckpoint{}, nil, false
	}
	digests := tuiRuntimePrefixDigests(req.scope, req.Messages)
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role != "assistant" {
			continue
		}
		key := tuiRuntimeNativeCheckpointKey(req.scope, digests[i])
		p, ok := store.Points[key]
		if !ok || (p.Model != "" && p.Model != model) {
			continue
		}
		// the suffix after this reply must be new user input only, not a
		// rewritten later reply or a tool result (those go through the tool
		// continuation / fork paths)
		since := req.Messages[i+1:]
		if !tuiRuntimeNativeCheckpointSuffixOK(since) {
			continue
		}
		return p, since, true
	}
	return tuiRuntimeNativeCheckpoint{}, nil, false
}

// nativeCheckpointSuffixOK says the messages after a checkpointed reply are
// a fresh turn: no later assistant reply and no tool results.
func tuiRuntimeNativeCheckpointSuffixOK(since []Message) bool {
	if len(since) == 0 {
		return false
	}
	for _, m := range since {
		if m.Role == "assistant" {
			return false
		}
		for _, p := range m.Parts {
			if p.Kind == ToolResult {
				return false
			}
		}
	}
	return true
}
