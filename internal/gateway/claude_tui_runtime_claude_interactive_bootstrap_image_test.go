package gateway

// DAV-1940: production image regression — an ordinary historical bootstrap
// (an outer conversation that predates persisted inner checkpoints, and the
// ephemeral side-thread bootstrap of the same shape) rejected the *current*
// turn's inline image with
//
//	"Claude interactive historical bootstrap supports only text in the
//	current user request"
//
// even though first-turn and persisted-resume interactive paths already
// route inline images through the private magpie_view_attached_image MCP
// tool. These tests pin the fix: the current user request of a historical
// bootstrap may carry inline base64 images (and may be image-only), the
// private viewer is registered for that turn only, historical tool traffic
// stays stripped, historical images stay inert, URL-only images keep the
// existing refusal, and the read-only recovery turn stays tool-free even
// when its merged row carries an image.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

var tuiRuntimeBootstrapImagePart = Part{Kind: Image, MediaType: "image/png", Data: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJ"}

// bootstrapImageRequest is the reported production shape: a long outer
// history that includes old tool traffic, then a current user request that
// carries text plus an inline image. The old code rejected it in
// validateTextUser with the exact 502 the user screenshotted.
func tuiRuntimeBootstrapImageRequest() *tuiRuntimeRequest {
	return &tuiRuntimeRequest{Request: Request{
		Model: "claude-opus-5-5", System: "CURRENT_SYSTEM", Tools: []Tool{{Name: "current_tool"}},
		Messages: []Message{
			{Role: "user", Parts: []Part{{Kind: Text, Text: "OLD_USER_TEXT"}}},
			{Role: "assistant", Parts: []Part{
				{Kind: Text, Text: "OLD_ASSISTANT_TEXT"},
				{Kind: ToolCall, ID: "old-call", Name: "dangerous_old_tool"},
			}},
			{Role: "user", Parts: []Part{{Kind: ToolResult, CallID: "old-call", Text: "STALE_TOOL_RESULT"}}},
			{Role: "assistant", Parts: []Part{{Kind: Text, Text: "POST_TOOL_VISIBLE_TEXT"}}},
			{Role: "user", Parts: []Part{
				{Kind: Text, Text: "CURRENT_WITH_IMAGE"},
				tuiRuntimeBootstrapImagePart,
			}},
		},
	}}
}

// RED->GREEN renderer contract: the bootstrap prompt names the private image
// tool and the attachment id inside the current-request block, keeps raw
// base64 out of the text, returns the bytes in the images map, and strips
// every historical tool name/result.
func TestClaudeTUIRuntimeInteractiveHistoricalBootstrapAcceptsCurrentInlineImage(t *testing.T) {
	req := tuiRuntimeBootstrapImageRequest()
	prompt, images, err := tuiRuntimeRenderInteractiveHistoricalBootstrapWithImages(req)
	if err != nil {
		t.Fatalf("current-turn inline image was rejected: %v", err)
	}
	id := tuiRuntimeInteractiveImageID(tuiRuntimeBootstrapImagePart)
	cur := prompt[strings.Index(prompt, "<current_user_request>"):]
	if !strings.Contains(cur, "CURRENT_WITH_IMAGE") ||
		!strings.Contains(cur, "mcp__magpie__"+tuiRuntimeInteractiveImageTool) ||
		!strings.Contains(cur, id) {
		t.Fatalf("current request block lacks image marker/id:\n%s", prompt)
	}
	if strings.Contains(prompt, tuiRuntimeBootstrapImagePart.Data) {
		t.Fatal("raw image bytes leaked into the text prompt")
	}
	if got, ok := images[id]; !ok || got.Data != tuiRuntimeBootstrapImagePart.Data || got.MediaType != tuiRuntimeBootstrapImagePart.MediaType {
		t.Fatalf("image map lost the bootstrap attachment: %#v, %v", got, ok)
	}
	if len(images) != 1 {
		t.Fatalf("unexpected extra attachments: %#v", images)
	}
	for _, want := range []string{"<historical_conversation_import>", "OLD_USER_TEXT", "OLD_ASSISTANT_TEXT", "POST_TOOL_VISIBLE_TEXT", "CURRENT_SYSTEM"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("bootstrap prompt missing %q:\n%s", want, prompt)
		}
	}
	for _, stale := range []string{"dangerous_old_tool", "STALE_TOOL_RESULT"} {
		if strings.Contains(prompt, stale) {
			t.Fatalf("historical bootstrap replayed stale work %q:\n%s", stale, prompt)
		}
	}
	// The legacy text-only wrapper stays source-compatible and returns the
	// same prompt bytes.
	plain, err := tuiRuntimeRenderInteractiveHistoricalBootstrap(req)
	if err != nil || plain != prompt {
		t.Fatalf("text wrapper diverged: err=%v same=%v", err, plain == prompt)
	}
}

