package gateway

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"hash"
	"slices"

	"github.com/yetone/magpie/internal/provider"
)

// A conversation's prefix is named by a digest of its messages exactly as
// the client sent them: every role, text, tool call's id, name and
// arguments, tool result with its error flag and images, image and file
// with a digest of its bytes, and thinking with its signature, in order.
// Nothing is ignored or matched loosely: a history whose digest is equal is
// the same history. Only representations the protocol defines as equal are
// made one: a tool call's arguments as one JSON object, whatever their key
// order and spacing. Where a part came from (SignedBy, SealedBy) and the
// client's cache marks are not content and are left out.
//
// Each digest is a hash of the one before and of one message, each field
// length-prefixed, so one pass over a request names every prefix of it:
// H0 = H(version, scope), Hi = H(Hi-1, message i). The scope is the
// isolation domain a native session may be found in (tuiRuntimePrefixScope).
const tuiRuntimePrefixVersion = "claude-tui-prefix-v1"

type tuiRuntimePrefixDigest [32]byte

// tuiRuntimePrefixScope is the isolation domain a native session is found
// in: the account it runs as (owner names the provider, account and MCP
// server) and the gateway key that asked. Two callers of one account never
// go on in each other's sessions.
func tuiRuntimePrefixScope(owner, callerKey string) string {
	return owner + "\x00key=" + callerKey
}

// tuiRuntimePrefixDigests is the digest of msgs[:i+1] for every i.
func tuiRuntimePrefixDigests(scope string, msgs []Message) []tuiRuntimePrefixDigest {
	out := make([]tuiRuntimePrefixDigest, len(msgs))
	prev := tuiRuntimePrefixStart(scope)
	for i, m := range msgs {
		prev = tuiRuntimePrefixNext(prev, m)
		out[i] = prev
	}
	return out
}

// tuiRuntimePrefixOf is the digest of msgs whole.
func tuiRuntimePrefixOf(scope string, msgs []Message) tuiRuntimePrefixDigest {
	d := tuiRuntimePrefixStart(scope)
	for _, m := range msgs {
		d = tuiRuntimePrefixNext(d, m)
	}
	return d
}

func tuiRuntimePrefixStart(scope string) tuiRuntimePrefixDigest {
	h := sha256.New()
	e := tuiRuntimePrefixEncoder{h}
	e.str(tuiRuntimePrefixVersion)
	e.str(scope)
	var d tuiRuntimePrefixDigest
	h.Sum(d[:0])
	return d
}

func tuiRuntimePrefixNext(prev tuiRuntimePrefixDigest, m Message) tuiRuntimePrefixDigest {
	h := sha256.New()
	h.Write(prev[:])
	e := tuiRuntimePrefixEncoder{h}
	e.str("message")
	e.str(m.Role)
	e.parts(m.Parts)
	var d tuiRuntimePrefixDigest
	h.Sum(d[:0])
	return d
}

type tuiRuntimePrefixEncoder struct{ h hash.Hash }

func (e tuiRuntimePrefixEncoder) n(v int) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(v))
	e.h.Write(b[:])
}

func (e tuiRuntimePrefixEncoder) str(s string) {
	e.n(len(s))
	e.h.Write([]byte(s))
}

func (e tuiRuntimePrefixEncoder) bytes(b []byte) {
	e.n(len(b))
	e.h.Write(b)
}

func (e tuiRuntimePrefixEncoder) flag(b bool) {
	if b {
		e.n(1)
	} else {
		e.n(0)
	}
}

// data names inline bytes by their digest: a request's image is hashed
// once here, not kept.
func (e tuiRuntimePrefixEncoder) data(s string) {
	sum := sha256.Sum256([]byte(s))
	e.n(len(s))
	e.h.Write(sum[:])
}

func (e tuiRuntimePrefixEncoder) parts(parts []Part) {
	e.n(len(parts))
	for _, p := range parts {
		e.str(string(p.Kind))
		e.str(p.Text)
		e.str(p.MediaType)
		e.data(p.Data)
		e.str(p.URL)
		e.str(p.Detail)
		e.str(p.ID)
		e.str(p.Name)
		if p.Kind == ToolCall {
			e.bytes(tuiRuntimeCanonicalArgs(p.Args))
		} else {
			e.bytes(p.Args)
		}
		e.str(p.CallID)
		e.flag(p.IsError)
		e.parts(p.Images)
		e.bytes(tuiRuntimeCanonicalJSON(p.Standalone))
		e.n(len(p.Loads))
		for _, l := range p.Loads {
			e.str(l)
		}
		e.str(p.Signature)
		e.bytes(p.Sealed)
		e.n(len(p.Hits))
		for _, hit := range p.Hits {
			e.str(hit.Title)
			e.str(hit.URL)
			e.str(hit.PageAge)
			e.str(hit.Cited)
		}
	}
}

