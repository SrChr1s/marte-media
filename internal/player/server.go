// Package player serves the audio and video streams yt-dlp found to the webview.
//
// The page plays them through <audio>/<video> elements pointed at http://127.0.0.1:<port>/s/<secret>/<id>.
// Going through here instead of using the googlevideo URL directly lets us send the headers yt-dlp used,
// refresh a signed URL that expired, and keep other pages and processes from using the proxy: it listens on
// the loopback only and every path carries a random secret.
package player

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DefaultChunk is how much of a stream is requested at once. YouTube throttles open-ended byte ranges to
// roughly 40 KB/s but serves bounded chunks at full speed, which is why yt-dlp downloads in 10 MiB pieces.
const DefaultChunk = 10 << 20

// Source is one stream to serve.
type Source struct {
	URL     string
	Headers map[string]string
	// Refresh returns a new signed URL (and headers) once the current one is rejected (403/410).
	Refresh func(ctx context.Context) (string, map[string]string, error)
	// ReadyAt is when the URL starts to be accepted (it answers 403 before); zero means now.
	ReadyAt time.Time
	// NextReady, when set, gives the ReadyAt of the URL the last Refresh returned.
	NextReady func() time.Time
	// Chunk caps the size of each upstream request; 0 means DefaultChunk.
	Chunk int64
}

type entry struct {
	mu  sync.Mutex
	src Source
}

func (e *entry) current() (string, map[string]string, time.Time) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.src.URL, e.src.Headers, e.src.ReadyAt
}

// refresh replaces the URL, unless another request already did (stale is the URL that just failed).
func (e *entry) refresh(ctx context.Context, stale string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.src.URL != stale {
		return nil
	}
	if e.src.Refresh == nil {
		return errors.New("el enlace caducó y no se puede renovar")
	}
	u, h, err := e.src.Refresh(ctx)
	if err != nil {
		return err
	}
	e.src.URL, e.src.Headers = u, h
	e.src.ReadyAt = time.Time{}
	if e.src.NextReady != nil {
		e.src.ReadyAt = e.src.NextReady()
	}
	return nil
}

type Server struct {
	secret string
	ln     net.Listener
	srv    *http.Server
	client *http.Client
	max    int

	mu      sync.Mutex
	sources map[string]*entry
	order   []string
}

// New binds 127.0.0.1 on a random port. Call Start to serve.
func New() (*Server, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	s := &Server{
		secret:  token(16),
		ln:      ln,
		sources: map[string]*entry{},
		max:     256,
		client: &http.Client{Transport: &http.Transport{
			ResponseHeaderTimeout: 25e9,
			DisableCompression:    true, // we forward byte ranges; they must refer to the bytes on the wire
			MaxIdleConnsPerHost:   4,
		}},
	}
	s.srv = &http.Server{Handler: http.HandlerFunc(s.serve)}
	return s, nil
}

func token(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func (s *Server) Start()       { go s.srv.Serve(s.ln) }
func (s *Server) Close() error { return s.srv.Close() }
func (s *Server) Addr() string { return s.ln.Addr().String() }

// Add registers a stream and returns the local URL the page should play.
func (s *Server) Add(src Source) string {
	id := token(8)
	if src.Chunk <= 0 {
		src.Chunk = DefaultChunk
	}
	s.mu.Lock()
	s.sources[id] = &entry{src: src}
	s.order = append(s.order, id)
	for len(s.order) > s.max {
		delete(s.sources, s.order[0])
		s.order = s.order[1:]
	}
	s.mu.Unlock()
	return fmt.Sprintf("http://%s/s/%s/%s", s.Addr(), s.secret, id)
}

func (s *Server) lookup(id string) *entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sources[id]
}

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	// A page on another origin could still reach the port; only requests addressed to us are served.
	if r.Host != s.Addr() {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(parts) != 3 || parts[0] != "s" || subtle.ConstantTimeCompare([]byte(parts[1]), []byte(s.secret)) != 1 {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	e := s.lookup(parts[2])
	if e == nil {
		http.NotFound(w, r)
		return
	}

	resp, err := s.fetch(r, e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	h := w.Header()
	for _, k := range []string{"Content-Type", "Content-Length", "Content-Range", "Last-Modified", "Etag"} {
		if v := resp.Header.Get(k); v != "" {
			h.Set(k, v)
		}
	}
	h.Set("Accept-Ranges", "bytes")
	h.Set("Cache-Control", "no-store")
	w.WriteHeader(resp.StatusCode)
	// Send the headers now: the player needs the type and length before the first byte of data arrives.
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	if r.Method != http.MethodHead {
		io.Copy(w, resp.Body) // ends when either side closes: the request context cancels the upstream read
	}
}

var byteRange = regexp.MustCompile(`^bytes=(\d+)-(\d*)$`)

// clampRange limits "bytes=N-" and over-long ranges to one chunk. The answer is then a 206 whose
// Content-Range says what was really sent, which is all a media element needs to ask for the next piece.
// Suffix ("bytes=-N") and multi-range requests are left alone.
func clampRange(h string, chunk int64) string {
	m := byteRange.FindStringSubmatch(h)
	if m == nil || chunk <= 0 {
		return h
	}
	start, _ := strconv.ParseInt(m[1], 10, 64)
	if m[2] != "" {
		if end, _ := strconv.ParseInt(m[2], 10, 64); end-start+1 <= chunk {
			return h
		}
	}
	return fmt.Sprintf("bytes=%d-%d", start, start+chunk-1)
}

// fetch asks the upstream for the stream (the same byte range the page asked for). If the signed URL was
// rejected it is renewed once and the request repeated.
func (s *Server) fetch(r *http.Request, e *entry) (*http.Response, error) {
	for attempt := 0; ; attempt++ {
		u, headers, ready := e.current()
		if err := waitUntil(r.Context(), ready); err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		for k, v := range headers {
			switch strings.ToLower(k) {
			case "range", "host", "accept-encoding", "content-length":
			default:
				req.Header.Set(k, v)
			}
		}
		req.Header.Set("Accept-Encoding", "identity")
		if rg := r.Header.Get("Range"); rg != "" {
			req.Header.Set("Range", clampRange(rg, e.src.Chunk))
		}
		resp, err := s.client.Do(req)
		if err != nil {
			return nil, err
		}
		if (resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusGone) && attempt == 0 {
			resp.Body.Close()
			if err := e.refresh(r.Context(), u); err != nil {
				return nil, fmt.Errorf("no se pudo renovar el enlace: %w", err)
			}
			continue
		}
		return resp, nil
	}
}

// waitUntil blocks until t (or the context ends). Fresh YouTube URLs answer 403 for a few seconds.
func waitUntil(ctx context.Context, t time.Time) error {
	d := time.Until(t)
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
