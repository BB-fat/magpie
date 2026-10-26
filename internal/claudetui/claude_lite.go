package claudetui

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// A client of the Messages API sends its whole conversation with every
// request and runs every client tool itself. Native Claude gets those
// tools through MCP, plus WebSearch when the gateway is not searching for
// it. So the request is the truth. It continues the live or checkpointed native session where it
// exactly can; otherwise the native session is rebuilt from the request's
// own structured history — tool calls and their results included, so the
// model sees what already ran — instead of being refused or answered by a
// tool-free recovery turn. A rebuild cannot run a tool twice: only the
// client runs tools, and only the calls it is handed.
//
// A conversation's helper calls (titles, subagents, other models) keep
// upstream's routing: they run stateless, beside the conversation's native
// session and checkpoint. So does a request without a conversation id.

// liteTurnWait bounds how long a conversation's turn waits for its previous
// turn to let go of it (a reply still unwinding after its client left, a
// parked run being stopped). A conversation still generating after it is
// answered 503 with Retry-After: a second native run of one conversation,
// able to call tools beside the first, is never started.
var liteTurnWait = 60 * time.Second

// liteContinueText is the turn a rebuilt session is given when the request
// ends in tool results with nothing else to say.
const liteContinueText = "Continue from the tool results above."

// liteResultsReceived closes a rebuilt history that ends in tool results.
// Resuming a session whose last message is the user's, Claude Code answers
// that message itself ("No response requested.") into the new reply.
const liteResultsReceived = "Tool results received."

// statelessSession is the signed bridge session of a request that has no
// conversation of its own (the plugin found no stable conversation id).
const statelessSession = "stateless"

