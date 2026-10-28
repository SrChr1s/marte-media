package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"marte-media/internal/player"
	"marte-media/internal/ytdlp"
)

// playState is everything the player view needs on the Go side.
type playState struct {
	srv *player.Server
	res *player.Resolver

	searchMu     sync.Mutex
	searchCancel context.CancelFunc

	warm chan struct{} // slots for background resolutions (Prefetch)
}

func (a *App) startPlayer() {
	srv, err := player.New()
	if err != nil {
		runtime.LogErrorf(a.ctx, "no se pudo iniciar el servidor del reproductor: %v", err)
		return
	}
	srv.Start()
	a.play = &playState{srv: srv, warm: make(chan struct{}, 2), res: &player.Resolver{
		Srv:   srv,
		Cache: player.NewMediaCache(),
		Fetch: func(ctx context.Context, id string) (ytdlp.Media, error) {
			env, err := a.bins.Env()
			if err != nil {
				return ytdlp.Media{}, err
			}
			return ytdlp.Resolve(ctx, env, id)
		},
	}}
}

func (a *App) shutdown(context.Context) {
	if a.play != nil {
		a.play.srv.Close()
	}
}

// ResolvePrefs is what the page says it can decode and what the user wants. Codec lists are families in
// order of preference ("vp9", "av01", "avc1" / "opus", "aac"), already filtered with canPlayType().
type ResolvePrefs struct {
	Height      int      `json:"height"`
	VideoCodecs []string `json:"videoCodecs"`
	AudioCodecs []string `json:"audioCodecs"`
}

// Search looks up videos (kind "video") or songs (kind "music"). One search runs at a time: a new one
// cancels the previous, and CancelSearch stops it (it then fails with "cancelado").
func (a *App) Search(query, kind string, limit int) ([]ytdlp.Entry, error) {
	env, err := a.bins.Env()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(a.ctx, 60*time.Second)
	defer cancel()

	if a.play != nil {
		a.play.searchMu.Lock()
		if a.play.searchCancel != nil {
			a.play.searchCancel()
		}
		a.play.searchCancel = cancel
		a.play.searchMu.Unlock()
	}
	return ytdlp.Search(ctx, env, query, kind, limit)
}

func (a *App) CancelSearch() {
	if a.play == nil {
		return
	}
	a.play.searchMu.Lock()
	defer a.play.searchMu.Unlock()
	if a.play.searchCancel != nil {
		a.play.searchCancel()
	}
}

// Resolve finds the streams for one video and returns local URLs the page can play. Results are cached, so
// changing quality or resolving the next track ahead of time is cheap.
func (a *App) Resolve(id string, p ResolvePrefs) (player.StreamInfo, error) {
	if a.play == nil {
		return player.StreamInfo{}, errors.New("el reproductor no está disponible")
	}
	ctx, cancel := context.WithTimeout(a.ctx, 90*time.Second)
	defer cancel()
	return a.play.res.Resolve(ctx, id, ytdlp.Prefs{Height: p.Height, VideoCodecs: p.VideoCodecs, AudioCodecs: p.AudioCodecs})
}

// Prefetch resolves a video in the background so that playing it later skips the 3 s yt-dlp run and part of
// the few seconds YouTube makes a fresh link wait. It is best effort: when two are already running the call
// is dropped, and any error is left for the real Resolve to report.
func (a *App) Prefetch(id string) {
	if a.play == nil || id == "" {
		return
	}
	select {
	case a.play.warm <- struct{}{}:
	default:
		return
	}
	go func() {
		defer func() { <-a.play.warm }()
		ctx, cancel := context.WithTimeout(a.ctx, 90*time.Second)
		defer cancel()
		a.play.res.Warm(ctx, id)
	}()
}
