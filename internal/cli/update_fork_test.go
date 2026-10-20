package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/yetone/magpie/internal/settings"
	"github.com/yetone/magpie/internal/update"
)

func TestUpdateCmdDisabledInFork(t *testing.T) {
	os.Unsetenv("MAGPIE_UPDATE_ALLOW_OFFICIAL")
	var buf bytes.Buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := updateCmd([]string{"update"})
	w.Close()
	os.Stdout = old
	io.Copy(&buf, r)
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("BB-fat/magpie")) {
		t.Fatalf("update output = %q, want fork releases message", out)
	}
}

func TestUpdateAutoOnRejectedInFork(t *testing.T) {
	os.Unsetenv("MAGPIE_UPDATE_ALLOW_OFFICIAL")
	groupsHome(t)
	if err := updateCmd([]string{"update", "auto", "on"}); err == nil {
		t.Fatal("update auto on = nil, want error")
	}
}

func TestUpdateAutoOnCannotPersistInFork(t *testing.T) {
	os.Unsetenv("MAGPIE_UPDATE_ALLOW_OFFICIAL")
	groupsHome(t)
	if err := os.MkdirAll(settings.Dir(), 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"noAutoUpdate":false}`
	if err := os.WriteFile(settings.Path(), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := settings.Save(settings.Load()); err != nil {
		t.Fatal(err)
	}
	if !settings.Load().NoAutoUpdate {
		t.Fatal("NoAutoUpdate still false after Save in fork")
	}
	if !update.SelfUpdateDisabled() {
		t.Fatal("SelfUpdateDisabled = false in fork test")
	}
}
