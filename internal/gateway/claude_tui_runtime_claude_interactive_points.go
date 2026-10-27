package gateway

import (
	"bytes"
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

// A cut point is where a native transcript holds a complete reply: the
// conversation up to and including that reply (its conv key, which names
// the owner) and the transcript row that ends it. Native transcripts only
// grow, so a point never changes once written. A request goes on from the
// deepest point its history extends: in that session itself when no run
// writes it and it still ends at the point, else in a copy of the session
// up to the point (a fork). Either way Anthropic is sent the prefix it has
// cached, where a session rebuilt from the request writes all of it again.
// Branches of one conversation (side questions, compactions, edits of an
// earlier turn, the same history under another session id) each continue
// their own transcript, side by side.
type tuiRuntimeInteractivePoint struct {
	SessionID      string    `json:"session_id"`
	Row            string    `json:"row"`   // uuid of the reply's last transcript row
	Work           string    `json:"work"`  // state key naming the session's cwd
	Outer          string    `json:"outer"` // state key of the conversation that wrote it
	Model          string    `json:"model,omitempty"`
	InstructionKey string    `json:"instruction_key"`
	ReplyKey       string    `json:"reply_key,omitempty"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type tuiRuntimeInteractivePointStore struct {
	Version int                                   `json:"version"`
	Points  map[string]tuiRuntimeInteractivePoint `json:"points"` // conv key → point
}

const (
	tuiRuntimeInteractivePointStoreVersion = 1
	tuiRuntimeInteractivePointsMost        = 4096 // oldest go first
	tuiRuntimeInteractivePointsPerOuter    = 128
	tuiRuntimeInteractivePointsLongest     = 30 * 24 * time.Hour
)

// pointsMu serializes the point store's read-modify-write across bridges.
var tuiRuntimeInteractivePointsMu sync.Mutex

func tuiRuntimeInteractivePointsPath() string {
	return filepath.Join(tuiRuntimeMustStateDir(), "claude-interactive-points.json")
}

// loadInteractivePoints reads the store. A store that doesn't read is an
// error, never taken for an empty one, so it is not written over either.
func tuiRuntimeLoadInteractivePoints() (tuiRuntimeInteractivePointStore, error) {
	empty := tuiRuntimeInteractivePointStore{Version: tuiRuntimeInteractivePointStoreVersion, Points: map[string]tuiRuntimeInteractivePoint{}}
	data, err := os.ReadFile(tuiRuntimeInteractivePointsPath())
	if errors.Is(err, os.ErrNotExist) {
		return empty, nil
	}
	if err != nil {
		return empty, err
	}
	var store tuiRuntimeInteractivePointStore
	if err := json.Unmarshal(data, &store); err != nil {
		return empty, fmt.Errorf("cut point store: %w", err)
	}
	if store.Version != tuiRuntimeInteractivePointStoreVersion {
		return empty, fmt.Errorf("cut point store version %d", store.Version)
	}
	if store.Points == nil {
		store.Points = map[string]tuiRuntimeInteractivePoint{}
	}
	return store, nil
}

func tuiRuntimeWriteInteractivePoints(store tuiRuntimeInteractivePointStore) error {
	cutoff := time.Now().Add(-tuiRuntimeInteractivePointsLongest)
	perOuter := map[string][]string{}
	for k, p := range store.Points {
		if p.UpdatedAt.Before(cutoff) {
			delete(store.Points, k)
			continue
		}
		perOuter[p.Outer] = append(perOuter[p.Outer], k)
	}
	older := func(a, b string) int { return store.Points[b].UpdatedAt.Compare(store.Points[a].UpdatedAt) }
	for _, keys := range perOuter {
		if len(keys) > tuiRuntimeInteractivePointsPerOuter {
			slices.SortFunc(keys, older)
			for _, k := range keys[tuiRuntimeInteractivePointsPerOuter:] {
				delete(store.Points, k)
			}
		}
	}
	if len(store.Points) > tuiRuntimeInteractivePointsMost {
		keys := make([]string, 0, len(store.Points))
		for k := range store.Points {
			keys = append(keys, k)
		}
		slices.SortFunc(keys, older)
		for _, k := range keys[tuiRuntimeInteractivePointsMost:] {
			delete(store.Points, k)
		}
	}
	data, err := json.Marshal(store)
	if err != nil {
		return err
	}
	path := tuiRuntimeInteractivePointsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".claude-interactive-points-*")
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

// saveInteractivePoint records the point at conv.
func tuiRuntimeSaveInteractivePoint(conv string, p tuiRuntimeInteractivePoint) error {
	if conv == "" || p.SessionID == "" || p.Row == "" || p.Work == "" || p.InstructionKey == "" {
		return errors.New("incomplete cut point")
	}
	tuiRuntimeInteractivePointsMu.Lock()
	defer tuiRuntimeInteractivePointsMu.Unlock()
	store, err := tuiRuntimeLoadInteractivePoints()
	if err != nil {
		return err
	}
	p.UpdatedAt = time.Now().UTC()
	store.Points[conv] = p
	return tuiRuntimeWriteInteractivePoints(store)
}

// findInteractivePoint finds the deepest point req's history goes on from:
// by the conversation up to a reply (any of the owner's conversations, so
// the same history under another session id is found too), else, in this
// conversation, by a reply the client kept while it rewrote what came
// before. It returns the point and the messages after it.
func tuiRuntimeFindInteractivePoint(owner, outerSession, model string, req *tuiRuntimeRequest) (tuiRuntimeInteractivePoint, []Message, bool) {
	tuiRuntimeInteractivePointsMu.Lock()
	store, err := tuiRuntimeLoadInteractivePoints()
	tuiRuntimeInteractivePointsMu.Unlock()
	if err != nil {
		log.Printf("Claude interactive cut points unreadable, not resuming from them: %v", err)
		return tuiRuntimeInteractivePoint{}, nil, false
	}
	if len(store.Points) == 0 {
		return tuiRuntimeInteractivePoint{}, nil, false
	}
	msgs := req.Messages
	usable := func(p tuiRuntimeInteractivePoint) bool { return p.Model == "" || p.Model == model }
	best, bestAt := tuiRuntimeInteractivePoint{}, -1
	for _, k := range slices.Backward(tuiRuntimeConversationKeyIndex(owner, msgs)) {
		if p, ok := store.Points[k.key]; ok && usable(p) {
			if _, status := tuiRuntimeSafeConversationSuffixAfterDetailed(msgs, k.at); status == tuiRuntimeConversationSuffixOK {
				best, bestAt = p, k.at
			}
			break
		}
	}
	if outerSession != "" {
		outer := tuiRuntimeInteractiveSessionStateKey(owner, outerSession)
		for _, p := range store.Points {
			if p.Outer != outer || p.ReplyKey == "" || !usable(p) {
				continue
			}
			since, status := tuiRuntimeConversationSuffixAfterReplyDetailed(msgs, p.ReplyKey)
			if at := len(msgs) - len(since) - 1; status == tuiRuntimeConversationSuffixOK && at > bestAt {
				best, bestAt = p, at
			}
		}
	}
	if bestAt < 0 {
		return tuiRuntimeInteractivePoint{}, nil, false
	}
	return best, msgs[bestAt+1:], true
}

type tuiRuntimeConversationKeyAt struct {
	key string
	at  int // index of the reply the key ends with
}

// conversationKeyIndex is conversationKeys with the index of each reply.
func tuiRuntimeConversationKeyIndex(owner string, msgs []Message) []tuiRuntimeConversationKeyAt {
	var out []tuiRuntimeConversationKeyAt
	for i, k := range tuiRuntimeConversationKeys(owner, msgs) {
		out = append(out, tuiRuntimeConversationKeyAt{key: k, at: tuiRuntimeNthAssistant(msgs, i)})
	}
	return out
}

func tuiRuntimeNthAssistant(msgs []Message, n int) int {
	for i, m := range msgs {
		if m.Role == "assistant" {
			if n == 0 {
				return i
			}
			n--
		}
	}
	return -1
}

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