// configDir is the Claude config folder the account runs in ("" for
// Claude Code's own, as the plugin's runtime ran).
func (rt *runtimeTUI) serveLite(out responder, r *http.Request, model, owner, outerSession, configDir string, body []byte) (int, string) {
	req, err := parseAnthropic(body)
	if err != nil {
		return out.Fail(http.StatusBadRequest, err.Error())
	}
	req.Model = model
	if len(req.Safeguards) > 0 {
		req.SafeguardBeta = "dangerous-tool-use-2026-09-03"
		for _, beta := range strings.Split(r.Header.Get("anthropic-beta"), ",") {
			if beta = strings.TrimSpace(beta); strings.HasPrefix(beta, "dangerous-tool-use-") {
				req.SafeguardBeta = beta
				break
			}
		}
	}
	// Rebuilding is not permission to erase a mismatched tool result. When
	// the caller's visible final assistant turn contains tool_use, every
	// returned ID must match that turn, even if the old native run is gone.
	// Compacted histories without a visible call remain eligible for rebuild.
	if err := liteVisibleToolHandoff(req); err != nil {
		msg := "Claude interactive tool result does not match its visible tool call; refusing"
		return out.Fail(http.StatusConflict, msg)
	}
	// The same tool results again, after their reply was delivered: a client
	// retrying. Only the very same request gets that reply again; the same
	// history asked under another system prompt, tools or settings is run.
	req.replayKey = liteReplayKey(owner, req)
	if completed, ok := rt.bridge.liteCompletedToolTurn(owner, outerSession, req); ok {
		// The original native run already returned a complete answer.
		// Return precisely that answer; starting another native inference
		// could generate the same tool_use a second time.
		log.Printf("Claude interactive tool results already answered; resending complete text without a new native run")
		if completed.reply == "" {
			completed.reply = "The tool result was already processed. Send a new user message to continue."
		}
		events := make(chan Event, 3)
		events <- Event{Kind: KStart, Model: model, ID: "msg_" + newID()}
		events <- Event{Kind: KText, Text: completed.reply}
		events <- Event{Kind: KStop, Stop: "stop"}
		close(events)
		return out.Relay(req, events, func() {}, func(string, string, bool) {})
	}
	upstream := func() (int, string) {
		start := func(ctx context.Context, req *Request) (*subscriptionRun, <-chan Event, error) {
			return rt.bridge.startInteractiveRequest(ctx, req, model, configDir, owner, outerSession)
		}
		return rt.serveMessagesInner(out, r, "Claude Code", model, owner, outerSession, body, start)
	}
	// Desktop's own compaction and suggestion calls already run isolated.
	if claudeDesktopCompaction(req) || claudeDesktopSuggestion(req) {
		return upstream()
	}
	// The exact continuation of a live run — its tool results, perhaps with
	// what the user said meanwhile — goes on in that native session.
	if run, results, how := rt.bridge.match(req); run != nil && liteContinues(owner, model, req, run, results, how) {
		return upstream()
	}
	if outerSession != "" {
		if err := rt.bridge.verifyInteractiveSessionStore(); err != nil {
			log.Printf("Claude interactive checkpoint store integrity rejected: %v", err)
			msg := "Claude interactive checkpoint store invalid"
			return out.Fail(http.StatusInternalServerError, msg)
		}
	}
	if outerSession != "" && !rt.bridge.liteCanonical(owner, outerSession, model, req) {
		outerSession = ""
	}
	if outerSession != "" {
		if !rt.bridge.takeLiteTurn(r.Context(), owner, outerSession) {
			msg := "Claude interactive conversation is still answering another request; retry shortly"
			out.Header().Set("Retry-After", "5")
			return out.Fail(http.StatusServiceUnavailable, msg)
		}
		defer rt.bridge.releaseInteractiveOuterTurn(owner, outerSession)
	}
	var search Tool
	if req.WebSearch && !searching(r.Context()) && canSearch() {
		search = searchTool(req.Tools)
		req.Tools = append(slices.Clone(req.Tools), search)
		req.SearchTool = search.Name
	}
	run, events, err := rt.bridge.startLite(r.Context(), req, model, owner, outerSession, configDir)
	if err != nil {
		return out.Fail(startStatus(err), "Claude Code: "+err.Error())
	}
	defer run.releaseHTTP()
	run.mu.Lock()
	run.httpConsumers++
	keepImage := len(run.images) > 0
	run.tools = map[string]bool{}
	for _, t := range req.Tools {
		run.tools[t.Name] = true
	}
	if keepImage {
		run.tools[interactiveImageTool] = true
	}
	if search.Name != "" && config.SearchHandler != nil {
		run.search, run.searchName = config.SearchHandler, search.Name
	}
	run.mu.Unlock()
	run.heard(req.Messages)
	began := time.Now()
	var letting atomic.Bool
	gone := context.AfterFunc(r.Context(), func() {
		letting.Store(true)
		log.Printf("Claude Code run on %s stopped: the client went away %s into the reply", model, time.Since(began).Round(time.Millisecond))
		// A stateless run belongs to no conversation that could take it up
		// again: it is stopped, not kept waiting for its client.
		if outerSession == "" {
			run.abort()
			return
		}
		run.letGoFor(events)
	})
	defer gone()
	abort := func() {
		if !letting.Load() {
			run.abort()
		}
	}
	return out.Relay(req, events, abort, func(said, stop string, ok bool) {
		if !gone() {
			return
		}
		run.ended(req, said, stop, ok)
	})
}

type liteCompletedTurn struct {
	key   string
	reply string
}