// tuiRuntimeCanonicalArgs is a tool call's arguments as one JSON object:
// keys sorted, no spacing, numbers as written. Absent or null arguments are
// the empty object, as the Messages API takes them. Bytes that aren't JSON
// are kept as they are, and match only themselves.
func tuiRuntimeCanonicalArgs(raw json.RawMessage) []byte {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return []byte("{}")
	}
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	dec.UseNumber()
	var v any
	if dec.Decode(&v) != nil || dec.More() {
		return slices.Clone(trimmed)
	}
	out, err := json.Marshal(v)
	if err != nil {
		return slices.Clone(trimmed)
	}
	return out
}

func tuiRuntimeCanonicalJSON(v map[string]any) []byte {
	if v == nil {
		return nil
	}
	out, err := json.Marshal(v)
	if err != nil {
		return []byte("?")
	}
	return tuiRuntimeCanonicalArgs(out)
}

// tuiRuntimeReplyRecorder keeps the reply a run's turn gave its client, as
// the client is told it and sends it back in its next request: thinking
// blocks with their signatures, text, and tool calls with their arguments,
// one part each, in order.
//
// A Messages client keeps the reply as the blocks it was sent, and those
// are the encoder's: a search is two blocks of its own, and after it each
// line naming a page found is a text block (search_cite.go). For one, the
// reply is collected as the relay collects it and laid out by
// renderAnthropic, whose blocks are the stream's, then read back as its
// next request is (messagesReply).
type tuiRuntimeReplyRecorder struct {
	parts []Part
	// open is the kind the last part is still being written as
	open Kind
	// messages: the client speaks Anthropic Messages, and col collects
	// the reply for messagesReply
	messages bool
	col      collector
}

// tuiRuntimeNewReplyRecorder records the reply for a client of the API
// client.
func tuiRuntimeNewReplyRecorder(client provider.Protocol) *tuiRuntimeReplyRecorder {
	return &tuiRuntimeReplyRecorder{messages: client == provider.Anthropic}
}

func (rec *tuiRuntimeReplyRecorder) note(ev Event) {
	if rec.messages {
		rec.col.add(ev)
		return
	}
	last := func() *Part { return &rec.parts[len(rec.parts)-1] }
	switch ev.Kind {
	case KThinkStart:
		rec.parts = append(rec.parts, Part{Kind: Thinking})
		rec.open = Thinking
	case KThink:
		if rec.open != Thinking {
			rec.parts = append(rec.parts, Part{Kind: Thinking})
			rec.open = Thinking
		}
		last().Text += ev.Text
	case KSig:
		if rec.open != Thinking {
			rec.parts = append(rec.parts, Part{Kind: Thinking})
			rec.open = Thinking
		}
		last().Signature += ev.Text
	case KText:
		if rec.open != Text {
			rec.parts = append(rec.parts, Part{Kind: Text})
			rec.open = Text
		}
		last().Text += ev.Text
	case KToolStart:
		rec.parts = append(rec.parts, Part{Kind: ToolCall, ID: ev.ID, Name: ev.Name})
		rec.open = ToolCall
	case KToolArgs:
		if rec.open == ToolCall {
			last().Args = append(last().Args, ev.Text...)
		}
	}
}

// message is the reply as the client's next request holds it. Text the
// relay sends as one block is one part here too.
func (rec *tuiRuntimeReplyRecorder) message() Message {
	if rec.messages {
		return tuiRuntimeMessagesReply(rec.col.finish())
	}
	parts := slices.Clone(rec.parts)
	for i := range parts {
		if parts[i].Kind == ToolCall {
			parts[i].Args = json.RawMessage(tuiRuntimeCanonicalArgs(parts[i].Args))
		}
	}
	return Message{Role: "assistant", Parts: parts}
}

// messagesReply is res as a Messages client sends it back: the blocks
// renderAnthropic lays it out in, read by the parser its next request goes
// through. A reply that doesn't read back is one with no parts, which
// nothing is resumed at.
func tuiRuntimeMessagesReply(res Result) Message {
	var rendered struct {
		Content json.RawMessage `json:"content"`
	}
	if json.Unmarshal(renderAnthropic(res, ""), &rendered) != nil {
		return Message{Role: "assistant"}
	}
	body, err := json.Marshal(map[string]any{"messages": []any{
		map[string]any{"role": "assistant", "content": rendered.Content},
	}})
	if err != nil {
		return Message{Role: "assistant"}
	}
	req, err := parseAnthropic(body)
	if err != nil || len(req.Messages) != 1 {
		return Message{Role: "assistant"}
	}
	return req.Messages[0]
}

// tee forwards events as they come, noting each, and closes its own channel
// when events closes.
func (rec *tuiRuntimeReplyRecorder) tee(events <-chan Event) <-chan Event {
	out := make(chan Event, cap(events))
	go func() {
		defer close(out)
		for ev := range events {
			rec.note(ev)
			out <- ev
		}
	}()
	return out
}

// tuiRuntimeDrain reads what is left of events once its relay is done, so
// the run's sends to it never wait on a reader that left.
func tuiRuntimeDrain(events <-chan Event) {
	go func() {
		for range events {
		}
	}()
}

// tuiRuntimeLastAssistant is the index of msgs' last assistant message, -1
// with none.
func tuiRuntimeLastAssistant(msgs []Message) int {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == "assistant" {
			return i
		}
	}
	return -1
}
