package gateway

// Append-only, checksum-verified evidence of tool handoffs. This is an
// opt-in first-stage recovery ledger, NOT an execution replay log: storing
// "registered" does not prove a Desktop tool actually ran; storing
// "result_verified" does not prove native Claude consumed that result.
// Neither event grants permission to invoke a side-effecting tool again.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const tuiRuntimeToolJournalMaxBytes = 16 << 20

type tuiRuntimeInteractiveToolEvidence struct {
	Version  int    `json:"version"`
	Session  string `json:"session"` // hash(owner, outerSession), never raw identity
	Run      string `json:"run"`     // hash(ephemeral callback token)
	Call     string `json:"call"`    // hash(native tool_use_id)
	Tool     string `json:"tool"`
	ArgsHash string `json:"args_sha256,omitempty"`
	Result   string `json:"result_sha256,omitempty"`
	Stage    string `json:"stage"`
	At       string `json:"at"`
	Checksum string `json:"checksum"`
}

type tuiRuntimeInteractiveToolJournal struct {
	mu   sync.Mutex
	path string
}

func tuiRuntimeToolEvidenceHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// Native MCP and Anthropic tool_use can serialize identical JSON arguments
// with different insignificant whitespace or key ordering. Hash the parsed
// value rather than the spelling while retaining strict value equality.
func tuiRuntimeToolEvidenceArgumentsHash(raw json.RawMessage) (string, error) {
	var value any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber() // float64 would merge distinct integers > 2^53
	if len(raw) == 0 || dec.Decode(&value) != nil {
		return "", errors.New("invalid tool arguments")
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return "", errors.New("trailing tool argument data")
	}
	if _, ok := value.(map[string]any); !ok {
		return "", errors.New("tool arguments must be an object")
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return tuiRuntimeToolEvidenceHash(string(canonical)), nil
}

func tuiRuntimeToolEvidenceChecksum(e tuiRuntimeInteractiveToolEvidence) string {
	e.Checksum = ""
	data, _ := json.Marshal(e)
	return tuiRuntimeToolEvidenceHash(string(data))
}

func (j *tuiRuntimeInteractiveToolJournal) readLocked() ([]tuiRuntimeInteractiveToolEvidence, error) {
	info, statErr := os.Lstat(j.path)
	if statErr == nil {
		if !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
			return nil, errors.New("tool journal is not a private regular file")
		}
		if info.Size() > tuiRuntimeToolJournalMaxBytes {
			return nil, errors.New("tool journal exceeded maximum size")
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return nil, statErr
	}
	b, err := os.ReadFile(j.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(b) > tuiRuntimeToolJournalMaxBytes || (len(b) > 0 && b[len(b)-1] != '\n') {
		return nil, errors.New("tool journal oversized or truncated; refusing unsafe recovery")
	}
	var out []tuiRuntimeInteractiveToolEvidence
	byCall := map[string]tuiRuntimeInteractiveToolEvidence{}
	for _, line := range bytes.Split(bytes.TrimSuffix(b, []byte{'\n'}), []byte{'\n'}) {
		if len(line) == 0 {
			if len(b) == 0 {
				break
			}
			return nil, errors.New("empty tool journal record")
		}
		var e tuiRuntimeInteractiveToolEvidence
		if json.Unmarshal(line, &e) != nil || e.Version != 1 || e.Checksum == "" || e.Checksum != tuiRuntimeToolEvidenceChecksum(e) || e.Session == "" || e.Call == "" || e.Run == "" {
			return nil, errors.New("invalid tool journal record; refusing unsafe recovery")
		}
		key := e.Session + ":" + e.Call
		prev, exists := byCall[key]
		switch e.Stage {
		case "registered":
			if exists || e.ArgsHash == "" || e.Result != "" || e.Tool == "" {
				return nil, errors.New("invalid/duplicate tool registration")
			}
		case "result_verified":
			if !exists || prev.Stage != "registered" || e.Result == "" || prev.Run != e.Run || prev.Tool != e.Tool || prev.ArgsHash != e.ArgsHash {
				return nil, errors.New("unowned or repeated tool result")
			}
		case "auto_recovery_claimed":
			if !exists || (prev.Stage != "registered" && prev.Stage != "result_verified") || e.Result == "" || prev.Run != e.Run || prev.Tool != e.Tool || prev.ArgsHash != e.ArgsHash || (prev.Result != "" && prev.Result != e.Result) {
				return nil, errors.New("invalid or duplicated automatic recovery claim")
			}
		case "auto_recovery_retry_claimed":
			// Only a previously claimed read-only reconstruction may be
			// retried; identical IDs and result bytes are mandatory.
			if !exists || prev.Stage != "auto_recovery_claimed" || e.Result == "" || prev.Result != e.Result || prev.Run != e.Run || prev.Tool != e.Tool || prev.ArgsHash != e.ArgsHash {
				return nil, errors.New("invalid or duplicated read-only recovery retry")
			}
		default:
			return nil, errors.New("unsupported tool journal stage")
		}
		byCall[key] = e
		out = append(out, e)
	}
	return out, nil
}

func (j *tuiRuntimeInteractiveToolJournal) append(e tuiRuntimeInteractiveToolEvidence) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	unlock, err := tuiRuntimeLockInteractiveToolJournal(j.path)
	if err != nil {
		return err
	}
	defer unlock()
	old, err := j.readLocked()
	if err != nil {
		return err
	}
	// Reuse exactly the same monotone transition guard as the reader.
	key := e.Session + ":" + e.Call
	var prev *tuiRuntimeInteractiveToolEvidence
	for i := range old {
		if old[i].Session+":"+old[i].Call == key {
			prev = &old[i]
		}
	}
	if e.Stage == "registered" {
		if prev != nil || e.Tool == "" || e.ArgsHash == "" || e.Result != "" {
			return errors.New("tool ID already registered or malformed")
		}
	} else if e.Stage == "result_verified" {
		if prev == nil || prev.Stage != "registered" || prev.Tool != e.Tool || prev.ArgsHash != e.ArgsHash || prev.Run != e.Run || e.Result == "" {
			return errors.New("tool result does not match outstanding registration")
		}
	} else {
		return errors.New("invalid tool evidence stage")
	}
	e.Version, e.At = 1, time.Now().UTC().Format(time.RFC3339Nano)
	if e.Session == "" || e.Call == "" || e.Run == "" {
		return errors.New("missing tool ownership evidence")
	}
	e.Checksum = tuiRuntimeToolEvidenceChecksum(e)
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if len(line)+1+tuiRuntimeLenMustFileSize(j.path) > tuiRuntimeToolJournalMaxBytes {
		return errors.New("tool journal full; refusing unsafe continuation")
	}
	if err := os.MkdirAll(filepath.Dir(j.path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(j.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(append(line, '\n')); err != nil {
		return err
	}
	return f.Sync()
}

func tuiRuntimeLenMustFileSize(path string) int {
	st, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return int(st.Size())
}

func (j *tuiRuntimeInteractiveToolJournal) read() ([]tuiRuntimeInteractiveToolEvidence, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	unlock, err := tuiRuntimeLockInteractiveToolJournal(j.path)
	if err != nil {
		return nil, err
	}
	defer unlock()
	return j.readLocked()
}

func tuiRuntimeNewInteractiveToolJournal() *tuiRuntimeInteractiveToolJournal {
	if os.Getenv("MAGPIE_CLAUDE_DURABLE_TOOL_JOURNAL") != "1" {
		return nil
	}
	return &tuiRuntimeInteractiveToolJournal{path: filepath.Join(tuiRuntimeMustStateDir(), "claude-interactive-tools.jsonl")}
}

func tuiRuntimeToolEvidenceFor(run *tuiRuntimeSubscriptionRun, tool, id, argsDigest, resultDigest, stage string) (tuiRuntimeInteractiveToolEvidence, error) {
	if run == nil || !run.interactive || run.outerSession == "" || run.owner == "" || run.token == "" || id == "" {
		return tuiRuntimeInteractiveToolEvidence{}, fmt.Errorf("no canonical interactive tool owner")
	}
	return tuiRuntimeInteractiveToolEvidence{Session: tuiRuntimeInteractiveSessionStateKey(run.owner, run.outerSession), Run: tuiRuntimeToolEvidenceHash(run.token), Call: tuiRuntimeToolEvidenceHash(id), Tool: tool, ArgsHash: argsDigest, Result: resultDigest, Stage: stage}, nil
}

// verifyResults records only results that the normal handoff guard has
// already accepted. Never infer execution success from a registered callback,
// and never store a tool result body on disk. A checksum-match merely says
// Desktop supplied that result: it does NOT mean native Claude consumed it.
func (j *tuiRuntimeInteractiveToolJournal) verifyResults(run *tuiRuntimeSubscriptionRun, results []Part) error {
	if j == nil || run == nil || !run.interactive || run.outerSession == "" {
		return nil
	}
	records, err := j.read()
	if err != nil {
		return err
	}
	session, runID := tuiRuntimeInteractiveSessionStateKey(run.owner, run.outerSession), tuiRuntimeToolEvidenceHash(run.token)
	for _, result := range results {
		if result.Kind != ToolResult {
			return errors.New("non-tool result in durable handoff")
		}
		id := tuiRuntimeToolEvidenceHash(result.CallID)
		var registered *tuiRuntimeInteractiveToolEvidence
		for i := range records {
			e := &records[i]
			if e.Session == session && e.Call == id {
				registered = e
			}
		}
		if registered == nil || registered.Stage != "registered" || registered.Run != runID {
			return errors.New("tool result has no outstanding durable registration")
		}
		value, marshalErr := json.Marshal(result)
		if marshalErr != nil {
			return marshalErr
		}
		e := *registered
		e.Result = tuiRuntimeToolEvidenceHash(string(value))
		e.Stage = "result_verified"
		e.Checksum = ""
		if err := j.append(e); err != nil {
			return err
		}
		records = append(records, e)
	}
	return nil
}

// claimOrphanRecovery authorizes exactly ONE new, tool-free *model turn*,
// not execution of an external tool. Validation is deliberately narrower
// than the live callback path: the last assistant tool_use set must be
// present on the wire with every result, and each item must match a durable
// registration for the same outer session and same native run. Compacting
// the assistant tool_use away is fail-closed on this cross-process path.
// The claim is committed before a recovery process may start, so retrying
// the same orphaned tool_result never starts another turn.
func (j *tuiRuntimeInteractiveToolJournal) claimOrphanRecovery(owner, outer string, req *Request) error {
	if j == nil || owner == "" || outer == "" || req == nil {
		return errors.New("durable recovery unavailable")
	}
	if err := tuiRuntimeValidateInteractiveToolContinuation(req); err != nil {
		return err
	}
	clean, _, err := tuiRuntimeInteractiveRequestWithoutSystem(req)
	if err != nil {
		return err
	}
	msgs := clean.Messages
	start := len(msgs)
	for start > 0 && msgs[start-1].Role == "user" {
		start--
	}
	firstAssistant := start
	for firstAssistant > 0 && msgs[firstAssistant-1].Role == "assistant" {
		firstAssistant--
	}
	tools := map[string]Part{}
	for _, m := range msgs[firstAssistant:start] {
		for _, p := range m.Parts {
			if p.Kind == ToolCall {
				tools[p.ID] = p
			}
		}
	}
	results := map[string]Part{}
	for _, m := range msgs[start:] {
		for _, p := range m.Parts {
			if p.Kind == ToolResult {
				results[p.CallID] = p
			}
		}
	}
	if len(tools) == 0 || len(tools) != len(results) {
		return errors.New("incomplete tool continuation")
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	unlock, err := tuiRuntimeLockInteractiveToolJournal(j.path)
	if err != nil {
		return err
	}
	defer unlock()
	old, err := j.readLocked()
	if err != nil {
		return err
	}
	session := tuiRuntimeInteractiveSessionStateKey(owner, outer)
	latest := map[string]tuiRuntimeInteractiveToolEvidence{}
	mostRecentCall := ""
	for _, e := range old {
		if e.Session != session {
			continue
		}
		latest[e.Call] = e
		if e.Stage == "registered" {
			mostRecentCall = e.Call
		}
	}
	if mostRecentCall == "" {
		return errors.New("no durable native callback for this conversation")
	}
	resultIDs := tuiRuntimeResultsByHash(results)
	// A stale completed exchange cannot be mistaken for the interrupted
	// exchange when a newer native callback exists in the same conversation.
	if _, ok := resultIDs[mostRecentCall]; !ok {
		return errors.New("a newer native call exists")
	}
	claimed := make([]tuiRuntimeInteractiveToolEvidence, 0, len(tools))
	var runID string
	priorAutoClaim := false
	firstTool := true
	for rawID, tool := range tools {
		id := tuiRuntimeToolEvidenceHash(rawID)
		e, ok := latest[id]
		if !ok || (e.Stage != "registered" && e.Stage != "result_verified" && e.Stage != "auto_recovery_claimed") || e.Tool != tool.Name || e.ArgsHash == "" {
			return errors.New("unverified tool call identity or already recovered result")
		}
		thisClaim := e.Stage == "auto_recovery_claimed"
		if firstTool {
			firstTool, priorAutoClaim = false, thisClaim
		} else if thisClaim != priorAutoClaim {
			return errors.New("mixed one-shot claim state across parallel tools")
		}
		if runID == "" {
			runID = e.Run
		}
		if e.Run != runID {
			return errors.New("tool results have multiple native owners")
		}
		hash, err := tuiRuntimeToolEvidenceArgumentsHash(tool.Args)
		if err != nil || hash != e.ArgsHash {
			return errors.New("tool arguments do not match native callback")
		}
		value, err := json.Marshal(results[rawID])
		if err != nil {
			return err
		}
		resultHash := tuiRuntimeToolEvidenceHash(string(value))
		if e.Result != "" && e.Result != resultHash {
			return errors.New("tool result changed since verification")
		}
		e.Result = resultHash
		if priorAutoClaim {
			if os.Getenv("MAGPIE_CLAUDE_AUTO_RECOVER_RETRY") != "1" {
				return errors.New("read-only recovery was already claimed")
			}
			e.Stage = "auto_recovery_retry_claimed"
		} else {
			e.Stage = "auto_recovery_claimed"
		}
		e.At = time.Now().UTC().Format(time.RFC3339Nano)
		e.Checksum = tuiRuntimeToolEvidenceChecksum(e)
		claimed = append(claimed, e)
	}
	// A parallel sibling with an unresolved native callback may still be
	// executing. Never start reconstruction while it is unaccounted for.
	for id, e := range latest {
		if e.Run == runID && e.Stage == "registered" {
			if _, exists := resultIDs[id]; !exists {
				return errors.New("unresolved sibling callback in same native run")
			}
		}
	}
	// Write one fsynced batch: a crash during the write yields a truncated
	// ledger that refuses recovery instead of granting ambiguous permission.
	var batch []byte
	for _, e := range claimed {
		line, err := json.Marshal(e)
		if err != nil {
			return err
		}
		batch = append(batch, line...)
		batch = append(batch, '\n')
	}
	if tuiRuntimeLenMustFileSize(j.path)+len(batch) > tuiRuntimeToolJournalMaxBytes {
		return errors.New("tool journal full")
	}
	f, err := os.OpenFile(j.path, os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(batch); err != nil {
		return err
	}
	return f.Sync()
}

func tuiRuntimeResultsByHash(results map[string]Part) map[string]bool {
	out := make(map[string]bool, len(results))
	for id := range results {
		out[tuiRuntimeToolEvidenceHash(id)] = true
	}
	return out
}
