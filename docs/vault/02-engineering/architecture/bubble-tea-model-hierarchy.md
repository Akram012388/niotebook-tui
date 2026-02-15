---
title: "Bubble Tea Model Hierarchy"
created: 2026-02-15
updated: 2026-02-15
status: accepted
tags: [engineering, tui, architecture, bubbletea]
---

# Bubble Tea Model Hierarchy

This document specifies the exact model structure, message flow, and component composition for the Niotebook TUI. It is the blueprint for all TUI code.

## Root Model: AppModel

The TUI has a single root model that owns all state and delegates to sub-models based on the current view.

```go
// internal/tui/app/model.go

type View int

const (
    ViewLogin View = iota
    ViewRegister
    ViewTimeline
    ViewProfile
    ViewOwnProfile
)

type AppModel struct {
    // Current view
    currentView View

    // Sub-models (each manages its own state)
    login    LoginModel
    register RegisterModel
    timeline TimelineModel
    profile  ProfileModel

    // Shared state
    client     *api.Client   // HTTP client wrapper
    user       *models.User  // Authenticated user (nil if not logged in)
    windowSize tea.WindowSizeMsg

    // Overlay state
    showCompose bool
    compose     ComposeModel
    showHelp    bool
    help        HelpModel

    // Status bar
    statusBar StatusBarModel
}
```

## Init() Behavior

```go
func (m AppModel) Init() tea.Cmd {
    // On startup:
    // 1. Load config from ~/.config/niotebook/config.yaml
    // 2. Check for stored JWT in ~/.config/niotebook/auth.json
    // 3. If valid token exists: set user, switch to ViewTimeline, return cmd to fetch timeline
    // 4. If no token or expired: switch to ViewLogin, return nil
    return tea.Batch(
        loadConfig,      // reads config file
        checkAuthToken,  // reads auth.json, validates expiry
    )
}
```

Init() runs once at startup. Sub-model Init() methods are called when switching to that view (not at app startup).

## Update() — Message Routing

```go
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd

    // Global messages handled first (regardless of current view)
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.windowSize = msg
        // Propagate to all sub-models
        m.timeline, _ = m.timeline.Update(msg)
        m.profile, _ = m.profile.Update(msg)
        return m, nil

    case tea.KeyMsg:
        // Global keys (when no text input is focused)
        if !m.isTextInputFocused() {
            switch msg.String() {
            case "ctrl+c":
                return m, tea.Quit
            case "?":
                m.showHelp = !m.showHelp
                return m, nil
            }
        }

    case msgAuthSuccess:
        m.user = msg.User
        m.client.SetToken(msg.Tokens.AccessToken)
        m.currentView = ViewTimeline
        m.statusBar.SetSuccess("Welcome, @" + msg.User.Username)
        return m, m.timeline.Init() // fetch timeline

    case msgAuthExpired:
        m.user = nil
        m.currentView = ViewLogin
        m.statusBar.SetError("Session expired. Please log in again.")
        return m, nil

    case msgAPIError:
        m.statusBar.SetError(msg.Message)
        return m, nil
    }

    // Overlay routing (compose and help take priority)
    if m.showHelp {
        m.help, cmd = m.help.Update(msg)
        if m.help.dismissed {
            m.showHelp = false
        }
        return m, cmd
    }

    if m.showCompose {
        m.compose, cmd = m.compose.Update(msg)
        if m.compose.submitted {
            m.showCompose = false
            m.statusBar.SetSuccess("Post published.")
            // Refresh timeline to show new post
            return m, m.timeline.FetchLatest()
        }
        if m.compose.cancelled {
            m.showCompose = false
        }
        return m, cmd
    }

    // View-specific routing
    switch m.currentView {
    case ViewLogin:
        m.login, cmd = m.login.Update(msg)
    case ViewRegister:
        m.register, cmd = m.register.Update(msg)
    case ViewTimeline:
        m.timeline, cmd = m.timeline.Update(msg)
        // Handle timeline-specific key actions
        if key, ok := msg.(tea.KeyMsg); ok {
            switch key.String() {
            case "n":
                m.showCompose = true
                m.compose = NewComposeModel(m.client, m.windowSize)
                return m, m.compose.Init()
            case "enter":
                if post := m.timeline.SelectedPost(); post != nil {
                    m.profile = NewProfileModel(m.client, post.Author.ID, m.windowSize)
                    m.currentView = ViewProfile
                    return m, m.profile.Init()
                }
            case "p":
                m.profile = NewProfileModel(m.client, m.user.ID, m.windowSize)
                m.currentView = ViewOwnProfile
                return m, m.profile.Init()
            }
        }
    case ViewProfile, ViewOwnProfile:
        m.profile, cmd = m.profile.Update(msg)
        if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
            m.currentView = ViewTimeline
            return m, nil // timeline state preserved
        }
    }

    return m, cmd
}
```

## View() — Rendering

