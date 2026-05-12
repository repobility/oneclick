// OneClick — encrypted one-time URL redirect service.
//
// Trust model: the server holds only ciphertext blobs. The AES-GCM key is
// generated in the browser and travels in the URL fragment (#k=...),
// which the browser never sends in HTTP requests — so the server cannot
// decrypt the destination URL even with full access to its own logs.
//
// Wire shape: client posts {ciphertext, iv} (both base64) to /api/links.
// Server stores them keyed by an unguessable 22-char urlsafe-base64 id.
// Recipient opens /l/<id>#k=<key>, browser fetches ciphertext via the
// API, decrypts, and redirects.
package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite"
)

const (
	maxCiphertextBytes = 16 * 1024 // 16 KiB after base64 expansion
	minCiphertextBytes = 24        // base64-encoded AES-GCM 16-byte tag
	ivBase64Len        = 16        // 12 raw bytes -> 16 base64 chars
	maxTTLSeconds      = 7 * 24 * 3600
	minTTLSeconds      = 60
	defaultTTLSeconds  = 24 * 3600
	maxClicks          = 100
	defaultClicks      = 1
	maxBodyBytes       = 64 * 1024
)

var (
	b64Re      = regexp.MustCompile(`^[A-Za-z0-9+/]+={0,2}$`)
	idRe       = regexp.MustCompile(`^[A-Za-z0-9_-]{20,32}$`)
	tmplCreate *template.Template
	tmplView   *template.Template
	tmplGone   *template.Template
)

func main() {
	host := os.Getenv("HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("ONECLICK_DB")
	if dbPath == "" {
		dbPath = "data/oneclick.db"
	}
	_ = os.MkdirAll(filepath.Dir(dbPath), 0o755)

	store, err := openStore(dbPath)
	if err != nil {
		slog.Error("open store failed", "err", err)
		os.Exit(1)
	}
	defer store.Close()

	parseTemplates()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(securityHeaders)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		n, _ := store.Count()
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "stored": n})
	})

	r.Post("/api/links", handleCreate(store))
	r.Get("/api/links/{id}/meta", handleMeta(store))
	r.Get("/api/links/{id}", handleConsume(store))

	r.Get("/", handleHome)
	r.Get("/l/{id}", handleView)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Periodic sweep.
	go func() {
		t := time.NewTicker(60 * time.Second)
		defer t.Stop()
		for range t.C {
			_ = store.Sweep()
		}
	}()

	addr := host + ":" + port
	slog.Info("OneClick listening", "addr", "http://"+addr)
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server exited", "err", err)
		os.Exit(1)
	}
}

func parseTemplates() {
	tmplCreate = template.Must(template.ParseFiles("templates/create.html"))
	tmplView = template.Must(template.ParseFiles("templates/view.html"))
	tmplGone = template.Must(template.ParseFiles("templates/gone.html"))
}

func handleHome(w http.ResponseWriter, _ *http.Request) {
	if err := tmplCreate.Execute(w, nil); err != nil {
		slog.Warn("create template failed", "err", err)
	}
}

func handleView(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !idRe.MatchString(id) {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := tmplView.Execute(w, map[string]string{"ID": id}); err != nil {
		slog.Warn("view template failed", "err", err)
	}
}

func handleCreate(store *Store) http.HandlerFunc {
	type req struct {
		Ciphertext string `json:"ciphertext"`
		IV         string `json:"iv"`
		TTLSeconds int    `json:"ttl_seconds"`
		MaxClicks  int    `json:"max_clicks"`
	}
	type res struct {
		ID                string `json:"id"`
		ExpiresAt         int64  `json:"expires_at"`
		ClicksRemaining   int    `json:"clicks_remaining"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		var p req
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid_payload")
			return
		}
		if !isValidCiphertext(p.Ciphertext) {
			writeErr(w, http.StatusBadRequest, "invalid_ciphertext")
			return
		}
		if !isValidIV(p.IV) {
			writeErr(w, http.StatusBadRequest, "invalid_iv")
			return
		}
		ttl := clamp(p.TTLSeconds, minTTLSeconds, maxTTLSeconds, defaultTTLSeconds)
		clicks := clamp(p.MaxClicks, 1, maxClicks, defaultClicks)

		id, expires, err := store.Insert(p.Ciphertext, p.IV, ttl, clicks)
		if err != nil {
			slog.Warn("store insert failed", "err", err)
			writeErr(w, http.StatusInternalServerError, "store_failed")
			return
		}
		writeJSON(w, http.StatusOK, res{ID: id, ExpiresAt: expires, ClicksRemaining: clicks})
	}
}

func handleMeta(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !idRe.MatchString(id) {
			writeErr(w, http.StatusBadRequest, "invalid_id")
			return
		}
		rec, err := store.Get(id)
		if err != nil {
			writeErr(w, http.StatusNotFound, "not_found")
			return
		}
		if rec.ExpiresAt <= time.Now().Unix() {
			_ = store.Delete(id)
			writeErr(w, http.StatusGone, "expired")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"expires_at":       rec.ExpiresAt,
			"clicks_remaining": rec.ClicksRemaining,
			"created_at":       rec.CreatedAt,
		})
	}
}

func handleConsume(store *Store) http.HandlerFunc {
	type res struct {
		Ciphertext       string `json:"ciphertext"`
		IV               string `json:"iv"`
		ExpiresAt        int64  `json:"expires_at"`
		ClicksRemaining  int    `json:"clicks_remaining"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !idRe.MatchString(id) {
			writeErr(w, http.StatusBadRequest, "invalid_id")
			return
		}
		rec, remaining, err := store.Consume(id)
		if err != nil {
			writeErr(w, http.StatusNotFound, "not_found")
			return
		}
		writeJSON(w, http.StatusOK, res{
			Ciphertext:      rec.Ciphertext,
			IV:              rec.IV,
			ExpiresAt:       rec.ExpiresAt,
			ClicksRemaining: remaining,
		})
	}
}

// ----- helpers -----

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strict same-origin policy. Inline styles allowed only because the
		// templated pages use small inline classes; no inline JS.
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; img-src 'self' data:; "+
				"script-src 'self'; style-src 'self' 'unsafe-inline'; "+
				"connect-src 'self'; base-uri 'self'; "+
				"form-action 'none'; frame-ancestors 'none'")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func isValidCiphertext(s string) bool {
	return len(s) >= minCiphertextBytes && len(s) <= maxCiphertextBytes && b64Re.MatchString(s)
}

func isValidIV(s string) bool {
	return len(s) == ivBase64Len && b64Re.MatchString(s)
}

func clamp(v, lo, hi, def int) int {
	if v <= 0 {
		v = def
	}
	if v < lo {
		v = lo
	}
	if v > hi {
		v = hi
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}

// 16 random bytes, urlsafe-base64 → 22-char id, ~128-bit search space.
func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

