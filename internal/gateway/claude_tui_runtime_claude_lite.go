package gateway

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
// Identified helper calls run beside the conversation's native session
// and checkpoint. A changed history alone is not a helper: it replaces
// the checkpoint when it cannot resume it.

// liteTurnWait bounds how long a conversation's turn waits for its previous
// turn to let go of it (a reply still unwinding after its client left, a
// parked run being stopped). A conversation still generating after it is
// answered 503 with Retry-After: a second native run of one conversation,
// able to call tools beside the first, is never started.
var tuiRuntimeLiteTurnWait = 60 * time.Second

// liteContinueText is the turn a rebuilt session is given when the request
// ends in tool results with nothing else to say.
const tuiRuntimeLiteContinueText = "Continue from the tool results above."

// liteResultsReceived closes a rebuilt history that ends in tool results.
// Resuming a session whose last message is the user's, Claude Code answers
// that message itself ("No response requested.") into the new reply.
const tuiRuntimeLiteResultsReceived = "Tool results received."

// statelessSession is the signed bridge session of a request that has no
// conversation of its own (the plugin found no stable conversation id).
const tuiRuntimeStatelessSession = "stateless"

// autoModeRefusal answers Claude Code's auto mode classifier on a TUI
// account. Auto mode asks the API to review each action (safeguards) and
// reads the verdict from safeguard_results; the native Claude Code a TUI
// run is keeps that verdict to itself (its transcript has none), so the
// client finds none and classifies every action locally instead: its whole
// transcript in one request, which the PTY can only type as one prompt
// block, written to the cache again each time (au-exit 10-27: 31 of them,
// ~300k tokens each, 9.4M in half an hour). Refused, Claude Code asks the
// user about the action, or stops the turn after several in a row: never
// an action let through unreviewed.
const tuiRuntimeAutoModeRefusal = "Claude Code's auto mode can't review actions on an account in TUI mode: its classifier is refused here. Switch the account to standard (magpie claude-code mode <account> standard) or leave auto mode."

// autoModeClassifier reports whether req is Claude Code's auto mode
// classifier (its "<transcript>" request): no tools, only the user's
// messages, the first beginning the transcript. A segmented transcript
// gives each block a user message of its own.
func tuiRuntimeAutoModeClassifier(req *tuiRuntimeRequest) bool {
	if len(req.Tools) > 0 || len(req.Messages) == 0 {
		return false
	}
	for _, m := range req.Messages {
		if m.Role != "user" {
			return false
		}
	}
	first := req.Messages[0].Parts
	return len(first) > 0 && first[0].Kind == Text && strings.HasPrefix(first[0].Text, "<transcript>\n")
}

