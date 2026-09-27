---
title: Load Factor and Resizing
quiz:
  - question: |
      During `grow`, why can't you just copy each old bucket into the same index of
      the new, bigger bucket slice?
    options:
      - text: The new slice has a different element type
      - text: A key's bucket is `hash % len(buckets)`, so when the length changes, most keys belong in a different bucket
        correct: true
      - text: Copying slices in Go is not allowed
      - text: You can; it's just slower
    explanation: |
      With 8 buckets, a hash of 13 goes to bucket 5. With 16 buckets, it goes to bucket 13.
      If you left it in bucket 5, `Get` would look in bucket 13 and never find it. Every
      entry must be *rehashed* into the new slice.
  - question: 'Resizing copies every entry, which is O(n). How can `Set` still be "O(1) amortized"?'
    options:
      - text: The resize runs in a background goroutine
      - text: 'Because the table doubles, resizes get rarer as it grows: n inserts trigger copies totalling less than 2n, so the average cost per insert is constant'
        correct: true
      - text: Resizing is actually O(1) because slices are pointers
      - text: It can't; the claim is wrong
    explanation: |
      Doubling means the copies cost roughly 8 + 16 + 32 + … + n, which adds up to
      less than 2n. Spread over n inserts, that's a constant amount of extra work each.
      It's the same argument that makes `append` O(1) amortized.
