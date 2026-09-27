---
title: Table Tests for Routes
quiz:
  - question: Why give every row of a route table a `name` and run it with `t.Run(tt.name, ...)`?
    options:
      - text: '`t.Run` makes the rows run faster'
      - text: Failures are reported per row by name, one failing row doesn't hide the others, and you can run a single row with `-run`
        correct: true
      - text: The testing package requires names
      - text: It makes the tests run in random order
    explanation: |
      Subtests show up as `TestRoutes/whiskers_cannot_delete_pip's_squeak`, keep going
      after a `t.Fatal` in another row, and can be selected with
      `go test -run 'TestRoutes/whiskers'`.
  - question: A route table checks that Pip can delete Pip's squeak, and nothing else about delete. What important case is missing?
    options:
      - text: Deleting with a very long ID
      - text: Someone *other* than the author trying to delete it, which must fail with 403
        correct: true
      - text: Deleting on a Sunday
      - text: Deleting with `GET` instead of `DELETE`
    explanation: |
      Permission bugs hide in the paths you don't test. For every protected route, test
      at least: no token (401), a valid token for the wrong user (403), and the happy path.
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
    	"slices"
    	"strings"
    	"sync"
    	"uuid"
    )

    // ---- storage (no bugs here) ----

    var ErrNotFound = errors.New("not found")

    type Squeak struct {
    	ID       uuid.UUID `json:"id"`
    	AuthorID uuid.UUID `json:"author_id"`
    	Body     string    `json:"body"`
    }

    type MemoryStore struct {
    	mu      sync.RWMutex
    	squeaks []Squeak
    }

    func (s *MemoryStore) Create(authorID uuid.UUID, body string) Squeak {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	sq := Squeak{ID: uuid.NewV7(), AuthorID: authorID, Body: body}
    	s.squeaks = append(s.squeaks, sq)
    	return sq
    }

    func (s *MemoryStore) Get(id uuid.UUID) (Squeak, error) {
    	s.mu.RLock()
    	defer s.mu.RUnlock()
    	for _, sq := range s.squeaks {
    		if sq.ID == id {
    			return sq, nil
    		}
    	}
    	return Squeak{}, ErrNotFound
    }

    func (s *MemoryStore) List() []Squeak {
    	s.mu.RLock()
    	defer s.mu.RUnlock()
    	return slices.Clone(s.squeaks)
    }

    func (s *MemoryStore) Delete(id uuid.UUID) error {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	i := slices.IndexFunc(s.squeaks, func(sq Squeak) bool { return sq.ID == id })
    	if i < 0 {
    		return ErrNotFound
    	}
    	s.squeaks = slices.Delete(s.squeaks, i, i+1)
    	return nil
    }

    // ---- helpers (no bugs here) ----

    func respondWithJSON(w http.ResponseWriter, code int, payload any) {
    	data, _ := json.Marshal(payload)
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

    // ---- the app: there are four bugs somewhere below ----

    type apiConfig struct {
    	squeaks *MemoryStore
    	tokens  map[string]uuid.UUID // bearer token -> user ID (a stand-in for real JWTs)
    }

    func (cfg *apiConfig) requireAuth(next http.Handler) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		token, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
    		userID, ok := cfg.tokens[token]
    		if !ok {
    			respondWithError(w, http.StatusUnauthorized, "invalid or missing token")
    		}
    		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDKey{}, userID)))
    	})
    }

    func (cfg *apiConfig) handleList(w http.ResponseWriter, r *http.Request) {
    	respondWithJSON(w, http.StatusOK, cfg.squeaks.List())
    }

    func (cfg *apiConfig) handleGet(w http.ResponseWriter, r *http.Request) {
    	id, err := uuid.Parse(r.PathValue("id"))
    	if err != nil {
    		respondWithError(w, http.StatusBadRequest, "invalid squeak id")
    		return
    	}
    	sq, err := cfg.squeaks.Get(id)
    	if err != nil {
    		respondWithError(w, http.StatusNotFound, "squeak not found")
    		return
    	}
    	respondWithJSON(w, http.StatusOK, sq)
    }

    func (cfg *apiConfig) handleCreate(w http.ResponseWriter, r *http.Request) {
    	userID, _ := userIDFrom(r.Context())
    	var params struct {
    		Body string `json:"body"`
    	}
    	if err := json.UnmarshalRead(r.Body, &params); err != nil || strings.TrimSpace(params.Body) == "" {
    		respondWithError(w, http.StatusBadRequest, "body is required")
    		return
    	}
    	respondWithJSON(w, http.StatusOK, cfg.squeaks.Create(userID, params.Body))
    }

    func (cfg *apiConfig) handleDelete(w http.ResponseWriter, r *http.Request) {
    	id, err := uuid.Parse(r.PathValue("id"))
    	if err != nil {
    		respondWithError(w, http.StatusBadRequest, "invalid squeak id")
    		return
    	}
    	if _, err := cfg.squeaks.Get(id); err != nil {
    		respondWithError(w, http.StatusNotFound, "squeak not found")
    		return
    	}
    	if err := cfg.squeaks.Delete(id); err != nil {
    		respondWithError(w, http.StatusNotFound, "squeak not found")
    		return
    	}
    	w.WriteHeader(http.StatusNoContent)
    }

    func newRouter(cfg *apiConfig) *http.ServeMux {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/squeaks/", cfg.handleList)
    	mux.HandleFunc("GET /api/squeaks/{id}", cfg.handleGet)
    	mux.Handle("POST /api/squeaks", cfg.requireAuth(http.HandlerFunc(cfg.handleCreate)))
    	mux.Handle("DELETE /api/squeaks/{id}", cfg.requireAuth(http.HandlerFunc(cfg.handleDelete)))
    	return mux
    }

    // ---- a route table, like the one in the lesson ----

    func main() {
    	pip, whiskers := uuid.NewV7(), uuid.NewV7()
    	cfg := &apiConfig{
    		squeaks: &MemoryStore{},
    		tokens:  map[string]uuid.UUID{"pip-token": pip, "whiskers-token": whiskers},
    	}
    	mux := newRouter(cfg)
    	pipSqueak := cfg.squeaks.Create(pip, "pip was here").ID.String()

    	for _, tt := range []struct {
    		name, method, path, token, body string
    		want                            int
    	}{
    		{"list is public", "GET", "/api/squeaks", "", "", 200},
    		{"get one", "GET", "/api/squeaks/" + pipSqueak, "", "", 200},
    		{"get bad id", "GET", "/api/squeaks/nibble", "", "", 400},
    		{"post needs auth", "POST", "/api/squeaks", "", `{"body":"sneaky"}`, 401},
    		{"post as pip", "POST", "/api/squeaks", "pip-token", `{"body":"hi"}`, 201},
    		{"post empty", "POST", "/api/squeaks", "pip-token", `{"body":""}`, 400},
    		{"whiskers can't delete pip's squeak", "DELETE", "/api/squeaks/" + pipSqueak, "whiskers-token", "", 403},
    		{"pip deletes pip's squeak", "DELETE", "/api/squeaks/" + pipSqueak, "pip-token", "", 204},
    	} {
    		req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
    		if tt.token != "" {
    			req.Header.Set("Authorization", "Bearer "+tt.token)
    		}
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, req)
    		mark := "ok "
    		if rec.Code != tt.want {
    			mark = "BUG"
    		}
    		fmt.Printf("%s %-36s got %d, want %d\n", mark, tt.name, rec.Code, tt.want)
    	}
    	for _, sq := range cfg.squeaks.List() {
    		if sq.AuthorID == uuid.Nil() {
    			fmt.Printf("BUG a squeak with no author got stored: %q\n", sq.Body)
    		}
    	}
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
    	"slices"
    	"strings"
    	"sync"
    	"uuid"
    )

    // ---- storage (no bugs here) ----

    var ErrNotFound = errors.New("not found")

    type Squeak struct {
    	ID       uuid.UUID `json:"id"`
    	AuthorID uuid.UUID `json:"author_id"`
    	Body     string    `json:"body"`
    }

    type MemoryStore struct {
    	mu      sync.RWMutex
    	squeaks []Squeak
    }

    func (s *MemoryStore) Create(authorID uuid.UUID, body string) Squeak {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	sq := Squeak{ID: uuid.NewV7(), AuthorID: authorID, Body: body}
    	s.squeaks = append(s.squeaks, sq)
    	return sq
    }

    func (s *MemoryStore) Get(id uuid.UUID) (Squeak, error) {
    	s.mu.RLock()
    	defer s.mu.RUnlock()
    	for _, sq := range s.squeaks {
    		if sq.ID == id {
    			return sq, nil
    		}
    	}
    	return Squeak{}, ErrNotFound
    }

    func (s *MemoryStore) List() []Squeak {
    	s.mu.RLock()
    	defer s.mu.RUnlock()
    	return slices.Clone(s.squeaks)
    }

    func (s *MemoryStore) Delete(id uuid.UUID) error {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	i := slices.IndexFunc(s.squeaks, func(sq Squeak) bool { return sq.ID == id })
    	if i < 0 {
    		return ErrNotFound
    	}
    	s.squeaks = slices.Delete(s.squeaks, i, i+1)
    	return nil
    }

    // ---- helpers (no bugs here) ----

    func respondWithJSON(w http.ResponseWriter, code int, payload any) {
    	data, _ := json.Marshal(payload)
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

    // ---- the app ----

    type apiConfig struct {
    	squeaks *MemoryStore
    	tokens  map[string]uuid.UUID // bearer token -> user ID (a stand-in for real JWTs)
    }

    func (cfg *apiConfig) requireAuth(next http.Handler) http.Handler {
    	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		token, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
    		userID, ok := cfg.tokens[token]
    		if !ok {
    			respondWithError(w, http.StatusUnauthorized, "invalid or missing token")
    			return
    		}
    		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDKey{}, userID)))
    	})
    }

    func (cfg *apiConfig) handleList(w http.ResponseWriter, r *http.Request) {
    	respondWithJSON(w, http.StatusOK, cfg.squeaks.List())
    }

    func (cfg *apiConfig) handleGet(w http.ResponseWriter, r *http.Request) {
    	id, err := uuid.Parse(r.PathValue("id"))
    	if err != nil {
    		respondWithError(w, http.StatusBadRequest, "invalid squeak id")
    		return
    	}
    	sq, err := cfg.squeaks.Get(id)
    	if err != nil {
    		respondWithError(w, http.StatusNotFound, "squeak not found")
    		return
    	}
    	respondWithJSON(w, http.StatusOK, sq)
    }

    func (cfg *apiConfig) handleCreate(w http.ResponseWriter, r *http.Request) {
    	userID, _ := userIDFrom(r.Context())
    	var params struct {
    		Body string `json:"body"`
    	}
    	if err := json.UnmarshalRead(r.Body, &params); err != nil || strings.TrimSpace(params.Body) == "" {
    		respondWithError(w, http.StatusBadRequest, "body is required")
    		return
    	}
    	respondWithJSON(w, http.StatusCreated, cfg.squeaks.Create(userID, params.Body))
    }

    func (cfg *apiConfig) handleDelete(w http.ResponseWriter, r *http.Request) {
    	id, err := uuid.Parse(r.PathValue("id"))
    	if err != nil {
    		respondWithError(w, http.StatusBadRequest, "invalid squeak id")
    		return
    	}
    	sq, err := cfg.squeaks.Get(id)
    	if err != nil {
    		respondWithError(w, http.StatusNotFound, "squeak not found")
    		return
    	}
    	userID, ok := userIDFrom(r.Context())
    	if !ok || sq.AuthorID != userID {
    		respondWithError(w, http.StatusForbidden, "you can only delete your own squeaks")
    		return
    	}
    	if err := cfg.squeaks.Delete(id); err != nil {
    		respondWithError(w, http.StatusNotFound, "squeak not found")
    		return
    	}
    	w.WriteHeader(http.StatusNoContent)
    }

    func newRouter(cfg *apiConfig) *http.ServeMux {
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/squeaks", cfg.handleList)
    	mux.HandleFunc("GET /api/squeaks/{id}", cfg.handleGet)
    	mux.Handle("POST /api/squeaks", cfg.requireAuth(http.HandlerFunc(cfg.handleCreate)))
    	mux.Handle("DELETE /api/squeaks/{id}", cfg.requireAuth(http.HandlerFunc(cfg.handleDelete)))
    	return mux
    }

    // ---- a route table, like the one in the lesson ----

    func main() {
    	pip, whiskers := uuid.NewV7(), uuid.NewV7()
    	cfg := &apiConfig{
    		squeaks: &MemoryStore{},
    		tokens:  map[string]uuid.UUID{"pip-token": pip, "whiskers-token": whiskers},
    	}
    	mux := newRouter(cfg)
    	pipSqueak := cfg.squeaks.Create(pip, "pip was here").ID.String()

    	for _, tt := range []struct {
    		name, method, path, token, body string
    		want                            int
    	}{
    		{"list is public", "GET", "/api/squeaks", "", "", 200},
    		{"get one", "GET", "/api/squeaks/" + pipSqueak, "", "", 200},
    		{"get bad id", "GET", "/api/squeaks/nibble", "", "", 400},
    		{"post needs auth", "POST", "/api/squeaks", "", `{"body":"sneaky"}`, 401},
    		{"post as pip", "POST", "/api/squeaks", "pip-token", `{"body":"hi"}`, 201},
    		{"post empty", "POST", "/api/squeaks", "pip-token", `{"body":""}`, 400},
    		{"whiskers can't delete pip's squeak", "DELETE", "/api/squeaks/" + pipSqueak, "whiskers-token", "", 403},
    		{"pip deletes pip's squeak", "DELETE", "/api/squeaks/" + pipSqueak, "pip-token", "", 204},
    	} {
    		req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
    		if tt.token != "" {
    			req.Header.Set("Authorization", "Bearer "+tt.token)
    		}
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, req)
    		mark := "ok "
    		if rec.Code != tt.want {
    			mark = "BUG"
    		}
    		fmt.Printf("%s %-36s got %d, want %d\n", mark, tt.name, rec.Code, tt.want)
    	}
    	for _, sq := range cfg.squeaks.List() {
    		if sq.AuthorID == uuid.Nil() {
    			fmt.Printf("BUG a squeak with no author got stored: %q\n", sq.Body)
    		}
    	}
    }
  tests: |
    package main

    import (
    	"net/http/httptest"
    	"strings"
    	"testing"
    	"uuid"
    )

    type testApp struct {
    	cfg           *apiConfig
    	pip, whiskers uuid.UUID
    }

    func newTestApp() *testApp {
    	pip, whiskers := uuid.NewV7(), uuid.NewV7()
    	return &testApp{
    		cfg: &apiConfig{
    			squeaks: &MemoryStore{},
    			tokens:  map[string]uuid.UUID{"pip-token": pip, "whiskers-token": whiskers},
    		},
    		pip:      pip,
    		whiskers: whiskers,
    	}
    }

    func TestRoutes(t *testing.T) {
    	for _, tt := range []struct {
    		name, method, path, token, body string
    		wantCode                        int
    	}{
    		{"list is public", "GET", "/api/squeaks", "", "", 200},
    		{"get one", "GET", "/api/squeaks/{pip}", "", "", 200},
    		{"get missing", "GET", "/api/squeaks/" + uuid.NewV7().String(), "", "", 404},
    		{"get bad id", "GET", "/api/squeaks/nibble", "", "", 400},
    		{"post needs auth", "POST", "/api/squeaks", "", `{"body":"sneaky"}`, 401},
    		{"post with junk token", "POST", "/api/squeaks", "junk", `{"body":"sneaky"}`, 401},
    		{"post as pip", "POST", "/api/squeaks", "pip-token", `{"body":"hi"}`, 201},
    		{"post empty", "POST", "/api/squeaks", "pip-token", `{"body":"  "}`, 400},
    		{"delete needs auth", "DELETE", "/api/squeaks/{pip}", "", "", 401},
    		{"whiskers can't delete pip's squeak", "DELETE", "/api/squeaks/{pip}", "whiskers-token", "", 403},
    		{"pip deletes pip's squeak", "DELETE", "/api/squeaks/{pip}", "pip-token", "", 204},
    		{"delete missing", "DELETE", "/api/squeaks/" + uuid.NewV7().String(), "pip-token", "", 404},
    		{"wrong method", "PUT", "/api/squeaks", "pip-token", "", 405},
    	} {
    		t.Run(tt.name, func(t *testing.T) {
    			app := newTestApp() // fresh state for every row
    			pipSqueak := app.cfg.squeaks.Create(app.pip, "pip was here")
    			path := strings.Replace(tt.path, "{pip}", pipSqueak.ID.String(), 1)

    			req := httptest.NewRequest(tt.method, path, strings.NewReader(tt.body))
    			if tt.token != "" {
    				req.Header.Set("Authorization", "Bearer "+tt.token)
    			}
    			rec := httptest.NewRecorder()
    			newRouter(app.cfg).ServeHTTP(rec, req)

    			if rec.Code != tt.wantCode {
    				t.Errorf("%s %s: status = %d, want %d; body: %s", tt.method, tt.path, rec.Code, tt.wantCode, rec.Body)
    			}
    			after := app.cfg.squeaks.List()
    			for _, sq := range after {
    				if sq.AuthorID == uuid.Nil() {
    					t.Errorf("%s %s stored a squeak with no author (%q); rejected requests must not reach the handler", tt.method, tt.path, sq.Body)
    				}
    			}
    			if tt.wantCode >= 400 && len(after) != 1 {
    				t.Errorf("%s %s was rejected, but the store went from 1 squeak to %d", tt.method, tt.path, len(after))
    			}
    		})
    	}
    }
