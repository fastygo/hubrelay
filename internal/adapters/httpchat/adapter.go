package httpchat

import (
	"context"
	"log"
	"net/http"
	"time"

	"sshbot/internal/adapters/shared"
	"sshbot/internal/core"
	proxymgr "sshbot/internal/proxy"
)

type Adapter struct {
	bindAddress string
	service     *core.Service
	proxy       *proxymgr.Manager
	server      *http.Server
}

func New(bindAddress string, service *core.Service, proxyManager *proxymgr.Manager) *Adapter {
	return &Adapter{
		bindAddress: bindAddress,
		service:     service,
		proxy:       proxyManager,
	}
}

func (a *Adapter) Name() string {
	return "http_chat"
}

func (a *Adapter) Start(ctx context.Context) error {
	mux := a.buildMux()

	a.server = &http.Server{
		Addr:              a.bindAddress,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			log.Printf("http chat shutdown failed: %v", err)
		}
	}()

	err := a.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (a *Adapter) buildMux() *http.ServeMux {
	return shared.NewMux(a.service, a.proxy, a.Name())
}
