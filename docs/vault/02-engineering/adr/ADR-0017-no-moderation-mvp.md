---
title: "ADR-0017: No Content Moderation in MVP"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, product, moderation]
---

# ADR-0017: No Content Moderation in MVP

## Status

Accepted

## Context

Any social platform eventually needs content moderation. The question is whether to build moderation tooling for the MVP launch.

## Decision

**Defer all moderation tooling to post-MVP.** The MVP ships with no automated content filtering, no reporting system, and no admin tools for content management.

### Mitigation Strategy

- MVP launches to a small, invite-only community (personal network, developer friends)
- Admin (project owner) can moderate directly via SQL if absolutely necessary: `DELETE FROM posts WHERE id = '...'`
- Community norms are established early through personal relationships, not tooling

### When to Revisit

Moderation becomes a priority when:
- The platform opens to public registration
- User count exceeds what one person can informally moderate (~100+ active users)
- Any instance of harassment or spam occurs

## Consequences

### Positive

- Reduces MVP scope — no moderation UI, no reporting API, no filter logic
- Forces focus on core features
- Small community self-polices effectively

### Negative

- No formal process for handling problematic content
- Admin moderation via raw SQL is error-prone and not auditable
- If the platform grows unexpectedly, there's no safety net

### Neutral

- This is a standard approach for MVP social platforms. Moderation tooling is built when the community requires it, not before.
