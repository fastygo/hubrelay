package config

import (
	"fmt"
	"os"
	"strings"

	coreconfig "github.com/fastygo/hubcore/config"
)

const (
	TransportHTTP     = "http"
	TransportGRPC     = "grpc"
	TransportUnix     = "unix"
	DataSourceLive    = "live"
	DataSourceFixture = "fixture"
)

type Config struct {
	AppBind            string
	DataDir            string
	WebhookToken       string
	QdrantURL          string
	Auth               AuthConfig
	DataSource         string
	Transport          string
	HubRelayBaseURL    string
	HubRelayGRPCTarget string
	HubRelaySocketPath string
}

type AuthConfig struct {
	AdminUser string
	AdminPass string
	Disabled  bool
}

func Load() (Config, error) {
	cfg := Config{
		AppBind:      coreconfig.DefaultString(os.Getenv("APP_BIND"), "127.0.0.1:8081"),
		DataDir:      coreconfig.DefaultString(os.Getenv("APP_DATA_DIR"), "./data"),
		WebhookToken: strings.TrimSpace(os.Getenv("APP_WEBHOOK_TOKEN")),
		QdrantURL:    strings.TrimSpace(os.Getenv("QDRANT_URL")),
		Auth: AuthConfig{
			AdminUser: coreconfig.DefaultStringFromEnv([]string{os.Getenv("APP_ADMIN_USER"), os.Getenv("PAAS_ADMIN_USER")}, "admin"),
			AdminPass: coreconfig.DefaultStringFromEnv([]string{os.Getenv("APP_ADMIN_PASS"), os.Getenv("PAAS_ADMIN_PASS")}, "admin@123"),
			Disabled:  coreconfig.DefaultBoolFromEnv([]string{os.Getenv("APP_AUTH_DISABLED")}, false),
		},
		DataSource:         strings.ToLower(coreconfig.DefaultString(os.Getenv("APP_DATA_SOURCE"), DataSourceLive)),
		Transport:          strings.ToLower(coreconfig.DefaultString(os.Getenv("HUBRELAY_TRANSPORT"), TransportHTTP)),
		HubRelayBaseURL:    coreconfig.DefaultString(os.Getenv("HUBRELAY_BASE_URL"), "http://127.0.0.1:5500"),
		HubRelayGRPCTarget: coreconfig.DefaultString(os.Getenv("HUBRELAY_GRPC_TARGET"), "127.0.0.1:5501"),
		HubRelaySocketPath: coreconfig.DefaultString(os.Getenv("HUBRELAY_SOCKET_PATH"), "/run/hubrelay/hubrelay.sock"),
	}

	switch cfg.DataSource {
	case DataSourceLive:
		switch cfg.Transport {
		case TransportHTTP:
			if strings.TrimSpace(cfg.HubRelayBaseURL) == "" {
				return Config{}, fmt.Errorf("HUBRELAY_BASE_URL must not be empty for http transport")
			}
		case TransportGRPC:
			if strings.TrimSpace(cfg.HubRelayGRPCTarget) == "" {
				return Config{}, fmt.Errorf("HUBRELAY_GRPC_TARGET must not be empty for grpc transport")
			}
		case TransportUnix:
			if strings.TrimSpace(cfg.HubRelaySocketPath) == "" {
				return Config{}, fmt.Errorf("HUBRELAY_SOCKET_PATH must not be empty for unix transport")
			}
		default:
			return Config{}, fmt.Errorf("unsupported HUBRELAY_TRANSPORT %q", cfg.Transport)
		}
	case DataSourceFixture:
	default:
		return Config{}, fmt.Errorf("unsupported APP_DATA_SOURCE %q", cfg.DataSource)
	}

	if strings.TrimSpace(cfg.AppBind) == "" {
		return Config{}, fmt.Errorf("APP_BIND must not be empty")
	}
	if strings.TrimSpace(cfg.DataDir) == "" {
		return Config{}, fmt.Errorf("APP_DATA_DIR must not be empty")
	}

	return cfg, nil
}
