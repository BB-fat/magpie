package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const tuiRuntimeInteractiveSessionStateVersion = 1

type tuiRuntimeInteractiveRestoreStatus string

const (
	tuiRuntimeInteractiveRestoreExact              tuiRuntimeInteractiveRestoreStatus = "exact"
	tuiRuntimeInteractiveRestoreReplyAnchor        tuiRuntimeInteractiveRestoreStatus = "reply_anchor"
	tuiRuntimeInteractiveRestoreCompanion          tuiRuntimeInteractiveRestoreStatus = "turn_companion"
	tuiRuntimeInteractiveRestoreStateMissing       tuiRuntimeInteractiveRestoreStatus = "state_missing"
	tuiRuntimeInteractiveRestoreStateInvalid       tuiRuntimeInteractiveRestoreStatus = "state_invalid"
	tuiRuntimeInteractiveRestoreDirty              tuiRuntimeInteractiveRestoreStatus = "dirty"
	tuiRuntimeInteractiveRestoreContextMismatch    tuiRuntimeInteractiveRestoreStatus = "context_mismatch"
	tuiRuntimeInteractiveRestoreCheckpointNotFound tuiRuntimeInteractiveRestoreStatus = "checkpoint_not_found"
	tuiRuntimeInteractiveRestoreSuffixRejected     tuiRuntimeInteractiveRestoreStatus = "suffix_rejected"
)

type tuiRuntimeConversationSuffixStatus uint8

const (
	tuiRuntimeConversationSuffixOK tuiRuntimeConversationSuffixStatus = iota
	tuiRuntimeConversationSuffixCheckpointNotFound
	tuiRuntimeConversationSuffixRejected
)