```go
func (m AppModel) View() string {
    // Calculate content area height
    contentHeight := m.windowSize.Height - 2 // minus header and status bar

    // Minimum terminal size check
    if m.windowSize.Width < 40 || m.windowSize.Height < 10 {
        return lipgloss.Place(
            m.windowSize.Width, m.windowSize.Height,
            lipgloss.Center, lipgloss.Center,
            "Terminal too small.\nResize to at least 40x10.",
        )
    }

    // Build header
    header := m.renderHeader()

    // Build content based on current view
    var content string
    switch m.currentView {
    case ViewLogin:
        content = m.login.View()
    case ViewRegister:
        content = m.register.View()
    case ViewTimeline:
        content = m.timeline.View()
    case ViewProfile, ViewOwnProfile:
        content = m.profile.View()
    }

    // Apply content area constraints
    content = lipgloss.NewStyle().
        Height(contentHeight).
        MaxHeight(contentHeight).
        Render(content)

    // Build status bar
    statusBar := m.statusBar.View(m.currentView, m.windowSize.Width)

    // Compose vertical layout
    screen := lipgloss.JoinVertical(lipgloss.Left, header, content, statusBar)

    // Overlay compose modal if open
    if m.showCompose {
        screen = m.overlayCompose(screen)
    }

    // Overlay help if open
    if m.showHelp {
        screen = m.overlayHelp(screen)
    }

    return screen
}
```

## Sub-Model Contracts

Each sub-model implements `tea.Model` and follows these contracts:

### TimelineModel

```go
type TimelineModel struct {
    posts      []models.Post
    cursor     int              // index of selected post
    nextCursor string           // pagination cursor from server
    hasMore    bool
    loading    bool
    viewport   viewport.Model
    client     *api.Client
}

// Init() returns a Cmd that fetches the first page of the timeline
func (m TimelineModel) Init() tea.Cmd

// SelectedPost() returns the currently highlighted post (for profile navigation)
func (m TimelineModel) SelectedPost() *models.Post

// FetchLatest() returns a Cmd that refreshes from the top (no cursor)
func (m TimelineModel) FetchLatest() tea.Cmd
```

### ComposeModel

```go
type ComposeModel struct {
    textarea  textarea.Model
    client    *api.Client
    submitted bool    // true when post was published
    cancelled bool    // true when user pressed Esc
    posting   bool    // true while HTTP request in flight
    err       error
}
```

### ProfileModel

```go
type ProfileModel struct {
    user     *models.User
    posts    []models.Post
    isOwn    bool          // true if viewing own profile
    editing  bool          // true if edit modal is open
    viewport viewport.Model
    client   *api.Client
}
```

### LoginModel / RegisterModel

```go
type LoginModel struct {
    emailInput    textinput.Model
    passwordInput textinput.Model
    focusIndex    int           // 0 = email, 1 = password
    submitting    bool
    err           string        // validation error message
    errField      string        // which field has the error
    client        *api.Client
}
```

## Message Types

All custom messages returned by async Cmds:

```go
// Auth
type msgAuthSuccess struct {
    User   *models.User
    Tokens *models.TokenPair
}
type msgAuthExpired struct{}
type msgAuthError struct{ Message string; Field string }

// Timeline
type msgTimelineLoaded struct {
    Posts      []models.Post
    NextCursor string
    HasMore    bool
}
type msgTimelineRefreshed struct {
    Posts      []models.Post
    NextCursor string
    HasMore    bool
}

// Posts
type msgPostPublished struct{ Post models.Post }

// Profile
type msgProfileLoaded struct {
    User  *models.User
    Posts []models.Post
}
type msgProfileUpdated struct{ User *models.User }

// Generic
type msgAPIError struct{ Message string }
type msgStatusClear struct{} // sent after 5s timer to clear status bar
```

## Async HTTP Pattern

All API calls follow this pattern — return a `tea.Cmd` that performs the HTTP request and sends back a message:

```go
func fetchTimeline(client *api.Client, cursor string, limit int) tea.Cmd {
    return func() tea.Msg {
        resp, err := client.GetTimeline(cursor, limit)
        if err != nil {
            return msgAPIError{Message: err.Error()}
        }
        return msgTimelineLoaded{
            Posts:      resp.Posts,
            NextCursor: resp.NextCursor,
            HasMore:    resp.HasMore,
        }
    }
}
```

This is Bubble Tea's standard async pattern. The Cmd runs in a goroutine managed by Bubble Tea's runtime. The returned message is sent back to Update() on the main thread. No manual goroutine management needed.

## isTextInputFocused()

Determines whether key events should be routed to a text input (suppressing global shortcuts like `q`):

```go
func (m AppModel) isTextInputFocused() bool {
    if m.showCompose {
        return true
    }
    switch m.currentView {
    case ViewLogin:
        return true
    case ViewRegister:
        return true
    }
    if m.currentView == ViewOwnProfile && m.profile.editing {
        return true
    }
    return false
}
```

## State Preservation

- **Timeline → Profile → Timeline:** Timeline state (scroll position, loaded posts, cursor) is **preserved**. Pressing `Esc` from profile returns to the same scroll position.
- **Timeline → Compose → Timeline:** Timeline state preserved. Compose modal is an overlay, not a view switch.
- **Login → Timeline:** Login state is discarded after successful auth.
- **Refresh (r):** Replaces all loaded posts with fresh data. Scroll position resets to top.

## Help Overlay

Rendered as a centered bordered box over the current view (same overlay technique as compose modal). Content is a static list of keybindings for the current view, generated from a map:

```go
var helpBindings = map[View][]HelpEntry{
    ViewTimeline: {
        {"j/k", "Scroll up/down"},
        {"n", "New post"},
        {"r", "Refresh"},
        {"Enter", "View profile"},
        {"p", "Own profile"},
        {"g/G", "Top/bottom"},
        {"?", "Close help"},
        {"q", "Quit"},
    },
    // ... other views
}
```

Dismissed by pressing `?`, `Esc`, or `q`.
