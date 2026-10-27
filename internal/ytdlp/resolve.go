package ytdlp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Format is the part of a yt-dlp format the player needs.
type Format struct {
	ID       string
	Ext      string
	VCodec   string
	ACodec   string
	Protocol string
	Note     string
	Height   int
	FPS      float64
	ABR      float64
	Channels int
	URL      string
	Headers  map[string]string
	// AvailableAt is when the URL starts to work: some clients' URLs answer 403 until then (yt-dlp sleeps for it).
	AvailableAt time.Time
}

// Media is a resolved video: its metadata and every direct format.
type Media struct {
	ID        string
	Title     string
	Channel   string
	Artist    string // set for YouTube Music tracks
	Track     string
	Thumbnail string
	Duration  float64
	Formats   []Format
	ExpiresAt time.Time // when the signed URLs stop working
}

type rawMedia struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Channel   string  `json:"channel"`
	Uploader  string  `json:"uploader"`
	Artist    string  `json:"artist"`
	Track     string  `json:"track"`
	Duration  float64 `json:"duration"`
	Thumbnail string  `json:"thumbnail"`
	Live      string  `json:"live_status"`
	Formats   []struct {
		ID       string            `json:"format_id"`
		Ext      string            `json:"ext"`
		VCodec   string            `json:"vcodec"`
		ACodec   string            `json:"acodec"`
		Protocol string            `json:"protocol"`
		Note     string            `json:"format_note"`
		Height   int               `json:"height"`
		FPS      float64           `json:"fps"`
		ABR      float64           `json:"abr"`
		Channels int               `json:"audio_channels"`
		URL      string            `json:"url"`
		Headers  map[string]string `json:"http_headers"`
		Avail    int64             `json:"available_at"`
	} `json:"formats"`
}

// ParseMedia reads the JSON that `yt-dlp -J --no-playlist <video>` prints.
func ParseMedia(data []byte) (Media, error) {
	var r rawMedia
	if err := json.Unmarshal(data, &r); err != nil {
		return Media{}, fmt.Errorf("respuesta de yt-dlp inválida: %w", err)
	}
	if r.Live == "is_live" || r.Live == "is_upcoming" {
		return Media{}, fmt.Errorf("las transmisiones en vivo todavía no se pueden reproducir")
	}
	m := Media{ID: r.ID, Title: r.Title, Channel: firstNonEmpty(r.Channel, r.Uploader), Artist: r.Artist, Track: r.Track, Thumbnail: r.Thumbnail, Duration: r.Duration}
	for _, f := range r.Formats {
		if f.URL == "" {
			continue
		}
		var avail time.Time
		if f.Avail > 0 {
			avail = time.Unix(f.Avail, 0)
		}
		m.Formats = append(m.Formats, Format{
			ID: f.ID, Ext: f.Ext, VCodec: norm(f.VCodec), ACodec: norm(f.ACodec), Protocol: f.Protocol, Note: f.Note,
			Height: f.Height, FPS: f.FPS, ABR: f.ABR, Channels: f.Channels, URL: f.URL, Headers: f.Headers, AvailableAt: avail,
		})
		if t := expiry(f.URL); !t.IsZero() && (m.ExpiresAt.IsZero() || t.Before(m.ExpiresAt)) {
			m.ExpiresAt = t
		}
	}
	if len(m.Formats) == 0 {
		return Media{}, fmt.Errorf("no hay formatos reproducibles para este video")
	}
	return m, nil
}

// Resolve fetches the formats of one video (no playlist, no download).
func Resolve(ctx context.Context, env Env, id string) (Media, error) {
	out, err := runJSON(ctx, env, "--no-playlist", "--", "https://www.youtube.com/watch?v="+id)
	if err != nil {
		return Media{}, err
	}
	return ParseMedia(out)
}

func norm(s string) string {
	if s == "none" {
		return ""
	}
	return s
}

// expiry reads the "expire" (unix seconds) parameter of a googlevideo URL.
func expiry(raw string) time.Time {
	u, err := url.Parse(raw)
	if err != nil {
		return time.Time{}
	}
	if n, err := strconv.ParseInt(u.Query().Get("expire"), 10, 64); err == nil && n > 0 {
		return time.Unix(n, 0)
	}
	return time.Time{}
}

