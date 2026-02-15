---
title: "ADR-0005: MVP Scope — Timeline + Posts + Profiles Only"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, product, scope]
---

# ADR-0005: MVP Scope — Timeline + Posts + Profiles Only

## Status

Accepted

## Context

The initial vision for Niotebook includes a rich feature set: follows, likes, reposts, replies, threads, search, notifications, DMs, media handling, offline mode, plugins, and federation. Building all of these for launch would delay shipping significantly and introduce risk.

We need to identify the smallest feature set that constitutes a usable social platform.

## Decision

The MVP ships with **three core features only:**

1. **Home Timeline** — a global feed of all posts, sorted chronologically, with manual refresh
2. **Post Composition** — compose and publish text posts from the TUI
3. **User Profiles** — view bios and post history; edit own profile

Everything else is deferred to post-MVP releases.

### What This Means Concretely

**Included:**
- User registration and login (email + password)
- Global timeline (all posts from all users — no follow graph filtering)
- Text-only post creation with character limit
- Profile viewing and editing
- Keyboard-driven navigation (j/k scroll, n to compose, Enter to view profile)
- Manual refresh (press key to fetch new posts)

**Excluded from MVP:**
- Follow/unfollow and personalized timeline
- Likes, reposts, quote posts
- Replies and threaded conversations
- Search and hashtags
- Notifications
- Direct messages
- Media attachments (images, links with previews)
- Real-time updates
- Everything in "Advanced Features" from the initial notes

### Why a Global Timeline?

Without follow/unfollow, there's no personalized feed. The MVP shows **all posts from all users** in reverse chronological order. This has a side benefit: early adopters see all activity, which makes the platform feel alive even with few users. Personalized feeds come with the follow system in v1.1.

## Consequences

### Positive

- Drastically reduced scope — faster path to a working, shippable product
- Forces focus on core quality: the posting and reading experience must be excellent
- Simpler backend: fewer endpoints, simpler data model, fewer edge cases
- Global timeline helps with cold start — every post is visible to every user

### Negative

- Platform may feel "too simple" compared to established social media
- No engagement features (likes, replies) means less stickiness — users can post and read but can't interact
- Global timeline won't scale past a few hundred active users without becoming noisy

### Neutral

- The MVP is a foundation. Each deferred feature is a clear, well-scoped increment to add in subsequent releases.
