<!-- Thanks for contributing! Keep the trust model intact. -->

## Summary

<!-- One paragraph: what changed and why. Link the issue if there is one. -->

## Trust-model impact

- [ ] No change. (Pure refactor, docs, or UI tweak.)
- [ ] Touches the wire protocol — `SECURITY.md` and `.repobility/access.yml` are updated.
- [ ] Touches the server validation rules — added/updated tests in `handlers_test.go` or `store_test.go`.
- [ ] Touches the cryptographic core — added/updated tests in `auth_test.go` and `static/crypto.js`.
- [ ] Changes the capability-URL contract — `.repobility/access.yml` is updated and `auth_test.go` covers the new behavior.

## Checklist

- [ ] `go test -race ./...` passes locally.
- [ ] `gofmt -l .` is empty.
- [ ] `go vet ./...` clean.
- [ ] No new third-party origins in CSP allowlist.
- [ ] Browser code uses `crypto.getRandomValues` / `crypto.subtle` — no `Math.random` in security-adjacent code.
- [ ] User-visible changes noted in `CHANGELOG.md` under `[Unreleased]`.

## Test plan

<!-- How did you verify this? Manual two-tab flow, curl against the API,
     go test, etc. -->
