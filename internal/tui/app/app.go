package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/tui/client"
	"github.com/Akram012388/niotebook-tui/internal/tui/components"
	"github.com/Akram012388/niotebook-tui/internal/tui/config"
)

// View identifiers for the root model.
type View int

const (
	ViewLogin View = iota
	ViewRegister
	ViewTimeline
	ViewProfile
)

// ViewModel is the interface that all view sub-models must implement.
type ViewModel interface {
	Init() tea.Cmd
	Update(tea.Msg) (ViewModel, tea.Cmd)
	View() string
	HelpText() string
}

// ComposeViewModel is the interface for the compose overlay.
type ComposeViewModel interface {
	ViewModel
	Submitted() bool
	Cancelled() bool
	IsTextInputFocused() bool
}

// HelpViewModel is the interface for the help overlay.
type HelpViewModel interface {
	ViewModel
	Dismissed() bool
}

// ProfileViewModel is the interface for the profile view.
type ProfileViewModel interface {
	ViewModel
	Editing() bool
	Dismissed() bool
}

// TimelineViewModel is the interface for the timeline view.
type TimelineViewModel interface {
	ViewModel
	FetchLatest() tea.Cmd
}

// ViewFactory creates view sub-models. This breaks the import cycle between
// the app and views packages.
type ViewFactory interface {
	NewLogin(c *client.Client) ViewModel
	NewRegister(c *client.Client) ViewModel
	NewTimeline(c *client.Client) TimelineViewModel
	NewProfile(c *client.Client, userID string, isOwn bool) ProfileViewModel
	NewCompose(c *client.Client) ComposeViewModel
	NewHelp(viewName string) HelpViewModel
}

// Help view name constants.
const (
	HelpViewTimeline = "timeline"
	HelpViewProfile  = "profile"
	HelpViewCompose  = "compose"
)

// AppModel is the root Bubble Tea model that manages all sub-models,
// shared state, and view routing.
type AppModel struct {
	// Shared state
	client  *client.Client
	user    *models.User
	tokens  *models.TokenPair
	width   int
	height  int
	factory ViewFactory

	// Current view
	currentView View

	// Sub-models
	login    ViewModel
	register ViewModel
	timeline TimelineViewModel
	profile  ProfileViewModel

	// Overlays
	compose ComposeViewModel
	help    HelpViewModel

	// Components
	statusBar components.StatusBarModel
}

// NewAppModel creates the root app model. If storedAuth has a token, the model
// can skip login (but for now, always starts on login).
func NewAppModel(c *client.Client, storedAuth *config.StoredAuth) AppModel {
	m := AppModel{
		client:      c,
		currentView: ViewLogin,
		statusBar:   components.NewStatusBarModel(),
	}
	return m
}

// NewAppModelWithFactory creates the root app model with a view factory to
// break the import cycle between app and views.
func NewAppModelWithFactory(c *client.Client, storedAuth *config.StoredAuth, f ViewFactory) AppModel {
	m := AppModel{
		client:      c,
		currentView: ViewLogin,
		factory:     f,
		login:       f.NewLogin(c),
		register:    f.NewRegister(c),
		timeline:    f.NewTimeline(c),
		statusBar:   components.NewStatusBarModel(),
	}
	return m
}

// CurrentView returns the active view identifier.
func (m AppModel) CurrentView() View {
	return m.currentView
}

// IsComposeOpen returns whether the compose overlay is active.
func (m AppModel) IsComposeOpen() bool {
	return m.compose != nil
}

// isTextInputFocused returns true when a text input is focused, so global
// shortcuts like 'n', 'q', '?' should not fire.
func (m AppModel) isTextInputFocused() bool {
	if m.compose != nil {
		return true
	}
	switch m.currentView {
	case ViewLogin, ViewRegister:
		return true
	case ViewProfile:
		if m.profile != nil {
			return m.profile.Editing()
		}
	}
	return false
}

