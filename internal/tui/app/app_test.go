package app_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/client"
	"github.com/Akram012388/niotebook-tui/internal/tui/config"
	"github.com/Akram012388/niotebook-tui/internal/tui/layout"
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
	expanded  bool
}

func (s *stubCompose) Submitted() bool          { return false }
func (s *stubCompose) Cancelled() bool          { return s.cancelled }
func (s *stubCompose) Expanded() bool           { return s.expanded }
func (s *stubCompose) IsTextInputFocused() bool { return s.expanded }
func (s *stubCompose) Expand()                  { s.expanded = true }
func (s *stubCompose) Update(msg tea.Msg) (app.ViewModel, tea.Cmd) {
	if _, ok := msg.(app.MsgPostPublished); ok {
		s.expanded = false
	}
	return s, nil
}

type stubHelp struct{ stubViewModel }

func (s *stubHelp) Dismissed() bool { return false }

type stubSplash struct{ stubViewModel }

func (s *stubSplash) Done() bool         { return false }
func (s *stubSplash) Failed() bool       { return false }
func (s *stubSplash) ErrorMessage() string { return "" }

type stubFactory struct{}

func (f *stubFactory) NewSplash(_ string) app.SplashViewModel             { return &stubSplash{} }
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

// connectAndAuth simulates the splash -> server connected -> auth success flow
// that most tests need to reach the authenticated state.
func connectAndAuth(m app.AppModel, user *models.User, tokens *models.TokenPair) app.AppModel {
	m = update(m, app.MsgServerConnected{})
	m = update(m, app.MsgAuthSuccess{User: user, Tokens: tokens})
	return m
}

// connectServer simulates the splash -> server connected flow to reach login.
func connectServer(m app.AppModel) app.AppModel {
	return update(m, app.MsgServerConnected{})
}

func TestAppModelStartsOnLogin(t *testing.T) {
	m := app.NewAppModel(nil, nil) // no stored auth
	if m.CurrentView() != app.ViewLogin {
		t.Errorf("initial view = %v, want ViewLogin", m.CurrentView())
	}
}

func TestAppModelWithFactoryStartsOnSplash(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	if m.CurrentView() != app.ViewSplash {
		t.Errorf("initial view = %v, want ViewSplash", m.CurrentView())
	}
}

func TestAppModelServerConnectedTransitionsToLogin(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = update(m, app.MsgServerConnected{})
	if m.CurrentView() != app.ViewLogin {
		t.Errorf("view after server connected = %v, want ViewLogin", m.CurrentView())
	}
}

func TestAppModelServerConnectedWithStoredAuthGoesToTimeline(t *testing.T) {
	storedAuth := &config.StoredAuth{
		AccessToken:  "stored-token",
		RefreshToken: "stored-refresh",
	}
	m := app.NewAppModelWithFactory(nil, storedAuth, &stubFactory{}, "")
	m = update(m, app.MsgServerConnected{})
	if m.CurrentView() != app.ViewTimeline {
		t.Errorf("view = %v, want ViewTimeline with stored auth", m.CurrentView())
	}
}

func TestAppModelServerFailedStaysOnSplash(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = update(m, app.MsgServerFailed{Err: "connection refused"})
	if m.CurrentView() != app.ViewSplash {
		t.Errorf("view = %v, want ViewSplash after server failed", m.CurrentView())
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
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if !m.IsComposeOpen() {
		t.Error("expected compose to be open after pressing n")
	}
}

func TestAppModelQuestionMarkOpensHelp(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	// Help should now be open — view name should reflect it
	if m.View() == "" {
		t.Error("expected non-empty view with help overlay")
	}
}

func TestAppModelQQuitsOnTimeline(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
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
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectServer(m)
	m = update(m, app.MsgSwitchToRegister{})
	if m.CurrentView() != app.ViewRegister {
		t.Errorf("view = %v, want ViewRegister", m.CurrentView())
	}
}

func TestAppModelSwitchToLogin(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectServer(m)
	m = update(m, app.MsgSwitchToRegister{}) // go to register first
	m = update(m, app.MsgSwitchToLogin{})
	if m.CurrentView() != app.ViewLogin {
		t.Errorf("view = %v, want ViewLogin", m.CurrentView())
	}
}

func TestAppModelAPIErrorShowsInStatusBar(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = update(m, app.MsgAPIError{Message: "server error"})
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after API error")
	}
}

