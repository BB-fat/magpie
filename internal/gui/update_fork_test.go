package gui

import (
	"testing"
	"time"

	"github.com/yetone/magpie/internal/settings"
)

func TestUpdateDueNeverInFork(t *testing.T) {
	t.Setenv("MAGPIE_UPDATE_ALLOW_OFFICIAL", "")
	s := settings.Settings{UpdateEvery: 30}
	last := time.Now().Add(-time.Hour)
	if updateDue(s, last, time.Now()) {
		t.Fatal("updateDue true with automatic updates on in fork")
	}
}
