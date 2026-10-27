package player

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

var body = strings.Repeat("0123456789", 100) // 1000 bytes

func newServer(t *testing.T) *Server {
	t.Helper()
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	s.Start()
	t.Cleanup(func() { s.Close() })
	return s
}

// upstream mimics googlevideo: byte ranges via ServeContent, and it records what it was sent.
func upstream(t *testing.T, seen *http.Header) *httptest.Server {
	t.Helper()
	u := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if seen != nil {
			*seen = r.Header.Clone()
		}
		w.Header().Set("Content-Type", "audio/webm")
		http.ServeContent(w, r, "a.webm", time.Time{}, strings.NewReader(body))
	}))
	t.Cleanup(u.Close)
	return u
}

func get(t *testing.T, url string, hdr map[string]string) (*http.Response, string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp, string(b)
}

func TestRangePassthrough(t *testing.T) {
	s := newServer(t)
	local := s.Add(Source{URL: upstream(t, nil).URL})

	resp, b := get(t, local, nil)
	if resp.StatusCode != 200 || len(b) != 1000 || resp.Header.Get("Content-Type") != "audio/webm" {
		t.Fatalf("full: %d %d %q", resp.StatusCode, len(b), resp.Header.Get("Content-Type"))
	}
	resp, b = get(t, local, map[string]string{"Range": "bytes=100-199"})
	if resp.StatusCode != 206 || resp.Header.Get("Content-Range") != "bytes 100-199/1000" || b != body[100:200] {
		t.Fatalf("range: %d %q %q", resp.StatusCode, resp.Header.Get("Content-Range"), b)
	}
	if resp.Header.Get("Accept-Ranges") != "bytes" {
		t.Error("must advertise byte ranges so the player can seek")
	}
	resp, _ = get(t, local, map[string]string{"Range": "bytes=5000-"})
	if resp.StatusCode != 416 {
		t.Errorf("out of range should be 416, got %d", resp.StatusCode)
	}
}

func TestRangeIsChunked(t *testing.T) {
	var seen http.Header
	s := newServer(t)
	local := s.Add(Source{URL: upstream(t, &seen).URL, Chunk: 100})

	cases := []struct{ ask, upstream, content string }{
		{"bytes=0-", "bytes=0-99", "bytes 0-99/1000"},          // open-ended: only one chunk
		{"bytes=500-", "bytes=500-599", "bytes 500-599/1000"},  // from the middle
		{"bytes=0-9999", "bytes=0-99", "bytes 0-99/1000"},      // longer than a chunk
		{"bytes=10-19", "bytes=10-19", "bytes 10-19/1000"},     // small: untouched
		{"bytes=950-", "bytes=950-1049", "bytes 950-999/1000"}, // near the end: upstream trims
		{"bytes=-50", "bytes=-50", "bytes 950-999/1000"},       // suffix: untouched
	}
	for _, c := range cases {
		resp, b := get(t, local, map[string]string{"Range": c.ask})
		if seen.Get("Range") != c.upstream {
			t.Errorf("%s: upstream was asked for %q, want %q", c.ask, seen.Get("Range"), c.upstream)
		}
		if resp.StatusCode != 206 || resp.Header.Get("Content-Range") != c.content || len(b) > 100 {
			t.Errorf("%s: got %d %q (%d bytes), want 206 %q", c.ask, resp.StatusCode, resp.Header.Get("Content-Range"), len(b), c.content)
		}
	}
	// without a Range header the request is not turned into a partial one
	if resp, b := get(t, local, nil); resp.StatusCode != 200 || len(b) != 1000 {
		t.Errorf("no Range: want the whole body, got %d with %d bytes", resp.StatusCode, len(b))
	}
}

func TestForwardsHeadersButNotEncoding(t *testing.T) {
	var seen http.Header
	s := newServer(t)
	local := s.Add(Source{URL: upstream(t, &seen).URL, Headers: map[string]string{"User-Agent": "yt-test", "Accept-Language": "en-us", "Range": "bytes=0-1"}})
	get(t, local, map[string]string{"Accept-Encoding": "gzip"})
	if seen.Get("User-Agent") != "yt-test" || seen.Get("Accept-Language") != "en-us" {
		t.Errorf("yt-dlp headers not forwarded: %v", seen)
	}
	if seen.Get("Accept-Encoding") != "identity" {
		t.Errorf("must ask for identity encoding, got %q", seen.Get("Accept-Encoding"))
	}
	if seen.Get("Range") != "" {
		t.Errorf("a Range from yt-dlp's headers must not leak in: %q", seen.Get("Range"))
	}
}

