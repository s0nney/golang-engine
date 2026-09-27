---
title: Collisions
quiz:
  - question: |
      In the linear-probing example, `"nova"` hashes to slot 7. Slots 7, 0 and 1 are
      full. Where does it go, and how does `Get("nova")` find it?
    options:
      - text: It's rejected, because its home slot is taken
      - text: It overwrites `"bo"` in slot 7
      - text: 'It goes in slot 2; `Get` starts at 7 and steps forward, wrapping around, until it finds `"nova"`'
        correct: true
      - text: It goes in a separate overflow list attached to slot 7
    explanation: |
      Linear probing tries the next slot, then the next, wrapping around with `% len`.
      Lookups follow exactly the same path, so they find the key where insertion
      put it, or stop at the first empty slot if it isn't there.
  - question: |
      With open addressing, why can't `Delete` simply mark a slot empty?
    options:
      - text: Go doesn't allow modifying slice elements
      - text: An empty slot would cut the probe chain, and lookups for keys stored after it would stop early and report "not found"
        correct: true
      - text: It would make the table resize
      - text: Deleting from open addressing tables is impossible
    explanation: |
      `Get` stops at the first empty slot. If you empty a slot in the middle of a
      chain, everything placed after it becomes unreachable. That's why deletions
      leave a *tombstone* that lookups skip over but insertions can reuse.
---

Two different usernames, one bucket. Collisions are unavoidable: there are endless
possible usernames and only so many buckets. What matters is how a hashmap handles
them. There are two big families.

## Strategy 1: separate chaining

This is what you built in the last lesson. Each bucket holds a list of entries, and
colliding keys just join the list.

- **Pros:** simple, and deletion is easy (remove from the list).
- **Cons:** every bucket is a separate allocation, and following it to another
  place in memory is slow for the CPU cache.

## Strategy 2: open addressing

Store the entries directly in one flat slice of **slots**, one entry per slot. If a
key's home slot is taken, *probe* for another slot using a fixed rule. The simplest
rule is **linear probing**: try the next slot, then the next, wrapping around at the end.

```go
package main

import (
	"fmt"
	"hash/fnv"
)

type slot struct {
	key   string
	val   int
	inUse bool
}

type ProbeMap struct {
	slots []slot
	size  int
}

func (m *ProbeMap) home(key string) int {
	h := fnv.New64a()
	h.Write([]byte(key))
	return int(h.Sum64() % uint64(len(m.slots)))
}

// Set assumes the table never fills up; the next lesson fixes that.
func (m *ProbeMap) Set(key string, val int) {
	i := m.home(key)
	for m.slots[i].inUse && m.slots[i].key != key {
		i = (i + 1) % len(m.slots) // linear probing: try the next slot
	}
	if !m.slots[i].inUse {
		m.size++
	}
	m.slots[i] = slot{key, val, true}
}

func (m *ProbeMap) Get(key string) (int, bool) {
	i := m.home(key)
	for m.slots[i].inUse { // an empty slot ends the search
		if m.slots[i].key == key {
			return m.slots[i].val, true
		}
		i = (i + 1) % len(m.slots)
	}
	return 0, false
}

func main() {
	m := &ProbeMap{slots: make([]slot, 8)}
	for _, name := range []string{"mira", "bo", "ash", "luna", "nova"} {
		fmt.Printf("%-4s home=%d\n", name, m.home(name))
		m.Set(name, len(name))
	}
	for i, s := range m.slots {
		fmt.Printf("%d %q\n", i, s.key)
	}
	fmt.Println(m.Get("nova"))
}
```

Output:

```
mira home=6
bo   home=6
ash  home=7
luna home=7
nova home=7
0 "ash"
1 "luna"
2 "nova"
3 ""
4 ""
5 ""
6 "mira"
7 "bo"
4 true
```

Look at what happened. `"mira"` took its home slot 6, so `"bo"` (also home 6) moved
to 7. `"ash"` wanted 7 and wrapped around to 0. `"luna"` wanted 7, found 7 and 0 full,
and landed in 1. `"nova"` probed 7, 0 and 1 and settled in 2. To find `"nova"`, `Get`
retraces that exact path: start at its home slot and step forward until it either
finds the key or hits an **empty** slot, which proves the key isn't there.

The `% len(m.slots)` is what makes the probe wrap from the last slot back to slot 0.

## Clustering

Notice how the keys formed one solid run in slots 6, 7, 0, 1, 2. That's
**primary clustering**: once a run forms, any key hashing anywhere into it lands at
the end and makes it longer. Long runs mean long probes. Fancier probing schemes
(quadratic probing, double hashing) jump around to break clusters up. Keeping the
table from getting too full matters even more, and that's next lesson's topic.

## Deleting with tombstones

You can't simply clear a slot in an open-addressing table. If you emptied slot 0
(`"ash"`), a later `Get("nova")` would start at 7, see `"bo"`, step to 0, find it
empty and give up, even though `"nova"` is sitting in slot 2.

The fix is a **tombstone**: a marker meaning "something used to be here". Lookups
treat a tombstone as occupied and keep probing past it. Inserts are allowed to reuse
it. Tombstones pile up over time, so tables occasionally rebuild themselves to clear
them out.

## Which one wins?

| | Chaining | Open addressing |
|---|---|---|
| Memory layout | Pointers to separate lists | One flat slice |
| Cache friendliness | Worse | Better |
| Deletion | Easy | Needs tombstones |
| Very full tables | Degrade gracefully | Degrade sharply |

Modern high-performance hashmaps, including Go's own since Go 1.24, use open
addressing, because on today's hardware memory access patterns matter more than
almost anything else. You'll see how Go does it in two lessons.
