// Store tests — exercise the SQLite layer in isolation.

package main

import (
	"path/filepath"
	"testing"
	"time"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := openStore(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("openStore: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestInsertReturnsUnguessableID(t *testing.T) {
	s := openTestStore(t)
	id, exp, err := s.Insert("Y2lwaGVydGV4dGN0Y3Q=", "bm9uY2VfYmFzZTY0YWE=", 60, 1)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if !idRe.MatchString(id) {
		t.Errorf("id %q doesn't match expected shape", id)
	}
	if exp <= time.Now().Unix() {
		t.Errorf("expiry %d is in the past", exp)
	}
}

func TestGetMissingReturnsError(t *testing.T) {
	s := openTestStore(t)
	_, err := s.Get("nonexistent_id_zzzzz")
	if err == nil {
		t.Fatal("expected error for missing record")
	}
}

func TestConsumeAtomicallyDeletesOnLastClick(t *testing.T) {
	s := openTestStore(t)
	id, _, err := s.Insert("Y2lwaGVydGV4dGN0Y3Q=", "bm9uY2VfYmFzZTY0YWE=", 60, 1)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	_, remaining, err := s.Consume(id)
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if remaining != 0 {
		t.Errorf("remaining = %d, want 0", remaining)
	}
	_, _, err = s.Consume(id)
	if err == nil {
		t.Fatal("second consume must fail (record gone)")
	}
}

func TestConsumeDecrementsWhenMoreClicksRemaining(t *testing.T) {
	s := openTestStore(t)
	id, _, _ := s.Insert("Y2lwaGVydGV4dGN0Y3Q=", "bm9uY2VfYmFzZTY0YWE=", 60, 3)
	for want := 2; want >= 0; want-- {
		_, remaining, err := s.Consume(id)
		if err != nil {
			t.Fatalf("consume %d: %v", want, err)
		}
		if remaining != want {
			t.Errorf("after click, remaining=%d want %d", remaining, want)
		}
	}
	_, _, err := s.Consume(id)
	if err == nil {
		t.Fatal("fourth consume must fail")
	}
}

func TestSweepDeletesExpiredRows(t *testing.T) {
	s := openTestStore(t)
	id, _, _ := s.Insert("Y2lwaGVydGV4dGN0Y3Q=", "bm9uY2VfYmFzZTY0YWE=", 60, 1)
	// Force the row to be expired.
	if _, err := s.db.Exec("UPDATE links SET expires_at = 1 WHERE id = ?", id); err != nil {
		t.Fatalf("force-expire: %v", err)
	}
	if err := s.Sweep(); err != nil {
		t.Fatalf("sweep: %v", err)
	}
	_, err := s.Get(id)
	if err == nil {
		t.Fatal("expected swept record to be gone")
	}
}

func TestConsumeOnExpiredReturnsError(t *testing.T) {
	s := openTestStore(t)
	id, _, _ := s.Insert("Y2lwaGVydGV4dGN0Y3Q=", "bm9uY2VfYmFzZTY0YWE=", 60, 1)
	if _, err := s.db.Exec("UPDATE links SET expires_at = 1 WHERE id = ?", id); err != nil {
		t.Fatalf("force-expire: %v", err)
	}
	_, _, err := s.Consume(id)
	if err == nil {
		t.Fatal("expected consume on expired record to fail")
	}
}
