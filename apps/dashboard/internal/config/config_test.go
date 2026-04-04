package config

import "testing"

func TestLoadDefaultsToHubRole(t *testing.T) {
	t.Setenv("DASHBOARD_ROLE", "")
	t.Setenv("APP_BIND", "127.0.0.1:8080")
	t.Setenv("APP_DATA_SOURCE", DataSourceFixture)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Role != RoleHub {
		t.Fatalf("expected default role %q, got %q", RoleHub, cfg.Role)
	}
}

func TestLoadAcceptsRemoteRole(t *testing.T) {
	t.Setenv("DASHBOARD_ROLE", RoleRemote)
	t.Setenv("APP_BIND", "127.0.0.1:8080")
	t.Setenv("APP_DATA_SOURCE", DataSourceFixture)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Role != RoleRemote {
		t.Fatalf("expected role %q, got %q", RoleRemote, cfg.Role)
	}
}

func TestLoadRejectsUnsupportedRole(t *testing.T) {
	t.Setenv("DASHBOARD_ROLE", "invalid")
	t.Setenv("APP_BIND", "127.0.0.1:8080")
	t.Setenv("APP_DATA_SOURCE", DataSourceFixture)

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load() to reject unsupported DASHBOARD_ROLE")
	}
}
