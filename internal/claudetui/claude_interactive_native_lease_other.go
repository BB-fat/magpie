//go:build !darwin && !linux

package claudetui

import (
	"errors"
	"os"
)

func acquireInteractiveNativeLease(string, string) (*os.File, error) {
	return nil, errors.New("native session fencing requires an OS-level process lease")
}
func releaseInteractiveNativeLease(f *os.File) {
	if f != nil {
		_ = f.Close()
	}
}
