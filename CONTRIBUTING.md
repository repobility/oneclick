# Contributing to OneClick

OneClick is a small, auditable codebase. Changes are weighted toward
simplicity and clarity over feature breadth.

## Ground rules

1. **Don't break the trust model.** The server must remain unable to
   decrypt destination URLs. Any change introducing server-side
   decryption, plaintext logging, or breaking the capability-URL
   property is out of scope.
2. **Trust crypto, not your own.** Use the Web Crypto API or another
   audited primitive. Do not write custom cryptography.
3. **Keep the surface small.** No build step beyond `go build`. No JS
   framework. No third-party JS origins.
4. **Tests must accompany behaviour changes.** Every protocol- or
   auth-relevant change lands with new cases under `*_test.go`.

## Local development

```bash
git clone https://github.com/repobility/oneclick.git
cd oneclick

go test -race ./...
gofmt -l .              # must be empty
go vet ./...
go run .                # http://127.0.0.1:8080
```

CI runs the same checks on Go 1.22, 1.23, and 1.24 + staticcheck +
govulncheck.

## Project layout

| Path                       | Purpose                                                |
| -------------------------- | ------------------------------------------------------ |
| `main.go`                  | HTTP routes, middleware, server boot.                  |
| `store.go`                 | SQLite-backed store with atomic click-consume.         |
| `templates/`               | `html/template` files (compose, view, gone).           |
| `static/crypto.js`         | Web Crypto AES-GCM helpers.                            |
| `static/compose.js`        | Compose flow JS.                                       |
| `static/view.js`           | Viewer flow JS.                                        |
| `static/style.css`         | UI styles.                                             |
| `static/{robots,humans,llms,sitemap,security}.{txt,xml}` | Hygiene. |
| `*_test.go`                | Go test suites.                                        |
| `.repobility/access.yml`   | Endpoint-by-endpoint authorization matrix.             |
| `.github/workflows/ci.yml` | CI: matrix tests, lint, vuln scan, build.              |
| `docs/`                    | Scan history, ADRs.                                    |

## Style

- Standard `gofmt` formatting (tabs, no manual choices).
- Type hints / receivers follow Go conventions.
- `errors.Is`/`errors.As` for matching; never `==`/string-match unless
  the package documents that pattern.
- All randomness goes through `crypto/rand` (Go) or
  `crypto.getRandomValues` (browser). Never `math/rand` or
  `Math.random` in security-adjacent code.

## Pull-request checklist

- [ ] `go test -race ./...` passes locally.
- [ ] `gofmt -l .` is empty.
- [ ] `go vet ./...` clean.
- [ ] Wire-protocol or threat-model changes update `SECURITY.md` and
      `.repobility/access.yml`.
- [ ] User-visible changes update `README.md` and `CHANGELOG.md`.
- [ ] No new third-party origins in CSP.

## Reporting security issues

See [SECURITY.md](SECURITY.md). Please do not file public issues for
vulnerabilities — email the maintainer first.