// liteReplayKey requires the complete semantic request, independently of
// the loose history matcher. Client metadata and response streaming are ignored.
func liteReplayKey(owner string, req *Request) string {
	if req.replayKey != "" {
		return req.replayKey
	}
	copied := *req
	copied.Metadata, copied.ClientMetadata, copied.Stream = nil, nil, false
	var normalize func([]Part) []Part
	normalize = func(parts []Part) []Part {
		parts = slices.Clone(parts)
		for i := range parts {
			if parts[i].Kind == ToolCall {
				parts[i].Args = argsOf(parts[i])
				if bytes.Equal(bytes.TrimSpace(parts[i].Args), []byte("null")) {
					parts[i].Args = json.RawMessage("{}")
				}
			}
			parts[i].Images = normalize(parts[i].Images)
		}
		return parts
	}
	copied.Messages = slices.Clone(req.Messages)
	for i := range copied.Messages {
		copied.Messages[i].Parts = normalize(copied.Messages[i].Parts)
	}
	raw, err := json.Marshal(copied)
	if err != nil {
		return ""
	}
	// UseNumber preserves integers beyond float64 precision while sorting
	// JSON object keys, including nested argument and schema objects.
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var value any
	if err := dec.Decode(&value); err != nil {
		return ""
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	h := sha256.New()
	fmt.Fprintf(h, "%s\x00", owner)
	h.Write(canonical)
	return hex.EncodeToString(h.Sum(nil))
}

func (b *subscriptionBridge) noteLiteCompletedToolTurn(owner, outerSession string, req *Request, reply ...string) {
	if b == nil || outerSession == "" || req == nil || !interactiveToolContinuation(req) {
		return
	}
	key := interactiveSessionStateKey(owner, outerSession)
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.liteCompleted == nil || len(b.liteCompleted) >= 1024 {
		b.liteCompleted = make(map[string]liteCompletedTurn)
	}
	var answered string
	if len(reply) != 0 {
		answered = reply[0]
	}
	b.liteCompleted[key] = liteCompletedTurn{key: liteReplayKey(owner, req), reply: answered}
}

func (b *subscriptionBridge) liteToolTurnWasDelivered(owner, outerSession string, req *Request) bool {
	_, ok := b.liteCompletedToolTurn(owner, outerSession, req)
	return ok
}

func (b *subscriptionBridge) liteCompletedToolTurn(owner, outerSession string, req *Request) (liteCompletedTurn, bool) {
	if b == nil || outerSession == "" || req == nil || !interactiveToolContinuation(req) {
		return liteCompletedTurn{}, false
	}
	key := interactiveSessionStateKey(owner, outerSession)
	b.mu.Lock()
	prior := b.liteCompleted[key]
	b.mu.Unlock()
	return prior, prior.key != "" && prior.key == liteReplayKey(owner, req)
}

func liteVisibleToolHandoff(req *Request) error {
	if req == nil {
		return errors.New("missing request")
	}
	clean, _, err := interactiveRequestWithoutSystem(req)
	if err != nil {
		return err
	}
	end := len(clean.Messages)
	userStart := end
	for userStart > 0 && clean.Messages[userStart-1].Role == "user" {
		userStart--
	}
	if userStart == end {
		return nil
	}
	assistantStart := userStart
	for assistantStart > 0 && clean.Messages[assistantStart-1].Role == "assistant" {
		assistantStart--
	}
	if assistantStart == userStart {
		return nil // compacted result with no visible assistant: cannot verify
	}
	visibleCall, returned := false, false
	for _, m := range clean.Messages[assistantStart:userStart] {
		for _, p := range m.Parts {
			visibleCall = visibleCall || p.Kind == ToolCall
		}
	}
	for _, m := range clean.Messages[userStart:] {
		for _, p := range m.Parts {
			returned = returned || p.Kind == ToolResult
		}
	}
	if visibleCall && returned {
		return validateInteractiveToolContinuation(clean)
	}
	return nil
}

// liteContinues reports whether req exactly continues run: the same owner,
// model, conversation, tools and settings it was started with. Only then is
// it handed to the run; anything else is a new turn.
func liteContinues(owner, model string, req *Request, run *subscriptionRun, results []Part, how string) bool {
	if !interactiveToolContinuation(req) || !run.interactive || run.owner != owner || run.model != model {
		return false
	}
	if !run.offers(req.Tools) && !run.exactInteractiveToolGrowth(owner, req, how) {
		return false
	}
	if validateInteractiveMatchedContinuation(req, run, results, how) != nil || !run.follows(req.Messages) {
		return false
	}
	run.mu.Lock()
	sameContext := run.interactiveContextKey == turnKey(owner, req, nil) && run.effort == req.Effort
	sameSafeguards := interactiveSameSafeguards(run.safeguards, req.Safeguards) &&
		(len(req.Safeguards) == 0 || run.safeguardBeta == req.SafeguardBeta)
	run.mu.Unlock()
	return sameSafeguards && (sameContext || run.interactiveEquivalentContext(req) || run.exactInteractiveToolGrowth(owner, req, how))
}

// liteCanonical is upstream's helper routing: a request on a conversation id
// is that conversation's own turn unless it is a different thread the
// checkpoint does not verify, or a one-message call for another model.
func (b *subscriptionBridge) liteCanonical(owner, outerSession, model string, req *Request) bool {
	shape, _, err := interactiveRequestWithoutSystem(req)
	if err != nil {
		return true
	}
	if active := b.outerThread(owner, outerSession); active != "" && active != firstWords(req) {
		_, _, status := b.restoreInteractiveSession(owner, outerSession, req)
		if status != interactiveRestoreExact && status != interactiveRestoreReplyAnchor && status != interactiveRestoreCompanion {
			return false
		}
	}
	if len(shape.Messages) == 1 {
		saved, savedOK, _ := b.loadInteractiveSession(owner, outerSession)
		if savedOK && ((saved.Model != "" && saved.Model != model) || (saved.Model == "" && claudeTier(model) == "haiku")) {
			return false
		}
	}
	return true
}

// takeLiteTurn makes this request the conversation's only turn. A turn still
// unwinding is waited out. A run parked on tool calls is stopped: this is
// the conversation's own next request (helpers were routed stateless before
// here), so those calls are either answered in its history — and served
// from it — or never reached the client. It reports false when another run
// of the conversation is still generating after liteTurnWait.
func (b *subscriptionBridge) takeLiteTurn(ctx context.Context, owner, outerSession string) bool {
	deadline := time.Now().Add(liteTurnWait)
	for {
		if b.claimInteractiveOuterTurn(owner, outerSession) {
			old := b.interactiveOuterOwner(owner, outerSession)
			if old == nil {
				return true
			}
			old.mu.Lock()
			idle := old.terminalDelivered || !old.parkedAt.IsZero()
			old.mu.Unlock()
			if idle {
				log.Printf("Claude interactive stopped the conversation's idle or parked run for its next request")
				old.abort()
			}
			b.releaseInteractiveOuterTurn(owner, outerSession)
		}
		if time.Now().After(deadline) {
			return false
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(100 * time.Millisecond):
		}
	}
}

// startLite starts the native run for a turn: the checkpointed session
// resumed when the request exactly follows it, a first turn as a first turn,
// and otherwise a session rebuilt from the request's history. Stateless
// requests (outerSession "") never read or write a checkpoint.
func (b *subscriptionBridge) startLite(ctx context.Context, req *Request, model, owner, outerSession, configDir string) (*subscriptionRun, <-chan Event, error) {
	shape, _, err := interactiveRequestWithoutSystem(req)
	if err != nil {
		return nil, nil, err
	}
	if outerSession != "" {
		if id, prompt, images, ok := b.liteParkedBranch(owner, outerSession, model, configDir, req); ok {
			log.Printf("Claude interactive resuming the conversation's own session with its tool results")
			b.anchorOuterThread(owner, outerSession, req)
			copied := *req
			copied.resumedCopy = true
			return b.startInteractiveWithPromptImages(ctx, &copied, model, configDir, owner, outerSession, prompt, id, images, false)
		}
		saved, savedOK, err := b.loadInteractiveSession(owner, outerSession)
		if err != nil {
			return nil, nil, err
		}
		if savedOK && !saved.Dirty && (saved.Model == "" || saved.Model == model) {
			if saved, since, status := b.restoreInteractiveSession(owner, outerSession, req); status == interactiveRestoreExact || status == interactiveRestoreReplyAnchor || status == interactiveRestoreCompanion {
				b.anchorOuterThread(owner, outerSession, req)
				contextChanged := saved.ContextKey != "" && saved.ContextKey != turnKey(owner, req, nil)
				return b.startInteractiveResumeTurn(ctx, req, model, configDir, owner, outerSession, saved, since, contextChanged)
			} else {
				log.Printf("Claude interactive rebuilding native session from the request: checkpoint %s", status)
			}
		} else if savedOK {
			log.Printf("Claude interactive rebuilding native session from the request: checkpoint dirty=%t model_match=%t", saved.Dirty, saved.Model == "" || saved.Model == model)
		}
		b.anchorOuterThread(owner, outerSession, req)
	}
	if len(shape.Messages) == 1 && validateInteractiveFirstTurn(shape) == nil {
		return b.startInteractiveOneTurn(ctx, req, model, configDir, owner, outerSession)
	}
	seeded, prompt, images, err := liteSeed(req)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("Claude interactive starting a native session rebuilt from the request: history=%d stateless=%t", len(seeded.seedHistory), outerSession == "")
	return b.startInteractiveWithPromptImages(ctx, seeded, model, configDir, owner, outerSession, prompt, "", images, false)
}

// anchorOuterThread records req's thread as the conversation's own, as
// upstream does once a canonical start is verified.
func (b *subscriptionBridge) anchorOuterThread(owner, outerSession string, req *Request) {
	thread := firstWords(req)
	if thread == "" {
		return
	}
	key := interactiveSessionStateKey(owner, outerSession)
	b.mu.Lock()
	b.outerThreads[key] = thread
	b.mu.Unlock()
}

// liteSeed splits req into the history a rebuilt native session resumes from
// and the turn it is given: the user's messages after the last assistant
// reply. Tool results there stay in the history, where the native session
// sees them answer its calls. Unanswered calls are dropped; orphaned
// results retain their content as labelled historical output.
func liteSeed(req *Request) (*Request, string, map[string]Part, error) {
	k, results, turn := liteTail(req.Messages)
	history := pairedToolTraffic(append(slices.Clone(req.Messages[:k]), results...))
	if n := len(history); n > 0 && history[n-1].Role != "assistant" {
		history = append(history, Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: liteResultsReceived}}})
	}
	prompt, images, err := liteTurnPrompt(req, turn)
	if err != nil {
		return nil, "", nil, err
	}
	seeded := *req
	seeded.seedHistory = history
	return &seeded, prompt, images, nil
}

