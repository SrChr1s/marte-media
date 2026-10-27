package player

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"marte-media/internal/ytdlp"
)

func fixture(t *testing.T) ytdlp.Media {
	t.Helper()
	b, err := os.ReadFile("../ytdlp/testdata/formats_bbb.json")
	if err != nil {
		t.Fatal(err)
	}
	m, err := ytdlp.ParseMedia(b)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

var prefs = ytdlp.Prefs{Height: 720, VideoCodecs: []string{"vp9", "av01"}, AudioCodecs: []string{"opus", "aac"}}

func TestCacheSharesInFlightAndReuses(t *testing.T) {
	c := NewMediaCache()
	var calls atomic.Int32
	release := make(chan struct{})
	fetch := func(context.Context) (ytdlp.Media, error) {
		calls.Add(1)
		<-release
		return ytdlp.Media{ID: "x"}, nil
	}
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := c.Get(context.Background(), "x", fetch); err != nil {
				t.Error(err)
			}
		}()
	}
	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()
	c.Get(context.Background(), "x", fetch) // already cached
	if calls.Load() != 1 {
		t.Errorf("fetch should run once for concurrent and repeated gets, ran %d times", calls.Load())
	}
}

func TestCacheExpiryErrorsAndRefresh(t *testing.T) {
	c := NewMediaCache()
	now := time.Now()
	c.now = func() time.Time { return now }
	var calls int
	fetch := func(context.Context) (ytdlp.Media, error) { calls++; return ytdlp.Media{ID: "x"}, nil }

	c.Get(context.Background(), "x", fetch)
	now = now.Add(29 * time.Minute)
	c.Get(context.Background(), "x", fetch)
	if calls != 1 {
		t.Fatalf("still fresh: want 1 call, got %d", calls)
	}
	now = now.Add(2 * time.Minute)
	c.Get(context.Background(), "x", fetch)
	if calls != 2 {
		t.Fatalf("after the TTL: want 2 calls, got %d", calls)
	}
	c.Refresh(context.Background(), "x", fetch)
	if calls != 3 {
		t.Fatalf("Refresh must bypass the cache: %d", calls)
	}

	// signed URLs that expire soon shorten the lifetime
	soon := ytdlp.Media{ID: "y", ExpiresAt: now.Add(12 * time.Minute)}
	c.Get(context.Background(), "y", func(context.Context) (ytdlp.Media, error) { calls++; return soon, nil })
	now = now.Add(3 * time.Minute) // 10 min before expiry
	c.Get(context.Background(), "y", func(context.Context) (ytdlp.Media, error) { calls++; return soon, nil })
	if calls != 5 {
		t.Errorf("entry should be dropped 10 minutes before its URLs expire, calls=%d", calls)
	}

	boom := errors.New("boom")
	var n int
	bad := func(context.Context) (ytdlp.Media, error) { n++; return ytdlp.Media{}, boom }
	c.Get(context.Background(), "z", bad)
	c.Get(context.Background(), "z", bad)
	if n != 2 {
		t.Errorf("errors must not be cached, fetch ran %d times", n)
	}
}

func newResolver(t *testing.T, fetch func(context.Context, string) (ytdlp.Media, error)) *Resolver {
	t.Helper()
	srv, err := New()
	if err != nil {
		t.Fatal(err)
	}
	srv.Start()
	t.Cleanup(func() { srv.Close() })
	return &Resolver{Srv: srv, Cache: NewMediaCache(), Fetch: fetch}
}

func TestResolveBuildsLocalStreams(t *testing.T) {
	m := fixture(t)
	r := newResolver(t, func(context.Context, string) (ytdlp.Media, error) { return m, nil })
	info, err := r.Resolve(context.Background(), m.ID, prefs)
	if err != nil {
		t.Fatal(err)
	}
	if info.Video == nil || info.Audio == nil || info.Muxed != nil {
		t.Fatalf("expected separate video and audio: %+v", info)
	}
	for _, s := range []*SourceInfo{info.Video, info.Audio} {
		if !strings.HasPrefix(s.URL, "http://127.0.0.1:") || strings.Contains(s.URL, "googlevideo") {
			t.Errorf("page must only see local URLs, got %s", s.URL)
		}
	}
	if info.Video.Mime != `video/webm; codecs="vp9"` || info.Audio.Mime != `audio/webm; codecs="opus"` || info.Video.Height != 720 {
		t.Errorf("mimes/height: %q %q %d", info.Video.Mime, info.Audio.Mime, info.Video.Height)
	}
	if info.Title == "" || info.Duration < 500 || len(info.Heights) < 6 {
		t.Errorf("metadata missing: %+v", info)
	}
}

func TestResolveNothingPlayable(t *testing.T) {
	m := fixture(t)
	r := newResolver(t, func(context.Context, string) (ytdlp.Media, error) { return m, nil })
	_, err := r.Resolve(context.Background(), m.ID, ytdlp.Prefs{VideoCodecs: []string{"theora"}, AudioCodecs: []string{"wma"}})
	if !errors.Is(err, ErrNothingPlayable) {
		t.Fatalf("want ErrNothingPlayable, got %v", err)
	}
}

// The signed URL of the audio expires: the proxy re-resolves the video and plays the same format id again.
func TestExpiredURLIsRenewedThroughResolver(t *testing.T) {
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("fresh-audio")) }))
	defer good.Close()
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "expired", 403) }))
	defer dead.Close()

	var fetches atomic.Int32
	base := fixture(t)
	fetch := func(context.Context, string) (ytdlp.Media, error) {
		m := base
		m.Formats = append([]ytdlp.Format{}, base.Formats...)
		for i := range m.Formats {
			if m.Formats[i].ID == "251" {
				m.Formats[i].URL = dead.URL // first answer: already expired
				if fetches.Add(1) > 1 {
					m.Formats[i].URL = good.URL // the re-resolve returns a fresh one
				}
			}
		}
		return m, nil
	}
	r := newResolver(t, fetch)
	info, err := r.Resolve(context.Background(), base.ID, prefs)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Get(info.Audio.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b := make([]byte, 64)
	n, _ := resp.Body.Read(b)
	if resp.StatusCode != 200 || string(b[:n]) != "fresh-audio" {
		t.Fatalf("expected the renewed stream, got %d %q", resp.StatusCode, b[:n])
	}
	if fetches.Load() != 2 {
		t.Errorf("one resolve plus one refresh expected, got %d", fetches.Load())
	}
}
