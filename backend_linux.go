//go:build linux

package main

import "os"

// WebKitGTK paints through the GPU (DMABUF) by default, which is what keeps animations at 60fps.
// On X11 with an NVIDIA GPU that path can leave the window blank, so there we turn it off. That
// falls back to CPU painting, hence lowGraphics(): the UI drops its decorative animations.
// Must run before GTK initialises, hence init(). Set WEBKIT_DISABLE_DMABUF_RENDERER yourself
// ("0" keeps the GPU path, "1" forces the fallback) to override.
func init() {
	if os.Getenv("WEBKIT_DISABLE_DMABUF_RENDERER") != "" {
		return
	}
	onX11 := os.Getenv("GDK_BACKEND") == "x11" || os.Getenv("WAYLAND_DISPLAY") == ""
	if onX11 {
		os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
	}
}

// lowGraphics reports whether the webview is painting on the CPU.
func lowGraphics() bool {
	v := os.Getenv("WEBKIT_DISABLE_DMABUF_RENDERER")
	return v != "" && v != "0"
}
