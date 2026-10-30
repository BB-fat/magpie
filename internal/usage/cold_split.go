package usage

import (
	"sync"
	"time"
)

// A first start, with no sums kept on disk, built the page's cells, the
// heatmap's days and Codex's effort before the cut from one pairing of
// the whole history: every gateway call priced, every session row's
// message id in one set, every call matched in one map, ~210 MB of a
// ~355 MB live heap and a ~500 MB footprint on a year's history (#2278).
// They are built as a day's start advances them (pageAdvance), a few
// weeks of history at a time: a span's chunks are those with a row a
// pairing of it reaches, and each span's pairing is let go before the
// next.

// coldSpan is the days of history a first build pairs at a time. Zero
// pairs them all at once (tests compare the two).
var coldSpan = 14

// coldSpanning, in tests, is told how many chunks each span pairs.
var coldSpanning func(chunks int)

// coldSpans calls fn for each span of days from since, or the day of the
// oldest row of sets, to cut, with the chunks of sets a pairing of that
// span reaches: those with a row from heatReach before it to heatReach
// after. A gateway block not priced yet is priced for its span alone, and
// let go after it. False: coldSpan is zero, and fn was not called.
func coldSpans(since, cut time.Time, sets [3][]*rowChunk, fn func(from, to time.Time, span [3][]*rowChunk)) bool {
	if coldSpan <= 0 {
		return false
	}
	first := chunkFirsts(sets)
	oldest := time.Time{}
	for _, set := range sets {
		for _, c := range set {
			if at := first(c); !at.IsZero() && (oldest.IsZero() || at.Before(oldest)) {
				oldest = at
			}
		}
	}
	from := since
	if oldest.IsZero() {
		from = cut
	} else if day := time.Date(oldest.Year(), oldest.Month(), oldest.Day(), 0, 0, 0, 0, cut.Location()); from.Before(day) {
		from = day
	}
	for from.Before(cut) {
		to := time.Date(from.Year(), from.Month(), from.Day()+coldSpan, 0, 0, 0, 0, cut.Location())
		if to.After(cut) {
			to = cut
		}
		lo, hi := from.Add(-heatReach), to.Add(heatReach)
		var span [3][]*rowChunk
		alone := map[*blockPricer]*blockPricer{}
		for k, set := range sets {
			for _, c := range set {
				at := first(c)
				if at.IsZero() || c.Latest.Before(lo) || !at.Before(hi) {
					continue
				}
				if b := c.lazy; b != nil {
					// priced into a copy: the query's stub stays one, and
					// the span's priced rows go back to the pool after it
					if alone[b] == nil {
						alone[b] = b.alone()
					}
					copy := *c
					copy.lazy = alone[b]
					c = &copy
				}
				span[k] = append(span[k], c)
			}
		}
		if coldSpanning != nil {
			coldSpanning(chunksIn(span))
		}
		fn(from, to, span)
		for _, b := range alone {
			b.release()
		}
		from = to
	}
	return true
}

// alone is a pricer of b's prices that keeps the blocks it prices to
// itself, until release.
func (b *blockPricer) alone() *blockPricer {
	o := *b
	o.shared, o.held, o.pooled = false, nil, nil
	return &o
}

// chunkFirsts is the time of each chunk's oldest row, zero for one with
// none. A named chunk's is kept by its serial: the page, the heatmap and
// Codex's effort each ask it of every chunk.
func chunkFirsts(sets [3][]*rowChunk) func(*rowChunk) time.Time {
	firsts.Lock()
	firsts.turn++
	turn := firsts.turn
	if firsts.by == nil {
		firsts.by = map[uint64]chunkFirst{}
	}
	for k, kept := range firsts.by {
		if kept.turn+hintsKept < turn {
			delete(firsts.by, k)
		}
	}
	firsts.Unlock()
	rows := lentRows()
	known := map[*rowChunk]time.Time{}
	for _, set := range sets {
		for _, c := range set {
			if c.Latest.IsZero() {
				continue
			}
			if c.serial != 0 {
				firsts.Lock()
				kept, ok := firsts.by[c.serial]
				if ok {
					kept.turn = turn
					firsts.by[c.serial] = kept
				}
				firsts.Unlock()
				if ok {
					known[c] = kept.at
					continue
				}
			}
			read := c
			if c.lazy != nil {
				read = c.raw // the block's times are its priced copy's
			}
			var at time.Time
			open := rows.of(read)
			for i := range open.Rows {
				if t := open.Rows[i].Time; at.IsZero() || t.Before(at) {
					at = t
				}
			}
			known[c] = at
			if c.serial != 0 {
				firsts.Lock()
				firsts.by[c.serial] = chunkFirst{at: at, turn: turn}
				firsts.Unlock()
			}
		}
	}
	rows.release()
	return func(c *rowChunk) time.Time { return known[c] }
}

type chunkFirst struct {
	at   time.Time
	turn uint64
}

var firsts struct {
	sync.Mutex
	by   map[uint64]chunkFirst
	turn uint64
}