// liteTail splits the messages after the last assistant reply (from k) into
// their tool results, which belong to the history, and the rest: the turn.
func liteTail(msgs []Message) (k int, results, turn []Message) {
	k = len(msgs)
	for k > 0 && msgs[k-1].Role != "assistant" {
		k--
	}
	for _, m := range msgs[k:] {
		var answers, rest []Part
		for _, p := range m.Parts {
			if p.Kind == ToolResult {
				answers = append(answers, p)
			} else {
				rest = append(rest, p)
			}
		}
		if len(answers) > 0 {
			results = append(results, Message{Role: "user", Parts: answers})
		}
		if len(rest) > 0 {
			turn = append(turn, Message{Role: m.Role, Parts: rest})
		}
	}
	return k, results, turn
}

// liteTurnPrompt is the prompt a resumed session is given for turn: what the
// user said, or liteContinueText when the turn is only tool results.
func liteTurnPrompt(req *Request, turn []Message) (string, map[string]Part, error) {
	if !hasUserText(turn) {
		turn = append(turn, Message{Role: "user", Parts: []Part{{Kind: Text, Text: liteContinueText}}})
	}
	turnReq := *req
	turnReq.Messages = turn
	return renderInteractiveFirstPromptWithImages(&turnReq)
}

// noteLiteParked keeps r, a conversation's run now waiting on its tool
// calls, for liteParkedBranch.
func (b *subscriptionBridge) noteLiteParked(r *subscriptionRun) {
	if b == nil || !r.interactive || r.outerSession == "" {
		return
	}
	key := interactiveSessionStateKey(r.owner, r.outerSession)
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.liteParked == nil {
		b.liteParked = make(map[string]*subscriptionRun)
	}
	for k, old := range b.liteParked {
		old.mu.Lock()
		stale := time.Since(old.parkedAt) > claudeCacheTTL
		old.mu.Unlock()
		if stale {
			delete(b.liteParked, k)
		}
	}
	b.liteParked[key] = r
}

