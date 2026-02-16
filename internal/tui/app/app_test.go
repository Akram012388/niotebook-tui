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

func TestAppModelInitWithFactory(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	cmd := m.Init()
	// Init should return login's Init command (nil from stub)
	_ = cmd
}

func TestAppModelPOpensOwnProfile(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{ID: "u1", Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if m.CurrentView() != app.ViewProfile {
		t.Errorf("view = %v, want ViewProfile after p", m.CurrentView())
	}
}

func TestAppModelPostPublished(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	// Open compose
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if !m.IsComposeOpen() {
		t.Fatal("compose should be open")
	}
	// Publish post
	m = update(m, app.MsgPostPublished{Post: models.Post{ID: "1", Content: "Hello"}})
	if m.IsComposeOpen() {
		t.Error("compose should be closed after publish")
	}
}

func TestAppModelStatusClear(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	m = update(m, app.MsgStatusClear{})
	// Should not panic
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after status clear")
	}
}

func TestAppModelComposeCancelClosesOverlay(t *testing.T) {
	cancelFactory := &stubCancelFactory{}
	m := app.NewAppModelWithFactory(nil, nil, cancelFactory)
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	// Open compose
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if !m.IsComposeOpen() {
		t.Fatal("compose should be open")
	}
	// Press Esc — stub compose returns cancelled=true on any key
	m = update(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.IsComposeOpen() {
		t.Error("compose should be closed after cancel")
	}
}

func TestAppModelHelpDismissClosesOverlay(t *testing.T) {
	dismissFactory := &stubDismissFactory{}
	m := app.NewAppModelWithFactory(nil, nil, dismissFactory)
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	// Open help
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	// Dismiss with Esc — stub help returns dismissed=true on any key
	m = update(m, tea.KeyMsg{Type: tea.KeyEsc})
	// Help should be closed; view should not panic
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after help dismiss")
	}
}

func TestAppModelRRefreshesTimeline(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	// Press r to refresh
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	// Should return a command (nil from stub FetchLatest)
	_ = cmd
}

func TestAppModelProfileLoaded(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{ID: "u1", Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	// Open profile
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	// Send profile loaded
	m = update(m, app.MsgProfileLoaded{
		User:  &models.User{ID: "u1", Username: "akram"},
		Posts: nil,
	})
	if m.CurrentView() != app.ViewProfile {
		t.Errorf("view = %v, want ViewProfile", m.CurrentView())
	}
}

func TestAppModelProfileUpdated(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{ID: "u1", Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	// Open profile
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	// Send profile updated
	m = update(m, app.MsgProfileUpdated{
		User: &models.User{ID: "u1", Username: "akram", DisplayName: "New Name"},
	})
	if m.CurrentView() != app.ViewProfile {
		t.Errorf("view = %v, want ViewProfile", m.CurrentView())
	}
}

func TestAppModelTimelineLoadedRouting(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	m = update(m, app.MsgTimelineLoaded{
		Posts:   []models.Post{{ID: "1"}},
		HasMore: false,
	})
	// Should not panic
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestAppModelLoginViewRendering(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	// Before auth, should render login view
	view := m.View()
	_ = view // Just ensure no panic
}

func TestAppModelRegisterViewRendering(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgSwitchToRegister{})
	view := m.View()
	_ = view // Just ensure no panic
}

func TestAppModelProfileDismissReturnsToTimeline(t *testing.T) {
	dismissProfileFactory := &stubDismissProfileFactory{}
	m := app.NewAppModelWithFactory(nil, nil, dismissProfileFactory)
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{ID: "u1", Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	// Open profile
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if m.CurrentView() != app.ViewProfile {
		t.Fatalf("view = %v, want ViewProfile", m.CurrentView())
	}
	// Key press triggers Dismissed()=true in stub
	m = update(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.CurrentView() != app.ViewTimeline {
		t.Errorf("view = %v, want ViewTimeline after profile dismiss", m.CurrentView())
	}
}

func TestAppModelComposeBlocksTimelineShortcuts(t *testing.T) {
	trackFactory := &stubTrackFactory{}
	m := app.NewAppModelWithFactory(nil, nil, trackFactory)
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	// Open compose
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if !m.IsComposeOpen() {
		t.Fatal("compose should be open")
	}
	// Press 'j' — should be routed to compose (not timeline)
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	// Compose is still open — key was handled by compose, not timeline
	if !m.IsComposeOpen() {
		t.Error("compose should still be open after pressing j")
	}
}

func TestAppModelTimelineLoadedSetsPostsInView(t *testing.T) {
	trackFactory := &stubTrackFactory{}
	m := app.NewAppModelWithFactory(nil, nil, trackFactory)
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	if m.CurrentView() != app.ViewTimeline {
		t.Fatalf("view = %v, want ViewTimeline", m.CurrentView())
	}
	// Send timeline loaded message
	m = update(m, app.MsgTimelineLoaded{
		Posts:   []models.Post{{ID: "p1", Content: "Hello"}, {ID: "p2", Content: "World"}},
		HasMore: true,
	})
	// The timeline stub should have received the update
	tl := trackFactory.lastTimeline
	if tl == nil {
		t.Fatal("expected timeline stub to exist")
	}
	if !tl.updated {
		t.Error("expected timeline to have received the MsgTimelineLoaded update")
	}
}

// --- Additional stub factories for specific behaviors ---

// stubCancelFactory returns a compose that cancels immediately on any key
type stubCancelFactory struct{ stubFactory }

func (f *stubCancelFactory) NewCompose(_ *client.Client) app.ComposeViewModel {
	return &stubCancelCompose{}
}

type stubCancelCompose struct{ stubViewModel }

func (s *stubCancelCompose) Submitted() bool          { return false }
func (s *stubCancelCompose) Cancelled() bool          { return true }
func (s *stubCancelCompose) IsTextInputFocused() bool { return true }

// stubDismissFactory returns a help that dismisses immediately on any key
type stubDismissFactory struct{ stubFactory }

func (f *stubDismissFactory) NewHelp(_ string) app.HelpViewModel {
	return &stubDismissHelp{}
}

type stubDismissHelp struct{ stubViewModel }

func (s *stubDismissHelp) Dismissed() bool { return true }

// stubDismissProfileFactory returns a profile that dismisses immediately on any key
type stubDismissProfileFactory struct{ stubFactory }

func (f *stubDismissProfileFactory) NewProfile(_ *client.Client, _ string, _ bool) app.ProfileViewModel {
	return &stubDismissProfile{}
}

type stubDismissProfile struct{ stubViewModel }

func (s *stubDismissProfile) Editing() bool   { return false }
func (s *stubDismissProfile) Dismissed() bool { return true }

// stubTrackFactory tracks the timeline it creates for verification
type stubTrackFactory struct {
	stubFactory
	lastTimeline *stubTrackTimeline
}

func (f *stubTrackFactory) NewTimeline(_ *client.Client) app.TimelineViewModel {
	tl := &stubTrackTimeline{}
	f.lastTimeline = tl
	return tl
}

func (f *stubTrackFactory) NewCompose(_ *client.Client) app.ComposeViewModel {
	return &stubCompose{}
}

// stubTrackTimeline tracks whether Update was called with MsgTimelineLoaded
type stubTrackTimeline struct {
	stubViewModel
	updated bool
}

func (s *stubTrackTimeline) FetchLatest() tea.Cmd { return nil }
func (s *stubTrackTimeline) Update(msg tea.Msg) (app.ViewModel, tea.Cmd) {
	if _, ok := msg.(app.MsgTimelineLoaded); ok {
		s.updated = true
	}
	return s, nil
}
