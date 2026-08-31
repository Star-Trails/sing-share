package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"sing-share/services"
)

//go:embed all:frontend/dist
var assets embed.FS

func init() {
	application.RegisterEvent[string]("config-dropped")
}

func main() {
	shareService := services.NewShareService(firstFileArgument(os.Args[1:]))
	app := application.New(application.Options{
		Name:        "sing-share",
		Description: "Share sing-box profiles with QRS animated QR codes",
		Services: []application.Service{
			application.NewService(shareService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:                       "main",
		Title:                      "sing-share",
		Width:                      560,
		Height:                     720,
		MinWidth:                   480,
		MinHeight:                  620,
		InitialPosition:            application.WindowCentered,
		BackgroundColour:           application.NewRGB(245, 244, 240),
		URL:                        "/",
		EnableFileDrop:             true,
		DefaultContextMenuDisabled: true,
	})
	window.OnWindowEvent(events.Common.WindowFilesDropped, func(event *application.WindowEvent) {
		for _, path := range event.Context().DroppedFiles() {
			if strings.EqualFold(filepath.Ext(path), ".json") {
				app.Event.Emit("config-dropped", path)
				return
			}
		}
		files := event.Context().DroppedFiles()
		if len(files) > 0 {
			app.Event.Emit("config-dropped", files[0])
		}
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func firstFileArgument(args []string) string {
	for _, arg := range args {
		if arg != "" && !strings.HasPrefix(arg, "-") {
			return arg
		}
	}
	return ""
}
