package update

import (
	"context"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/forkbuild"
)

func TestSelfUpdateDisabledInCustomFork(t *testing.T) {
	t.Setenv(forkbuild.AllowOfficialEnv, "")
	if !SelfUpdateDisabled() {
		t.Fatal("SelfUpdateDisabled = false, want true without test escape env")
	}
	if _, err := Latest(context.Background()); err != ErrSelfUpdateDisabled {
		t.Fatalf("Latest = %v, want ErrSelfUpdateDisabled", err)
	}
}

func TestSelfUpdateAllowEnvForTests(t *testing.T) {
	t.Setenv(forkbuild.AllowOfficialEnv, "1")
	if SelfUpdateDisabled() {
		t.Fatal("SelfUpdateDisabled = true with MAGPIE_UPDATE_ALLOW_OFFICIAL")
	}
}

func TestDisabledMessageMentionsForkReleases(t *testing.T) {
	if msg := DisabledMessage(); msg == "" || !strings.Contains(msg, "BB-fat/magpie") {
		t.Fatalf("DisabledMessage = %q", msg)
	}
}
