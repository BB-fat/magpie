package gateway

import (
	"context"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Done is evaluated by the request gate only when the retry actually waits.
// This observes the real blocking boundary without a sleep or production hook.
type tuiRuntimeRetryWaitContext struct {
	context.Context
	waiting chan struct{}
	once    sync.Once
}

func (c *tuiRuntimeRetryWaitContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.waiting) })
	return c.Context.Done()
}

func TestClaudeTUIRuntimeQueuedRetryRechecksCompletedReply(t *testing.T) {
	_, runs := tuiRuntimeScWrapper(t)
	tuiRuntimeWithConfig(t, &tuiRuntimeConfig{StateDir: t.TempDir()})
	rt, b := tuiRuntimeScRuntime(t)
	firstAtRoom, allowFirst, secondWaiting := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var roomCalls atomic.Int32

	tuiRuntimeWithConfig(t, &tuiRuntimeConfig{Room: func(context.Context, string) (func(bool), error) {
		if roomCalls.Add(1) == 1 {
			close(firstAtRoom)
			<-allowFirst
		}
		return func(bool) {}, nil
	}})
	out := make(chan int, 2)
	body := tuiRuntimeReplayBody("")
	ask := func(watch bool) {
		r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
		if watch {
			r = r.WithContext(&tuiRuntimeRetryWaitContext{Context: r.Context(), waiting: secondWaiting})
		}
		rec := httptest.NewRecorder()
		code, _ := rt.serveMessages(rec, r, "claude-opus-5-5", tuiRuntimeScOuter, []byte(body), &Usage{})
		out <- code
	}
	go ask(false)
	select {
	case <-firstAtRoom:
	case <-time.After(10 * time.Second):
		t.Fatal("first request did not reach Room")
	}
	go ask(true)
	select {
	case <-secondWaiting:
	case <-time.After(10 * time.Second):
		close(allowFirst)
		t.Fatal("retry did not wait on the first request")
	}
	close(allowFirst)
	for i := 0; i < 2; i++ {
		select {
		case code := <-out:
			if code != 200 {
				t.Errorf("request status %d", code)
			}
		case <-time.After(20 * time.Second):
			t.Fatal("requests did not finish")
		}
	}
	b.abortAll()
	if n := len(runs()); n != 1 {
		t.Fatalf("OneExecutionPerRetry violated: the queued identical request started %d native processes, want 1", n)
	}
}
