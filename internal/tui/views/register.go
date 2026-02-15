package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/client"
)

// MsgSwitchToLogin tells the parent to switch to the login view.
type MsgSwitchToLogin struct{}

const (
	usernameMinLen = 3
	usernameMaxLen = 15
	passwordMinLen = 8
)

// RegisterModel manages the register form state.
type RegisterModel struct {
	usernameInput textinput.Model
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

// NewRegisterModel creates a new register form model.
func NewRegisterModel(c *client.Client) RegisterModel {
	username := textinput.New()
	username.Placeholder = "username"
	username.CharLimit = usernameMaxLen
	username.Width = 30
	username.Focus()

	email := textinput.New()
	email.Placeholder = "you@example.com"
	email.CharLimit = 254
	email.Width = 30

	password := textinput.New()
	password.Placeholder = "min 8 characters"
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '●'
	password.CharLimit = 128
	password.Width = 30

	return RegisterModel{
		usernameInput: username,
		emailInput:    email,
		passwordInput: password,
		client:        c,
	}
}

// FocusIndex returns the currently focused field index.
func (m RegisterModel) FocusIndex() int {
	return m.focusIndex
}

// Init returns the initial command.
func (m RegisterModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages for the register view.
func (m RegisterModel) Update(msg tea.Msg) (RegisterModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case app.MsgAuthError:
		m.submitting = false
		m.err = msg.Message
		m.errField = msg.Field
		return m, nil

	case tea.KeyMsg:
		if m.err != "" && msg.Type != tea.KeyEnter {
			m.err = ""
			m.errField = ""
		}

		switch msg.Type {
		case tea.KeyTab:
			return m.nextField()

		case tea.KeyShiftTab:
			return m.prevField()

		case tea.KeyEnter:
			return m, m.submit()
		}
	}

	var cmd tea.Cmd
	switch m.focusIndex {
	case 0:
		m.usernameInput, cmd = m.usernameInput.Update(msg)
	case 1:
		m.emailInput, cmd = m.emailInput.Update(msg)
	case 2:
		m.passwordInput, cmd = m.passwordInput.Update(msg)
	}
	return m, cmd
}

// nextField moves focus to the next field.
func (m RegisterModel) nextField() (RegisterModel, tea.Cmd) {
	if m.focusIndex >= 2 {
		// Tab past last field → switch to login view
		return m, func() tea.Msg { return MsgSwitchToLogin{} }
	}

	m.blurAll()
	m.focusIndex++
	return m, m.focusCurrent()
}

// prevField moves focus to the previous field.
func (m RegisterModel) prevField() (RegisterModel, tea.Cmd) {
	if m.focusIndex <= 0 {
		return m, nil
	}

	m.blurAll()
	m.focusIndex--
	return m, m.focusCurrent()
}

func (m *RegisterModel) blurAll() {
	m.usernameInput.Blur()
	m.emailInput.Blur()
	m.passwordInput.Blur()
}

func (m *RegisterModel) focusCurrent() tea.Cmd {
	switch m.focusIndex {
	case 0:
		return m.usernameInput.Focus()
	case 1:
		return m.emailInput.Focus()
	case 2:
		return m.passwordInput.Focus()
	}
	return nil
}

// submit validates and submits the register form.
func (m *RegisterModel) submit() tea.Cmd {
	username := strings.TrimSpace(m.usernameInput.Value())
	email := strings.TrimSpace(m.emailInput.Value())
	password := m.passwordInput.Value()

	// Client-side validation
	if username == "" {
		m.err = "username is required"
		m.errField = "username"
		return nil
	}
	if len(username) < usernameMinLen {
		m.err = fmt.Sprintf("username must be at least %d characters", usernameMinLen)
		m.errField = "username"
		return nil
	}
	if email == "" {
		m.err = "email is required"
		m.errField = "email"
		return nil
	}
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		m.err = "invalid email format"
		m.errField = "email"
		return nil
	}
	if len(password) < passwordMinLen {
		m.err = fmt.Sprintf("password must be at least %d characters", passwordMinLen)
		m.errField = "password"
		return nil
	}

	m.submitting = true
	c := m.client

	return func() tea.Msg {
		if c == nil {
			return app.MsgAuthError{Message: "no server connection"}
		}
		resp, err := c.Register(username, email, password)
		if err != nil {
			return app.MsgAuthError{Message: err.Error()}
		}
		return app.MsgAuthSuccess{
			User:   resp.User,
			Tokens: resp.Tokens,
		}
	}
}

// View renders the register form.
func (m RegisterModel) View() string {
	var b strings.Builder

	b.WriteString(formTitleStyle.Render("Register"))
	b.WriteString("\n\n")

	// Username field
	usernameLen := len(m.usernameInput.Value())
	counterColor := "8"
	if usernameLen > 0 && usernameLen < usernameMinLen {
		counterColor = "1"
	}
	counterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(counterColor)).Faint(true)

	b.WriteString(labelStyle.Render("Username") + " " + counterStyle.Render(fmt.Sprintf("%d/%d", usernameLen, usernameMaxLen)))
	b.WriteString("\n")
	b.WriteString(m.usernameInput.View())
	b.WriteString("\n")
	if m.errField == "username" && m.err != "" {
		b.WriteString(errMsgStyle.Render(m.err))
		b.WriteString("\n")
	}
	b.WriteString("\n")

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
	pwLen := len(m.passwordInput.Value())
	pwCounterColor := "8"
	if pwLen > 0 && pwLen < passwordMinLen {
		pwCounterColor = "1"
	}
	pwCounterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(pwCounterColor)).Faint(true)

	b.WriteString(labelStyle.Render("Password") + " " + pwCounterStyle.Render(fmt.Sprintf("%d chars", pwLen)))
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
		b.WriteString(hintStyle.Render("Registering..."))
	} else {
		b.WriteString(buttonStyle.Render("[Enter] Register"))
	}
	b.WriteString("\n\n")

	// Login hint
	b.WriteString(hintStyle.Render("Already have an account?"))
	b.WriteString("\n")
	b.WriteString(hintStyle.Render("[Tab] Login"))

	form := formBoxStyle.Render(b.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, form)
}

// HelpText returns the status bar help text for the register view.
func (m RegisterModel) HelpText() string {
	return "Tab: switch fields/login  Enter: submit  q: quit"
}
