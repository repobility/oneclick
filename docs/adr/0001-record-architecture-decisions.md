# ADR 0001 — Record architecture decisions

- **Status**: Accepted
- **Date**: 2026-05-12

## Context

OneClick makes several non-obvious choices (Go for the backend,
capability-URL auth, no JS framework, pure-Go SQLite). Future
contributors and security reviewers need the *why* of each.

## Decision

Significant decisions live as ADRs under [`docs/adr/`](.), following
the [Cognitect convention](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions).
Each ADR has **Context**, **Decision**, **Consequences**, and is
immutable once accepted.

## Consequences

- A reviewer reading a PR that touches the wire protocol, the auth
  model, or the cryptographic core can find the relevant ADR.
- Out-of-date ADRs are explicitly visible (status: `Superseded`)
  instead of silently dropped.
