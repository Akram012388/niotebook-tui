package views

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

// MinSplashDuration is the minimum time the splash screen stays visible so the
// user can see the brand animation before transitioning.
const MinSplashDuration = 2500 * time.Millisecond

// Animation timing
const (
	revealTickInterval = 40 * time.Millisecond // per-character reveal speed
	taglinePause       = 200 * time.Millisecond
)

// splashPhase tracks the current animation phase.
type splashPhase int

const (
	phaseReveal     splashPhase = iota // typewriter logo reveal
	phaseTagline                       // tagline appearing
	phaseConnecting                    // spinner + status
)

// BlockSpinnerFrames returns the four styled frames for the custom block
// spinner. Each frame is a string of three block characters separated by
// spaces, progressively filling from light shade to full block.
func BlockSpinnerFrames() []string {
	border := lipgloss.NewStyle().Foreground(theme.Border)
	accent := lipgloss.NewStyle().Foreground(theme.Accent)

	light := border.Render("░")
	full := accent.Render("█")

	return []string{
		light + " " + light + " " + light, // ░ ░ ░
		full + " " + light + " " + light,  // █ ░ ░
		full + " " + full + " " + light,   // █ █ ░
		full + " " + full + " " + full,    // █ █ █
	}
}

// newBlockSpinner creates a spinner.Model configured with the custom block
// spinner frames at 300ms per frame.
func newBlockSpinner() spinner.Model {
	frames := BlockSpinnerFrames()
	s := spinner.New()
	s.Spinner = spinner.Spinner{
		Frames: frames,
		FPS:    300 * time.Millisecond,
	}
	return s
}

// SplashModel is the splash screen shown on app launch while connecting
// to the server.
type SplashModel struct {
	serverURL   string
	spinner     spinner.Model
	done        bool
	failed      bool
	err         string
	width       int
	height      int
	phase       splashPhase
	revealIndex int    // how many chars of the logo to show
	logoText    string // the full plain-text logo for counting
	showTagline bool
}

// NewSplashModel creates a new splash screen model.
func NewSplashModel(serverURL string) SplashModel {
	// The plain text of the splash logo for character counting: "n i o t e b o o k"
	logoPlain := "n i o t e b o o k"
	return SplashModel{
		serverURL: serverURL,
		spinner:   newBlockSpinner(),
		phase:     phaseReveal,
		logoText:  logoPlain,
	}
}

// Done returns whether the server connection succeeded.
func (m SplashModel) Done() bool { return m.done }

// Failed returns whether the server connection failed.
func (m SplashModel) Failed() bool { return m.failed }

// ErrorMessage returns the error message if the connection failed.
func (m SplashModel) ErrorMessage() string { return m.err }

// HelpText returns an empty string since the splash screen has no help.
func (m SplashModel) HelpText() string { return "" }

// Init returns the initial commands: start the typewriter reveal and check server health.
func (m SplashModel) Init() tea.Cmd {
	return tea.Batch(
		m.revealTick(),
		m.checkHealth(),
	)
}

func (m SplashModel) revealTick() tea.Cmd {
	return tea.Tick(revealTickInterval, func(_ time.Time) tea.Msg {
		return app.MsgRevealTick{}
	})
}

// Internal message for splash animation
type msgTaglineShow struct{}

func (m SplashModel) taglinePauseCmd() tea.Cmd {
	return tea.Tick(taglinePause, func(_ time.Time) tea.Msg {
		return msgTaglineShow{}
	})
}

// Update handles messages for the splash screen.
func (m SplashModel) Update(msg tea.Msg) (SplashModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case app.MsgServerConnected:
		m.done = true
		return m, nil

	case app.MsgServerFailed:
		m.failed = true
		m.err = msg.Err
		return m, nil

	case app.MsgRevealTick:
		if m.phase == phaseReveal {
			m.revealIndex++
			if m.revealIndex >= len([]rune(m.logoText)) {
				// Logo fully revealed, pause then show tagline
				m.phase = phaseTagline
				return m, m.taglinePauseCmd()
			}
			return m, m.revealTick()
		}
		return m, nil

	case msgTaglineShow:
		m.showTagline = true
		m.phase = phaseConnecting
		return m, m.spinner.Tick

	case tea.KeyMsg:
		if m.failed {
			switch {
			case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'r':
				// Retry connection
				m.failed = false
				m.err = ""
				return m, tea.Batch(m.spinner.Tick, m.checkHealth())
			}
		}
		switch {
		case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'q':
			return m, tea.Quit
		case msg.Type == tea.KeyCtrlC:
			return m, tea.Quit
		}
		return m, nil

	case spinner.TickMsg:
		if m.phase == phaseConnecting && !m.done && !m.failed {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	return m, nil
}

// View renders the splash screen.
func (m SplashModel) View() string {
	var b strings.Builder

	// Render logo with typewriter reveal
	b.WriteString(m.renderRevealLogo())
	b.WriteString("\n")

	// Tagline (only after reveal completes)
	if m.showTagline {
		b.WriteString(theme.TaglineSplash())
		b.WriteString("\n\n")
	} else {
		b.WriteString("\n\n")
	}

	// Connection status (only in connecting phase)
	if m.phase == phaseConnecting {
		if m.failed {
			errStyle := lipgloss.NewStyle().Foreground(theme.Error)
			b.WriteString(errStyle.Render(fmt.Sprintf("connection failed: %s", m.err)))
			b.WriteString("\n\n")
			b.WriteString(theme.Hint.Render("press r to retry · q to quit"))
		} else if !m.done {
			b.WriteString(m.spinner.View())
			b.WriteString("\n")
			b.WriteString(theme.Caption.Render("connecting..."))
		}
	}
	// If done, just show logo + tagline (about to transition)

	content := b.String()

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m SplashModel) renderRevealLogo() string {
	text := lipgloss.NewStyle().Bold(true).Foreground(theme.Text)
	accent := lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)

	letters := []struct {
		char  string
		style lipgloss.Style
	}{
		{"n", text}, {"i", accent}, {"o", text}, {"t", text},
		{"e", text}, {"b", text}, {"o", text}, {"o", text}, {"k", text},
	}

	// Build the spaced logo: "n i o t e b o o k"
	// Each letter takes 2 chars in the plain text (char + space), except the last
	var revealed strings.Builder
	charIndex := 0
	for i, l := range letters {
		if charIndex >= m.revealIndex {
			break
		}
		revealed.WriteString(l.style.Render(l.char))
		charIndex++
		// Add space between letters (except after last)
		if i < len(letters)-1 && charIndex < m.revealIndex {
			revealed.WriteString(" ")
			charIndex++
		}
	}

	return revealed.String()
}

// checkHealth returns a tea.Cmd that makes an HTTP GET request to the
// server's health endpoint. A minimum display time ensures the splash
// screen is visible even when the server responds instantly.
func (m SplashModel) checkHealth() tea.Cmd {
	url := m.serverURL
	return func() tea.Msg {
		start := time.Now()

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(url + "/health")

		// Ensure the splash is visible for at least MinSplashDuration.
		if elapsed := time.Since(start); elapsed < MinSplashDuration {
			time.Sleep(MinSplashDuration - elapsed)
		}

		if err != nil {
			return app.MsgServerFailed{Err: err.Error()}
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return app.MsgServerConnected{}
		}
		return app.MsgServerFailed{
			Err: fmt.Sprintf("server returned status %d", resp.StatusCode),
		}
	}
}
