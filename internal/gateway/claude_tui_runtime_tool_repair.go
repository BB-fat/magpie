package gateway

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/yetone/magpie/internal/claudebridge"
)

// A tool round with a bare name has no native MCP waiter. Keep its prefix
// and cwd until the client supplies the results, then resume a corrected copy.
type tuiRuntimeToolRepair struct {
	mu              sync.Mutex
	rows            [][]byte
	session, parent string
	work, workKey   string
	release         func()
	timer           *time.Timer
	claimed         bool
}

func (s *tuiRuntimeToolRepair) discard() {
	s.mu.Lock()
	release := s.release
	s.release, s.rows = nil, nil
	s.timer.Stop()
	s.mu.Unlock()
	if release != nil {
		release()
	}
}

func (s *tuiRuntimeToolRepair) claim() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.claimed {
		return false
	}
	s.claimed = true
	s.timer.Stop()
	return true
}

func (s *tuiRuntimeToolRepair) expire() {
	s.mu.Lock()
	expired := !s.claimed
	s.claimed = true
	s.mu.Unlock()
	if expired {
		s.discard()
	}
}

func (r *tuiRuntimeSubscriptionRun) keepToolRepair(buffered []tuiRuntimeShownCall) error {
	r.mu.Lock()
	calls := append(slices.Clone(r.shown), buffered...)
	sid := r.sessionID
	r.mu.Unlock()
	data, _, err := tuiRuntimeReadTranscript(r.nativeConfigDir, r.nativeWork, sid)
	if err != nil {
		return err
	}
	ids, names := make([]string, 0, len(calls)), map[string]string{}
	for _, c := range calls {
		ids = append(ids, c.id)
		names[c.id] = tuiRuntimeMcpToolPrefix() + claudebridge.ToolName(c.name)
	}
	rows, parent, err := tuiRuntimeToolRoundLines(data, ids, true)
	if err != nil {
		return err
	}
	for i, row := range rows {
		rows[i], err = tuiRuntimeRepairToolNames(row, names)
		if err != nil {
			return err
		}
	}
	s := &tuiRuntimeToolRepair{rows: rows, session: sid, parent: parent, work: r.nativeWork, workKey: r.workKey}
	r.mu.Lock()
	for _, call := range calls {
		if r.seenMCPCallIDs[call.id] {
			continue
		}
		if err := r.registerToolLocked(call.name, call.id, json.RawMessage(call.args)); err != nil {
			r.mu.Unlock()
			return err
		}
	}
	if err := r.persistFirstToolAnchorLocked(); err != nil {
		r.mu.Unlock()
		return err
	}
	// Stateless cleanup owns the cwd slot. Transfer it with the prefix so
	// another stateless launch cannot remove this branch's files underneath it.
	if release := r.releaseWork; release != nil {
		s.release = func() { release(sid) }
		r.releaseWork = nil
	}
	s.timer = time.AfterFunc(tuiRuntimeInteractiveParkLongest, s.expire)
	r.toolRepair = s
	r.mu.Unlock()
	return nil
}

func tuiRuntimeRepairToolNames(line []byte, names map[string]string) ([]byte, error) {
	var row map[string]json.RawMessage
	if err := json.Unmarshal(line, &row); err != nil {
		return nil, err
	}
	if string(row["type"]) != `"assistant"` {
		return line, nil
	}
	var message map[string]json.RawMessage
	if err := json.Unmarshal(row["message"], &message); err != nil {
		return nil, err
	}
	var blocks []map[string]json.RawMessage
	if err := json.Unmarshal(message["content"], &blocks); err != nil {
		return nil, err
	}
	changed := false
	for _, block := range blocks {
		if string(block["type"]) != `"tool_use"` {
			continue
		}
		var id, name string
		_ = json.Unmarshal(block["id"], &id)
		_ = json.Unmarshal(block["name"], &name)
		if canonical := names[id]; canonical != "" && canonical != name {
			block["name"], _ = json.Marshal(canonical)
			changed = true
		}
	}
	if !changed {
		return line, nil
	}
	message["content"], _ = json.Marshal(blocks)
	row["message"], _ = json.Marshal(message)
	return json.Marshal(row)
}

func (b *tuiRuntimeSubscriptionBridge) takeToolRepair(owner, model string, req *tuiRuntimeRequest) (*tuiRuntimeToolRepair, error) {
	_, results, _ := tuiRuntimeLiteTail(req.Messages)
	var run *tuiRuntimeSubscriptionRun
	var parts []Part
	b.mu.Lock()
	for _, m := range results {
		for _, p := range m.Parts {
			parts = append(parts, p)
			if run == nil {
				run = b.liteParked[p.CallID]
			}
		}
	}
	b.mu.Unlock()
	if run == nil || run.owner != owner || run.model != model {
		return nil, nil
	}
	run.mu.Lock()
	repair := run.toolRepair
	run.mu.Unlock()
	if repair == nil || !run.interactiveEquivalentContext(req) || !run.follows(req.Messages) ||
		tuiRuntimeValidateInteractiveMatchedContinuation(req, run, parts, tuiRuntimeByExactID) != nil || !repair.claim() {
		return nil, nil
	}
	if err := b.toolJournal.verifyResults(run, parts); err != nil {
		repair.discard()
		return nil, err
	}
	tuiRuntimeKeepRunMode(req, run)
	b.mu.Lock()
	for _, p := range parts {
		if b.liteParked[p.CallID] == run {
			delete(b.liteParked, p.CallID)
		}
	}
	b.mu.Unlock()
	return repair, nil
}

func (s *tuiRuntimeToolRepair) write(configDir, model string, messages []Message) (string, error) {
	dir := tuiRuntimeClaudeSessions(configDir, s.work)[0]
	id, f, err := tuiRuntimeCopyTranscript(dir, s.session, s.rows)
	if err != nil {
		return "", err
	}
	_, results, turn := tuiRuntimeLiteTail(messages)
	err = tuiRuntimeAppendClaudeTranscript(f, id, s.parent, s.work, model, nil, append(results, turn...), "", "")
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(filepath.Join(dir, id+".jsonl"))
		return "", err
	}
	return id, nil
}
