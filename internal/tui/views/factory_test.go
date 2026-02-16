package views_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akram012388/niotebook-tui/internal/tui/views"
)

func TestNewFactory(t *testing.T) {
	f := views.NewFactory()
	if f == nil {
		t.Fatal("NewFactory returned nil")
	}
}

func TestFactoryNewSplash(t *testing.T) {
	f := views.NewFactory()
	vm := f.NewSplash("http://localhost:8080")
	if vm == nil {
		t.Fatal("NewSplash returned nil")
	}
	_ = vm.View()
	_ = vm.HelpText()
	if vm.Done() {
		t.Error("new splash should not be done")
	}
	if vm.Failed() {
		t.Error("new splash should not have failed")
	}
	if vm.ErrorMessage() != "" {
		t.Error("new splash should have empty error message")
	}
	_, _ = vm.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
}

func TestFactoryNewLogin(t *testing.T) {
	f := views.NewFactory()
	vm := f.NewLogin(nil)
	if vm == nil {
		t.Fatal("NewLogin returned nil")
	}
	_ = vm.View()
	_ = vm.HelpText()
	_, _ = vm.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
}

func TestFactoryNewRegister(t *testing.T) {
	f := views.NewFactory()
	vm := f.NewRegister(nil)
	if vm == nil {
		t.Fatal("NewRegister returned nil")
	}
	_ = vm.View()
	_ = vm.HelpText()
	_, _ = vm.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
}

func TestFactoryNewTimeline(t *testing.T) {
	f := views.NewFactory()
	vm := f.NewTimeline(nil)
	if vm == nil {
		t.Fatal("NewTimeline returned nil")
	}
	_ = vm.View()
	_ = vm.HelpText()
	_, _ = vm.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
}

func TestFactoryNewProfile(t *testing.T) {
	f := views.NewFactory()
	vm := f.NewProfile(nil, "user-1", true)
	if vm == nil {
		t.Fatal("NewProfile returned nil")
	}
	_ = vm.View()
	_ = vm.HelpText()
	if vm.Editing() {
		t.Error("new profile should not be editing")
	}
	if vm.Dismissed() {
		t.Error("new profile should not be dismissed")
	}
	_, _ = vm.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
}

func TestFactoryNewCompose(t *testing.T) {
	f := views.NewFactory()
	vm := f.NewCompose(nil)
	if vm == nil {
		t.Fatal("NewCompose returned nil")
	}
	_ = vm.View()
	_ = vm.HelpText()
	if vm.Submitted() {
		t.Error("new compose should not be submitted")
	}
	if vm.Cancelled() {
		t.Error("new compose should not be cancelled")
	}
	_ = vm.IsTextInputFocused()
	_, _ = vm.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
}

func TestFactoryNewHelp(t *testing.T) {
	f := views.NewFactory()
	vm := f.NewHelp("timeline")
	if vm == nil {
		t.Fatal("NewHelp returned nil")
	}
	_ = vm.View()
	_ = vm.HelpText()
	if vm.Dismissed() {
		t.Error("new help should not be dismissed")
	}
	_, _ = vm.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
}

func TestFactoryTimelineFetchLatest(t *testing.T) {
	f := views.NewFactory()
	vm := f.NewTimeline(nil)
	// FetchLatest with nil client returns nil cmd
	cmd := vm.FetchLatest()
	_ = cmd
}
