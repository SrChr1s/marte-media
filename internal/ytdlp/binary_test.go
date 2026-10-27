package ytdlp

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func touch(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestLookupOrder(t *testing.T) {
	cfg, bundle := t.TempDir(), t.TempDir()
	old := bundledDirs
	bundledDirs = func() []string { return []string{bundle} }
	t.Cleanup(func() { bundledDirs = old })
	t.Setenv("PATH", t.TempDir()) // nothing on PATH

	b := NewBinaries(cfg)
	if _, err := b.YtDlp(); err == nil {
		t.Fatal("expected not installed")
	}
	if b.FFmpeg() != "" {
		t.Fatal("expected no ffmpeg")
	}

	bundledYt := filepath.Join(bundle, exeName("yt-dlp"))
	touch(t, bundledYt)
	touch(t, filepath.Join(bundle, exeName("ffmpeg")))
	if p, _ := b.YtDlp(); p != bundledYt {
		t.Fatalf("bundled copy should be found, got %q", p)
	}
	if b.FFmpeg() != filepath.Join(bundle, exeName("ffmpeg")) {
		t.Fatalf("bundled ffmpeg should be found, got %q", b.FFmpeg())
	}

	touch(t, b.managedPath()) // an app-updated copy must win over the bundled one
	if p, _ := b.YtDlp(); p != b.managedPath() {
		t.Fatalf("managed copy should win, got %q", p)
	}
}

func TestReleaseAssetKnownPlatform(t *testing.T) {
	a, err := releaseAsset()
	if err != nil {
		t.Skipf("unsupported platform %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	if runtime.GOOS == "windows" && filepath.Ext(a) != ".exe" {
		t.Fatalf("windows asset must be .exe, got %s", a)
	}
}

func TestJSRuntimeSelection(t *testing.T) {
	cfg, bundle, pathDir := t.TempDir(), t.TempDir(), t.TempDir()
	old := bundledDirs
	bundledDirs = func() []string { return []string{bundle} }
	t.Cleanup(func() { bundledDirs = old })
	t.Setenv("PATH", pathDir)
	b := NewBinaries(cfg)

	if n, _ := b.JSRuntime(); n != "" {
		t.Fatalf("no runtime expected, got %q", n)
	}
	if args := b.jsArgs(); len(args) != 0 {
		t.Fatalf("no args expected, got %v", args)
	}

	// Only node on PATH (this is the situation that broke YouTube extraction before).
	touch(t, filepath.Join(pathDir, exeName("node")))
	if n, _ := b.JSRuntime(); n != "node" {
		t.Fatalf("want node, got %q", n)
	}

	// A bundled deno wins and is passed with its full path, ahead of the system ones.
	bundled := filepath.Join(bundle, exeName("deno"))
	touch(t, bundled)
	if n, p := b.JSRuntime(); n != "deno" || p != bundled {
		t.Fatalf("want bundled deno, got %q %q", n, p)
	}
	args := b.jsArgs()
	if len(args) != 4 || args[0] != "--js-runtimes" || args[1] != "deno:"+bundled || args[3][:5] != "node:" {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestEnvNeedsYtDlp(t *testing.T) {
	old := bundledDirs
	bundledDirs = func() []string { return nil }
	t.Cleanup(func() { bundledDirs = old })
	t.Setenv("PATH", t.TempDir())
	if _, err := NewBinaries(t.TempDir()).Env(); err == nil {
		t.Fatal("expected an error without yt-dlp")
	}
}
