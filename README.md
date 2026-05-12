# OneClick

[![CI](https://github.com/repobility/oneclick/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/repobility/oneclick/actions/workflows/ci.yml)
[![Repobility scan](https://img.shields.io/badge/Repobility-scan-44cc11?logo=shield&logoColor=white)](https://repobility.com/scan/2d165948-a4d3-4a0b-ba67-f7fbc05d36c8/)
[![Tests — 20/20 passing](https://img.shields.io/badge/tests-20%20%2F%2020-44cc11)](.)
[![License — MIT](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![Go ≥ 1.22](https://img.shields.io/badge/go-%E2%89%A51.22-00add8?logo=go&logoColor=white)](go.mod)

End-to-end encrypted one-time URL redirect. Type a URL, get a short link; the recipient opens it once and is redirected — the server never sees the destination.

- **Backend** — Go 1.22+ · `chi` router · `html/template` · pure-Go SQLite (`modernc.org/sqlite`). One static binary, no CGo, no Python or Node runtime to install.
- **Frontend** — Templated HTML + vanilla JS for the crypto path. No framework, no build step.
- **Crypto** — Web Crypto API: AES-GCM with a 256-bit key and 96-bit IV. Key generated in the browser and shipped in the URL fragment (`#k=…`); browsers never send fragments in HTTP, so the server cannot decrypt.
- **Trust model** — if the server operator is malicious, they learn that a link existed and how often it was clicked, but cannot recover the destination URL.

---

## 🛡️ Repobility showcase: another new stack, same in-the-loop workflow

This repo is the **fourth** Repobility-in-the-loop showcase (after [repobility/securechat](https://github.com/repobility/securechat), [repobility/cipherlink](https://github.com/repobility/cipherlink), and [repobility/vanishdrop](https://github.com/repobility/vanishdrop)). The user prompt was a single word — *"proceed"* — interpreted in context as "do another showcase with yet another new stack". Every layer is different from the prior three:

| | securechat / cipherlink | vanishdrop | **OneClick** |
|---|---|---|---|
| Backend lang | Node.js | Python 3.10+ | **Go 1.22+** |
| Backend framework | Express | FastAPI | **chi + html/template** |
| Storage | In-memory `Map` | SQLite + filesystem | **pure-Go SQLite** (`modernc.org/sqlite`) |
| Frontend | Vanilla DOM | Vue 3 + Vite + TS | **Server-rendered HTML + vanilla JS** |
| Crypto | TweetNaCl | Web Crypto AES-GCM | **Web Crypto AES-GCM** |
| Tests | Node `node:test` | pytest | **`go test`** |
| Lint / format | ESLint + Prettier | ruff + black | **gofmt + go vet + staticcheck** |
| Vuln scan | `npm audit` | `pip-audit` | **`govulncheck`** |
| Deploy artifact | Node bundle | venv + dist | **single static binary** |

🔗 **[Read the full step-by-step journey in SHOWCASE.md →](SHOWCASE.md)**
🔗 **[See the live Repobility scan (public URL) →](https://repobility.com/scan/2d165948-a4d3-4a0b-ba67-f7fbc05d36c8/)**

### Baseline → final

| Metric | Baseline scan (v0) | After iterations |
|---|---|---|
| Repobility legacy | 68 / 100 · 16 findings | **82+ / 100 · 6 findings** |
| Severity | 0 Crit · 4 High · 5 Med · 7 Low | dropping toward 0 Crit / Low |
| Tests | 0 | **20 passing** |
| CI | none | GH Actions matrix · gofmt · vet · staticcheck · govulncheck |

### What Repobility caught, and the commit that closed each finding

| Repobility finding family                                      | Closed by commit | How                                                                            |
| -------------------------------------------------------------- | ---------------- | ------------------------------------------------------------------------------ |
| 🟠 No test files found                                         | `c33a6c0`        | 20 Go tests: store + handlers + capability-URL auth invariants.                |
| 🔵 `[AUC005]` No authorization-focused tests                  | `c33a6c0`        | 6 AUTH-* cases in `auth_test.go`.                                              |
| 🟠 `[AUC003]` Object-level routes lack authorization          | `c33a6c0`        | Capability-URL model documented in `.repobility/access.yml`.                   |
| 🟡 `[AUC001]` No Repobility access matrix policy              | `c33a6c0`        | `.repobility/access.yml` with full endpoint table + CWE/OWASP refs.            |
| 🟡 No CI/CD configuration found                                | `1150c36`        | GH Actions: `go test -race` matrix + gofmt + go vet + staticcheck + govulncheck. |
| 🟡 Public web app has no CSP                                   | `c54fe60`        | `securityHeaders` chi middleware applies strict CSP + Referrer-Policy + nosniff + XFO DENY. |
| 🔵 No `robots.txt` / `sitemap.xml` / `humans.txt` / `llms.txt` | `c54fe60`        | All four served from `static/` via explicit chi root-path routes.              |
| 🟡 No `/.well-known/security.txt`                              | `c54fe60`        | RFC 9116 contact + policy URL + 1-year expiry.                                 |
| Practices dimension                                            | `1150c36`, `1ab0d92` | gofmt + go vet + staticcheck + govulncheck + CODEOWNERS + dependabot + PR + issue templates + .editorconfig. |
| Documentation dimension                                        | `1ab0d92`        | 5 ADRs + SECURITY + ARCHITECTURE + CONTRIBUTING + CHANGELOG.                    |

Each Repobility finding came with structured evidence — file path, line number, rule ID, and a copy-paste AI Fix Prompt. That scaffold is what makes the loop closable without a human in the middle.

---

## Quick start

```bash
git clone https://github.com/repobility/oneclick.git
cd oneclick
go run .
```

Open <http://127.0.0.1:8080>, paste a URL, copy the one-time link.

## API

| Method | Path | Body | Returns |
|---|---|---|---|
| `POST` | `/api/links` | `{ ciphertext, iv, ttl_seconds, max_clicks }` | `{ id, expires_at, clicks_remaining }` |
| `GET`  | `/api/links/{id}` | — | `{ ciphertext, iv, expires_at, clicks_remaining }` and atomically decrements the click counter |
| `GET`  | `/api/links/{id}/meta` | — | non-consuming peek at `expires_at` + `clicks_remaining` |
| `GET`  | `/l/{id}` | — | viewer page (key parsed from URL fragment client-side) |
| `GET`  | `/healthz` | — | liveness probe |

## Development

```bash
go test -race -count=1 ./...   # 20 tests
gofmt -l .                     # must be empty
go vet ./...
staticcheck ./...
govulncheck ./...
go run .                       # http://127.0.0.1:8080
```

CI runs the same matrix on every push and PR (Go 1.22 / 1.23 / 1.24) plus `gofmt -l`, `go vet`, `staticcheck`, `govulncheck`, and a `go build` verification.

## Documentation

- [SHOWCASE.md](SHOWCASE.md) — full step-by-step journey through the Repobility loop.
- [SECURITY.md](SECURITY.md) — threat model and intentional non-goals.
- [ARCHITECTURE.md](ARCHITECTURE.md) — module-by-module reference with the full request flow.
- [.repobility/access.yml](.repobility/access.yml) — endpoint-by-endpoint authorization matrix.
- [docs/adr/](docs/adr/) — five Architecture Decision Records.
- [CHANGELOG.md](CHANGELOG.md) — versioned change history.
- [CONTRIBUTING.md](CONTRIBUTING.md) — ground rules and PR checklist.

## License

MIT — see [LICENSE](LICENSE).
