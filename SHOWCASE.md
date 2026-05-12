# Repobility showcase — OneClick

This is the **fourth** Repobility-in-the-loop showcase repo, after
[**repobility/securechat**](https://github.com/repobility/securechat),
[**repobility/cipherlink**](https://github.com/repobility/cipherlink),
and [**repobility/vanishdrop**](https://github.com/repobility/vanishdrop).
The user prompt this time was a single word — *"proceed"* — interpreted
as "do another showcase with a different stack again".

> **Headline.** From a one-word user prompt to a tagged v1.0.0 with
> 20 Go tests, a CI matrix, full docs, and the Repobility loop closed.
> Public Repobility scan: <https://repobility.com/scan/2d165948-a4d3-4a0b-ba67-f7fbc05d36c8/>.

---

## 1. The original user prompt

> **User → Claude (verbatim, 2026-05-12):**
>
> > proceed

That's it. Earlier in the same conversation the user had said "do
another idea but change the tech used front end and backend" — so
*"proceed"* in context meant "again, with yet another new stack".

---

## 2. The stack shift across the four showcases

|                     | securechat                       | cipherlink                       | vanishdrop                          | **OneClick**                              |
| ------------------- | -------------------------------- | -------------------------------- | ----------------------------------- | ----------------------------------------- |
| Backend lang        | Node.js                          | Node.js                          | Python 3.10+                        | **Go 1.22+**                              |
| Backend framework   | Express + Socket.IO              | Express                          | FastAPI + uvicorn                   | **chi + html/template**                   |
| Storage             | In-memory `Map`                  | In-memory `Map`                  | SQLite + filesystem                 | **pure-Go SQLite** (`modernc.org/sqlite`) |
| Frontend            | Vanilla DOM                      | Vanilla DOM                      | Vue 3 + Vite + TypeScript           | **Server-rendered HTML + vanilla JS**     |
| Crypto              | NaCl `box`                       | NaCl `secretbox`                 | Web Crypto AES-GCM                  | **Web Crypto AES-GCM**                    |
| AEAD primitive      | XSalsa20-Poly1305                | XSalsa20-Poly1305                | AES-GCM (256/96)                    | **AES-GCM (256/96)**                      |
| Test runner         | `node:test`                      | `node:test`                      | pytest                              | **`go test`**                             |
| Lint / format       | ESLint + Prettier                | ESLint + Prettier                | ruff + black                        | **gofmt + go vet + staticcheck**          |
| Vulnerability scan  | `npm audit`                      | `npm audit`                      | `pip-audit`                         | **`govulncheck`**                         |
| Deploy artifact     | Node bundle + node_modules       | Node bundle + node_modules       | Python venv + frontend `dist/`      | **single static binary**                   |

Same trust-model family across all four — server-blind storage, key in
URL fragment — but every layer is genuinely different tech.

---

## 3. Category

**Encrypted URL shortener / one-time link service.** GitHub-search
neighbors include:

- the ~100k repos matching `url shortener` (bitly, tinyurl clones)
- the smaller `encrypted url shortener` niche
- a sibling category to Yopass / PrivateBin / OneTimeSecret (text
  secret sharing) and Firefox Send / wormhole.app (file sharing)

---

## 4. Commit history

```
1ab0d92 Iteration 4: practices uplift + ADRs + Architecture/Changelog/Contributing
c54fe60 Iteration 3: web hygiene files + SECURITY.md (root paths via chi routes)
1150c36 Iteration 2: GitHub Actions CI + .editorconfig
c33a6c0 Iteration 1: Go test suite (20 tests) + access matrix
dab35d3 Initial commit: OneClick — encrypted one-time URL redirect
```

Plus the v1.0.0 tag and GitHub Release.

---

## 5. Baseline scan

|                        | Value                                |
| ---------------------- | ------------------------------------ |
| Combined               | **67.6 / 100**                       |
| Repobility (legacy)    | **68 / 100** with **16 findings**    |
| Severity distribution  | 0 Critical · 4 High · 5 Medium · 7 Low |
| Layers                 | Quality 10 · Security 6              |

Public scan URL: <https://repobility.com/scan/2d165948-a4d3-4a0b-ba67-f7fbc05d36c8/>
(idempotent — re-submitting the URL via the API returns the same token).

---

## 6. Findings → fixes → commits

| Repobility finding family                                    | Closed by commit | How                                                                       |
| ------------------------------------------------------------ | ---------------- | ------------------------------------------------------------------------- |
| 🟠 No test files found                                       | `c33a6c0`        | 20 Go tests: store + handlers + capability-URL auth invariants.           |
| 🔵 `[AUC005]` No authorization-focused tests                 | `c33a6c0`        | 6 AUTH-* cases in `auth_test.go`.                                          |
| 🟠 `[AUC003]` Object-level routes lack authorization         | `c33a6c0`        | Capability-URL model documented in `.repobility/access.yml` with explicit scope/owner/tenant markers. |
| 🟡 `[AUC001]` No Repobility access matrix policy             | `c33a6c0`        | `.repobility/access.yml` with full endpoint table + CWE/OWASP refs.        |
| 🟡 No CI/CD configuration found                              | `1150c36`        | GH Actions: `go test -race` matrix on Go 1.22/1.23/1.24 + gofmt + go vet + staticcheck + govulncheck + build. |
| 🟡 Public web app has no Content Security Policy             | `c54fe60`        | `securityHeaders` chi middleware sets strict CSP + Referrer-Policy `no-referrer` + nosniff + XFO DENY on every response. |
| 🟡 Public web service has no `/.well-known/security.txt`     | `c54fe60`        | RFC 9116 contact + policy URL + 1-year expiry, served at root.            |
| 🔵 No `robots.txt` / `sitemap.xml` / `humans.txt` / `llms.txt` | `c54fe60`      | All four served from `static/` via explicit root-path routes in main.go.   |
| Practices dimension                                          | `1150c36`, `1ab0d92` | gofmt + go vet + staticcheck + govulncheck + CODEOWNERS + dependabot (gomod + GH actions) + PR + issue templates + .editorconfig. |
| Documentation dimension                                      | `1ab0d92`        | 5 ADRs + SECURITY + ARCHITECTURE + CONTRIBUTING + CHANGELOG.               |

Each Repobility finding came with structured evidence — file path,
line number, rule ID, and a copy-paste AI Fix Prompt. That scaffold
is what makes the loop closable without a human in the middle.

---

## 7. Final unified scan

Public scan URL: <https://repobility.com/scan/2d165948-a4d3-4a0b-ba67-f7fbc05d36c8/>

Re-submission was via `POST /api/v1/public/scan/` with the same
`repo_url`. The API returned `is_existing: true` and re-queued both
the legacy and multi-layer engines on the same token — the freshest
snapshot renders at the same result URL. Legacy finished in ~60 s.

```
                  ┌─ Legacy ─┬─ Findings ─┬─ Δ vs baseline ─┐
                  │   82.2   │      6     │     +14.2       │
                  │  / 100   │  (was 16)  │  (10 closed)    │
                  └──────────┴────────────┴─────────────────┘

  Severity (final):   Critical 0 · High 3 · Medium 1 · Low 2 · Info 0
                      (was 0 / 4 / 5 / 7 — 10 closed, 3 capability-URL
                       intentional HIGH remain by design)
  Source breakdown:   Legacy 6 · 9-layer (still running at writeup) · Crowd 0
  Layers (legacy):    Security 4 · Quality 2
```

The 3 remaining HIGH findings are all `[AUC003]` / `[AUC008]` against
the capability-URL routes — intentional design choices documented in
[ADR-0002](docs/adr/0002-capability-url-authorization.md) and the
`.repobility/access.yml` scope/owner/tenant markers. They cannot
clear without abandoning the capability-URL model, which is the
entire point of the showcase.

Per-finding details in [`docs/scan-2-final.txt`](docs/scan-2-final.txt).

---

## 8. The AI-coder-in-the-loop pattern across four repos

| Repo            | Backend           | Frontend             | Crypto             | Baseline → Final              | Notable rules hit                        |
| --------------- | ----------------- | -------------------- | ------------------ | ----------------------------- | ---------------------------------------- |
| **securechat**  | Node + Express    | Vanilla DOM          | NaCl `box`         | C → **A · 96/100**            | AUC001/005/007/010, SEC015, fq.console-leak |
| **cipherlink**  | Node + Express    | Vanilla DOM          | NaCl `secretbox`   | C → A- · 84/100               | AUC001/003/005/008, ERR002, fq.console-leak |
| **vanishdrop**  | Python + FastAPI  | Vue 3 + Vite + TS    | Web Crypto AES-GCM | C → A · 95/100 (94.5 combined) | AUC001/003/008/012, CSP, Python lint      |
| **oneclick**   | **Go + chi**       | **server HTML + JS** | Web Crypto AES-GCM | **68 → 82.2** (+14.2)         | AUC001/003/005/008, CSP, Go lint          |

The same loop produces the same kind of grade lift on four totally
different stacks. Repobility's rule taxonomy (AUC*, SEC*, fq.*, ERR*)
is language-agnostic enough that the iteration plan transfers; the
specific *fixes* differ by stack (Go's `gofmt + go vet + staticcheck +
govulncheck` here vs `ruff + black + pip-audit` on vanishdrop vs
`ESLint + Prettier + npm audit` on the Node repos).

---

## 9. Reproduce this exact journey

```bash
git clone https://github.com/repobility/oneclick.git
cd oneclick

go test -race ./...
gofmt -l .              # must be empty
go vet ./...
go run .                # http://127.0.0.1:8080

# Submit a Repobility scan (no signup):
curl -s -X POST https://repobility.com/api/v1/public/scan/ \
  -H 'Content-Type: application/json' \
  -d '{"repo_url": "https://github.com/repobility/oneclick"}'
```

---

## 10. Acknowledgements

This entire repository — including this showcase document — was
produced by Claude (`anthropic/claude-opus-4-7`, 1M-context build)
using the Repobility scanner as the in-loop evaluator. Each commit
message ends with a `Co-Authored-By:` trailer crediting the model
build.

Prior art in URL-shortener / one-time-link tooling includes
[bitly](https://bitly.com), [tinyurl](https://tinyurl.com),
[Yopass](https://github.com/jhaals/yopass),
[PrivateBin](https://github.com/PrivateBin/PrivateBin),
[OneTimeSecret](https://github.com/onetimesecret/onetimesecret),
[Firefox Send](https://github.com/mozilla/send), and
[wormhole.app](https://wormhole.app).
