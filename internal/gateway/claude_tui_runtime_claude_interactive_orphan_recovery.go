package gateway

import (
	"log"
	"os"
	"slices"
)

// Automatic bounded read-only recovery for an orphan PTY tool handoff.
//
// A client-side compaction (Claude Desktop's /compact, or any harness that
// rewrites its own history) can land while a PTY tool call is still out. The
// live run is then let go, the checkpoint stays dirty, and the tool result the
// client still carries matches no run. The gateway refuses to replay history,
// which is right, but an unattended client then retries the same orphan result
// forever.
//
// This seam turns exactly that shape into the recovery the human would be
// asked to type: the request gains a real `/magpie-recover` user turn and the
// whole turn is sanitised tool-free (readOnly) by the existing opt-in path. It
// never replays a tool, never injects an unverified result as fact, and only
// fires when a dirty checkpoint proves the conversation had an owned native
// session whose verified prefix the compacted history no longer contains.
const tuiRuntimeAutoOrphanRecoveryEnv = "MAGPIE_CLAUDE_AUTO_READONLY_ORPHAN_HANDOFF"

func tuiRuntimeAutoOrphanRecoveryEnabled() bool {
	return os.Getenv(tuiRuntimeAutoOrphanRecoveryEnv) == "1"
}

// beginOrphanReadOnlyRecovery appends the opt-in sentinel when the request is
// an orphan tool handoff over a rewritten (compacted) dirty checkpoint. It
// returns true only when the request now carries a recognised sentinel as its
// last user turn; callers re-derive the index after sanitising.
func (b *tuiRuntimeSubscriptionBridge) beginOrphanReadOnlyRecovery(owner, outerSession, model string, req *tuiRuntimeRequest) bool {
	// Reasons are fixed enums only: never log identities, keys, history,
	// model names, tool parameters or raw storage errors.
	reject := func(reason string) bool {
		log.Printf("Claude interactive orphan auto-recovery refused: reason=%s", reason)
		return false
	}
	if !tuiRuntimeAutoOrphanRecoveryEnabled() {
		return reject("disabled")
	}
	if b == nil || req == nil || outerSession == "" {
		return reject("missing_input")
	}
	saved, ok, err := b.loadInteractiveSession(owner, outerSession)
	if err != nil {
		return reject("checkpoint_store_invalid")
	}
	if !ok {
		return reject("checkpoint_missing")
	}
	if !saved.Dirty {
		return reject("checkpoint_clean")
	}
	if saved.ConvKey == "" {
		return reject("checkpoint_anchor_missing")
	}
	if saved.Model != "" && saved.Model != model {
		return reject("model_mismatch")
	}
	// Only a rewritten history qualifies: the checkpoint's verified prefix must
	// be gone entirely. A live-but-mismatched prefix stays a normal refusal.
	if _, status := tuiRuntimeConversationSuffixAfterDetailed(owner, req.Messages, saved.ConvKey); status != tuiRuntimeConversationSuffixCheckpointNotFound {
		if status == tuiRuntimeConversationSuffixOK {
			return reject("checkpoint_prefix_present")
		}
		return reject("checkpoint_suffix_rejected")
	}
	req.Messages = append(slices.Clone(req.Messages), Message{Role: "user", Parts: []Part{{Kind: Text, Text: "/magpie-recover"}}})
	log.Printf("Claude interactive orphan tool handoff: bounded read-only recovery over a compacted dirty checkpoint")
	return true
}
