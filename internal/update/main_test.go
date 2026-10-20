package update

import (
	"os"
	"testing"

	"github.com/yetone/magpie/internal/forkbuild"
	"github.com/yetone/magpie/internal/testenv"
)

// The tests run in a home of their own, never the user's real files.
func TestMain(m *testing.M) {
	os.Setenv(forkbuild.AllowOfficialEnv, "1")
	testenv.Main(m)
}
