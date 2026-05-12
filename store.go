package main

import (
	"database/sql"
	"errors"
	"log/slog"
	"time"
)

// Store wraps a SQLite database holding only ciphertext + metadata.
// It never holds keys; it cannot decrypt anything.
type Store struct {
	db *sql.DB
}

// Record is the in-memory shape of a stored row.
type Record struct {
	ID              string
	Ciphertext      string
	IV              string
	ExpiresAt       int64
	ClicksRemaining int
	CreatedAt       int64
}

const schema = `
CREATE TABLE IF NOT EXISTS links (
	id TEXT PRIMARY KEY,
	ciphertext TEXT NOT NULL,
	iv TEXT NOT NULL,
	expires_at INTEGER NOT NULL,
	clicks_remaining INTEGER NOT NULL,
	created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_links_expires ON links(expires_at);
`

func openStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(on)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite is fine with one writer
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close releases the underlying DB handle.
func (s *Store) Close() error { return s.db.Close() }

// Count returns the number of stored records (live + expired-but-unswept).
func (s *Store) Count() (int, error) {
	var n int
	err := s.db.QueryRow("SELECT COUNT(*) FROM links").Scan(&n)
	return n, err
}

// Insert stores a new ciphertext blob and returns its generated id and expiry timestamp.
func (s *Store) Insert(ciphertext, iv string, ttlSeconds, clicks int) (string, int64, error) {
	now := time.Now().Unix()
	expires := now + int64(ttlSeconds)
	for attempt := 0; attempt < 5; attempt++ {
		id, err := generateID()
		if err != nil {
			return "", 0, err
		}
		_, err = s.db.Exec(
			"INSERT INTO links (id, ciphertext, iv, expires_at, clicks_remaining, created_at) VALUES (?, ?, ?, ?, ?, ?)",
			id, ciphertext, iv, expires, clicks, now,
		)
		if err == nil {
			return id, expires, nil
		}
		// Retry on PK collision; surface any other error.
		if !isUniqueViolation(err) {
			return "", 0, err
		}
	}
	return "", 0, errors.New("id_collision")
}

// Get returns the record without consuming a click. Returns sql.ErrNoRows
// if missing.
func (s *Store) Get(id string) (*Record, error) {
	r := &Record{ID: id}
	err := s.db.QueryRow(
		"SELECT ciphertext, iv, expires_at, clicks_remaining, created_at FROM links WHERE id = ?",
		id,
	).Scan(&r.Ciphertext, &r.IV, &r.ExpiresAt, &r.ClicksRemaining, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Consume returns the record and atomically decrements its click counter.
// If the counter hits zero or the record is expired, it is deleted.
// Returns (record, clicks_remaining_after, error).
func (s *Store) Consume(id string) (*Record, int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, 0, err
	}
	// Rollback on every exit. If Commit ran successfully Rollback returns
	// sql.ErrTxDone, which is the expected, ignorable case — anything else
	// is a real error we want in the logs.
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			slog.Warn("tx rollback failed", "err", rbErr)
		}
	}()

	r := &Record{ID: id}
	err = tx.QueryRow(
		"SELECT ciphertext, iv, expires_at, clicks_remaining, created_at FROM links WHERE id = ?",
		id,
	).Scan(&r.Ciphertext, &r.IV, &r.ExpiresAt, &r.ClicksRemaining, &r.CreatedAt)
	if err != nil {
		return nil, 0, err
	}
	if r.ExpiresAt <= time.Now().Unix() {
		if _, exErr := tx.Exec("DELETE FROM links WHERE id = ?", id); exErr != nil {
			slog.Warn("delete expired row failed", "id", id, "err", exErr)
		}
		if cmErr := tx.Commit(); cmErr != nil {
			slog.Warn("commit expired-delete failed", "id", id, "err", cmErr)
		}
		return nil, 0, errors.New("expired")
	}
	remaining := r.ClicksRemaining - 1
	if remaining <= 0 {
		if _, err := tx.Exec("DELETE FROM links WHERE id = ?", id); err != nil {
			return nil, 0, err
		}
	} else {
		if _, err := tx.Exec("UPDATE links SET clicks_remaining = ? WHERE id = ?", remaining, id); err != nil {
			return nil, 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}
	return r, remaining, nil
}

// Delete removes a record by id. Used by handlers on the expired path.
func (s *Store) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM links WHERE id = ?", id)
	return err
}

// Sweep deletes any record past its expiry. Returns the number of rows deleted.
func (s *Store) Sweep() error {
	_, err := s.db.Exec("DELETE FROM links WHERE expires_at <= ?", time.Now().Unix())
	return err
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return containsAny(err.Error(), "UNIQUE constraint failed", "constraint failed")
}

func containsAny(s string, needles ...string) bool {
	for _, n := range needles {
		if len(n) > 0 && len(s) >= len(n) {
			for i := 0; i+len(n) <= len(s); i++ {
				if s[i:i+len(n)] == n {
					return true
				}
			}
		}
	}
	return false
}
