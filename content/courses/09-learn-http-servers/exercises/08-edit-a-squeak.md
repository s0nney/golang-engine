---
title: Edit a Squeak
difficulty: medium
after: authorization-and-webhooks
hints:
  - 'Work top to bottom and `return` after every error response: caller (`userIDFrom`), path ID (`uuid.Parse(r.PathValue("id"))`), body (`http.MaxBytesReader` then `json.UnmarshalRead` with `json.RejectUnknownMembers(true)`), rules, then the store. `errors.AsType[*http.MaxBytesError](err)` tells a too-large body (413) from broken JSON (400).'
  - 'Decode into a struct with a **pointer** field, `Body *string`, so you can tell `{}` (nil: 400) from `{"body":""}` (present but empty: 422). Trim spaces before counting runes with `utf8.RuneCountInString`.'
  - 'For the store: `errors.Is(err, ErrNotFound)` is a 404, and any other error is a 500 whose message is a fixed string like `"couldn''t edit squeak"`. Never send `err.Error()`: it can contain database details. The window check is `cfg.now().Sub(sq.CreatedAt) > editWindow`.'
exercise:
  starter: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"sync"
    	"time"
    	"uuid"
    )

    type Squeak struct {
    	ID        uuid.UUID `json:"id"`
    	AuthorID  uuid.UUID `json:"author_id"`
    	Body      string    `json:"body"`
    	CreatedAt time.Time `json:"created_at"`
    	EditedAt  time.Time `json:"edited_at,omitzero"`
    }

    var ErrNotFound = errors.New("squeak not found")

    type SqueakStore interface {
    	GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error)
    	UpdateSqueak(ctx context.Context, sq Squeak) error
    }

    type MemoryStore struct {
    	mu      sync.Mutex
    	squeaks map[uuid.UUID]Squeak
    }

    func NewMemoryStore(sqs ...Squeak) *MemoryStore {
    	s := &MemoryStore{squeaks: make(map[uuid.UUID]Squeak)}
    	for _, sq := range sqs {
    		s.squeaks[sq.ID] = sq
    	}
    	return s
    }

    func (s *MemoryStore) GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error) {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	sq, ok := s.squeaks[id]
    	if !ok {
    		return Squeak{}, ErrNotFound
    	}
    	return sq, nil
    }

    func (s *MemoryStore) UpdateSqueak(ctx context.Context, sq Squeak) error {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	if _, ok := s.squeaks[sq.ID]; !ok {
    		return ErrNotFound
    	}
    	s.squeaks[sq.ID] = sq
    	return nil
    }

    type ctxKey struct{}

    // withUser returns r as seen by a handler behind the auth middleware.
    func withUser(r *http.Request, id uuid.UUID) *http.Request {
    	return r.WithContext(context.WithValue(r.Context(), ctxKey{}, id))
    }

    func userIDFrom(ctx context.Context) (uuid.UUID, bool) {
    	id, ok := ctx.Value(ctxKey{}).(uuid.UUID)
    	return id, ok
    }

    func respondWithJSON(w http.ResponseWriter, code int, payload any) {
    	data, err := json.Marshal(payload)
    	if err != nil {
    		w.WriteHeader(http.StatusInternalServerError)
    		return
    	}
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(code)
    	w.Write(data)
    }

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	respondWithJSON(w, code, map[string]string{"error": msg})
    }

    const (
    	maxEditBytes = 1 << 10
    	maxSqueakLen = 140
    	editWindow   = 10 * time.Minute
    )

    type apiConfig struct {
    	store SqueakStore
    	now   func() time.Time
    }

    func (cfg *apiConfig) handleEditSqueak(w http.ResponseWriter, r *http.Request) {
    	respondWithError(w, http.StatusNotImplemented, "not implemented")
    }

    func main() {
    	pip := uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	posted := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
    	sq := Squeak{ID: uuid.NewV7(), AuthorID: pip, Body: "teh cheese", CreatedAt: posted}
    	cfg := &apiConfig{store: NewMemoryStore(sq), now: func() time.Time { return posted.Add(time.Minute) }}

    	mux := http.NewServeMux()
    	mux.HandleFunc("PATCH /api/squeaks/{id}", cfg.handleEditSqueak)
    	req := httptest.NewRequest("PATCH", "/api/squeaks/"+sq.ID.String(), strings.NewReader(`{"body":"the cheese"}`))
    	rec := httptest.NewRecorder()
    	mux.ServeHTTP(rec, withUser(req, pip))
    	fmt.Println(rec.Code, rec.Body.String())
    }
  solution: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"sync"
    	"time"
    	"unicode/utf8"
    	"uuid"
    )

    type Squeak struct {
    	ID        uuid.UUID `json:"id"`
    	AuthorID  uuid.UUID `json:"author_id"`
    	Body      string    `json:"body"`
    	CreatedAt time.Time `json:"created_at"`
    	EditedAt  time.Time `json:"edited_at,omitzero"`
    }

    var ErrNotFound = errors.New("squeak not found")

    type SqueakStore interface {
    	GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error)
    	UpdateSqueak(ctx context.Context, sq Squeak) error
    }

    type MemoryStore struct {
    	mu      sync.Mutex
    	squeaks map[uuid.UUID]Squeak
    }

    func NewMemoryStore(sqs ...Squeak) *MemoryStore {
    	s := &MemoryStore{squeaks: make(map[uuid.UUID]Squeak)}
    	for _, sq := range sqs {
    		s.squeaks[sq.ID] = sq
    	}
    	return s
    }

    func (s *MemoryStore) GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error) {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	sq, ok := s.squeaks[id]
    	if !ok {
    		return Squeak{}, ErrNotFound
    	}
    	return sq, nil
    }

    func (s *MemoryStore) UpdateSqueak(ctx context.Context, sq Squeak) error {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	if _, ok := s.squeaks[sq.ID]; !ok {
    		return ErrNotFound
    	}
    	s.squeaks[sq.ID] = sq
    	return nil
    }

    type ctxKey struct{}

    func withUser(r *http.Request, id uuid.UUID) *http.Request {
    	return r.WithContext(context.WithValue(r.Context(), ctxKey{}, id))
    }

    func userIDFrom(ctx context.Context) (uuid.UUID, bool) {
    	id, ok := ctx.Value(ctxKey{}).(uuid.UUID)
    	return id, ok
    }

    func respondWithJSON(w http.ResponseWriter, code int, payload any) {
    	data, err := json.Marshal(payload)
    	if err != nil {
    		w.WriteHeader(http.StatusInternalServerError)
    		return
    	}
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(code)
    	w.Write(data)
    }

    func respondWithError(w http.ResponseWriter, code int, msg string) {
    	respondWithJSON(w, code, map[string]string{"error": msg})
    }

    const (
    	maxEditBytes = 1 << 10
    	maxSqueakLen = 140
    	editWindow   = 10 * time.Minute
    )

    type apiConfig struct {
    	store SqueakStore
    	now   func() time.Time
    }

    type editParams struct {
    	Body *string `json:"body"`
    }

    func (cfg *apiConfig) handleEditSqueak(w http.ResponseWriter, r *http.Request) {
    	userID, ok := userIDFrom(r.Context())
    	if !ok {
    		respondWithError(w, http.StatusUnauthorized, "login required")
    		return
    	}
    	id, err := uuid.Parse(r.PathValue("id"))
    	if err != nil {
    		respondWithError(w, http.StatusBadRequest, "invalid squeak ID")
    		return
    	}

    	var params editParams
    	r.Body = http.MaxBytesReader(w, r.Body, maxEditBytes)
    	if err := json.UnmarshalRead(r.Body, &params, json.RejectUnknownMembers(true)); err != nil {
    		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
    			respondWithError(w, http.StatusRequestEntityTooLarge, "request body too large")
    			return
    		}
    		respondWithError(w, http.StatusBadRequest, "invalid JSON")
    		return
    	}
    	if params.Body == nil {
    		respondWithError(w, http.StatusBadRequest, "body is required")
    		return
    	}
    	body := strings.TrimSpace(*params.Body)
    	if body == "" || utf8.RuneCountInString(body) > maxSqueakLen {
    		respondWithError(w, http.StatusUnprocessableEntity, fmt.Sprintf("body must be 1 to %d characters", maxSqueakLen))
    		return
    	}

    	sq, err := cfg.store.GetSqueak(r.Context(), id)
    	if errors.Is(err, ErrNotFound) {
    		respondWithError(w, http.StatusNotFound, "squeak not found")
    		return
    	}
    	if err != nil {
    		respondWithError(w, http.StatusInternalServerError, "couldn't edit squeak")
    		return
    	}
    	if sq.AuthorID != userID {
    		respondWithError(w, http.StatusForbidden, "you can only edit your own squeaks")
    		return
    	}
    	now := cfg.now()
    	if now.Sub(sq.CreatedAt) > editWindow {
    		respondWithError(w, http.StatusConflict, "squeaks can only be edited for 10 minutes")
    		return
    	}

    	sq.Body, sq.EditedAt = body, now
    	if err := cfg.store.UpdateSqueak(r.Context(), sq); err != nil {
    		if errors.Is(err, ErrNotFound) {
    			respondWithError(w, http.StatusNotFound, "squeak not found")
    			return
    		}
    		respondWithError(w, http.StatusInternalServerError, "couldn't edit squeak")
    		return
    	}
    	respondWithJSON(w, http.StatusOK, sq)
    }

    func main() {
    	pip := uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	posted := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
    	sq := Squeak{ID: uuid.NewV7(), AuthorID: pip, Body: "teh cheese", CreatedAt: posted}
    	cfg := &apiConfig{store: NewMemoryStore(sq), now: func() time.Time { return posted.Add(time.Minute) }}

    	mux := http.NewServeMux()
    	mux.HandleFunc("PATCH /api/squeaks/{id}", cfg.handleEditSqueak)
    	req := httptest.NewRequest("PATCH", "/api/squeaks/"+sq.ID.String(), strings.NewReader(`{"body":"the cheese"}`))
    	rec := httptest.NewRecorder()
    	mux.ServeHTTP(rec, withUser(req, pip))
    	fmt.Println(rec.Code, rec.Body.String())
    }
  tests: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"errors"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"testing"
    	"time"
    	"uuid"
    )

    var (
    	pip      = uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	whiskers = uuid.MustParse("0192f1e2-9999-7b4d-9e5f-a1b2c3d4e5f6")
    	posted   = time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
    	pipsID   = uuid.MustParse("0192f1e3-0000-7000-8000-000000000001")
    )

    type env struct {
    	cfg   *apiConfig
    	store *MemoryStore
    	now   time.Time
    }

    func newEnv() *env {
    	e := &env{store: NewMemoryStore(Squeak{ID: pipsID, AuthorID: pip, Body: "teh cheese", CreatedAt: posted}), now: posted.Add(time.Minute)}
    	e.cfg = &apiConfig{store: e.store, now: func() time.Time { return e.now }}
    	return e
    }

    func (e *env) edit(user *uuid.UUID, id, body string) *httptest.ResponseRecorder {
    	mux := http.NewServeMux()
    	mux.HandleFunc("PATCH /api/squeaks/{id}", e.cfg.handleEditSqueak)
    	req := httptest.NewRequest("PATCH", "/api/squeaks/"+id, strings.NewReader(body))
    	if user != nil {
    		req = withUser(req, *user)
    	}
    	rec := httptest.NewRecorder()
    	mux.ServeHTTP(rec, req)
    	return rec
    }

    func checkError(t *testing.T, what string, rec *httptest.ResponseRecorder, want int) {
    	t.Helper()
    	if rec.Code != want {
    		t.Errorf("%s: status %d, want %d (body %s)", what, rec.Code, want, rec.Body.String())
    		return
    	}
    	var e map[string]string
    	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil || e["error"] == "" {
    		t.Errorf("%s: body %s, want a JSON {\"error\":\"...\"}", what, rec.Body.String())
    	}
    }

    func (e *env) bodyNow(t *testing.T) string {
    	t.Helper()
    	sq, _ := e.store.GetSqueak(context.Background(), pipsID)
    	return sq.Body
    }

    func TestEditOwnSqueak(t *testing.T) {
    	e := newEnv()
    	rec := e.edit(&pip, pipsID.String(), `{"body":"  the cheese  "}`)
    	if rec.Code != 200 {
    		t.Fatalf("pip edits her own squeak: status %d %s, want 200", rec.Code, rec.Body.String())
    	}
    	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
    		t.Errorf("Content-Type = %q, want application/json", ct)
    	}
    	var got Squeak
    	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
    		t.Fatalf("response %s isn't a squeak: %v", rec.Body.String(), err)
    	}
    	if got.ID != pipsID || got.AuthorID != pip || got.Body != "the cheese" || !got.CreatedAt.Equal(posted) || !got.EditedAt.Equal(e.now) {
    		t.Errorf("response = %+v, want the squeak with body %q (trimmed), created_at unchanged, edited_at = now (%v)", got, "the cheese", e.now)
    	}
    	if b := e.bodyNow(t); b != "the cheese" {
    		t.Errorf("stored body after edit = %q, want %q", b, "the cheese")
    	}
    }

    func TestEditWindow(t *testing.T) {
    	e := newEnv()
    	e.now = posted.Add(10 * time.Minute)
    	if rec := e.edit(&pip, pipsID.String(), `{"body":"just in time"}`); rec.Code != 200 {
    		t.Errorf("edit exactly 10 minutes after posting: status %d, want 200", rec.Code)
    	}
    	e.now = posted.Add(10*time.Minute + time.Second)
    	checkError(t, "edit 10m1s after posting", e.edit(&pip, pipsID.String(), `{"body":"too late"}`), 409)
    	if b := e.bodyNow(t); b != "just in time" {
    		t.Errorf("stored body after the 10m and 10m1s edits = %q, want %q (the late one must be rejected)", b, "just in time")
    	}
    }

    func TestNotYours(t *testing.T) {
    	e := newEnv()
    	checkError(t, "whiskers edits pip's squeak", e.edit(&whiskers, pipsID.String(), `{"body":"pip smells"}`), 403)
    	checkError(t, "no logged-in user", e.edit(nil, pipsID.String(), `{"body":"hi"}`), 401)
    	if b := e.bodyNow(t); b != "teh cheese" {
    		t.Errorf("a rejected edit changed the squeak to %q", b)
    	}
    }

    func TestBadRequests(t *testing.T) {
    	id := pipsID.String()
    	tests := []struct {
    		name, id, body string
    		want           int
    	}{
    		{"not a UUID", "42", `{"body":"hi"}`, 400},
    		{"unknown squeak", uuid.NewV7().String(), `{"body":"hi"}`, 404},
    		{"broken JSON", id, `{"body":`, 400},
    		{"not an object", id, `"hi"`, 400},
    		{"wrong type", id, `{"body":42}`, 400},
    		{"unknown field", id, `{"body":"hi","author_id":"` + pip.String() + `"}`, 400},
    		{"field name case", id, `{"Body":"hi"}`, 400},
    		{"no body field", id, `{}`, 400},
    		{"trailing data", id, `{"body":"hi"}{"body":"again"}`, 400},
    		{"empty body", id, `{"body":""}`, 422},
    		{"only spaces", id, `{"body":"   \n "}`, 422},
    		{"141 characters", id, `{"body":"` + strings.Repeat("🧀", 141) + `"}`, 422},
    		{"over 1 KiB", id, `{"body":"` + strings.Repeat("x", 2000) + `"}`, 413},
    	}
    	for _, tt := range tests {
    		e := newEnv()
    		checkError(t, tt.name, e.edit(&pip, tt.id, tt.body), tt.want)
    		if b := e.bodyNow(t); b != "teh cheese" {
    			t.Errorf("%s: the rejected edit changed the squeak to %q", tt.name, b)
    		}
    	}
    	e := newEnv()
    	if rec := e.edit(&pip, id, `{"body":"`+strings.Repeat("🧀", 140)+`"}`); rec.Code != 200 {
    		t.Errorf("140 emoji (560 bytes): status %d, want 200; the limit is 140 characters", rec.Code)
    	}
    }

    type brokenStore struct{ getErr, updateErr error }

    func (b brokenStore) GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error) {
    	if b.getErr != nil {
    		return Squeak{}, b.getErr
    	}
    	return Squeak{ID: id, AuthorID: pip, Body: "old", CreatedAt: posted}, nil
    }

    func (b brokenStore) UpdateSqueak(ctx context.Context, sq Squeak) error { return b.updateErr }

    func TestStoreErrorsDontLeak(t *testing.T) {
    	secret := errors.New("dial tcp 10.0.3.7:5432: password authentication failed for user squeak_admin")
    	for _, st := range []brokenStore{{getErr: secret}, {updateErr: secret}} {
    		cfg := &apiConfig{store: st, now: func() time.Time { return posted }}
    		e := &env{cfg: cfg}
    		rec := e.edit(&pip, pipsID.String(), `{"body":"hi"}`)
    		checkError(t, "store fails", rec, 500)
    		if b := rec.Body.String(); strings.Contains(b, "10.0.3.7") || strings.Contains(b, "squeak_admin") || strings.Contains(b, "password") {
    			t.Errorf("500 response %s leaks the store's error text; send a fixed message", b)
    		}
    	}
    	cfg := &apiConfig{store: brokenStore{updateErr: ErrNotFound}, now: func() time.Time { return posted }}
    	checkError(t, "squeak deleted between Get and Update", (&env{cfg: cfg}).edit(&pip, pipsID.String(), `{"body":"hi"}`), 404)
    }
