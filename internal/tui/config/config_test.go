package config_test

import (
	"os"
	"path/filepath"
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
	os.WriteFile(path, []byte("server_url: http://localhost:8080\n"), 0600)

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
