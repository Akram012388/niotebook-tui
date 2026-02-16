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
)

var (
	profileUsernameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("6")).
				Bold(true)

	profileDisplayNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("7"))

	profileBioStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	profileJoinedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Faint(true)

	profileSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("5")).
				Bold(true)

	profileSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8"))
)

// ProfileModel manages the profile view state.
type ProfileModel struct {
	user      *models.User
	posts     []models.Post
	cursor    int
	scrollTop int
	loading   bool
	editing   bool
	dismissed bool
	isOwn     bool
	client    *client.Client
	width     int
	height    int
}

// NewProfileModel creates a new profile view model.
func NewProfileModel(c *client.Client, userID string, isOwn bool) ProfileModel {
	m := ProfileModel{
		client:  c,
		isOwn:   isOwn,
		loading: true,
	}
	return m
}

// Init returns the initial command to fetch the profile.
func (m ProfileModel) Init() tea.Cmd {
	return m.fetchProfile()
}

func (m ProfileModel) fetchProfile() tea.Cmd {
	c := m.client
	isOwn := m.isOwn
	var userID string
	if m.user != nil {
		userID = m.user.ID
	}
	return func() tea.Msg {
		if c == nil {
			return app.MsgAPIError{Message: "no server connection"}
		}

		id := userID
		if isOwn {
			id = "me"
		}

		user, err := c.GetUser(id)
		if err != nil {
			return app.MsgAPIError{Message: err.Error()}
		}

		resp, err := c.GetUserPosts(user.ID, "", 20)
		if err != nil {
			return app.MsgAPIError{Message: err.Error()}
		}

		return app.MsgProfileLoaded{
			User:  user,
			Posts: resp.Posts,
		}
	}
}

// User returns the loaded user, if any.
func (m ProfileModel) User() *models.User {
	return m.user
}

// Editing returns whether the edit modal is active.
func (m ProfileModel) Editing() bool {
	return m.editing
}

// Dismissed returns whether the user pressed Esc to leave.
func (m ProfileModel) Dismissed() bool {
	return m.dismissed
}

// IsOwn returns whether this is the current user's own profile.
func (m ProfileModel) IsOwn() bool {
	return m.isOwn
}

// Update handles messages for the profile view.
func (m ProfileModel) Update(msg tea.Msg) (ProfileModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case app.MsgProfileLoaded:
		m.loading = false
		m.user = msg.User
		m.posts = msg.Posts
		m.cursor = 0
		m.scrollTop = 0
		return m, nil

	case app.MsgProfileUpdated:
		m.user = msg.User
		m.editing = false
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m ProfileModel) handleKey(msg tea.KeyMsg) (ProfileModel, tea.Cmd) {
	switch {
	case msg.Type == tea.KeyEsc:
		m.dismissed = true
		return m, nil

	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'e':
		if m.isOwn {
			m.editing = true
		}
		return m, nil

	case msg.Type == tea.KeyDown || (msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'j'):
		if len(m.posts) > 0 && m.cursor < len(m.posts)-1 {
			m.cursor++
			visibleCount := m.visiblePostCount()
			if m.cursor >= m.scrollTop+visibleCount {
				m.scrollTop = m.cursor - visibleCount + 1
			}
		}
		return m, nil

	case msg.Type == tea.KeyUp || (msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'k'):
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.scrollTop {
				m.scrollTop = m.cursor
			}
		}
		return m, nil

	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'g':
		m.cursor = 0
		m.scrollTop = 0
		return m, nil

	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'G':
		if len(m.posts) > 0 {
			m.cursor = len(m.posts) - 1
			visibleCount := m.visiblePostCount()
			if len(m.posts) > visibleCount {
				m.scrollTop = len(m.posts) - visibleCount
			}
		}
		return m, nil
	}

	return m, nil
}

func (m ProfileModel) visiblePostCount() int {
	if m.height <= 0 {
		return 5
	}
	// Profile header takes ~8 lines, each post ~4 lines
	available := m.height - 8
	if available < 4 {
		available = 4
	}
	count := available / 4
	if count < 1 {
		count = 1
	}
	return count
}

// View renders the profile view.
func (m ProfileModel) View() string {
	if m.loading {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			loadingStyle.Render("Loading profile..."))
	}

	if m.user == nil {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			emptyStateStyle.Render("Profile not found."))
	}

	var b strings.Builder

	// Username
	b.WriteString(profileUsernameStyle.Render("@" + m.user.Username))
	b.WriteString("\n")

	// Display name
	if m.user.DisplayName != "" {
		b.WriteString(profileDisplayNameStyle.Render(m.user.DisplayName))
		b.WriteString("\n")
	}

	b.WriteString(profileSeparatorStyle.Render(strings.Repeat("─", m.width)))
	b.WriteString("\n")

	// Bio
	if m.user.Bio != "" {
		b.WriteString(profileBioStyle.Render(m.user.Bio))
		b.WriteString("\n")
	}

	// Joined date
	joined := m.user.CreatedAt.Format("Jan 2006")
	b.WriteString(profileJoinedStyle.Render("Joined " + joined))
	b.WriteString("\n")

	b.WriteString(profileSeparatorStyle.Render(strings.Repeat("─", m.width)))
	b.WriteString("\n\n")

	// Posts section
	b.WriteString(profileSectionStyle.Render("Posts"))
	b.WriteString("\n")
	b.WriteString(profileSeparatorStyle.Render(strings.Repeat("─", m.width)))
	b.WriteString("\n")

	if len(m.posts) == 0 {
		b.WriteString(emptyStateStyle.Render("No posts yet."))
		b.WriteString("\n")
	} else {
		now := time.Now()
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
	}

	// Edit hint for own profile
	if m.isOwn {
		b.WriteString("\n")
		b.WriteString(hintStyle.Render("[e] Edit profile"))
	}

	return b.String()
}

// HelpText returns the status bar help text for the profile view.
func (m ProfileModel) HelpText() string {
	if m.isOwn {
		return "j/k: scroll  e: edit bio  Esc: back  ?: help"
	}
	return "j/k: scroll  Esc: back  ?: help"
}
