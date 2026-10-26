//go:build !linux

package claudetui

import "os/exec"

// The TUI bridge runs on Linux only (Supported); elsewhere the tree is
// the wrapper's group alone.

func endWithParent(*exec.Cmd) {}

func descendants(int) map[int]uint64 { return nil }

func groupRSS(map[int]uint64) (int64, bool) { return 0, false }

func killSeen(map[int]uint64) {}
