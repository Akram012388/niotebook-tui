---
title: "ADR-0003: PostgreSQL as Primary Database"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, database, infrastructure]
---

# ADR-0003: PostgreSQL as Primary Database

## Status

Accepted

## Context

Niotebook needs persistent storage for users, posts, profiles, and eventually social graph relationships (follows, likes, reposts). The database choice affects query patterns, scaling characteristics, and operational complexity.

Candidates considered:

1. **PostgreSQL** — relational, mature, excellent Go support via `pgx`
2. **SQLite** — embedded, zero-dependency, good for single-server
3. **MongoDB** — document store, flexible schema

## Decision

Use **PostgreSQL** as the primary and only database.

### Why PostgreSQL Over Alternatives

**Social media data is inherently relational.** Users follow users. Users create posts. Users like posts. Users reply to posts. These are relationships between entities — exactly what relational databases model natively. A query like "posts from people I follow, ordered by time" is a straightforward SQL join with an index.

**Proven at social media scale.** Instagram runs on PostgreSQL. Twitter started on MySQL (relational). Mastodon uses PostgreSQL. The pattern is well-understood.

**Full-text search built in.** PostgreSQL's `tsvector`/`tsquery` provides good-enough search for MVP without adding Elasticsearch.

**Excellent Go driver.** `pgx` is one of the best database drivers in any language — high performance, low allocation, native PostgreSQL protocol support.

### Why Not SQLite

SQLite uses a single-writer lock. Concurrent social media operations (multiple users posting, reading, refreshing feeds simultaneously) would serialize at the write lock, creating a bottleneck. Fine for a personal tool, problematic for a multi-user platform.

### Why Not MongoDB

Social graph queries (followers, feed generation) become unnecessarily complex with document stores. Joins don't exist natively. The flexible schema advantage is minimal when your domain model is well-understood upfront.

## Consequences

### Positive

- Natural modeling of social relationships with foreign keys and joins
- Battle-tested scaling path (connection pooling, read replicas, partitioning)
- Built-in full-text search for MVP
- Strong data integrity via constraints and transactions
- Hand-written SQL via `pgx` — no ORM overhead, full control over queries

### Negative

- Requires running a PostgreSQL instance alongside the server (operational overhead vs. embedded SQLite)
- Schema migrations needed for model changes (mitigated by `golang-migrate`)
- Must manage connection pools and handle connection failures

### Neutral

- PostgreSQL is the default database for new Go projects. Extensive documentation, community support, and tooling available.
