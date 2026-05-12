# OneClick

End-to-end encrypted one-time URL redirect. Type a URL, get a short link; the recipient opens it once and is redirected — the server never sees the destination.

- **Backend** — Go 1.22+ · `chi` router · `html/template` · pure-Go SQLite (`modernc.org/sqlite`).
- **Frontend** — Templated HTML + vanilla JS for the crypto path.
- **Crypto** — Web Crypto API: AES-GCM with a 256-bit key and 96-bit IV. Key generated in the browser and shipped in the URL fragment (`#k=…`); browsers never send fragments in HTTP, so the server cannot decrypt.
- **Trust model** — if the server operator is malicious, they learn that a link existed and how often it was clicked, but cannot recover the destination URL.

## Quick start

```bash
git clone https://github.com/repobility/oneclick.git
cd oneclick
go run .
```

Open <http://127.0.0.1:8080>, paste a URL, copy the resulting one-time link.

## API

| Method | Path | Body | Returns |
|---|---|---|---|
| `POST` | `/api/links` | `{ ciphertext, iv, ttl_seconds, max_clicks }` | `{ id, expires_at, clicks_remaining }` |
| `GET`  | `/api/links/{id}` | — | `{ ciphertext, iv, expires_at, clicks_remaining }` and atomically decrements the click counter |
| `GET`  | `/api/links/{id}/meta` | — | non-consuming peek at `expires_at` + `clicks_remaining` |
| `GET`  | `/l/{id}` | — | viewer page (key parsed from URL fragment client-side) |
| `GET`  | `/healthz` | — | liveness probe |

## License

MIT.
