> **Note:** This document is the original brainstorm that seeded the project. It has been superseded by the documentation vault at `docs/vault/`. The vault is the source of truth for all product, engineering, and design decisions. Key divergences from this document include: Niotebook is a standalone platform (not an X client), uses stdlib `net/http` (not Cobra), PostgreSQL (not BoltDB/Badger), and the MVP scope excludes likes, reposts, replies, and search.

# Project Brief: Niotebook

## Project Overview
**Project Name:** Niotebook  
**Tagline:** "X for the Terminal: A Nerd's Notebook for Real-Time Social Streams"  
**Owner/Developer:** Shaikh Akram Ahmed (@CodeAkram on X)  
**Domain:** www.niotebook.com  
**Launch Date (Target):** Q2 2026 (initial MVP by April 2026)  
**Version:** 0.1.0 (MVP)  
**License:** MIT (open-source to encourage community contributions)  
**Repository:** github.com/CodeAkram/niotebook  

Niotebook is an innovative, terminal-based Text User Interface (TUI) application designed to replicate and enhance the core experience of X (formerly Twitter) within the confines of a command-line environment. Tailored for developers, sysadmins, power users, and "nerds" who prefer keyboard-driven workflows over graphical interfaces, Niotebook transforms the terminal into a lightweight, efficient social media client. It emphasizes speed, concurrency, and minimalism while delivering a premium UI/UX feel—think infinite scrolling timelines, threaded conversations, real-time notifications, and seamless posting, all without leaving your shell.

The name "Niotebook" draws from "NIO" (inspired by non-blocking I/O, symbolizing efficient data streams) and "notebook" (evoking a personal, always-on digital journal for public notes and interactions). This positions it as a "notebook for the open web," where users can jot down thoughts, engage in discussions, and follow feeds in a distraction-free terminal space. Unlike web or mobile apps, Niotebook prioritizes privacy (no tracking pixels), performance (runs natively with low resource usage), and extensibility (via plugins or scripts).

The project aims to bridge the gap between traditional social media and CLI tools, making it ideal for scenarios like server-side monitoring of trends, automated posting from scripts, or quick checks during coding sessions. It will integrate with X's API (or compatible alternatives like Mastodon for federation) to fetch and post content, with built-in support for handling rate limits and authentication securely.

## Project Objectives
- **Primary Goal:** Create a fully functional TUI client for X-like social interactions, optimized for terminal users who value efficiency and concurrency.
- **Key Objectives:**
  - Deliver a responsive, immersive TUI experience that rivals modern apps in usability, with smooth scrolling, keyboard shortcuts, and visual polish.
  - Ensure scalability through Go's native concurrency, supporting real-time updates and multiple background tasks without UI lag.
  - Maintain simplicity: Single binary distribution, no external runtimes, and easy installation (e.g., via `go install` or Homebrew).
  - Foster an open-source community: Encourage forks for custom features like theme engines or integrations with tools (e.g., Vim for composing posts).
  - Prioritize user privacy and security: Local token storage, optional offline mode for cached feeds, and no unnecessary data collection.
- **Target Audience:** CLI enthusiasts, developers, Linux/macOS users, remote workers on SSH sessions, and anyone frustrated with bloated web apps.
- **Success Metrics:** 1,000+ GitHub stars in first year, active contributions, positive feedback on r/golang and r/commandline, and integration into dotfiles setups.

## Detailed Description
Niotebook operates as a standalone TUI application that launches into a full-screen (or windowed) terminal interface, taking over the alt-screen for an immersive experience. Upon startup, users authenticate via OAuth (storing tokens securely in a local config file) and are greeted with a customizable home view: a scrolling timeline of posts from followed accounts.

The app mimics X's core loop—browse, engage, create—but adapts it for terminal constraints:
- **Navigation & Interaction:** Keyboard-centric (Vim/Emacs-inspired bindings: j/k for scroll, / for search, n for new post). Mouse support optional for selection.
- **Real-Time Elements:** Background goroutines poll or use WebSockets for updates, pushing notifications (e.g., bell sounds or popups) without interrupting the main view.
- **Customization:** Themes via config files (e.g., dark mode, color schemes), key remapping, and modular views (split-screen for timeline + trends).
- **Edge Cases:** Handles terminal resizes gracefully, supports Unicode/emoji for rich post rendering, and includes accessibility features like high-contrast modes.
- **Extensibility:** Plugin system (e.g., Go modules) for adding features like media previews (text-based ASCII art or external viewer integration) or bots.

Potential expansions post-MVP: Federation with ActivityPub (e.g., Mastodon compatibility), multi-account support, or a companion server for custom backends.

