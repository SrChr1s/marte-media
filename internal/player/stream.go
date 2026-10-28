package player

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"marte-media/internal/ytdlp"
)

// SourceInfo is one playable stream as the page sees it: a local URL and what it contains.
type SourceInfo struct {
	URL    string `json:"url"`
	Mime   string `json:"mime"`
	Height int    `json:"height"`
	Codec  string `json:"codec"`
}

// StreamInfo is what the page needs to play one video or song. Video and Audio are separate streams that the
// page plays in sync; Muxed is a single file used only when no separate video can be decoded.
type StreamInfo struct {
	ID        string      `json:"id"`
	Title     string      `json:"title"`
	Channel   string      `json:"channel"`
	Artist    string      `json:"artist"`
	Track     string      `json:"track"`
	Thumbnail string      `json:"thumbnail"`
	Duration  float64     `json:"duration"`
	Heights   []int       `json:"heights"`
	Video     *SourceInfo `json:"video"`
	Audio     *SourceInfo `json:"audio"`
	Muxed     *SourceInfo `json:"muxed"`
}

// Resolver turns a video id into streams served by the local Server.
type Resolver struct {
	Srv   *Server
	Cache *MediaCache
	Fetch func(ctx context.Context, id string) (ytdlp.Media, error) // normally ytdlp.Resolve
}

// Warm resolves a video into the cache without registering any stream.
func (r *Resolver) Warm(ctx context.Context, id string) {
	r.Cache.Get(ctx, id, func(ctx context.Context) (ytdlp.Media, error) { return r.Fetch(ctx, id) })
}

var ErrNothingPlayable = errors.New("este video no tiene un formato que la app pueda reproducir")

func (r *Resolver) Resolve(ctx context.Context, id string, p ytdlp.Prefs) (StreamInfo, error) {
	fetch := func(ctx context.Context) (ytdlp.Media, error) { return r.Fetch(ctx, id) }
	m, err := r.Cache.Get(ctx, id, fetch)
	if err != nil {
		return StreamInfo{}, err
	}
	picks := ytdlp.Pick(m, p)
	if picks.Audio == nil && picks.Video == nil && picks.Muxed == nil {
		return StreamInfo{}, ErrNothingPlayable
	}
	info := StreamInfo{
		ID: m.ID, Title: m.Title, Channel: m.Channel, Artist: m.Artist, Track: m.Track,
		Thumbnail: m.Thumbnail, Duration: m.Duration, Heights: picks.Heights,
	}
	info.Audio = r.add(id, picks.Audio, fetch, "audio")
	info.Video = r.add(id, picks.Video, fetch, "video")
	info.Muxed = r.add(id, picks.Muxed, fetch, "video")
	return info, nil
}

// add registers a format with the server. When its signed URL is rejected the video is resolved again
// (bypassing the cache) and the same format id is looked up in the new answer.
func (r *Resolver) add(id string, f *ytdlp.Format, fetch func(context.Context) (ytdlp.Media, error), kind string) *SourceInfo {
	if f == nil {
		return nil
	}
	formatID := f.ID
	var mu sync.Mutex
	renewed := f.AvailableAt
	local := r.Srv.Add(Source{
		URL:     f.URL,
		Headers: f.Headers,
		ReadyAt: f.AvailableAt,
		NextReady: func() time.Time {
			mu.Lock()
			defer mu.Unlock()
			return renewed
		},
		Refresh: func(ctx context.Context) (string, map[string]string, error) {
			m, err := r.Cache.Refresh(ctx, id, fetch)
			if err != nil {
				return "", nil, err
			}
			for _, nf := range m.Formats {
				if nf.ID == formatID {
					mu.Lock()
					renewed = nf.AvailableAt
					mu.Unlock()
					return nf.URL, nf.Headers, nil
				}
			}
			return "", nil, fmt.Errorf("el formato %s ya no está disponible", formatID)
		},
	})
	return &SourceInfo{URL: local, Mime: Mime(*f, kind), Height: f.Height, Codec: f.VCodec + f.ACodec}
}

// Mime is the type string for MediaSource/canPlayType style checks, e.g. `video/webm; codecs="vp9"`.
func Mime(f ytdlp.Format, kind string) string {
	if kind == "audio" {
		switch f.Ext {
		case "m4a", "mp4":
			return fmt.Sprintf(`audio/mp4; codecs="%s"`, f.ACodec)
		default:
			return fmt.Sprintf(`audio/%s; codecs="%s"`, f.Ext, f.ACodec)
		}
	}
	codecs := f.VCodec
	if f.ACodec != "" {
		codecs += ", " + f.ACodec
	}
	return fmt.Sprintf(`video/%s; codecs="%s"`, f.Ext, codecs)
}
