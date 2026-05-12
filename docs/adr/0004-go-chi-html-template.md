# ADR 0004 — Go + chi + html/template stack

- **Status**: Accepted
- **Date**: 2026-05-12

## Context

OneClick is the fourth Repobility-in-the-loop showcase. The prior
three covered Node + vanilla DOM (securechat, cipherlink) and
Python + Vue (vanishdrop). The user prompt was to use a different
stack — so we pick Go.

## Decision

- **Backend**: Go 1.22+ with `go-chi/chi` for routing.
- **Templates**: stdlib `html/template`.
- **Storage**: pure-Go SQLite via `modernc.org/sqlite` — no CGo, no
  external libsqlite. Single-binary deploy.
- **Frontend**: server-rendered HTML + vanilla JS for the crypto path.
  No build step, no JS framework.
- **Crypto**: Web Crypto API (AES-GCM).

## Rationale

- **Go** compiles to a single static binary — simplest possible
  deployment story.
- **chi** has middleware semantics that map cleanly to per-route
  authorization decisions, even though we don't need much auth here.
- **html/template** auto-escapes contextually, so the templates are
  safe by default.
- **modernc.org/sqlite** removes the CGo dependency that would
  otherwise complicate cross-compilation and CI.
- **No JS framework** is appropriate for a small app like this; the
  whole client is < 200 LOC of plain JS plus the Web Crypto path.

## Considered alternatives

- **Templ** (type-safe Go templates) — would have been nice, but adds
  a code-generation step. `html/template` is good enough.
- **HTMX** — considered, but the crypto flow has to happen in JS
  regardless, so HTMX would only have helped with the form submit. Not
  worth the extra dependency.
- **Echo / Gin** — fine alternatives to chi. chi was picked for its
  stdlib-like API.

## Consequences

- Repobility's scanner exercises Go-specific rules (gofmt, vet,
  staticcheck, govulncheck) it didn't on prior repos.
- CI matrix is `go test -race` across Go 1.22/1.23/1.24.
- Deployment is `./oneclick` — one binary, no runtime dependencies
  beyond the SQLite file it creates next to itself.
