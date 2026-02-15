---
title: "ADR-0009: VPS Deployment (DigitalOcean/Hetzner)"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, deployment, infrastructure]
---

# ADR-0009: VPS Deployment (DigitalOcean/Hetzner)

## Status

Accepted

## Context

The Niotebook server and PostgreSQL database need to run somewhere accessible to TUI clients over the internet. Options:

1. **VPS** (DigitalOcean, Hetzner, Vultr) — rent a Linux server, deploy manually or with simple scripts
2. **Managed platform** (Railway, Fly.io, Render) — deploy via Git push, managed Postgres included
3. **Cloud provider** (AWS, GCP) — full cloud infrastructure with IaC

## Decision

Deploy to a **VPS** (DigitalOcean or Hetzner) for the MVP.

### Deployment Model

- Single VPS running both the Go server binary and PostgreSQL
- Server binary managed by systemd (auto-restart on crash)
- PostgreSQL installed via OS package manager
- HTTPS via Let's Encrypt (certbot or Caddy as reverse proxy)
- Deployment: `scp` the binary + `systemctl restart` (or a simple Makefile target)

### Why VPS Over Managed Platforms

**Cost:** A Hetzner CX22 (2 vCPU, 4GB RAM) costs ~$4.50/month. Equivalent Railway/Render pricing for server + Postgres would be $10-25/month. At MVP scale with low traffic, this is pure cost savings.

**Control:** Full root access to configure PostgreSQL, set up backups, inspect logs, tune the OS. Managed platforms abstract this away, which is convenient but limiting when debugging production issues.

**Simplicity:** The Go binary has zero runtime dependencies. Deployment is literally: build binary, copy to server, restart service. No Docker, no Kubernetes, no buildpacks.

### Why Not AWS/GCP

Over-engineered for a single-server MVP. The operational overhead of VPCs, security groups, IAM roles, and Terraform state management is unjustified at this stage.

## Consequences

### Positive

- Cheapest hosting option ($4-10/month for server + database)
- Full control over the environment
- Simplest possible deployment (binary + systemd)
- No vendor lock-in to managed platform APIs or pricing models

### Negative

- Manual server maintenance (OS updates, PostgreSQL upgrades, disk monitoring)
- No automatic scaling — must manually upgrade VPS tier if traffic grows
- Backup strategy must be implemented manually (pg_dump cron job)
- SSL certificate management (mitigated by Caddy's auto-HTTPS)

### Neutral

- If traffic grows beyond a single server's capacity, migration to a managed platform or multi-server setup is straightforward. The Go binary and PostgreSQL are portable.