func TestAppModelAuthExpiredReturnsToLogin(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, app.MsgAuthExpired{})
	if m.CurrentView() != app.ViewLogin {
		t.Errorf("view = %v, want ViewLogin after auth expired", m.CurrentView())
	}
}

func TestAppModelWindowResize(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	// Should not panic
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after resize")
	}
}

func TestAppModelInitWithFactory(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	cmd := m.Init()
	// Init should return splash's Init command (nil from stub)
	_ = cmd
}

func TestAppModelPOpensOwnProfile(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{ID: "u1", Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if m.CurrentView() != app.ViewProfile {
		t.Errorf("view = %v, want ViewProfile after p", m.CurrentView())
	}
}

func TestAppModelPostPublished(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	// Open compose
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if !m.IsComposeOpen() {
		t.Fatal("compose should be open")
	}
	// Publish post — compose collapses but is not nil
	m = update(m, app.MsgPostPublished{Post: models.Post{ID: "1", Content: "Hello"}})
	if m.IsComposeOpen() {
		t.Error("compose should be collapsed (not expanded) after publish")
	}
}

func TestAppModelStatusClear(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = update(m, app.MsgStatusClear{})
	// Should not panic
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after status clear")
	}
}

func TestAppModelComposeCancelClosesOverlay(t *testing.T) {
	cancelFactory := &stubCancelFactory{}
	m := app.NewAppModelWithFactory(nil, nil, cancelFactory, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	// Open compose
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if !m.IsComposeOpen() {
		t.Fatal("compose should be open")
	}
	// Press Esc — compose collapses (not nil, just not expanded)
	m = update(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.IsComposeOpen() {
		t.Error("compose should be collapsed after cancel")
	}
}

func TestAppModelHelpDismissClosesOverlay(t *testing.T) {
	dismissFactory := &stubDismissFactory{}
	m := app.NewAppModelWithFactory(nil, nil, dismissFactory, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
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
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	// Press r to refresh
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	// Should return a command (nil from stub FetchLatest)
	_ = cmd
}

func TestAppModelProfileLoaded(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{ID: "u1", Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
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
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{ID: "u1", Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
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
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
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
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectServer(m)
	// After server connect, should render login view
	view := m.View()
	_ = view // Just ensure no panic
}

func TestAppModelRegisterViewRendering(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectServer(m)
	m = update(m, app.MsgSwitchToRegister{})
	view := m.View()
	_ = view // Just ensure no panic
}

func TestAppModelProfileDismissReturnsToTimeline(t *testing.T) {
	dismissProfileFactory := &stubDismissProfileFactory{}
	m := app.NewAppModelWithFactory(nil, nil, dismissProfileFactory, "")
	m = connectAndAuth(m, &models.User{ID: "u1", Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
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
	m := app.NewAppModelWithFactory(nil, nil, trackFactory, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
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
	m := app.NewAppModelWithFactory(nil, nil, trackFactory, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	if m.CurrentView() != app.ViewTimeline {
		t.Fatalf("view = %v, want ViewTimeline", m.CurrentView())
	}
	// Send timeline loaded message
	_ = update(m, app.MsgTimelineLoaded{
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

func TestAppModelHelpOnProfile(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{ID: "u1", Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	// Open profile
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if m.CurrentView() != app.ViewProfile {
		t.Fatalf("view = %v, want ViewProfile", m.CurrentView())
	}
	// Open help from profile
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view with help on profile")
	}
}

func TestAppModelWindowResizeWithOverlays(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{ID: "u1", Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	// Open profile
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	// Open compose
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	// Resize with compose+profile open
	m = update(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after resize with overlays")
	}
}

func TestAppModelWindowResizeWithHelp(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = update(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after resize with help")
	}
}

func TestAppModelNFromProfile(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{ID: "u1", Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	// Open profile
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	// Open compose from profile
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if !m.IsComposeOpen() {
		t.Error("expected compose to open from profile view")
	}
}

func TestAppModelComposeCreatedOnServerConnected(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	// Compose should exist (always present) but not expanded
	if m.IsComposeOpen() {
		t.Error("compose should not be expanded by default")
	}
	// The view should render without panic — compose bar visible in center column
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view with inline compose bar")
	}
}

func TestAppModelComposeBarRendersOnAuthenticatedViews(t *testing.T) {
	renderFactory := &stubRenderFactory{}
	m := app.NewAppModelWithFactory(nil, nil, renderFactory, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	// On timeline, compose bar output should be part of the view
	view := m.View()
	if !strings.Contains(view, "[compose-bar]") {
		t.Error("expected compose bar to render inline on timeline view")
	}

	// Open profile — compose bar should still render
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	view = m.View()
	if !strings.Contains(view, "[compose-bar]") {
		t.Error("expected compose bar to render inline on profile view")
	}
}

func TestAppModelComposeCreatedOnAuthSuccess(t *testing.T) {
	// Test that compose is created during handleAuthSuccess if not yet created
	m := app.NewAppModel(nil, nil)
	// No factory initially, so compose won't be created
	// But with factory and auth success, compose should be created
	m2 := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m2 = update(m2, app.MsgServerConnected{})
	// Now auth success — compose should exist after this
	m2 = update(m2, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	// Compose exists but is not expanded
	if m2.IsComposeOpen() {
		t.Error("compose should not be expanded after auth")
	}
	_ = m // silence unused
}

func TestAppModelRegisterKeyRouting(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectServer(m)
	m = update(m, app.MsgSwitchToRegister{})
	m = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	// Key routing to register view — should not panic
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	_ = m.View()
	if m.CurrentView() != app.ViewRegister {
		t.Errorf("view = %v, want ViewRegister", m.CurrentView())
	}
}

func TestAppModelProfileViewContent(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{ID: "u1", Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	view := m.View()
	if view == "" {
		t.Error("expected non-empty profile view content")
	}
}

func TestAppModelComposeViewName(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view with compose")
	}
}

func TestAppModelInitWithoutFactory(t *testing.T) {
	m := app.NewAppModel(nil, nil)
	cmd := m.Init()
	// Without factory, login is nil, should return nil
	if cmd != nil {
		t.Error("expected nil cmd from Init without factory")
	}
}

// --- Additional stub factories for specific behaviors ---

// stubCancelFactory returns a compose that cancels immediately on any key
type stubCancelFactory struct{ stubFactory }

func (f *stubCancelFactory) NewCompose(_ *client.Client) app.ComposeViewModel {
	return &stubCancelCompose{}
}

type stubCancelCompose struct {
	stubViewModel
	expanded bool
}

func (s *stubCancelCompose) Submitted() bool          { return false }
func (s *stubCancelCompose) Cancelled() bool          { return !s.expanded }
func (s *stubCancelCompose) Expanded() bool           { return s.expanded }
func (s *stubCancelCompose) IsTextInputFocused() bool { return s.expanded }
func (s *stubCancelCompose) Expand()                  { s.expanded = true }
func (s *stubCancelCompose) Update(msg tea.Msg) (app.ViewModel, tea.Cmd) {
	// Any key collapses — simulates Esc cancelling
	s.expanded = false
	return s, nil
}

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

// stubRenderFactory returns compose that renders identifiable output
type stubRenderFactory struct{ stubFactory }

func (f *stubRenderFactory) NewCompose(_ *client.Client) app.ComposeViewModel {
	return &stubRenderCompose{}
}

type stubRenderCompose struct {
	stubViewModel
	expanded bool
}

func (s *stubRenderCompose) Submitted() bool          { return false }
func (s *stubRenderCompose) Cancelled() bool          { return false }
func (s *stubRenderCompose) Expanded() bool           { return s.expanded }
func (s *stubRenderCompose) IsTextInputFocused() bool { return s.expanded }
func (s *stubRenderCompose) Expand()                  { s.expanded = true }
func (s *stubRenderCompose) View() string             { return "[compose-bar]" }

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
	return &stubCompose{expanded: false}
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

// --- Column focus navigation tests ---

func TestAppModelTabCyclesColumns(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	// Default is center.
	if got := m.FocusedColumn(); got != layout.FocusCenter {
		t.Fatalf("initial focus = %v, want FocusCenter", got)
	}

	// Tab -> right.
	m = update(m, tea.KeyMsg{Type: tea.KeyTab})
	if got := m.FocusedColumn(); got != layout.FocusRight {
		t.Errorf("after 1st Tab focus = %v, want FocusRight", got)
	}

	// Tab -> left.
	m = update(m, tea.KeyMsg{Type: tea.KeyTab})
	if got := m.FocusedColumn(); got != layout.FocusLeft {
		t.Errorf("after 2nd Tab focus = %v, want FocusLeft", got)
	}

	// Tab -> center.
	m = update(m, tea.KeyMsg{Type: tea.KeyTab})
	if got := m.FocusedColumn(); got != layout.FocusCenter {
		t.Errorf("after 3rd Tab focus = %v, want FocusCenter", got)
	}

	// Should render without errors
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after Tab cycling")
	}
}

func TestAppModelShiftTabCyclesReverse(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	// Shift+Tab from center -> left.
	m = update(m, tea.KeyMsg{Type: tea.KeyShiftTab})
	if got := m.FocusedColumn(); got != layout.FocusLeft {
		t.Errorf("after Shift+Tab focus = %v, want FocusLeft", got)
	}

	// Shift+Tab from left -> right.
	m = update(m, tea.KeyMsg{Type: tea.KeyShiftTab})
	if got := m.FocusedColumn(); got != layout.FocusRight {
		t.Errorf("after 2nd Shift+Tab focus = %v, want FocusRight", got)
	}

	// Shift+Tab from right -> center.
	m = update(m, tea.KeyMsg{Type: tea.KeyShiftTab})
	if got := m.FocusedColumn(); got != layout.FocusCenter {
		t.Errorf("after 3rd Shift+Tab focus = %v, want FocusCenter", got)
	}

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after Shift+Tab")
	}
}

func TestAppModelEscFromSideColumnReturnsToCenter(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	// Tab to right
	m = update(m, tea.KeyMsg{Type: tea.KeyTab})
	if got := m.FocusedColumn(); got != layout.FocusRight {
		t.Fatalf("after Tab focus = %v, want FocusRight", got)
	}

	// Esc should reset to center
	m = update(m, tea.KeyMsg{Type: tea.KeyEsc})
	if got := m.FocusedColumn(); got != layout.FocusCenter {
		t.Errorf("after Esc focus = %v, want FocusCenter", got)
	}

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after Esc from side column")
	}
}

func TestAppModelNFromLeftColumnOpensCompose(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	// Tab to right, then tab to left
	m = update(m, tea.KeyMsg{Type: tea.KeyTab})
	m = update(m, tea.KeyMsg{Type: tea.KeyTab})
	if got := m.FocusedColumn(); got != layout.FocusLeft {
		t.Fatalf("after 2 Tabs focus = %v, want FocusLeft", got)
	}

	// Press n — should open compose
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if !m.IsComposeOpen() {
		t.Error("n from left column should open compose")
	}
}
