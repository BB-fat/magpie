package gateway

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

// rssEvery is how often a run's processes are measured.
var tuiRuntimeRssEvery = 2 * time.Second

// readRSS measures a run's tree; tests give their own.
var tuiRuntimeReadRSS = (*tuiRuntimeProcTree).rss

// watchRSS measures the wrapper, its tmux, Claude Code and its MCP helper
// every every until done, and ends the run with an error once they hold
// more than config.MaxRSS together. It is a measure, not a hard limit: a
// run can grow past it between two readings. A reading that fails is not
// known, and ends nothing.
func (r *tuiRuntimeSubscriptionRun) watchRSS(t *tuiRuntimeProcTree, done <-chan struct{}, every time.Duration) {
	tick := time.NewTicker(every)
	defer tick.Stop()
	unknown := false
	for {
		select {
		case <-done:
			return
		case <-tick.C:
		}
		n, ok := tuiRuntimeReadRSS(t)
		if !ok {
			if !unknown {
				log.Printf("Claude TUI run on %s: its memory could not be read (pid %d)", r.model, t.pid)
			}
			unknown = true
			continue
		}
		unknown = false
		limit := tuiRuntimeSettings.MaxRSS
		if limit <= 0 || n <= limit {
			continue
		}
		msg := fmt.Sprintf("Claude TUI run used %d MiB, over its limit of %d MiB (claudeTuiMaxRSSMiB); it was ended", n>>20, limit>>20)
		log.Printf("Claude TUI run on %s (pid %d): %s", r.model, t.pid, msg)
		r.emit(Event{Kind: KError, Status: http.StatusBadGateway, Code: "memory_limit", Text: msg})
		r.endSegment()
		r.abort()
		return
	}
}

// roomError is a run that found no room among the gateway's Claude Code
// processes: no fault of the account's, so it is answered 503.
type tuiRuntimeRoomError struct{ error }

func (e tuiRuntimeRoomError) Unwrap() error { return e.error }

// startStatus is the status a run that failed to start is answered with.
func tuiRuntimeStartStatus(err error) int {
	var room tuiRuntimeRoomError
	if errors.As(err, &room) {
		return http.StatusServiceUnavailable
	}
	return http.StatusBadGateway
}
