---
title: Owner-Only Actions
quiz:
  - question: |
      Between `GetSqueak` (ownership checked) and `DeleteSqueak`, another request deletes
      the same squeak. What does a well-written handler return?
    options:
      - text: '`204`, because the ownership check passed'
      - text: '`404`, because `DeleteSqueak` reports `ErrNotFound`, and the handler checks its error too'
        correct: true
      - text: '`500`, because the squeak vanished'
      - text: '`403`'
    explanation: |
      Check-then-act has a gap. The handler must still handle `DeleteSqueak`'s error, since
      "not found" is a perfectly normal outcome of a race. A store method like
      `DeleteSqueakByAuthor(ctx, id, authorID)` that checks and deletes under one lock closes
      the gap entirely.
exercise:
  starter: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"log"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"sync"
    	"time"
    	"uuid"
    )

    var ErrNotFound = errors.New("not found")

    type Squeak struct {
    	ID        uuid.UUID `json:"id"`
    	AuthorID  uuid.UUID `json:"author_id"`
    	Body      string    `json:"body"`
    	CreatedAt time.Time `json:"created_at"`
    }

    type SqueakStore interface {
    	CreateSqueak(ctx context.Context, authorID uuid.UUID, body string) (Squeak, error)
    	GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error)
    	DeleteSqueak(ctx context.Context, id uuid.UUID) error
    }

    type MemoryStore struct {
    	mu      sync.RWMutex
    	squeaks map[uuid.UUID]Squeak
    }

    func NewMemoryStore() *MemoryStore {
    	return &MemoryStore{squeaks: make(map[uuid.UUID]Squeak)}
    }

    func (s *MemoryStore) CreateSqueak(ctx context.Context, authorID uuid.UUID, body string) (Squeak, error) {
    	sq := Squeak{ID: uuid.NewV7(), AuthorID: authorID, Body: body, CreatedAt: time.Now().UTC()}
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	s.squeaks[sq.ID] = sq
    	return sq, nil
    }

    func (s *MemoryStore) GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error) {
    	s.mu.RLock()
    	defer s.mu.RUnlock()
    	sq, ok := s.squeaks[id]
    	if !ok {
    		return Squeak{}, ErrNotFound
    	}
    	return sq, nil
    }

    func (s *MemoryStore) DeleteSqueak(ctx context.Context, id uuid.UUID) error {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	if _, ok := s.squeaks[id]; !ok {
    		return ErrNotFound
    	}
    	delete(s.squeaks, id)
    	return nil
    }

    type apiConfig struct {
    	squeaks SqueakStore
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

    type userIDKey struct{}

    func userIDFrom(ctx context.Context) (uuid.UUID, bool) {
    	id, ok := ctx.Value(userIDKey{}).(uuid.UUID)
    	return id, ok
    }

    // demoAuth stands in for requireAuth. To keep the exercise short, the
    // "token" is simply the user's ID. Never do this in a real app!
    func demoAuth(next http.Handler) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		id, err := uuid.Parse(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
    		if err != nil {
    			respondWithError(w, http.StatusUnauthorized, "missing or invalid token")
    			return
    		}
    		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDKey{}, id)))
    	})
    }

    // handleDeleteSqueak handles DELETE /api/squeaks/{id}. Only the author may delete.
    func (cfg *apiConfig) handleDeleteSqueak(w http.ResponseWriter, r *http.Request) {
    	id, _ := uuid.Parse(r.PathValue("id"))
    	cfg.squeaks.DeleteSqueak(r.Context(), id)
    	w.WriteHeader(http.StatusNoContent)
    	// ?
    }

    func newRouter(cfg *apiConfig) *http.ServeMux {
    	mux := http.NewServeMux()
    	mux.Handle("DELETE /api/squeaks/{id}", demoAuth(http.HandlerFunc(cfg.handleDeleteSqueak)))
    	return mux
    }

    func main() {
    	store := NewMemoryStore()
    	cfg := &apiConfig{squeaks: store}
    	mux := newRouter(cfg)

    	pip, whiskers := uuid.NewV7(), uuid.NewV7()
    	sq, _ := store.CreateSqueak(context.Background(), pip, "my cheese, my rules")

    	del := func(who string, user uuid.UUID) {
    		req := httptest.NewRequest("DELETE", "/api/squeaks/"+sq.ID.String(), nil)
    		req.Header.Set("Authorization", "Bearer "+user.String())
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, req)
    		fmt.Printf("%-8s deletes pip's squeak -> %d %s\n", who, rec.Code, rec.Body.String())
    	}
    	del("whiskers", whiskers)
    	del("pip", pip)
    	del("pip", pip)
    	_ = log.Printf
    }
  solution: |
    package main

    import (
    	"context"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"log"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"sync"
    	"time"
    	"uuid"
    )

    var ErrNotFound = errors.New("not found")

    type Squeak struct {
    	ID        uuid.UUID `json:"id"`
    	AuthorID  uuid.UUID `json:"author_id"`
    	Body      string    `json:"body"`
    	CreatedAt time.Time `json:"created_at"`
    }

    type SqueakStore interface {
    	CreateSqueak(ctx context.Context, authorID uuid.UUID, body string) (Squeak, error)
    	GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error)
    	DeleteSqueak(ctx context.Context, id uuid.UUID) error
    }

    type MemoryStore struct {
    	mu      sync.RWMutex
    	squeaks map[uuid.UUID]Squeak
    }

    func NewMemoryStore() *MemoryStore {
    	return &MemoryStore{squeaks: make(map[uuid.UUID]Squeak)}
    }

    func (s *MemoryStore) CreateSqueak(ctx context.Context, authorID uuid.UUID, body string) (Squeak, error) {
    	sq := Squeak{ID: uuid.NewV7(), AuthorID: authorID, Body: body, CreatedAt: time.Now().UTC()}
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	s.squeaks[sq.ID] = sq
    	return sq, nil
    }

    func (s *MemoryStore) GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error) {
    	s.mu.RLock()
    	defer s.mu.RUnlock()
    	sq, ok := s.squeaks[id]
    	if !ok {
    		return Squeak{}, ErrNotFound
    	}
    	return sq, nil
    }

    func (s *MemoryStore) DeleteSqueak(ctx context.Context, id uuid.UUID) error {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	if _, ok := s.squeaks[id]; !ok {
    		return ErrNotFound
    	}
    	delete(s.squeaks, id)
    	return nil
    }

    type apiConfig struct {
    	squeaks SqueakStore
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

    type userIDKey struct{}

    func userIDFrom(ctx context.Context) (uuid.UUID, bool) {
    	id, ok := ctx.Value(userIDKey{}).(uuid.UUID)
    	return id, ok
    }

    func demoAuth(next http.Handler) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		id, err := uuid.Parse(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
    		if err != nil {
    			respondWithError(w, http.StatusUnauthorized, "missing or invalid token")
    			return
    		}
    		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDKey{}, id)))
    	})
    }

    func (cfg *apiConfig) handleDeleteSqueak(w http.ResponseWriter, r *http.Request) {
    	userID, ok := userIDFrom(r.Context())
    	if !ok {
    		respondWithError(w, http.StatusUnauthorized, "not logged in")
    		return
    	}
    	id, err := uuid.Parse(r.PathValue("id"))
    	if err != nil {
    		respondWithError(w, http.StatusBadRequest, "invalid squeak id")
    		return
    	}

    	squeak, err := cfg.squeaks.GetSqueak(r.Context(), id)
    	if errors.Is(err, ErrNotFound) {
    		respondWithError(w, http.StatusNotFound, "squeak not found")
    		return
    	}
    	if err != nil {
    		log.Printf("getting squeak %s: %v", id, err)
    		respondWithError(w, http.StatusInternalServerError, "couldn't delete squeak")
    		return
    	}
    	if squeak.AuthorID != userID {
    		respondWithError(w, http.StatusForbidden, "you can only delete your own squeaks")
    		return
    	}

    	err = cfg.squeaks.DeleteSqueak(r.Context(), id)
    	if errors.Is(err, ErrNotFound) {
    		respondWithError(w, http.StatusNotFound, "squeak not found")
    		return
    	}
    	if err != nil {
    		log.Printf("deleting squeak %s: %v", id, err)
    		respondWithError(w, http.StatusInternalServerError, "couldn't delete squeak")
    		return
    	}
    	w.WriteHeader(http.StatusNoContent)
    }

    func newRouter(cfg *apiConfig) *http.ServeMux {
    	mux := http.NewServeMux()
    	mux.Handle("DELETE /api/squeaks/{id}", demoAuth(http.HandlerFunc(cfg.handleDeleteSqueak)))
    	return mux
    }

    func main() {
    	store := NewMemoryStore()
    	cfg := &apiConfig{squeaks: store}
    	mux := newRouter(cfg)

    	pip, whiskers := uuid.NewV7(), uuid.NewV7()
    	sq, _ := store.CreateSqueak(context.Background(), pip, "my cheese, my rules")

    	del := func(who string, user uuid.UUID) {
    		req := httptest.NewRequest("DELETE", "/api/squeaks/"+sq.ID.String(), nil)
    		req.Header.Set("Authorization", "Bearer "+user.String())
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, req)
    		fmt.Printf("%-8s deletes pip's squeak -> %d %s\n", who, rec.Code, rec.Body.String())
    	}
    	del("whiskers", whiskers)
    	del("pip", pip)
    	del("pip", pip)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"context"
    	"encoding/json/v2"
    	"errors"
    	"log"
    	"net/http"
    	"net/http/httptest"
    	"os"
    	"strings"
    	"testing"
    	"uuid"
    )

    func deleteAs(mux http.Handler, user uuid.UUID, id string) *httptest.ResponseRecorder {
    	req := httptest.NewRequest("DELETE", "/api/squeaks/"+id, nil)
    	if user != uuid.Nil() {
    		req.Header.Set("Authorization", "Bearer "+user.String())
    	}
    	rec := httptest.NewRecorder()
    	mux.ServeHTTP(rec, req)
    	return rec
    }

    func checkError(t *testing.T, what string, rec *httptest.ResponseRecorder, wantCode int) {
    	t.Helper()
    	if rec.Code != wantCode {
    		t.Errorf("%s: status = %d, want %d (body %q)", what, rec.Code, wantCode, rec.Body.String())
    		return
    	}
    	var body map[string]string
    	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["error"] == "" {
    		t.Errorf("%s: body = %q, want a single JSON object like {\"error\":\"...\"}", what, rec.Body.String())
    	}
    }

    func TestOwnerCanDelete(t *testing.T) {
    	store := NewMemoryStore()
    	mux := newRouter(&apiConfig{squeaks: store})
    	pip := uuid.NewV7()
    	sq, _ := store.CreateSqueak(t.Context(), pip, "mine")

    	rec := deleteAs(mux, pip, sq.ID.String())
    	if rec.Code != http.StatusNoContent {
    		t.Fatalf("author deleting own squeak: status = %d, want 204 (body %q)", rec.Code, rec.Body.String())
    	}
    	if rec.Body.Len() != 0 {
    		t.Errorf("204 response has a body: %q", rec.Body.String())
    	}
    	if _, err := store.GetSqueak(t.Context(), sq.ID); !errors.Is(err, ErrNotFound) {
    		t.Error("the squeak still exists after a 204")
    	}
    	checkError(t, "deleting the same squeak twice", deleteAs(mux, pip, sq.ID.String()), http.StatusNotFound)
    }

    func TestOthersCannotDelete(t *testing.T) {
    	store := NewMemoryStore()
    	mux := newRouter(&apiConfig{squeaks: store})
    	pip, whiskers := uuid.NewV7(), uuid.NewV7()
    	sq, _ := store.CreateSqueak(t.Context(), pip, "mine")

    	checkError(t, "whiskers deleting pip's squeak", deleteAs(mux, whiskers, sq.ID.String()), http.StatusForbidden)
    	if _, err := store.GetSqueak(t.Context(), sq.ID); err != nil {
    		t.Error("whiskers managed to delete pip's squeak!")
    	}
    }

    func TestBadRequests(t *testing.T) {
    	store := NewMemoryStore()
    	mux := newRouter(&apiConfig{squeaks: store})
    	pip := uuid.NewV7()
    	checkError(t, "invalid squeak id", deleteAs(mux, pip, "not-a-uuid"), http.StatusBadRequest)
    	checkError(t, "unknown squeak id", deleteAs(mux, pip, uuid.NewV7().String()), http.StatusNotFound)
    	checkError(t, "no token", deleteAs(mux, uuid.Nil(), uuid.NewV7().String()), http.StatusUnauthorized)
    }

    type brokenStore struct{ SqueakStore }

    func (brokenStore) GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error) {
    	return Squeak{}, errors.New("disk on fire at /var/lib/squeak")
    }

    func TestStoreFailure(t *testing.T) {
    	var logs bytes.Buffer
    	log.SetOutput(&logs)
    	t.Cleanup(func() { log.SetOutput(os.Stderr) })

    	mux := newRouter(&apiConfig{squeaks: brokenStore{}})
    	rec := deleteAs(mux, uuid.NewV7(), uuid.NewV7().String())
    	checkError(t, "store failure", rec, http.StatusInternalServerError)
    	if strings.Contains(rec.Body.String(), "disk on fire") {
    		t.Errorf("the 500 response leaks the internal error: %q", rec.Body.String())
    	}
    }

    func TestHandlerWithoutAuth(t *testing.T) {
    	cfg := &apiConfig{squeaks: NewMemoryStore()}
    	req := httptest.NewRequest("DELETE", "/api/squeaks/x", nil)
    	req.SetPathValue("id", uuid.NewV7().String())
    	rec := httptest.NewRecorder()
    	cfg.handleDeleteSqueak(rec, req)
    	if rec.Code != http.StatusUnauthorized {
    		t.Errorf("handler called without auth middleware: status = %d, want 401", rec.Code)
    	}
    }
