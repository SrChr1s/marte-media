package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	opts := &options.App{
		Title:     "Marte Media",
		Width:     1100,
		Height:    760,
		MinWidth:  820,
		MinHeight: 560,
		MaxWidth:  16384, // 0 would cap the window at the size of the monitor it started on
		MaxHeight: 16384,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 15, B: 15, A: 1},   // #0f0f0f
		Linux:            &linux.Options{ProgramName: "marte-media"}, // Wayland app_id, matches marte-media.desktop
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	}
	err := wails.Run(opts)

	if err != nil {
		println("Error:", err.Error())
	}
}