// Init satisfies tea.Model.
func (m AppModel) Init() tea.Cmd {
	if m.login != nil {
		return m.login.Init()
	}
	return nil
}

// Update satisfies tea.Model. Routing order:
// 1. Window size → propagate to all
// 2. Global keys (q to quit, unless text input focused)
// 3. Overlay routing (compose, help)
// 4. App-level messages (auth, post published, etc.)
// 5. View-specific routing
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m.propagateWindowSize(msg)

	case tea.KeyMsg:
		// Quit on ctrl+c always
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		// Quit on q when no text input is focused
		if !m.isTextInputFocused() && msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'q' {
			return m, tea.Quit
		}

		// Route to overlays first
		if m.help != nil {
			return m.updateHelp(msg)
		}
		if m.compose != nil {
			return m.updateCompose(msg)
		}

		// Global shortcuts (only when no text input focused)
		if !m.isTextInputFocused() {
			switch {
			case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'n':
				if m.currentView == ViewTimeline || m.currentView == ViewProfile {
					return m.openCompose()
				}
			case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == '?':
				return m.openHelp()
			case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'r':
				if m.currentView == ViewTimeline && m.timeline != nil {
					cmd := m.timeline.FetchLatest()
					return m, cmd
				}
			case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'p':
				if m.currentView == ViewTimeline && m.user != nil {
					return m.openProfile(m.user.ID, true)
				}
			}
		}

		// Fall through to view-specific key handling
		return m.updateCurrentView(msg)

	// App-level messages
	case MsgAuthSuccess:
		return m.handleAuthSuccess(msg)

	case MsgAuthError:
		return m.updateCurrentView(msg)

	case MsgAuthExpired:
		m.user = nil
		m.tokens = nil
		m.currentView = ViewLogin
		if m.factory != nil {
			m.login = m.factory.NewLogin(m.client)
		}
		cmd := m.statusBar.SetError("Session expired. Please log in again.")
		return m, cmd

	case MsgTimelineLoaded, MsgTimelineRefreshed:
		if m.timeline != nil {
			var updated ViewModel
			var cmd tea.Cmd
			updated, cmd = m.timeline.Update(msg)
			if tl, ok := updated.(TimelineViewModel); ok {
				m.timeline = tl
			}
			return m, cmd
		}
		return m, nil

	case MsgPostPublished:
		m.compose = nil
		cmd := m.statusBar.SetSuccess("Post published!")
		var fetchCmd tea.Cmd
		if m.timeline != nil {
			fetchCmd = m.timeline.FetchLatest()
		}
		return m, tea.Batch(cmd, fetchCmd)

	case MsgProfileLoaded:
		if m.profile != nil {
			var updated ViewModel
			var cmd tea.Cmd
			updated, cmd = m.profile.Update(msg)
			if pv, ok := updated.(ProfileViewModel); ok {
				m.profile = pv
			}
			return m, cmd
		}
		return m, nil

	case MsgProfileUpdated:
		if m.profile != nil {
			var updated ViewModel
			var cmd tea.Cmd
			updated, cmd = m.profile.Update(msg)
			if pv, ok := updated.(ProfileViewModel); ok {
				m.profile = pv
			}
			if msg.User != nil && m.user != nil && msg.User.ID == m.user.ID {
				m.user = msg.User
			}
			return m, cmd
		}
		return m, nil

	case MsgAPIError:
		cmd := m.statusBar.SetError(msg.Message)
		return m, cmd

	case components.MsgStatusClear:
		m.statusBar.Clear()
		return m, nil

	case MsgSwitchToRegister:
		m.currentView = ViewRegister
		if m.register != nil {
			return m, m.register.Init()
		}
		return m, nil

	case MsgSwitchToLogin:
		m.currentView = ViewLogin
		if m.login != nil {
			return m, m.login.Init()
		}
		return m, nil
	}

	return m.updateCurrentView(msg)
}

