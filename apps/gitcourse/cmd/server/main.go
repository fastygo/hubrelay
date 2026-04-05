package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"gitcourse/internal/config"
	"gitcourse/internal/content"
	"gitcourse/internal/git"
	"gitcourse/internal/handlers"
	"gitcourse/internal/middleware"
	appmodule "gitcourse/internal/module"
	askmodule "gitcourse/internal/modules/ask"
	coursemodule "gitcourse/internal/modules/course"
	"gitcourse/internal/progress"
	"gitcourse/internal/relay"
	"gitcourse/internal/source"
	"gitcourse/internal/store"
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

	switch cfg.DataSource {
	case config.DataSourceFixture:
		for _, locale := range locales {
			fixtureSources[locale], err = source.NewFixture(locale)
			if err != nil {
				log.Fatalf("create fixture source for %s: %v", locale, err)
			}
		}
	case config.DataSourceLive:
		courseStore, err := store.NewJSONFileStore(cfg.DataDir)
		if err != nil {
			log.Fatalf("init store: %v", err)
		}
		reader := git.NewHTTPReader()
		progressCache := progress.NewCache(reader, 5*time.Minute)
		client, err := relay.New(cfg)
		if err != nil {
			log.Fatalf("create relay client: %v", err)
		}
		defer client.Close()
		liveSource = source.NewLive(cfg, courseStore, reader, progressCache, client)
	default:
		log.Fatalf("unsupported data source %q", cfg.DataSource)
	}

	auth := middleware.NewSessionAuth(cfg.Auth.AdminUser, cfg.Auth.AdminPass)
	app := handlers.New(catalogs, locales, liveSource, fixtureSources, auth, cfg.Auth.Disabled, cfg.WebhookToken)
	mux := http.NewServeMux()

	modules := []appmodule.Module{
		coursemodule.New(app),
		askmodule.New(app),
	}
	app.SetNavItems(collectNavItems(modules))

	mux.HandleFunc("/login", app.Login)
	mux.HandleFunc("/logout", app.Logout)
	for _, module := range modules {
		module.Routes(mux)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	next := middleware.WithPrincipal(mux)
	if !cfg.Auth.Disabled {
		next = auth.Middleware()(next)
	}

	server := &http.Server{
		Addr:              cfg.AppBind,
		Handler:           next,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("gitcourse listening on %s", cfg.AppBind)
		if cfg.Auth.Disabled {
			log.Printf("gitcourse auth disabled")
		} else {
			log.Printf("gitcourse login at http://%s/login", cfg.AppBind)
		}
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

func collectNavItems(modules []appmodule.Module) []appmodule.NavItem {
	var items []appmodule.NavItem
	for _, module := range modules {
		items = append(items, module.NavItems()...)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Order == items[j].Order {
			return items[i].Path < items[j].Path
		}
		return items[i].Order < items[j].Order
	})
	return items
}