---

Squeak now has a dozen routes, several middleware and three kinds of caller (anonymous,
a regular mouse, an admin). Writing one test function per combination would be
exhausting. A **table test** over the whole router covers a lot of ground in very little
code, and it reads like a spec of your API.

## The table

Each row is one request and what should come back:

```go
func TestRoutes(t *testing.T) {
	app := newTestApp(t) // builds cfg, the real router, and a couple of users
	pipSqueak := app.createSqueak(t, app.pip, "pip was here")

	tests := []struct {
		name     string
		method   string
		path     string
		token    string // "" means no Authorization header
		body     string
		wantCode int
	}{
		{"health check", "GET", "/api/healthz", "", "", 200},
		{"list is public", "GET", "/api/squeaks", "", "", 200},
		{"get one", "GET", "/api/squeaks/" + pipSqueak.ID.String(), "", "", 200},
		{"get bad id", "GET", "/api/squeaks/not-a-uuid", "", "", 400},
		{"get missing", "GET", "/api/squeaks/" + uuid.NewV7().String(), "", "", 404},

		{"post needs auth", "POST", "/api/squeaks", "", `{"body":"hi"}`, 401},
		{"post with junk token", "POST", "/api/squeaks", "junk", `{"body":"hi"}`, 401},
		{"post as pip", "POST", "/api/squeaks", app.pipToken, `{"body":"hi"}`, 201},
		{"post empty", "POST", "/api/squeaks", app.pipToken, `{"body":""}`, 400},

		{"whiskers can't delete pip's squeak", "DELETE", "/api/squeaks/" + pipSqueak.ID.String(), app.whiskersToken, "", 403},
		{"metrics are admin only", "GET", "/admin/metrics", app.pipToken, "", 403},
		{"wrong method", "PUT", "/api/squeaks", app.pipToken, "", 405},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			rec := httptest.NewRecorder()
			app.handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("%s %s: status = %d, want %d; body: %s",
					tt.method, tt.path, rec.Code, tt.wantCode, rec.Body)
			}
		})
	}
}
```

