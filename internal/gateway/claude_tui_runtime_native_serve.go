package gateway

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"slices"
	"strings"
	"time"
)

// A run whose wrapper is started with --desplega-serve keeps its native
// Claude Code for more than one turn. Each turn is typed in by a submit
// control line on the wrapper's stdin; the wrapper writes the turn's
// stream-json lines and its result, then bridge_ready once Claude Code's
// input box takes a prompt again (serve-protocol.ts). dispatchServe hands
// each native turn's lines to a reader of its own (readInteractiveOneTurn),
// so a turn's end is not the process's.
//
// A turn that ends with a complete reply leaves the process waiting for
// the conversation's next turn, idle, for tuiRuntimeIdleTTL: its session
// ends at that reply, and the request that goes on from exactly it (its
// history, then the reply as the client was told it, then new user input)
// is typed into the same Claude Code. Nothing is said to Claude Code while
// it waits; the wait is not tied to the upstream cache's.
//
//	starting → running ⇄ waiting on tool results
//	running → (result, bridge_ready, reply kept) → idle → claimed → running
//	idle → (TTL, room, pool bound) → stopping → exited
//
// A turn the client gave up on, a failed one, or one with any doubt about
// where the session ends is never kept: its process is ended, and its
// session's cut points stay as they were.

// tuiRuntimeDefaultIdleTTL is how long a process waits idle for its
// conversation's next turn when the settings say nothing.
var tuiRuntimeDefaultIdleTTL = 60 * time.Second

// tuiRuntimeServeIdleMost bounds the processes one bridge keeps idle; past
// it, the one idle longest is ended. claudeRoom ends idle ones first too.
var tuiRuntimeServeIdleMost = 8

// tuiRuntimeIdleTTL is how long a process with a complete reply waits for
// the next turn; 0 keeps none, and every run ends with its reply.
func tuiRuntimeIdleTTL() time.Duration {
	if tuiRuntimeSettings.IdleTTL == nil {
		return tuiRuntimeDefaultIdleTTL
	}
	return max(tuiRuntimeSettings.IdleTTL(), 0)
}

// tuiRuntimeHotConfig is what a native process was started with that a
// turn typed into it can't change: its account and caller, model,
// instructions and their place, effort, thinking, the MCP tool definitions
// its helper serves (byte for byte), native search and safeguards. A
// request whose own launch would differ in any of them is started anew
// from its prefix instead.
type tuiRuntimeHotConfig struct {
	scope, model, instructions, effort string
	outerSession, configDir            string
	lean, thinkOff, nativeSearch       bool
	// imageTool says the helper serves the private image tool; tools is
	// the digest of the definitions with it when it does
	imageTool, held bool
	tools           [32]byte
	safeguards      [32]byte
	safeguardBeta   string
}

func tuiRuntimeHotConfigOf(req *tuiRuntimeRequest, model, outerSession, configDir string, imageTool bool) tuiRuntimeHotConfig {
	tools, _ := json.Marshal(tuiRuntimeInteractiveTools(req, imageTool))
	return tuiRuntimeHotConfig{
		scope: req.scope, model: model, instructions: tuiRuntimeInstructionKey(req), effort: req.Effort,
		outerSession: outerSession, configDir: configDir,
		lean: req.leanPrompt, thinkOff: req.ThinkOff, nativeSearch: req.WebSearch && req.tuiSearchTool == "",
		imageTool: imageTool, tools: sha256.Sum256(tools),
		safeguards: sha256.Sum256(req.Safeguards), safeguardBeta: req.SafeguardBeta,
	}
}

// beginServeTurn starts the reader of the native turn about to be typed in.
// It is started before the submit, so an MCP callback of the turn that
// comes before any of its lines has its reader waiting for it.
func (r *tuiRuntimeSubscriptionRun) beginServeTurn() {
	rd, w := io.Pipe()
	done := make(chan struct{})
	r.mu.Lock()
	r.turnOut, r.turnDone = w, done
	r.turnsSent++
	r.mu.Unlock()
	go func() {
		r.readInteractiveOneTurn(rd)
		// a reader that ended early (an error, an API failure) reads on, so
		// the dispatcher never blocks on a turn no one reads
		_, _ = io.Copy(io.Discard, rd)
		close(done)
	}()
}

