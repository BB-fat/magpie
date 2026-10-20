package update

import (
	"errors"

	"github.com/yetone/magpie/internal/forkbuild"
)

// CustomFork mirrors forkbuild.CustomFork for callers that already import update.
const CustomFork = forkbuild.CustomFork

// ForkReleasesURL is where this fork's releases are published.
const ForkReleasesURL = forkbuild.ReleasesURL

// ErrSelfUpdateDisabled is returned when a check or install would reach the
// official update path in a custom fork build.
var ErrSelfUpdateDisabled = errors.New("self-update is disabled in this BB-fat custom build")

// SelfUpdateDisabled reports whether magpie must not check, download, or
// install from the built-in update machinery.
func SelfUpdateDisabled() bool {
	return forkbuild.SelfUpdateDisabled()
}

// DisabledMessage tells the user where to get custom builds.
func DisabledMessage() string {
	return "Self-update is disabled in this BB-fat custom build. Install from " + ForkReleasesURL
}

func guardSelfUpdate() error {
	if SelfUpdateDisabled() {
		return ErrSelfUpdateDisabled
	}
	return nil
}