// liteParkedBranch serves the results of a conversation's tool calls whose
// run was let go while it waited on them (interactiveParkLongest). Its own
// session is copied up to those calls, the results are appended, and the
// copy resumed: Anthropic is asked what the run would have asked, a prefix
// it has cached, where a session rebuilt from the request (liteSeed) would
// write the whole conversation again. It returns the copy's id and the
// turn's prompt, or false when req is not exactly those results.
func (b *subscriptionBridge) liteParkedBranch(owner, outerSession, model, configDir string, req *Request) (string, string, map[string]Part, bool) {
	key := interactiveSessionStateKey(owner, outerSession)
	b.mu.Lock()
	run := b.liteParked[key]
	delete(b.liteParked, key)
	b.mu.Unlock()
	if run == nil {
		return "", "", nil, false
	}
	run.mu.Lock()
	closed, sid, runModel := run.closed, run.sessionID, run.model
	sameContext := run.interactiveContextKey == turnKey(owner, req, nil)
	calls := make([]string, 0, len(run.shown))
	for _, c := range run.shown {
		calls = append(calls, c.id)
	}
	run.mu.Unlock()
	if !closed || sid == "" || runModel != model || len(calls) == 0 ||
		!(sameContext || run.interactiveEquivalentContext(req)) || !run.follows(req.Messages) {
		return "", "", nil, false
	}
	_, results, turn := liteTail(req.Messages)
	answered := map[string]bool{}
	for _, m := range results {
		for _, p := range m.Parts {
			answered[p.CallID] = true
		}
	}
	if len(answered) != len(calls) {
		return "", "", nil, false
	}
	for _, id := range calls {
		if !answered[id] {
			return "", "", nil, false
		}
	}
	work, err := interactiveSessionWorkDir(owner, outerSession)
	if err != nil {
		return "", "", nil, false
	}
	prompt, images, err := liteTurnPrompt(req, turn)
	if err != nil {
		return "", "", nil, false
	}
	id, err := liteBranchTranscript(work, configDir, model, sid, calls, results)
	if err != nil {
		log.Printf("Claude interactive could not resume the conversation's own session with its tool results: %v", err)
		return "", "", nil, false
	}
	return id, prompt, images, true
}