// configDir is the Claude config folder the account runs in ("" for
// Claude Code's own, as the plugin's runtime ran).
func (rt *tuiRuntime) serveLite(out tuiRuntimeResponder, r *http.Request, model, owner, outerSession, configDir string, body []byte) (int, string) {
	req, err := tuiRuntimeParseRequest(body)
	if err != nil {
		return out.Fail(http.StatusBadRequest, err.Error())
	}
	req.Model = model
	// A subagent may share its parent's session header, but must not
	// replace the parent's checkpoint when its history differs.
	if subOf(r.Header) != "" {
		outerSession = ""
	}
	// Auto mode's own classifier: refused before any native run starts.
	if tuiRuntimeAutoModeClassifier(req) {
		log.Printf("Claude interactive refused auto mode's local classifier request: a TUI account can't review auto mode's actions")
		out.Header().Set("x-should-retry", "false")
		return out.Fail(http.StatusBadRequest, tuiRuntimeAutoModeRefusal)
	}
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
	if err := tuiRuntimeLiteVisibleToolHandoff(req); err != nil {
		msg := "Claude interactive tool result does not match its visible tool call; refusing"
		return out.Fail(http.StatusConflict, msg)
	}
	// The same tool results again, after their reply was delivered: a client
	// retrying. Only the very same request gets that reply again; the same
	// history asked under another system prompt, tools or settings is run.
	req.replayKey = tuiRuntimeLiteReplayKey(owner, req)
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
		start := func(ctx context.Context, req *tuiRuntimeRequest) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
			return rt.bridge.startInteractiveRequest(ctx, req, model, configDir, owner, outerSession)
		}
		return rt.serveMessagesInner(out, r, "Claude Code", model, owner, outerSession, body, start)
	}
	// Desktop's own compaction and suggestion calls already run isolated.
	if tuiRuntimeClaudeDesktopCompaction(req) || tuiRuntimeClaudeDesktopSuggestion(req) {
		return upstream()
	}
	// Compare the same effective tools that the native run was started with.
	// Adding the private search tool only after matching makes every search-
	// enabled continuation appear to have changed its tool definitions.
	var search Tool
	if req.WebSearch && !searching(r.Context()) && tuiRuntimeCanSearch() {
		search = searchTool(req.Tools)
		req.Tools = append(slices.Clone(req.Tools), search)
		req.tuiSearchTool = search.Name
	}
	// The exact continuation of a live run — its tool results, perhaps with
	// what the user said meanwhile — goes on in that native session.
	if run, results, how := rt.bridge.match(req); run != nil && tuiRuntimeLiteContinues(owner, model, req, run, results, how) {
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
		rt.bridge.letGoAbandonedRuns(owner, req)
		release, ok := rt.bridge.enterLiteTurn(r.Context(), req.replayKey)
		if !ok {
			msg := "Claude interactive is still answering this same request; retry shortly"
			out.Header().Set("Retry-After", "5")
			return out.Fail(http.StatusServiceUnavailable, msg)
		}
		defer release()
	}
	run, events, err := rt.bridge.startLite(r.Context(), req, model, owner, outerSession, configDir)
	if err != nil {
		return out.Fail(tuiRuntimeStartStatus(err), "Claude Code: "+err.Error())
	}
	defer run.releaseHTTP()
	run.mu.Lock()
	run.startKey = req.replayKey
	run.httpConsumers++
	keepImage := len(run.images) > 0
	run.tools = map[string]bool{}
	for _, t := range req.Tools {
		run.tools[t.Name] = true
	}
	if keepImage {
		run.tools[tuiRuntimeInteractiveImageTool] = true
	}
	if search.Name != "" && tuiRuntimeSettings.SearchHandler != nil {
		run.search, run.searchName = tuiRuntimeSettings.SearchHandler, search.Name
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

type tuiRuntimeLiteCompletedTurn struct {
	key   string
	reply string
}

// liteReplayKey requires the complete semantic request, independently of
// the loose history matcher. Client metadata and response streaming are ignored.
func tuiRuntimeLiteReplayKey(owner string, req *tuiRuntimeRequest) string {
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

func (b *tuiRuntimeSubscriptionBridge) noteLiteCompletedToolTurn(owner, outerSession string, req *tuiRuntimeRequest, reply ...string) {
	if b == nil || outerSession == "" || req == nil || !tuiRuntimeInteractiveToolContinuation(req) {
		return
	}
	key := tuiRuntimeInteractiveSessionStateKey(owner, outerSession)
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.liteCompleted == nil || len(b.liteCompleted) >= 1024 {
		b.liteCompleted = make(map[string]tuiRuntimeLiteCompletedTurn)
	}
	var answered string
	if len(reply) != 0 {
		answered = reply[0]
	}
	b.liteCompleted[key] = tuiRuntimeLiteCompletedTurn{key: tuiRuntimeLiteReplayKey(owner, req), reply: answered}
}

func (b *tuiRuntimeSubscriptionBridge) liteToolTurnWasDelivered(owner, outerSession string, req *tuiRuntimeRequest) bool {
	_, ok := b.liteCompletedToolTurn(owner, outerSession, req)
	return ok
}

func (b *tuiRuntimeSubscriptionBridge) liteCompletedToolTurn(owner, outerSession string, req *tuiRuntimeRequest) (tuiRuntimeLiteCompletedTurn, bool) {
	if b == nil || outerSession == "" || req == nil || !tuiRuntimeInteractiveToolContinuation(req) {
		return tuiRuntimeLiteCompletedTurn{}, false
	}
	key := tuiRuntimeInteractiveSessionStateKey(owner, outerSession)
	b.mu.Lock()
	prior := b.liteCompleted[key]
	b.mu.Unlock()
	return prior, prior.key != "" && prior.key == tuiRuntimeLiteReplayKey(owner, req)
}

func tuiRuntimeLiteVisibleToolHandoff(req *tuiRuntimeRequest) error {
	if req == nil {
		return errors.New("missing request")
	}
	clean, _, err := tuiRuntimeInteractiveRequestWithoutSystem(req)
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
		return tuiRuntimeValidateInteractiveToolContinuation(clean)
	}
	return nil
}

// liteContinues reports whether req exactly continues run: the same owner,
// model, conversation, tools and settings it was started with. Only then is
// it handed to the run; anything else is a new turn.
func tuiRuntimeLiteContinues(owner, model string, req *tuiRuntimeRequest, run *tuiRuntimeSubscriptionRun, results []Part, how string) bool {
	if !tuiRuntimeInteractiveToolContinuation(req) || !run.interactive || run.owner != owner || run.model != model {
		return false
	}
	if !run.offers(req.Tools) && !run.exactInteractiveToolGrowth(owner, req, how) {
		return false
	}
	if tuiRuntimeValidateInteractiveMatchedContinuation(req, run, results, how) != nil || !run.follows(req.Messages) {
		return false
	}
	run.mu.Lock()
	sameSafeguards := tuiRuntimeInteractiveSameSafeguards(run.safeguards, req.Safeguards) &&
		(len(req.Safeguards) == 0 || run.safeguardBeta == req.SafeguardBeta)
	run.mu.Unlock()
	return sameSafeguards && (run.interactiveEquivalentContext(req) || run.exactInteractiveToolGrowth(owner, req, how))
}

// liteCanonical keeps upstream's one-message, other-model helper routing.
// A changed first message is not evidence of a helper: client compaction
// or edits can replace the history while keeping the same conversation id.
func (b *tuiRuntimeSubscriptionBridge) liteCanonical(owner, outerSession, model string, req *tuiRuntimeRequest) bool {
	shape, _, err := tuiRuntimeInteractiveRequestWithoutSystem(req)
	if err != nil {
		return true
	}
	if len(shape.Messages) == 1 {
		saved, savedOK, _ := b.loadInteractiveSession(owner, outerSession)
		if savedOK && ((saved.Model != "" && saved.Model != model) || (saved.Model == "" && claudeTier(model) == "haiku")) {
			return false
		}
	}
	return true
}

// enterLiteTurn admits a lite request. Different requests of one
// conversation run side by side, each in its own native session; only the
// very same request (a client's retry) waits for the one in flight, so it
// is never answered by two native runs at once. It reports false when that
// one is still running after liteTurnWait.
func (b *tuiRuntimeSubscriptionBridge) enterLiteTurn(ctx context.Context, key string) (func(), bool) {
	wait := time.NewTimer(tuiRuntimeLiteTurnWait)
	defer wait.Stop()
	for {
		b.mu.Lock()
		if b.liteInFlight == nil {
			b.liteInFlight = map[string]chan struct{}{}
		}
		ch := b.liteInFlight[key]
		if ch == nil {
			ch = make(chan struct{})
			b.liteInFlight[key] = ch
			b.mu.Unlock()
			return func() {
				b.mu.Lock()
				delete(b.liteInFlight, key)
				b.mu.Unlock()
				close(ch)
			}, true
		}
		b.mu.Unlock()
		select {
		case <-ch:
		case <-ctx.Done():
			return nil, false
		case <-wait.C:
			return nil, false
		}
	}
}

// letGoAbandonedRuns stops the runs req shows its conversation went on
// without: one started for this very request whose client is gone (a
// retry), and one parked on tool calls that req's history holds but does
// not answer. A run of another branch — its calls not in req's history —
// is left alone: that branch may still come back with its results.
func (b *tuiRuntimeSubscriptionBridge) letGoAbandonedRuns(owner string, req *tuiRuntimeRequest) {
	called := map[string]bool{}
	for _, m := range req.Messages {
		for _, p := range m.Parts {
			if m.Role == "assistant" && p.Kind == ToolCall {
				called[p.ID] = true
			}
		}
	}
	_, results, _ := tuiRuntimeLiteTail(req.Messages)
	answered := map[string]bool{}
	for _, m := range results {
		for _, p := range m.Parts {
			answered[p.CallID] = true
		}
	}
	var gone []*tuiRuntimeSubscriptionRun
	for _, run := range b.allRuns() {
		run.mu.Lock()
		if run.interactive && !run.closed && run.owner == owner && run.httpConsumers == 0 {
			abandoned := run.startKey != "" && run.startKey == req.replayKey
			if !abandoned && !run.parkedAt.IsZero() {
				inHistory, all := false, true
				for _, c := range run.shown {
					inHistory = inHistory || called[c.id]
					all = all && answered[c.id]
				}
				abandoned = inHistory && !all
			}
			if abandoned {
				gone = append(gone, run)
			}
		}
		run.mu.Unlock()
	}
	for _, run := range gone {
		log.Printf("Claude interactive stopped a run its conversation went on without")
		run.abort()
	}
}

// startLite starts the native run for a turn, keeping as much of the
// prefix Anthropic cached as it can: the session of the deepest reply the
// request goes on from, in place or forked there (startLiteAtPoint); the
// conversation's checkpoint as earlier versions saved it; else a first turn
// as a first turn, and a session rebuilt from the request's history.
// Stateless requests (outerSession "") never read or write either.
func (b *tuiRuntimeSubscriptionBridge) startLite(ctx context.Context, req *tuiRuntimeRequest, model, owner, outerSession, configDir string) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
	if outerSession != "" {
		if id, prompt, images, work, ok := b.liteParkedBranch(owner, outerSession, model, configDir, req); ok {
			log.Printf("Claude interactive resuming the conversation's own session with its tool results")
			b.anchorOuterThread(owner, outerSession, req)
			copied := *req
			copied.resumedCopy, copied.workKey = true, work
			return b.startInteractiveWithPromptImages(ctx, &copied, model, configDir, owner, outerSession, prompt, id, images, false)
		}
		point, pointSince, havePoint := tuiRuntimeInteractivePoint{}, []Message(nil), false
		if tuiRuntimeClaudeDesktopTurnCompanionSuffix(req) == nil {
			point, pointSince, havePoint = tuiRuntimeFindInteractivePoint(owner, outerSession, model, req)
		}
		saved, savedOK, err := b.loadInteractiveSession(owner, outerSession)
		if err != nil {
			return nil, nil, err
		}
		var since []Message
		status := tuiRuntimeInteractiveRestoreStateMissing
		if savedOK && !saved.Dirty && (saved.Model == "" || saved.Model == model) {
			saved, since, status = b.restoreInteractiveSession(owner, outerSession, req)
		}
		legacy := status == tuiRuntimeInteractiveRestoreExact || status == tuiRuntimeInteractiveRestoreReplyAnchor || status == tuiRuntimeInteractiveRestoreCompanion
		// The deeper of the two goes on; a point also when they are level,
		// since it can be forked if its session is busy.
		if havePoint && (!legacy || status != tuiRuntimeInteractiveRestoreCompanion && len(pointSince) <= len(since)) {
			if run, events, err, ok := b.startLiteAtPoint(ctx, req, model, owner, outerSession, configDir, point, pointSince); ok {
				b.anchorOuterThread(owner, outerSession, req)
				return run, events, err
			}
		}
		if legacy {
			if release, ok := b.claimSession(saved.SessionID); ok {
				defer release()
				b.anchorOuterThread(owner, outerSession, req)
				return b.startInteractiveResumeTurn(ctx, req, model, configDir, owner, outerSession, saved, since)
			}
			log.Printf("Claude interactive rebuilding native session from the request: its checkpointed session is busy")
		} else if savedOK && (saved.Dirty || saved.Model != "" && saved.Model != model) {
			log.Printf("Claude interactive rebuilding native session from the request: checkpoint dirty=%t model_match=%t", saved.Dirty, saved.Model == "" || saved.Model == model)
		} else if savedOK {
			log.Printf("Claude interactive rebuilding native session from the request: checkpoint %s", status)
		}
		b.anchorOuterThread(owner, outerSession, req)
	}
	return b.startInteractiveRebuild(ctx, req, model, owner, outerSession, configDir)
}

