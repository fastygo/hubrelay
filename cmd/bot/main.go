package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"sshbot/internal/adapters/email"
	"sshbot/internal/adapters/httpchat"
	"sshbot/internal/adapters/unixsock"
	"sshbot/internal/ai"
	"sshbot/internal/buildprofile"
	"sshbot/internal/core"
	"sshbot/internal/egress"
	"sshbot/internal/outbound"
	askplugin "sshbot/internal/plugins/ask"
	egressplugin "sshbot/internal/plugins/egress"
	"sshbot/internal/plugins/systeminfo"
	proxymgr "sshbot/internal/proxy"
	"sshbot/internal/storage"
)

func main() {
	profile := buildprofile.Current()
	log.Printf("profile=%s ai_enabled=%v ai_has_key=%v ai_model=%s ai_base_url=%s ai_mode=%s proxy_session=%v proxy_force=%v",
		profile.ID,
		profile.OpenAI.Enabled,
		profile.OpenAI.HasAPIKey,
		profile.OpenAI.Model,
		profile.OpenAI.BaseURL,
		profile.OpenAI.APIMode,
		profile.ProxySession.Enabled,
		profile.ProxySession.Force,
	)
	if profile.ProxySession.Force && !profile.ProxySession.Enabled {
		log.Fatalf("proxy session force requires proxy session support to be enabled")
	}
	dbPath := os.Getenv("BOT_DB_FILE")
	if dbPath == "" {
		dbPath = "data/bot.db"
	}

	store, err := storage.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open runtime store: %v", err)
	}
	defer func() {
		if closeErr := store.Close(); closeErr != nil {
			log.Printf("failed to close store: %v", closeErr)
		}
	}()

	var proxyManager *proxymgr.Manager
	if profile.ProxySession.Enabled && profile.OpenAI.Enabled && profile.OpenAI.HasAPIKey {
		proxyManager = proxymgr.NewManager(ai.NewOpenAIProxyProber(profile.OpenAI.APIKey, profile.OpenAI.BaseURL, 12*time.Second))
	}
	privateEgressChecker := outbound.PrivateEgressChecker(outbound.NewInterfaceEgressChecker())
	var egressManager *egress.Manager
	if profile.Egress.UseManager && len(profile.Egress.Gateways) > 0 {
		egressManager = egress.NewManager(egress.Config{
			Gateways:      profile.Egress.Gateways,
			CheckInterval: profile.Egress.CheckInterval,
		}, nil)
		privateEgressChecker = outbound.NewEgressManagerChecker(egressManager)
	}

	var provider ai.Provider
	if profile.OpenAI.Enabled && profile.OpenAI.HasAPIKey {
		providerInstance, providerErr := ai.NewOpenAICompatibleProvider(
			profile.OpenAI.Provider,
			profile.OpenAI.APIKey,
			profile.OpenAI.BaseURL,
			profile.OpenAI.Model,
			profile.OpenAI.APIMode,
			outbound.Policy{RequireProxy: profile.ProxySession.Force},
			profile.PrivateEgress,
			privateEgressChecker,
			proxyManager,
		)
		if providerErr != nil {
			log.Fatalf("failed to configure ai provider: %v", providerErr)
		}
		provider = providerInstance
	}

	registry := core.NewPluginRegistry()
	registry.Register("system", systeminfo.Factory)
	registry.Register("ask", askplugin.Factory)
	registry.Register("egress", egressplugin.Factory)

	plugins, err := registry.BuildAll(core.PluginFactoryContext{
		ProfileID:    profile.ID,
		Capabilities: append([]core.Capability(nil), profile.Capabilities...),
		Config:       cloneStringMap(profile.Config),
		Store:        store,
		Deps: map[string]any{
			"egress_manager": egressManager,
			"proxy_manager":  proxyManager,
			"ai_provider":    provider,
		},
	})
	if err != nil {
		log.Fatalf("build plugin registry: %v", err)
	}

	serviceOptions := make([]core.ServiceOption, 0, 1)
	if egressManager != nil {
		serviceOptions = append(serviceOptions, core.WithActiveGatewayProvider(egressManager))
	}

	service, err := core.NewService(profile, store, plugins, serviceOptions...)
	if err != nil {
		log.Fatalf("failed to build service: %v", err)
	}

	adapters := make([]core.Adapter, 0, 2)
	if profile.HTTPChat.Enabled {
		adapters = append(adapters, httpchat.New(profile.HTTPChat.BindAddress, service, proxyManager))
	}
	if profile.UnixSocket.Enabled && profile.UnixSocket.SocketPath != "" {
		adapters = append(adapters, unixsock.New(profile.UnixSocket.SocketPath, service, proxyManager))
	}
	if profile.Email.Enabled {
		adapters = append(adapters, email.New(profile.Email.Provider, profile.Email.Mode))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if proxyManager != nil {
		go proxyManager.Start(ctx)
	}
	if egressManager != nil {
		go egressManager.Start(ctx)
	}

	errCh := make(chan error, len(adapters))
	var wg sync.WaitGroup
	for _, adapter := range adapters {
		current := adapter
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("starting adapter: %s", current.Name())
			if runErr := current.Start(ctx); runErr != nil {
				errCh <- runErr
			}
		}()
	}

	select {
	case <-ctx.Done():
		log.Printf("shutdown requested")
	case runErr := <-errCh:
		log.Printf("adapter failed: %v", runErr)
		cancel()
	}

	wg.Wait()
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}