// liteBranchTranscript copies session sid's rows up to the last of its
// assistant rows calling calls, byte for byte but for the session id, then
// appends results and liteResultsReceived. What the session wrote after the
// calls (an interruption, as its run was let go) is left out.
func liteBranchTranscript(work, configDir, model, sid string, calls []string, results []Message) (string, error) {
	var dir string
	var data []byte
	err := os.ErrNotExist
	for _, d := range claudeSessions(configDir, work) {
		if data, err = os.ReadFile(filepath.Join(d, sid+".jsonl")); err == nil {
			dir = d
			break
		}
	}
	if err != nil {
		return "", err
	}
	want := map[string]bool{}
	for _, c := range calls {
		want[c] = true
	}
	lines := bytes.SplitAfter(data, []byte("\n"))
	cut, parent, found := -1, "", map[string]bool{}
	for i, line := range lines {
		var row struct {
			Type    string
			UUID    string
			Message struct {
				Content []struct{ Type, ID string }
			}
		}
		if json.Unmarshal(line, &row) != nil || row.Type != "assistant" {
			continue
		}
		for _, c := range row.Message.Content {
			if c.Type == "tool_use" && want[c.ID] {
				found[c.ID] = true
				cut, parent = i, row.UUID
			}
		}
	}
	if cut < 0 || parent == "" || len(found) != len(want) {
		return "", fmt.Errorf("session %s does not hold the calls", sid)
	}
	id := randomToken()[:32]
	id = id[:8] + "-" + id[8:12] + "-4" + id[13:16] + "-a" + id[17:20] + "-" + id[20:]
	path := filepath.Join(dir, id+".jsonl")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", err
	}
	defer f.Close()
	from, to := []byte(`"sessionId":"`+sid+`"`), []byte(`"sessionId":"`+id+`"`)
	for _, line := range lines[:cut+1] {
		line = bytes.ReplaceAll(line, from, to)
		if !bytes.HasSuffix(line, []byte("\n")) {
			line = append(line, '\n')
		}
		if _, err := f.Write(line); err != nil {
			_ = os.Remove(path)
			return "", err
		}
	}
	if err := appendClaudeTranscript(f, id, parent, work, model, nil, results, "", liteResultsReceived); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	return id, nil
}