Read the table from top to bottom and it's a description of Squeak's access rules.
When someone asks "can anyone post without logging in?", the answer is on one line.

## Order and independence

Rows that *change* state are a trap. If "pip deletes pip's squeak" runs before "get one",
the second row fails, and if someone reorders the table, different rows fail. Some
options:

- Keep the route table to requests that **don't destroy** shared fixtures, and put
  multi-step flows ("create, then delete, then get gives 404") in their own tests.
- Or build fresh state **inside** each subtest, so rows can't affect each other.

Either way, avoid `t.Parallel()` in rows that share a store unless you've thought
about it. The store is concurrency-safe, but the *assertions* may depend on order.

## The test app helper

The helper that builds everything is worth investing in, because every server test uses
it:

```go
type testApp struct {
	cfg                     *apiConfig
	handler                 http.Handler
	pip, whiskers           User
	pipToken, whiskersToken string
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()
	cfg := &apiConfig{
		squeaks:   NewMemoryStore(),
		users:     NewMemoryUserStore(),
		jwtSecret: []byte("test-secret"),
	}
	app := &testApp{cfg: cfg, handler: newRouter(cfg)}
	app.pip = app.createUser(t, "pip@squeak.dev")
	app.whiskers = app.createUser(t, "whiskers@squeak.dev")
	app.pipToken = app.tokenFor(t, app.pip)
	app.whiskersToken = app.tokenFor(t, app.whiskers)
	return app
}
```

