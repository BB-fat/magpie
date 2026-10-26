package claudetui

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// endWithParent has the wrapper asked to end (SIGTERM, which it tears its
// tmux down on) when the gateway that started it dies.
func endWithParent(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Pdeathsig = syscall.SIGTERM
}

// procStat is a process's parent and start time (clock ticks since boot),
// from /proc/<pid>/stat.
func procStat(pid int) (ppid int, start uint64, ok bool) {
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return 0, 0, false
	}
	// the command may hold spaces and parentheses: fields follow its last ")"
	s := string(b)
	i := strings.LastIndexByte(s, ')')
	if i < 0 {
		return 0, 0, false
	}
	f := strings.Fields(s[i+1:])
	// f[0] state, f[1] ppid, … f[19] starttime
	if len(f) < 20 {
		return 0, 0, false
	}
	p, err1 := strconv.Atoi(f[1])
	st, err2 := strconv.ParseUint(f[19], 10, 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return p, st, true
}

// descendants is every process under root now, each with its start time.
func descendants(root int) map[int]uint64 {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}
	children := map[int][]int{}
	starts := map[int]uint64{}
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		ppid, start, ok := procStat(pid)
		if !ok {
			continue
		}
		children[ppid] = append(children[ppid], pid)
		starts[pid] = start
	}
	out := map[int]uint64{}
	queue := []int{root}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, c := range children[p] {
			if _, done := out[c]; !done {
				out[c] = starts[c]
				queue = append(queue, c)
			}
		}
	}
	return out
}

// same says pid still runs as the process that started at start (0: the
// wrapper itself, whose pid the gateway holds until it is reaped).
func same(pid int, start uint64) bool {
	_, st, ok := procStat(pid)
	return ok && (start == 0 || st == start)
}

// groupRSS sums the resident memory of pids still running as they were
// seen; false when the wrapper's own could not be read.
func groupRSS(pids map[int]uint64) (int64, bool) {
	var total int64
	known := false
	for pid, start := range pids {
		if !same(pid, start) {
			continue
		}
		n, ok := pidRSS(pid)
		if !ok {
			if start == 0 {
				return 0, false
			}
			continue
		}
		total += n
		known = true
	}
	return total, known
}

func pidRSS(pid int) (int64, bool) {
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/status")
	if err != nil {
		return 0, false
	}
	for _, line := range strings.Split(string(b), "\n") {
		if f := strings.Fields(line); len(f) >= 2 && f[0] == "VmRSS:" {
			kb, err := strconv.ParseInt(f[1], 10, 64)
			if err != nil {
				return 0, false
			}
			return kb << 10, true
		}
	}
	// a kernel thread or a zombie holds none
	return 0, true
}

// killSeen kills each process still running as it was seen.
func killSeen(seen map[int]uint64) {
	for pid, start := range seen {
		if start != 0 && same(pid, start) {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	}
}
