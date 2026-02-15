---
title: "ADR-0012: Cursor-Based Pagination"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, api, performance]
---

# ADR-0012: Cursor-Based Pagination

## Status

Accepted

## Context

The timeline API must return posts in pages. Two standard approaches:

1. **Offset-based:** `GET /timeline?offset=50&limit=50` — "skip 50, give me 50"
2. **Cursor-based:** `GET /timeline?cursor=<post_id>&limit=50` — "give me 50 posts older than this one"

## Decision

Use **cursor-based pagination** with the post's `created_at` timestamp as the cursor.

### API Format

**Request:**
```
GET /api/v1/timeline?cursor=2026-02-15T10:30:00Z&limit=50
```

If no cursor is provided, return the newest posts.

**Response:**
```json
{
  "posts": [...],
  "next_cursor": "2026-02-15T09:15:00Z",
  "has_more": true
}
```

### Why Not Offset

Offset pagination breaks when new posts are inserted. If a user is viewing page 2 (posts 51-100) and a new post is created, what was post 51 becomes post 52. Refreshing page 2 duplicates the last post from page 1 or skips one. This is especially problematic for a social feed where new posts arrive constantly.

Cursor pagination is stable: "posts older than X" always returns the same set regardless of new inserts.

## Consequences

### Positive

- Stable pagination — no duplicated or skipped posts when new content arrives
- Efficient database query: `WHERE created_at < $cursor ORDER BY created_at DESC LIMIT $limit` uses the `idx_posts_created_at` index directly
- Stateless — no server-side session tracking for pagination state

### Negative

- Cannot "jump to page N" (no random access) — must scroll sequentially
- Slightly more complex client logic (must track cursor from last response)

### Neutral

- Cursor-based pagination is the standard for social media feeds (X, Instagram, Mastodon all use it).
