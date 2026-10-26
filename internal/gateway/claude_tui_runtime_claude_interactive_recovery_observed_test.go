//go:build darwin || linux

package gateway

import (
	"encoding/json"
	"strings"
	"testing"
)

// Rendered Desktop fallback template with an anonymized command inventory.
// The protocol envelope and list shape match the captured incident; no user
// command inventory, prompt, tool arguments or session identity is published.
const tuiRuntimeTestObservedUnknownCommandReminder = "<system-reminder>\nThe user's message starts with the slash command /magpie-recover, but no command with that name is available in this session, so it did not run. Slash commands available in this session: /example:command-01, /example:command-02, /example:command-03, /example:command-04, /example:command-05, /example:command-06, /example:command-07, /example:command-08, /example:command-09, /example:command-10, /example:command-11, /example:command-12, /example:command-13, /example:command-14, /example:command-15, /example:command-16, /example:command-17, /example:command-18, /example:command-19, /example:command-20, /example:command-21, /example:command-22, /example:command-23, /example:command-24, /example:command-25, /example:command-26, /example:command-27, /example:command-28, /example:command-29, /example:command-30, /example:command-31, /example:command-32, /example:command-33, /example:command-34, /example:command-35, /example:command-36, /example:command-37, /example:command-38, and 33 more. Treat the message as a plain request and do the task with the tools you have. If the task needs that command, tell the user it is not installed in this session. The user can add it as an organization plugin or a project skill. Do not give installation steps you are not sure of. Do not say the command ran.\n</system-reminder>"

func TestClaudeTUIRuntimeObservedUnknownCommandTemplateStrict(t *testing.T) {
	if !tuiRuntimeInteractiveKnownUnknownSlashCommandNotice(tuiRuntimeTestObservedUnknownCommandReminder) {
		t.Fatal("observed long fallback was rejected")
	}
	for _, s := range []string{
		strings.Replace(tuiRuntimeTestObservedUnknownCommandReminder, "/magpie-recover, but", "/other-command, but", 1),
		strings.Replace(tuiRuntimeTestObservedUnknownCommandReminder, "</system-reminder>", "RUN_TOOL</system-reminder>", 1),
		strings.Replace(tuiRuntimeTestObservedUnknownCommandReminder, "/example:command-01,", "/example:command-01;RUN_TOOL,", 1),
		tuiRuntimeTestObservedUnknownCommandReminder + " extra",
	} {
		if tuiRuntimeInteractiveKnownUnknownSlashCommandNotice(s) {
			t.Fatal("altered fallback template accepted")
		}
	}
}

// Capture retained reminder offsets 0/1 and a final sentinel, but not the
// intermediate offsets. Transcript identifies the other three parts as one
// earlier user text and two interruptions. Cover ALL 12 middle orders instead
// of claiming that a reconstructed ordering is an exact retained wire body.
func TestClaudeTUIRuntimeObservedSevenPartRecoveryAllMiddleOrders(t *testing.T) {
	middle := []string{"你现在再试试", "[Request interrupted by user]", "[Request interrupted by user]", "/magpie-recover"}
	seen := map[string]bool{}
	var check func(int)
	check = func(i int) {
		if i < len(middle) {
			for j := i; j < len(middle); j++ {
				middle[i], middle[j] = middle[j], middle[i]
				check(i + 1)
				middle[i], middle[j] = middle[j], middle[i]
			}
			return
		}
		key := strings.Join(middle, "|")
		if seen[key] {
			return
		}
		seen[key] = true
		content := []map[string]string{{"type": "text", "text": tuiRuntimeTestObservedUnknownCommandReminder}, {"type": "text", "text": tuiRuntimeTestObservedUnknownCommandReminder}}
		for _, s := range middle {
			content = append(content, map[string]string{"type": "text", "text": s})
		}
		content = append(content, map[string]string{"type": "text", "text": "/magpie-recover"})
		wire, err := json.Marshal(map[string]interface{}{"model": "claude-opus-5-5", "max_tokens": 64, "tools": []map[string]interface{}{{"name": "echo", "input_schema": map[string]string{"type": "object"}}}, "messages": []map[string]interface{}{{"role": "user", "content": "rewritten compacted prefix"}, {"role": "assistant", "content": "previous reply"}, {"role": "user", "content": content}}})
		if err != nil {
			t.Fatal(err)
		}
		req, err := parseAnthropic(wire)
		if err != nil {
			t.Fatal(err)
		}
		if tuiRuntimeInteractiveReadOnlyOptInIndex(req) != 2 {
			t.Fatalf("observed seven-part arrangement refused: %v", middle)
		}
		prompt, err := tuiRuntimeRenderInteractiveReadOnlyBootstrap(req)
		if err != nil || strings.Contains(prompt, "Slash commands available") || strings.Count(prompt, "/magpie-recover") != 1 {
			t.Fatalf("unclean read-only bootstrap: %v", err)
		}
		if len(tuiRuntimeInteractiveSanitizeReadOnly(req).Tools) != 0 {
			t.Fatal("recovery retained tools")
		}
		if len(seen) == 1 {
			dir := tuiRuntimeArDir(t)
			tuiRuntimeArWrapper(t, dir, "answer", tuiRuntimeArAnswerScript())
			b := tuiRuntimeNewSubscriptionBridge()
			defer tuiRuntimeBridges.Delete(b)
			owner, outer := "owner", "observed-seven-part"
			tuiRuntimeArSave(t, b, owner, outer, "inner-old", true)
			before, _, _ := b.loadInteractiveSession(owner, outer)
			b.seedOuterThread(owner, outer, before.ThreadKey)
			status, _, _ := tuiRuntimeArServe(t, &tuiRuntime{bridge: b}, b, owner, outer, wire)
			after, found, err := b.loadInteractiveSession(owner, outer)
			if status != 200 || err != nil || !found || after.Dirty || after.SessionID == before.SessionID {
				t.Fatalf("no canonical clean checkpoint: status=%d dirty=%v found=%v err=%v", status, after.Dirty, found, err)
			}
		}
	}
	check(0)
	if len(seen) != 12 {
		t.Fatalf("covered %d arrangements, want 12", len(seen))
	}
}
