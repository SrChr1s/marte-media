//go:build e2e

package player

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"marte-media/internal/ytdlp"
)

// Run with: go test -tags e2e ./internal/player -run E2E -v
// Needs network and build/deps/linux-amd64 (scripts/fetch-deps.sh). Real search, real resolution, and the real
// signed URLs fetched through the local proxy.
func realEnv(t *testing.T) ytdlp.Env {
	t.Helper()
	deps, _ := filepath.Abs("../../build/deps/linux-amd64")
	env := ytdlp.Env{Bin: filepath.Join(deps, "yt-dlp"), FFmpeg: filepath.Join(deps, "ffmpeg"), JSArgs: []string{"--js-runtimes", "deno:" + filepath.Join(deps, "deno")}}
	return env
}

func TestE2ESearchAndPlayThroughProxy(t *testing.T) {
	env := realEnv(t)
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	videos, err := ytdlp.Search(ctx, env, "marte documental", ytdlp.KindVideo, 5)
	if err != nil || len(videos) == 0 {
		t.Fatalf("video search: %v (%d)", err, len(videos))
	}
	t.Logf("video search: %d results, first %q by %q (%.0fs, %d views)", len(videos), videos[0].Title, videos[0].Channel, videos[0].Duration, videos[0].ViewCount)

	songs, err := ytdlp.Search(ctx, env, "daft punk get lucky", ytdlp.KindMusic, 20)
	if err != nil || len(songs) == 0 {
		t.Fatalf("music search: %v (%d)", err, len(songs))
	}
	t.Logf("music search: %d songs, first %q", len(songs), songs[0].Title)

	srv, err := New()
	if err != nil {
		t.Fatal(err)
	}
	srv.Start()
	defer srv.Close()
	r := &Resolver{Srv: srv, Cache: NewMediaCache(), Fetch: func(ctx context.Context, id string) (ytdlp.Media, error) { return ytdlp.Resolve(ctx, env, id) }}

	info, err := r.Resolve(ctx, "aqz-KE-bpKQ", ytdlp.Prefs{Height: 720, VideoCodecs: []string{"vp9", "av01"}, AudioCodecs: []string{"opus", "aac"}})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("resolved %q: heights=%v video=%s audio=%s", info.Title, info.Heights, info.Video.Mime, info.Audio.Mime)

	for name, s := range map[string]*SourceInfo{"audio": info.Audio, "video": info.Video} {
		req, _ := http.NewRequest(http.MethodGet, s.URL, nil)
		req.Header.Set("Range", "bytes=0-4095")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		head, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		resp.Body.Close()
		if resp.StatusCode != http.StatusPartialContent || len(head) != 4096 {
			t.Fatalf("%s: want 206 with 4096 bytes through the proxy, got %d with %d (%s)", name, resp.StatusCode, len(head), resp.Header.Get("Content-Range"))
		}
		if !bytes.HasPrefix(head, []byte{0x1A, 0x45, 0xDF, 0xA3}) { // EBML header: it is really WebM
			t.Fatalf("%s: not a WebM stream, starts with % x (type %q)", name, head[:8], resp.Header.Get("Content-Type"))
		}
		t.Logf("%s: 206 %s, %s, WebM header ok", name, resp.Header.Get("Content-Range"), resp.Header.Get("Content-Type"))
	}
}
