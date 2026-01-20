package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"maoni/app/core/auth"
	"maoni/app/core/bq"
	"maoni/app/core/fire"
	"maoni/app/core/mid"
	"maoni/app/core/web"
	"maoni/app/core/webx"
	"maoni/app/modules/admin"
	"maoni/app/modules/base"
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
	l := log.New(os.Stdout, "MAONI : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
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
		ServiceAcc      string `conf:",mask"`
		FsDbID          string `conf:"default:(default)"`
		BqDatasetID     string `conf:"default:MAONI"`
	}{}

	// Initial parse to capture command-line flags like --dev
	help, err := conf.Parse("", &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	confStr, _ := conf.String(&cfg)

	// If in dev mode, load env.dev.yaml and re-parse config
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

		// Re-parse to load values from the environment variables set from the yaml file.
		if _, err = conf.Parse("", &cfg); err != nil {
			return fmt.Errorf("parsing config after loading env.dev.yaml: %w", err)
		}

		if str, err := conf.String(&cfg); err == nil {
			confStr = str
		}
	}

	l.Println("Config:\n", confStr)

	// Manually check for required fields after all parsing attempts.
	if cfg.GoogleKey == "" {
		return errors.New("required field GoogleKey is missing value")
	}
	if cfg.GoogleSecret == "" {
		return errors.New("required field GoogleSecret is missing value")
	}
	if cfg.GoogleProjectID == "" {
		return errors.New("required field GoogleProjectID is missing value")
	}
	if cfg.Host == "" {
		return errors.New("required field Host is missing value")
	}
	if cfg.SessionSecret == "" {
		return errors.New("required field SessionSecret is missing value")
	}

	// --- Services ---
	ctx := context.Background()
	fsDb, err := fire.Store(ctx, cfg.GoogleProjectID, cfg.FsDbID)
	if err != nil {
		return fmt.Errorf("failed to connect to firestore: %w", err)
	}
	defer fsDb.Close()

	bqClient, err := bq.NewClient(ctx, cfg.GoogleProjectID)
	if err != nil {
		return fmt.Errorf("failed to create bigquery client: %w", err)
	}
	defer bqClient.Close()

	if err := bq.EnsureSchema(ctx, bqClient, cfg.BqDatasetID); err != nil {
		return fmt.Errorf("failed to ensure bigquery schema: %w", err)
	}
	surveyStore := bq.NewSurveyStore(bqClient, cfg.BqDatasetID)

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

	// Goth needs a session store; this one is temporary for the auth dance.
	cookieStore := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	cookieStore.Options.MaxAge = int(auth.SessionDuration.Seconds()) // Use the project's session duration
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

	// Initialize modules/routes
	base.InitModule(l, app, sessionStore, surveyStore)
	admin.InitModule(l, app, sessionStore, surveyStore)

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
		serverErr <- server.ListenAndServe()
	}()

	l.Printf("Starting server on %s", cfg.Host)

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %s", err)
	case sig := <-shutdown:
		l.Printf("starting server shutdown: %s", sig)
		defer l.Println("server shutdown complete")

		ctx, cancel := context.WithTimeout(context.Background(), web.DefaultShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			server.Close()
			return fmt.Errorf("could not gracefully shutdown server: %s", err)
		}
	}

	return nil
}
