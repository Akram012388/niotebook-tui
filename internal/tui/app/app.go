package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/tui/client"
	"github.com/Akram012388/niotebook-tui/internal/tui/components"
	"github.com/Akram012388/niotebook-tui/internal/tui/config"
	"github.com/Akram012388/niotebook-tui/internal/tui/layout"
)

// View identifiers for the root model.
type View int

const (
	ViewSplash View = iota
	ViewLogin
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

// ComposeViewModel is the interface for the inline compose bar.
type ComposeViewModel interface {
	ViewModel
	Submitted() bool
	Cancelled() bool
	Expanded() bool
	IsTextInputFocused() bool
	Expand()
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

// SplashViewModel is the interface for the splash screen.
type SplashViewModel interface {
	ViewModel
	Done() bool
	Failed() bool
	ErrorMessage() string
}

// TimelineViewModel is the interface for the timeline view.
type TimelineViewModel interface {
	ViewModel
	FetchLatest() tea.Cmd
}

// ViewFactory creates view sub-models. This breaks the import cycle between
// the app and views packages.
type ViewFactory interface {
	NewSplash(serverURL string) SplashViewModel
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
	client    *client.Client
	serverURL string
	user      *models.User
	tokens    *models.TokenPair
	width     int
	height    int
	factory   ViewFactory

	// Stored auth for post-splash transition
	storedAuth *config.StoredAuth

	// Current view
	currentView View

	// Column focus
	focus layout.FocusState

	// Sub-models
	splash   SplashViewModel
	login    ViewModel
	register ViewModel
	timeline TimelineViewModel
	profile  ProfileViewModel

	// Inline compose bar (always present on authenticated views)
	compose ComposeViewModel

	// Overlays
	help HelpViewModel

	// Interactive state for sidebar and discover panels
	sidebarState  components.SidebarState
	discoverState components.DiscoverState
}

// NewAppModel creates the root app model. If storedAuth has a token, the model
// can skip login (but for now, always starts on login).
func NewAppModel(c *client.Client, storedAuth *config.StoredAuth) AppModel {
	m := AppModel{
		client:      c,
		currentView: ViewLogin,
		focus:       layout.NewFocusState(),
	}
	return m
}

// NewAppModelWithFactory creates the root app model with a view factory to
// break the import cycle between app and views.
func NewAppModelWithFactory(c *client.Client, storedAuth *config.StoredAuth, f ViewFactory, serverURL string) AppModel {
	m := AppModel{
		client:      c,
		serverURL:   serverURL,
		currentView: ViewSplash,
		focus:       layout.NewFocusState(),
		factory:     f,
		storedAuth:  storedAuth,
		splash:      f.NewSplash(serverURL),
	}
	return m
}

// CurrentView returns the active view identifier.
func (m AppModel) CurrentView() View {
	return m.currentView
}

// IsComposeOpen returns whether the compose bar is expanded.
func (m AppModel) IsComposeOpen() bool {
	return m.compose != nil && m.compose.Expanded()
}

// FocusedColumn returns which column currently has keyboard focus.
func (m AppModel) FocusedColumn() layout.FocusColumn {
	return m.focus.Active()
}

// isTextInputFocused returns true when a text input is focused, so global
// shortcuts like 'n', 'q', '?' should not fire.
func (m AppModel) isTextInputFocused() bool {
	if m.compose != nil && m.compose.Expanded() {
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
	if m.currentView == ViewSplash && m.splash != nil {
		return m.splash.Init()
	}
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
		if m.compose != nil && m.compose.Expanded() {
			return m.updateCompose(msg)
		}

		// Column navigation (disabled when text input focused)
		if !m.isTextInputFocused() {
			switch msg.Type {
			case tea.KeyTab:
				m.focus.Next()
				return m, nil
			case tea.KeyShiftTab:
				m.focus.Prev()
				return m, nil
			}
		}

		// Global shortcuts (only when no text input focused)
		if !m.isTextInputFocused() {
			switch {
			case msg.Type == tea.KeyEsc:
				if m.focus.Active() != layout.FocusCenter {
					m.focus.Reset()
					return m, nil
				}
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

		// Route j/k/Enter to sidebar when left column is focused
		if m.focus.Active() == layout.FocusLeft && !m.isTextInputFocused() {
			switch {
			case msg.Type == tea.KeyDown || (msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'j'):
				if m.sidebarState.NavCursor < components.NavItemCount-1 {
					m.sidebarState.NavCursor++
				}
				return m, nil
			case msg.Type == tea.KeyUp || (msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'k'):
				if m.sidebarState.NavCursor > 0 {
					m.sidebarState.NavCursor--
				}
				return m, nil
			case msg.Type == tea.KeyEnter:
				switch m.sidebarState.NavCursor {
				case 0: // Home
					if m.currentView != ViewTimeline {
						m.currentView = ViewTimeline
						if m.timeline != nil {
							return m, m.timeline.FetchLatest()
						}
					}
				case 1: // Profile
					if m.user != nil {
						return m.openProfile(m.user.ID, true)
					}
				}
				return m, nil
			}
		}

		// Route j/k/Enter to discover panel when right column is focused
		if m.focus.Active() == layout.FocusRight && !m.isTextInputFocused() {
			switch {
			case msg.Type == tea.KeyDown || (msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'j'):
				if m.discoverState.ActiveSection == components.SectionTrending {
					if m.discoverState.TrendingCursor < components.TrendingCount()-1 {
						m.discoverState.TrendingCursor++
						m.adjustDiscoverScroll()
					}
				} else {
					if m.discoverState.WritersCursor < components.WritersCount()-1 {
						m.discoverState.WritersCursor++
						m.adjustDiscoverScroll()
					}
				}
				return m, nil
			case msg.Type == tea.KeyUp || (msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'k'):
				if m.discoverState.ActiveSection == components.SectionTrending {
					if m.discoverState.TrendingCursor > 0 {
						m.discoverState.TrendingCursor--
						m.adjustDiscoverScroll()
					}
				} else {
					if m.discoverState.WritersCursor > 0 {
						m.discoverState.WritersCursor--
						m.adjustDiscoverScroll()
					}
				}
				return m, nil
			case msg.Type == tea.KeyEnter:
				// Toggle between Trending and Writers sections
				if m.discoverState.ActiveSection == components.SectionTrending {
					m.discoverState.ActiveSection = components.SectionWriters
				} else {
					m.discoverState.ActiveSection = components.SectionTrending
				}
				return m, nil
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
		return m, nil

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
		var cmds []tea.Cmd
		if m.compose != nil {
			updated, cmd := m.compose.Update(msg)
			if cv, ok := updated.(ComposeViewModel); ok {
				m.compose = cv
			}
			cmds = append(cmds, cmd)
		}
		if m.timeline != nil {
			cmds = append(cmds, m.timeline.FetchLatest())
		}
		return m, tea.Batch(cmds...)

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

	case MsgServerConnected:
		// Server is reachable. Create login/register/timeline/compose views now.
		if m.factory != nil {
			m.login = m.factory.NewLogin(m.client)
			m.register = m.factory.NewRegister(m.client)
			m.timeline = m.factory.NewTimeline(m.client)
			m.compose = m.factory.NewCompose(m.client)
		}
		// Give newly created views their correct dimensions.
		if m.width > 0 {
			cols := layout.ComputeColumns(m.width)
			centerMsg := tea.WindowSizeMsg{Width: cols.Center, Height: m.height}
			fullMsg := tea.WindowSizeMsg{Width: m.width, Height: m.height}
			if m.login != nil {
				m.login, _ = m.login.Update(fullMsg)
			}
			if m.register != nil {
				m.register, _ = m.register.Update(fullMsg)
			}
			if m.timeline != nil {
				updated, _ := m.timeline.Update(centerMsg)
				if tl, ok := updated.(TimelineViewModel); ok {
					m.timeline = tl
				}
			}
			if m.compose != nil {
				updated, _ := m.compose.Update(centerMsg)
				if cv, ok := updated.(ComposeViewModel); ok {
					m.compose = cv
				}
			}
		}
		// If stored auth has tokens, skip login and go to timeline.
		if m.storedAuth != nil && m.storedAuth.AccessToken != "" {
			if m.client != nil {
				m.client.SetToken(m.storedAuth.AccessToken)
				m.client.SetRefreshToken(m.storedAuth.RefreshToken)
			}
			m.currentView = ViewTimeline
			if m.timeline != nil {
				return m, m.timeline.FetchLatest()
			}
			return m, nil
		}
		m.currentView = ViewLogin
		if m.login != nil {
			return m, m.login.Init()
		}
		return m, nil

	case MsgServerFailed:
		// Stay on splash — splash handles retry internally.
		if m.splash != nil {
			var updated ViewModel
			var cmd tea.Cmd
			updated, cmd = m.splash.Update(msg)
			if sv, ok := updated.(SplashViewModel); ok {
				m.splash = sv
			}
			return m, cmd
		}
		return m, nil
	}

	// Route to splash if it's the current view
	if m.currentView == ViewSplash {
		if m.splash != nil {
			var updated ViewModel
			var cmd tea.Cmd
			updated, cmd = m.splash.Update(msg)
			if sv, ok := updated.(SplashViewModel); ok {
				m.splash = sv
			}
			return m, cmd
		}
		return m, nil
	}

	return m.updateCurrentView(msg)
}

// View satisfies tea.Model.
func (m AppModel) View() string {
	// Splash: full-screen, no chrome
	if m.currentView == ViewSplash {
		if m.splash != nil {
			return m.splash.View()
		}
		return ""
	}

	// Auth views: centered, no sidebar
	if m.currentView == ViewLogin || m.currentView == ViewRegister {
		return m.viewCurrentContent()
	}

	// Authenticated views: three-column layout
	cols := layout.ComputeColumns(m.width)

	// Reserve 1 line for the column header bar rendered by RenderColumns.
	contentHeight := m.height - 1
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Center content: compose bar at top + main content below.
	var centerParts []string

	// Inline compose bar (always shown on authenticated views).
	if m.compose != nil && (m.currentView == ViewTimeline || m.currentView == ViewProfile) {
		centerParts = append(centerParts, m.compose.View())
	}

	// Main content.
	if m.help != nil {
		centerParts = append(centerParts, m.help.View())
	} else {
		centerParts = append(centerParts, m.viewCurrentContent())
	}

	centerContent := strings.Join(centerParts, "\n")

	// Left sidebar
	leftContent := components.RenderSidebar(
		m.user,
		components.View(m.currentView),
		m.focus.Active() == layout.FocusLeft,
		&m.sidebarState,
		cols.Left,
		contentHeight,
	)

	// Right sidebar — discover / trending panel
	rightContent := components.RenderDiscover(
		m.focus.Active() == layout.FocusRight,
		&m.discoverState,
		cols.Right,
		contentHeight,
	)

	return layout.RenderColumns(m.width, m.height, m.focus.Active(), leftContent, centerContent, rightContent)
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

	// Ensure compose bar exists for authenticated views.
	if m.compose == nil && m.factory != nil {
		m.compose = m.factory.NewCompose(m.client)
	}

	// Fetch timeline
	if m.timeline != nil {
		cmd := m.timeline.FetchLatest()
		return m, cmd
	}
	return m, nil
}

// openCompose expands the inline compose bar.
func (m AppModel) openCompose() (AppModel, tea.Cmd) {
	if m.compose != nil {
		m.compose.Expand()
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
	if m.compose != nil && m.compose.Expanded() {
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

// adjustDiscoverScroll keeps the cursor visible within the discover panel's
// scrollable sections by adjusting the scroll offset when the cursor moves
// outside the visible window. The visible item count mirrors the calculation
// in discover.go's renderTrendingSection/renderWritersSection.
func (m *AppModel) adjustDiscoverScroll() {
	// Compute visible items per section (mirrors discover.go logic)
	searchLines := 4
	availableHeight := m.height - searchLines - 2
	if availableHeight < 4 {
		availableHeight = 4
	}
	halfHeight := availableHeight / 2
	headerLines := 2
	itemHeight := 2
	visibleItems := (halfHeight - headerLines) / itemHeight
	if visibleItems < 1 {
		visibleItems = 1
	}

	if m.discoverState.ActiveSection == components.SectionTrending {
		if m.discoverState.TrendingCursor >= m.discoverState.TrendingScroll+visibleItems {
			m.discoverState.TrendingScroll = m.discoverState.TrendingCursor - visibleItems + 1
		}
		if m.discoverState.TrendingCursor < m.discoverState.TrendingScroll {
			m.discoverState.TrendingScroll = m.discoverState.TrendingCursor
		}
	} else {
		if m.discoverState.WritersCursor >= m.discoverState.WritersScroll+visibleItems {
			m.discoverState.WritersScroll = m.discoverState.WritersCursor - visibleItems + 1
		}
		if m.discoverState.WritersCursor < m.discoverState.WritersScroll {
			m.discoverState.WritersScroll = m.discoverState.WritersCursor
		}
	}
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

// updateCompose routes messages to the compose bar.
func (m AppModel) updateCompose(msg tea.Msg) (AppModel, tea.Cmd) {
	var updated ViewModel
	var cmd tea.Cmd
	updated, cmd = m.compose.Update(msg)
	if cv, ok := updated.(ComposeViewModel); ok {
		m.compose = cv
	}
	// Compose auto-collapses on Esc/publish via its own Update logic.
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
	case ViewSplash:
		if m.splash != nil {
			var updated ViewModel
			updated, cmd = m.splash.Update(msg)
			if sv, ok := updated.(SplashViewModel); ok {
				m.splash = sv
			}
		}
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

// propagateWindowSize sends the window size to all sub-models. Views that
// render inside the center column receive the center column width so they can
// word-wrap and center content correctly.
func (m AppModel) propagateWindowSize(msg tea.WindowSizeMsg) (AppModel, tea.Cmd) {
	var cmds []tea.Cmd

	// Center column dimensions for views rendered inside the three-column layout.
	// Subtract 1 from height for the column header bar.
	cols := layout.ComputeColumns(msg.Width)
	centerMsg := tea.WindowSizeMsg{Width: cols.Center, Height: msg.Height - 1}

	// Splash: full-screen, no columns.
	if m.splash != nil {
		var updated ViewModel
		var cmd tea.Cmd
		updated, cmd = m.splash.Update(msg)
		if sv, ok := updated.(SplashViewModel); ok {
			m.splash = sv
		}
		cmds = append(cmds, cmd)
	}

	// Auth views: full-screen, no columns.
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

	// Timeline and profile: center column width.
	if m.timeline != nil {
		var updated ViewModel
		var cmd tea.Cmd
		updated, cmd = m.timeline.Update(centerMsg)
		if tl, ok := updated.(TimelineViewModel); ok {
			m.timeline = tl
		}
		cmds = append(cmds, cmd)
	}
	if m.profile != nil {
		var updated ViewModel
		var cmd tea.Cmd
		updated, cmd = m.profile.Update(centerMsg)
		if pv, ok := updated.(ProfileViewModel); ok {
			m.profile = pv
		}
		cmds = append(cmds, cmd)
	}

	// Inline compose bar: center column width.
	if m.compose != nil {
		var updated ViewModel
		var cmd tea.Cmd
		updated, cmd = m.compose.Update(centerMsg)
		if cv, ok := updated.(ComposeViewModel); ok {
			m.compose = cv
		}
		cmds = append(cmds, cmd)
	}
	if m.help != nil {
		var updated ViewModel
		var cmd tea.Cmd
		updated, cmd = m.help.Update(centerMsg)
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
	case ViewSplash:
		if m.splash != nil {
			return m.splash.View()
		}
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
