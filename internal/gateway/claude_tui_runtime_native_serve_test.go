package gateway

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/access"
)

// A serve run's wrapper here is this test binary again
// (TestClaudeTUIRuntimeServeFakeWrapper): it speaks the wrapper's control
// protocol, answers each turn "reply <n> from <pid>", and notes each turn
// it was given, with its pid, in MAGPIE_TUI_SERVE_FAKE.
const tuiRuntimeServeModel = "claude-sonnet-5-5"

type tuiRuntimeServeTurn struct {
	PID   int    `json:"pid"`
	Turn  int    `json:"turn"`
	Input string `json:"input"`
}

func TestClaudeTUIRuntimeServeFakeWrapper(t *testing.T) {
	logPath := os.Getenv("MAGPIE_TUI_SERVE_FAKE")
	if logPath == "" {
		t.Skip("the serve tests run it as their wrapper")
	}
	tuiRuntimeServeFake(logPath)
	os.Exit(0)
}

func tuiRuntimeServeFake(logPath string) {
	out := bufio.NewWriter(os.Stdout)
	say := func(v any) {
		b, _ := json.Marshal(v)
		out.Write(append(b, '\n'))
		out.Flush()
	}
	pid := os.Getpid()
	sid := fmt.Sprintf("fake-%d", pid)
	say(map[string]any{"type": "bridge_ready", "turns": 0})
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 64<<10), 64<<20)
	turns := 0
	for in.Scan() {
		var c struct{ Op, Segment, Input string }
		if json.Unmarshal(in.Bytes(), &c) != nil || c.Op != "submit" {
			return
		}
		turns++
		if f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600); err == nil {
			b, _ := json.Marshal(tuiRuntimeServeTurn{PID: pid, Turn: turns, Input: c.Input})
			f.Write(append(b, '\n'))
			f.Close()
		}
		say(map[string]any{"type": "bridge_accepted", "segment": c.Segment})
		if strings.Contains(c.Input, "HOLD") {
			select {} // the turn never ends: its client goes, or the test ends it
		}
		reply := fmt.Sprintf("reply %d from %d", turns, pid)
		say(map[string]any{"type": "assistant", "uuid": fmt.Sprintf("%s-%d", sid, turns), "session_id": sid,
			"message": map[string]any{"id": fmt.Sprintf("m%d", turns), "model": tuiRuntimeServeModel,
				"usage":   map[string]any{"input_tokens": 1, "output_tokens": 2},
				"content": []any{map[string]any{"type": "text", "text": reply}}}})
		say(map[string]any{"type": "result", "session_id": sid, "stop_reason": "end_turn", "result": reply, "is_error": false})
		say(map[string]any{"type": "bridge_ready", "turns": turns})
	}
}

// tuiRuntimeServeTest readies a bridge whose runs keep their process ttl
// for the next turn, with the fake wrapper; it gives the bridge, the
// account folder runs are in and the fake's turn log.
func tuiRuntimeServeTest(t *testing.T, ttl time.Duration) (*tuiRuntimeBridge, string, string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the Claude TUI runs on Linux")
	}
	logPath := filepath.Join(t.TempDir(), "turns.jsonl")
	t.Setenv("MAGPIE_TUI_SERVE_FAKE", logPath)
	old := tuiRuntimeSettings
	t.Cleanup(func() { tuiRuntimeSettings = old })
	var mu sync.Mutex
	initTUIRuntime(&tuiRuntimeConfig{
		StateDir:   t.TempDir(),
		HelperArgv: []string{"/bin/false"}, // the fake makes no MCP calls
		IdleTTL: func() time.Duration {
			mu.Lock()
			defer mu.Unlock()
			return ttl
		},
	})
	tuiRuntimeSetWrapper(t, os.Args[0], "-test.run=^TestClaudeTUIRuntimeServeFakeWrapper$", "--")
	b := newTUIBridge()
	t.Cleanup(func() {
		b.Shutdown()
		deadline := time.Now().Add(30 * time.Second)
		for b.Busy() && time.Now().Before(deadline) {
			time.Sleep(20 * time.Millisecond)
		}
	})
	return b, t.TempDir(), logPath
}

