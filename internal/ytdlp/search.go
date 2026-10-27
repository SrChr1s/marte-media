package ytdlp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	KindVideo = "video"
	KindMusic = "music"
)

// Search looks up videos (YouTube) or songs (YouTube Music).
//
// YouTube Music's search mixes songs with albums and artists, and for songs yt-dlp's flat result carries
// only the title (no artist, duration or cover). Only the playable entries are kept, with the cover taken
// from i.ytimg.com; the rest of the metadata arrives when the track is resolved.
func Search(ctx context.Context, env Env, query, kind string, limit int) ([]Entry, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("escribe algo para buscar")
	}
	if limit <= 0 {
		limit = 30
	}
	target := fmt.Sprintf("ytsearch%d:%s", limit, query)
	if kind == KindMusic {
		target = "https://music.youtube.com/search?q=" + url.QueryEscape(query)
	}
	out, err := runJSON(ctx, env, "--flat-playlist", "--playlist-end", strconv.Itoa(limit), "--", target)
	if err != nil {
		return nil, err
	}
	return ParseSearch(out, kind)
}

// ParseSearch turns the JSON of a search into playable entries.
func ParseSearch(data []byte, kind string) ([]Entry, error) {
	var raw rawEntry
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("respuesta de yt-dlp inválida: %w", err)
	}
	entries := []Entry{}
	seen := map[string]bool{}
	for _, r := range raw.Entries {
		// Albums and artists come back as "YoutubeTab" pages; live streams have no direct audio/video file.
		if r.ID == "" || len(r.ID) != 11 || seen[r.ID] {
			continue
		}
		if r.IEKey != "" && r.IEKey != "Youtube" {
			continue
		}
		if r.LiveStatus == "is_live" || r.LiveStatus == "is_upcoming" {
			continue
		}
		seen[r.ID] = true
		e := r.toEntry()
		e.URL = "https://www.youtube.com/watch?v=" + r.ID
		if e.Thumbnail == "" {
			e.Thumbnail = "https://i.ytimg.com/vi/" + r.ID + "/hqdefault.jpg"
		}
		entries = append(entries, e)
	}
	return entries, nil
}