---

Time to close the IDOR hole from the last lesson. Only a squeak's author may delete it.

The starter includes the squeak store, the JSON helpers and a `demoAuth` middleware.
To keep things short, `demoAuth` treats the bearer token *as* the user ID. That's
wildly insecure, but it lets the tests act as any user. In the real Squeak it's the
`requireAuth` you wrote in the last chapter.

## Your task

Rewrite `handleDeleteSqueak` for `DELETE /api/squeaks/{id}`:

1. Get the caller's ID with `userIDFrom`. If there isn't one, respond **401**. (The
   middleware should always set it, but a handler registered without the middleware must
   fail closed.)
2. Parse the `id` path value with `uuid.Parse`. Invalid: **400**.
3. Load the squeak with `GetSqueak`. `ErrNotFound`: **404**. Any other error: log it and
   respond **500** with a generic message. Don't include the error text in the response.
4. If the squeak's `AuthorID` isn't the caller: **403**.
5. Delete it. Handle `DeleteSqueak`'s error the same way as in step 3 (it can still say
   `ErrNotFound` if another request got there first).
6. Success: **204** with no body.

Every error response must be JSON with an `error` field, sent with `respondWithError`.

**Run** shows the problem: with the starter, Whiskers happily deletes Pip's squeak.

The `_ = log.Printf` line at the end of `main` only keeps the `log` import compiling.
Delete it once your handler logs errors.
