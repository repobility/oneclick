# ADR 0002 — Capability URL as the authorization model

- **Status**: Accepted
- **Date**: 2026-05-12

## Context

A one-time link service needs *some* way to scope who can resolve the
destination. Options:

- **Accounts + ACL** — heavy: needs a user database, password reset,
  session tokens.
- **Pre-shared passphrase** — pushes the coordination problem onto the
  user.
- **Capability URL** — the URL itself is the credential.

## Decision

OneClick uses a capability URL of the form:

```
https://host/l/<id>#k=<urlsafe-base64-key>
```

- `id`: 16 random bytes via `crypto/rand` → 22-char urlsafe base64
  → ~128-bit search space. Path-visible DB key.
- `k`: 32-byte AES-GCM key from `crypto.getRandomValues`. In the URL
  fragment.

The browser **never** sends fragments in HTTP. The server cannot learn
the key, even from logs. Possession of the *full* URL is the credential.

## Rationale

- Removes the need for accounts.
- Aligns with prior art: PrivateBin, Yopass, OneTimeSecret, Firefox
  Send, wormhole.app — all use capability URLs.
- The threat surface shrinks: there's no auth code to have a bug. The
  server treats every fetcher equally; only the URL distinguishes who
  can decrypt.

## Consequences

- A leaked URL is a leaked destination. Mitigated by one-time-click
  default; once consumed, the URL is worthless.
- The creator cannot retract after sending. TTL is enforced
  server-side (default 24 h).
- Static analyzers (Repobility's AUC003 / AUC008) flag every
  object-level endpoint as "missing auth". We document the model
  explicitly in `.repobility/access.yml` and this ADR.
