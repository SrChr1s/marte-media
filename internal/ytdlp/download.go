package ytdlp

import (
	"bufio"
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	progressPrefix = "MARTE|"
	filePrefix     = "MARTEFILE|"
)

type Options struct {
	URL          string
	Dir          string
	Mode         string // video | audio
	Height       int    // 0 = best
	VideoFormat  string // mp4 | webm
	AudioFormat  string // mp3 | m4a | opus
	AudioBitrate int    // kbps
	FFmpeg       string
}

type Progress struct {
	Phase      string // downloading | processing
	Percent    float64
	Downloaded int64
	Total      int64
	Speed      float64 // bytes/s
	ETA        int     // seconds, -1 unknown
}

// BuildArgs returns the yt-dlp arguments for the given options.
func BuildArgs(o Options) []string {
	args := []string{
		"--newline", "--no-colors", "--no-warnings", "--no-playlist",
		// --print implies --quiet, which also hides progress; ask for it back so --progress-template prints.
		"--progress",
		"-o", filepath.Join(o.Dir, "%(title).150B [%(id)s].%(ext)s"),
		"--progress-template", "download:" + progressPrefix +
			"%(progress.status)s|%(progress.downloaded_bytes)s|%(progress.total_bytes)s|" +
			"%(progress.total_bytes_estimate)s|%(progress.speed)s|%(progress.eta)s",
		"--print", "after_move:" + filePrefix + "%(filepath)s",
	}
	if o.FFmpeg != "" {
		args = append(args, "--ffmpeg-location", o.FFmpeg)
	}
	if o.Mode == "audio" {
		args = append(args, "-x", "--audio-format", o.AudioFormat, "--embed-metadata")
		if o.AudioBitrate > 0 {
			args = append(args, "--audio-quality", strconv.Itoa(o.AudioBitrate)+"K")
		}
		// Thumbnails can't be embedded in opus by ffmpeg's mux path; skip there.
		if o.AudioFormat != "opus" {
			args = append(args, "--embed-thumbnail")
		}
	} else {
		args = append(args, "-f", videoSelector(o), "--merge-output-format", o.VideoFormat)
	}
	return append(args, "--", o.URL)
}

func videoSelector(o Options) string {
	h := ""
	if o.Height > 0 {
		h = fmt.Sprintf("[height<=%d]", o.Height)
	}
	if o.VideoFormat == "mp4" {
		// Prefer H.264 + AAC for broad player compatibility, fall back to anything.
		return fmt.Sprintf("bv*%s[vcodec^=avc1]+ba[ext=m4a]/bv*%s+ba/b%s", h, h, h)
	}
	return fmt.Sprintf("bv*%s+ba/b%s", h, h)
}

// ParseLine interprets one stdout line. It returns a progress update, a final
// file path, or neither.
func ParseLine(line string) (p *Progress, file string) {
	line = strings.TrimSpace(line)
	switch {
	case strings.HasPrefix(line, filePrefix):
		return nil, strings.TrimPrefix(line, filePrefix)
	case strings.HasPrefix(line, progressPrefix):
		f := strings.Split(strings.TrimPrefix(line, progressPrefix), "|")
		if len(f) < 6 {
			return nil, ""
		}
		if f[0] == "finished" {
			return &Progress{Phase: "processing", Percent: 100, ETA: -1}, ""
		}
		down := atoi64(f[1])
		total := atoi64(f[2])
		if total == 0 {
			total = atoi64(f[3])
		}
		pr := &Progress{
			Phase:      "downloading",
			Downloaded: down,
			Total:      total,
			Speed:      atof(f[4]),
			ETA:        -1,
		}
		if eta, err := strconv.Atoi(strings.TrimSpace(f[5])); err == nil {
			pr.ETA = eta
		}
		if total > 0 {
			pr.Percent = float64(down) * 100 / float64(total)
		}
		return pr, ""
	case strings.HasPrefix(line, "[Merger]"), strings.HasPrefix(line, "[ExtractAudio]"),
		strings.HasPrefix(line, "[EmbedThumbnail]"), strings.HasPrefix(line, "[Metadata]"):
		return &Progress{Phase: "processing", Percent: 100, ETA: -1}, ""
	}
	return nil, ""
}

func atoi64(s string) int64 {
	s = strings.TrimSpace(s)
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int64(f)
	}
	return 0
}

func atof(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}

// Download runs yt-dlp and reports progress. It returns the final file path.
// Cancelling ctx kills the process.
func Download(ctx context.Context, env Env, o Options, onProgress func(Progress)) (string, error) {
	if o.FFmpeg == "" {
		o.FFmpeg = env.FFmpeg
	}
	cmd := newCmd(ctx, env.Bin, append(append([]string{}, env.JSArgs...), BuildArgs(o)...)...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return "", err
	}
	var file string
	sc := bufio.NewScanner(stdout)
	for sc.Scan() {
		p, f := ParseLine(sc.Text())
		if f != "" {
			file = f
		}
		if p != nil && onProgress != nil {
			onProgress(*p)
		}
	}
	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", fmt.Errorf("%s", lastLine(stderr.String()))
	}
	return file, nil
}