// submitServe types prompt in as the turn beginServeTurn readied. Writing
// can wait on a wrapper reading slowly, so no lock is held but serveMu.
func (r *tuiRuntimeSubscriptionRun) submitServe(prompt string) error {
	r.mu.Lock()
	segment := fmt.Sprintf("turn-%d", r.turnsSent)
	r.mu.Unlock()
	line, err := json.Marshal(map[string]string{"op": "submit", "segment": segment, "input": prompt})
	if err != nil {
		return err
	}
	r.serveMu.Lock()
	defer r.serveMu.Unlock()
	r.mu.Lock()
	in := r.serveIn
	r.mu.Unlock()
	if in == nil {
		return errors.New("the native Claude Code has ended")
	}
	_, err = in.Write(append(line, '\n'))
	return err
}

// dispatchServe reads a serve run's wrapper for its whole life: each native
// turn's lines go to its reader until its result, and the wrapper's own
// lines say when the process takes its next turn. It returns once the
// wrapper's output ends and the last turn's reader is done.
func (r *tuiRuntimeSubscriptionRun) dispatchServe(rd io.Reader) {
	scanner := bufio.NewScanner(rd)
	scanner.Buffer(make([]byte, 64<<10), 64<<20)
	for scanner.Scan() {
		line := scanner.Bytes()
		var head struct {
			Type  string `json:"type"`
			Turns int    `json:"turns"`
		}
		_ = json.Unmarshal(line, &head) // a line that isn't JSON goes to the turn, whose reader refuses it
		switch head.Type {
		case "bridge_ready":
			r.noteServeReady(head.Turns)
			continue
		case "bridge_accepted":
			continue
		}
		r.mu.Lock()
		out, done := r.turnOut, r.turnDone
		r.mu.Unlock()
		if out == nil {
			log.Printf("Claude interactive wrapper wrote a %q line between turns; dropped", head.Type)
			continue
		}
		_, _ = out.Write(append(slices.Clone(line), '\n'))
		if head.Type == "result" {
			_ = out.Close()
			<-done
			r.mu.Lock()
			if r.turnOut == out {
				r.turnOut, r.turnDone = nil, nil
			}
			r.mu.Unlock()
		}
	}
	r.mu.Lock()
	out, done := r.turnOut, r.turnDone
	r.turnOut, r.turnDone = nil, nil
	r.mu.Unlock()
	if out != nil {
		// the turn's reader sees its end without a result, and says so
		_ = out.CloseWithError(scanner.Err())
		<-done
	}
}

// turn is the native turn the run is on, by number: a request's
// cancellation acts on the process only while its turn is the latest.
func (r *tuiRuntimeSubscriptionRun) turn() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.turnsSent
}

// letGoServeTurn is told the client of turn went away. A turn handed to
// a later request, or a process idle for the next turn, is not this
// client's to end. A turn that still has tool calls in the client's hands
// keeps its process for their results (letGoFor); any other is ended, its
// session left where its last kept reply is.
func (r *tuiRuntimeSubscriptionRun) letGoServeTurn(segment <-chan Event, turn int) {
	r.mu.Lock()
	mine := r.turnsSent == turn && !r.hotIdle
	waiting := !r.closed && (len(r.shown) > 0 || len(r.pending) > 0)
	r.mu.Unlock()
	switch {
	case !mine:
		return
	case waiting:
		r.letGoFor(segment)
	default:
		r.abort()
	}
}

// noteTurnReply keeps the reply this turn's client was told, as its next
// request holds it (tuiRuntimeReplyRecorder).
func (r *tuiRuntimeSubscriptionRun) noteTurnReply(reply Message) {
	r.mu.Lock()
	r.turnReply = reply
	r.mu.Unlock()
}

// afterServeTurn decides, once a serve run's turn ended with a complete
// reply, whether its process waits for the conversation's next turn. Only a
// run whose session ends exactly at that reply, on the request's scope, is
// kept; any other process is ended.
func (r *tuiRuntimeSubscriptionRun) afterServeTurn(req *tuiRuntimeRequest) {
	r.mu.Lock()
	reply := r.turnReply
	r.turnReply = Message{}
	keep := r.serve && !r.closed && !r.schema && !r.readOnlyRecovery && r.toolRepair == nil &&
		req.scope != "" && req.scope == r.hotConfig.scope && len(reply.Parts) > 0 && tuiRuntimeIdleTTL() > 0
	r.mu.Unlock()
	if !keep {
		go r.abort()
		return
	}
	msgs := append(req.Messages[:len(req.Messages):len(req.Messages)], reply)
	r.keepForNextTurn(tuiRuntimePrefixOf(req.scope, msgs))
}