func TestRejectsBadAccess(t *testing.T) {
	s := newServer(t)
	local := s.Add(Source{URL: upstream(t, nil).URL})
	id := local[strings.LastIndex(local, "/")+1:]

	for name, u := range map[string]string{
		"wrong secret": fmt.Sprintf("http://%s/s/%s/%s", s.Addr(), "deadbeef", id),
		"unknown id":   strings.TrimSuffix(local, id) + "nope",
		"bad path":     fmt.Sprintf("http://%s/", s.Addr()),
	} {
		if resp, _ := get(t, u, nil); resp.StatusCode != 404 {
			t.Errorf("%s: want 404, got %d", name, resp.StatusCode)
		}
	}

	// DNS-rebinding style request: right path, wrong Host header
	req, _ := http.NewRequest(http.MethodGet, local, nil)
	req.Host = "evil.example"
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Errorf("wrong Host: want 404, got %d", resp.StatusCode)
	}

	req, _ = http.NewRequest(http.MethodPost, local, nil)
	resp, _ = http.DefaultClient.Do(req)
	resp.Body.Close()
	if resp.StatusCode != 405 {
		t.Errorf("POST: want 405, got %d", resp.StatusCode)
	}
}

func TestHeadHasNoBody(t *testing.T) {
	s := newServer(t)
	local := s.Add(Source{URL: upstream(t, nil).URL})
	resp, err := http.Head(local)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 || resp.Header.Get("Content-Length") != "1000" || len(b) != 0 {
		t.Errorf("HEAD: %d len=%s body=%d", resp.StatusCode, resp.Header.Get("Content-Length"), len(b))
	}
}

func TestRefreshOnExpiredURL(t *testing.T) {
	expired := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "expired", 403) }))
	defer expired.Close()
	fresh := upstream(t, nil)

	var refreshes atomic.Int32
	s := newServer(t)
	local := s.Add(Source{URL: expired.URL, Refresh: func(context.Context) (string, map[string]string, error) {
		refreshes.Add(1)
		return fresh.URL, nil, nil
	}})

	for i := 0; i < 3; i++ {
		resp, b := get(t, local, map[string]string{"Range": "bytes=0-9"})
		if resp.StatusCode != 206 || b != body[:10] {
			t.Fatalf("request %d: %d %q", i, resp.StatusCode, b)
		}
	}
	if refreshes.Load() != 1 {
		t.Errorf("the URL should be renewed once and then reused, refreshed %d times", refreshes.Load())
	}
}

func TestRefreshFailureIsBadGateway(t *testing.T) {
	gone := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "gone", 410) }))
	defer gone.Close()
	s := newServer(t)

	failing := s.Add(Source{URL: gone.URL, Refresh: func(context.Context) (string, map[string]string, error) { return "", nil, errors.New("boom") }})
	if resp, _ := get(t, failing, nil); resp.StatusCode != 502 {
		t.Errorf("failed refresh: want 502, got %d", resp.StatusCode)
	}
	noRefresh := s.Add(Source{URL: gone.URL})
	if resp, _ := get(t, noRefresh, nil); resp.StatusCode != 502 {
		t.Errorf("no refresh function: want 502, got %d", resp.StatusCode)
	}
}

func TestOldSourcesArePruned(t *testing.T) {
	s := newServer(t)
	s.max = 3
	up := upstream(t, nil)
	first := s.Add(Source{URL: up.URL})
	for i := 0; i < 3; i++ {
		s.Add(Source{URL: up.URL})
	}
	if resp, _ := get(t, first, nil); resp.StatusCode != 404 {
		t.Errorf("oldest source should be gone, got %d", resp.StatusCode)
	}
}

func TestClientCancelStopsUpstream(t *testing.T) {
	done := make(chan struct{})
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		<-r.Context().Done() // returns only when the proxy drops the upstream connection
		close(done)
	}))
	t.Cleanup(slow.Close)
	t.Cleanup(slow.CloseClientConnections) // runs first: a failing test must not hang waiting for the handler
	s := newServer(t)
	local := s.Add(Source{URL: slow.URL})

	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, local, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	resp.Body.Close()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("upstream connection was not released after the client left")
	}
}

func TestWaitsUntilReady(t *testing.T) {
	var firstHit time.Time
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if firstHit.IsZero() {
			firstHit = time.Now()
		}
		w.Write([]byte("ok"))
	}))
	defer up.Close()
	s := newServer(t)
	ready := time.Now().Add(300 * time.Millisecond)
	local := s.Add(Source{URL: up.URL, ReadyAt: ready})
	resp, _ := get(t, local, nil)
	if resp.StatusCode != 200 || firstHit.Before(ready) {
		t.Fatalf("status %d, upstream hit %v before it was ready", resp.StatusCode, ready.Sub(firstHit))
	}
}
