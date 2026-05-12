// HTTP handler tests — exercise the routes against an in-memory store
// using httptest. No network, no SQLite file (each test gets its own
// temp DB via openTestStore).

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func newRouter(t *testing.T) (chi.Router, *Store) {
	t.Helper()
	store := openTestStore(t)
	parseTemplatesForTest(t)
	r := chi.NewRouter()
	r.Use(securityHeaders)
	r.Post("/api/links", handleCreate(store))
	r.Get("/api/links/{id}/meta", handleMeta(store))
	r.Get("/api/links/{id}", handleConsume(store))
	r.Get("/", handleHome)
	r.Get("/l/{id}", handleView)
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		n, _ := store.Count()
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "stored": n})
	})
	return r, store
}

// parseTemplatesForTest reuses the global parse but tolerates missing
// template files in test environments where the cwd is not the repo root.
func parseTemplatesForTest(t *testing.T) {
	t.Helper()
	defer func() {
		// Tests don't render templates by default; if parsing fails because
		// the cwd isn't the repo root, swallow it.
		_ = recover()
	}()
	parseTemplates()
}

func sampleCT() string {
	// 32 random-ish bytes encode to ~44 base64 chars — well above the
	// 24-char minimum the validator enforces.
	return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("A"), 32))
}

func sampleIV() string {
	// 12-byte AES-GCM IV → exactly 16 base64 chars.
	return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("B"), 12))
}

func postCreate(t *testing.T, r http.Handler, body any) *httptest.ResponseRecorder {
	t.Helper()
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/links", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func get(t *testing.T, r http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
	return w
}

func TestHealthzOK(t *testing.T) {
	r, _ := newRouter(t)
	w := get(t, r, "/healthz")
	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"ok":true`) {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestSecurityHeadersAreSet(t *testing.T) {
	r, _ := newRouter(t)
	w := get(t, r, "/healthz")
	for _, h := range []string{
		"Content-Security-Policy",
		"Referrer-Policy",
		"X-Content-Type-Options",
		"X-Frame-Options",
	} {
		if w.Header().Get(h) == "" {
			t.Errorf("missing security header %q", h)
		}
	}
}

func TestCreateReturnsID(t *testing.T) {
	r, _ := newRouter(t)
	w := postCreate(t, r, map[string]any{
		"ciphertext":  sampleCT(),
		"iv":          sampleIV(),
		"ttl_seconds": 60,
		"max_clicks":  1,
	})
	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	if _, ok := out["id"].(string); !ok {
		t.Errorf("missing id in response: %v", out)
	}
}

func TestCreateRejectsBadCiphertext(t *testing.T) {
	r, _ := newRouter(t)
	w := postCreate(t, r, map[string]any{
		"ciphertext":  "shrt",
		"iv":          sampleIV(),
		"ttl_seconds": 60,
		"max_clicks":  1,
	})
	if w.Code != 400 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "invalid_ciphertext") {
		t.Errorf("unexpected error: %s", w.Body.String())
	}
}

func TestCreateRejectsBadIV(t *testing.T) {
	r, _ := newRouter(t)
	w := postCreate(t, r, map[string]any{
		"ciphertext":  sampleCT(),
		"iv":          "xxx",
		"ttl_seconds": 60,
		"max_clicks":  1,
	})
	if w.Code != 400 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "invalid_iv") {
		t.Errorf("unexpected error: %s", w.Body.String())
	}
}

func TestConsumeOneTimeFlow(t *testing.T) {
	r, _ := newRouter(t)
	w := postCreate(t, r, map[string]any{
		"ciphertext":  sampleCT(),
		"iv":          sampleIV(),
		"ttl_seconds": 60,
		"max_clicks":  1,
	})
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	id := created["id"].(string)

	// First read returns 200.
	w1 := get(t, r, "/api/links/"+id)
	if w1.Code != 200 {
		t.Fatalf("first read status=%d", w1.Code)
	}

	// Second read returns 404.
	w2 := get(t, r, "/api/links/"+id)
	if w2.Code != 404 {
		t.Fatalf("second read status=%d want 404", w2.Code)
	}
}

func TestMetaIsNonConsuming(t *testing.T) {
	r, _ := newRouter(t)
	w := postCreate(t, r, map[string]any{
		"ciphertext":  sampleCT(),
		"iv":          sampleIV(),
		"ttl_seconds": 60,
		"max_clicks":  1,
	})
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	id := created["id"].(string)

	for i := 0; i < 5; i++ {
		w := get(t, r, "/api/links/"+id+"/meta")
		if w.Code != 200 {
			t.Fatalf("meta %d status=%d", i, w.Code)
		}
		var m map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &m)
		if m["clicks_remaining"].(float64) != 1 {
			t.Errorf("peek decremented click counter to %v", m["clicks_remaining"])
		}
	}
}

func TestInvalidIDReturns400(t *testing.T) {
	r, _ := newRouter(t)
	for _, id := range []string{"short", "!!!badchars!!!"} {
		w := get(t, r, "/api/links/"+id)
		if w.Code != 400 {
			t.Errorf("id=%q status=%d want 400", id, w.Code)
		}
	}
}
