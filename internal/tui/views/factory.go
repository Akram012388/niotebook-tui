package views

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/client"
)

// Factory implements app.ViewFactory, creating concrete view models.
type Factory struct{}

// NewFactory returns a new view factory.
func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) NewSplash(serverURL string) app.SplashViewModel {
	m := NewSplashModel(serverURL)
	return &splashAdapter{m}
}

func (f *Factory) NewLogin(c *client.Client) app.ViewModel {
	m := NewLoginModel(c)
	return &loginAdapter{m}
}

func (f *Factory) NewRegister(c *client.Client) app.ViewModel {
	m := NewRegisterModel(c)
	return &registerAdapter{m}
}

func (f *Factory) NewTimeline(c *client.Client) app.TimelineViewModel {
	m := NewTimelineModel(c)
	return &timelineAdapter{m}
}

func (f *Factory) NewProfile(c *client.Client, userID string, isOwn bool) app.ProfileViewModel {
	m := NewProfileModel(c, userID, isOwn)
	return &profileAdapter{m}
}

func (f *Factory) NewCompose(c *client.Client) app.ComposeViewModel {
	m := NewComposeModel(c)
	return &composeAdapter{m}
}

func (f *Factory) NewHelp(viewName string) app.HelpViewModel {
	m := NewHelpModel(viewName)
	return &helpAdapter{m}
}

// loginAdapter wraps LoginModel to implement app.ViewModel.
type loginAdapter struct {
	model LoginModel
}

func (a *loginAdapter) Init() tea.Cmd                       { return a.model.Init() }
func (a *loginAdapter) View() string                        { return a.model.View() }
func (a *loginAdapter) HelpText() string                    { return a.model.HelpText() }
func (a *loginAdapter) Update(msg tea.Msg) (app.ViewModel, tea.Cmd) {
	m, cmd := a.model.Update(msg)
	a.model = m
	return a, cmd
}

// registerAdapter wraps RegisterModel to implement app.ViewModel.
type registerAdapter struct {
	model RegisterModel
}

func (a *registerAdapter) Init() tea.Cmd                       { return a.model.Init() }
func (a *registerAdapter) View() string                        { return a.model.View() }
func (a *registerAdapter) HelpText() string                    { return a.model.HelpText() }
func (a *registerAdapter) Update(msg tea.Msg) (app.ViewModel, tea.Cmd) {
	m, cmd := a.model.Update(msg)
	a.model = m
	return a, cmd
}

// timelineAdapter wraps TimelineModel to implement app.TimelineViewModel.
type timelineAdapter struct {
	model TimelineModel
}

func (a *timelineAdapter) Init() tea.Cmd     { return a.model.Init() }
func (a *timelineAdapter) View() string      { return a.model.View() }
func (a *timelineAdapter) HelpText() string  { return a.model.HelpText() }
func (a *timelineAdapter) FetchLatest() tea.Cmd { return a.model.FetchLatest() }
func (a *timelineAdapter) Update(msg tea.Msg) (app.ViewModel, tea.Cmd) {
	m, cmd := a.model.Update(msg)
	a.model = m
	return a, cmd
}

// profileAdapter wraps ProfileModel to implement app.ProfileViewModel.
type profileAdapter struct {
	model ProfileModel
}

func (a *profileAdapter) Init() tea.Cmd     { return a.model.Init() }
func (a *profileAdapter) View() string      { return a.model.View() }
func (a *profileAdapter) HelpText() string  { return a.model.HelpText() }
func (a *profileAdapter) Editing() bool     { return a.model.Editing() }
func (a *profileAdapter) Dismissed() bool   { return a.model.Dismissed() }
func (a *profileAdapter) Update(msg tea.Msg) (app.ViewModel, tea.Cmd) {
	m, cmd := a.model.Update(msg)
	a.model = m
	return a, cmd
}

// composeAdapter wraps ComposeModel to implement app.ComposeViewModel.
type composeAdapter struct {
	model ComposeModel
}

func (a *composeAdapter) Init() tea.Cmd            { return a.model.Init() }
func (a *composeAdapter) View() string             { return a.model.View() }
func (a *composeAdapter) HelpText() string         { return a.model.HelpText() }
func (a *composeAdapter) Submitted() bool          { return a.model.Submitted() }
func (a *composeAdapter) Cancelled() bool          { return a.model.Cancelled() }
func (a *composeAdapter) IsTextInputFocused() bool { return a.model.IsTextInputFocused() }
func (a *composeAdapter) Update(msg tea.Msg) (app.ViewModel, tea.Cmd) {
	m, cmd := a.model.Update(msg)
	a.model = m
	return a, cmd
}

// helpAdapter wraps HelpModel to implement app.HelpViewModel.
type helpAdapter struct {
	model HelpModel
}

func (a *helpAdapter) Init() tea.Cmd     { return a.model.Init() }
func (a *helpAdapter) View() string      { return a.model.View() }
func (a *helpAdapter) HelpText() string  { return a.model.HelpText() }
func (a *helpAdapter) Dismissed() bool   { return a.model.Dismissed() }
func (a *helpAdapter) Update(msg tea.Msg) (app.ViewModel, tea.Cmd) {
	m, cmd := a.model.Update(msg)
	a.model = m
	return a, cmd
}

// splashAdapter wraps SplashModel to implement app.SplashViewModel.
type splashAdapter struct {
	model SplashModel
}

func (a *splashAdapter) Init() tea.Cmd        { return a.model.Init() }
func (a *splashAdapter) View() string         { return a.model.View() }
func (a *splashAdapter) HelpText() string     { return a.model.HelpText() }
func (a *splashAdapter) Done() bool           { return a.model.Done() }
func (a *splashAdapter) Failed() bool         { return a.model.Failed() }
func (a *splashAdapter) ErrorMessage() string { return a.model.ErrorMessage() }
func (a *splashAdapter) Update(msg tea.Msg) (app.ViewModel, tea.Cmd) {
	m, cmd := a.model.Update(msg)
	a.model = m
	return a, cmd
}