## Features
### Core Features (MVP)
- **Timeline/Feed:** Infinite scrolling viewport for home, mentions, or search results. Lazy loading of posts with pagination.
- **Posting & Composing:** Modal textarea for new posts/DMs/replies, with character count, emoji picker, and thread support.
- **Engagement:** Like, repost, reply, quote—keyboard shortcuts trigger API calls asynchronously.
- **Search & Discovery:** Keyword search with filters (e.g., users, hashtags), displaying results in a selectable list.
- **Notifications & DMs:** Real-time-ish polling for alerts; dedicated views for threads and private messages.
- **Profile Viewing:** Display user bios, followers, and post history in side panels.
- **Media Handling:** Text previews for images/videos (e.g., alt-text or links); optional external open for full view.

### Advanced Features (Post-MVP)
- **Offline Mode:** Cache feeds for reading without internet.
- **Analytics:** Basic stats on engagement (e.g., view counts).
- **Integration:** Hooks for scripting (e.g., pipe posts to grep) or tools like tmux splits.
- **Security:** Rate-limit handling, token refresh, and proxy support.

## Tech Stack
Built entirely in Go for performance, concurrency, and simplicity—leveraging its standard library and a focused set of third-party packages. No JavaScript or external runtimes required; compiles to a single, portable binary.

### Core Language & Runtime
- **Go:** Version 1.22+ (for native concurrency with goroutines and channels, ensuring non-blocking I/O and scalability).

### TUI Framework
- **Bubble Tea (github.com/charmbracelet/bubbletea):** The primary framework for building the interactive TUI. Uses an Elm-inspired architecture (Model-Update-View) for managing state and rendering. Handles event loops, keyboard/mouse input, and full-screen rendering efficiently.
  - Why? Proven for complex TUIs; supports smooth updates without redraw overhead.

### UI Components & Styling
- **Bubbles (github.com/charmbracelet/bubbles):** Companion library providing reusable widgets like viewport (for scrolling feeds), list (for search results/notifications), textarea (for post composition), spinner (loading indicators), and paginator.
- **Lip Gloss (github.com/charmbracelet/lipgloss):** CSS-like styling engine for terminal output. Enables premium aesthetics: borders, colors, alignments, gradients, and layouts for post cards, headers, and modals.
  - Achieves "OpenCode-style" polish: Clean, colorful, responsive designs that feel modern in any terminal emulator.

### Networking & API Integration
- **Go Standard Library (net/http, encoding/json):** For API requests to X's endpoints (e.g., timelines, posting). Goroutines handle concurrent fetches.
- **Gorilla WebSocket (github.com/gorilla/websocket):** For real-time updates if using push notifications or a custom backend.
- **OAuth Library (golang.org/x/oauth2):** Secure authentication flows.

### Data Handling & Storage
- **BoltDB or Badger (github.com/dgraph-io/badger):** Lightweight embedded DB for caching posts, configs, and tokens locally.
- **YAML/JSON Parsing (gopkg.in/yaml.v3):** For user config files (e.g., themes, keybindings).

### Testing & Utilities
- **Go Testing Stdlib:** Unit/integration tests for models and API mocks.
- **Logrus or Zerolog:** Structured logging for debugging.
- **Cobra or Urfave/cli:** CLI flags for launch options (e.g., `niotebook --config path`).

### Build & Deployment
- **Go Build Tools:** Cross-compile for Linux, macOS, Windows.
- **CI/CD:** GitHub Actions for testing and releases.
- **Dependencies Management:** Go Modules (go.mod).

This stack keeps the project lean (~5-10 dependencies total), with a focus on native Go strengths for "amazing concurrent support" (e.g., spawning goroutines for each API poll or WebSocket listener).

## Architecture Overview
- **MVC Pattern via Bubble Tea:** Models hold state (e.g., TimelineModel with []Post); Updates handle events/API responses; Views render styled output.
- **Concurrency Model:** Main event loop in foreground; background goroutines + channels for async tasks (e.g., fetch new posts every 30s).
- **Modular Design:** Separate packages for UI components, API clients, and storage to ease maintenance.
- **Error Handling:** Graceful fallbacks (e.g., retry on API failures) with user-friendly messages.

## Development Roadmap
1. **Week 1-2 (Prototype):** Set up Bubble Tea skeleton with basic viewport and mock feed.
2. **Week 3-4 (Core Features):** Implement auth, timeline fetching, and posting.
3. **Week 5-6 (Polish):** Add styling, scrolling, and concurrency tests.
4. **Week 7+ (MVP Release):** Integration testing, docs on niotebook.com, and GitHub launch.
5. **Ongoing:** Community feedback, feature additions.

This brief provides a solid foundation—feel free to iterate on it, Shaikh! If you need sample code snippets, wireframes, or help with the landing page content for www.niotebook.com, just say the word. 🚀