// startLiteAtPoint goes on from point with since, the messages after it:
// in its session when no run writes it and nothing was said in it after
// the point, else in a copy of the session up to the point. Changed
// instructions rebuild: the old ones are in the session's first prompt.
// It reports false, starting nothing, when the session can't be read.
func (b *tuiRuntimeSubscriptionBridge) startLiteAtPoint(ctx context.Context, req *tuiRuntimeRequest, model, owner, outerSession, configDir string, point tuiRuntimeInteractivePoint, since []Message) (*tuiRuntimeSubscriptionRun, <-chan Event, error, bool) {
	if point.InstructionKey != tuiRuntimeInstructionKey(req) {
		log.Printf("Claude interactive rebuilding native session: external instructions changed")
		run, events, err := b.startInteractiveRebuild(ctx, req, model, owner, outerSession, configDir)
		return run, events, err, true
	}
	work, err := tuiRuntimeInteractiveWorkDirForKey(point.Work)
	if err != nil {
		return nil, nil, err, true
	}
	prompt, images, err := tuiRuntimeRenderInteractiveTurnPromptWithImages(since)
	if err != nil {
		return nil, nil, err, true
	}
	copied := *req
	copied.workKey = point.Work
	if len(req.Tools) == 0 && len(images) == 0 && point.ToolsHash != "" {
		copied.heldTools = tuiRuntimeLoadInteractiveTools(point.ToolsHash)
	}
	prompt = tuiRuntimeRenderInteractiveTurnConstraint(&copied) + prompt
	if release, ok := b.claimSession(point.SessionID); ok {
		defer release()
		if tuiRuntimeTranscriptEndsAt(configDir, work, point.SessionID, point.Row) {
			log.Printf("Claude interactive resuming the native session at the reply the request goes on from")
			copied.pointResume = true
			run, events, err := b.startInteractiveWithPromptImages(ctx, &copied, model, configDir, owner, outerSession, prompt, point.SessionID, images, false)
			return run, events, err, true
		}
	}
	id, err := tuiRuntimeForkTranscript(configDir, work, point.SessionID, point.Row)
	if err != nil {
		log.Printf("Claude interactive could not fork the native session at the reply the request goes on from: %v", err)
		return nil, nil, nil, false
	}
	log.Printf("Claude interactive forking the native session at the reply the request goes on from")
	copied.resumedCopy = true
	run, events, err := b.startInteractiveWithPromptImages(ctx, &copied, model, configDir, owner, outerSession, prompt, id, images, false)
	return run, events, err, true
}

