package ytdlp

import (
	"slices"
	"testing"
)

func TestParseLineProgress(t *testing.T) {
	p, f := ParseLine("MARTE|downloading|5242880|10485760|NA|1048576.5|5")
	if f != "" || p == nil {
		t.Fatalf("expected progress, got %v %q", p, f)
	}
	if p.Percent != 50 || p.Total != 10485760 || p.ETA != 5 || p.Speed != 1048576.5 {
		t.Fatalf("bad progress: %+v", p)
	}
}

func TestParseLineEstimateAndUnknownETA(t *testing.T) {
	p, _ := ParseLine("MARTE|downloading|100|NA|400|NA|NA")
	if p == nil || p.Total != 400 || p.Percent != 25 || p.ETA != -1 {
		t.Fatalf("bad progress: %+v", p)
	}
}

func TestParseLineFileAndPhases(t *testing.T) {
	if _, f := ParseLine("MARTEFILE|/tmp/a [x].mp4"); f != "/tmp/a [x].mp4" {
		t.Fatalf("file = %q", f)
	}
	if p, _ := ParseLine("[Merger] Merging formats into \"a.mp4\""); p == nil || p.Phase != "processing" {
		t.Fatalf("merger: %+v", p)
	}
	if p, f := ParseLine("[youtube] Extracting URL"); p != nil || f != "" {
		t.Fatal("noise should be ignored")
	}
}

func TestBuildArgsVideo(t *testing.T) {
	a := BuildArgs(Options{URL: "u", Dir: "/d", Mode: "video", Height: 720, VideoFormat: "mp4", FFmpeg: "/usr/bin/ffmpeg"})
	for _, want := range []string{"--merge-output-format", "mp4", "--ffmpeg-location", "--no-playlist", "--progress"} {
		if !slices.Contains(a, want) {
			t.Errorf("missing %q in %v", want, a)
		}
	}
	i := slices.Index(a, "-f")
	if i < 0 || a[i+1] != "bv*[height<=720][vcodec^=avc1]+ba[ext=m4a]/bv*[height<=720]+ba/b[height<=720]" {
		t.Errorf("bad selector: %v", a)
	}
	if a[len(a)-2] != "--" || a[len(a)-1] != "u" {
		t.Errorf("url must follow --: %v", a)
	}
}

func TestBuildArgsAudio(t *testing.T) {
	a := BuildArgs(Options{URL: "u", Dir: "/d", Mode: "audio", AudioFormat: "mp3", AudioBitrate: 192})
	for _, want := range []string{"-x", "--audio-format", "mp3", "192K", "--embed-thumbnail"} {
		if !slices.Contains(a, want) {
			t.Errorf("missing %q in %v", want, a)
		}
	}
	o := BuildArgs(Options{URL: "u", Dir: "/d", Mode: "audio", AudioFormat: "opus"})
	if slices.Contains(o, "--embed-thumbnail") {
		t.Error("opus should skip thumbnail embedding")
	}
}

func TestParseInfoVideoAndPlaylist(t *testing.T) {
	v, err := ParseInfo([]byte(`{"id":"abc","title":"T","channel":"C","duration":61,"thumbnail":"th",
		"webpage_url":"https://y/watch?v=abc",
		"formats":[{"height":360,"vcodec":"avc1"},{"height":1080,"vcodec":"vp9"},{"height":1080,"vcodec":"avc1"},{"height":0,"vcodec":"none"}]}`))
	if err != nil || v.Kind != "video" || !slices.Equal(v.Heights, []int{1080, 360}) || v.Title != "T" {
		t.Fatalf("video: %+v %v", v, err)
	}
	pl, err := ParseInfo([]byte(`{"_type":"playlist","id":"PL","title":"List","entries":[{"id":"a","title":"A"},{"id":"b","title":"B","url":"https://y/b"}]}`))
	if err != nil || pl.Kind != "playlist" || pl.PlaylistN != 2 || pl.Entries[0].URL != "https://www.youtube.com/watch?v=a" || pl.Entries[1].URL != "https://y/b" {
		t.Fatalf("playlist: %+v %v", pl, err)
	}
}