// keepForNextTurn makes the run idle at tip once the wrapper is ready too.
func (r *tuiRuntimeSubscriptionRun) keepForNextTurn(tip tuiRuntimePrefixDigest) {
	b := r.bridge
	b.mu.Lock()
	r.mu.Lock()
	r.hotTip = &tip
	gen, idle := r.enterIdleLocked()
	r.mu.Unlock()
	b.mu.Unlock()
	if idle {
		r.armIdle(gen)
	}
}

// noteServeReady is the wrapper saying Claude Code takes a prompt again,
// turns turns done.
func (r *tuiRuntimeSubscriptionRun) noteServeReady(turns int) {
	b := r.bridge
	b.mu.Lock()
	r.mu.Lock()
	r.turnsReady = turns
	gen, idle := r.enterIdleLocked()
	r.mu.Unlock()
	b.mu.Unlock()
	if idle {
		r.armIdle(gen)
	}
}

// enterIdleLocked makes the run idle when its turn's reply was kept and its
// wrapper is ready for the next: both, in either order. b.mu and r.mu held.
func (r *tuiRuntimeSubscriptionRun) enterIdleLocked() (uint64, bool) {
	if r.hotIdle || r.hotTip == nil || r.closed || r.left || r.turnsReady != r.turnsSent || r.turnOut != nil {
		return 0, false
	}
	r.hotIdle = true
	r.hotGen++
	r.hotSince = time.Now()
	b := r.bridge
	if b.hot == nil {
		b.hot = map[tuiRuntimePrefixDigest][]*tuiRuntimeSubscriptionRun{}
	}
	b.hot[*r.hotTip] = append(b.hot[*r.hotTip], r)
	return r.hotGen, true
}

// armIdle starts gen's wait: the TTL, the pool's bound and the room's
// waiters, who may end it for room.
func (r *tuiRuntimeSubscriptionRun) armIdle(gen uint64) {
	ttl := tuiRuntimeIdleTTL()
	t := time.AfterFunc(ttl, func() { r.letGoIdle(gen, "it waited "+ttl.String()+" for its next turn") })
	r.mu.Lock()
	if r.hotIdle && r.hotGen == gen {
		r.hotTimer = t
	} else {
		t.Stop() // claimed or let go meanwhile
	}
	r.mu.Unlock()
	if r.timer != nil {
		r.timer.Stop() // the active turn's bound; the TTL bounds the wait
	}
	r.bridge.boundIdle()
	if tuiRuntimeSettings.Nudge != nil {
		tuiRuntimeSettings.Nudge()
	}
}

// letGoIdle ends the run if it is still idle in wait gen: a claim or an
// earlier let-go since takes gen with it, so a late timer ends no turn.
func (r *tuiRuntimeSubscriptionRun) letGoIdle(gen uint64, why string) bool {
	b := r.bridge
	b.mu.Lock()
	r.mu.Lock()
	ok := r.hotIdle && r.hotGen == gen
	if ok {
		r.leaveIdleLocked()
	}
	model := r.model
	r.mu.Unlock()
	b.mu.Unlock()
	if ok {
		log.Printf("Claude interactive idle process on %s ended: %s", model, why)
		// counted as leaving at once, for the room's waiters; the teardown
		// itself waits on the dispatcher this may be called from
		r.markLeaving()
		go r.abort()
	}
	return ok
}

// markLeaving counts the run's process as one magpie asked to end, once.
func (r *tuiRuntimeSubscriptionRun) markLeaving() {
	r.mu.Lock()
	leaving := r.tree != nil && !r.left && !r.closed
	if leaving {
		r.left = true
	}
	r.mu.Unlock()
	if leaving {
		tuiRuntimeSettings.Leave()
	}
}

// leaveIdleLocked takes the run out of the idle pool. b.mu and r.mu held.
func (r *tuiRuntimeSubscriptionRun) leaveIdleLocked() {
	if !r.hotIdle {
		return
	}
	r.hotIdle = false
	r.hotGen++
	if r.hotTimer != nil {
		r.hotTimer.Stop()
		r.hotTimer = nil
	}
	b := r.bridge
	if r.hotTip != nil {
		runs := slices.DeleteFunc(b.hot[*r.hotTip], func(o *tuiRuntimeSubscriptionRun) bool { return o == r })
		if len(runs) == 0 {
			delete(b.hot, *r.hotTip)
		} else {
			b.hot[*r.hotTip] = runs
		}
	}
	r.hotTip = nil
}

