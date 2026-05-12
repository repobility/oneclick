# ADR 0005 — No click receipts or analytics

- **Status**: Accepted
- **Date**: 2026-05-12

## Context

URL shorteners commonly offer click analytics — the creator can see
that their link was clicked, when, from what IP/UA.

## Decision

OneClick does not offer click receipts. The creator cannot learn —
from the service — whether or when their link was opened.

## Rationale

- The recipient's privacy outweighs the sender's curiosity. Many
  legitimate recipients open one-time links in moments where being
  identified is harmful (incident response, journalism, abuse-victim
  support).
- The sender can already infer the outcome out-of-band: the recipient
  acks via the channel they got the URL on.
- Adding click receipts would require either (a) an account for the
  sender (kills ADR-0002) or (b) a long-lived sender token attached
  to each upload, which becomes a new credential to protect.

## Consequences

- `/healthz` reports only a count of stored records, no per-record
  state.
- The compose UI shows the URL, expiry, and max-clicks, never an
  "opened at" stamp.
- Operators should keep server logs minimal — even per-fetch IP logs
  would effectively become click receipts for anyone who can read
  them. The default chi `RealIP` middleware extracts the client IP
  for `RemoteAddr` but does not log it.
