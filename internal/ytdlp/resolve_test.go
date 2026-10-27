package ytdlp

import (
	"os"
	"slices"
	"testing"
)

func loadMedia(t *testing.T) Media {
	t.Helper()
	b, err := os.ReadFile("testdata/formats_bbb.json")
	if err != nil {
		t.Fatal(err)
	}
	m, err := ParseMedia(b)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// What a Linux webview with GStreamer can decode (no H.264).
var linuxPrefs = Prefs{Height: 720, VideoCodecs: []string{"vp9", "av01"}, AudioCodecs: []string{"opus", "aac"}}

func TestParseMediaReal(t *testing.T) {
	m := loadMedia(t)
	if m.ID != "aqz-KE-bpKQ" || m.Duration < 500 || len(m.Formats) < 40 {
		t.Fatalf("unexpected media: %+v (%d formats)", m.ID, len(m.Formats))
	}
	if m.ExpiresAt.IsZero() {
		t.Fatal("expiry should be read from the signed URLs")
	}
}

func TestPickLinux720(t *testing.T) {
	p := Pick(loadMedia(t), linuxPrefs)
	if p.Video == nil || p.Audio == nil {
		t.Fatalf("missing picks: %+v", p)
	}
	if p.Video.ID != "247" { // VP9 720p30 over https (not 302 which is 60 fps, not the HLS 612, not AV1 398)
		t.Errorf("video = %s (%s %dp %s)", p.Video.ID, p.Video.VCodec, p.Video.Height, p.Video.Protocol)
	}
	if p.Audio.ID != "251" { // Opus 128k, not the DRC variant, not AAC
		t.Errorf("audio = %s (%s)", p.Audio.ID, p.Audio.ACodec)
	}
	if p.Muxed != nil {
		t.Error("no muxed fallback expected when a separate video exists")
	}
	if !slices.Contains(p.Heights, 1080) || !slices.Contains(p.Heights, 2160) || p.Heights[0] != 2160 {
		t.Errorf("heights should list the full ladder, tallest first: %v", p.Heights)
	}
}

func TestPickHeights(t *testing.T) {
	m := loadMedia(t)
	cases := map[int]int{1080: 1080, 2160: 2160, 500: 480, 100: 144} // wanted -> chosen (never above if avoidable)
	for want, got := range cases {
		p := linuxPrefs
		p.Height = want
		if v := Pick(m, p).Video; v == nil || v.Height != got {
			t.Errorf("want %dp: got %+v", want, v)
		}
	}
	p := linuxPrefs
	p.Height = 1080
	if v := Pick(m, p).Video; v.ID != "303" { // VP9 1080p60 over AV1 399 because VP9 comes first
		t.Errorf("1080 should be VP9 (303), got %s", v.ID)
	}
}

func TestPickAudioPreferences(t *testing.T) {
	m := loadMedia(t)
	p := Pick(m, Prefs{VideoCodecs: []string{"vp9"}, AudioCodecs: []string{"aac"}})
	if p.Audio == nil || p.Audio.ID != "140" { // stereo AAC 129k; 256/258 are 5.1, 139 is lower quality
		t.Errorf("aac audio = %+v", p.Audio)
	}
	if p := Pick(m, Prefs{VideoCodecs: []string{"vp9"}}); p.Audio != nil {
		t.Errorf("no audio codec allowed -> no audio, got %+v", p.Audio)
	}
}

func TestPickWindowsH264(t *testing.T) {
	// H.264 first, as WebView2 would report: the mp4 avc1 stream is chosen, not VP9.
	p := Pick(loadMedia(t), Prefs{Height: 720, VideoCodecs: []string{"avc1", "vp9"}, AudioCodecs: []string{"aac", "opus"}})
	if p.Video == nil || p.Video.ID != "136" || p.Audio == nil || p.Audio.ID != "140" {
		t.Errorf("unexpected picks: video=%+v audio=%+v", p.Video, p.Audio)
	}
}

func TestPickMuxedFallbackAndNothing(t *testing.T) {
	m := Media{Formats: []Format{
		{ID: "18", Ext: "mp4", VCodec: "avc1.42001E", ACodec: "mp4a.40.2", Protocol: "https", Height: 360, URL: "u"},
		{ID: "hls", Ext: "mp4", VCodec: "avc1", ACodec: "mp4a.40.2", Protocol: "m3u8_native", Height: 1080, URL: "u"},
	}}
	p := Pick(m, Prefs{VideoCodecs: []string{"avc1"}, AudioCodecs: []string{"aac"}})
	if p.Video != nil || p.Muxed == nil || p.Muxed.ID != "18" || !slices.Equal(p.Heights, []int{360}) {
		t.Errorf("want muxed 18 only, got %+v", p)
	}
	// a webview that decodes none of the codecs gets nothing to play
	if p := Pick(m, Prefs{VideoCodecs: []string{"vp9"}, AudioCodecs: []string{"opus"}}); p.Video != nil || p.Muxed != nil || p.Audio != nil || len(p.Heights) != 0 {
		t.Errorf("nothing decodable should give nothing: %+v", p)
	}
}

func TestParseMediaRejectsLiveAndEmpty(t *testing.T) {
	if _, err := ParseMedia([]byte(`{"id":"x","live_status":"is_live","formats":[{"format_id":"1","url":"u"}]}`)); err == nil {
		t.Error("live streams must be rejected")
	}
	if _, err := ParseMedia([]byte(`{"id":"x","formats":[]}`)); err == nil {
		t.Error("no formats must be an error")
	}
}