func TestClaudeTUIRuntimeInteractiveHistoricalBootstrapImageOnlyCurrentRequest(t *testing.T) {
	req := tuiRuntimeBootstrapImageRequest()
	msgs := req.Messages
	msgs[len(msgs)-1].Parts = []Part{tuiRuntimeBootstrapImagePart}
	prompt, images, err := tuiRuntimeRenderInteractiveHistoricalBootstrapWithImages(req)
	if err != nil {
		t.Fatalf("image-only current request rejected: %v", err)
	}
	if len(images) != 1 || !strings.Contains(prompt, tuiRuntimeInteractiveImageID(tuiRuntimeBootstrapImagePart)) {
		t.Fatalf("image-only request lost its attachment: %d images\n%s", len(images), prompt)
	}
}

func TestClaudeTUIRuntimeInteractiveHistoricalBootstrapMultipleImagesAndSystemNotice(t *testing.T) {
	second := Part{Kind: Image, MediaType: "image/png", Data: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSK"}
	req := tuiRuntimeBootstrapImageRequest()
	req.Messages = append(req.Messages[:len(req.Messages)-1],
		Message{Role: "system", Parts: []Part{{Kind: Text, Text: "TAIL_SYSTEM_NOTICE"}}},
		Message{Role: "user", Parts: []Part{
			{Kind: Text, Text: "CURRENT_WITH_IMAGES"},
			tuiRuntimeBootstrapImagePart, second,
		}})
	prompt, images, err := tuiRuntimeRenderInteractiveHistoricalBootstrapWithImages(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 2 {
		t.Fatalf("expected two attachments, got %d", len(images))
	}
	for _, p := range []Part{tuiRuntimeBootstrapImagePart, second} {
		if _, ok := images[tuiRuntimeInteractiveImageID(p)]; !ok {
			t.Fatalf("attachment %s missing", tuiRuntimeInteractiveImageID(p))
		}
	}
	if !strings.Contains(prompt, "TAIL_SYSTEM_NOTICE") {
		t.Fatal("trailing system notice was dropped from external instructions")
	}
	// Both markers live inside the current-request block, not the history.
	cur := prompt[strings.Index(prompt, "<current_user_request>"):]
	if strings.Count(cur, "[Attached image id=") != 2 {
		t.Fatalf("expected two image markers in current block:\n%s", cur)
	}
}

// A current-turn URL-only image keeps the existing refusal rule: no remote
// fetch, no silent drop.
func TestClaudeTUIRuntimeInteractiveHistoricalBootstrapRejectsURLImage(t *testing.T) {
	req := tuiRuntimeBootstrapImageRequest()
	msgs := req.Messages
	msgs[len(msgs)-1].Parts = []Part{
		{Kind: Text, Text: "CURRENT"},
		{Kind: Image, URL: "https://example.com/a.png"},
	}
	if _, _, err := tuiRuntimeRenderInteractiveHistoricalBootstrapWithImages(req); err == nil || !strings.Contains(err.Error(), "URL image") {
		t.Fatalf("URL-only image was not refused: %v", err)
	}
}

// Historical images stay inert: they are not re-registered as attachments
// and their bytes are never copied into the prompt.
func TestClaudeTUIRuntimeInteractiveHistoricalBootstrapDoesNotReviveHistoricalImages(t *testing.T) {
	old := Part{Kind: Image, MediaType: "image/png", Data: "T0xEX0lNQUdFX0JZVEVT"}
	req := tuiRuntimeBootstrapImageRequest()
	msgs := req.Messages
	msgs[0].Parts = append(msgs[0].Parts, old)
	prompt, images, err := tuiRuntimeRenderInteractiveHistoricalBootstrapWithImages(req)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := images[tuiRuntimeInteractiveImageID(old)]; ok {
		t.Fatal("a historical image was re-activated as a current attachment")
	}
	if len(images) != 1 {
		t.Fatalf("expected only the current attachment, got %d", len(images))
	}
	if strings.Contains(prompt, old.Data) || strings.Contains(prompt, tuiRuntimeInteractiveImageID(old)) {
		t.Fatalf("historical image bytes/id leaked into the prompt:\n%s", prompt)
	}
}

// End-to-end GREEN on the ordinary outer-session bootstrap path: the PTY
// receives the rendered prompt, the private viewer tool is advertised for
// this turn, a real callback with the correct id returns the inline bytes,
// and a wrong id is refused. The clean checkpoint is written on completion.
func TestClaudeTUIRuntimeInteractiveHistoricalBootstrapImageEndToEnd(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "bootstrap-image-bridge")
	logPath := filepath.Join(dir, "bootstrap-image.log")
	script := `#!/bin/sh
prompt=$(cat)
printf 'args:%s\nprompt:%s\n' "$*" "$prompt" > "$INTERACTIVE_BOOTSTRAP_IMAGE_LOG"
printf '%s\n' '{"type":"assistant","uuid":"row","session_id":"inner-boot-img","message":{"id":"m","model":"claude-opus-5-5","content":[{"type":"text","text":"BOOT_IMAGE_OK"}]}}'
# Hold the run open so its MCP callback stays reachable while we probe it;
# readInteractiveOneTurn waits on the PTY pipes either way.
sleep 0.5
printf '%s\n' '{"type":"result","session_id":"inner-boot-img","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_BOOTSTRAP_IMAGE_LOG", logPath)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer, model := "owner", "old-desktop-session-img", "claude-opus-5-5"
	req := tuiRuntimeBootstrapImageRequest()
	run, events, err := b.startInteractiveRequest(context.Background(), req, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var text, stop string
	eventsDone := make(chan struct{})
	go func() {
		defer close(eventsDone)
		for ev := range events {
			switch ev.Kind {
			case KText:
				mu.Lock()
				text += ev.Text
				mu.Unlock()
			case KStop:
				mu.Lock()
				stop = ev.Stop
				mu.Unlock()
			case KError:
				t.Error(ev.Text)
			}
		}
	}()
	// The reply text row arrives while the wrapper still holds the run
	// open; wait for it before probing the callback (the run is removed
	// from b.runs the moment the result row finishes it).
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		done := strings.Contains(text, "BOOT_IMAGE_OK")
		mu.Unlock()
		if done {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	// The launched run must carry the private image tool and the attachment.
	run.mu.Lock()
	_, toolOK := run.tools[tuiRuntimeInteractiveImageTool]
	_, imgOK := run.images[tuiRuntimeInteractiveImageID(tuiRuntimeBootstrapImagePart)]
	run.mu.Unlock()
	if !toolOK || !imgOK {
		t.Fatalf("run lacks private viewer (%v) or attachment (%v)", toolOK, imgOK)
	}
	toolsPath := filepath.Join(run.tmp, "tools.json")
	toolBytes, err := os.ReadFile(toolsPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(toolBytes), tuiRuntimeInteractiveImageTool) || !strings.Contains(string(toolBytes), "current_tool") {
		t.Fatalf("tools.json missing private viewer or caller tool: %s", toolBytes)
	}
	// A real callback with the owned id returns the inline bytes; a wrong
	// id is refused — and neither is registered as a client tool call.
	call := func(imageID string) tuiRuntimeMcpToolResult {
		t.Helper()
		body := fmt.Sprintf(`{"tool_call_id":"call-img","name":%q,"arguments":{"id":%q}}`, tuiRuntimeInteractiveImageTool, imageID)
		r := httptest.NewRequest(http.MethodPost, "/cb/"+run.token, strings.NewReader(body))
		r.SetPathValue("token", run.token)
		rec := httptest.NewRecorder()
		b.mcpCall(rec, r)
		if rec.Code != http.StatusOK {
			t.Fatalf("image MCP status %d: %s", rec.Code, rec.Body.String())
		}
		var got tuiRuntimeMcpToolResult
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		return got
	}
	got := call(tuiRuntimeInteractiveImageID(tuiRuntimeBootstrapImagePart))
	if got.IsError || len(got.Content) != 2 || got.Content[1]["type"] != "image" ||
		got.Content[1]["data"] != tuiRuntimeBootstrapImagePart.Data || got.Content[1]["mimeType"] != tuiRuntimeBootstrapImagePart.MediaType {
		t.Fatalf("private image MCP returned wrong content: %#v", got)
	}
	missing := call("img_missing")
	if !missing.IsError {
		t.Fatalf("unknown image id was not refused: %#v", missing)
	}
	if len(b.calls) != 0 {
		t.Fatalf("private image MCP was registered as a client tool call: %#v", b.calls)
	}
	<-eventsDone
	mu.Lock()
	defer mu.Unlock()
	if text != "BOOT_IMAGE_OK" || stop != "stop" {
		t.Fatalf("image bootstrap = %q %q", text, stop)
	}
	run.ended(req, text, stop, true)
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	log := string(logBytes)
	for _, want := range []string{"<current_user_request>", "CURRENT_WITH_IMAGE", tuiRuntimeInteractiveImageTool, tuiRuntimeInteractiveImageID(tuiRuntimeBootstrapImagePart)} {
		if !strings.Contains(log, want) {
			t.Fatalf("bootstrap prompt missing %q:\n%s", want, log)
		}
	}
	for _, stale := range []string{"dangerous_old_tool", "STALE_TOOL_RESULT", tuiRuntimeBootstrapImagePart.Data, "--resume"} {
		if strings.Contains(log, stale) {
			t.Fatalf("bootstrap leaked/replayed %q:\n%s", stale, log)
		}
	}
	saved, ok := b.interactiveSession(owner, outer)
	if !ok || saved.SessionID != "inner-boot-img" || saved.Dirty {
		t.Fatalf("image bootstrap did not establish a clean checkpoint: %#v, %v", saved, ok)
	}
}

// The ephemeral side-thread bootstrap (no saved checkpoint, outer session
// already claimed by a different first-user fingerprint) gets the same image
// support but must never touch the canonical checkpoint.
func TestClaudeTUIRuntimeInteractiveHistoricalBootstrapImageSideThreadEphemeral(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "side-image-bridge")
	logPath := filepath.Join(dir, "side-image.log")
	script := `#!/bin/sh
prompt=$(cat)
printf '%s\n' '---' >> "$INTERACTIVE_SIDE_IMAGE_LOG"
printf 'args:%s\nprompt:%s\n' "$*" "$prompt" >> "$INTERACTIVE_SIDE_IMAGE_LOG"
case "$prompt" in
  *"CURRENT_WITH_IMAGE"*) sid=sess-side; text=SIDE_IMAGE_OK ;;
  *) sid=sess-main; text=MAIN ;;
