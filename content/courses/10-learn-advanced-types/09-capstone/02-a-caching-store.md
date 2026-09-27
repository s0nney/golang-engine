---
title: 'A Caching Store'
quiz:
  - question: |
      `Cached.Delete` removes a key from the store but forgets to remove it from the LRU. What goes wrong?
    options:
      - text: Nothing; the LRU will evict it eventually
      - text: '`Get` keeps returning the deleted value from the cache until it happens to be evicted'
        correct: true
      - text: The next `Set` panics
      - text: '`All` returns the key twice'
    explanation: |
      Reads check the cache first, so a stale entry wins over the store's "not found".
      Every write path (`Set` and `Delete`) must keep the cache consistent with the store.
  - question: Why is `var _ Store[string, int] = (*Cached[string, int])(nil)` useful?
    options:
      - text: It creates a global cache
      - text: It's a compile-time check that `*Cached` implements `Store` (for one instantiation), costing nothing at runtime
        correct: true
      - text: It registers `Cached` with the `Store` interface
      - text: It's required for generic types to satisfy interfaces
    explanation: |
      If a method is missing or has the wrong signature, this line fails to compile, right
      next to the type, instead of somewhere far away where a `*Cached` is first used as a
      `Store`.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"iter"
    	"slices"
    )

    // ---- From chapter 5: OrderedMap ----

    // OrderedMap is a map that remembers insertion order.
    type OrderedMap[K comparable, V any] struct {
    	keys []K
    	vals map[K]V
    }

    func (m *OrderedMap[K, V]) Set(k K, v V) {
    	if m.vals == nil {
    		m.vals = make(map[K]V)
    	}
    	if _, ok := m.vals[k]; !ok {
    		m.keys = append(m.keys, k)
    	}
    	m.vals[k] = v
    }

    func (m *OrderedMap[K, V]) Get(k K) (V, bool) {
    	v, ok := m.vals[k]
    	return v, ok
    }

    func (m *OrderedMap[K, V]) Delete(k K) bool {
    	if _, ok := m.vals[k]; !ok {
    		return false
    	}
    	delete(m.vals, k)
    	i := slices.Index(m.keys, k)
    	m.keys = slices.Delete(m.keys, i, i+1)
    	return true
    }

    func (m *OrderedMap[K, V]) All() iter.Seq2[K, V] {
    	return func(yield func(K, V) bool) {
    		for _, k := range m.keys {
    			if !yield(k, m.vals[k]) {
    				return
    			}
    		}
    	}
    }

    // ---- From chapter 5: LRU (now with Remove) ----

    type node[K comparable, V any] struct {
    	key        K
    	val        V
    	prev, next *node[K, V]
    }

    // LRU is a fixed-capacity cache that evicts the least recently used entry.
    type LRU[K comparable, V any] struct {
    	capacity int
    	items    map[K]*node[K, V]
    	root     node[K, V]
    }

    func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
    	c := &LRU[K, V]{capacity: max(capacity, 1), items: make(map[K]*node[K, V])}
    	c.root.next, c.root.prev = &c.root, &c.root
    	return c
    }

    func (c *LRU[K, V]) unlink(n *node[K, V]) {
    	n.prev.next = n.next
    	n.next.prev = n.prev
    }

    func (c *LRU[K, V]) pushFront(n *node[K, V]) {
    	n.prev = &c.root
    	n.next = c.root.next
    	c.root.next.prev = n
    	c.root.next = n
    }

    func (c *LRU[K, V]) Get(k K) (V, bool) {
    	n, ok := c.items[k]
    	if !ok {
    		var zero V
    		return zero, false
    	}
    	c.unlink(n)
    	c.pushFront(n)
    	return n.val, true
    }

    func (c *LRU[K, V]) Put(k K, v V) {
    	if n, ok := c.items[k]; ok {
    		n.val = v
    		c.unlink(n)
    		c.pushFront(n)
    		return
    	}
    	if len(c.items) >= c.capacity {
    		oldest := c.root.prev
    		c.unlink(oldest)
    		delete(c.items, oldest.key)
    	}
    	n := &node[K, V]{key: k, val: v}
    	c.pushFront(n)
    	c.items[k] = n
    }

    // Remove drops k from the cache and reports whether it was there.
    func (c *LRU[K, V]) Remove(k K) bool {
    	n, ok := c.items[k]
    	if !ok {
    		return false
    	}
    	c.unlink(n)
    	delete(c.items, k)
    	return true
    }

    // ---- New: the Store interface and a caching decorator ----

    // Store is anything that stores values by key. *OrderedMap satisfies it.
    type Store[K comparable, V any] interface {
    	Get(k K) (V, bool)
    	Set(k K, v V)
    	Delete(k K) bool
    	All() iter.Seq2[K, V]
    }

    // Cached wraps a Store with an LRU cache for reads. It's a Store too.
    type Cached[K comparable, V any] struct {
    	store        Store[K, V]
    	cache        *LRU[K, V]
    	hits, misses int
    }

    // NewCached returns a Cached in front of store, caching up to capacity values.
    func NewCached[K comparable, V any](store Store[K, V], capacity int) *Cached[K, V] {
    	return &Cached[K, V]{store: store, cache: NewLRU[K, V](capacity)}
    }

    // Get serves k from the cache (a hit) or else from the store (a miss),
    // caching what the store returns.
    func (c *Cached[K, V]) Get(k K) (V, bool) {
    	// ?
    	return c.store.Get(k)
    }

    // Set writes through: to the store first, then the cache.
    func (c *Cached[K, V]) Set(k K, v V) {
    	// ?
    	c.store.Set(k, v)
    }

    // Delete removes k from the store and the cache.
    func (c *Cached[K, V]) Delete(k K) bool {
    	// ?
    	return c.store.Delete(k)
    }

    // All iterates the underlying store.
    func (c *Cached[K, V]) All() iter.Seq2[K, V] { return c.store.All() }

    // Stats reports cache hits and misses so far.
    func (c *Cached[K, V]) Stats() (hits, misses int) { return c.hits, c.misses }

    // Compile-time checks: both are Stores.
    var (
    	_ Store[string, int] = (*OrderedMap[string, int])(nil)
    	_ Store[string, int] = (*Cached[string, int])(nil)
    )

    func main() {
    	var backing OrderedMap[string, int]
    	c := NewCached(&backing, 2) // K and V inferred from the methods
    	c.Set("logo.png", 2048)
    	c.Set("app.js", 512)
    	backing.Set("old.css", 64) // written behind the cache's back

    	for _, k := range []string{"logo.png", "old.css", "old.css", "missing"} {
    		v, ok := c.Get(k)
    		fmt.Println(k, v, ok)
    	}
    	fmt.Println(c.Delete("app.js"), c.Delete("app.js"))
    	for k, v := range c.All() {
    		fmt.Print(k, "=", v, " ")
    	}
    	fmt.Println()
    	fmt.Println(c.Stats())
    }
  solution: |
    package main

    import (
    	"fmt"
    	"iter"
    	"slices"
    )

    // ---- From chapter 5: OrderedMap ----

    // OrderedMap is a map that remembers insertion order.
    type OrderedMap[K comparable, V any] struct {
    	keys []K
    	vals map[K]V
    }

    func (m *OrderedMap[K, V]) Set(k K, v V) {
    	if m.vals == nil {
    		m.vals = make(map[K]V)
    	}
    	if _, ok := m.vals[k]; !ok {
    		m.keys = append(m.keys, k)
    	}
    	m.vals[k] = v
    }

    func (m *OrderedMap[K, V]) Get(k K) (V, bool) {
    	v, ok := m.vals[k]
    	return v, ok
    }

    func (m *OrderedMap[K, V]) Delete(k K) bool {
    	if _, ok := m.vals[k]; !ok {
    		return false
    	}
    	delete(m.vals, k)
    	i := slices.Index(m.keys, k)
    	m.keys = slices.Delete(m.keys, i, i+1)
    	return true
    }

    func (m *OrderedMap[K, V]) All() iter.Seq2[K, V] {
    	return func(yield func(K, V) bool) {
    		for _, k := range m.keys {
    			if !yield(k, m.vals[k]) {
    				return
    			}
    		}
    	}
    }

    // ---- From chapter 5: LRU (now with Remove) ----

    type node[K comparable, V any] struct {
    	key        K
    	val        V
    	prev, next *node[K, V]
    }

    // LRU is a fixed-capacity cache that evicts the least recently used entry.
    type LRU[K comparable, V any] struct {
    	capacity int
    	items    map[K]*node[K, V]
    	root     node[K, V]
    }

    func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
    	c := &LRU[K, V]{capacity: max(capacity, 1), items: make(map[K]*node[K, V])}
    	c.root.next, c.root.prev = &c.root, &c.root
    	return c
    }

    func (c *LRU[K, V]) unlink(n *node[K, V]) {
    	n.prev.next = n.next
    	n.next.prev = n.prev
    }

    func (c *LRU[K, V]) pushFront(n *node[K, V]) {
    	n.prev = &c.root
    	n.next = c.root.next
    	c.root.next.prev = n
    	c.root.next = n
    }

    func (c *LRU[K, V]) Get(k K) (V, bool) {
    	n, ok := c.items[k]
    	if !ok {
    		var zero V
    		return zero, false
    	}
    	c.unlink(n)
    	c.pushFront(n)
    	return n.val, true
    }

    func (c *LRU[K, V]) Put(k K, v V) {
    	if n, ok := c.items[k]; ok {
    		n.val = v
    		c.unlink(n)
    		c.pushFront(n)
    		return
    	}
    	if len(c.items) >= c.capacity {
    		oldest := c.root.prev
    		c.unlink(oldest)
    		delete(c.items, oldest.key)
    	}
    	n := &node[K, V]{key: k, val: v}
    	c.pushFront(n)
    	c.items[k] = n
    }

    // Remove drops k from the cache and reports whether it was there.
    func (c *LRU[K, V]) Remove(k K) bool {
    	n, ok := c.items[k]
    	if !ok {
    		return false
    	}
    	c.unlink(n)
    	delete(c.items, k)
    	return true
    }

    // ---- New: the Store interface and a caching decorator ----

    // Store is anything that stores values by key. *OrderedMap satisfies it.
    type Store[K comparable, V any] interface {
    	Get(k K) (V, bool)
    	Set(k K, v V)
    	Delete(k K) bool
    	All() iter.Seq2[K, V]
    }

    // Cached wraps a Store with an LRU cache for reads. It's a Store too.
    type Cached[K comparable, V any] struct {
    	store        Store[K, V]
    	cache        *LRU[K, V]
    	hits, misses int
    }

    // NewCached returns a Cached in front of store, caching up to capacity values.
    func NewCached[K comparable, V any](store Store[K, V], capacity int) *Cached[K, V] {
    	return &Cached[K, V]{store: store, cache: NewLRU[K, V](capacity)}
    }

    // Get serves k from the cache (a hit) or else from the store (a miss),
    // caching what the store returns.
    func (c *Cached[K, V]) Get(k K) (V, bool) {
    	if v, ok := c.cache.Get(k); ok {
    		c.hits++
    		return v, true
    	}
    	c.misses++
    	v, ok := c.store.Get(k)
    	if ok {
    		c.cache.Put(k, v)
    	}
    	return v, ok
    }

    // Set writes through: to the store first, then the cache.
    func (c *Cached[K, V]) Set(k K, v V) {
    	c.store.Set(k, v)
    	c.cache.Put(k, v)
    }

    // Delete removes k from the store and the cache.
    func (c *Cached[K, V]) Delete(k K) bool {
    	c.cache.Remove(k)
    	return c.store.Delete(k)
    }

    // All iterates the underlying store.
    func (c *Cached[K, V]) All() iter.Seq2[K, V] { return c.store.All() }

    // Stats reports cache hits and misses so far.
    func (c *Cached[K, V]) Stats() (hits, misses int) { return c.hits, c.misses }

    // Compile-time checks: both are Stores.
    var (
    	_ Store[string, int] = (*OrderedMap[string, int])(nil)
    	_ Store[string, int] = (*Cached[string, int])(nil)
    )

    func main() {
    	var backing OrderedMap[string, int]
    	c := NewCached(&backing, 2) // K and V inferred from the methods
    	c.Set("logo.png", 2048)
    	c.Set("app.js", 512)
    	backing.Set("old.css", 64) // written behind the cache's back

    	for _, k := range []string{"logo.png", "old.css", "old.css", "missing"} {
    		v, ok := c.Get(k)
    		fmt.Println(k, v, ok)
    	}
    	fmt.Println(c.Delete("app.js"), c.Delete("app.js"))
    	for k, v := range c.All() {
    		fmt.Print(k, "=", v, " ")
    	}
    	fmt.Println()
    	fmt.Println(c.Stats())
    }
  tests: |
    package main

    import (
    	"iter"
    	"testing"
    )

    // countingStore wraps an OrderedMap and counts reads that reach it.
    type countingStore struct {
    	m     OrderedMap[string, int]
    	reads int
    }

    func (s *countingStore) Get(k string) (int, bool) {
    	s.reads++
    	return s.m.Get(k)
    }
    func (s *countingStore) Set(k string, v int)         { s.m.Set(k, v) }
    func (s *countingStore) Delete(k string) bool        { return s.m.Delete(k) }
    func (s *countingStore) All() iter.Seq2[string, int] { return s.m.All() }

    func TestReadThrough(t *testing.T) {
    	var s countingStore
    	s.m.Set("a", 1)
    	c := NewCached[string, int](&s, 2)
    	for range 3 {
    		if v, ok := c.Get("a"); !ok || v != 1 {
    			t.Fatalf(`Get("a") = (%d, %v), want (1, true)`, v, ok)
    		}
    	}
    	if s.reads != 1 {
    		t.Errorf("3 Gets of the same key reached the store %d times, want 1 (then cached)", s.reads)
    	}
    	if h, m := c.Stats(); h != 2 || m != 1 {
    		t.Errorf("Stats() = (%d hits, %d misses), want (2, 1)", h, m)
    	}
    	if _, ok := c.Get("nope"); ok {
    		t.Errorf(`Get("nope") found something`)
    	}
    	if _, ok := c.cache.Get("nope"); ok {
    		t.Errorf("a miss for a missing key must not be cached")
    	}
    }

    func TestWriteThrough(t *testing.T) {
    	var s countingStore
    	c := NewCached[string, int](&s, 2)
    	c.Set("k", 7)
    	if v, ok := s.m.Get("k"); !ok || v != 7 {
    		t.Errorf("after Set, the store has (%d, %v), want (7, true)", v, ok)
    	}
    	if v, ok := c.Get("k"); !ok || v != 7 || s.reads != 0 {
    		t.Errorf("Get after Set = (%d, %v) with %d store reads; want (7, true) served from cache", v, ok, s.reads)
    	}
    	c.Set("k", 8)
    	if v, _ := c.Get("k"); v != 8 {
    		t.Errorf("Get after overwrite = %d, want 8", v)
    	}
    }

    func TestDeleteInvalidates(t *testing.T) {
    	var s countingStore
    	c := NewCached[string, int](&s, 2)
    	c.Set("k", 1)
    	c.Get("k")
    	if !c.Delete("k") {
    		t.Errorf(`Delete("k") = false, want true`)
    	}
    	if v, ok := c.Get("k"); ok {
    		t.Errorf(`Get("k") after Delete = (%d, true): stale value served from the cache`, v)
    	}
    	if c.Delete("k") {
    		t.Errorf(`second Delete("k") = true, want false`)
    	}
    }

    func TestAllAndEviction(t *testing.T) {
    	var s countingStore
    	c := NewCached[string, int](&s, 1)
    	c.Set("a", 1)
    	c.Set("b", 2)
    	c.Get("a") // evicted by b: must come from the store
    	if s.reads != 1 {
    		t.Errorf("Get of an evicted key: %d store reads, want 1", s.reads)
    	}
    	var keys []string
    	for k := range c.All() {
    		keys = append(keys, k)
    	}
    	if len(keys) != 2 || keys[0] != "a" || keys[1] != "b" {
    		t.Errorf("All() keys = %v, want [a b] (the store's order)", keys)
    	}
    }
