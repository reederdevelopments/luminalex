package main

import (
	"context"
	"controlroom/app/backend/auth"
	"controlroom/app/backend/mid"
	"controlroom/app/backend/web"
	"controlroom/app/backend/webx"
	costing "controlroom/app/frontend/costing"
	base "controlroom/app/frontend/home"
	kb "controlroom/app/frontend/kb"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ardanlabs/conf/v3"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"gopkg.in/yaml.v2"
)

//go:embed assets
var assets embed.FS

func main() {
	l := log.New(os.Stdout, "CONTROLROOM : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	if err := run(l); err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
}

func run(l *log.Logger) error {
	// --- Configuration ---
	cfg := struct {
		Port            int    `conf:"default:3080"`
		Dev             bool   `conf:"flag:dev"`
		GoogleKey       string `conf:",noumask"`
		GoogleSecret    string `conf:",mask"`
		GoogleProjectID string `conf:",noumask"`
		Host            string `conf:",noumask"`
		SessionSecret   string `conf:",mask"`
		FsDbID          string `conf:"default:(default)"`
	}{}

	help, err := conf.Parse("", &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	confStr, _ := conf.String(&cfg)

	if cfg.Dev {
		f, err := os.Open("env.dev.yaml")
		if err != nil {
			return fmt.Errorf("opening env.dev.yaml: %w", err)
		}
		defer f.Close()

		env := map[string]string{}
		if err := yaml.NewDecoder(f).Decode(&env); err != nil {
			return fmt.Errorf("decoding env.dev.yaml: %w", err)
		}

		for k, v := range env {
			os.Setenv(k, v)
		}

		if _, err = conf.Parse("", &cfg); err != nil {
			return fmt.Errorf("parsing config after loading env.dev.yaml: %w", err)
		}

		if str, err := conf.String(&cfg); err == nil {
			confStr = str
		}
	}

	l.Println("Config:\n", confStr)

	if cfg.GoogleKey == "" || cfg.GoogleSecret == "" || cfg.GoogleProjectID == "" || cfg.Host == "" || cfg.SessionSecret == "" {
		return errors.New("missing required configuration fields")
	}

	ctx := context.Background()

	// --- Services ---
	fsDb, err := auth.NewFirestoreClient(ctx, cfg.GoogleProjectID, cfg.FsDbID)
	if err != nil {
		return fmt.Errorf("failed to connect to firestore: %w", err)
	}
	defer fsDb.Close()

	// --- Authentication ---
	authService := auth.NewService(cfg.SessionSecret)
	sessionStore := auth.NewFireStore(l, fsDb, authService, cfg.Dev)

	goth.UseProviders(
		google.New(
			cfg.GoogleKey,
			cfg.GoogleSecret,
			fmt.Sprintf("%s/auth/google/callback", cfg.Host),
			"email", "profile",
		),
	)

	cookieStore := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	cookieStore.Options.MaxAge = int(auth.SessionDuration.Seconds())
	cookieStore.Options.Path = "/"
	cookieStore.Options.HttpOnly = true
	cookieStore.Options.Secure = !cfg.Dev
	gothic.Store = cookieStore

	// --- HTTP Server ---
	app := web.NewApp()

	// --- Static Assets ---
	staticHandler := webx.StaticHandler(assets, l, cfg.Dev)
	app.HandleStd(
		http.MethodGet,
		"/assets/*",
		staticHandler,
		mid.Log(l),
		mid.CatchErr(l),
		mid.CatchPanic(),
		mid.TryGzip,
	)

	// Initialize base module (Auth & Core routes)
	// We now pass coreDBs down the chain.
	base.InitModule(l, app, sessionStore)
	costing.InitModule(l, app, sessionStore)
	kb.InitModule(l, app, sessionStore)

	// Start Server
	host := fmt.Sprintf("0.0.0.0:%d", cfg.Port)
	server := http.Server{
		Addr:    host,
		Handler: app,
	}

	shutdown := make(chan os.Signal, 1)
	serverErr := make(chan error, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		l.Printf("Starting server on %s", cfg.Host)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-shutdown:
		l.Printf("starting server shutdown: %s", sig)
		defer l.Println("server shutdown complete")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), web.DefaultShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			server.Close()
			return fmt.Errorf("could not gracefully shutdown server: %s", err)
		}
	}

	return nil
}
