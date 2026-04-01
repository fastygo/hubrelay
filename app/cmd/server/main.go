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
	"hubrelay-dashboard/internal/handlers"
	"hubrelay-dashboard/internal/middleware"
	"hubrelay-dashboard/internal/relay"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	client, err := relay.New(cfg)
	if err != nil {
		log.Fatalf("create relay client: %v", err)
	}
	defer client.Close()

	app := handlers.New(client)
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
