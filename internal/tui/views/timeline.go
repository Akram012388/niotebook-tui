package views

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/client"
	"github.com/Akram012388/niotebook-tui/internal/tui/components"
	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

var (
	emptyStateStyle = lipgloss.NewStyle().
			Foreground(theme.TextMuted).Italic(true)
	tlLoadingStyle = lipgloss.NewStyle().
			Foreground(theme.Warning)
)

// TimelineModel manages the timeline view state.
type TimelineModel struct {
	posts      []models.Post
	cursor     int
	scrollTop  int
	nextCursor string
	hasMore    bool
	loading    bool
	client     *client.Client
	width      int
	height     int
}

// NewTimelineModel creates a new timeline view model.
func NewTimelineModel(c *client.Client) TimelineModel {
	return TimelineModel{
		client: c,
	}
}

// SetPosts replaces the current posts list.
func (m *TimelineModel) SetPosts(posts []models.Post) {
	m.posts = posts
	m.cursor = 0
	m.scrollTop = 0
}

// CursorIndex returns the current cursor position.
func (m TimelineModel) CursorIndex() int {
	return m.cursor
}

// SelectedPost returns the currently highlighted post, or nil if none.
func (m TimelineModel) SelectedPost() *models.Post {
	if len(m.posts) == 0 || m.cursor < 0 || m.cursor >= len(m.posts) {
		return nil
	}
	return &m.posts[m.cursor]
}

// Init returns the initial command to fetch the timeline.
func (m TimelineModel) Init() tea.Cmd {
	return m.fetchTimeline("")
}

// FetchLatest returns a command that refreshes the timeline from the top.
func (m TimelineModel) FetchLatest() tea.Cmd {
	return m.fetchTimeline("")
}

func (m TimelineModel) fetchTimeline(cursor string) tea.Cmd {
	c := m.client
	return func() tea.Msg {
		if c == nil {
			return app.MsgAPIError{Message: "no server connection"}
		}
		resp, err := c.GetTimeline(cursor, 20)
		if err != nil {
			return app.MsgAPIError{Message: err.Error()}
		}
		nextCursor := ""
		if resp.NextCursor != nil {
			nextCursor = *resp.NextCursor
		}
		return app.MsgTimelineLoaded{
			Posts:      resp.Posts,
			NextCursor: nextCursor,
			HasMore:    resp.HasMore,
		}
	}
}


// Update handles messages for the timeline view.
func (m TimelineModel) Update(msg tea.Msg) (TimelineModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case app.MsgTimelineLoaded:
		m.loading = false
		m.posts = msg.Posts
		m.nextCursor = msg.NextCursor
		m.hasMore = msg.HasMore
		m.cursor = 0
		m.scrollTop = 0
		return m, nil

	case app.MsgTimelineRefreshed:
		m.loading = false
		m.posts = msg.Posts
		m.nextCursor = msg.NextCursor
		m.hasMore = msg.HasMore
		m.cursor = 0
		m.scrollTop = 0
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m TimelineModel) handleKey(msg tea.KeyMsg) (TimelineModel, tea.Cmd) {
	postCount := len(m.posts)
	if postCount == 0 {
		return m, nil
	}

	visibleCount := m.visiblePostCount()

	switch {
	// j or down arrow: move cursor down
	case msg.Type == tea.KeyDown || (msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'j'):
		if m.cursor < postCount-1 {
			m.cursor++
			if m.cursor >= m.scrollTop+visibleCount {
				m.scrollTop = m.cursor - visibleCount + 1
			}
		}

	// k or up arrow: move cursor up
	case msg.Type == tea.KeyUp || (msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'k'):
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.scrollTop {
				m.scrollTop = m.cursor
			}
		}

	// g: jump to top
	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'g':
		m.cursor = 0
		m.scrollTop = 0

	// G: jump to bottom
	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'G':
		m.cursor = postCount - 1
		if postCount > visibleCount {
			m.scrollTop = postCount - visibleCount
		}

	// Space: page down
	case msg.Type == tea.KeySpace:
		m.cursor += visibleCount
		if m.cursor >= postCount {
			m.cursor = postCount - 1
		}
		m.scrollTop = m.cursor - visibleCount + 1
		if m.scrollTop < 0 {
			m.scrollTop = 0
		}

	// b: page up
	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'b':
		m.cursor -= visibleCount
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.scrollTop = m.cursor
	}

	return m, nil
}

func (m TimelineModel) visiblePostCount() int {
	if m.height <= 0 {
		return 5
	}
	// Estimate ~4 lines per post card (header + content + separator + spacing)
	count := m.height / 4
	if count < 1 {
		count = 1
	}
	return count
}

// View renders the timeline view.
func (m TimelineModel) View() string {
	if m.loading {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			tlLoadingStyle.Render("Loading timeline..."))
	}

	if len(m.posts) == 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			emptyStateStyle.Render("No posts yet. Press n to compose one!"))
	}

	now := time.Now()
	var b strings.Builder

	visibleCount := m.visiblePostCount()
	end := m.scrollTop + visibleCount
	if end > len(m.posts) {
		end = len(m.posts)
	}

	for i := m.scrollTop; i < end; i++ {
		selected := i == m.cursor
		card := components.RenderPostCard(m.posts[i], m.width, selected, now)
		b.WriteString(card)
		b.WriteString("\n")
	}

	return b.String()
}

// HelpText returns the status bar help text for the timeline view.
func (m TimelineModel) HelpText() string {
	return "j/k: navigate  n: compose  r: refresh  ?: help  q: quit"
}