---

The first new piece is a **decorator**: a `Store` that wraps another `Store` and adds behaviour, here an LRU read cache. Because it has the same methods, it's a `Store` itself, so callers can't tell the difference, except that repeated reads get faster.

## Read-through, write-through

`Cached[K, V]` follows two classic policies:

- **Read-through**: `Get` checks the cache first. On a **hit**, it returns the cached value without touching the store. On a **miss**, it reads the store and, if the key exists, puts the value in the cache for next time. Missing keys aren't cached.
- **Write-through**: `Set` writes to the store first (the source of truth), then updates the cache so the next read is a hit.

And `Delete` removes the key from **both**, or reads would keep finding the stale cached value.

`All` just delegates to the underlying store: iterating everything through a cache would only churn it.

## The shape of the code

```go
type Cached[K comparable, V any] struct {
	store        Store[K, V]
	cache        *LRU[K, V]
	hits, misses int
}

func NewCached[K comparable, V any](store Store[K, V], capacity int) *Cached[K, V]
```

Two design details:

- The constructor can't infer `K` and `V` from the `int` capacity, but it can from the store. If you pass a `Store[string, int]` interface value, that's ordinary unification. More surprisingly, `NewCached(&backing, 2)` with a concrete `*OrderedMap[string, int]` works too: since Go 1.21, when an argument is passed to a parameter of generic interface type, inference also matches the argument's **methods** against the interface's, so `Get(k string) (int, bool)` reveals `K = string` and `V = int`.
- A pair of `var _ Store[...] = ...` lines at the bottom check at compile time that both `*OrderedMap` and `*Cached` implement `Store`.

The LRU from chapter 5 gains one method, `Remove(k) bool`, so `Delete` can invalidate entries.

## A note on concurrency

`Cached` isn't safe for concurrent use: even `Get` modifies the LRU's order and the counters. A concurrent version would put a `sync.Mutex` around each method, as in Learn Concurrency. Stash keeps it single-goroutine so the generic design stays in focus.

## Your turn

Complete `Cached`:

- `Get(k)`: on a cache hit, count a hit and return it. Otherwise count a miss, read the store, and cache the value if the store had it.
- `Set(k, v)`: write to the store, then put the value in the cache.
- `Delete(k)`: remove the key from the cache and the store, returning whether the store had it.

The tests use a store that counts its reads, so they can tell whether the cache really saves trips.
