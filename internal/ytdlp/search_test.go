package ytdlp

import (
	"os"
	"strings"
	"testing"
)

func TestParseSearchVideo(t *testing.T) {
	b, _ := os.ReadFile("testdata/search_video.json")
	es, err := ParseSearch(b, KindVideo)
	if err != nil {
		t.Fatal(err)
	}
	if len(es) != 2 { // the third result is a live stream and is dropped
		t.Fatalf("want 2 results, got %d: %+v", len(es), es)
	}
	e := es[0]
	if e.ID == "" || e.Title == "" || e.Channel == "" || e.Duration <= 0 || e.ViewCount <= 0 || !strings.HasPrefix(e.Thumbnail, "https://i.ytimg.com/") {
		t.Errorf("incomplete entry: %+v", e)
	}
	if e.URL != "https://www.youtube.com/watch?v="+e.ID {
		t.Errorf("url = %s", e.URL)
	}
}

func TestParseSearchMusicKeepsOnlySongs(t *testing.T) {
	b, _ := os.ReadFile("testdata/search_music.json")
	es, err := ParseSearch(b, KindMusic)
	if err != nil {
		t.Fatal(err)
	}
	if len(es) == 0 || len(es) >= 9 {
		t.Fatalf("albums/artists (YoutubeTab) must be dropped; got %d of 9", len(es))
	}
	for _, e := range es {
		if len(e.ID) != 11 || strings.Contains(e.URL, "music.youtube.com") || e.Thumbnail == "" || e.Title == "" {
			t.Errorf("bad song entry: %+v", e)
		}
	}
}

func TestParseSearchDedupesAndRejectsGarbage(t *testing.T) {
	es, err := ParseSearch([]byte(`{"entries":[{"id":"aaaaaaaaaaa","title":"a"},{"id":"aaaaaaaaaaa","title":"dup"},{"id":"short","title":"x"}]}`), KindVideo)
	if err != nil || len(es) != 1 || es[0].Title != "a" {
		t.Fatalf("got %+v %v", es, err)
	}
	if _, err := ParseSearch([]byte(`not json`), KindVideo); err == nil {
		t.Error("invalid JSON must be an error")
	}
}
