---
title: Niotebook Vault
created: 2026-02-15
updated: 2026-02-16
tags: [home, index]
---

# Niotebook Documentation Vault

Welcome to the Niotebook project documentation. This vault contains all product requirements, architecture decisions, and design documents for the Niotebook TUI social media platform.

## Navigation

### Product
- **[[01-product/PRD|Product Requirements Document]]** — what we're building and why, user stories, non-functional requirements

### Engineering
- **[[02-engineering/adr/index|Architecture Decision Records]]** — 23 ADRs covering every technical decision
- **[[02-engineering/architecture/system-overview|System Architecture]]** — high-level system design, component diagram, tech stack
- **[[02-engineering/architecture/database-schema|Database Schema]]** — tables, constraints, indexes, migrations, key queries
- **[[02-engineering/architecture/bubble-tea-model-hierarchy|Bubble Tea Model Hierarchy]]** — TUI model structure, message types, async patterns, state management
- **[[02-engineering/architecture/server-internals|Server Internals]]** — middleware, logging, graceful shutdown, health endpoint, background jobs
- **[[02-engineering/api/api-specification|API Specification]]** — full REST API with exact request/response JSON formats
- **[[02-engineering/api/jwt-implementation|JWT Implementation]]** — token lifecycle, claims, refresh flow, security measures
- **[[02-engineering/architecture/build-and-dev-workflow|Build & Dev Workflow]]** — Makefile, local setup, CLI flags, deployment, git conventions
- **[[02-engineering/testing/testing-strategy|Testing Strategy]]** — unit, integration, TUI tests, CI pipeline, coverage targets

### Plans
- **[[04-plans/index|Implementation Plans Index]]** — phased implementation plans
- **[[04-plans/2026-02-16-mvp-implementation|MVP Implementation Plan]]** — 24 tasks across 7 phases, TDD throughout

### Design
- **[[03-design/index|Design Documents Index]]** — all design docs
- **[[03-design/tui-layout-and-navigation|TUI Layout & Navigation]]** — screen structure, views, transitions
- **[[03-design/keybindings|Key Bindings]]** — full keymap for every view
- **[[03-design/post-card-component|Post Card Component]]** — visual design, anatomy, empty states
- **[[03-design/auth-flow-ux|Auth Flow UX]]** — login/register screens, token refresh

## Project Summary

**Niotebook** is a standalone social media platform built as a TUI (Text User Interface) application. It replicates the core X/Twitter experience for developers and power users who live in terminal environments — coding with agentic tools like Claude Code, Codex, OpenCode, Cursor, etc.

It is **not** an X client. It is its own platform with its own backend, user accounts, and social graph.

## Key Decisions Summary

| Decision | Choice | ADR |
|----------|--------|-----|
| Platform type | Standalone social media (not X client) | [[02-engineering/adr/ADR-0001-standalone-platform\|0001]] |
| Backend | Go monolith | [[02-engineering/adr/ADR-0002-go-monolith\|0002]] |
| Database | PostgreSQL | [[02-engineering/adr/ADR-0003-postgresql\|0003]] |
| Auth | Email + password, JWT tokens | [[02-engineering/adr/ADR-0004-email-password-auth\|0004]] |
| MVP scope | Timeline + posts + profiles | [[02-engineering/adr/ADR-0005-mvp-scope\|0005]] |
| Feed updates | Manual refresh | [[02-engineering/adr/ADR-0006-manual-refresh\|0006]] |
| Repo structure | Monorepo, shared types | [[02-engineering/adr/ADR-0007-monorepo\|0007]] |
| API style | REST/JSON, API-first | [[02-engineering/adr/ADR-0008-api-first-rest\|0008]] |
| Deployment | VPS (DigitalOcean/Hetzner) | [[02-engineering/adr/ADR-0009-vps-deployment\|0009]] |
| Post limit | 140 characters | [[02-engineering/adr/ADR-0010-post-character-limit\|0010]] |
| Usernames | 3-15 chars, alphanumeric + underscores | [[02-engineering/adr/ADR-0011-username-rules\|0011]] |
| Pagination | Cursor-based (timestamp) | [[02-engineering/adr/ADR-0012-cursor-pagination\|0012]] |
| Error UX | Inline status bar messages | [[02-engineering/adr/ADR-0013-error-handling-ux\|0013]] |
| Rate limiting | Per-IP token bucket | [[02-engineering/adr/ADR-0014-rate-limiting\|0014]] |
| Theme | Single dark theme | [[02-engineering/adr/ADR-0015-dark-theme-only\|0015]] |
| Config path | ~/.config/niotebook/ (XDG) | [[02-engineering/adr/ADR-0016-config-xdg\|0016]] |
| Moderation | Deferred (invite-only launch) | [[02-engineering/adr/ADR-0017-no-moderation-mvp\|0017]] |
| TUI layout | Header + content + status bar | [[02-engineering/adr/ADR-0018-tui-layout\|0018]] |
| Post cards | Compact (username + time + content) | [[02-engineering/adr/ADR-0019-compact-post-cards\|0019]] |
| Compose | Inline modal overlay | [[02-engineering/adr/ADR-0020-compose-inline-modal\|0020]] |
| Testing | Comprehensive (unit + integration + TUI) | [[02-engineering/adr/ADR-0021-comprehensive-testing\|0021]] |
| Multi-line posts | Newlines allowed, count toward 140 limit | [[02-engineering/adr/ADR-0022-multiline-posts\|0022]] |
| Health endpoint | GET /health for monitoring | [[02-engineering/adr/ADR-0023-health-endpoint\|0023]] |

## Current Status

**Phase:** Implementation Planning — COMPLETE
**Next:** Implementation (see [[04-plans/2026-02-16-mvp-implementation|MVP Implementation Plan]])
**Target MVP:** Q2 2026
