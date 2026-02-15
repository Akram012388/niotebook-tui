---
title: "Post Card Component"
created: 2026-02-15
updated: 2026-02-15
status: accepted
tags: [design, tui, component]
---

# Post Card Component

The post card is the most-rendered component in the app. It must be compact, scannable, and clear.

## Layout

### Unselected Post

```
  @username · 2m
  Post content goes here, up to 140 characters. It wraps to the
  next line if it exceeds the terminal width minus the left padding.
────────────────────────────────────────────────────────────────────
```

### Selected Post (Cursor on This Post)

```
▸ @username · 2m
  Post content goes here, up to 140 characters. It wraps to the
  next line if it exceeds the terminal width minus the left padding.
────────────────────────────────────────────────────────────────────
```

The `▸` marker and the username are highlighted with the accent color (magenta) when selected. The rest of the post text may have a subtle background tint or remain unchanged depending on terminal capability.

## Anatomy

```
[marker] @[username] · [relative_time]
  [content, word-wrapped to terminal width - 2 (left padding)]
[separator line]
```

### Fields

| Field | Source | Styling | Notes |
|-------|--------|---------|-------|
| Marker | Selection state | Accent color, `▸` when selected, space when not | 1 character + 1 space padding |
| Username | `post.author.username` | Cyan, bold | Prefixed with `@` |
| Separator dot | Static | Dim gray | ` · ` (space, middle dot, space) |
| Relative time | `post.created_at` | Dim gray | See relative time rules below |
| Content | `post.content` | White/default | Word-wrapped |
| Separator | Static | Dark gray | `─` repeated to terminal width |

### Relative Time Rules

| Age | Display |
|-----|---------|
| < 1 minute | `now` |
| 1-59 minutes | `Xm` (e.g., `5m`) |
| 1-23 hours | `Xh` (e.g., `3h`) |
| 1-6 days | `Xd` (e.g., `2d`) |
| 7-29 days | `Xw` (e.g., `2w`) |
| 30+ days | `Jan 15` (month + day) |
| Different year | `Jan 15, 2025` (month + day + year) |

Time is calculated client-side based on the TUI machine's clock.

## Sizing & Wrapping

- **Left padding:** 2 characters (for marker + space)
- **Content area width:** `terminal_width - 2` (left padding)
- **Word wrapping:** Break at word boundaries. If a single word exceeds the content width (e.g., a very long URL), break mid-word.
- **Maximum post height:** No limit — a 140-char post at 80-column terminal is at most 2-3 lines. No truncation needed.
- **Separator:** Full terminal width, single `─` character, dark gray

## Empty States

### No Posts in Timeline

```
┌──────────────────────────────────────────────────────────────────────────┐
│  niotebook  @akram                                        Timeline  ↻   │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│                                                                          │
│                                                                          │
│                       No posts yet.                                      │
│                       Be the first to post!                              │
│                       Press n to compose.                                │
│                                                                          │
│                                                                          │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  n: new post  r: refresh  ?: help                                        │
└──────────────────────────────────────────────────────────────────────────┘
```

### Loading State (Initial Fetch / Refresh)

```
  ⠋ Loading timeline...
```

Spinner (Bubbles spinner component) centered in the content area during initial load. During refresh (`r`), the spinner appears in the status bar: `"Refreshing..."` and existing posts remain visible.

### Pagination Loading (Scrolled to Bottom)

```
  @last_user · 1h
  Last loaded post content here.
────────────────────────────────────────────────────────────────────
  ⠋ Loading more posts...
```

Spinner appears below the last post while fetching the next page.

## Unicode & Special Characters

- Posts may contain emoji and Unicode. The TUI renders them as-is — terminal emulator handles glyph rendering.
- Character counting for the 140 limit uses `len([]rune(content))` (Go rune count), not byte length. This means emoji and CJK characters count as 1 character each.
- If a terminal doesn't support certain Unicode glyphs, they'll appear as `?` or boxes — this is acceptable and expected for a terminal app.