esac
printf '{"type":"assistant","uuid":"row-%s","session_id":"%s","message":{"id":"m-%s","model":"claude-opus-5-5","content":[{"type":"text","text":"%s"}]}}\n' "$text" "$sid" "$text" "$text"
printf '{"type":"result","session_id":"%s","stop_reason":"end_turn"}\n' "$sid"
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	t.Setenv("INTERACTIVE_SIDE_IMAGE_LOG", logPath)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer, model := "owner", "desktop-side-image", "claude-opus-5-5"
	mainReq := &tuiRuntimeRequest{Request: Request{Model: model, Messages: []Message{{Role: "user", Parts: []Part{{Kind: Text, Text: "main thread"}}}}}}
	main, mainEvents, err := b.startInteractiveRequest(context.Background(), mainReq, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	var mainText, mainStop string
	for ev := range mainEvents {
		switch ev.Kind {
		case KText:
			mainText += ev.Text
		case KStop:
			mainStop = ev.Stop
		case KError:
			t.Fatal(ev.Text)
		}
	}
	main.ended(mainReq, mainText, mainStop, true)
	before, ok := b.interactiveSession(owner, outer)
	if !ok || before.SessionID != "sess-main" {
		t.Fatalf("main checkpoint missing: %#v, %v", before, ok)
	}

	// A side thread reuses the outer id but starts a different logical
	// conversation; it hits the historical bootstrap with outerSession "".
	sideReq := tuiRuntimeBootstrapImageRequest()
	side, sideEvents, err := b.startInteractiveRequest(context.Background(), sideReq, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	var sideText, sideStop string
	for ev := range sideEvents {
		switch ev.Kind {
		case KText:
			sideText += ev.Text
		case KStop:
			sideStop = ev.Stop
		case KError:
			t.Fatal(ev.Text)
		}
	}
	if sideText != "SIDE_IMAGE_OK" || sideStop != "stop" {
		t.Fatalf("side image bootstrap = %q %q", sideText, sideStop)
	}
	side.mu.Lock()
	_, toolOK := side.tools[tuiRuntimeInteractiveImageTool]
	_, imgOK := side.images[tuiRuntimeInteractiveImageID(tuiRuntimeBootstrapImagePart)]
	outerClaimed := side.outerSession
	side.mu.Unlock()
	if !toolOK || !imgOK {
		t.Fatalf("side run lacks private viewer (%v) or attachment (%v)", toolOK, imgOK)
	}
	if outerClaimed != "" {
		t.Fatalf("side thread claimed the canonical outer session %q", outerClaimed)
	}
	side.ended(sideReq, sideText, sideStop, true)
	after, ok := b.interactiveSession(owner, outer)
	if !ok || after.SessionID != before.SessionID || after.ConvKey != before.ConvKey || after.ReplyKey != before.ReplyKey {
		t.Fatalf("ephemeral side bootstrap advanced the main checkpoint:\nbefore=%#v\nafter=%#v", before, after)
	}
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	log := string(logBytes)
	if !strings.Contains(log, "CURRENT_WITH_IMAGE") || !strings.Contains(log, tuiRuntimeInteractiveImageID(tuiRuntimeBootstrapImagePart)) {
		t.Fatalf("side bootstrap prompt lacks image marker:\n%s", log)
	}
	if strings.Contains(log, "--resume sess-main") && strings.Contains(log, "CURRENT_WITH_IMAGE") {
		t.Fatalf("side bootstrap resumed the main inner session:\n%s", log)
	}
}

// The read-only recovery boundary is unchanged: even when the merged opt-in
// row carries an inline image, the recovery turn starts with zero tools —
// no private viewer — and any image callback is refused.
func TestClaudeTUIRuntimeInteractiveReadOnlyRecoveryKeepsToolsZeroWithImage(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	wrapper := filepath.Join(dir, "ro-image-bridge")
	script := `#!/bin/sh
cat >/dev/null
printf '%s\n' '{"type":"assistant","uuid":"row","session_id":"inner-ro-img","message":{"id":"m","model":"claude-opus-5-5","content":[{"type":"text","text":"RO_OK"}]}}'
# Hold the run open so the refused callback below still reaches it.
sleep 0.5
printf '%s\n' '{"type":"result","session_id":"inner-ro-img","stop_reason":"end_turn"}'
`
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	tuiRuntimeSetWrapper(t, wrapper)
	b := tuiRuntimeNewSubscriptionBridge()
	defer tuiRuntimeBridges.Delete(b)
	owner, outer, model := "owner", "outer-ro-img", "claude-opus-5-5"
	base := []Message{
		{Role: "user", Parts: []Part{{Kind: Text, Text: "original request"}}},
		{Role: "assistant", Parts: []Part{{Kind: Text, Text: "answer"}, {Kind: ToolCall, ID: "failed-call", Name: "write_file"}}},
	}
	baseReq := &tuiRuntimeRequest{Request: Request{Model: model, Messages: base}}
	keys := tuiRuntimeConversationKeys(owner, base)
	b.saveInteractiveSession(owner, outer, tuiRuntimeInteractiveSessionEntry{
		SessionID: "inner-ro-img", Model: model, ThreadKey: firstWords(&baseReq.Request),
		ConvKey: keys[0], ReplyKey: tuiRuntimeAssistantReplyKey(base[1]), ContextKey: tuiRuntimeTurnKey(owner, baseReq, nil), Dirty: true,
	})
	// Merged opt-in row: stale tool_result + current image + sentinel.
	req := &tuiRuntimeRequest{Request: Request{Model: model, Messages: []Message{
		base[0], base[1],
		{Role: "user", Parts: []Part{
			{Kind: ToolResult, CallID: "failed-call", Text: "write completed"},
			tuiRuntimeBootstrapImagePart,
			{Kind: Text, Text: "/magpie-recover"},
		}},
	}}}
	if tuiRuntimeInteractiveReadOnlyOptInIndex(req) < 0 {
		t.Fatal("merged opt-in row with an image was not recognized as a recovery turn")
	}
	run, events, err := b.startInteractiveRequest(context.Background(), req, model, "", owner, outer)
	if err != nil {
		t.Fatal(err)
	}
	var mu2 sync.Mutex
	var text, stop string
	eventsDone := make(chan struct{})
	go func() {
		defer close(eventsDone)
		for ev := range events {
			switch ev.Kind {
			case KText:
				mu2.Lock()
				text += ev.Text
				mu2.Unlock()
			case KStop:
				mu2.Lock()
				stop = ev.Stop
				mu2.Unlock()
			case KError:
				t.Error(ev.Text)
			}
		}
	}()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		mu2.Lock()
		done := strings.Contains(text, "RO_OK")
		mu2.Unlock()
		if done {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	run.mu.Lock()
	viewer := run.tools[tuiRuntimeInteractiveImageTool]
	nTools := len(run.tools)
	nImages := len(run.images)
	run.mu.Unlock()
	if viewer || nTools != 0 || nImages != 0 {
		t.Fatalf("read-only recovery gained tools/images: viewer=%v tools=%d images=%d", viewer, nTools, nImages)
	}
	toolBytes, err := os.ReadFile(filepath.Join(run.tmp, "tools.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(toolBytes), tuiRuntimeInteractiveImageTool) {
		t.Fatalf("read-only tools.json advertises the image viewer: %s", toolBytes)
	}
	// An image callback on this run is refused as a read-only violation.
	body := fmt.Sprintf(`{"tool_call_id":"call-img","name":%q,"arguments":{"id":%q}}`, tuiRuntimeInteractiveImageTool, tuiRuntimeInteractiveImageID(tuiRuntimeBootstrapImagePart))
	r := httptest.NewRequest(http.MethodPost, "/cb/"+run.token, strings.NewReader(body))
	r.SetPathValue("token", run.token)
	rec := httptest.NewRecorder()
	b.mcpCall(rec, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("read-only image callback status %d: %s", rec.Code, rec.Body.String())
	}
	var got tuiRuntimeMcpToolResult
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if !got.IsError || len(got.Content) == 0 || !strings.Contains(fmt.Sprint(got.Content[0]["text"]), "read-only") {
		t.Fatalf("read-only image callback was not refused: %#v", got)
	}
	<-eventsDone
	mu2.Lock()
	defer mu2.Unlock()
	if text != "RO_OK" || stop != "stop" {
		t.Fatalf("read-only recovery = %q %q", text, stop)
	}
	run.ended(req, text, stop, true)
}

// The bootstrap text prefix/layout is unchanged: same import header, same
// external-instructions block, same budget trim marker.
func TestClaudeTUIRuntimeInteractiveHistoricalBootstrapImageKeepsPrefixAndBudget(t *testing.T) {
	textOnly := &tuiRuntimeRequest{Request: Request{
		Model: "claude-opus-5-5", System: "SYS",
		Messages: []Message{
			{Role: "user", Parts: []Part{{Kind: Text, Text: "OLD"}}},
			{Role: "assistant", Parts: []Part{{Kind: Text, Text: "ANSWER"}}},
			{Role: "user", Parts: []Part{{Kind: Text, Text: "NEW"}}},
		},
	}}
	before, err := tuiRuntimeRenderInteractiveHistoricalBootstrap(textOnly)
	if err != nil {
		t.Fatal(err)
	}
	after, _, err := tuiRuntimeRenderInteractiveHistoricalBootstrapWithImages(textOnly)
	if err != nil || after != before {
		t.Fatalf("text-only bootstrap changed: err=%v same=%v", err, after == before)
	}
	// UTF-8 budget still enforced with an image in the current request.
	big := tuiRuntimeBootstrapImageRequest()
	big.Messages[0].Parts = []Part{{Kind: Text, Text: strings.Repeat("中", 220_000)}}
	prompt, _, err := tuiRuntimeRenderInteractiveHistoricalBootstrapWithImages(big)
	if err != nil {
		t.Fatal(err)
	}
	if len(prompt) > tuiRuntimeInteractiveHistoricalBootstrapMaxBytes {
		t.Fatalf("bootstrap is %d bytes, limit %d", len(prompt), tuiRuntimeInteractiveHistoricalBootstrapMaxBytes)
	}
	if !strings.Contains(prompt, "[Earlier historical text was omitted") {
		t.Fatal("trim marker missing")
	}
}
