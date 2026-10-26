package plugin

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// BunPath finds magpie's own Bun as it is: never downloads one, and leaves
// one downloaded before under its old name (Bun alone renames it).
func TestBunPathNeverDownloadsOrRenames(t *testing.T) {
	bunHome(t, BunVersion, time.Now())
	asked := bunReleases(t, []byte("would be downloaded"))
	if _, err := BunPath(BunVersion); err == nil || !strings.Contains(err.Error(), "MAGPIE_BUN") {
		t.Fatalf("no Bun: %v, want it said how to get one", err)
	}
	old := filepath.Join(bunDirOf(BunVersion), bunExe())
	if err := os.MkdirAll(filepath.Dir(old), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(old, []byte("downloaded before"), 0o755); err != nil {
		t.Fatal(err)
	}
	exe, err := BunPath(BunVersion)
	if err != nil || exe != old {
		t.Fatalf("BunPath = %s, %v, want the one downloaded before at %s", exe, err, old)
	}
	if _, err := os.Stat(bunExeOf(BunVersion)); !os.IsNotExist(err) {
		t.Fatal("BunPath renamed the Bun downloaded before")
	}
	if err := os.WriteFile(bunExeOf(BunVersion), []byte("magpie-bun"), 0o755); err != nil {
		t.Fatal(err)
	}
	if exe, err := BunPath(BunVersion); err != nil || exe != bunExeOf(BunVersion) {
		t.Fatalf("BunPath = %s, %v, want magpie-bun", exe, err)
	}
	if asked.Load() != 0 {
		t.Fatalf("BunPath downloaded %d times", asked.Load())
	}
}

// $MAGPIE_BUN is taken when it is new enough, and refused, saying why,
// when it is too old, isn't a program or won't say its version.
func TestBunPathTakesMagpieBun(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a shell script stands in for Bun")
	}
	bunHome(t, BunVersion, time.Now())
	dir := t.TempDir()
	bun := func(name, version string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("#!/bin/sh\necho "+version+"\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		return p
	}
	t.Setenv("MAGPIE_BUN", bun("new", "1.4.2"))
	if exe, err := BunPath("1.1.0"); err != nil || exe != filepath.Join(dir, "new") {
		t.Fatalf("BunPath = %s, %v", exe, err)
	}
	t.Setenv("MAGPIE_BUN", bun("old", "1.0.36"))
	if _, err := BunPath("1.1.0"); err == nil || !strings.Contains(err.Error(), "1.1.0 or newer") {
		t.Fatalf("an old $MAGPIE_BUN: %v", err)
	}
	t.Setenv("MAGPIE_BUN", bun("mute", ""))
	if _, err := BunPath("1.1.0"); err == nil || !strings.Contains(err.Error(), "doesn't say its version") {
		t.Fatalf("a $MAGPIE_BUN that says no version: %v", err)
	}
	t.Setenv("MAGPIE_BUN", filepath.Join(dir, "missing"))
	if _, err := BunPath("1.1.0"); err == nil || !strings.Contains(err.Error(), "not a program") {
		t.Fatalf("a missing $MAGPIE_BUN: %v", err)
	}
}
