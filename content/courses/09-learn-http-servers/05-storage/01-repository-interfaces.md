---
title: Repository Interfaces
quiz:
  - question: Why do Squeak's handlers depend on a `SqueakStore` interface instead of a concrete `*MemoryStore`?
    options:
      - text: Interfaces are faster than structs
      - text: So the storage can be swapped (in-memory for tests, a real database in production) without changing the handlers
        correct: true
      - text: Go doesn't allow structs as struct fields
      - text: Because the interface stores the data for you
    explanation: |
      Handlers only need *some* way to create, get, list and delete squeaks. Depending on
      the behaviour rather than one implementation means a test fake, the in-memory store
      and a Postgres-backed store all plug in unchanged.
  - question: |
      `GetSqueak` can't find the ID. Which return is the most useful to the handler?
    options:
      - text: '`Squeak{}, nil`'
      - text: '`Squeak{}, errors.New("not found")` created fresh inside the method'
      - text: '`Squeak{}, ErrNotFound` where `ErrNotFound` is a package-level sentinel'
        correct: true
      - text: It should panic
    explanation: |
      A sentinel lets the handler write `errors.Is(err, ErrNotFound)` and answer 404,
      while any *other* error becomes a 500. Returning a zero value with `nil` makes
      "missing" look like a real, empty squeak.
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

    // SqueakStore stores squeaks. Implementations must be safe for concurrent use.
    type SqueakStore interface {
    	CreateSqueak(ctx context.Context, authorID uuid.UUID, body string) (Squeak, error)
    	GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error)
    	ListSqueaks(ctx context.Context, authorID uuid.UUID) ([]Squeak, error)
    	DeleteSqueak(ctx context.Context, id uuid.UUID) error
    }

    // fakeStore is a tiny SqueakStore for this exercise. It only implements
    // GetSqueak; the embedded interface stands in for the other methods.
    type fakeStore struct {
    	SqueakStore
    	squeaks map[uuid.UUID]Squeak
    }

    func (f fakeStore) GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error) {
    	sq, ok := f.squeaks[id]
    	if !ok {
    		return Squeak{}, ErrNotFound
    	}
    	return sq, nil
    }

    type apiConfig struct {
    	squeaks SqueakStore
    }

    func respondWithJSON(w http.ResponseWriter, code int, payload any) {
    	data, err := json.Marshal(payload)
    	if err != nil {
    		log.Printf("encoding response: %v", err)
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

    // handleGetSqueak handles GET /api/squeaks/{id}. See the lesson for the rules.
    func (cfg *apiConfig) handleGetSqueak(w http.ResponseWriter, r *http.Request) {
    	// ?
    	respondWithJSON(w, http.StatusOK, Squeak{})
    }

    func main() {
    	log.SetFlags(0)
    	pip := uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	sq := Squeak{
    		ID:        uuid.MustParse("0192f1e3-0000-7000-8000-000000000001"),
    		AuthorID:  pip,
    		Body:      "first squeak!",
    		CreatedAt: time.Date(2026, 9, 1, 9, 30, 0, 0, time.UTC),
    	}
    	cfg := &apiConfig{squeaks: fakeStore{squeaks: map[uuid.UUID]Squeak{sq.ID: sq}}}
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/squeaks/{id}", cfg.handleGetSqueak)

    	for _, id := range []string{sq.ID.String(), "not-a-uuid", uuid.NewV7().String()} {
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/squeaks/"+id, nil))
    		fmt.Println(rec.Code, rec.Body.String())
    	}
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

    // SqueakStore stores squeaks. Implementations must be safe for concurrent use.
    type SqueakStore interface {
    	CreateSqueak(ctx context.Context, authorID uuid.UUID, body string) (Squeak, error)
    	GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error)
    	ListSqueaks(ctx context.Context, authorID uuid.UUID) ([]Squeak, error)
    	DeleteSqueak(ctx context.Context, id uuid.UUID) error
    }

    // fakeStore is a tiny SqueakStore for this exercise. It only implements
    // GetSqueak; the embedded interface stands in for the other methods.
    type fakeStore struct {
    	SqueakStore
    	squeaks map[uuid.UUID]Squeak
    }

    func (f fakeStore) GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error) {
    	sq, ok := f.squeaks[id]
    	if !ok {
    		return Squeak{}, ErrNotFound
    	}
    	return sq, nil
    }

    type apiConfig struct {
    	squeaks SqueakStore
    }

    func respondWithJSON(w http.ResponseWriter, code int, payload any) {
    	data, err := json.Marshal(payload)
    	if err != nil {
    		log.Printf("encoding response: %v", err)
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

    func (cfg *apiConfig) handleGetSqueak(w http.ResponseWriter, r *http.Request) {
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
    		respondWithError(w, http.StatusInternalServerError, "couldn't get squeak")
    		return
    	}
    	respondWithJSON(w, http.StatusOK, squeak)
    }

    func main() {
    	log.SetFlags(0)
    	pip := uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6")
    	sq := Squeak{
    		ID:        uuid.MustParse("0192f1e3-0000-7000-8000-000000000001"),
    		AuthorID:  pip,
    		Body:      "first squeak!",
    		CreatedAt: time.Date(2026, 9, 1, 9, 30, 0, 0, time.UTC),
    	}
    	cfg := &apiConfig{squeaks: fakeStore{squeaks: map[uuid.UUID]Squeak{sq.ID: sq}}}
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/squeaks/{id}", cfg.handleGetSqueak)

    	for _, id := range []string{sq.ID.String(), "not-a-uuid", uuid.NewV7().String()} {
    		rec := httptest.NewRecorder()
    		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/squeaks/"+id, nil))
    		fmt.Println(rec.Code, rec.Body.String())
    	}
    }
  tests: |
    package main

    import (
    	"bytes"
    	"context"
    	"encoding/json/v2"
    	"errors"
    	"fmt"
    	"log"
    	"net/http"
    	"net/http/httptest"
    	"os"
    	"strings"
    	"testing"
    	"time"
    	"uuid"
    )

    var testSqueak = Squeak{
    	ID:        uuid.MustParse("0192f1e3-0000-7000-8000-00000000abcd"),
    	AuthorID:  uuid.MustParse("0192f1e2-8c3a-7b4d-9e5f-a1b2c3d4e5f6"),
    	Body:      "cheese at noon",
    	CreatedAt: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
    }

    type brokenStore struct{ SqueakStore }

    func (brokenStore) GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error) {
    	return Squeak{}, errors.New("disk on fire at /var/lib/squeak")
    }

    // wrappedStore returns ErrNotFound wrapped with extra context.
    type wrappedStore struct{ SqueakStore }

    func (wrappedStore) GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error) {
    	return Squeak{}, fmt.Errorf("squeak %s: %w", id, ErrNotFound)
    }

    func get(store SqueakStore, id string) *httptest.ResponseRecorder {
    	cfg := &apiConfig{squeaks: store}
    	mux := http.NewServeMux()
    	mux.HandleFunc("GET /api/squeaks/{id}", cfg.handleGetSqueak)
    	rec := httptest.NewRecorder()
    	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/squeaks/"+id, nil))
    	return rec
    }

    func errorOf(t *testing.T, rec *httptest.ResponseRecorder) string {
    	t.Helper()
    	var body struct {
    		Error string `json:"error"`
    	}
    	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Error == "" {
    		t.Fatalf("status %d body %q: want a JSON body like {\"error\":\"...\"}", rec.Code, rec.Body.String())
    	}
    	return body.Error
    }

    func TestFound(t *testing.T) {
    	store := fakeStore{squeaks: map[uuid.UUID]Squeak{testSqueak.ID: testSqueak}}
    	for _, id := range []string{testSqueak.ID.String(), strings.ToUpper(testSqueak.ID.String())} {
    		rec := get(store, id)
    		if rec.Code != http.StatusOK {
    			t.Fatalf("GET /api/squeaks/%s: status = %d, want 200 (body %q)", id, rec.Code, rec.Body.String())
    		}
    		var got Squeak
    		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
    			t.Fatalf("response %q is not a squeak: %v", rec.Body.String(), err)
    		}
    		if got.ID != testSqueak.ID || got.Body != testSqueak.Body || !got.CreatedAt.Equal(testSqueak.CreatedAt) {
    			t.Errorf("GET /api/squeaks/%s returned %+v, want %+v", id, got, testSqueak)
    		}
    	}
    }

    func TestBadID(t *testing.T) {
    	for _, id := range []string{"42", "not-a-uuid", "0192f1e3-0000-7000-8000"} {
    		rec := get(fakeStore{}, id)
    		if rec.Code != http.StatusBadRequest {
    			t.Errorf("GET /api/squeaks/%s: status = %d, want 400", id, rec.Code)
    			continue
    		}
    		errorOf(t, rec)
    	}
    }

    func TestNotFound(t *testing.T) {
    	for name, store := range map[string]SqueakStore{
    		"missing squeak":      fakeStore{squeaks: map[uuid.UUID]Squeak{}},
    		"wrapped ErrNotFound": wrappedStore{},
    	} {
    		rec := get(store, uuid.NewV7().String())
    		if rec.Code != http.StatusNotFound {
    			t.Errorf("%s: status = %d, want 404 (use errors.Is)", name, rec.Code)
    			continue
    		}
    		errorOf(t, rec)
    	}
    }

    func TestStoreFailure(t *testing.T) {
    	var logs bytes.Buffer
    	log.SetOutput(&logs)
    	t.Cleanup(func() { log.SetOutput(os.Stderr) })

    	rec := get(brokenStore{}, uuid.NewV7().String())
    	if rec.Code != http.StatusInternalServerError {
    		t.Fatalf("store failure: status = %d, want 500", rec.Code)
    	}
    	if msg := errorOf(t, rec); strings.Contains(msg, "disk on fire") {
    		t.Errorf("500 response %q leaks the internal error; send a generic message", msg)
    	}
    	if !strings.Contains(logs.String(), "disk on fire") {
    		t.Errorf("log output = %q, want the store's error logged", logs.String())
    	}
    }