// View satisfies tea.Model. Layout: header + content + status bar.
func (m AppModel) View() string {
	// Before auth, render only the auth view (no header/status bar)
	if m.currentView == ViewLogin || m.currentView == ViewRegister {
		return m.viewCurrentContent()
	}

	// Header
	username := ""
	if m.user != nil {
		username = m.user.Username
	}
	viewName := m.viewName()
	header := components.RenderHeader("niotebook", username, viewName, m.width)

	// Content area: total height minus header (1 line) and status bar (1 line)
	contentHeight := m.height - 2
	if contentHeight < 1 {
		contentHeight = 1
	}

	content := m.viewCurrentContent()

	// If an overlay is open, render it on top of the content
	if m.help != nil {
		content = m.help.View()
	} else if m.compose != nil {
		content = m.compose.View()
	}

	content = lipgloss.NewStyle().Height(contentHeight).Render(content)

	// Status bar
	helpText := m.currentHelpText()
	statusBar := m.statusBar.View(helpText, m.width)

	return lipgloss.JoinVertical(lipgloss.Left, header, content, statusBar)
}

// handleAuthSuccess transitions from login/register to the timeline.
func (m AppModel) handleAuthSuccess(msg MsgAuthSuccess) (AppModel, tea.Cmd) {
	m.user = msg.User
	m.tokens = msg.Tokens
	m.currentView = ViewTimeline

	if m.client != nil && msg.Tokens != nil {
		m.client.SetToken(msg.Tokens.AccessToken)
		m.client.SetRefreshToken(msg.Tokens.RefreshToken)
	}

	// Fetch timeline
	if m.timeline != nil {
		cmd := m.timeline.FetchLatest()
		return m, cmd
	}
	return m, nil
}

// openCompose creates a new compose overlay.
func (m AppModel) openCompose() (AppModel, tea.Cmd) {
	if m.factory != nil {
		m.compose = m.factory.NewCompose(m.client)
		cmd := m.compose.Init()
		return m, cmd
	}
	return m, nil
}

// openHelp creates a new help overlay for the current context.
func (m AppModel) openHelp() (AppModel, tea.Cmd) {
	if m.factory == nil {
		return m, nil
	}
	var viewName string
	if m.compose != nil {
		viewName = HelpViewCompose
	} else {
		switch m.currentView {
		case ViewTimeline:
			viewName = HelpViewTimeline
		case ViewProfile:
			viewName = HelpViewProfile
		default:
			viewName = HelpViewTimeline
		}
	}
	m.help = m.factory.NewHelp(viewName)
	updated, _ := m.help.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	if hv, ok := updated.(HelpViewModel); ok {
		m.help = hv
	}
	return m, nil
}

// openProfile navigates to the profile view, preserving timeline state.
func (m AppModel) openProfile(userID string, isOwn bool) (AppModel, tea.Cmd) {
	if m.factory != nil {
		m.profile = m.factory.NewProfile(m.client, userID, isOwn)
		m.currentView = ViewProfile
		cmd := m.profile.Init()
		return m, cmd
	}
	return m, nil
}

// updateCompose routes messages to the compose overlay.
func (m AppModel) updateCompose(msg tea.Msg) (AppModel, tea.Cmd) {
	var updated ViewModel
	var cmd tea.Cmd
	updated, cmd = m.compose.Update(msg)
	if cv, ok := updated.(ComposeViewModel); ok {
		m.compose = cv
	}

	if m.compose.Cancelled() {
		m.compose = nil
		return m, nil
	}

	return m, cmd
}

// updateHelp routes messages to the help overlay.
func (m AppModel) updateHelp(msg tea.Msg) (AppModel, tea.Cmd) {
	var updated ViewModel
	var cmd tea.Cmd
	updated, cmd = m.help.Update(msg)
	if hv, ok := updated.(HelpViewModel); ok {
		m.help = hv
	}

	if m.help.Dismissed() {
		m.help = nil
		return m, nil
	}

	return m, cmd
}

