package forkbuild

import "os"

// CustomFork is true in BB-fat/magpie builds.
const CustomFork = true

// ReleasesURL is where this fork's releases are published.
const ReleasesURL = "https://github.com/BB-fat/magpie/releases"

const allowOfficialEnv = "MAGPIE_UPDATE_ALLOW_OFFICIAL"

// SelfUpdateDisabled reports whether built-in self-update must not run.
func SelfUpdateDisabled() bool {
	if !CustomFork {
		return false
	}
	return os.Getenv(allowOfficialEnv) == ""
}

// AllowOfficialEnv is the test-only escape hatch env var name.
const AllowOfficialEnv = allowOfficialEnv
