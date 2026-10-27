package store

import (
	"testing"
	"time"
)

func TestSettingsDefaultsAndRoundTrip(t *testing.T) {
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Settings(); got.MaxConcurrent != 2 || got.AudioFormat != "mp3" {
		t.Fatalf("unexpected defaults: %+v", got)
	}
	cfg := s.Settings()
	cfg.MaxConcurrent = 4
	cfg.DownloadDir = "/tmp/x"
	if err := s.SaveSettings(cfg); err != nil {
		t.Fatal(err)
	}
	if got := s.Settings(); got.MaxConcurrent != 4 || got.DownloadDir != "/tmp/x" {
		t.Fatalf("not persisted: %+v", got)
	}
}

func TestHistoryOrderAndDelete(t *testing.T) {
	s, _ := New(t.TempDir())
	_ = s.AddHistory(HistoryEntry{ID: "a", FinishedAt: time.Now()})
	_ = s.AddHistory(HistoryEntry{ID: "b", FinishedAt: time.Now()})
	h := s.History()
	if len(h) != 2 || h[0].ID != "b" {
		t.Fatalf("want newest first, got %+v", h)
	}
	_ = s.DeleteHistory("b")
	h = s.History()
	if len(h) != 1 || h[0].ID != "a" {
		t.Fatalf("delete failed: %+v", h)
	}
}
