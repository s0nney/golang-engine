---
title: Building a Hashmap
quiz:
  - question: |
      What does this print, using the `HashMap` from this lesson?

      ```go
      m := NewHashMap[int](8)
      m.Set("mira", 10)
      m.Set("mira", 25)
      v, ok := m.Get("mira")
      fmt.Println(v, ok, m.Len())
      ```
    options:
      - text: '`10 true 2`'
      - text: '`25 true 2`'
      - text: '`25 true 1`'
        correct: true
      - text: '`10 true 1`'
    explanation: |
      `Set` first scans the bucket for the key. It finds `"mira"` and overwrites the
      value in place, so there's still one entry and `Len` stays at 1.
  - question: |
      Why does `Set` loop with `for i := range *b` and write `(*b)[i].val = val`,
      instead of writing `for _, e := range *b { if e.key == key { e.val = val } }`?
    options:
      - text: The second version doesn't compile
      - text: '`e` is a copy of the entry, so assigning to `e.val` changes the copy and the map never sees the update'
        correct: true
      - text: The second version is slower
      - text: Ranging over a slice of structs is not allowed
    explanation: |
      A `range` loop's value variable holds a *copy* of each element. To modify the
      element stored in the slice, index into the slice. This is one of the most
      common Go gotchas.
  - question: A hashmap with 8 buckets holds 8,000 entries, spread evenly. Roughly how many key comparisons does a `Get` do?
    options:
      - text: '1'
      - text: About 13 (log₂ 8,000)
      - text: About 1,000
        correct: true
      - text: '8,000'
    explanation: |
      Each bucket holds about 8,000 / 8 = 1,000 entries, and `Get` scans one bucket.
      That's still 8 times faster than a plain slice, but nowhere near O(1). Too few
      buckets is exactly the problem that resizing fixes.
---

Let's build a hashmap for the studio: username → player level. We'll use
**separate chaining**, the simplest collision strategy: each bucket is a small slice
of entries, and keys that hash to the same bucket just share it.

```
buckets (8 of them, hashed with FNV-1a)
  0: [ {kai 40} ]
  1: []
  ...
  5: []
  6: [ {mira 12} {bo 3} ]    <- "mira" and "bo" collide
  7: []
```

## The types

For now, keys are strings (usernames) and the value type is generic. We'll lift the
string-only restriction at the end of the chapter.

```go
type entry[V any] struct {
	key string
	val V
}

type HashMap[V any] struct {
	buckets [][]entry[V]
	size    int
}

func NewHashMap[V any](n int) *HashMap[V] {
	return &HashMap[V]{buckets: make([][]entry[V], n)}
}
```

Unlike our BST, the zero value isn't usable: with zero buckets, `hash % 0` would
panic with a division by zero. So we provide a constructor.

## Set, Get and Delete

All three start the same way: hash the key, pick a bucket, scan that bucket.

```go
package main

import (
	"fmt"
	"hash/fnv"
)

type entry[V any] struct {
	key string
	val V
}

type HashMap[V any] struct {
	buckets [][]entry[V]
	size    int
}

func NewHashMap[V any](n int) *HashMap[V] {
	return &HashMap[V]{buckets: make([][]entry[V], n)}
}

func (m *HashMap[V]) bucket(key string) *[]entry[V] {
	h := fnv.New64a()
	h.Write([]byte(key))
	return &m.buckets[h.Sum64()%uint64(len(m.buckets))]
}

func (m *HashMap[V]) Set(key string, val V) {
	b := m.bucket(key)
	for i := range *b {
		if (*b)[i].key == key {
			(*b)[i].val = val // update existing
			return
		}
	}
	*b = append(*b, entry[V]{key, val})
	m.size++
}

func (m *HashMap[V]) Get(key string) (V, bool) {
	for _, e := range *m.bucket(key) {
		if e.key == key {
			return e.val, true
		}
	}
	var zero V
	return zero, false
}

func (m *HashMap[V]) Delete(key string) {
	b := m.bucket(key)
	for i, e := range *b {
		if e.key == key {
			// Order inside a bucket doesn't matter: move the last entry into the gap.
			last := len(*b) - 1
			(*b)[i] = (*b)[last]
			*b = (*b)[:last]
			m.size--
			return
		}
	}
}

func (m *HashMap[V]) Len() int { return m.size }

func main() {
	levels := NewHashMap[int](8)
	levels.Set("mira", 12)
	levels.Set("kai", 40)
	levels.Set("bo", 3)
	levels.Set("mira", 13) // level up!

	fmt.Println(levels.Get("mira"))
	fmt.Println(levels.Get("zed"))

	levels.Delete("kai")
	_, ok := levels.Get("kai")
	fmt.Println(ok, levels.Len())
}
```

Output:

```
13 true
0 false
false 2
```

## Things to notice

**`bucket` returns a pointer to the slice.** `Set` and `Delete` need to change the
bucket's length (via `append` or re-slicing). A slice header is a small struct
(pointer, length, capacity), so modifying a *copy* of it wouldn't stick. With a
pointer, `*b = append(*b, ...)` updates the real bucket inside `m.buckets`.

**`Set` indexes into the bucket.** `(*b)[i].val = val` writes to the element in the
slice. A `for _, e := range` loop would give you a copy.

**`Get` uses the comma-ok idiom**, returning the zero value and `false` when the key
is missing, just like `v, ok := m[key]` on a built-in map.

**`Delete` swaps with the last entry** instead of shifting everything down. Bucket
order means nothing, so this makes deletion O(1) once the key is found.

## How fast is it?

Hashing is O(length of the key), which we treat as constant. Then we scan one bucket.
If n entries spread evenly over k buckets, each bucket holds about n/k entries. With a
fixed 8 buckets and a million players, that's 125,000 entries per bucket, which is
basically a slow slice. For O(1) we need the bucket count to **grow with n**. That's
the resizing lesson, coming up after we look at collisions more closely.
