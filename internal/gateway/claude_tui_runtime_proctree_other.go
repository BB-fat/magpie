//go:build !linux

package gateway

import "os/exec"

// The TUI bridge runs on Linux only (Supported); elsewhere the tree is
// the wrapper's group alone.

func tuiRuntimeEndWithParent(*exec.Cmd) {}

func tuiRuntimeDescendants(int) map[int]uint64 { return nil }

func tuiRuntimeGroupRSS(map[int]uint64) (int64, bool) { return 0, false }

func tuiRuntimeKillSeen(map[int]uint64) {}
