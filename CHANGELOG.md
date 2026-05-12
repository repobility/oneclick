# Changelog

All notable changes to OneClick are documented here. Format:
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Versioning: [SemVer](https://semver.org/).

## [1.0.0] — 2026-05-12

### Added

- Initial release.
- End-to-end encrypted one-time URL redirect using Web Crypto AES-GCM
  (256-bit key, 96-bit IV).
- Go server with chi router, html/template, and a pure-Go SQLite store
  (`modernc.org/sqlite`).
- Atomic one-time-click semantics; record + row deleted in the same
  transaction on the final read.
- Strict security headers middleware: CSP, Referrer-Policy: no-referrer,
  XCTO nosniff, X-Frame-Options DENY.
- 20 Go tests across `store_test.go`, `handlers_test.go`, `auth_test.go`.
- `.repobility/access.yml` documenting the capability-URL model with
  per-endpoint scope/owner/tenant markers.
- GitHub Actions CI: test matrix on Go 1.22/1.23/1.24, gofmt + go vet +
  staticcheck, govulncheck, build verification.
- Web hygiene: robots.txt, sitemap.xml, humans.txt, llms.txt,
  /.well-known/security.txt (all served at root paths).
- `SECURITY.md`, `ARCHITECTURE.md`, `CONTRIBUTING.md`, ADRs.

[1.0.0]: https://github.com/repobility/oneclick/releases/tag/v1.0.0
