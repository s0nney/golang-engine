---
title: Build the Squeak Store
quiz:
  - question: |
      Why does `DeleteSqueak` take the exclusive `Lock` instead of `RLock`, even though it
      first *reads* the map to check the squeak exists?
    options:
      - text: '`RLock` can''t be used with `delete`'
      - text: The check and the delete must happen as one step; with a read lock another goroutine could change the map in between, and deleting under a read lock is a data race anyway
        correct: true
      - text: '`RLock` is slower'
      - text: It doesn't matter which one you use
    explanation: |
      Check-then-act must be atomic. Releasing a read lock and then taking a write lock
      opens a gap where another request could delete (or create) the same squeak. Take
      `Lock` once and do both steps under it.
exercise:
  starter: |
    package main

    import (
    	"cmp"
    	"context"
    	"errors"
    	"fmt"
    	"maps"
    	"slices"
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
    	ListSqueaks(ctx context.Context, authorID uuid.UUID) ([]Squeak, error)
    	DeleteSqueak(ctx context.Context, id uuid.UUID) error
    }

    // MemoryStore must be safe for concurrent use.
    type MemoryStore struct {
    	mu      sync.RWMutex
    	squeaks map[uuid.UUID]Squeak
    }

    var _ SqueakStore = (*MemoryStore)(nil)

    func NewMemoryStore() *MemoryStore {
    	return &MemoryStore{squeaks: make(map[uuid.UUID]Squeak)}
    }

    // CreateSqueak stores a new squeak with a uuid.NewV7 ID and CreatedAt = now in UTC.
    func (s *MemoryStore) CreateSqueak(ctx context.Context, authorID uuid.UUID, body string) (Squeak, error) {
    	sq := Squeak{ID: uuid.NewV7(), AuthorID: authorID, Body: body}
    	s.squeaks[sq.ID] = sq
    	return sq, nil
    }

    // GetSqueak returns the squeak with the given ID, or ErrNotFound.
    func (s *MemoryStore) GetSqueak(ctx context.Context, id uuid.UUID) (Squeak, error) {
    	// ?
    	return Squeak{}, nil
    }

    // ListSqueaks returns every squeak, oldest first (ties broken by ID).
    // If authorID isn't uuid.Nil(), only that author's squeaks are returned.
    func (s *MemoryStore) ListSqueaks(ctx context.Context, authorID uuid.UUID) ([]Squeak, error) {
    	// ?
    	return nil, nil
    }

    // DeleteSqueak removes the squeak with the given ID, or returns ErrNotFound.
    func (s *MemoryStore) DeleteSqueak(ctx context.Context, id uuid.UUID) error {
    	// ?
    	return nil
    }

    func main() {
    	ctx := context.Background()
    	store := NewMemoryStore()
    	pip, whiskers := uuid.NewV7(), uuid.NewV7()

    	first, _ := store.CreateSqueak(ctx, pip, "first squeak!")
    	store.CreateSqueak(ctx, whiskers, "cheese at noon?")
    	store.CreateSqueak(ctx, pip, "yes please")

    	all, _ := store.ListSqueaks(ctx, uuid.Nil())
    	fmt.Println("all squeaks:", len(all))
    	mine, _ := store.ListSqueaks(ctx, pip)
    	fmt.Println("pip's squeaks:", len(mine))

    	got, err := store.GetSqueak(ctx, first.ID)
    	fmt.Printf("get first: %q %v\n", got.Body, err)
    	fmt.Println("delete first:", store.DeleteSqueak(ctx, first.ID))
    	fmt.Println("delete again:", store.DeleteSqueak(ctx, first.ID))
    	_, err = store.GetSqueak(ctx, first.ID)
    	fmt.Println("get deleted:", err)

    	_, _, _ = cmp.Or[int], maps.Values[map[int]int], slices.Sort[[]int]
    }
  solution: |
    package main

    import (
    	"cmp"
    	"context"
    	"errors"
    	"fmt"
    	"maps"
    	"slices"
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
    	ListSqueaks(ctx context.Context, authorID uuid.UUID) ([]Squeak, error)
    	DeleteSqueak(ctx context.Context, id uuid.UUID) error
    }

    type MemoryStore struct {
    	mu      sync.RWMutex
    	squeaks map[uuid.UUID]Squeak
    }

    var _ SqueakStore = (*MemoryStore)(nil)

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

    func (s *MemoryStore) ListSqueaks(ctx context.Context, authorID uuid.UUID) ([]Squeak, error) {
    	s.mu.RLock()
    	list := slices.Collect(maps.Values(s.squeaks))
    	s.mu.RUnlock()

    	if authorID != uuid.Nil() {
    		list = slices.DeleteFunc(list, func(sq Squeak) bool { return sq.AuthorID != authorID })
    	}
    	slices.SortFunc(list, func(a, b Squeak) int {
    		return cmp.Or(a.CreatedAt.Compare(b.CreatedAt), a.ID.Compare(b.ID))
    	})
    	return list, nil
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

    func main() {
    	ctx := context.Background()
    	store := NewMemoryStore()
    	pip, whiskers := uuid.NewV7(), uuid.NewV7()

    	first, _ := store.CreateSqueak(ctx, pip, "first squeak!")
    	store.CreateSqueak(ctx, whiskers, "cheese at noon?")
    	store.CreateSqueak(ctx, pip, "yes please")

    	all, _ := store.ListSqueaks(ctx, uuid.Nil())
    	fmt.Println("all squeaks:", len(all))
    	mine, _ := store.ListSqueaks(ctx, pip)
    	fmt.Println("pip's squeaks:", len(mine))

    	got, err := store.GetSqueak(ctx, first.ID)
    	fmt.Printf("get first: %q %v\n", got.Body, err)
    	fmt.Println("delete first:", store.DeleteSqueak(ctx, first.ID))
    	fmt.Println("delete again:", store.DeleteSqueak(ctx, first.ID))
    	_, err = store.GetSqueak(ctx, first.ID)
    	fmt.Println("get deleted:", err)
    }
  tests: |
    package main

    import (
    	"errors"
    	"sync"
    	"testing"
    	"time"
    	"uuid"
    )

    func TestCreateAndGet(t *testing.T) {
    	ctx := t.Context()
    	s := NewMemoryStore()
    	author := uuid.NewV7()
    	before := time.Now()
    	sq, err := s.CreateSqueak(ctx, author, "hello")
    	if err != nil {
    		t.Fatalf("CreateSqueak error: %v", err)
    	}
    	if sq.ID == uuid.Nil() {
    		t.Error("CreateSqueak returned a squeak with a nil ID")
    	}
    	if sq.ID.String()[14] != '7' {
    		t.Errorf("ID %s is not a version 7 UUID; use uuid.NewV7()", sq.ID)
    	}
    	if sq.CreatedAt.IsZero() || sq.CreatedAt.Before(before.Add(-time.Second)) {
    		t.Errorf("CreatedAt = %v, want the current time", sq.CreatedAt)
    	}
    	if sq.CreatedAt.Location() != time.UTC {
    		t.Errorf("CreatedAt location = %v, want UTC", sq.CreatedAt.Location())
    	}
    	got, err := s.GetSqueak(ctx, sq.ID)
    	if err != nil {
    		t.Fatalf("GetSqueak(%s) error: %v", sq.ID, err)
    	}
    	if got != sq {
    		t.Errorf("GetSqueak = %+v, want %+v", got, sq)
    	}
    	if _, err := s.GetSqueak(ctx, uuid.NewV7()); !errors.Is(err, ErrNotFound) {
    		t.Errorf("GetSqueak(unknown id) error = %v, want ErrNotFound", err)
    	}
    }

    func TestListOrderAndFilter(t *testing.T) {
    	ctx := t.Context()
    	s := NewMemoryStore()
    	pip, whiskers := uuid.NewV7(), uuid.NewV7()
    	var want []uuid.UUID
    	for i := range 20 {
    		author := pip
    		if i%3 == 0 {
    			author = whiskers
    		}
    		sq, _ := s.CreateSqueak(ctx, author, "squeak")
    		want = append(want, sq.ID)
    	}
    	all, err := s.ListSqueaks(ctx, uuid.Nil())
    	if err != nil {
    		t.Fatalf("ListSqueaks error: %v", err)
    	}
    	if len(all) != 20 {
    		t.Fatalf("ListSqueaks(uuid.Nil()) returned %d squeaks, want 20", len(all))
    	}
    	for i := range all {
    		if all[i].ID != want[i] {
    			t.Fatalf("ListSqueaks order wrong at index %d: want oldest first (by CreatedAt, then ID)", i)
    		}
    	}
    	w, _ := s.ListSqueaks(ctx, whiskers)
    	if len(w) != 7 {
    		t.Errorf("ListSqueaks(whiskers) returned %d squeaks, want 7", len(w))
    	}
    	for _, sq := range w {
    		if sq.AuthorID != whiskers {
    			t.Errorf("ListSqueaks(whiskers) included a squeak by %s", sq.AuthorID)
    		}
    	}
    	if none, _ := s.ListSqueaks(ctx, uuid.NewV7()); len(none) != 0 {
    		t.Errorf("ListSqueaks(unknown author) returned %d squeaks, want 0", len(none))
    	}

    	all[0].Body = "tampered"
    	again, _ := s.ListSqueaks(ctx, uuid.Nil())
    	if again[0].Body == "tampered" {
    		t.Error("changing a listed squeak changed the store; return copies")
    	}
    }

    func TestDelete(t *testing.T) {
    	ctx := t.Context()
    	s := NewMemoryStore()
    	sq, _ := s.CreateSqueak(ctx, uuid.NewV7(), "bye")
    	if err := s.DeleteSqueak(ctx, sq.ID); err != nil {
    		t.Fatalf("DeleteSqueak error: %v", err)
    	}
    	if _, err := s.GetSqueak(ctx, sq.ID); !errors.Is(err, ErrNotFound) {
    		t.Errorf("GetSqueak after delete: error = %v, want ErrNotFound", err)
    	}
    	if err := s.DeleteSqueak(ctx, sq.ID); !errors.Is(err, ErrNotFound) {
    		t.Errorf("deleting twice: error = %v, want ErrNotFound", err)
    	}
    }

    func TestConcurrentUse(t *testing.T) {
    	ctx := t.Context()
    	s := NewMemoryStore()
    	author := uuid.NewV7()
    	var wg sync.WaitGroup
    	var mu sync.Mutex
    	var created []uuid.UUID
    	for range 20 {
    		wg.Go(func() {
    			for range 50 {
    				sq, err := s.CreateSqueak(ctx, author, "hammer")
    				if err != nil {
    					t.Errorf("CreateSqueak error: %v", err)
    					return
    				}
    				s.GetSqueak(ctx, sq.ID)
    				s.ListSqueaks(ctx, author)
    				mu.Lock()
    				created = append(created, sq.ID)
    				mu.Unlock()
    			}
    		})
    	}
    	wg.Wait()
    	all, _ := s.ListSqueaks(ctx, uuid.Nil())
    	if len(all) != 1000 {
    		t.Fatalf("after 1000 concurrent creates, ListSqueaks has %d squeaks", len(all))
    	}

    	var deleted sync.Map
    	for g := range 10 {
    		wg.Go(func() {
    			for i := g; i < len(created); i += 10 {
    				if err := s.DeleteSqueak(ctx, created[i]); err == nil {
    					if _, dup := deleted.LoadOrStore(created[i], true); dup {
    						t.Errorf("squeak %s was deleted twice", created[i])
    					}
    				}
    				s.ListSqueaks(ctx, uuid.Nil())
    			}
    		})
    	}
    	wg.Wait()
    	if rest, _ := s.ListSqueaks(ctx, uuid.Nil()); len(rest) != 0 {
    		t.Errorf("after deleting everything concurrently, %d squeaks remain", len(rest))
    	}
    }
---

Time to build the store for real. Squeak's handlers will use it for the rest of the course.

## Your task

Finish `MemoryStore` so it satisfies `SqueakStore` **and is safe for concurrent use**:

1. **`CreateSqueak`** already builds a squeak, but it doesn't set `CreatedAt` and it
   writes to the map without a lock. Set `CreatedAt` to the current time in UTC and
   protect the write.
2. **`GetSqueak`** returns the squeak, or `ErrNotFound` if there's no such ID.
3. **`ListSqueaks`** returns all squeaks, or only `authorID`'s when it isn't
   `uuid.Nil()`, sorted **oldest first**, with ties broken by ID. Callers must be able to
   modify the returned slice without affecting the store.
4. **`DeleteSqueak`** removes the squeak, or returns `ErrNotFound` if it doesn't exist.
   Check and delete under **one** exclusive lock.

Use `RLock` for reads and `Lock` for writes.

The tests create 1,000 squeaks from 20 goroutines while reading and listing, then delete
them all from 10 goroutines. An unprotected map will crash the test with
`fatal error: concurrent map writes` (or `concurrent map read and map write`).

## Hints

- `slices.Collect(maps.Values(m))` copies a map's values into a new slice.
- `slices.DeleteFunc(list, func(sq Squeak) bool { ... })` drops the entries you don't want.
- `cmp.Or(a.CreatedAt.Compare(b.CreatedAt), a.ID.Compare(b.ID))` compares by time,
  then by ID.
- The last line of the starter's `main` just keeps unused imports compiling. Delete it
  once you use them.
