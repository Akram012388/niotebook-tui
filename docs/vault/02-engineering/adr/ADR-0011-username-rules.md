---
title: "ADR-0011: Username Validation Rules"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, auth, validation]
---

# ADR-0011: Username Validation Rules

## Status

Accepted

## Context

Usernames are the primary identity on Niotebook (displayed as `@username`). Rules must balance expressiveness with safety (no impersonation, no confusing characters).

## Decision

Username constraints:

- **Length:** 3-15 characters
- **Allowed characters:** alphanumeric (`a-z`, `A-Z`, `0-9`) and underscores (`_`)
- **Case-insensitive uniqueness:** `@Akram` and `@akram` are the same user. Stored lowercase in the database, displayed as entered during registration.
- **No leading/trailing underscores:** `_akram_` is rejected
- **No consecutive underscores:** `akram__dev` is rejected
- **Reserved usernames:** `admin`, `root`, `system`, `niotebook`, `api`, `help`, `support`, `me`, `about`, `settings`, `login`, `register`, `auth`, `posts`, `users`, `timeline`, `search`, `explore` (prevents confusion with system routes, API paths, and UI elements). Authoritative list maintained in `internal/server/service/` at the application level.

### Regex

```
^[a-zA-Z0-9](?:[a-zA-Z0-9_]*[a-zA-Z0-9])?$
```

With additional checks for length (3-15) and reserved names.

## Consequences

### Positive

- Clean, URL-safe handles (`/users/akram` works without encoding)
- No confusing lookalike characters (no dots, hyphens, special Unicode)
- Case-insensitive prevents impersonation (`@Akram` vs `@akram`)

### Negative

- Some users may want hyphens or dots in their username
- 15-char limit may be too short for some preferred handles

### Neutral

- Username changes are not supported in MVP. Users pick once at registration.
