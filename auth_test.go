// Authorization-focused tests — capability-URL invariants.
//
// OneClick uses capability-URL auth: the unguessable 22-char id is one
// half of the credential, the 32-byte AES-GCM key in the URL fragment is
// the other. No user table, no session, no role. These tests prove the
// properties that fall out of that model.

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAUTH01_IDReturnsCiphertextNotPlaintext(t *testing.T) {
	r, _ := newRouter(t)
	// Use a ciphertext we recognize so we can prove the server echoes
	// it byte-for-byte without ever decrypting.
	ctBytes := bytes.Repeat([]byte("Z"), 64)
	ct := base64.StdEncoding.EncodeToString(ctBytes)
	w := postCreate(t, r, map[string]any{
		"ciphertext":  ct,
		"iv":          sampleIV(),
		"ttl_seconds": 60,
		"max_clicks":  1,
	})
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	id := created["id"].(string)

	w2 := get(t, r, "/api/links/"+id)
	if w2.Code != 200 {
		t.Fatalf("consume status=%d", w2.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(w2.Body.Bytes(), &body)
	if body["ciphertext"] != ct {
		t.Errorf("server must echo ciphertext verbatim; got mismatch")
	}
	// The plaintext URL never appears anywhere in the response body.
	if bytes.Contains(w2.Body.Bytes(), ctBytes) {
		t.Errorf("raw plaintext leaked in response body")
	}
}

func TestAUTH02_OneTimeReadIsAtomic(t *testing.T) {
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

	first := get(t, r, "/api/links/"+id)
	if first.Code != http.StatusOK {
		t.Fatalf("first consume status=%d", first.Code)
	}
	second := get(t, r, "/api/links/"+id)
	if second.Code != http.StatusNotFound {
		t.Errorf("second consume status=%d want 404", second.Code)
	}
}

func TestAUTH03_IDsAreUnguessable(t *testing.T) {
	r, _ := newRouter(t)
	seen := make(map[string]struct{})
	for i := 0; i < 50; i++ {
		w := postCreate(t, r, map[string]any{
			"ciphertext":  sampleCT(),
			"iv":          sampleIV(),
			"ttl_seconds": 60,
			"max_clicks":  1,
		})
		var created map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &created)
		id, _ := created["id"].(string)
		if !idRe.MatchString(id) {
			t.Errorf("id %q doesn't match expected shape", id)
		}
		seen[id] = struct{}{}
	}
	if len(seen) != 50 {
		t.Errorf("only %d unique ids out of 50 — entropy too low", len(seen))
	}
}

func TestAUTH04_UnknownIDReturnsGenericNotFound(t *testing.T) {
	r, _ := newRouter(t)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/api/links/aaaaaaaaaaaaaaaaaaaa", nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("status=%d want 404", w.Code)
	}
}

func TestAUTH05_MetaIsNonConsuming(t *testing.T) {
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
			t.Fatalf("peek %d status=%d", i, w.Code)
		}
	}
	// And then the actual consume still works.
	if get(t, r, "/api/links/"+id).Code != 200 {
		t.Errorf("consume after peeks should still succeed")
	}
}

func TestAUTH06_NoListOrDeleteEndpoints(t *testing.T) {
	r, _ := newRouter(t)
	// No GET on /api/links collection.
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/api/links", nil))
	if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Errorf("/api/links GET status=%d want 404 or 405", w.Code)
	}

	// No DELETE / PUT on a specific id.
	for _, m := range []string{"DELETE", "PUT", "PATCH"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(m, "/api/links/aaaaaaaaaaaaaaaaaaaa", nil))
		if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s status=%d want 404 or 405", m, w.Code)
		}
	}
}
