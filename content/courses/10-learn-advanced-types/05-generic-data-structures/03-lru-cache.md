---
title: 'Stash: An LRU Cache'
quiz:
  - question: |
      An `LRU[string, int]` has capacity 2. What's evicted by the last line?

      ```go
      c.Put("a", 1)
      c.Put("b", 2)
      c.Get("a")
      c.Put("c", 3)
      ```
    options:
      - text: '`"a"`, because it was added first'
      - text: '`"b"`, because `Get("a")` made `"a"` the most recently used'
        correct: true
      - text: '`"c"`, because the cache is full'
      - text: Nothing; the capacity grows
    explanation: |
      "Least recently *used*", not "least recently added". Reading `"a"` moves it to the
      front, which leaves `"b"` at the back to be evicted when `"c"` arrives.
  - question: Why does Stash's LRU use its own generic linked list instead of `container/list`?
    options:
      - text: '`container/list` doesn''t support removal'
      - text: '`container/list` stores values as `any`, so every read needs a type assertion and non-pointer values get boxed; a generic node type keeps everything typed'
        correct: true
      - text: '`container/list` is deprecated'
      - text: Generic code can't import `container/list`
    explanation: |
      `container/list` predates generics: `Element.Value` is an `any`. It works, but you
      lose static types and pay for interface conversions. A dozen lines of generic
      pointer juggling gives you a typed list.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"iter"
    )

    type node[K comparable, V any] struct {
    	key        K
    	val        V
    	prev, next *node[K, V]
    }

    // LRU is a fixed-capacity cache that evicts the least recently used entry.
    type LRU[K comparable, V any] struct {
    	capacity int
    	items    map[K]*node[K, V]
    	root     node[K, V] // sentinel: root.next is newest, root.prev is oldest

    	// OnEvict, if set, is called with each entry the cache evicts.
    	OnEvict func(K, V)
    }

    // NewLRU returns an empty cache holding at most capacity entries (at least 1).
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

    // Len reports how many entries are cached.
    func (c *LRU[K, V]) Len() int { return len(c.items) }

    // Keys yields the keys from most to least recently used.
    func (c *LRU[K, V]) Keys() iter.Seq[K] {
    	return func(yield func(K) bool) {
    		for n := c.root.next; n != &c.root; n = n.next {
    			if !yield(n.key) {
    				return
    			}
    		}
    	}
    }

    // Get returns the value for k and marks it as most recently used.
    func (c *LRU[K, V]) Get(k K) (V, bool) {
    	// ?
    	var zero V
    	return zero, false
    }

    // Put stores v under k as the most recently used entry. If k is new and
    // the cache is full, it first evicts the least recently used entry,
    // calling OnEvict if it's set.
    func (c *LRU[K, V]) Put(k K, v V) {
    	// ?
    }

    func main() {
    	c := NewLRU[string, int](2)
    	c.OnEvict = func(k string, v int) { fmt.Println("evicted", k, v) }
    	c.Put("a", 1)
    	c.Put("b", 2)
    	c.Get("a")
    	c.Put("c", 3)
    	for k := range c.Keys() {
    		fmt.Print(k, " ")
    	}
    	fmt.Println(c.Len())
    }
  solution: |
    package main

    import (
    	"fmt"
    	"iter"
    )

    type node[K comparable, V any] struct {
    	key        K
    	val        V
    	prev, next *node[K, V]
    }

    // LRU is a fixed-capacity cache that evicts the least recently used entry.
    type LRU[K comparable, V any] struct {
    	capacity int
    	items    map[K]*node[K, V]
    	root     node[K, V] // sentinel: root.next is newest, root.prev is oldest

    	// OnEvict, if set, is called with each entry the cache evicts.
    	OnEvict func(K, V)
    }

    // NewLRU returns an empty cache holding at most capacity entries (at least 1).
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

    // Len reports how many entries are cached.
    func (c *LRU[K, V]) Len() int { return len(c.items) }

    // Keys yields the keys from most to least recently used.
    func (c *LRU[K, V]) Keys() iter.Seq[K] {
    	return func(yield func(K) bool) {
    		for n := c.root.next; n != &c.root; n = n.next {
    			if !yield(n.key) {
    				return
    			}
    		}
    	}
    }

    // Get returns the value for k and marks it as most recently used.
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

    // Put stores v under k as the most recently used entry. If k is new and
    // the cache is full, it first evicts the least recently used entry,
    // calling OnEvict if it's set.
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
    		if c.OnEvict != nil {
    			c.OnEvict(oldest.key, oldest.val)
    		}
    	}
    	n := &node[K, V]{key: k, val: v}
    	c.pushFront(n)
    	c.items[k] = n
    }

    func main() {
    	c := NewLRU[string, int](2)
    	c.OnEvict = func(k string, v int) { fmt.Println("evicted", k, v) }
    	c.Put("a", 1)
    	c.Put("b", 2)
    	c.Get("a")
    	c.Put("c", 3)
    	for k := range c.Keys() {
    		fmt.Print(k, " ")
    	}
    	fmt.Println(c.Len())
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    func keys(c *LRU[string, int]) []string { return slices.Collect(c.Keys()) }

    func TestPutAndGet(t *testing.T) {
    	c := NewLRU[string, int](3)
    	c.Put("a", 1)
    	c.Put("b", 2)
    	if v, ok := c.Get("a"); !ok || v != 1 {
    		t.Errorf(`Get("a") = (%d, %v), want (1, true)`, v, ok)
    	}
    	if v, ok := c.Get("zzz"); ok || v != 0 {
    		t.Errorf(`Get("zzz") = (%d, %v), want (0, false)`, v, ok)
    	}
    	if c.Len() != 2 {
    		t.Errorf("Len() = %d, want 2", c.Len())
    	}
    	if got := keys(c); !slices.Equal(got, []string{"a", "b"}) {
    		t.Errorf("after Put a, Put b, Get a: Keys() = %v, want [a b] (most recent first)", got)
    	}
    }

    func TestEvictsLeastRecentlyUsed(t *testing.T) {
    	c := NewLRU[string, int](2)
    	var evicted []string
    	c.OnEvict = func(k string, v int) { evicted = append(evicted, k) }
    	c.Put("a", 1)
    	c.Put("b", 2)
    	c.Get("a")
    	c.Put("c", 3)
    	if _, ok := c.Get("b"); ok {
    		t.Errorf(`"b" should have been evicted (least recently used)`)
    	}
    	if !slices.Equal(evicted, []string{"b"}) {
    		t.Errorf("OnEvict saw %v, want [b]", evicted)
    	}
    	if c.Len() != 2 {
    		t.Errorf("Len() = %d, want 2 (the capacity)", c.Len())
    	}
    	if got := keys(c); !slices.Equal(got, []string{"c", "a"}) {
    		t.Errorf("Keys() = %v, want [c a]", got)
    	}
    }

    func TestUpdateMovesToFront(t *testing.T) {
    	c := NewLRU[string, int](2)
    	c.Put("a", 1)
    	c.Put("b", 2)
    	c.Put("a", 10) // update, not a new entry
    	if c.Len() != 2 {
    		t.Fatalf("updating an existing key changed Len() to %d, want 2", c.Len())
    	}
    	c.Put("c", 3) // should evict b, not a
    	if v, ok := c.Get("a"); !ok || v != 10 {
    		t.Errorf(`Get("a") = (%d, %v), want (10, true)`, v, ok)
    	}
    	if _, ok := c.Get("b"); ok {
    		t.Errorf(`"b" should have been evicted`)
    	}
    }

    func TestCapacityOneAndNoCallback(t *testing.T) {
    	c := NewLRU[string, int](1)
    	for i, k := range []string{"x", "y", "z"} {
    		c.Put(k, i)
    	}
    	if got := keys(c); !slices.Equal(got, []string{"z"}) {
    		t.Errorf("capacity 1 after x, y, z: Keys() = %v, want [z]", got)
    	}
    }
---

Stash's next piece is a cache. A cache can't grow forever, so when it's full, something has to go. The classic policy is **LRU: evict the least recently used entry**. Each `Get` or `Put` marks an entry as fresh; the stalest entry is the one to drop.

## Two structures again

An LRU needs two things to be fast:

- **O(1) lookup** by key: a map.
- **O(1) "move to front"** and **"remove from the back"**: a doubly linked list ordered by recency.

The map points straight at list nodes, so a hit can unlink its node and move it to the front without searching.

## A generic node

```go
type node[K comparable, V any] struct {
	key        K
	val        V
	prev, next *node[K, V]
}
```

A generic type can refer to itself, as long as it uses **the same type parameters in the same order**: `*node[K, V]` inside `node[K, V]`. The node stores the key too, because when we evict the last node we need its key to delete it from the map.

## A sentinel makes the list easy

Rather than tracking `head` and `tail` pointers and special-casing an empty list, the cache holds one dummy node, `root`, in a ring:

```go
type LRU[K comparable, V any] struct {
	capacity int
	items    map[K]*node[K, V]
	root     node[K, V] // sentinel: root.next is newest, root.prev is oldest

	OnEvict func(K, V)
}

func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
	c := &LRU[K, V]{capacity: max(capacity, 1), items: make(map[K]*node[K, V])}
	c.root.next, c.root.prev = &c.root, &c.root // empty ring
	return c
}
```

An empty cache is a ring of just `root`. The newest entry is always `root.next` and the oldest is `root.prev`, and the two helpers never need an `if`:

```go
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
```

`container/list` uses exactly this trick internally. Stash rolls its own because `container/list` stores `any` values: every read would need `e.Value.(*entry)`, and every non-pointer value would be boxed into an interface.

Two small design notes:

- `LRU` has a constructor, `NewLRU[string, int](2)`, because a cache without a capacity makes no sense. Neither type argument can be inferred from an `int`, so callers write both. (Put `K` and `V` in the order people say them: "a cache from string to int".)
- `OnEvict` is an exported **field** of function type. It's the lightest way to offer a hook, and `nil` means "no callback".

## Your turn

Write the two operations that make it a cache:

- `Get(k)`: on a hit, move the node to the front and return its value and `true`. On a miss, return the zero value and `false`.
- `Put(k, v)`:
  - if `k` is already cached, update its value and move it to the front (the size doesn't change);
  - otherwise, if the cache is full, evict the oldest node (`c.root.prev`): unlink it, delete it from `items`, and call `OnEvict` if it isn't `nil`. Then add a new node at the front.
