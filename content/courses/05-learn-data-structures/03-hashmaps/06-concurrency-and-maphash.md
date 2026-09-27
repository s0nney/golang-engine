---
title: Concurrency and hash/maphash
quiz:
  - question: Two goroutines write to the same plain `map[string]int` at the same time, with no locking. What happens?
    options:
      - text: Go serialises the writes automatically, so it's safe
      - text: One write is silently lost, but nothing else goes wrong
      - text: It's a data race; the runtime may crash the program with "concurrent map writes"
        correct: true
      - text: It doesn't compile
    explanation: |
      Built-in maps are not safe for concurrent writes (or a write alongside reads).
      The runtime detects many of these races and kills the program with a fatal error
      that `recover` can't catch. Protect the map with a `sync.Mutex`, or use `sync.Map`.
  - question: When is `sync.Map` a better fit than a regular map guarded by a `sync.Mutex`?
    options:
      - text: Always; it's the concurrent version of `map`
      - text: 'For caches where keys are written once and read many times, or goroutines work on disjoint sets of keys'
        correct: true
      - text: When you need the keys in sorted order
      - text: When you want compile-time type checking of keys and values
    explanation: |
      `sync.Map` is specialised for those two patterns. For everything else, the Go
      docs recommend a plain map plus a mutex: it's type-safe and easier to keep
      consistent with other state. `sync.Map` stores `any`, so you lose type safety.
  - question: Why does `maphash` make you supply a random `Seed`?
    options:
      - text: So that hashes are the same across every run of the program
      - text: So attackers can't predict which keys collide and flood one bucket on purpose
        correct: true
      - text: Seeds make hashing faster
      - text: To make the hash cryptographically secure
    explanation: |
      If the hash function were fixed and public, someone could precompute thousands of
      usernames that all land in one bucket, turning O(1) lookups into O(n) (a
      "hash flooding" attack). A random per-table seed makes that impossible. It also
      means hash values differ between runs, so never store them on disk.
---

Two last things before we leave hashmaps: using maps from many goroutines at once,
and hashing arbitrary keys in your own data structures.

## Maps and goroutines

The game server handles each player's connection in its own goroutine. They all
update one `map[string]int` of scores. That's a problem: **built-in maps are not safe
for concurrent use** when any goroutine writes. Internally, a write may be moving
slots around or splitting a table while another goroutine reads, and the results
would be garbage. Rather than corrupting memory quietly, the runtime usually detects
it and stops the whole program:

```
fatal error: concurrent map writes
```

This isn't a `panic`. You can't `recover` from it. The race detector
(`go run -race`, `go test -race`) finds these bugs reliably, so use it.

## Fix 1: a mutex

The standard fix is to wrap the map and a `sync.Mutex` in a struct, and lock
around every access:

```go
package main

import (
	"fmt"
	"sync"
)

type Scores struct {
	mu     sync.Mutex
	byName map[string]int
}

func (s *Scores) Add(name string, points int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byName[name] += points
}

func (s *Scores) Get(name string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.byName[name]
}

func main() {
	s := &Scores{byName: make(map[string]int)}
	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(func() { s.Add("mira", 1) }) // 1,000 goroutines at once
	}
	wg.Wait()
	fmt.Println(s.Get("mira"))
}
```

This prints `1000` every time. `wg.Go` (Go 1.25) starts a goroutine and tracks it
in one call. Pass the struct around by pointer: copying a `sync.Mutex` breaks it, and
`go vet` warns you if you try. If reads vastly outnumber writes, `sync.RWMutex`
lets many readers in at once.

## Fix 2: sync.Map

`sync.Map` is a concurrent map in the standard library, with methods such as
`Load`, `Store`, `LoadOrStore`, `Delete` and `Range`. Its docs are unusually frank:
most code should use a plain map with a mutex instead. `sync.Map` shines in two
situations: caches where each key is written once and read many times, and
goroutines working on disjoint sets of keys. Its keys and values are `any`, so you
also give up type safety and need type assertions on every `Load`.

## hash/maphash: hashing any key

Our from-scratch `HashMap` only took `string` keys, because FNV hashes bytes. The
`hash/maphash` package hashes **any comparable value** with the same fast, seeded
hash the runtime uses. That lets us make the key type generic:

```go
package main

import (
	"fmt"
	"hash/maphash"
)

type entry[K comparable, V any] struct {
	key K
	val V
}

type HashMap[K comparable, V any] struct {
	seed    maphash.Seed
	buckets [][]entry[K, V]
}

func NewHashMap[K comparable, V any](n int) *HashMap[K, V] {
	return &HashMap[K, V]{seed: maphash.MakeSeed(), buckets: make([][]entry[K, V], n)}
}

func (m *HashMap[K, V]) bucket(key K) *[]entry[K, V] {
	h := maphash.Comparable(m.seed, key)
	return &m.buckets[h%uint64(len(m.buckets))]
}

func (m *HashMap[K, V]) Set(key K, val V) {
	b := m.bucket(key)
	for i := range *b {
		if (*b)[i].key == key {
			(*b)[i].val = val
			return
		}
	}
	*b = append(*b, entry[K, V]{key, val})
}

func (m *HashMap[K, V]) Get(key K) (V, bool) {
	for _, e := range *m.bucket(key) {
		if e.key == key {
			return e.val, true
		}
	}
	var zero V
	return zero, false
}

type Zone struct{ X, Y int }

func main() {
	names := NewHashMap[Zone, string](16)
	names.Set(Zone{0, 0}, "Spawn Village")
	names.Set(Zone{3, -1}, "Goblin Caves")
	fmt.Println(names.Get(Zone{3, -1}))
	fmt.Println(names.Get(Zone{9, 9}))
}
```

Output:

```
Goblin Caves true
 false
```

(The second line starts with a space because the zero value of `string` is empty.)
Struct keys like `Zone` just work, as long as all their fields are comparable.

## Why the seed?

`maphash.MakeSeed()` returns a random seed, and hashes only make sense relative to
it. That's deliberate. If your hash function were fixed, an attacker who knows it
could sign up thousands of usernames crafted to land in the same bucket, and every
lookup would crawl through one enormous chain. That's called **hash flooding**, and
random seeds defeat it. Go's built-in map uses a random seed per map for the same
reason. The flip side: hash values change every run, so never save them to disk or
send them over the network.

## Further reading

- [Go by Example: Mutexes](https://gobyexample.com/mutexes)
- `go doc sync.Map` and `go doc hash/maphash`
