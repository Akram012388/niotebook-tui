package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/client"
)

var (
	formBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(1, 2)

	formTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("5")).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	buttonStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("5"))

	errMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1"))

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Faint(true)
)

// LoginModel manages the login form state.
type LoginModel struct {
	emailInput    textinput.Model
	passwordInput textinput.Model
	focusIndex    int
	submitting    bool
	err           string
	errField      string
	client        *client.Client
	width         int
	height        int
}

// NewLoginModel creates a new login form model.
func NewLoginModel(c *client.Client) LoginModel {
	email := textinput.New()
	email.Placeholder = "you@example.com"
	email.CharLimit = 254
	email.Width = 30
	email.Focus()

	password := textinput.New()
	password.Placeholder = "password"
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '●'
	password.CharLimit = 128
	password.Width = 30

	return LoginModel{
		emailInput:    email,
		passwordInput: password,
		client:        c,
	}
}

// FocusIndex returns the currently focused field index.
func (m LoginModel) FocusIndex() int {
	return m.focusIndex
}

// Init returns the initial command (cursor blink).
func (m LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages for the login view.
func (m LoginModel) Update(msg tea.Msg) (LoginModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case app.MsgAuthError:
		m.submitting = false
		m.err = msg.Message
		m.errField = msg.Field
		if msg.Field == "password" || msg.Field == "" {
			m.passwordInput.SetValue("")
		}
		return m, nil

	case tea.KeyMsg:
		// Clear previous error on any keypress
		if m.err != "" && msg.Type != tea.KeyEnter {
			m.err = ""
			m.errField = ""
		}

		switch msg.Type {
		case tea.KeyTab:
			if m.focusIndex == 0 {
				m.focusIndex = 1
				m.emailInput.Blur()
				return m, m.passwordInput.Focus()
			}
			// Tab on password field → switch to register view
			return m, func() tea.Msg { return app.MsgSwitchToRegister{} }

		case tea.KeyShiftTab:
			if m.focusIndex == 1 {
				m.focusIndex = 0
				m.passwordInput.Blur()
				return m, m.emailInput.Focus()
			}
			return m, nil

		case tea.KeyEnter:
			return m, m.submit()
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	if m.focusIndex == 0 {
		m.emailInput, cmd = m.emailInput.Update(msg)
	} else {
		m.passwordInput, cmd = m.passwordInput.Update(msg)
	}
	return m, cmd
}

// submit validates and submits the login form.
func (m *LoginModel) submit() tea.Cmd {
	email := strings.TrimSpace(m.emailInput.Value())
	password := m.passwordInput.Value()

	if email == "" {
		m.err = "email is required"
		m.errField = "email"
		return nil
	}
	if password == "" {
		m.err = "password is required"
		m.errField = "password"
		return nil
	}

	m.submitting = true
	c := m.client

	return func() tea.Msg {
		if c == nil {
			return app.MsgAuthError{Message: "no server connection"}
		}
		resp, err := c.Login(email, password)
		if err != nil {
			return app.MsgAuthError{Message: err.Error()}
		}
		return app.MsgAuthSuccess{
			User:   resp.User,
			Tokens: resp.Tokens,
		}
	}
}

// View renders the login form.
func (m LoginModel) View() string {
	var b strings.Builder

	b.WriteString(formTitleStyle.Render("Login"))
	b.WriteString("\n\n")

	// Email field
	b.WriteString(labelStyle.Render("Email"))
	b.WriteString("\n")
	b.WriteString(m.emailInput.View())
	b.WriteString("\n")
	if m.errField == "email" && m.err != "" {
		b.WriteString(errMsgStyle.Render(m.err))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Password field
	b.WriteString(labelStyle.Render("Password"))
	b.WriteString("\n")
	b.WriteString(m.passwordInput.View())
	b.WriteString("\n")
	if m.errField == "password" && m.err != "" {
		b.WriteString(errMsgStyle.Render(m.err))
		b.WriteString("\n")
	}
	if m.errField == "" && m.err != "" {
		b.WriteString(errMsgStyle.Render(m.err))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Submit button
	if m.submitting {
		b.WriteString(hintStyle.Render("Logging in..."))
	} else {
		b.WriteString(buttonStyle.Render("[Enter] Login"))
	}
	b.WriteString("\n\n")

	// Register hint
	b.WriteString(hintStyle.Render("No account?"))
	b.WriteString("\n")
	b.WriteString(hintStyle.Render("[Tab] Register"))

	form := formBoxStyle.Render(b.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, form)
}

// HelpText returns the status bar help text for the login view.
func (m LoginModel) HelpText() string {
	return "Tab: switch to register  Enter: submit  q: quit"
}
