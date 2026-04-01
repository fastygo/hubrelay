package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	TransportHTTP     = "http"
	TransportUnix     = "unix"
	DataSourceLive    = "live"
	DataSourceFixture = "fixture"
)

type Config struct {
	AppBind            string
	DataSource         string
	Transport          string
	HubRelayBaseURL    string
	HubRelaySocketPath string
}

func Load() (Config, error) {
	cfg := Config{
		AppBind:            defaultString(os.Getenv("APP_BIND"), "0.0.0.0:8080"),
		DataSource:         strings.ToLower(defaultString(os.Getenv("APP_DATA_SOURCE"), DataSourceLive)),
		Transport:          strings.ToLower(defaultString(os.Getenv("HUBRELAY_TRANSPORT"), TransportHTTP)),
		HubRelayBaseURL:    defaultString(os.Getenv("HUBRELAY_BASE_URL"), "http://127.0.0.1:5500"),
		HubRelaySocketPath: defaultString(os.Getenv("HUBRELAY_SOCKET_PATH"), "/run/hubrelay/hubrelay.sock"),
	}

	switch cfg.DataSource {
	case DataSourceLive:
		switch cfg.Transport {
		case TransportHTTP:
			if strings.TrimSpace(cfg.HubRelayBaseURL) == "" {
				return Config{}, fmt.Errorf("HUBRELAY_BASE_URL must not be empty for http transport")
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

	return cfg, nil
}

func defaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