// boundIdle ends the processes idle longest past tuiRuntimeServeIdleMost.
func (b *tuiRuntimeSubscriptionBridge) boundIdle() {
	type idle struct {
		run *tuiRuntimeSubscriptionRun
		gen uint64
		at  time.Time
	}
	var all []idle
	b.mu.Lock()
	for _, runs := range b.hot {
		for _, r := range runs {
			r.mu.Lock()
			all = append(all, idle{r, r.hotGen, r.hotSince})
			r.mu.Unlock()
		}
	}
	b.mu.Unlock()
	if len(all) <= tuiRuntimeServeIdleMost {
		return
	}
	slices.SortFunc(all, func(a, b idle) int { return a.at.Compare(b.at) })
	for _, w := range all[:len(all)-tuiRuntimeServeIdleMost] {
		w.run.letGoIdle(w.gen, fmt.Sprintf("more than %d processes were idle", tuiRuntimeServeIdleMost))
	}
}

// tuiRuntimeLetGoIdleForRoom ends the process idle longest, of any bridge,
// for claudeRoom: an idle process goes before one working or waiting on its
// client's tools. It says whether it found one.
func tuiRuntimeLetGoIdleForRoom(most int) bool {
	for {
		var oldest *tuiRuntimeSubscriptionRun
		var gen uint64
		var at time.Time
		tuiRuntimeBridges.Range(func(k, _ any) bool {
			b := k.(*tuiRuntimeSubscriptionBridge)
			b.mu.Lock()
			for _, runs := range b.hot {
				for _, r := range runs {
					r.mu.Lock()
					if oldest == nil || r.hotSince.Before(at) {
						oldest, gen, at = r, r.hotGen, r.hotSince
					}
					r.mu.Unlock()
				}
			}
			b.mu.Unlock()
			return true
		})
		if oldest == nil {
			return false
		}
		if oldest.letGoIdle(gen, fmt.Sprintf("another Claude Code needed its room (%d may run at once)", most)) {
			return true
		}
		// claimed meanwhile: look again
	}
}

// startHot goes on with req in an idle process whose session ends exactly
// where req's history does, before its new user turn: the same scope, the
// same messages and the reply as its client was told it, and a launch
// req's own would not change. The claim takes the process from the pool
// before anything is typed in, so one process serves one request.
func (b *tuiRuntimeSubscriptionBridge) startHot(req *tuiRuntimeRequest, model, owner, outerSession, configDir string) (*tuiRuntimeSubscriptionRun, <-chan Event, bool) {
	if tuiRuntimeIdleTTL() <= 0 || req.scope == "" {
		return nil, nil, false
	}
	// the same request guards a fresh launch holds to: a turn typed into a
	// kept process is still this request, not a way past them
	if err := tuiRuntimeValidateInteractiveSafeguards(req); err != nil {
		return nil, nil, false
	}
	switch {
	case req.ToolChoice == "", req.ToolChoice == "auto", req.ToolChoice == "none",
		req.ToolChoice == "required", strings.HasPrefix(req.ToolChoice, "name:"):
	default:
		return nil, nil, false
	}
	last := tuiRuntimeLastAssistant(req.Messages)
	if last < 0 || last == len(req.Messages)-1 {
		return nil, nil, false
	}
	since := req.Messages[last+1:]
	for _, m := range since {
		for _, p := range m.Parts {
			if p.Kind == ToolResult {
				return nil, nil, false // results answer calls: a continuation, not a new turn
			}
		}
	}
	tip := tuiRuntimePrefixOf(req.scope, req.Messages[:last+1])
	b.mu.Lock()
	candidates := len(b.hot[tip])
	b.mu.Unlock()
	if candidates == 0 {
		return nil, nil, false
	}
	prompt, images, err := tuiRuntimeRenderInteractiveTurnPromptWithImages(since)
	if err != nil {
		return nil, nil, false
	}
	prompt = tuiRuntimeRenderInteractiveTurnConstraint(req) + prompt
	want := map[bool]tuiRuntimeHotConfig{
		false: tuiRuntimeHotConfigOf(req, model, outerSession, configDir, false),
		true:  tuiRuntimeHotConfigOf(req, model, outerSession, configDir, true),
	}
	run, why := b.claimHot(tip, want, len(images) > 0)
	if run == nil {
		log.Printf("Claude interactive idle process at the request's prefix not taken: %s", why)
		return nil, nil, false
	}
	if outerSession != "" {
		// the conversation's checkpoint names this session at an earlier
		// reply: dirty before the turn's first possible side effect, as a
		// resume makes it (markInteractiveSessionDirty)
		saved, ok, err := b.loadInteractiveSession(owner, outerSession)
		if err == nil && ok && saved.SessionID == run.sessionIDNow() {
			err = b.markInteractiveSessionDirty(owner, outerSession)
		}
		if err != nil {
			log.Printf("Claude interactive idle process not taken: its checkpoint can't be marked: %v", err)
			go run.abort()
			return nil, nil, false
		}
	}
	events, err := run.submitHotTurn(prompt, images)
	if err != nil {
		log.Printf("Claude interactive idle process failed to take its turn: %v", err)
		go run.abort()
		return nil, nil, false
	}
	return run, events, true
}

