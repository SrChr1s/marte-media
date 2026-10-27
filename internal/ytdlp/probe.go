package ytdlp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

type Entry struct {
	ID        string  `json:"id"`
	URL       string  `json:"url"`
	Title     string  `json:"title"`
	Channel   string  `json:"channel"`
	Duration  float64 `json:"duration"`
	Thumbnail string  `json:"thumbnail"`
	ViewCount int64   `json:"viewCount"`
}

type MediaInfo struct {
	Kind      string  `json:"kind"` // video | playlist
	Entry             // for kind=video
	Heights   []int   `json:"heights"`
	Entries   []Entry `json:"entries"` // for kind=playlist
	PlaylistN int     `json:"playlistCount"`
}

type rawFormat struct {
	Height int    `json:"height"`
	VCodec string `json:"vcodec"`
}

type rawEntry struct {
	Type       string  `json:"_type"`
	ID         string  `json:"id"`
	URL        string  `json:"url"`
	WebpageURL string  `json:"webpage_url"`
	Title      string  `json:"title"`
	Channel    string  `json:"channel"`
	Uploader   string  `json:"uploader"`
	Duration   float64 `json:"duration"`
	Thumbnail  string  `json:"thumbnail"`
	ViewCount  int64   `json:"view_count"`
	IEKey      string  `json:"ie_key"`
	LiveStatus string  `json:"live_status"`
	Thumbnails []struct {
		URL string `json:"url"`
	} `json:"thumbnails"`
	Formats []rawFormat `json:"formats"`
	Entries []rawEntry  `json:"entries"`
}

func (r rawEntry) toEntry() Entry {
	e := Entry{
		ID:        r.ID,
		URL:       firstNonEmpty(r.WebpageURL, r.URL),
		Title:     r.Title,
		Channel:   firstNonEmpty(r.Channel, r.Uploader),
		Duration:  r.Duration,
		Thumbnail: r.Thumbnail,
		ViewCount: r.ViewCount,
	}
	if e.Thumbnail == "" && len(r.Thumbnails) > 0 {
		e.Thumbnail = r.Thumbnails[len(r.Thumbnails)-1].URL
	}
	if e.URL == "" && r.ID != "" {
		e.URL = "https://www.youtube.com/watch?v=" + r.ID
	}
	return e
}

// ParseInfo turns `yt-dlp -J` output into MediaInfo.
func ParseInfo(data []byte) (MediaInfo, error) {
	var raw rawEntry
	if err := json.Unmarshal(data, &raw); err != nil {
		return MediaInfo{}, fmt.Errorf("respuesta de yt-dlp inválida: %w", err)
	}
	if raw.Type == "playlist" || len(raw.Entries) > 0 {
		info := MediaInfo{Kind: "playlist"}
		info.Entry = raw.toEntry()
		for _, e := range raw.Entries {
			if e.ID == "" {
				continue
			}
			info.Entries = append(info.Entries, e.toEntry())
		}
		info.PlaylistN = len(info.Entries)
		return info, nil
	}
	info := MediaInfo{Kind: "video", Entry: raw.toEntry()}
	seen := map[int]bool{}
	for _, f := range raw.Formats {
		if f.Height > 0 && f.VCodec != "none" && !seen[f.Height] {
			seen[f.Height] = true
			info.Heights = append(info.Heights, f.Height)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(info.Heights)))
	return info, nil
}

// ErrCanceled is returned by Probe when its context was canceled.
var ErrCanceled = errors.New("cancelado")

// Probe fetches metadata for a video or playlist URL.
func Probe(ctx context.Context, env Env, url string) (MediaInfo, error) {
	out, err := runJSON(ctx, env, "--flat-playlist", "--", url)
	if err != nil {
		return MediaInfo{}, err
	}
	return ParseInfo(out)
}

// runJSON runs yt-dlp with -J (plus the JS runtime args and the given ones) and returns its JSON output.
// A canceled context becomes ErrCanceled and a timeout a readable message, for every caller alike.
func runJSON(ctx context.Context, env Env, args ...string) ([]byte, error) {
	all := append(append([]string{}, env.JSArgs...), "-J", "--no-warnings")
	all = append(all, args...)
	out, err := newCmd(ctx, env.Bin, all...).Output()
	if err != nil {
		switch ctx.Err() {
		case context.Canceled:
			return nil, ErrCanceled
		case context.DeadlineExceeded:
			return nil, errors.New("la consulta tardó demasiado")
		}
		return nil, fmt.Errorf("%s", lastLine(stderrOf(err)))
	}
	return out, nil
}

func stderrOf(err error) string {
	if ee, ok := err.(*exec.ExitError); ok {
		return string(ee.Stderr)
	}
	return err.Error()
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	return strings.TrimPrefix(lines[len(lines)-1], "ERROR: ")
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}
