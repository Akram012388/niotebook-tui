---
title: Design Documents
created: 2026-02-15
updated: 2026-02-15
tags: [design, index]
---

# Design Documents

Detailed component and feature design documents for Niotebook.

## Documents

| Document | Status | Description |
|----------|--------|-------------|
| [[tui-layout-and-navigation\|TUI Layout & Navigation]] | Accepted | Screen structure, views, transitions, scroll, resize |
| [[keybindings\|Key Bindings]] | Accepted | Full keymap for every view, reserved future keys |
| [[post-card-component\|Post Card Component]] | Accepted | Visual design, anatomy, relative time, empty states |
| [[auth-flow-ux\|Auth Flow UX]] | Accepted | Login/register screens, token refresh, startup flow |

## Related Engineering Docs

| Document | Location | Description |
|----------|----------|-------------|
| API Specification | [[02-engineering/api/api-specification\|API Spec]] | Full REST API with request/response formats |
| JWT Implementation | [[02-engineering/api/jwt-implementation\|JWT Details]] | Token lifecycle, claims, refresh flow, security |
| Database Schema | [[02-engineering/architecture/database-schema\|DB Schema]] | Tables, constraints, indexes, migrations, key queries |
| Bubble Tea Models | [[02-engineering/architecture/bubble-tea-model-hierarchy\|Model Hierarchy]] | TUI model structure, message types, async patterns |
| Server Internals | [[02-engineering/architecture/server-internals\|Internals]] | Middleware, logging, shutdown, health, background jobs |
| Build & Dev Workflow | [[02-engineering/architecture/build-and-dev-workflow\|Build/Dev]] | Makefile, dev setup, CLI flags, deployment, git conventions |
| Testing Strategy | [[02-engineering/testing/testing-strategy\|Testing]] | Unit, integration, TUI tests, CI pipeline, coverage targets |

## Design Principles

These principles guide all design decisions:

1. **Keyboard-first.** Every action must be achievable without a mouse. Mouse support is optional enhancement, never required.
2. **Information density.** Terminal real estate is limited. Show maximum useful information per screen. No decorative whitespace.
3. **Responsive to terminal size.** Handle resize events gracefully. Degrade content (truncate, hide secondary info) rather than break layout.
4. **Familiar to terminal users.** Use conventions from Vim, less, man pages. j/k for scroll, / for search, q to quit, ? for help.
5. **Fast.** No action should feel sluggish. Optimistic UI where possible (show the post immediately, confirm with server in background).