// Prefs says what the webview can decode and what the user wants. Codec lists are families in order of
// preference: video "vp9", "av01", "avc1"; audio "opus", "aac". Only what canPlayType() accepts goes in.
type Prefs struct {
	Height      int // wanted height; 0 = 720
	VideoCodecs []string
	AudioCodecs []string
}

// Picks is the chosen set of formats.
type Picks struct {
	Video   *Format // video-only, to be played together with Audio
	Audio   *Format
	Muxed   *Format // audio+video in one file, only when no separate video fits
	Heights []int   // every height available with a decodable codec, tallest first
}

func vfamily(c string) string {
	switch {
	case strings.HasPrefix(c, "vp9"), strings.HasPrefix(c, "vp09"):
		return "vp9"
	case strings.HasPrefix(c, "av01"):
		return "av01"
	case strings.HasPrefix(c, "avc1"):
		return "avc1"
	}
	return ""
}

func afamily(c string) string {
	switch {
	case strings.HasPrefix(c, "opus"):
		return "opus"
	case strings.HasPrefix(c, "mp4a"):
		return "aac"
	}
	return ""
}

func rank(list []string, v string) int {
	for i, x := range list {
		if x == v {
			return i
		}
	}
	return -1
}

// direct reports whether a format is a plain file over HTTP (no HLS playlist, no storyboard).
func direct(f Format) bool { return f.Protocol == "https" || f.Protocol == "http" }

// Pick chooses the formats to play. Audio never uses the "DRC" (compressed dynamic range) variants, and
// formats the webview cannot decode are never offered.
func Pick(m Media, p Prefs) Picks {
	want := p.Height
	if want <= 0 {
		want = 720
	}
	var out Picks

	// audio: stereo only (the 5.1 tracks are HE-AAC/AC-3 and play badly), best decodable, preferring the codec
	// order, capped so the huge streams are not chosen
	for i := range m.Formats {
		f := &m.Formats[i]
		fam := afamily(f.ACodec)
		if f.VCodec != "" || fam == "" || !direct(*f) || f.Channels > 2 || strings.Contains(f.ID, "drc") || strings.Contains(strings.ToLower(f.Note), "drc") {
			continue
		}
		r := rank(p.AudioCodecs, fam)
		if r < 0 || f.ABR > 200 {
			continue
		}
		if out.Audio == nil {
			out.Audio = f
			continue
		}
		cur := rank(p.AudioCodecs, afamily(out.Audio.ACodec))
		if r < cur || (r == cur && f.ABR > out.Audio.ABR) {
			out.Audio = f
		}
	}

	// video: video-only formats, grouped by height
	heights := map[int]bool{}
	var vids []*Format
	for i := range m.Formats {
		f := &m.Formats[i]
		if f.VCodec == "" || f.ACodec != "" || !direct(*f) || f.Height <= 0 {
			continue
		}
		if rank(p.VideoCodecs, vfamily(f.VCodec)) < 0 {
			continue
		}
		heights[f.Height] = true
		vids = append(vids, f)
	}
	for h := range heights {
		out.Heights = append(out.Heights, h)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(out.Heights)))

	if len(out.Heights) > 0 {
		target := out.Heights[len(out.Heights)-1] // smallest, if everything is taller than wanted
		for _, h := range out.Heights {
			if h <= want {
				target = h
				break
			}
		}
		for _, f := range vids {
			if f.Height != target {
				continue
			}
			if out.Video == nil {
				out.Video = f
				continue
			}
			rf, rc := rank(p.VideoCodecs, vfamily(f.VCodec)), rank(p.VideoCodecs, vfamily(out.Video.VCodec))
			// same height: codec order first, then the lower frame rate (60 fps costs twice the CPU to decode)
			if rf < rc || (rf == rc && f.FPS < out.Video.FPS) {
				out.Video = f
			}
		}
		return out
	}

	// no separate video we can decode: a muxed file (YouTube's format 18) is the last resort
	for i := range m.Formats {
		f := &m.Formats[i]
		if f.VCodec != "" && f.ACodec != "" && direct(*f) && rank(p.VideoCodecs, vfamily(f.VCodec)) >= 0 && rank(p.AudioCodecs, afamily(f.ACodec)) >= 0 {
			if out.Muxed == nil || f.Height > out.Muxed.Height {
				out.Muxed = f
			}
		}
	}
	if out.Muxed != nil {
		out.Heights = []int{out.Muxed.Height}
	}
	return out
}
