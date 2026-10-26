//go:build darwin || linux

package gateway

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
)

// An entire native wrapper lifetime owns a canonical outer conversation.
// The acquired file description is deliberately inherited by the wrapper
// through exec.Cmd.ExtraFiles. If Runtime crashes, the wrapper continues to
// hold this advisory lock until it exits; a different Runtime must refuse
// another native writer while that wrapper is running.
//
// The lock alone cannot prove detached grandchildren died after an abnormal
// wrapper exit; callers must still require a dirty checkpoint, complete
// authenticated tool evidence and read-only recovery (never side effects).
func tuiRuntimeAcquireInteractiveNativeLease(owner, outer string) (*os.File, error) {
	if owner == "" || outer == "" {
		return nil, errors.New("native lease requires a canonical owner and outer session")
	}
	dir := filepath.Join(tuiRuntimeMustStateDir(), "claude-interactive-native-leases")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, tuiRuntimeInteractiveSessionStateKey(owner, outer)+".lock")
	if info, err := os.Lstat(path); err == nil {
		if !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
			return nil, errors.New("unsafe native lease file")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	if info, err := f.Stat(); err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		f.Close()
		return nil, errors.New("unsafe opened native lease file")
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		if err == syscall.EWOULDBLOCK || err == syscall.EAGAIN {
			return nil, tuiRuntimeErrInteractiveOuterLive
		}
		return nil, err
	}
	return f, nil
}

func tuiRuntimeReleaseInteractiveNativeLease(f *os.File) {
	if f != nil {
		_ = f.Close()
	} // No LOCK_UN: a surviving wrapper owns this same open file description.
}
