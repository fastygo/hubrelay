package unixsock

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"sshbot/internal/adapters/shared"
	"sshbot/internal/core"
	proxymgr "sshbot/internal/proxy"
)

type Adapter struct {
	socketPath string
	service    *core.Service
	proxy      *proxymgr.Manager
	server     *http.Server
	listener   net.Listener
}

func New(socketPath string, service *core.Service, proxyManager *proxymgr.Manager) *Adapter {
	return &Adapter{
		socketPath: socketPath,
		service:    service,
		proxy:      proxyManager,
	}
}

func (a *Adapter) Name() string {
	return "unix_socket"
}

func (a *Adapter) Start(ctx context.Context) error {
	if err := os.MkdirAll(filepath.Dir(a.socketPath), 0o755); err != nil {
		return err
	}
	if err := os.Remove(a.socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	listener, err := net.Listen("unix", a.socketPath)
	if err != nil {
		return err
	}
	a.listener = listener

	if err := os.Chmod(a.socketPath, 0o660); err != nil {
		_ = listener.Close()
		return err
	}

	a.server = &http.Server{
		Handler:           shared.NewMux(a.service, a.proxy, a.Name()),
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
			log.Printf("unix socket shutdown failed: %v", err)
		}
		_ = a.listener.Close()
		_ = os.Remove(a.socketPath)
	}()

	err = a.server.Serve(listener)
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	_ = os.Remove(a.socketPath)
	return nil
}
