package views

import (
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
)

// mockUser creates a models.User with the given username and display name.
func mockUser(username, displayName string) *models.User {
	return &models.User{
		ID:          "mock-" + username,
		Username:    username,
		DisplayName: displayName,
		CreatedAt:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
	}
}

// GenerateMockPosts returns 55 diverse mock posts for dev/testing.
func GenerateMockPosts() []models.Post {
	now := time.Now()

	authors := map[string]*models.User{
		"akram":        mockUser("akram", "Akram"),
		"alice":        mockUser("alice", "Alice Chen"),
		"bob":          mockUser("bob", "Bob Builder"),
		"dev_sarah":    mockUser("dev_sarah", "Sarah Dev"),
		"terminal_fan": mockUser("terminal_fan", "Terminal Fan"),
		"gopher_grace": mockUser("gopher_grace", "Grace Gopher"),
		"rust_rover":   mockUser("rust_rover", "Rover Rust"),
		"vim_master":   mockUser("vim_master", "Vim Master"),
		"cloud_nina":   mockUser("cloud_nina", "Nina Cloud"),
		"code_poet":    mockUser("code_poet", "Code Poet"),
	}

	posts := []models.Post{
		{ID: "p01", Author: authors["akram"], Content: "just launched niotebook TUI! the social terminal is alive", CreatedAt: now.Add(-2 * time.Minute)},
		{ID: "p02", Author: authors["alice"], Content: "loving the terminal vibes here\nthis is what social media should feel like\nno ads, no algorithm, just text", CreatedAt: now.Add(-5 * time.Minute)},
		{ID: "p03", Author: authors["bob"], Content: "building in public is easier when your tools are simple #terminal-life", CreatedAt: now.Add(-8 * time.Minute)},
		{ID: "p04", Author: authors["dev_sarah"], Content: "Go + Bubble Tea = perfect combo for TUIs", CreatedAt: now.Add(-12 * time.Minute)},
		{ID: "p05", Author: authors["terminal_fan"], Content: "who needs a browser when you have a terminal?", CreatedAt: now.Add(-15 * time.Minute)},
		{ID: "p06", Author: authors["gopher_grace"], Content: "the compose bar is so clean\ntype your thoughts\nhit ctrl+enter\ndone", CreatedAt: now.Add(-18 * time.Minute)},
		{ID: "p07", Author: authors["vim_master"], Content: "j/k navigation feels like home #vim #terminal-love", CreatedAt: now.Add(-22 * time.Minute)},
		{ID: "p08", Author: authors["rust_rover"], Content: "respect to the Go devs out there. Bubble Tea is a gem.", CreatedAt: now.Add(-25 * time.Minute)},
		{ID: "p09", Author: authors["cloud_nina"], Content: "deploying niotebook on my homelab this weekend", CreatedAt: now.Add(-30 * time.Minute)},
		{ID: "p10", Author: authors["code_poet"], Content: "code is poetry\nterminals are canvases\nwe paint with text", CreatedAt: now.Add(-35 * time.Minute)},
		{ID: "p11", Author: authors["akram"], Content: "working on the three-column layout. left sidebar, center feed, right discover. X-style but for the terminal.", CreatedAt: now.Add(-40 * time.Minute)},
		{ID: "p12", Author: authors["alice"], Content: "just discovered #claude-code and my productivity doubled", CreatedAt: now.Add(-45 * time.Minute)},
		{ID: "p13", Author: authors["bob"], Content: "hot take: the best UI is no UI. just give me a prompt.", CreatedAt: now.Add(-50 * time.Minute)},
		{ID: "p14", Author: authors["dev_sarah"], Content: "TDD is not optional\nwrite the test first\nwatch it fail\nmake it pass\nrefactor\nrepeat", CreatedAt: now.Add(-55 * time.Minute)},
		{ID: "p15", Author: authors["terminal_fan"], Content: "my terminal color scheme: gruvbox dark. fight me.", CreatedAt: now.Add(-1 * time.Hour)},
		{ID: "p16", Author: authors["gopher_grace"], Content: "#opencode is the future of development tools", CreatedAt: now.Add(-1*time.Hour - 5*time.Minute)},
		{ID: "p17", Author: authors["vim_master"], Content: "protip: use lipgloss adaptive colors so your TUI works in both dark and light terminals", CreatedAt: now.Add(-1*time.Hour - 10*time.Minute)},
		{ID: "p18", Author: authors["rust_rover"], Content: "async in Go: goroutines + channels = simplicity\nasync in Rust: Pin<Box<dyn Future>> = pain", CreatedAt: now.Add(-1*time.Hour - 15*time.Minute)},
		{ID: "p19", Author: authors["cloud_nina"], Content: "docker compose up -d niotebook\nthat's the whole deploy", CreatedAt: now.Add(-1*time.Hour - 20*time.Minute)},
		{ID: "p20", Author: authors["code_poet"], Content: "a", CreatedAt: now.Add(-1*time.Hour - 25*time.Minute)},
		{ID: "p21", Author: authors["akram"], Content: "keyboard shortcuts make everything better. Tab to switch columns, j/k to navigate, n to compose. #niotebook", CreatedAt: now.Add(-1*time.Hour - 30*time.Minute)},
		{ID: "p22", Author: authors["alice"], Content: "#mcp-servers are the new APIs. context-aware tool integration changes everything.", CreatedAt: now.Add(-2 * time.Hour)},
		{ID: "p23", Author: authors["bob"], Content: "spent 3 hours debugging a CSS layout. switched to terminal UI. fixed in 20 minutes.", CreatedAt: now.Add(-2*time.Hour - 10*time.Minute)},
		{ID: "p24", Author: authors["dev_sarah"], Content: "cursor-based pagination > offset pagination\nchange my mind", CreatedAt: now.Add(-2*time.Hour - 20*time.Minute)},
		{ID: "p25", Author: authors["terminal_fan"], Content: "the best feature of niotebook? no notifications. you check it when YOU want to.", CreatedAt: now.Add(-2*time.Hour - 30*time.Minute)},
		{ID: "p26", Author: authors["gopher_grace"], Content: "reading through the Bubble Tea source code and it's so clean. Elm architecture in Go just works.", CreatedAt: now.Add(-3 * time.Hour)},
		{ID: "p27", Author: authors["vim_master"], Content: ":wq", CreatedAt: now.Add(-3*time.Hour - 10*time.Minute)},
		{ID: "p28", Author: authors["rust_rover"], Content: "#agentic-coding is the next paradigm. not AI writing code for you. AI working WITH you.", CreatedAt: now.Add(-3*time.Hour - 20*time.Minute)},
		{ID: "p29", Author: authors["cloud_nina"], Content: "postgresql > everything else\nfight me\n(you won't win)", CreatedAt: now.Add(-3*time.Hour - 30*time.Minute)},
		{ID: "p30", Author: authors["code_poet"], Content: "exactly one hundred and forty characters is the perfect length for a thought on the social terminal platform niotebook check it out now!!!!", CreatedAt: now.Add(-4 * time.Hour)},
		{ID: "p31", Author: authors["akram"], Content: "the splash screen animation is coming along nicely. typewriter effect for the logo reveal.", CreatedAt: now.Add(-4*time.Hour - 15*time.Minute)},
		{ID: "p32", Author: authors["alice"], Content: "why does every app need to be in the browser?\nsome things are better in the terminal\nlike social media\napparently", CreatedAt: now.Add(-4*time.Hour - 30*time.Minute)},
		{ID: "p33", Author: authors["bob"], Content: "#codex #skills #openclaw — the open source AI tooling ecosystem is exploding", CreatedAt: now.Add(-5 * time.Hour)},
		{ID: "p34", Author: authors["dev_sarah"], Content: "JWT refresh tokens with single-use rotation. the right way to do auth.", CreatedAt: now.Add(-5*time.Hour - 20*time.Minute)},
		{ID: "p35", Author: authors["terminal_fan"], Content: "tmux + neovim + niotebook = the holy trinity", CreatedAt: now.Add(-5*time.Hour - 40*time.Minute)},
		{ID: "p36", Author: authors["gopher_grace"], Content: "interface segregation principle in Go: define interfaces where they're consumed, not where they're implemented", CreatedAt: now.Add(-6 * time.Hour)},
		{ID: "p37", Author: authors["vim_master"], Content: "hjkl is muscle memory at this point", CreatedAt: now.Add(-6*time.Hour - 30*time.Minute)},
		{ID: "p38", Author: authors["rust_rover"], Content: "monorepo life: one repo, two binaries, zero drama", CreatedAt: now.Add(-7 * time.Hour)},
		{ID: "p39", Author: authors["cloud_nina"], Content: "make build && make test && make deploy\nthat's the whole CI pipeline", CreatedAt: now.Add(-8 * time.Hour)},
		{ID: "p40", Author: authors["code_poet"], Content: "bits and bytes\nflow through wires\ntext on screens\na world entire", CreatedAt: now.Add(-9 * time.Hour)},
		{ID: "p41", Author: authors["akram"], Content: "three column layout done. responsive breakpoints at 100 and 80 cols.", CreatedAt: now.Add(-10 * time.Hour)},
		{ID: "p42", Author: authors["alice"], Content: "the amber accent color is perfect. warm. inviting. #niotebook", CreatedAt: now.Add(-12 * time.Hour)},
		{ID: "p43", Author: authors["bob"], Content: "convention: lowercase error messages, no trailing punctuation. it's the Go way.", CreatedAt: now.Add(-14 * time.Hour)},
		{ID: "p44", Author: authors["dev_sarah"], Content: "slog > logrus > zap\nstdlib wins again", CreatedAt: now.Add(-16 * time.Hour)},
		{ID: "p45", Author: authors["terminal_fan"], Content: "alacritty with JetBrains Mono. that's the setup.", CreatedAt: now.Add(-18 * time.Hour)},
		{ID: "p46", Author: authors["gopher_grace"], Content: "table-driven tests are the backbone of good Go testing", CreatedAt: now.Add(-20 * time.Hour)},
		{ID: "p47", Author: authors["vim_master"], Content: "mapped my caps lock to escape years ago. best decision ever.", CreatedAt: now.Add(-1 * 24 * time.Hour)},
		{ID: "p48", Author: authors["rust_rover"], Content: "pgx v5 > database/sql\nconnection pooling, prepared statements, batch queries, all built in", CreatedAt: now.Add(-1*24*time.Hour - 6*time.Hour)},
		{ID: "p49", Author: authors["cloud_nina"], Content: "the social terminal. I didn't know I needed this until now.", CreatedAt: now.Add(-1*24*time.Hour - 12*time.Hour)},
		{ID: "p50", Author: authors["code_poet"], Content: "first!", CreatedAt: now.Add(-2 * 24 * time.Hour)},
		{ID: "p51", Author: authors["akram"], Content: "welcome to niotebook. the social terminal.\nwrite your thoughts. share with the world.\n140 characters at a time.", CreatedAt: now.Add(-2*24*time.Hour - 6*time.Hour)},
		{ID: "p52", Author: authors["alice"], Content: "thisisaverylongwordwithnospacestotesthowthetextWrappingHandlesItInThePostCardRenderingComponentOfTheNiotebookTUI", CreatedAt: now.Add(-2*24*time.Hour - 12*time.Hour)},
		{ID: "p53", Author: authors["bob"], Content: "\n\n\n", CreatedAt: now.Add(-3 * 24 * time.Hour)},
		{ID: "p54", Author: authors["dev_sarah"], Content: "line1\nline2\nline3\nline4\nline5", CreatedAt: now.Add(-3*24*time.Hour - 12*time.Hour)},
		{ID: "p55", Author: authors["terminal_fan"], Content: "short", CreatedAt: now.Add(-4 * 24 * time.Hour)},
	}

	// Set AuthorID for all posts
	for i := range posts {
		if posts[i].Author != nil {
			posts[i].AuthorID = posts[i].Author.ID
		}
	}

	return posts
}
