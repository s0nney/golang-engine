---
title: An In-Memory Store
quiz:
  - question: |
      What happens when `ListSqueaks` is called?

      ```go
      func (s *MemoryStore) ListSqueaks(ctx context.Context, authorID uuid.UUID) ([]Squeak, error) {
      	s.mu.RLock()
      	defer s.mu.RUnlock()
      	for _, sq := range s.squeaks {
      		if _, err := s.GetSqueak(ctx, sq.ID); err != nil { // GetSqueak also calls RLock
      			return nil, err
      		}
      	}
      	// ...
      }
      ```
    options:
      - text: It works; read locks can always be taken twice
      - text: It can deadlock, because `sync.RWMutex` isn't reentrant and a waiting writer blocks new readers
        correct: true
      - text: It panics with "concurrent map read"
      - text: It fails to compile
    explanation: |
      Recursive read locking is explicitly forbidden by the `sync.RWMutex` docs. If a writer
      calls `Lock` between the two `RLock` calls, the writer waits for the first reader, and
      the second `RLock` waits for the writer: deadlock. Have locked methods call unexported
      helpers that assume the lock is already held.
  - question: When is `sync.RWMutex` a better choice than `sync.Mutex` for Squeak's store?
    options:
      - text: Always; it's strictly faster
      - text: When reads far outnumber writes, so many readers can hold the lock at once
        correct: true
      - text: When there are more writes than reads
      - text: Only when using channels
    explanation: |
      `RLock` lets many readers in together, and `Lock` is exclusive. A timeline is read far
      more often than it's written, which is the case `RWMutex` is for. With mostly writes,
      its extra bookkeeping makes it slightly slower than a plain `Mutex`.
---

Here's the store Squeak will run on until the day it meets a real database: a map guarded
by a mutex.

## The struct

```go
type MemoryStore struct {
	mu      sync.RWMutex
	squeaks map[uuid.UUID]Squeak
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{squeaks: make(map[uuid.UUID]Squeak)}
}

// Compile-time check that *MemoryStore implements SqueakStore.
var _ SqueakStore = (*MemoryStore)(nil)
```

- The **constructor** creates the map. Writing to a nil map panics, which is the classic
  bug of a `MemoryStore{}` literal.
- The **mutex sits next to the data it protects**, lowercase so nothing outside the
  type can touch either.
- The `var _ SqueakStore = ...` line costs nothing at runtime and turns "forgot a method"
  into a compile error right here, not somewhere far away where the store gets used.

## Why a lock at all?

Go maps are **not** safe for concurrent use. Two handlers writing at the same time can
corrupt the map, and the runtime detects that and kills the whole program with
`fatal error: concurrent map writes`. That's not a panic you can recover from. And since
every request has its own goroutine, "two writes at once" is just two mice squeaking at
the same moment.

## Writes take Lock, reads take RLock

```go
func (s *MemoryStore) CreateSqueak(ctx context.Context, authorID uuid.UUID, body string) (Squeak, error) {
	sq := Squeak{
		ID:        uuid.NewV7(),
		AuthorID:  authorID,
		Body:      body,
		CreatedAt: time.Now().UTC(),
	}
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
```

`sync.RWMutex` has two modes. `Lock` is exclusive: one writer, nobody else. `RLock` is
shared: any number of readers at once, but no writer. Squeak's timeline is read far more
than it's written, which is exactly the case it's for.

Notice that `CreateSqueak` builds the squeak *before* taking the lock. Keep critical
sections short: only the map access needs protecting.

## Return copies, not pointers

The map holds `Squeak` **values**, and methods return values. The caller gets a copy,
so it can change its copy without a lock and without affecting the store.

If the map held `*Squeak`, a handler could modify a squeak through the returned pointer
while another goroutine reads it: a data race the mutex can't prevent, because the
access happens *outside* the lock. The same goes for slices and maps *inside* your
structs, which share their backing data when copied. Copy them with `slices.Clone` or
`maps.Clone` before returning.

## Listing

```go
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
```

Map iteration order is random, so sort before returning. Here the lock is released
*before* filtering and sorting: the new slice belongs to this call alone, so there's no reason to
make writers wait while it's sorted.

`maps.Values` returns an iterator, and `slices.Collect` gathers it into a fresh slice,
which is a new backing array that no other goroutine can see.

## Don't call yourself while locked

`sync.RWMutex` is **not reentrant**. A method that holds the lock and calls another
method that takes it again can deadlock, even with two read locks (a writer waiting in
between blocks the second `RLock`). If you need to share logic, move it into an
unexported helper documented as "caller must hold s.mu".

## The race detector

Your best friend here is `go test -race`, which you met in the concurrency course. It
instruments memory accesses and reports any unsynchronized read/write pair it sees,
even when the run happened to produce correct output. Run your store's tests with it,
with lots of goroutines. The next exercise does exactly that kind of hammering.