// claimHot takes from the pool an idle process at tip whose launch matches
// want (by whether the turn has images), leaving it out of the pool and
// its wait. It says why none was taken.
func (b *tuiRuntimeSubscriptionBridge) claimHot(tip tuiRuntimePrefixDigest, want map[bool]tuiRuntimeHotConfig, images bool) (*tuiRuntimeSubscriptionRun, string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	why := "none idle there"
	for _, r := range slices.Clone(b.hot[tip]) {
		r.mu.Lock()
		have := r.hotConfig
		w := want[have.imageTool]
		switch {
		case !r.hotIdle || r.closed:
			why = "it was ending"
		case images && !have.imageTool:
			why = "the turn has images and the process no image tool"
		case have.held:
			why = "the process holds earlier tool definitions"
		case have != w:
			why = "the request's launch differs: " + have.differs(w)
		default:
			r.leaveIdleLocked()
			r.mu.Unlock()
			return r, ""
		}
		r.mu.Unlock()
	}
	return nil, why
}

// differs names the fields of c that o doesn't have, never their values.
func (c tuiRuntimeHotConfig) differs(o tuiRuntimeHotConfig) string {
	var out []string
	for _, f := range []struct {
		name string
		same bool
	}{
		{"scope", c.scope == o.scope}, {"model", c.model == o.model}, {"instructions", c.instructions == o.instructions},
		{"effort", c.effort == o.effort}, {"conversation", c.outerSession == o.outerSession}, {"account folder", c.configDir == o.configDir},
		{"lean prompt", c.lean == o.lean}, {"thinking", c.thinkOff == o.thinkOff}, {"native search", c.nativeSearch == o.nativeSearch},
		{"tools", c.tools == o.tools}, {"safeguards", c.safeguards == o.safeguards && c.safeguardBeta == o.safeguardBeta},
	} {
		if !f.same {
			out = append(out, f.name)
		}
	}
	return fmt.Sprint(out)
}

func (r *tuiRuntimeSubscriptionRun) sessionIDNow() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.sessionID
}

// submitHotTurn types the next turn into a claimed process and gives the
// segment its events come on. What the last turn noted of its reply and
// calls is forgotten; a callback queued with no turn to take it is refused.
func (r *tuiRuntimeSubscriptionRun) submitHotTurn(prompt string, images map[string]Part) (<-chan Event, error) {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil, errors.New("the native Claude Code has ended")
	}
	if r.hotConfig.imageTool {
		r.images = images
	}
	r.replied, r.terminalDelivered, r.resumeBase = false, false, nil
	r.asked, r.shown, r.parkedAt = nil, nil, time.Time{}
	r.abandonedByClient = false
	queued := r.queuedCallbacks
	r.queuedCallbacks = nil
	r.mu.Unlock()
	for _, c := range queued {
		log.Printf("Claude interactive MCP call made between turns refused: tool=%q", c.name)
		r.rejectUnshownInteractiveMCP(c.id)
	}
	if r.timer != nil {
		r.timer.Reset(tuiRuntimeInteractiveActiveLongest)
	}
	segment := r.attach()
	r.beginServeTurn()
	if err := r.submitServe(prompt); err != nil {
		return nil, err
	}
	return segment, nil
}
