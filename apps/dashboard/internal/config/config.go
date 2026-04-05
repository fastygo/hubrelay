package config

import (
	"fmt"
	"os"
	"strings"

	coreconfig "github.com/fastygo/hubcore/config"
)

const (
	RoleHub           = "hub"
	RoleRemote        = "remote"
	TransportHTTP     = "http"
	TransportGRPC     = "grpc"
	TransportUnix     = "unix"
	DataSourceLive    = "live"
	DataSourceFixture = "fixture"
)

type Config struct {
	Role               string
	AppBind            string
	Auth               AuthConfig
	DataSource         string
	Transport          string
	HubRelayBaseURL    string
	HubRelaySocketPath string
}

type AuthConfig struct {
	AdminUser string
	AdminPass string
	Disabled  bool
}

func Load() (Config, error) {
	cfg := Config{
		Role:    strings.ToLower(coreconfig.DefaultString(os.Getenv("DASHBOARD_ROLE"), RoleHub)),
		AppBind: coreconfig.DefaultString(os.Getenv("APP_BIND"), "0.0.0.0:8080"),
		Auth: AuthConfig{
			AdminUser: coreconfig.DefaultStringFromEnv([]string{os.Getenv("APP_ADMIN_USER"), os.Getenv("PAAS_ADMIN_USER")}, "admin"),
			AdminPass: coreconfig.DefaultStringFromEnv([]string{os.Getenv("APP_ADMIN_PASS"), os.Getenv("PAAS_ADMIN_PASS")}, "admin@123"),
			Disabled:  coreconfig.DefaultBoolFromEnv([]string{os.Getenv("APP_AUTH_DISABLED"), os.Getenv("DASHBOARD_AUTH_DISABLED")}, false),
		},
		DataSource:         strings.ToLower(coreconfig.DefaultString(os.Getenv("APP_DATA_SOURCE"), DataSourceLive)),
		Transport:          strings.ToLower(coreconfig.DefaultString(os.Getenv("HUBRELAY_TRANSPORT"), TransportHTTP)),
		HubRelayBaseURL:    coreconfig.DefaultString(os.Getenv("HUBRELAY_BASE_URL"), "http://127.0.0.1:5500"),
		HubRelaySocketPath: coreconfig.DefaultString(os.Getenv("HUBRELAY_SOCKET_PATH"), "/run/hubrelay/hubrelay.sock"),
	}

	switch cfg.Role {
	case RoleHub, RoleRemote:
	default:
		return Config{}, fmt.Errorf("unsupported DASHBOARD_ROLE %q", cfg.Role)
	}

	switch cfg.DataSource {
	case DataSourceLive:
		switch cfg.Transport {
		case TransportHTTP:
			if strings.TrimSpace(cfg.HubRelayBaseURL) == "" {
				return Config{}, fmt.Errorf("HUBRELAY_BASE_URL must not be empty for http transport")
			}
		case TransportGRPC:
			// Reserved for the future gRPC transport.
			// Keep config validation open so the transport slot is stable across refactors.
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

	return cfg, nil
}
