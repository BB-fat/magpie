package claudetui

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInteractiveSyntheticAPIErrorIsFailure(t *testing.T) {
	for _, readOnly := range []bool{false, true} {
		name := "ordinary"
		if readOnly {
			name = "read_only"
		}
		t.Run(name, func(t *testing.T) {
			run := &subscriptionRun{interactive: true, readOnlyRecovery: readOnly}
			events := run.attach()
			go run.readInteractiveOneTurn(strings.NewReader(`{"type":"assistant","message":{"id":"native-error","model":"<synthetic>","content":[{"type":"text","text":"API Error: upstream rejected this request. Details: reasoning_extraction"}]}}` + "\n" + `{"type":"result","is_error":false,"stop_reason":"end_turn","result":"API Error: upstream rejected this request"}` + "\n"))
			w := httptest.NewRecorder()
			endedOK := false
			endedCalls := 0
			status, _ := relay(w, httptest.NewRequest("POST", "/v1/messages", nil), protoAnthropic, "Claude Code", &Request{}, events, &Usage{}, func() {}, func(_, _ string, ok bool) { endedCalls++; endedOK = ok })
			if status != 502 || w.Code != 502 || endedOK {
				t.Fatalf("synthetic upstream failure reported as success: status=%d http=%d ended=%d ok=%t body=%s", status, w.Code, endedCalls, endedOK, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), "upstream") {
				t.Fatalf("lost upstream failure: %s", w.Body.String())
			}
		})
	}
}

func TestInteractiveNormalAPIErrorQuotationRemainsText(t *testing.T) {
	run := &subscriptionRun{interactive: true}
	events := run.attach()
	go run.readInteractiveOneTurn(strings.NewReader(`{"type":"assistant","message":{"id":"normal","model":"claude-opus-5-5","content":[{"type":"text","text":"API Error: is the text you asked me to explain."}]}}` + "\n" + `{"type":"result","is_error":false,"stop_reason":"end_turn"}` + "\n"))
	var text string
	var failed bool
	for e := range events {
		if e.Kind == KText {
			text += e.Text
		}
		if e.Kind == KError {
			failed = true
		}
	}
	if failed || text == "" {
		t.Fatalf("ordinary model quotation rejected: failed=%t text=%q", failed, text)
	}
}

func TestInteractiveTypedAPIErrorKeepsNonRetryableStatus(t *testing.T) {
	run := &subscriptionRun{interactive: true}
	events := run.attach()
	go run.readInteractiveOneTurn(strings.NewReader(`{"type":"assistant","isApiErrorMessage":true,"error":"invalid_request","request_id":"fixture-request","message":{"model":"<synthetic>","content":[{"type":"text","text":"API Error: upstream refusal"}]}}` + "\n"))
	w := httptest.NewRecorder()
	status, _ := relay(w, httptest.NewRequest("POST", "/v1/messages", nil), protoAnthropic, "Claude Code", &Request{}, events, &Usage{}, func() {}, func(_, _ string, ok bool) {
		if ok {
			t.Error("API refusal committed as clean")
		}
	})
	if status != 400 {
		t.Fatalf("nonretryable upstream refusal became HTTP %d: %s", status, w.Body.String())
	}
}
