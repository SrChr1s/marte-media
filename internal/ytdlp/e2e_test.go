//go:build e2e

package ytdlp

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Run with: go test -tags e2e ./internal/ytdlp -run E2E -v
// Needs network, and build/deps/<os>-amd64 filled by scripts/fetch-deps.sh. It checks that the app finds
// the bundled deno even when PATH has no JavaScript runtime, and that the result has the full format list.
func TestE2EBundledDenoProbe(t *testing.T) {
	deps, err := filepath.Abs("../../build/deps/linux-amd64")
	if err != nil {
		t.Fatal(err)
	}
	old := bundledDirs
	bundledDirs = func() []string { return []string{deps} }
	t.Cleanup(func() { bundledDirs = old })
	t.Setenv("PATH", "/usr/bin:/bin") // no node, bun or deno here

	b := NewBinaries(t.TempDir())
	if name, path := b.JSRuntime(); name != "deno" || path != filepath.Join(deps, "deno") {
		t.Fatalf("bundled deno not selected: %q %q", name, path)
	}
	env, err := b.Env()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("js args: %v", env.JSArgs)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	info, err := Probe(ctx, env, "https://www.youtube.com/watch?v=aqz-KE-bpKQ")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%q heights=%v", info.Title, info.Heights)
	if len(info.Heights) < 6 || info.Heights[0] < 2160 {
		t.Fatalf("expected the full ladder up to 4K, got %v", info.Heights)
	}
}

// A real (19 s) download as video and as audio, through the same Download() the queue uses.
func TestE2EDownload(t *testing.T) {
	deps, _ := filepath.Abs("../../build/deps/linux-amd64")
	old := bundledDirs
	bundledDirs = func() []string { return []string{deps} }
	t.Cleanup(func() { bundledDirs = old })
	t.Setenv("PATH", "/usr/bin:/bin")
	env, err := NewBinaries(t.TempDir()).Env()
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range []Options{
		{URL: "https://www.youtube.com/watch?v=jNQXAC9IVRw", Dir: t.TempDir(), Mode: "video", Height: 240, VideoFormat: "mp4"},
		{URL: "https://www.youtube.com/watch?v=jNQXAC9IVRw", Dir: t.TempDir(), Mode: "audio", AudioFormat: "mp3", AudioBitrate: 128},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		updates := 0
		path, err := Download(ctx, env, o, func(Progress) { updates++ })
		cancel()
		if err != nil {
			t.Fatalf("%s: %v", o.Mode, err)
		}
		st, serr := os.Stat(path)
		if serr != nil || st.Size() < 10_000 {
			t.Fatalf("%s: bad output %q (%v)", o.Mode, path, serr)
		}
		t.Logf("%s -> %s (%d bytes, %d progress updates)", o.Mode, filepath.Base(path), st.Size(), updates)
	}
}
