//go:build !windows

package gateway

import (
	"os"
	"strconv"
	"syscall"
)

// ownedByMe says the file belongs to the user the runtime runs as.
func tuiRuntimeOwnedByMe(fi os.FileInfo) bool {
	st, ok := fi.Sys().(*syscall.Stat_t)
	return ok && int(st.Uid) == os.Getuid()
}

// othersCanWrite says the group or anyone else may write in the folder.
func tuiRuntimeOthersCanWrite(fi os.FileInfo) bool { return fi.Mode().Perm()&0o022 != 0 }

// claudeWorkName is the work folder's name in the temp folder, which on
// Linux every user shares.
func tuiRuntimeClaudeWorkName() string { return "magpie-claude-work-" + strconv.Itoa(os.Getuid()) }
