# ADR 0003 — AES-GCM via the Web Crypto API as the AEAD

- **Status**: Accepted
- **Date**: 2026-05-12

## Context

OneClick needs a symmetric AEAD for short ciphertexts (URLs, capped at
2000 chars plaintext, plus a 16-byte GCM tag).

## Decision

We use **AES-GCM via the Web Crypto API**:

- 256-bit key from `crypto.getRandomValues(new Uint8Array(32))`.
- 96-bit IV from `crypto.getRandomValues(new Uint8Array(12))`.
- `crypto.subtle.encrypt({name:'AES-GCM',iv}, cryptoKey, bytes)`.

Same primitive VanishDrop uses. The Web Crypto API is universal across
modern browsers and hardware-accelerated via AES-NI.

## Considered alternatives

- **TweetNaCl secretbox** (XSalsa20-Poly1305) — fine, but the prior
  Node-based showcases (securechat, cipherlink) already used it. Web
  Crypto AES-GCM is what differentiates the more recent showcases.
- **AES-CBC + HMAC** — MAC-then-encrypt foot-gun; AES-GCM is the
  modern default.

## Consequences

- Confidentiality + integrity against any adversary without the key.
- Zero crypto-related third-party origins in the browser bundle.
- Web Crypto requires HTTPS (or `localhost`) — fine for any realistic
  deployment.
