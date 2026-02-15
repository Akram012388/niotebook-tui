---
title: "Product Requirements Document: Niotebook MVP"
version: 0.1.0
created: 2026-02-15
updated: 2026-02-15
status: accepted
tags: [prd, product, mvp]
---

# Product Requirements Document: Niotebook MVP

## 1. Product Overview

### 1.1 What Is Niotebook?

Niotebook is a **standalone social media platform** delivered as a TUI (Text User Interface) application. It provides an X/Twitter-like social experience entirely within the terminal for developers and power users who work in CLI-first environments.

### 1.2 What Niotebook Is NOT

- **Not an X/Twitter client.** Niotebook is its own platform with its own backend, user accounts, and social graph. No X API integration. No X branding or content.
- **Not a wrapper around existing social networks.** All content is native to Niotebook.
- **Not a web application.** The primary (and for MVP, only) interface is the TUI.

### 1.3 Why Niotebook?

Developers using agentic coding tools (Claude Code, Codex, OpenCode, Cursor, VS Code terminal) spend their workday in the terminal. Switching to a browser for social media is a context switch. Niotebook eliminates that by placing the social experience directly in the terminal.

The name combines **NIO** (non-blocking I/O, symbolizing efficient data streams) and **notebook** (a personal, always-on digital journal).

### 1.4 Target Audience

- Developers and sysadmins who live in the terminal
- Users of agentic coding tools (Claude Code, Codex, OpenCode, Cursor)
- CLI enthusiasts, Linux/macOS power users
- Remote workers on SSH sessions
- Anyone frustrated with bloated web-based social media

## 2. MVP Scope

### 2.1 MVP Feature Set

The MVP ships with the **absolute minimum** to be a functional social platform:

| Feature | Description |
|---------|-------------|
| **User Registration & Login** | Email + password authentication via the TUI. JWT-based sessions stored locally. |
| **Home Timeline** | Scrollable feed showing posts from all users (global feed for MVP — no follow graph yet). |
| **Post Composition** | Compose and publish text posts from the TUI. Character limit TBD. |
| **User Profiles** | View a user's bio and post history. Edit own profile. |
| **Manual Feed Refresh** | User presses a key to fetch latest posts. No auto-polling or real-time push. |

### 2.2 Explicitly NOT in MVP

These features are deferred to v1.1+:

- Follow/unfollow (and personalized timeline)
- Likes, reposts, quote posts
- Replies and threaded conversations
- Search and discovery
- Notifications
- Direct messages
- Media attachments
- Real-time updates (WebSockets/polling)
- Offline mode / caching
- Multi-account support
- Federation (ActivityPub/Mastodon)
- Plugin system
- SSH key authentication

### 2.3 MVP Success Criteria

- A user can register, log in, post, and browse a timeline entirely from the TUI
- The TUI renders cleanly in common terminal emulators (iTerm2, Alacritty, kitty, macOS Terminal, tmux)
- The server runs on a single VPS with PostgreSQL
- The project compiles to two Go binaries (`niotebook-server`, `niotebook-tui`)

## 3. User Stories

### 3.1 Registration & Authentication

- **As a new user**, I can create an account by providing a username, email, and password in the TUI, so that I can start using Niotebook.
- **As a returning user**, I can log in with my email and password, so that I can resume my session.
- **As a logged-in user**, my session persists across TUI restarts (JWT stored locally), so that I don't need to log in every time.

### 3.2 Timeline

- **As a user**, I can see a feed of recent posts from all users when I open the app, so that I can discover content.
- **As a user**, I can scroll through the feed using keyboard shortcuts (j/k or arrow keys), so that I can browse efficiently.
- **As a user**, I can press a key to refresh the feed, so that I can see new posts.
- **As a user**, I can see the author, timestamp, and content of each post, so that I can understand who posted what and when.

### 3.3 Posting

- **As a user**, I can open a compose modal by pressing a key (e.g., `n`), so that I can write a new post.
- **As a user**, I can see a character count while composing, so that I know how much space I have left.
- **As a user**, I can publish my post by pressing a key (e.g., Ctrl+Enter), so that it appears in the global timeline.
- **As a user**, I can cancel composition and return to the timeline, so that I can change my mind.

### 3.4 Profiles

- **As a user**, I can view another user's profile by selecting their name in the timeline, so that I can learn about them.
- **As a user**, I can see a user's bio and their recent posts on their profile page.
- **As a user**, I can edit my own bio and display name.

## 4. Non-Functional Requirements

### 4.1 Performance

- TUI startup to usable timeline: < 2 seconds on local network
- Post publish round-trip: < 500ms
- Feed refresh: < 1 second for 50 posts

### 4.2 Security

- Passwords hashed with bcrypt (cost 12+)
- JWT tokens with expiry (24h access, 7d refresh)
- All API communication over HTTPS in production
- No plaintext password storage anywhere
- Input sanitization on all user-provided content

### 4.3 Compatibility

- Go 1.22+
- Terminal support: 256-color and true-color terminals
- OS support: macOS, Linux (primary), Windows (best-effort)
- Terminal emulators: iTerm2, Alacritty, kitty, macOS Terminal, tmux, screen

### 4.4 Deployment

- Server: Single Go binary + PostgreSQL on a VPS (DigitalOcean/Hetzner)
- TUI client: Single Go binary distributed via `go install`, GitHub releases, or Homebrew
- No Docker required (but Dockerfile provided for convenience)

## 5. Resolved Questions

- [x] Character limit for posts? **140 characters.** See [[02-engineering/adr/ADR-0010-post-character-limit|ADR-0010]].
- [x] Global timeline ordering? **Cursor-based pagination, reverse chronological.** See [[02-engineering/adr/ADR-0012-cursor-pagination|ADR-0012]].
- [x] Username constraints? **3-15 chars, alphanumeric + underscores, case-insensitive.** See [[02-engineering/adr/ADR-0011-username-rules|ADR-0011]].
- [x] Rate limiting strategy? **Per-IP token bucket (60 req/min general).** See [[02-engineering/adr/ADR-0014-rate-limiting|ADR-0014]].
- [x] Content moderation for MVP? **Deferred. Small invite-only community.** See [[02-engineering/adr/ADR-0017-no-moderation-mvp|ADR-0017]].