type tuiRuntimeInteractiveSessionEntry struct {
	SessionID  string    `json:"session_id"`
	Model      string    `json:"model,omitempty"`
	ThreadKey  string    `json:"thread_key,omitempty"`
	ConvKey    string    `json:"conv_key"`
	ReplyKey   string    `json:"reply_key,omitempty"`
	ContextKey string    `json:"context_key"`
	Dirty      bool      `json:"dirty,omitempty"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type tuiRuntimeInteractiveSessionState struct {
	Version  int                                          `json:"version"`
	Sessions map[string]tuiRuntimeInteractiveSessionEntry `json:"sessions"`
}

func tuiRuntimeInteractiveSessionStatePath() string {
	return filepath.Join(tuiRuntimeMustStateDir(), "claude-interactive-sessions.json")
}

func tuiRuntimeInteractiveSessionStateKey(owner, outerSession string) string {
	sum := sha256.Sum256([]byte(owner + "\x00" + outerSession))
	return hex.EncodeToString(sum[:])
}

func tuiRuntimeInteractiveSessionWorkDir(owner, outerSession string) (string, error) {
	// Keep the cwd stable for this outer conversation, but outside the home:
	// Claude Code includes git status for a repository containing its cwd in
	// the system prompt, so a cache directory under a dotfiles-managed home can
	// both leak unrelated paths and invalidate the prompt-cache prefix. This is
	// the same concern addressed by #626 for headless subscription runs; the
	// interactive backend needs a per-conversation directory because --resume
	// resolves sessions by project/cwd.
	return tuiRuntimeOwnedWorkDir(filepath.Join(tuiRuntimeTempRoot(), "magpie-claude-interactive-"+tuiRuntimeInteractiveSessionStateKey(owner, outerSession)))
}

// ownedWorkDir makes dir, a fixed cwd for native runs, and checks this user
// alone can write it.
func tuiRuntimeOwnedWorkDir(dir string) (string, error) {
	if err := os.Mkdir(dir, 0o700); err != nil && !os.IsExist(err) {
		return "", err
	}
	fi, err := os.Lstat(dir)
	if err != nil {
		return "", err
	}
	if !fi.IsDir() {
		return "", fmt.Errorf("%s is not a folder", dir)
	}
	if tuiRuntimeOthersCanWrite(fi) {
		return "", fmt.Errorf("%s can be written by others (%v)", dir, fi.Mode().Perm())
	}
	if !tuiRuntimeOwnedByMe(fi) {
		return "", fmt.Errorf("%s belongs to another user", dir)
	}
	// An owned folder may still be read-only or inaccessible.
	probe, err := os.CreateTemp(dir, ".magpie-owner-check-*")
	if err != nil {
		return "", fmt.Errorf("%s is not writable by this user: %w", dir, err)
	}
	probeName := probe.Name()
	if err := probe.Close(); err != nil {
		_ = os.Remove(probeName)
		return "", err
	}
	_ = os.Remove(probeName)
	return dir, nil
}

// Reads and mutations use the same bounded validator under stateMu. A
// preflight cannot protect a later lookup from a replaced/truncated file;
// a damaged store must never become a missing conversation.
const tuiRuntimeMaxInteractiveSessionStoreBytes = 16 << 20

var tuiRuntimeErrInteractiveSessionStoreInvalid = errors.New("Claude interactive checkpoint store invalid; refusing native history replay")

func tuiRuntimeLoadInteractiveSessionStateStrict() (tuiRuntimeInteractiveSessionState, error) {
	empty := tuiRuntimeInteractiveSessionState{Version: tuiRuntimeInteractiveSessionStateVersion, Sessions: make(map[string]tuiRuntimeInteractiveSessionEntry)}
	path := tuiRuntimeInteractiveSessionStatePath()
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return empty, nil
	}
	if err != nil {
		return empty, fmt.Errorf("checkpoint store stat failed: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		return empty, errors.New("checkpoint store is not a private regular file")
	}
	if info.Size() > tuiRuntimeMaxInteractiveSessionStoreBytes {
		return empty, errors.New("checkpoint store exceeds size limit")
	}
	f, err := os.Open(path)
	if err != nil {
		return empty, fmt.Errorf("checkpoint store open failed: %w", err)
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !os.SameFile(info, opened) {
		return empty, errors.New("checkpoint store changed during open")
	}
	raw, err := io.ReadAll(io.LimitReader(f, tuiRuntimeMaxInteractiveSessionStoreBytes+1))
	if err != nil {
		return empty, fmt.Errorf("checkpoint store read failed: %w", err)
	}
	if len(raw) > tuiRuntimeMaxInteractiveSessionStoreBytes {
		return empty, errors.New("checkpoint store exceeds size limit")
	}
	var state tuiRuntimeInteractiveSessionState
	if json.Unmarshal(raw, &state) != nil || state.Version != tuiRuntimeInteractiveSessionStateVersion || state.Sessions == nil {
		return empty, errors.New("checkpoint store is malformed or has an incompatible version")
	}
	for _, entry := range state.Sessions {
		// ReplyKey is optional for a dirty first-tool anchor, but these
		// three fields are mandatory for both clean and dirty checkpoints.
		// An empty nested JSON object is not a valid older native session.
		if entry.SessionID == "" || entry.ConvKey == "" || entry.ContextKey == "" {
			return empty, errors.New("checkpoint store contains an incomplete session entry")
		}
	}
	return state, nil
}

// Check before matching live runs as well as at each later state lookup.
func (b *tuiRuntimeSubscriptionBridge) verifyInteractiveSessionStore() error {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	_, err := tuiRuntimeLoadInteractiveSessionStateStrict()
	return err
}

func tuiRuntimeWriteInteractiveSessionState(state tuiRuntimeInteractiveSessionState) error {
	dir := filepath.Dir(tuiRuntimeInteractiveSessionStatePath())
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	// A bounded state file is enough for continuity while avoiding an
	// unbounded collection of conversations that can no longer share a live
	// Anthropic prompt cache anyway.
	cutoff := time.Now().Add(-30 * 24 * time.Hour)
	for k, v := range state.Sessions {
		if !v.UpdatedAt.IsZero() && v.UpdatedAt.Before(cutoff) {
			delete(state.Sessions, k)
		}
	}
	b, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	if len(b) > tuiRuntimeMaxInteractiveSessionStoreBytes {
		return errors.New("checkpoint update exceeds store size limit")
	}
	tmp, err := os.CreateTemp(dir, ".claude-interactive-sessions-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(b); err != nil {
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
	if err := os.Rename(name, tuiRuntimeInteractiveSessionStatePath()); err != nil {
		// Never delete the last checkpoint to retry a failed replacement.
		return err
	}
	parent, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer parent.Close()
	return parent.Sync()
}

func (b *tuiRuntimeSubscriptionBridge) loadInteractiveSession(owner, outerSession string) (tuiRuntimeInteractiveSessionEntry, bool, error) {
	if outerSession == "" {
		return tuiRuntimeInteractiveSessionEntry{}, false, nil
	}
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	state, err := tuiRuntimeLoadInteractiveSessionStateStrict()
	if err != nil {
		log.Printf("Claude interactive checkpoint lookup rejected: %v", err)
		return tuiRuntimeInteractiveSessionEntry{}, false, tuiRuntimeErrInteractiveSessionStoreInvalid
	}
	e, ok := state.Sessions[tuiRuntimeInteractiveSessionStateKey(owner, outerSession)]
	return e, ok, nil
}

func (b *tuiRuntimeSubscriptionBridge) saveInteractiveSession(owner, outerSession string, e tuiRuntimeInteractiveSessionEntry) error {
	if outerSession == "" || e.SessionID == "" || e.ConvKey == "" || e.ContextKey == "" {
		return errors.New("incomplete interactive checkpoint entry")
	}
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	state, err := tuiRuntimeLoadInteractiveSessionStateStrict()
	if err != nil {
		return fmt.Errorf("refusing to overwrite checkpoint store: %w", err)
	}
	e.UpdatedAt = time.Now().UTC()
	state.Sessions[tuiRuntimeInteractiveSessionStateKey(owner, outerSession)] = e
	if err := tuiRuntimeWriteInteractiveSessionState(state); err != nil {
		log.Printf("Claude interactive session state: %v", err)
		return err
	}
	return nil
}

// Persist a first-turn recovery anchor only after the real native Claude
// session ID and the first registered callback are both known. This anchor
// is always Dirty, lacks an assistant reply, and can authorize only the
// existing tool-free recovery flow; it MUST NOT be treated as a completed
// transcript checkpoint or used to resume an uncertain native tool.
func (r *tuiRuntimeSubscriptionRun) persistFirstToolAnchorLocked() error {
	if r == nil || r.firstTurnAnchorKey == "" || r.sessionID == "" || r.outerSession == "" ||
		r.bridge == nil || r.bridge.toolJournal == nil || len(r.seenMCPCallIDs) == 0 {
		return nil
	}
	b := r.bridge
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	state, err := tuiRuntimeLoadInteractiveSessionStateStrict()
	if err != nil {
		return fmt.Errorf("refusing first-turn anchor: %w", err)
	}
	key := tuiRuntimeInteractiveSessionStateKey(r.owner, r.outerSession)
	if old, ok := state.Sessions[key]; ok {
		if old.SessionID == r.sessionID && old.Dirty && old.ConvKey == r.firstTurnAnchorKey && old.ReplyKey == "" {
			return nil // second callback or a repeated native init event
		}
		return errors.New("a distinct checkpoint already owns this conversation")
	}
	state.Sessions[key] = tuiRuntimeInteractiveSessionEntry{
		SessionID: r.sessionID, Model: r.model, ThreadKey: r.threadKey,
		ConvKey: r.firstTurnAnchorKey, ContextKey: r.interactiveContextKey,
		Dirty: true, UpdatedAt: time.Now().UTC(),
	}
	return tuiRuntimeWriteInteractiveSessionState(state)
}

func (b *tuiRuntimeSubscriptionBridge) markInteractiveSessionDirty(owner, outerSession string) error {
	if outerSession == "" {
		return nil
	}
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	state, err := tuiRuntimeLoadInteractiveSessionStateStrict()
	if err != nil {
		return fmt.Errorf("refusing dirty update: %w", err)
	}
	key := tuiRuntimeInteractiveSessionStateKey(owner, outerSession)
	e, ok := state.Sessions[key]
	if !ok {
		return errors.New("checkpoint missing; refusing native resume without dirty state")
	}
	if e.Dirty {
		return nil
	}
	e.Dirty = true
	e.UpdatedAt = time.Now().UTC()
	state.Sessions[key] = e
	if err := tuiRuntimeWriteInteractiveSessionState(state); err != nil {
		return fmt.Errorf("persisting dirty checkpoint: %w", err)
	}
	return nil
}

// claimInteractiveOuterTurn gives one in-flight interactive turn per
// (owner, outerSession). Desktop retries can deliver two requests for the
// same outer conversation concurrently; only the first may start a run,
// whichever kind of turn it is (a fresh first turn, a persisted resume, or
// an opt-in read-only recovery). The claim is released by the caller when
// the HTTP reply finishes — never held across turns and never persisted —
// so a restart drops it, and a replay then costs at most one extra
// tool-free turn, which is the documented bound.
func (b *tuiRuntimeSubscriptionBridge) claimInteractiveOuterTurn(owner, outerSession string) bool {
	if outerSession == "" {
		return false
	}
	key := tuiRuntimeInteractiveSessionStateKey(owner, outerSession)
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.outerTurns[key] {
		return false
	}
	b.outerTurns[key] = true
	return true
}

func (b *tuiRuntimeSubscriptionBridge) releaseInteractiveOuterTurn(owner, outerSession string) {
	if outerSession == "" {
		return
	}
	key := tuiRuntimeInteractiveSessionStateKey(owner, outerSession)
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.outerTurns, key)
}

// errInteractiveOuterLive refuses a canonical start on an outer
// conversation that a live canonical run still owns; serveMessagesInner
// maps it to 409, like the per-request outerTurns claim.
var tuiRuntimeErrInteractiveOuterLive = errors.New("Claude interactive conversation has a live run; refusing a second PTY")

// interactiveOuterLive reports whether a canonical interactive run owns
// (owner, outerSession): the run was published against it and its child
// has not been reaped and removed yet. Callers on the HTTP path hold the
// per-request outerTurns claim, so the check serializes with every
// canonical start: a run registers liveOuter while its starting request
// still holds that claim, which means any later start attempt sees it.
func (b *tuiRuntimeSubscriptionBridge) interactiveOuterLive(owner, outerSession string) bool {
	if outerSession == "" {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.liveOuter[tuiRuntimeInteractiveSessionStateKey(owner, outerSession)] != nil
}

// interactiveOuterOwner returns the canonical run that currently owns
// (owner, outerSession), or nil when it is free.
func (b *tuiRuntimeSubscriptionBridge) interactiveOuterOwner(owner, outerSession string) *tuiRuntimeSubscriptionRun {
	if outerSession == "" {
		return nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.liveOuter[tuiRuntimeInteractiveSessionStateKey(owner, outerSession)]
}

// interactiveTerminalCleanupWait bounds how long a next-turn request
// waits for a proven terminal-delivered owner to finish reaping — a
// wrapper's tmux EXIT_HOLD teardown is ~2s; the bound is short so a run
// that is merely slow, stuck or actually still working can never be
// mistaken for a finished one by a caller that waits longer.
var tuiRuntimeInteractiveTerminalCleanupWait = 6 * time.Second

// errInteractiveTerminalBusy refuses a canonical start while the outer's
// terminal-delivered owner did not finish reaping inside the wait bound:
// fail closed like errInteractiveOuterLive, never a second PTY.
var tuiRuntimeErrInteractiveTerminalBusy = errors.New("Claude interactive conversation's last run is still cleaning up; refusing a second PTY")

// interactiveTerminalDelivered reports whether ownerRun is the proven
// terminal-delivered canonical run for (owner, outerSession): it is the
// run still registered in liveOuter, and ended() marked it
// terminalDelivered — which is only ever set after a complete non-tool
// reply was delivered AND the clean canonical checkpoint for this same
// (owner, outerSession) was written. The persisted entry is re-checked
// to be clean and to belong to this run's inner session, so a stale or
// dirty checkpoint, a helper run, a parked tool turn, a failed/cancelled
// turn and a read-only-signaled turn are all refused, not waited on.
func (b *tuiRuntimeSubscriptionBridge) interactiveTerminalDelivered(ownerRun *tuiRuntimeSubscriptionRun, saved tuiRuntimeInteractiveSessionEntry, savedOK bool) bool {
	if ownerRun == nil || !savedOK || saved.Dirty {
		return false
	}
	ownerRun.mu.Lock()
	terminal, sessionID := ownerRun.terminalDelivered, ownerRun.sessionID
	ownerRun.mu.Unlock()
	// ownerRun was taken from liveOuter[(owner,outerSession)], which binds
	// it to this conversation; the checkpoint must be the clean one this
	// same run's inner session produced.
	return terminal && sessionID != "" && sessionID == saved.SessionID
}

// waitInteractiveTerminalOwner waits, outside b.mu, for a proven
// terminal-delivered canonical run to release its outer owner slot.
//
// ownerRun must already be classified terminal-delivered by the caller
// (ended() delivered a complete non-tool reply AND saved the clean
// canonical checkpoint — run.terminalDelivered). Anything else — a live
// executing run, a parked tool turn, a failed/cancelled turn, a
// read-only-signaled turn — is refused, not waited on.
//
// The wait never forces the slot free, never clears a dirty checkpoint
// and never starts a second PTY: on ctx cancel or the bounded timeout it
// reports busy so the caller fails closed while the old run keeps owning
// the outer until it is actually reaped and removeRun drops the slot.
// Because ownerRun is the liveOuter *value* published under b.mu, a new
// run can only take the slot after the old one's removeRun deleted it —
// so observing the slot empty or re-owned by a different run proves the
// old child was reaped; observing the same run still there proves it has
// not been reaped yet and the wait must keep failing closed.
func (b *tuiRuntimeSubscriptionBridge) waitInteractiveTerminalOwner(ctx context.Context, owner, outerSession string, ownerRun *tuiRuntimeSubscriptionRun) error {
	key := tuiRuntimeInteractiveSessionStateKey(owner, outerSession)
	deadline := time.Now().Add(tuiRuntimeInteractiveTerminalCleanupWait)
	for {
		b.mu.Lock()
		current := b.liveOuter[key]
		b.mu.Unlock()
		if current != ownerRun {
			return nil // slot freed (nil) or re-owned by a reaped-and-replaced run
		}
		if !time.Now().Before(deadline) {
			return tuiRuntimeErrInteractiveTerminalBusy
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Millisecond):
		}
	}
}

// assistantReplyKey is a history-independent anchor for the last completed
// reply saved in the inner Claude session. Unlike ConvKey it survives an
// outer client compacting or rewriting messages that came before that reply.
func tuiRuntimeAssistantReplyKey(msg Message) string {
	if msg.Role != "assistant" {
		return ""
	}
	h := sha256.New()
	tuiRuntimeHashMessages(h, []Message{msg}, nil)
	return hex.EncodeToString(h.Sum(nil))
}

func tuiRuntimeClaudeDesktopTurnCompanionMessage(m Message) bool {
	if m.Role != "user" {
		return false
	}
	// A synthetic continuation must never smuggle a client-side tool
	// result into a segment that the resume checker may discard.
	for _, p := range m.Parts {
		if p.Kind != Text {
			return false
		}
	}
	for _, p := range m.Parts {
		if p.Kind != Text {
			continue
		}
		text := strings.TrimSpace(p.Text)
		if strings.HasPrefix(text, "[Your previous response had no visible output.") ||
			strings.HasPrefix(text, "Your response above was cut off mid-stream.") {
			return true
		}
	}
	return false
}

func tuiRuntimeClaudeDesktopTurnCompanionSuffix(req *tuiRuntimeRequest) []Message {
	for i := len(req.Messages) - 1; i >= 0; i-- {
		m := req.Messages[i]
		if m.Role != "user" {
			continue
		}
		if tuiRuntimeClaudeDesktopTurnCompanionMessage(m) {
			return []Message{m}
		}
		return nil
	}
	return nil
}

// A Desktop-generated interruption hint can only continue the saved native
// thread if it is the *only semantic* post-checkpoint turn. Desktop may
// interleave volatile token-budget system notices: permit only that known
// generated form, not an assistant reply, old MCP result, other user
// instruction, duplicate companion or arbitrary system message.
func tuiRuntimeSafeDesktopCompanionTail(msgs []Message) bool {
	seen := false
	for _, m := range msgs {
		if tuiRuntimeClaudeDesktopTurnCompanionMessage(m) {
			if seen {
				return false
			}
			seen = true
			continue
		}
		if m.Role != "system" || len(tuiRuntimeWithoutDynamicTokenBudgetNotices([]Message{m})) != 0 {
			return false
		}
	}
	return seen
}

func tuiRuntimeClaudeDesktopFailedTurnCompanionReply(m Message) bool {
	if m.Role != "assistant" || len(m.Parts) == 0 {
		return false
	}
	var text strings.Builder
	for _, p := range m.Parts {
		if p.Kind != Text {
			// This recognition is used to *discard* a failed transport
			// artifact. Never discard an actual tool call or image.
			return false
		}
		text.WriteString(p.Text)
		if text.Len() > 8192 {
			return false
		}
	}
	s := strings.TrimSpace(text.String())
	if !strings.HasPrefix(s, "API Error: 502 ") {
		return false
	}
	// A clean checkpoint can reject a Desktop-generated follow-up for
	// several reasons. In all these cases startInteractiveRequest exited
	// before starting a new native run. Restrict the elision to the known
	// transport signature and a synthetic, tool-free Desktop companion.
	for _, reason := range []tuiRuntimeInteractiveRestoreStatus{
		tuiRuntimeInteractiveRestoreDirty,
		tuiRuntimeInteractiveRestoreSuffixRejected,
		tuiRuntimeInteractiveRestoreCheckpointNotFound,
		tuiRuntimeInteractiveRestoreContextMismatch,
		tuiRuntimeInteractiveRestoreStateMissing,
	} {
		if strings.Contains(s, "Claude interactive persisted resume failed: "+string(reason)+"; refusing full history replay") {
			return true
		}
	}
	return false
}

// restoreInteractiveSession decides whether a persisted inner Claude session
// can safely accept only the caller's new suffix. It never authorizes replay
// of the entire outer history.
func (b *tuiRuntimeSubscriptionBridge) restoreInteractiveSession(owner, outerSession string, req *tuiRuntimeRequest) (tuiRuntimeInteractiveSessionEntry, []Message, tuiRuntimeInteractiveRestoreStatus) {
	saved, ok, err := b.loadInteractiveSession(owner, outerSession)
	if err != nil {
		return tuiRuntimeInteractiveSessionEntry{}, nil, tuiRuntimeInteractiveRestoreStateInvalid
	}
	if !ok {
		return tuiRuntimeInteractiveSessionEntry{}, nil, tuiRuntimeInteractiveRestoreStateMissing
	}
	if suffix := tuiRuntimeClaudeDesktopTurnCompanionSuffix(req); suffix != nil {
		// A recognizable last user message is not sufficient evidence to
		// resume a native session. Validate the saved conversation hash or
		// unique last-reply anchor *before* skipping any outer history.
		// This applies even to a dirty checkpoint: an unrelated client or
		// previously completed tool result must never bypass this check.
		try := func(since []Message, status tuiRuntimeConversationSuffixStatus) (tuiRuntimeInteractiveRestoreStatus, bool) {
			switch status {
			case tuiRuntimeConversationSuffixOK:
				if tuiRuntimeSafeDesktopCompanionTail(since) {
					return tuiRuntimeInteractiveRestoreCompanion, true
				}
				return tuiRuntimeInteractiveRestoreSuffixRejected, true
			case tuiRuntimeConversationSuffixRejected:
				return tuiRuntimeInteractiveRestoreSuffixRejected, true
			default:
				return tuiRuntimeInteractiveRestoreCheckpointNotFound, false
			}
		}
		since, status := tuiRuntimeConversationSuffixAfterDetailed(owner, req.Messages, saved.ConvKey)
		if decided, found := try(since, status); found {
			if decided == tuiRuntimeInteractiveRestoreCompanion {
				return saved, suffix, decided
			}
			return saved, nil, decided
		}
		if saved.ReplyKey != "" {
			since, status = tuiRuntimeConversationSuffixAfterReplyDetailed(req.Messages, saved.ReplyKey)
			if decided, found := try(since, status); found {
				if decided == tuiRuntimeInteractiveRestoreCompanion {
					return saved, suffix, decided
				}
				return saved, nil, decided
			}
		}
		return saved, nil, tuiRuntimeInteractiveRestoreCheckpointNotFound
	}
	if saved.Dirty {
		return saved, nil, tuiRuntimeInteractiveRestoreDirty
	}
	since, status := tuiRuntimeConversationSuffixAfterDetailed(owner, req.Messages, saved.ConvKey)
	switch status {
	case tuiRuntimeConversationSuffixOK:
		return saved, since, tuiRuntimeInteractiveRestoreExact
	case tuiRuntimeConversationSuffixRejected:
		return saved, nil, tuiRuntimeInteractiveRestoreSuffixRejected
	}
	if saved.ReplyKey == "" {
		return saved, nil, tuiRuntimeInteractiveRestoreCheckpointNotFound
	}
	since, status = tuiRuntimeConversationSuffixAfterReplyDetailed(req.Messages, saved.ReplyKey)
	switch status {
	case tuiRuntimeConversationSuffixOK:
		return saved, since, tuiRuntimeInteractiveRestoreReplyAnchor
	case tuiRuntimeConversationSuffixRejected:
		return saved, nil, tuiRuntimeInteractiveRestoreSuffixRejected
	default:
		return saved, nil, tuiRuntimeInteractiveRestoreCheckpointNotFound
	}
}

func tuiRuntimeConversationSuffixAfterDetailed(owner string, msgs []Message, convKey string) ([]Message, tuiRuntimeConversationSuffixStatus) {
	if convKey == "" {
		return nil, tuiRuntimeConversationSuffixCheckpointNotFound
	}
	h := sha256.New()
	_, _ = h.Write([]byte(owner + "\x00"))
	match := -1
	tuiRuntimeHashMessages(h, msgs, func(i int) {
		if msgs[i].Role == "assistant" && hex.EncodeToString(h.Sum(nil)) == convKey {
			match = i
		}
	})
	return tuiRuntimeSafeConversationSuffixAfterDetailed(msgs, match)
}

func tuiRuntimeConversationSuffixAfterReplyDetailed(msgs []Message, replyKey string) ([]Message, tuiRuntimeConversationSuffixStatus) {
	if replyKey == "" {
		return nil, tuiRuntimeConversationSuffixCheckpointNotFound
	}
	match := -1
	for i, m := range msgs {
		if m.Role != "assistant" || tuiRuntimeAssistantReplyKey(m) != replyKey {
			continue
		}
		if match >= 0 {
			return nil, tuiRuntimeConversationSuffixRejected
		}
		match = i
	}
	return tuiRuntimeSafeConversationSuffixAfterDetailed(msgs, match)
}

func tuiRuntimeSafeConversationSuffixAfterDetailed(msgs []Message, match int) ([]Message, tuiRuntimeConversationSuffixStatus) {
	if match < 0 || match >= len(msgs)-1 {
		if match < 0 {
			return nil, tuiRuntimeConversationSuffixCheckpointNotFound
		}
		return nil, tuiRuntimeConversationSuffixRejected
	}
	since := msgs[match+1:]
	// A failed hidden turn-companion plus Magpie's synthetic 502 reply was
	// never accepted by the inner Claude session. Skip only that exact pair.
	for len(since) >= 2 && tuiRuntimeClaudeDesktopTurnCompanionMessage(since[0]) && tuiRuntimeClaudeDesktopFailedTurnCompanionReply(since[1]) {
		since = since[2:]
	}
	if len(since) == 0 {
		return nil, tuiRuntimeConversationSuffixRejected
	}
	for _, m := range since {
		if m.Role == "assistant" {
			return nil, tuiRuntimeConversationSuffixRejected
		}
		for _, p := range m.Parts {
			if p.Kind == ToolResult {
				return nil, tuiRuntimeConversationSuffixRejected
			}
		}
	}
	return since, tuiRuntimeConversationSuffixOK
}
