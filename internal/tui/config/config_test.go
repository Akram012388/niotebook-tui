package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/tui/config"
)

func TestLoadConfigDefaults(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.ServerURL != "https://api.niotebook.com" {
		t.Errorf("ServerURL = %q, want default", cfg.ServerURL)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("server_url: http://localhost:8080\n"), 0600); err != nil {
		t.Fatalf("setup WriteFile: %v", err)
	}

	cfg, err := config.LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile: %v", err)
	}
	if cfg.ServerURL != "http://localhost:8080" {
		t.Errorf("ServerURL = %q, want %q", cfg.ServerURL, "http://localhost:8080")
	}
}

func TestSaveAndLoadAuthTokens(t *testing.T) {
	dir := t.TempDir()
	authPath := filepath.Join(dir, "auth.json")

	tokens := &config.StoredAuth{
		AccessToken:  "access-123",
		RefreshToken: "refresh-456",
		ExpiresAt:    "2026-02-16T22:00:00Z",
	}
	if err := config.SaveAuth(authPath, tokens); err != nil {
		t.Fatalf("SaveAuth: %v", err)
	}

	loaded, err := config.LoadAuth(authPath)
	if err != nil {
		t.Fatalf("LoadAuth: %v", err)
	}
	if loaded.AccessToken != "access-123" {
		t.Errorf("AccessToken = %q, want %q", loaded.AccessToken, "access-123")
	}
}

func TestLoadAuthFileNotFound(t *testing.T) {
	_, err := config.LoadAuth("/nonexistent/auth.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestConfigDirXDGOverride(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/test-xdg")
	dir := config.ConfigDir()
	want := "/tmp/test-xdg/niotebook"
	if dir != want {
		t.Errorf("ConfigDir() = %q, want %q", dir, want)
	}
}

func TestConfigDirDefaultWithoutXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	dir := config.ConfigDir()
	if !strings.HasSuffix(dir, "/.config/niotebook") {
		t.Errorf("ConfigDir() = %q, want suffix /.config/niotebook", dir)
	}
}
