package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Akram012388/niotebook-tui/internal/build"
	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/client"
	"github.com/Akram012388/niotebook-tui/internal/tui/config"
	"github.com/Akram012388/niotebook-tui/internal/tui/views"
)

func main() {
	serverURL := flag.String("server", "", "server URL (overrides config)")
	configPath := flag.String("config", "", "config file path")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("niotebook-tui %s (%s)\n", build.Version, build.CommitSHA)
		os.Exit(0)
	}

	// Load config
	cfgDir := config.ConfigDir()
	cfgFile := filepath.Join(cfgDir, "config.yaml")
	if *configPath != "" {
		cfgFile = *configPath
	}

	cfg, err := config.LoadFromFile(cfgFile)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	if *serverURL != "" {
		cfg.ServerURL = *serverURL
	}

	// Load stored auth
	authFile := filepath.Join(cfgDir, "auth.json")
	storedAuth, _ := config.LoadAuth(authFile)

	// Create HTTP client
	c := client.New(cfg.ServerURL)
	if storedAuth != nil {
		c.SetToken(storedAuth.AccessToken)
		c.SetRefreshToken(storedAuth.RefreshToken)
	}

	// Set up token persistence callback
	c.OnTokenRefresh(func(accessToken, refreshToken string) {
		_ = config.SaveAuth(authFile, &config.StoredAuth{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		})
	})

	// Create and run app
	factory := views.NewFactory()
	model := app.NewAppModelWithFactory(c, storedAuth, factory)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		slog.Error("TUI error", "err", err)
		os.Exit(1)
	}
}
