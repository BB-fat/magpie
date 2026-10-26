//go:build !darwin && !linux

package gateway

import (
	"errors"
	"os"
)

func tuiRuntimeAcquireInteractiveNativeLease(string, string) (*os.File, error) {
	return nil, errors.New("native session fencing requires an OS-level process lease")
}
func tuiRuntimeReleaseInteractiveNativeLease(f *os.File) {
	if f != nil {
		_ = f.Close()
	}
}