exercise:
  starter: |
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

    const maxLoad = 0.75

    func NewHashMap[V any]() *HashMap[V] {
    	return &HashMap[V]{buckets: make([][]entry[V], 8)}
    }

    func (m *HashMap[V]) bucket(key string) *[]entry[V] {
    	h := fnv.New64a()
    	h.Write([]byte(key))
    	return &m.buckets[h.Sum64()%uint64(len(m.buckets))]
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

    func (m *HashMap[V]) Len() int { return m.size }

    // Set stores val under key, updating it if the key already exists.
    // After adding a NEW key, if the load factor (size / buckets) is
    // above maxLoad, it calls grow.
    func (m *HashMap[V]) Set(key string, val V) {
    	// ?
    }

    // grow doubles the number of buckets and re-inserts every entry.
    func (m *HashMap[V]) grow() {
    	// ?
    }

    func main() {
    	levels := NewHashMap[int]()
    	for i := range 20 {
    		levels.Set(fmt.Sprintf("player%d", i), i)
    	}
    	levels.Set("player7", 99) // level up: an update, not a new player
    	lvl, ok := levels.Get("player7")
    	fmt.Println("player7:", lvl, ok)             // want 99 true
    	fmt.Println("players:", levels.Len())        // want 20
    	fmt.Println("buckets:", len(levels.buckets)) // want 32
    }
  solution: |
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

    const maxLoad = 0.75

    func NewHashMap[V any]() *HashMap[V] {
    	return &HashMap[V]{buckets: make([][]entry[V], 8)}
    }

    func (m *HashMap[V]) bucket(key string) *[]entry[V] {
    	h := fnv.New64a()
    	h.Write([]byte(key))
    	return &m.buckets[h.Sum64()%uint64(len(m.buckets))]
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

    func (m *HashMap[V]) Len() int { return m.size }

    func (m *HashMap[V]) Set(key string, val V) {
    	b := m.bucket(key)
    	for i := range *b {
    		if (*b)[i].key == key {
    			(*b)[i].val = val
    			return
    		}
    	}
    	*b = append(*b, entry[V]{key, val})
    	m.size++
    	if float64(m.size)/float64(len(m.buckets)) > maxLoad {
    		m.grow()
    	}
    }

    func (m *HashMap[V]) grow() {
    	old := m.buckets
    	m.buckets = make([][]entry[V], 2*len(old))
    	for _, b := range old {
    		for _, e := range b {
    			nb := m.bucket(e.key)
    			*nb = append(*nb, e)
    		}
    	}
    }

    func main() {
    	levels := NewHashMap[int]()
    	for i := range 20 {
    		levels.Set(fmt.Sprintf("player%d", i), i)
    	}
    	levels.Set("player7", 99) // level up: an update, not a new player
    	lvl, ok := levels.Get("player7")
    	fmt.Println("player7:", lvl, ok)             // want 99 true
    	fmt.Println("players:", levels.Len())        // want 20
    	fmt.Println("buckets:", len(levels.buckets)) // want 32
    }
  tests: |
    package main

    import (
    	"fmt"
    	"testing"
    )

    func TestSetAndGet(t *testing.T) {
    	m := NewHashMap[int]()
    	m.Set("mira", 12)
    	m.Set("kai", 40)
    	for key, want := range map[string]int{"mira": 12, "kai": 40} {
    		if got, ok := m.Get(key); !ok || got != want {
    			t.Errorf("Get(%q) = %d, %v; want %d, true", key, got, ok, want)
    		}
    	}
    	if _, ok := m.Get("zed"); ok {
    		t.Errorf(`Get("zed") found a value, but "zed" was never set`)
    	}
    	if m.Len() != 2 {
    		t.Errorf("Len() = %d, want 2", m.Len())
    	}
    }

    func TestUpdateDoesNotAdd(t *testing.T) {
    	m := NewHashMap[int]()
    	m.Set("mira", 12)
    	m.Set("mira", 13)
    	if got, _ := m.Get("mira"); got != 13 {
    		t.Errorf(`Get("mira") = %d after updating to 13`, got)
    	}
    	if m.Len() != 1 {
    		t.Errorf("Len() = %d after setting the same key twice, want 1", m.Len())
    	}
    }

    func TestGrowsAtLoadFactor(t *testing.T) {
    	m := NewHashMap[int]()
    	for i := range 6 {
    		m.Set(fmt.Sprintf("p%d", i), i)
    	}
    	if got := len(m.buckets); got != 8 {
    		t.Errorf("with 6 entries (load 0.75) there should still be 8 buckets, got %d", got)
    	}
    	m.Set("p6", 6) // 7/8 > 0.75
    	if got := len(m.buckets); got != 16 {
    		t.Errorf("after the 7th entry there should be 16 buckets, got %d", got)
    	}
    	before := len(m.buckets)
    	for range 5 {
    		m.Set("p6", 6) // updates must not trigger growth
    	}
    	if got := len(m.buckets); got != before {
    		t.Errorf("updating an existing key changed the bucket count from %d to %d", before, got)
    	}
    }

    func TestManyPlayers(t *testing.T) {
    	m := NewHashMap[int]()
    	for i := range 1000 {
    		m.Set(fmt.Sprintf("player%d", i), i)
    	}
    	if m.Len() != 1000 {
    		t.Fatalf("Len() = %d, want 1000", m.Len())
    	}
    	if got := len(m.buckets); got != 2048 {
    		t.Errorf("1000 entries should end up in 2048 buckets, got %d", got)
    	}
    	for i := range 1000 {
    		key := fmt.Sprintf("player%d", i)
    		if got, ok := m.Get(key); !ok || got != i {
    			t.Fatalf("Get(%q) = %d, %v after growing; want %d, true (did grow rehash every entry?)", key, got, ok, i)
    		}
    	}
    	longest := 0
    	for _, b := range m.buckets {
    		longest = max(longest, len(b))
    	}
    	if longest > 10 {
    		t.Errorf("longest bucket has %d entries; entries aren't being spread out", longest)
    	}
    }
---

Our hashmap has a fatal flaw: 8 buckets forever. Add a million players and each
bucket becomes a 125,000-entry list. The fix is to watch how full the table is and
grow it before things get crowded.

## Load factor

The **load factor** is the average number of entries per bucket:

```
load factor = entries / buckets
```

Low load factor: short chains, fast lookups, but lots of empty buckets wasting
memory. High load factor: compact, but chains get long (or, with open addressing,
probe runs get long). A common rule is to **double the bucket count** whenever the
load factor passes a threshold like 0.75.

## Rehashing

Growing isn't just a bigger slice. A key's bucket is `hash % len(buckets)`, so when
the length changes, nearly every key's bucket changes too. Growing means building a
new bucket slice and **re-inserting every entry** into it.

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
	resizes int
}

const maxLoad = 0.75

func NewHashMap[V any]() *HashMap[V] {
	return &HashMap[V]{buckets: make([][]entry[V], 8)}
}

func hashKey(key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}

func (m *HashMap[V]) bucket(key string) *[]entry[V] {
	return &m.buckets[hashKey(key)%uint64(len(m.buckets))]
}

func (m *HashMap[V]) Set(key string, val V) {
	b := m.bucket(key)
	for i := range *b {
		if (*b)[i].key == key {
			(*b)[i].val = val
			return
		}
	}
	*b = append(*b, entry[V]{key, val})
	m.size++
	if float64(m.size)/float64(len(m.buckets)) > maxLoad {
		m.grow()
	}
}

// grow doubles the bucket count and re-inserts every entry.
func (m *HashMap[V]) grow() {
	old := m.buckets
	m.buckets = make([][]entry[V], 2*len(old))
	for _, b := range old {
		for _, e := range b {
			nb := m.bucket(e.key) // a new home, based on the new length
			*nb = append(*nb, e)
		}
	}
	m.resizes++
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

func main() {
	m := NewHashMap[int]()
	for i := range 100_000 {
		m.Set(fmt.Sprintf("player%d", i), i)
	}

	longest := 0
	for _, b := range m.buckets {
		longest = max(longest, len(b))
	}
	fmt.Println("entries:", m.size)
	fmt.Println("buckets:", len(m.buckets))
	fmt.Println("resizes:", m.resizes)
	fmt.Printf("load factor: %.2f\n", float64(m.size)/float64(len(m.buckets)))
	fmt.Println("longest chain:", longest)
	fmt.Println(m.Get("player31337"))
}
```

Output:

```
entries: 100000
buckets: 262144
resizes: 15
load factor: 0.38
longest chain: 6
31337 true
```

The map started with 8 buckets and doubled 15 times to 262,144. The longest chain
among 100,000 players is just 6 entries, so every lookup is a hash plus a handful of
comparisons, no matter how big the map gets. That's O(1) on average.

A few details:

- `float64(m.size)/float64(len(m.buckets))` converts both sides first. With plain
  `int` division, `5 / 8` is `0` and the check would never fire until the load hit 1.
- The resize check happens only when a **new** key is added, not on an update.
- `grow` calls `m.bucket` *after* replacing `m.buckets`, so each entry is placed by
  the new length.

## The cost of growing

A single `Set` that triggers `grow` copies every entry: O(n). That sounds bad, but
because the table **doubles**, resizes get rarer and rarer. To reach n entries, the
total copying is about 8 + 16 + 32 + … + n < 2n. Averaged over all n inserts, each one
pays a constant amount extra. This is called **amortized O(1)**, and it's exactly the
same reasoning behind `append` on a slice.

The occasional slow insert still happens, though, and for a game server handling a
login rush, one insert that suddenly copies 10 million entries is a noticeable
stall. Go's built-in map avoids that by splitting the map into many small tables that
grow independently, as you'll see next.

## Tips

- If you know roughly how many entries you'll need, allocate enough up front and skip
  the resizes. Go's built-in map supports this: `make(map[string]int, 100_000)`.
- Growing is automatic, but shrinking usually isn't. A map that once held a million
  entries keeps its big bucket array after you delete them. `clear(m)` empties a
  built-in map, but keeps its allocated space for reuse. If you need the memory back,
  make a new map and let the old one be garbage collected.

## Your turn

The exercise's `HashMap` can already `Get`. Complete `Set` so it updates an existing key
in place, or appends a new entry and grows the map when the load factor goes **above**
0.75. Then complete `grow`, which doubles the bucket count and re-inserts every entry
into its new bucket. Starting from 8 buckets, the 7th player should trigger the first
resize to 16. Updating an existing player's level must never trigger a resize.