---

Typos happen: "teh cheese". Squeak now lets authors fix a squeak with
`PATCH /api/squeaks/{id}` and a JSON body like `{"body": "the cheese"}`, but only
their own squeaks, and only for **10 minutes** after posting.

Complete `cfg.handleEditSqueak`. The caller's ID comes from `userIDFrom(r.Context())`
(the auth middleware put it there), and the current time comes from `cfg.now()`
so the tests can control the clock. Respond with `respondWithError` for every
failure:

| Situation | Status |
|---|---|
| no user in the context | `401` |
| `id` isn't a UUID | `400` |
| request body over `maxEditBytes` (1 KiB) | `413` |
| not valid JSON, not an object, unknown fields, trailing data, or no `body` field | `400` |
| `body` empty after trimming spaces, or over 140 characters | `422` |
| no such squeak | `404` |
| the squeak belongs to someone else | `403` |
| more than 10 minutes since `CreatedAt` | `409` |
| any other store error | `500` |

On success, save the squeak with its trimmed body and `EditedAt` set to
`cfg.now()`, and respond `200` with the updated squeak as JSON.

## Example

```
PATCH /api/squeaks/0192f1e3-…   (as pip, 1 minute after posting)
{"body": "the cheese"}

200 {"id":"0192f1e3-…","author_id":"…","body":"the cheese",
     "created_at":"2026-09-01T12:00:00Z","edited_at":"2026-09-01T12:01:00Z"}
```

## Constraints

- A failed request must not change the stored squeak.
- A squeak can be edited at exactly 10 minutes, not after.
- A 500 response must not include the store's error text, which can contain
  hostnames and usernames. `UpdateSqueak` can also return `ErrNotFound` if the
  squeak was deleted a moment ago: that's still a `404`.