// updateCurrentView routes messages to the active view.
func (m AppModel) updateCurrentView(msg tea.Msg) (AppModel, tea.Cmd) {
	var cmd tea.Cmd
	switch m.currentView {
	case ViewLogin:
		if m.login != nil {
			var updated ViewModel
			updated, cmd = m.login.Update(msg)
			m.login = updated
		}
	case ViewRegister:
		if m.register != nil {
			var updated ViewModel
			updated, cmd = m.register.Update(msg)
			m.register = updated
		}
	case ViewTimeline:
		if m.timeline != nil {
			var updated ViewModel
			updated, cmd = m.timeline.Update(msg)
			if tl, ok := updated.(TimelineViewModel); ok {
				m.timeline = tl
			}
		}
	case ViewProfile:
		if m.profile != nil {
			var updated ViewModel
			updated, cmd = m.profile.Update(msg)
			if pv, ok := updated.(ProfileViewModel); ok {
				m.profile = pv
			}
			if m.profile.Dismissed() {
				m.currentView = ViewTimeline
				return m, nil
			}
		}
	}
	return m, cmd
}

// propagateWindowSize sends the window size to all sub-models.
func (m AppModel) propagateWindowSize(msg tea.WindowSizeMsg) (AppModel, tea.Cmd) {
	var cmds []tea.Cmd

	if m.login != nil {
		var cmd tea.Cmd
		m.login, cmd = m.login.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.register != nil {
		var cmd tea.Cmd
		m.register, cmd = m.register.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.timeline != nil {
		var updated ViewModel
		var cmd tea.Cmd
		updated, cmd = m.timeline.Update(msg)
		if tl, ok := updated.(TimelineViewModel); ok {
			m.timeline = tl
		}
		cmds = append(cmds, cmd)
	}
	if m.profile != nil {
		var updated ViewModel
		var cmd tea.Cmd
		updated, cmd = m.profile.Update(msg)
		if pv, ok := updated.(ProfileViewModel); ok {
			m.profile = pv
		}
		cmds = append(cmds, cmd)
	}
	if m.compose != nil {
		var updated ViewModel
		var cmd tea.Cmd
		updated, cmd = m.compose.Update(msg)
		if cv, ok := updated.(ComposeViewModel); ok {
			m.compose = cv
		}
		cmds = append(cmds, cmd)
	}
	if m.help != nil {
		var updated ViewModel
		var cmd tea.Cmd
		updated, cmd = m.help.Update(msg)
		if hv, ok := updated.(HelpViewModel); ok {
			m.help = hv
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// viewCurrentContent returns the rendered content for the active view.
func (m AppModel) viewCurrentContent() string {
	switch m.currentView {
	case ViewLogin:
		if m.login != nil {
			return m.login.View()
		}
	case ViewRegister:
		if m.register != nil {
			return m.register.View()
		}
	case ViewTimeline:
		if m.timeline != nil {
			return m.timeline.View()
		}
	case ViewProfile:
		if m.profile != nil {
			return m.profile.View()
		}
	}
	return ""
}

// viewName returns a display name for the current view.
func (m AppModel) viewName() string {
	if m.compose != nil {
		return "Compose"
	}
	if m.help != nil {
		return "Help"
	}
	switch m.currentView {
	case ViewTimeline:
		return "Timeline"
	case ViewProfile:
		return "Profile"
	default:
		return ""
	}
}

// currentHelpText returns the status bar help text for the active view/overlay.
func (m AppModel) currentHelpText() string {
	if m.help != nil {
		return m.help.HelpText()
	}
	if m.compose != nil {
		return m.compose.HelpText()
	}
	switch m.currentView {
	case ViewLogin:
		if m.login != nil {
			return m.login.HelpText()
		}
	case ViewRegister:
		if m.register != nil {
			return m.register.HelpText()
		}
	case ViewTimeline:
		if m.timeline != nil {
			return m.timeline.HelpText()
		}
	case ViewProfile:
		if m.profile != nil {
			return m.profile.HelpText()
		}
	}
	return ""
}