// Rebuild from the caller's current history, never from an old transcript
// containing superseded external instructions.
func (b *tuiRuntimeSubscriptionBridge) startInteractiveRebuild(ctx context.Context, req *tuiRuntimeRequest, model, owner, outerSession, configDir string) (*tuiRuntimeSubscriptionRun, <-chan Event, error) {
	shape, _, err := tuiRuntimeInteractiveRequestWithoutSystem(req)
	if err != nil {
		return nil, nil, err
	}
	if len(shape.Messages) == 1 && tuiRuntimeValidateInteractiveFirstTurn(shape) == nil {
		return b.startInteractiveOneTurn(ctx, req, model, configDir, owner, outerSession)
	}
	seeded, prompt, images, err := tuiRuntimeLiteSeed(req)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("Claude interactive starting a native session rebuilt from the request: history=%d stateless=%t", len(seeded.seedHistory), outerSession == "")
	return b.startInteractiveWithPromptImages(ctx, seeded, model, configDir, owner, outerSession, prompt, "", images, false)
}

// anchorOuterThread records req's thread as the conversation's own, as
// upstream does once a canonical start is verified.
func (b *tuiRuntimeSubscriptionBridge) anchorOuterThread(owner, outerSession string, req *tuiRuntimeRequest) {
	thread := firstWords(&req.Request)
	if thread == "" {
		return
	}
	key := tuiRuntimeInteractiveSessionStateKey(owner, outerSession)
	b.mu.Lock()
	b.outerThreads[key] = thread
	b.mu.Unlock()
}

