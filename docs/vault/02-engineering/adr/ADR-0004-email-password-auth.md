---
title: "ADR-0004: Email + Password Authentication for MVP"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, auth, security]
---

# ADR-0004: Email + Password Authentication for MVP

## Status

Accepted

## Context

Users need to authenticate to Niotebook. Since the primary interface is a TUI, the auth flow must work without browser redirects or GUI elements. Options considered:

1. **Email + password** — classic registration and login via the TUI
2. **Magic link / device code** — email-based passwordless auth
3. **CLI token + web signup** — register on niotebook.com, paste API token into TUI
4. **SSH key-based** — authenticate using existing SSH keys (like GitHub CLI's `gh auth`)

## Decision

Use **email + password authentication** with JWT tokens for the MVP.

### Flow

1. **Registration:** User provides username, email, password in TUI. Server validates, hashes password (bcrypt, cost 12+), creates account, returns JWT pair.
2. **Login:** User provides email + password. Server verifies, returns JWT pair (access + refresh).
3. **Session persistence:** TUI stores JWT in `~/.config/niotebook/config.yaml`. On startup, TUI uses stored access token. If expired, uses refresh token to get new access token. If refresh expired, prompts for login.
4. **Token lifecycle:** Access token: 24h. Refresh token: 7 days.

### Why Not Other Options

**Magic link:** Requires an email delivery service (SendGrid, SES, Postmark). Adds an external dependency, operational cost, and failure mode (email deliverability). Overkill for MVP.

**CLI token + web signup:** Requires building niotebook.com as a web app with registration UI before the TUI is useful. Adds a separate frontend project to MVP scope.

**SSH key auth:** Great fit for the target audience but complex to implement correctly (key registration, challenge-response, signature verification). Planned as a v1.1 addition — very on-brand for a "nerds' platform."

## Consequences

### Positive

- Users can register and log in entirely within the TUI — zero external dependencies
- Familiar pattern, no learning curve
- Self-contained: no email service, no web UI, no external auth provider
- JWT tokens enable stateless API authentication — server doesn't need session storage

### Negative

- Passwords are a security liability (users reuse passwords, password storage is a target)
- Must implement password reset flow eventually (which will require email delivery anyway)
- Less "nerd-cred" than SSH key auth

### Neutral

- SSH key auth is planned for v1.1 as an additional auth method, not a replacement
- bcrypt with cost 12+ provides strong hashing. Rate limiting on login endpoint prevents brute force.
