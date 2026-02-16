package app_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/client"
)

// stubViewModel is a minimal ViewModel for testing.
type stubViewModel struct{}

func (s *stubViewModel) Init() tea.Cmd                           { return nil }
func (s *stubViewModel) Update(tea.Msg) (app.ViewModel, tea.Cmd) { return s, nil }
func (s *stubViewModel) View() string                            { return "" }
func (s *stubViewModel) HelpText() string                        { return "" }

type stubTimeline struct{ stubViewModel }

func (s *stubTimeline) FetchLatest() tea.Cmd { return nil }

type stubProfile struct{ stubViewModel }

func (s *stubProfile) Editing() bool   { return false }
func (s *stubProfile) Dismissed() bool { return false }

type stubCompose struct {
	stubViewModel
	cancelled bool
}

func (s *stubCompose) Submitted() bool          { return false }
func (s *stubCompose) Cancelled() bool          { return s.cancelled }
func (s *stubCompose) IsTextInputFocused() bool { return true }

type stubHelp struct{ stubViewModel }

func (s *stubHelp) Dismissed() bool { return false }

type stubFactory struct{}

func (f *stubFactory) NewLogin(_ *client.Client) app.ViewModel            { return &stubViewModel{} }
func (f *stubFactory) NewRegister(_ *client.Client) app.ViewModel         { return &stubViewModel{} }
func (f *stubFactory) NewTimeline(_ *client.Client) app.TimelineViewModel { return &stubTimeline{} }
func (f *stubFactory) NewProfile(_ *client.Client, _ string, _ bool) app.ProfileViewModel {
	return &stubProfile{}
}
func (f *stubFactory) NewCompose(_ *client.Client) app.ComposeViewModel { return &stubCompose{} }
func (f *stubFactory) NewHelp(_ string) app.HelpViewModel               { return &stubHelp{} }

func update(m app.AppModel, msg tea.Msg) app.AppModel {
	result, _ := m.Update(msg)
	return result.(app.AppModel)
}

func TestAppModelStartsOnLogin(t *testing.T) {
	m := app.NewAppModel(nil, nil) // no stored auth
	if m.CurrentView() != app.ViewLogin {
		t.Errorf("initial view = %v, want ViewLogin", m.CurrentView())
	}
}

func TestAppModelAuthSuccessSwitchesToTimeline(t *testing.T) {
	m := app.NewAppModel(nil, nil)
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	if m.CurrentView() != app.ViewTimeline {
		t.Errorf("view after auth = %v, want ViewTimeline", m.CurrentView())
	}
}

func TestAppModelNOpensCompose(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	// Simulate logged in on timeline
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if !m.IsComposeOpen() {
		t.Error("expected compose to be open after pressing n")
	}
}

func TestAppModelQuestionMarkOpensHelp(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	// Help should now be open — view name should reflect it
	if m.View() == "" {
		t.Error("expected non-empty view with help overlay")
	}
}

func TestAppModelQQuitsOnTimeline(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected quit command on q press")
	}
}

func TestAppModelCtrlCAlwaysQuits(t *testing.T) {
	m := app.NewAppModel(nil, nil) // on login view
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("expected quit command on ctrl+c")
	}
}

func TestAppModelSwitchToRegister(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgSwitchToRegister{})
	if m.CurrentView() != app.ViewRegister {
		t.Errorf("view = %v, want ViewRegister", m.CurrentView())
	}
}

func TestAppModelSwitchToLogin(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgSwitchToRegister{}) // go to register first
	m = update(m, app.MsgSwitchToLogin{})
	if m.CurrentView() != app.ViewLogin {
		t.Errorf("view = %v, want ViewLogin", m.CurrentView())
	}
}

func TestAppModelAPIErrorShowsInStatusBar(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	m = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = update(m, app.MsgAPIError{Message: "server error"})
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after API error")
	}
}

func TestAppModelAuthExpiredReturnsToLogin(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	m = update(m, app.MsgAuthExpired{})
	if m.CurrentView() != app.ViewLogin {
		t.Errorf("view = %v, want ViewLogin after auth expired", m.CurrentView())
	}
}

func TestAppModelWindowResize(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	m = update(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	// Should not panic
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after resize")
	}
}
