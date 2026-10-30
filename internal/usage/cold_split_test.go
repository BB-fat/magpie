package usage

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/sessions"
)

// pairAtOnce has a first build pair every chunk at once, for the tests
// of what a pairing of all of them is told apart from.
func pairAtOnce(t *testing.T) {
	span := coldSpan
	coldSpan = 0
	t.Cleanup(func() { coldSpan = span })
}

// A first start pairs the history a few weeks at a time, not at once
// (#2278), and builds the page, the heatmap and Codex's effort a pairing
// of every chunk does: a call whose gateway row is after a span's end is
// paired, and a span pairs only the chunks it reaches.
func TestFirstBuildPairsSpans(t *testing.T) {
	yesterday, _ := advanceHistory(t)
	// a call just before midnight, its gateway row just after
	at := time.Date(2026, 9, 23, 23, 59, 30, 0, time.Local)
	path := filepath.Join(sessions.CodexDir(), "sessions", at.Format("2006/01/02"), "rollout-"+at.Format("2006-01-02T15-04-05")+"-z.jsonl")
	writeCodexCall(t, path, "z", "sz", `"vscode"`, "resp-z", "m", "high", at, 300, 30)
	Append(Record{Time: at.Add(time.Minute), Agent: "codex", Provider: "openai", Model: "m", Input: 300, Output: 30, Status: 200, ResponseID: "resp-z", Session: "sz"})
	// a call through magpie alone, 30 days back
	Append(Record{Time: yesterday.AddDate(0, 0, -30), Agent: "claude-code", Provider: "anthropic", Model: "c", Input: 50, Output: 5, Status: 200})
	restart := func() {
		requestCache.Lock()
		requestCache.parts, requestCache.pageOld, requestCache.heatParts, requestCache.heatOld, requestCache.heats, requestCache.pairs, requestCache.pages = nil, nil, nil, nil, nil, nil, map[pageKey]RequestPage{}
		requestCache.Unlock()
		codexEffort.Lock()
		codexEffort.periods = map[Period]*effortEntry{}
		codexEffort.Unlock()
		codexEffortBuilding.Lock()
		effortOlds = map[Period]*effortOld{}
		codexEffortBuilding.Unlock()
		for _, path := range []string{pageDiskPath(), heatDiskPath(), effortDiskPath()} {
			os.Remove(path)
		}
	}
	restart()
	t.Cleanup(restart)
	type built struct {
		page    RequestPage
		heat    Heatmap
		effort  map[string][]sessions.ModelEffort
		written bool
	}
	build := func() built {
		var b built
		b.page = QueryPage(All, Filter{}, 0, 50)
		b.heat = HeatmapOf(Filter{})
		b.effort = CodexModelEffort(All)
		_, page := os.Stat(pageDiskPath())
		_, heat := os.Stat(heatDiskPath())
		_, effort := os.Stat(effortDiskPath())
		b.written = page == nil && heat == nil && effort == nil
		return b
	}
	paired := []int{}
	pageSplitting = func(n int) { paired = append(paired, n) }
	whole := build()
	pageSplitting = nil
	if !whole.written || len(paired) == 0 {
		t.Fatal("a pairing of every chunk wrote no sums")
	}
	all := paired[0]
	for _, days := range []int{1, 14} {
		coldSpan = days
		restart()
		spans := []int{}
		coldSpanning = func(n int) { spans = append(spans, n) }
		got := build()
		coldSpanning = nil
		if !got.written {
			t.Fatalf("%d days a span: no sums written", days)
		}
		if len(spans) < 2 {
			t.Fatalf("%d days a span: %d spans", days, len(spans))
		}
		pairedBelow(t, "a span", spans, 0, all)
		if !reflect.DeepEqual(got.page, whole.page) {
			t.Fatalf("%d days a span: page %+v, paired at once %+v", days, got.page, whole.page)
		}
		if !reflect.DeepEqual(got.heat, whole.heat) {
			t.Fatalf("%d days a span: heatmap %+v, paired at once %+v", days, got.heat, whole.heat)
		}
		if !reflect.DeepEqual(got.effort, whole.effort) {
			t.Fatalf("%d days a span: effort %+v, paired at once %+v", days, got.effort, whole.effort)
		}
	}
	if whole.page.Total != 9 {
		t.Fatalf("page: %d rows", whole.page.Total)
	}
}
