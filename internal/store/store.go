package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Settings struct {
	DownloadDir   string `json:"downloadDir"`
	MaxConcurrent int    `json:"maxConcurrent"`
	Mode          string `json:"mode"`         // video | audio
	VideoHeight   int    `json:"videoHeight"`  // 0 = best
	VideoFormat   string `json:"videoFormat"`  // mp4 | webm
	AudioFormat   string `json:"audioFormat"`  // mp3 | m4a | opus
	AudioBitrate  int    `json:"audioBitrate"` // kbps
}

type HistoryEntry struct {
	ID         string    `json:"id"`
	URL        string    `json:"url"`
	Title      string    `json:"title"`
	Channel    string    `json:"channel"`
	Thumbnail  string    `json:"thumbnail"`
	Mode       string    `json:"mode"`
	Format     string    `json:"format"`
	Path       string    `json:"path"`
	FinishedAt time.Time `json:"finishedAt"`
}

type Store struct {
	dir string
	mu  sync.Mutex
}

// New creates the store under dir, creating it if needed.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

// DefaultDir is ~/.config/marte-media.
func DefaultDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "marte-media")
	// The app used to be called marte-desc: carry its settings and history over once.
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		_ = os.Rename(filepath.Join(base, "marte-desc"), dir)
	}
	return dir, nil
}

func (s *Store) Dir() string { return s.dir }

func defaultSettings() Settings {
	dl := "."
	if home, err := os.UserHomeDir(); err == nil {
		dl = filepath.Join(home, "Downloads")
	}
	return Settings{
		DownloadDir:   dl,
		MaxConcurrent: 2,
		Mode:          "video",
		VideoHeight:   1080,
		VideoFormat:   "mp4",
		AudioFormat:   "mp3",
		AudioBitrate:  192,
	}
}

func (s *Store) Settings() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	cfg := defaultSettings()
	_ = s.read("settings.json", &cfg)
	if cfg.MaxConcurrent < 1 {
		cfg.MaxConcurrent = 1
	}
	return cfg
}

func (s *Store) SaveSettings(cfg Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.write("settings.json", cfg)
}

// History returns entries newest first.
func (s *Store) History() []HistoryEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.history()
}

func (s *Store) AddHistory(e HistoryEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	h := append([]HistoryEntry{e}, s.history()...)
	return s.write("history.json", h)
}

func (s *Store) DeleteHistory(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	h := s.history()
	out := h[:0]
	for _, e := range h {
		if e.ID != id {
			out = append(out, e)
		}
	}
	return s.write("history.json", out)
}

func (s *Store) history() []HistoryEntry {
	h := []HistoryEntry{}
	_ = s.read("history.json", &h)
	return h
}

func (s *Store) read(name string, v any) error {
	b, err := os.ReadFile(filepath.Join(s.dir, name))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return json.Unmarshal(b, v)
}

// write is atomic: tmp file + rename.
func (s *Store) write(name string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(s.dir, name)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
