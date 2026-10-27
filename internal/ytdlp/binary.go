package ytdlp

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const releaseBase = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/"

type Status struct {
	Installed bool   `json:"installed"`
	Path      string `json:"path"`
	Version   string `json:"version"`
	FFmpeg    string `json:"ffmpeg"` // empty if missing
	// JSRuntime is the JavaScript engine yt-dlp will use to solve YouTube's challenges
	// ("deno", "node", "bun" or "quickjs"); empty if none was found.
	JSRuntime string `json:"jsRuntime"`
	JSPath    string `json:"jsPath"`
}

// Env is everything a yt-dlp invocation needs besides its own arguments.
type Env struct {
	Bin    string
	FFmpeg string
	JSArgs []string // --js-runtimes ... (yt-dlp only enables deno by default)
}

// Binaries locates yt-dlp and ffmpeg. Lookup order, first hit wins:
//  1. the copy the app downloaded into binDir (so updates beat the bundled one),
//  2. the copy shipped next to the executable (installer),
//  3. PATH.
type Binaries struct {
	binDir string
}

func NewBinaries(configDir string) *Binaries {
	return &Binaries{binDir: filepath.Join(configDir, "bin")}
}

func exeName(base string) string {
	if runtime.GOOS == "windows" {
		return base + ".exe"
	}
	return base
}

// releaseAsset is the yt-dlp release file for this OS/arch.
func releaseAsset() (string, error) {
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "linux/amd64":
		return "yt-dlp_linux", nil
	case "linux/arm64":
		return "yt-dlp_linux_aarch64", nil
	case "windows/amd64":
		return "yt-dlp.exe", nil
	case "windows/arm64":
		return "yt-dlp_arm64.exe", nil
	}
	return "", fmt.Errorf("plataforma no soportada: %s/%s", runtime.GOOS, runtime.GOARCH)
}

func (b *Binaries) managedPath() string { return filepath.Join(b.binDir, exeName("yt-dlp")) }

// bundledDirs are the places the installer puts yt-dlp and ffmpeg.
var bundledDirs = func() []string {
	exe, err := os.Executable()
	if err != nil {
		return nil
	}
	dir := filepath.Dir(exe)
	return []string{dir, filepath.Join(dir, "bin")}
}

func isFile(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

// find returns the first existing candidate for name, falling back to PATH.
func (b *Binaries) find(name string, managed string) string {
	cands := []string{}
	if managed != "" {
		cands = append(cands, managed)
	}
	for _, d := range bundledDirs() {
		cands = append(cands, filepath.Join(d, exeName(name)))
	}
	for _, c := range cands {
		if isFile(c) {
			return c
		}
	}
	p, _ := exec.LookPath(name)
	return p
}

func (b *Binaries) YtDlp() (string, error) {
	if p := b.find("yt-dlp", b.managedPath()); p != "" {
		return p, nil
	}
	return "", fmt.Errorf("yt-dlp no está instalado")
}

func (b *Binaries) FFmpeg() string { return b.find("ffmpeg", "") }

// jsRuntimes in yt-dlp's own order of preference.
var jsRuntimes = []string{"deno", "node", "quickjs", "bun"}

// JSRuntime returns the first available JavaScript runtime: the one shipped with the app (deno) or one
// on PATH. yt-dlp needs it to extract from YouTube, and it only looks for deno unless told otherwise.
func (b *Binaries) JSRuntime() (name, path string) {
	for _, n := range jsRuntimes {
		if p := b.find(n, ""); p != "" {
			return n, p
		}
	}
	return "", ""
}

// jsArgs enables every runtime we can find, deno (bundled) first.
func (b *Binaries) jsArgs() []string {
	var args []string
	for _, n := range jsRuntimes {
		if p := b.find(n, ""); p != "" {
			args = append(args, "--js-runtimes", n+":"+p)
		}
	}
	return args
}

// Env resolves the tools once, for one yt-dlp call.
func (b *Binaries) Env() (Env, error) {
	bin, err := b.YtDlp()
	if err != nil {
		return Env{}, err
	}
	return Env{Bin: bin, FFmpeg: b.FFmpeg(), JSArgs: b.jsArgs()}, nil
}

func (b *Binaries) Status(ctx context.Context) Status {
	st := Status{FFmpeg: b.FFmpeg()}
	st.JSRuntime, st.JSPath = b.JSRuntime()
	path, err := b.YtDlp()
	if err != nil {
		return st
	}
	st.Installed, st.Path = true, path
	cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if out, err := newCmd(cctx, path, "--version").Output(); err == nil {
		st.Version = strings.TrimSpace(string(out))
	}
	return st
}

// Install downloads the latest yt-dlp release into binDir, verifying it against
// the release's SHA2-256SUMS. It is also how updates work.
func (b *Binaries) Install(ctx context.Context) error {
	asset, err := releaseAsset()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(b.binDir, 0o755); err != nil {
		return err
	}
	want, err := expectedSum(ctx, asset)
	if err != nil {
		return err
	}
	resp, err := get(ctx, releaseBase+asset)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	tmp := b.managedPath() + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	h := sha256.New()
	_, err = io.Copy(io.MultiWriter(f, h), resp.Body)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil && hex.EncodeToString(h.Sum(nil)) != want {
		err = fmt.Errorf("la suma SHA-256 de yt-dlp no coincide; descarga descartada")
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, b.managedPath())
}

// Update fetches the newest yt-dlp and reports the resulting version.
func (b *Binaries) Update(ctx context.Context) (string, error) {
	if err := b.Install(ctx); err != nil {
		return "", err
	}
	return "yt-dlp actualizado a " + b.Status(ctx).Version, nil
}

func get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("descarga falló (%s): HTTP %d", url, resp.StatusCode)
	}
	return resp, nil
}

// expectedSum reads the release's SHA2-256SUMS ("<hex>  <file>" lines).
func expectedSum(ctx context.Context, asset string) (string, error) {
	resp, err := get(ctx, releaseBase+"SHA2-256SUMS")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) == 2 && strings.TrimPrefix(f[1], "*") == asset {
			return strings.ToLower(f[0]), nil
		}
	}
	return "", fmt.Errorf("no hay suma SHA-256 publicada para %s", asset)
}
