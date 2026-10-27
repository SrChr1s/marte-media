//go:build !windows

package ytdlp

import (
	"context"
	"os/exec"
	"syscall"
	"time"
)

// newCmd runs the command in its own process group and, on cancel, kills the
// whole group: yt-dlp's PyInstaller binary spawns a child, and ffmpeg is a
// grandchild, so killing only the parent would leave them running.
func newCmd(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 3 * time.Second
	return cmd
}