---

So far Squeak has forgotten everything the moment a request finished. Time for storage.

Squeak's storage will live **in memory**: a map protected by a mutex. It's fast, it
needs no setup, and it loses everything on restart. In chapter 9 you'll swap
in a real SQL database. The trick that makes that swap painless is the subject of
this lesson: put storage behind an **interface**.

## The repository pattern

A *repository* is a type whose only job is storing and fetching one kind of thing. The
rest of the app talks to it in domain terms ("create a squeak") rather than storage
terms ("INSERT INTO", "lock the map"):

```go
// Squeak is a stored squeak.
type Squeak struct {
	ID        uuid.UUID `json:"id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

var ErrNotFound = errors.New("not found")

// SqueakStore stores squeaks. Implementations must be safe for concurrent use.
type SqueakStore interface {
	CreateSqueak(ctx context.Context, authorID uuid.UUID, body string) (Squeak, error)
	GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error)
	ListSqueaks(ctx context.Context, authorID uuid.UUID) ([]Squeak, error)
	DeleteSqueak(ctx context.Context, id uuid.UUID) error
}
```

(`uuid.UUID` is Go 1.27's new UUID type. It gets its own lesson next.)

And the handlers depend only on the interface:

```go
type apiConfig struct {
	squeaks SqueakStore
	hits    atomic.Int64
}
```

## Why the methods look like that

- **`context.Context` first.** An in-memory map doesn't need it, but a database does:
  if the client hangs up, `r.Context()` is cancelled and a well-behaved database query
  stops too. Putting `ctx` in the interface now means nothing changes later.
- **Every method returns an `error`.** Maps don't fail, but networks and disks do. The
  interface describes what *any* store might do.
- **The store creates the ID and timestamp.** `CreateSqueak` takes only what the client
  decides (author and body) and returns the full squeak. The store owns identity, so
  two callers can never pick the same ID.
- **`ListSqueaks` takes an `authorID`** so a profile page can show one mouse's squeaks.
  Passing the all-zeros `uuid.Nil()` means "every author".
- **`ErrNotFound` is a sentinel** the handler can check with `errors.Is` to choose a 404.
  A database implementation would translate `sql.ErrNoRows` into it, so handlers never
  see storage-specific errors.
- **"Safe for concurrent use"** is in the doc comment, because the compiler can't check
  it. Every handler runs on its own goroutine, so every implementation must hold up to it.

## A handler using the interface

A handler that shows one squeak now reads like a description of the job:

1. Parse the `{id}` path value with `uuid.Parse`. Not a UUID? That's the client's
   mistake: **400**.
2. Ask the store: `cfg.squeaks.GetSqueak(r.Context(), id)`.
3. `errors.Is(err, ErrNotFound)`? **404**.
4. Any other error is *your* problem: log it and answer a generic **500**.
5. Otherwise, **200** with the squeak as JSON.

That handler has no idea whether squeaks live in a map, Postgres or a text file, and
you'll write it at the end of this lesson.

## How big should the interface be?

Go folk wisdom says "the bigger the interface, the weaker the abstraction", and
"accept interfaces, return structs". A repository is a reasonable exception to
"keep interfaces tiny", since it's a cohesive set of operations on one thing. Still:

- One interface per kind of thing (`SqueakStore`, `UserStore`), not one giant `Store`.
- Only the methods the app actually uses. Add `UpdateSqueak` when a handler needs it.
- Define the interface where it's **used** (next to the handlers), and let the
  implementations just have the right methods. Go's implicit interfaces make that easy.

## Test fakes for free

Because handlers take an interface, a test can hand them a fake that fails on purpose:

```go
type brokenStore struct{ SqueakStore } // embed to satisfy the interface

func (brokenStore) GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error) {
	return Squeak{}, errors.New("disk on fire")
}
```

Embedding the interface means you only write the methods the test calls. Calling any
other method panics on the nil embedded value, which is fine in a focused test. Now you
can check your handler really answers 500, a path that's hard to hit with a real store.

## Your task

Write `handleGetSqueak` for `GET /api/squeaks/{id}`, following the five steps above:

- an `id` that `uuid.Parse` rejects: **400** with `respondWithError`,
- `ErrNotFound` from the store (even when wrapped): **404**,
- any other store error: log it with `log.Printf` and respond **500** with a generic
  message that doesn't include the error text,
- otherwise **200** with the squeak, using `respondWithJSON`.

The starter's `fakeStore` implements just `GetSqueak`, with the embedding trick from
above. The tests also use a `brokenStore` like the one above to check your 500 path.
