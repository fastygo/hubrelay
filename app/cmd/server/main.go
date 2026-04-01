package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hubrelay-dashboard/internal/config"
	"hubrelay-dashboard/internal/content"
	"hubrelay-dashboard/internal/handlers"
	"hubrelay-dashboard/internal/middleware"
	"hubrelay-dashboard/internal/relay"
	"hubrelay-dashboard/internal/source"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	locales, err := content.AvailableLocales()
	if err != nil {
		log.Fatalf("discover locales: %v", err)
	}
	catalogs := make(map[string]content.Catalog, len(locales))
	for _, locale := range locales {
		catalogs[locale], err = content.Load(locale)
		if err != nil {
			log.Fatalf("load content for %s: %v", locale, err)
		}
	}

	var liveSource source.Source
	fixtureSources := map[string]source.Source{}
	var client *relay.Client
	switch cfg.DataSource {
	case config.DataSourceFixture:
		for _, locale := range locales {
			fixtureSources[locale], err = source.NewFixture(locale)
			if err != nil {
				log.Fatalf("create fixture source for %s: %v", locale, err)
			}
		}
	case config.DataSourceLive:
		client, err = relay.New(cfg)
		if err != nil {
			log.Fatalf("create relay client: %v", err)
		}
		defer client.Close()
		liveSource = source.NewLive(client)
	default:
		log.Fatalf("unsupported data source %q", cfg.DataSource)
	}

	app := handlers.New(catalogs, locales, liveSource, fixtureSources)
	mux := http.NewServeMux()

	mux.HandleFunc("/", app.Health)
	mux.HandleFunc("/capabilities", app.Capabilities)
	mux.HandleFunc("/ask", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.AskPage(w, r)
		case http.MethodPost:
			app.AskSubmit(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/ask/stream", app.AskStream)
	mux.HandleFunc("/egress", app.Egress)
	mux.HandleFunc("/audit", app.Audit)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	server := &http.Server{
		Addr:              cfg.AppBind,
		Handler:           middleware.WithPrincipal(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("hubrelay dashboard listening on %s", cfg.AppBind)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
