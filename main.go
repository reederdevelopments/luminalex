package main

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"luminalex/core"
)

//go:embed assets
var assets embed.FS

func main() {
	_ = godotenv.Load()

	if err := run(); err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	app := core.NewApp()
	handler := core.NewHandler(app)
	assetHandler := http.FileServer(http.FS(assets))

	err := wails.Run(&options.App{
		Title:  "LuminaLex",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
			Middleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.HasPrefix(r.URL.Path, "/assets/") {
						assetHandler.ServeHTTP(w, r)
						return
					}
					if strings.HasPrefix(r.URL.Path, "/api/") {
						handler.ServeHTTP(w, r)
						return
					}
					if r.URL.Path == "/" || r.URL.Path == "/login" || r.URL.Path == "/home" || r.URL.Path == "/contacts" {
						handler.ServeHTTP(w, r)
						return
					}
					next.ServeHTTP(w, r)
				})
			},
		},
		OnStartup: app.Startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		return fmt.Errorf("starting wails app: %w", err)
	}

	return nil
}
