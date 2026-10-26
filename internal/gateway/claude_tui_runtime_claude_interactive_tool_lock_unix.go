//go:build darwin || linux

package gateway

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// The tool ledger's individual records and one-shot recovery claims need a
// writer lock that survives multiple Runtime instances using the same state
// directory. Do not remove or replace this lock file while the ledger exists.
func tuiRuntimeLockInteractiveToolJournal(path string) (func(), error) {
	lockPath := path + ".lock"
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o700); err != nil {
		return nil, err
	}
	if info, err := os.Lstat(lockPath); err == nil {
		if !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
			return nil, errors.New("unsafe tool journal lock file")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if info, err := f.Stat(); err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		f.Close()
		return nil, errors.New("unsafe opened tool journal lock")
	}
	deadline := time.Now().Add(3 * time.Second)
	for {
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN); _ = f.Close() }, nil
		}
		if err != syscall.EWOULDBLOCK && err != syscall.EAGAIN {
			_ = f.Close()
			return nil, err
		}
		if time.Now().After(deadline) {
			_ = f.Close()
			return nil, errors.New("tool journal is locked by another process; refusing unsafe continuation")
		}
		time.Sleep(15 * time.Millisecond)
	}
}
