package player

import (
	"context"
	"sync"
	"time"

	"marte-media/internal/ytdlp"
)

// MediaCache remembers resolved videos so that changing quality, or resolving the next track ahead of time,
// does not run yt-dlp again. Concurrent requests for the same id share one call. Signed URLs stop working at
// their "expire" time, so entries are dropped well before that.
type MediaCache struct {
	mu  sync.Mutex
	m   map[string]*cached
	now func() time.Time
	ttl time.Duration
}

type cached struct {
	done  chan struct{}
	media ytdlp.Media
	err   error
	until time.Time
}

func NewMediaCache() *MediaCache {
	return &MediaCache{m: map[string]*cached{}, now: time.Now, ttl: 30 * time.Minute}
}

func (c *MediaCache) deadline(m ytdlp.Media) time.Time {
	t := c.now().Add(c.ttl)
	if !m.ExpiresAt.IsZero() {
		if e := m.ExpiresAt.Add(-10 * time.Minute); e.Before(t) {
			t = e
		}
	}
	return t
}

// Get returns the cached media, calling fetch if there is none or it is stale. Errors are not cached.
func (c *MediaCache) Get(ctx context.Context, id string, fetch func(context.Context) (ytdlp.Media, error)) (ytdlp.Media, error) {
	return c.get(ctx, id, fetch, false)
}

// Refresh always calls fetch (the signed URLs were rejected) and replaces the entry.
func (c *MediaCache) Refresh(ctx context.Context, id string, fetch func(context.Context) (ytdlp.Media, error)) (ytdlp.Media, error) {
	return c.get(ctx, id, fetch, true)
}

func (c *MediaCache) get(ctx context.Context, id string, fetch func(context.Context) (ytdlp.Media, error), force bool) (ytdlp.Media, error) {
	for {
		c.mu.Lock()
		e, ok := c.m[id]
		if ok && !force {
			select {
			case <-e.done: // finished
				if e.err == nil && c.now().Before(e.until) {
					c.mu.Unlock()
					return e.media, nil
				}
			default: // in flight: wait for it
				c.mu.Unlock()
				select {
				case <-e.done:
				case <-ctx.Done():
					return ytdlp.Media{}, ctx.Err()
				}
				if e.err == nil {
					return e.media, nil
				}
				return ytdlp.Media{}, e.err
			}
		}
		e = &cached{done: make(chan struct{})}
		c.m[id] = e
		c.mu.Unlock()

		e.media, e.err = fetch(ctx)
		if e.err == nil {
			e.until = c.deadline(e.media)
		}
		close(e.done)
		if e.err != nil {
			c.mu.Lock()
			if c.m[id] == e {
				delete(c.m, id)
			}
			c.mu.Unlock()
			return ytdlp.Media{}, e.err
		}
		return e.media, nil
	}
}