// liteSeed splits req into the history a rebuilt native session resumes from
// and the turn it is given: the user's messages after the last assistant
// reply. Tool results there stay in the history, where the native session
// sees them answer its calls. Unanswered calls are dropped; orphaned
// results retain their content as labelled historical output.
func tuiRuntimeLiteSeed(req *tuiRuntimeRequest) (*tuiRuntimeRequest, string, map[string]Part, error) {
	k, results, turn := tuiRuntimeLiteTail(req.Messages)
	history := tuiRuntimePairedToolTraffic(append(slices.Clone(req.Messages[:k]), results...))
	if n := len(history); n > 0 && history[n-1].Role != "assistant" {
		history = append(history, Message{Role: "assistant", Parts: []Part{{Kind: Text, Text: tuiRuntimeLiteResultsReceived}}})
	}
	prompt, images, err := tuiRuntimeLiteTurnPrompt(req, turn)
	if err != nil {
		return nil, "", nil, err
	}
	seeded := *req
	seeded.seedHistory = history
	return &seeded, prompt, images, nil
}

// liteTail splits the messages after the last assistant reply (from k) into
// their tool results, which belong to the history, and the rest: the turn.
func tuiRuntimeLiteTail(msgs []Message) (k int, results, turn []Message) {
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
func tuiRuntimeLiteTurnPrompt(req *tuiRuntimeRequest, turn []Message) (string, map[string]Part, error) {
	if !tuiRuntimeHasUserText(turn) {
		turn = append(turn, Message{Role: "user", Parts: []Part{{Kind: Text, Text: tuiRuntimeLiteContinueText}}})
	}
	turnReq := *req
	turnReq.Messages = turn
	return tuiRuntimeRenderInteractiveFirstPromptWithImages(&turnReq)
}

// noteLiteParked keeps r, a run now waiting on its tool calls, under each
// of their ids for liteParkedBranch.
func (b *tuiRuntimeSubscriptionBridge) noteLiteParked(r *tuiRuntimeSubscriptionRun) {
	if b == nil || !r.interactive || r.outerSession == "" {
		return
	}
	r.mu.Lock()
	ids := make([]string, 0, len(r.shown))
	for _, c := range r.shown {
		ids = append(ids, c.id)
	}
	r.mu.Unlock()
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.liteParked == nil {
		b.liteParked = make(map[string]*tuiRuntimeSubscriptionRun)
	}
	for k, old := range b.liteParked {
		old.mu.Lock()
		stale := time.Since(old.parkedAt) > tuiRuntimeClaudeCacheTTL
		old.mu.Unlock()
		if stale {
			delete(b.liteParked, k)
		}
	}
	for _, id := range ids {
		b.liteParked[id] = r
	}
}

// liteParkedBranch serves the results of a run's tool calls that the live
// run did not take: the run was let go while it waited on them
// (interactiveParkLongest), or the results came with other tools or
// settings. The run is stopped if it still waits, its own session is
// copied up to those calls, the results are appended, and the copy
// resumed: Anthropic is asked what the run would have asked, a prefix it
// has cached, where a session rebuilt from the request (liteSeed) would
// write the whole conversation again. It returns the copy's id, the turn's
// prompt and the state key of the session's cwd, or false when req is not
// exactly those results. Other conversations' parked runs are not touched.
func (b *tuiRuntimeSubscriptionBridge) liteParkedBranch(owner, outerSession, model, configDir string, req *tuiRuntimeRequest) (string, string, map[string]Part, string, bool) {
	_, results, turn := tuiRuntimeLiteTail(req.Messages)
	answered := map[string]bool{}
	for _, m := range results {
		for _, p := range m.Parts {
			answered[p.CallID] = true
		}
	}
	var run *tuiRuntimeSubscriptionRun
	b.mu.Lock()
	for id := range answered {
		if r := b.liteParked[id]; r != nil {
			run = r
			break
		}
	}
	b.mu.Unlock()
	if run == nil {
		return "", "", nil, "", false
	}
	run.mu.Lock()
	sid, runModel, work, runOwner := run.sessionID, run.model, run.workKey, run.owner
	calls := make([]string, 0, len(run.shown))
	for _, c := range run.shown {
		calls = append(calls, c.id)
	}
	run.mu.Unlock()
	if sid == "" || runOwner != owner || runModel != model || work == "" || len(calls) == 0 || len(answered) != len(calls) ||
		!run.interactiveEquivalentContext(req) || !run.follows(req.Messages) {
		return "", "", nil, "", false
	}
	for _, id := range calls {
		if !answered[id] {
			return "", "", nil, "", false
		}
	}
	b.mu.Lock()
	for _, id := range calls {
		if b.liteParked[id] == run {
			delete(b.liteParked, id)
		}
	}
	b.mu.Unlock()
	// Whatever still runs in its session stops before it is copied.
	run.abort()
	dir, err := tuiRuntimeInteractiveWorkDirForKey(work)
	if err != nil {
		return "", "", nil, "", false
	}
	if !tuiRuntimeHasUserText(turn) {
		turn = append(turn, Message{Role: "user", Parts: []Part{{Kind: Text, Text: tuiRuntimeLiteContinueText}}})
	}
	prompt, images, err := tuiRuntimeRenderInteractiveTurnPromptWithImages(turn)
	if err != nil {
		return "", "", nil, "", false
	}
	prompt = tuiRuntimeRenderInteractiveTurnConstraint(req) + prompt
	id, err := tuiRuntimeLiteBranchTranscript(dir, configDir, model, sid, calls, results)
	if err != nil {
		log.Printf("Claude interactive could not resume the conversation's own session with its tool results: %v", err)
		return "", "", nil, "", false
	}
	return id, prompt, images, work, true
}

// liteBranchTranscript copies session sid's rows up to the last of its
// assistant rows calling calls, byte for byte but for the session id, then
// appends results and liteResultsReceived. What the session wrote after the
// calls (an interruption, as its run was let go) is left out.
func tuiRuntimeLiteBranchTranscript(work, configDir, model, sid string, calls []string, results []Message) (string, error) {
	data, dir, err := tuiRuntimeReadTranscript(configDir, work, sid)
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
	id, f, err := tuiRuntimeCopyTranscript(dir, sid, lines[:cut+1])
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := tuiRuntimeAppendClaudeTranscript(f, id, parent, work, model, nil, results, "", tuiRuntimeLiteResultsReceived); err != nil {
		_ = os.Remove(filepath.Join(dir, id+".jsonl"))
		return "", err
	}
	return id, nil
}

// Stateless runs borrow fixed cwds, the lowest free first, so one after
// another runs where the last did and is asked the same system prompt.
var tuiRuntimeLiteStateless struct {
	sync.Mutex
	busy []bool
}

func tuiRuntimeLiteStatelessWorkDir() (string, func(), error) {
	tuiRuntimeLiteStateless.Lock()
	i := slices.Index(tuiRuntimeLiteStateless.busy, false)
	if i < 0 {
		i = len(tuiRuntimeLiteStateless.busy)
		tuiRuntimeLiteStateless.busy = append(tuiRuntimeLiteStateless.busy, true)
	} else {
		tuiRuntimeLiteStateless.busy[i] = true
	}
	tuiRuntimeLiteStateless.Unlock()
	var once sync.Once
	release := func() {
		once.Do(func() {
			tuiRuntimeLiteStateless.Lock()
			tuiRuntimeLiteStateless.busy[i] = false
			tuiRuntimeLiteStateless.Unlock()
		})
	}
	dir, err := tuiRuntimeOwnedWorkDir(filepath.Join(tuiRuntimeTempRoot(), fmt.Sprintf("magpie-claude-stateless-%d-%d", os.Getuid(), i)))
	if err != nil {
		release()
		return "", func() {}, err
	}
	return dir, release, nil
}

func tuiRuntimeHasUserText(msgs []Message) bool {
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
func tuiRuntimePairedToolTraffic(msgs []Message) []Message {
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
				if i+1 < len(runs) && tuiRuntimeAnswers(runs[i+1], p.ID) {
					p.Args = argsOf(p)
					if bytes.Equal(bytes.TrimSpace(p.Args), []byte("null")) {
						p.Args = json.RawMessage("{}")
					}
					keep = append(keep, p)
				}
			case p.Kind == ToolResult:
				if i > 0 && tuiRuntimeCalls(runs[i-1], p.CallID) {
					keep = append(keep, p)
				} else {
					var text strings.Builder
					result := p
					result.Images = nil
					tuiRuntimeRenderParts(nil, &text, []Part{result})
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

func tuiRuntimeAnswers(m Message, id string) bool {
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

func tuiRuntimeCalls(m Message, id string) bool {
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

var tuiRuntimeLimitWords = regexp.MustCompile(`(?i)hit your .*limit|(session|usage|weekly|5-hour) limit|limit.{0,24}resets|rate.?limit|too many requests`)

// liteStatus is the status a failed turn is answered with. A native error
// that only carries the generic 502 but says the account hit its usage or
// rate limit is a 429: clients then wait instead of retrying at once.
func tuiRuntimeLiteStatus(status int, msg string) int {
	if (status == 0 || status == http.StatusBadGateway) && tuiRuntimeLimitWords.MatchString(msg) {
		return http.StatusTooManyRequests
	}
	return status
}