func tuiRuntimeServeTurns(t *testing.T, logPath string) []tuiRuntimeServeTurn {
	t.Helper()
	data, _ := os.ReadFile(logPath)
	var turns []tuiRuntimeServeTurn
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var turn tuiRuntimeServeTurn
		if err := json.Unmarshal([]byte(line), &turn); err != nil {
			t.Fatalf("fake's log: %v", err)
		}
		turns = append(turns, turn)
	}
	return turns
}

// tuiRuntimeServeAsk serves body as caller key on b, for outer.
func tuiRuntimeServeAsk(t *testing.T, ctx context.Context, b *tuiRuntimeBridge, profile, outer, key, body string) (*tuiRuntimeRecorder, int, string) {
	t.Helper()
	out := &tuiRuntimeRecorder{}
	r := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
	r = r.WithContext(access.WithIdentity(ctx, access.Identity{KeyID: key}))
	code, msg := b.Serve(r, out, tuiRuntimeServeModel, tuiRuntimeOwner("claude", "serve@example.com", tuiRuntimeSettings.MCPServer), outer, profile, []byte(body))
	return out, code, msg
}

// tuiRuntimeServeIdle waits for n processes to be idle in b's pool.
func tuiRuntimeServeIdle(t *testing.T, b *tuiRuntimeBridge, n int) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for {
		got := tuiRuntimeServeIdleCount(b)
		if got == n {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("%d processes idle, want %d", got, n)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func tuiRuntimeServeIdleCount(b *tuiRuntimeBridge) int {
	br := b.rt.bridge
	br.mu.Lock()
	defer br.mu.Unlock()
	n := 0
	for _, runs := range br.hot {
		n += len(runs)
	}
	return n
}

func tuiRuntimeServeBody(messages ...string) string {
	var ms []string
	for i, text := range messages {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		ms = append(ms, fmt.Sprintf(`{"role":%q,"content":%s}`, role, tuiRuntimeJsonString(text)))
	}
	return `{"model":"` + tuiRuntimeServeModel + `","max_tokens":64,"stream":true,"messages":[` + strings.Join(ms, ",") + `]}`
}

// The turn after a complete reply goes on in the same process, with no
// session header: the request's own history, then the reply as it was told,
// is where the process's session ends.
func TestClaudeTUIRuntimeServeNextTurnSameProcess(t *testing.T) {
	for _, outer := range []string{"", "conv-serve"} {
		t.Run("outer="+outer, func(t *testing.T) {
			b, profile, logPath := tuiRuntimeServeTest(t, time.Minute)
			ctx := context.Background()
			first, code, msg := tuiRuntimeServeAsk(t, ctx, b, profile, outer, "", tuiRuntimeServeBody("hello"))
			if code != 200 {
				t.Fatalf("first turn: %d %s", code, msg)
			}
			tuiRuntimeServeIdle(t, b, 1)
			second, code, msg := tuiRuntimeServeAsk(t, ctx, b, profile, outer, "", tuiRuntimeServeBody("hello", first.text(), "and again"))
			if code != 200 {
				t.Fatalf("second turn: %d %s", code, msg)
			}
			turns := tuiRuntimeServeTurns(t, logPath)
			if len(turns) != 2 || turns[0].PID != turns[1].PID || turns[1].Turn != 2 {
				t.Fatalf("turns %+v: the second was not typed into the first's process", turns)
			}
			if !strings.Contains(turns[1].Input, "and again") || strings.Contains(turns[1].Input, "hello") {
				t.Fatalf("the second turn was given %q: only its new user turn goes in", turns[1].Input)
			}
			if want := fmt.Sprintf("reply 2 from %d", turns[0].PID); second.text() != want {
				t.Fatalf("second reply %q, want %q", second.text(), want)
			}
			tuiRuntimeServeIdle(t, b, 1)
		})
	}
}

// Only the exact prefix goes on: a reply the client changed, another
// caller key on the same account, or a request whose launch would differ
// starts a process of its own, and the idle one stays for its own next turn.
func TestClaudeTUIRuntimeServeStrictPrefix(t *testing.T) {
	b, profile, logPath := tuiRuntimeServeTest(t, time.Minute)
	ctx := context.Background()
	first, code, msg := tuiRuntimeServeAsk(t, ctx, b, profile, "", "key-a", tuiRuntimeServeBody("hello"))
	if code != 200 {
		t.Fatalf("first turn: %d %s", code, msg)
	}
	tuiRuntimeServeIdle(t, b, 1)
	pid := tuiRuntimeServeTurns(t, logPath)[0].PID
	cases := []struct {
		name, key, body string
	}{
		{"reply edited", "key-a", tuiRuntimeServeBody("hello", first.text()+" ", "next")},
		{"earlier turn edited", "key-a", tuiRuntimeServeBody("hello!", first.text(), "next")},
		{"another caller key", "key-b", tuiRuntimeServeBody("hello", first.text(), "next")},
		{"tools offered", "key-a", strings.Replace(tuiRuntimeServeBody("hello", first.text(), "next"), `"messages"`, `"tools":`+tuiRuntimeRealTools+`,"messages"`, 1)},
		{"other instructions", "key-a", strings.Replace(tuiRuntimeServeBody("hello", first.text(), "next"), `"messages"`, `"system":"be brief","messages"`, 1)},
	}
	for _, c := range cases {
		before := len(tuiRuntimeServeTurns(t, logPath))
		if _, code, msg := tuiRuntimeServeAsk(t, ctx, b, profile, "", c.key, c.body); code != 200 {
			t.Fatalf("%s: %d %s", c.name, code, msg)
		}
		turns := tuiRuntimeServeTurns(t, logPath)
		if len(turns) != before+1 || turns[len(turns)-1].PID == pid || turns[len(turns)-1].Turn != 1 {
			t.Fatalf("%s: went on in the idle process: %+v", c.name, turns[before:])
		}
	}
	// the first process is still idle at its own prefix, and takes it
	before := len(tuiRuntimeServeTurns(t, logPath))
	if _, code, msg := tuiRuntimeServeAsk(t, ctx, b, profile, "", "key-a", tuiRuntimeServeBody("hello", first.text(), "next")); code != 200 {
		t.Fatalf("its own next turn: %d %s", code, msg)
	}
	if turns := tuiRuntimeServeTurns(t, logPath); turns[before].PID != pid || turns[before].Turn != 2 {
		t.Fatalf("its own next turn ran in %+v, want process %d's second", turns[before], pid)
	}
}

// A process idle past its TTL is ended, and the next turn is started anew;
// none is kept at all with a TTL of 0.
func TestClaudeTUIRuntimeServeIdleTTL(t *testing.T) {
	t.Run("expired", func(t *testing.T) {
		b, profile, logPath := tuiRuntimeServeTest(t, 300*time.Millisecond)
		ctx := context.Background()
		first, code, msg := tuiRuntimeServeAsk(t, ctx, b, profile, "", "", tuiRuntimeServeBody("hello"))
		if code != 200 {
			t.Fatalf("first turn: %d %s", code, msg)
		}
		deadline := time.Now().Add(30 * time.Second)
		for b.Busy() {
			if time.Now().After(deadline) {
				t.Fatal("the idle process outlived its TTL")
			}
			time.Sleep(20 * time.Millisecond)
		}
		if _, code, msg := tuiRuntimeServeAsk(t, ctx, b, profile, "", "", tuiRuntimeServeBody("hello", first.text(), "next")); code != 200 {
			t.Fatalf("after the TTL: %d %s", code, msg)
		}
		if turns := tuiRuntimeServeTurns(t, logPath); len(turns) != 2 || turns[1].PID == turns[0].PID {
			t.Fatalf("turns %+v: the turn after the TTL went on in the ended process", turns)
		}
	})
	t.Run("off", func(t *testing.T) {
		b, profile, _ := tuiRuntimeServeTest(t, 0)
		tuiRuntimeSetWrapper(t, tuiRuntimeFakeBridge(t)) // the one-turn wrapper: no serve run is asked for
		if _, code, msg := tuiRuntimeServeAsk(t, context.Background(), b, profile, "", "", tuiRuntimeServeBody("hello")); code != 200 {
			t.Fatalf("one-turn run: %d %s", code, msg)
		}
		if n := tuiRuntimeServeIdleCount(b); n != 0 {
			t.Fatalf("%d processes kept with the TTL off", n)
		}
	})
}

// A claim and the TTL racing: the wait the claim took is gone, so its
// timer ends nothing, even fired late; a later wait has its own.
func TestClaudeTUIRuntimeServeClaimBeatsTTL(t *testing.T) {
	b, profile, _ := tuiRuntimeServeTest(t, time.Minute)
	ctx := context.Background()
	if _, code, msg := tuiRuntimeServeAsk(t, ctx, b, profile, "", "", tuiRuntimeServeBody("hello")); code != 200 {
		t.Fatalf("first turn: %d %s", code, msg)
	}
	tuiRuntimeServeIdle(t, b, 1)
	br := b.rt.bridge
	var run *tuiRuntimeSubscriptionRun
	var tip tuiRuntimePrefixDigest
	br.mu.Lock()
	for d, runs := range br.hot {
		run, tip = runs[0], d
	}
	br.mu.Unlock()
	run.mu.Lock()
	gen, want := run.hotGen, run.hotConfig
	run.mu.Unlock()
	claimed, why := br.claimHot(tip, map[bool]tuiRuntimeHotConfig{false: want, true: want}, false)
	if claimed != run {
		t.Fatalf("claim: %v %s", claimed, why)
	}
	if run.letGoIdle(gen, "a late timer") {
		t.Fatal("a timer of the wait the claim took ended the process")
	}
	run.mu.Lock()
	closed, left := run.closed, run.left
	run.mu.Unlock()
	if closed || left {
		t.Fatal("the claimed process is ending")
	}
	if again, _ := br.claimHot(tip, map[bool]tuiRuntimeHotConfig{false: want, true: want}, false); again != nil {
		t.Fatal("one idle process was claimed twice")
	}
}

// The client of a turn going late, after the process took the next
// request's turn, ends nothing; the client of the turn in flight going
// ends it, and nothing of it is kept.
func TestClaudeTUIRuntimeServeLateCancel(t *testing.T) {
	b, profile, logPath := tuiRuntimeServeTest(t, time.Minute)
	ctx := context.Background()
	first, code, msg := tuiRuntimeServeAsk(t, ctx, b, profile, "", "", tuiRuntimeServeBody("hello"))
	if code != 200 {
		t.Fatalf("first turn: %d %s", code, msg)
	}
	tuiRuntimeServeIdle(t, b, 1)
	ctx2, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		tuiRuntimeServeAsk(t, ctx2, b, profile, "", "", tuiRuntimeServeBody("hello", first.text(), "HOLD on"))
	}()
	var run *tuiRuntimeSubscriptionRun
	deadline := time.Now().Add(30 * time.Second)
	for run == nil {
		for _, r := range b.rt.bridge.allRuns() {
			if r.turn() == 2 {
				run = r
			}
		}
		if time.Now().After(deadline) {
			t.Fatal("the second turn never started")
		}
		time.Sleep(10 * time.Millisecond)
	}
	for len(tuiRuntimeServeTurns(t, logPath)) < 2 {
		time.Sleep(10 * time.Millisecond)
	}
	// turn 1's client going now is not turn 2's
	run.letGoServeTurn(nil, 1)
	run.mu.Lock()
	closed, left := run.closed, run.left
	run.mu.Unlock()
	if closed || left {
		t.Fatal("a late cancellation of the first turn ended the second")
	}
	cancel()
	<-done
	deadline = time.Now().Add(30 * time.Second)
	for b.Busy() {
		if time.Now().After(deadline) {
			t.Fatal("the turn whose client went was not ended")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if n := tuiRuntimeServeIdleCount(b); n != 0 {
		t.Fatalf("%d processes kept after a cancelled turn", n)
	}
}

// Idle processes are bounded and go first for room: past
// tuiRuntimeServeIdleMost the one idle longest ends, and claudeRoom's
// let-go ends an idle one.
func TestClaudeTUIRuntimeServeIdleBounds(t *testing.T) {
	b, profile, _ := tuiRuntimeServeTest(t, time.Minute)
	old := tuiRuntimeServeIdleMost
	tuiRuntimeServeIdleMost = 1
	t.Cleanup(func() { tuiRuntimeServeIdleMost = old })
	ctx := context.Background()
	for _, text := range []string{"one", "two"} {
		if _, code, msg := tuiRuntimeServeAsk(t, ctx, b, profile, "", "", tuiRuntimeServeBody(text)); code != 200 {
			t.Fatalf("%s: %d %s", text, code, msg)
		}
	}
	tuiRuntimeServeIdle(t, b, 1)
	if !tuiRuntimeLetGoIdleForRoom(1) {
		t.Fatal("the room found no idle process to end")
	}
	tuiRuntimeServeIdle(t, b, 0)
	if tuiRuntimeLetGoIdleForRoom(1) {
		t.Fatal("the room ended a process with none idle")
	}
}