// Stateless runs borrow fixed cwds, the lowest free first, so one after
// another runs where the last did and is asked the same system prompt.
var liteStateless struct {
	sync.Mutex
	busy []bool
}

func liteStatelessWorkDir() (string, func(), error) {
	liteStateless.Lock()
	i := slices.Index(liteStateless.busy, false)
	if i < 0 {
		i = len(liteStateless.busy)
		liteStateless.busy = append(liteStateless.busy, true)
	} else {
		liteStateless.busy[i] = true
	}
	liteStateless.Unlock()
	var once sync.Once
	release := func() {
		once.Do(func() {
			liteStateless.Lock()
			liteStateless.busy[i] = false
			liteStateless.Unlock()
		})
	}
	dir, err := ownedWorkDir(filepath.Join(tempRoot(), fmt.Sprintf("magpie-claude-stateless-%d-%d", os.Getuid(), i)))
	if err != nil {
		release()
		return "", func() {}, err
	}
	return dir, release, nil
}

func hasUserText(msgs []Message) bool {
	for _, m := range msgs {
		if m.Role != "user" {
			continue
		}
		for _, p := range m.Parts {
			if (p.Kind == Text && p.Text != "") || p.Kind == Image || p.Kind == File {
				return true
			}
		}
	}
	return false
}

// pairedToolTraffic pairs adjacent assistant/user role runs. Orphan output
// stays historical text and media, never a native tool_result without a call.
func pairedToolTraffic(msgs []Message) []Message {
	runs := make([]Message, 0, len(msgs))
	for _, m := range msgs {
		if n := len(runs); n > 0 && runs[n-1].Role == m.Role {
			runs[n-1].Parts = append(runs[n-1].Parts, m.Parts...)
		} else {
			runs = append(runs, Message{Role: m.Role, Parts: slices.Clone(m.Parts)})
		}
	}
	out := make([]Message, 0, len(runs))
	for i, m := range runs {
		var keep []Part
		for _, p := range m.Parts {
			switch {
			case p.Kind == ToolCall && m.Role == "assistant":
				if i+1 < len(runs) && answers(runs[i+1], p.ID) {
					p.Args = argsOf(p)
					if bytes.Equal(bytes.TrimSpace(p.Args), []byte("null")) {
						p.Args = json.RawMessage("{}")
					}
					keep = append(keep, p)
				}
			case p.Kind == ToolResult:
				if i > 0 && calls(runs[i-1], p.CallID) {
					keep = append(keep, p)
				} else {
					var text strings.Builder
					result := p
					result.Images = nil
					renderParts(nil, &text, []Part{result})
					keep = append(keep, Part{Kind: Text, Text: text.String()})
					keep = append(keep, p.Images...)
					keep = append(keep, Part{Kind: Text, Text: fmt.Sprintf("\n[end of tool result id=%s]\n", p.CallID)})
				}
			default:
				keep = append(keep, p)
			}
		}
		if len(keep) > 0 {
			out = append(out, Message{Role: m.Role, Parts: keep})
		}
	}
	return out
}

func answers(m Message, id string) bool {
	if m.Role != "user" {
		return false
	}
	for _, p := range m.Parts {
		if p.Kind == ToolResult && p.CallID == id {
			return true
		}
	}
	return false
}

func calls(m Message, id string) bool {
	if m.Role != "assistant" {
		return false
	}
	for _, p := range m.Parts {
		if p.Kind == ToolCall && p.ID == id {
			return true
		}
	}
	return false
}

var limitWords = regexp.MustCompile(`(?i)hit your .*limit|(session|usage|weekly|5-hour) limit|limit.{0,24}resets|rate.?limit|too many requests`)

// liteStatus is the status a failed turn is answered with. A native error
// that only carries the generic 502 but says the account hit its usage or
// rate limit is a 429: clients then wait instead of retrying at once.
func liteStatus(status int, msg string) int {
	if (status == 0 || status == http.StatusBadGateway) && limitWords.MatchString(msg) {
		return http.StatusTooManyRequests
	}
	return status
}
