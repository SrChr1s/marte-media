package main

import (
	"context"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"marte-desc/internal/queue"
	"marte-desc/internal/store"
	"marte-desc/internal/ytdlp"
)

// App is bound to the frontend; every exported method is callable from Svelte.
type App struct {
	ctx   context.Context
	store *store.Store
	bins  *ytdlp.Binaries
	queue *queue.Queue

	probeMu     sync.Mutex
	probeCancel context.CancelFunc // cancels the analysis in flight, if any

	play *playState // search + local stream server, see app_player.go
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	dir, err := store.DefaultDir()
	if err == nil {
		a.store, err = store.New(dir)
	}
	if err != nil {
		runtime.LogErrorf(ctx, "no se pudo inicializar el almacenamiento: %v", err)
		return
	}
	a.bins = ytdlp.NewBinaries(dir)
	a.queue = queue.New(queue.Deps{
		Env:       a.bins.Env,
		MaxActive: func() int { return a.store.Settings().MaxConcurrent },
		OnUpdate:  func(j queue.Job) { runtime.EventsEmit(ctx, "job:update", j) },
		OnDone:    a.recordHistory,
	})
	a.startPlayer()
}

func (a *App) recordHistory(j queue.Job) {
	format := j.Request.VideoFormat
	if j.Request.Mode == "audio" {
		format = j.Request.AudioFormat
	}
	_ = a.store.AddHistory(store.HistoryEntry{
		ID: j.ID, URL: j.Request.URL, Title: j.Request.Title, Channel: j.Request.Channel,
		Thumbnail: j.Request.Thumbnail, Mode: j.Request.Mode, Format: format,
		Path: j.Path, FinishedAt: time.Now(),
	})
	runtime.EventsEmit(a.ctx, "history:changed")
}

// LowGraphics is true when the webview paints on the CPU; the UI then drops decorative animations.
func (a *App) LowGraphics() bool { return lowGraphics() }

// Probe returns metadata for a video or playlist URL.
func (a *App) Probe(url string) (ytdlp.MediaInfo, error) {
	env, err := a.bins.Env()
	if err != nil {
		return ytdlp.MediaInfo{}, err
	}
	ctx, cancel := context.WithTimeout(a.ctx, 60*time.Second)
	defer cancel()

	// One analysis at a time: a new one cancels the previous.
	a.probeMu.Lock()
	if a.probeCancel != nil {
		a.probeCancel()
	}
	a.probeCancel = cancel
	a.probeMu.Unlock()

	return ytdlp.Probe(ctx, env, url)
}

// CancelProbe stops the analysis in flight; its Probe call fails with "cancelado".
func (a *App) CancelProbe() {
	a.probeMu.Lock()
	defer a.probeMu.Unlock()
	if a.probeCancel != nil {
		a.probeCancel()
	}
}

// Enqueue adds one job per request; empty Dir/format fields fall back to the saved settings.
func (a *App) Enqueue(reqs []queue.Request) []string {
	cfg := a.store.Settings()
	for i := range reqs {
		r := &reqs[i]
		if r.Dir == "" {
			r.Dir = cfg.DownloadDir
		}
		if r.Mode == "" {
			r.Mode = cfg.Mode
		}
		if r.VideoFormat == "" {
			r.VideoFormat = cfg.VideoFormat
		}
		if r.AudioFormat == "" {
			r.AudioFormat = cfg.AudioFormat
		}
		if r.AudioBitrate == 0 {
			r.AudioBitrate = cfg.AudioBitrate
		}
	}
	return a.queue.Enqueue(reqs)
}

func (a *App) ListJobs() []queue.Job { return a.queue.List() }
func (a *App) Cancel(id string)      { a.queue.Cancel(id) }
func (a *App) Retry(id string)       { a.queue.Retry(id) }
func (a *App) ClearFinished()        { a.queue.ClearFinished() }

func (a *App) History() []store.HistoryEntry { return a.store.History() }
func (a *App) DeleteHistory(id string) error { return a.store.DeleteHistory(id) }

func (a *App) GetSettings() store.Settings { return a.store.Settings() }
func (a *App) SaveSettings(s store.Settings) error {
	return a.store.SaveSettings(s)
}

// PickFolder opens a native directory chooser; returns "" if the user cancels.
func (a *App) PickFolder(start string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Carpeta de descargas", DefaultDirectory: start,
	})
}

// OpenFile opens a file with the default app.
func (a *App) OpenFile(path string) error {
	switch goruntime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	}
	return exec.Command("xdg-open", path).Start()
}

// OpenFolder shows the folder holding path (selecting the file where supported).
func (a *App) OpenFolder(path string) error {
	switch goruntime.GOOS {
	case "windows":
		return exec.Command("explorer", "/select,"+filepath.Clean(path)).Start()
	case "darwin":
		return exec.Command("open", "-R", path).Start()
	}
	return exec.Command("xdg-open", filepath.Dir(path)).Start()
}

func (a *App) YtdlpStatus() ytdlp.Status { return a.bins.Status(a.ctx) }

func (a *App) InstallYtdlp() error {
	ctx, cancel := context.WithTimeout(a.ctx, 5*time.Minute)
	defer cancel()
	return a.bins.Install(ctx)
}

func (a *App) UpdateYtdlp() (string, error) {
	ctx, cancel := context.WithTimeout(a.ctx, 2*time.Minute)
	defer cancel()
	return a.bins.Update(ctx)
}