Two important details:

1. **`handler: newRouter(cfg)`** is the same function `main` calls, with the same
   middleware. Tests exercise what production runs.
2. **The config is built from scratch** with an in-memory store and a known secret. No
   global state leaks between tests, which is one more payoff from the storage interface
   and from keeping dependencies in `apiConfig`.

## Checking more than the status

The status code catches most regressions, but a table can carry more expectations:
a `wantBody` substring, a `wantHeader` map, or a `check func(t *testing.T, rec *httptest.ResponseRecorder)`
field for the odd row that needs custom assertions. Resist cramming everything in, though.
When a row needs three custom fields, it probably wants its own test function.

## Negative cases are the point

It's natural to test that things work. For a server, the valuable tests are the ones
that check things **fail correctly**: no token, a forged token, somebody else's token,
malformed JSON, the wrong method, a missing ID. Every protected route deserves at least
three rows: no token (401), wrong user (403) and the owner (2xx).

## Your task

The starter is a stripped-down Squeak with a route table in `main`, and **four bugs**
in the app code (the store and helpers are fine). **Run** it: every `BUG` line is a row
that doesn't behave as the table says.

Fix the app so that every row passes. The hidden tests run a bigger table, building a
fresh app for each row, and they also check that rejected requests don't change the
store.

Two hints:

- One `BUG` line in `main`'s table is a knock-on effect of another. Its table shares
  one store across rows, which is exactly the order trap described above. Fix the real
  bugs and it goes away.
- A 307 on a `GET` is the mux talking, not your handler. Look at the patterns.
