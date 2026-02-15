---
title: "ADR-0001: Standalone Platform, Not an X Client"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, product, architecture]
---

# ADR-0001: Standalone Platform, Not an X Client

## Status

Accepted

## Context

The initial concept for Niotebook described it as "X for the Terminal," which could be interpreted two ways:

1. **An X/Twitter API client** — a TUI that wraps X's API to browse and post to X from the terminal (like `rainbowstream` or `turses`).
2. **A standalone social platform** — an independent social network with its own backend, users, and data, inspired by X's UX patterns but living entirely in the TUI.

The target users are developers using agentic coding tools (Claude Code, Codex, OpenCode, Cursor) who want social media without leaving the terminal. They don't necessarily want X's specific content — they want the X-like *experience* in a terminal-native context.

## Decision

Niotebook is a **standalone social media platform** with its own backend server, user accounts, social graph, and content. It draws UX inspiration from X/Twitter but has no dependency on, integration with, or reference to X's APIs, branding, or content.

This means:
- We build and operate our own backend (API server, database, auth)
- Users register Niotebook-specific accounts
- All content is native to the Niotebook platform
- The TUI client communicates exclusively with the Niotebook server

## Consequences

### Positive

- Full control over the platform: features, moderation, data, and roadmap
- No dependency on X's API (which has restrictive rate limits, expensive access tiers, and a history of breaking changes)
- Can build features X doesn't have (terminal-native workflows, pipe-friendly output, scripting hooks)
- Community ownership — users' content lives on a platform they can trust
- Open-source server means anyone can self-host

### Negative

- Requires building and operating a full backend (significant engineering scope)
- Cold start problem: no content until users join. The platform is only valuable with an active community
- Must handle all infrastructure: hosting, scaling, moderation, abuse prevention

### Neutral

- Federation (ActivityPub) can be added later to bridge with Mastodon and other platforms, mitigating the cold start problem
