---
title: Go's Built-in Map
quiz:
  - question: |
      What happens when you run this?

      ```go
      var levels map[string]int
      fmt.Println(levels["mira"])
      levels["mira"] = 12
      ```
    options:
      - text: It prints `0`, then stores 12
      - text: It prints `0`, then panics on the assignment
        correct: true
      - text: It panics on the `Println`
      - text: It doesn't compile
    explanation: |
      A nil map behaves like an empty map for *reads*: lookups return the zero value,
      and `len` is 0. But it has no storage, so *writing* to it panics with
      "assignment to entry in nil map". Always create maps with `make` or a literal.
  - question: |
      You range over the same `map[string]int` twice in one program. What's
      guaranteed about the order?
    options:
      - text: Both loops visit keys in insertion order
      - text: Both loops visit keys in sorted order
      - text: Both loops visit keys in the same order, but it's unspecified
      - text: Nothing; the two loops may visit keys in different orders
        correct: true
    explanation: |
      Go deliberately randomises where each map iteration starts, so even two loops
      over an unchanged map can differ. If you need an order, sort the keys, for
      example with `slices.Sorted(maps.Keys(m))`.
  - question: In Go's Swiss-table map, what's stored in each byte of a group's control word?
    options:
      - text: The full 64-bit hash of the slot's key
      - text: A pointer to the key
      - text: A marker for empty or deleted, or else 7 bits of the key's hash
        correct: true
      - text: The number of probes needed to reach the slot
    explanation: |
      Each control byte says whether its slot is empty, deleted or full, and for a full
      slot it holds the low 7 bits of the hash (`h2`). Comparing those 8 bytes against
      the target's `h2` all at once rules out most slots without touching the keys.
---

You've built a hashmap from scratch, so let's look at how the pros do it. Go's `map`
is a hashmap built into the language and runtime, and since **Go 1.24** it's a
**Swiss table**, a design originally from Google's C++ libraries.

## Swiss tables in a nutshell

A Swiss table uses open addressing, like the linear-probing map from the collisions
lesson, but with two clever twists.

**Slots come in groups of 8.** Each group has a 64-bit **control word**: one byte per
slot. A control byte either marks its slot as empty or deleted (a tombstone), or
holds 7 bits of the key's hash.

**The hash is split in two.** The upper 57 bits (`h1`) pick which group to start
probing at. The lower 7 bits (`h2`) go into the control byte.

To look up a key, Go computes its hash, jumps to the group chosen by `h1`, and
compares `h2` against all 8 control bytes *at once* using a few bit tricks (or SIMD
instructions on CPUs that have them). Only slots whose 7 bits match need a full key
comparison, and a random mismatch only slips through about 1 time in 128. If the
group contains an empty slot and no match, the key isn't there. If the group is
full, probing moves on to another group.

Because checking 8 slots costs about the same as checking one, Swiss tables stay fast
even when quite full (up to about 7/8 of the slots), which saves memory.

**Growth is incremental.** A Go map is actually a *directory* of small Swiss tables,
each holding at most 1,024 entries. When one table fills up, only that table grows or
splits in two. So no single insert ever has to copy a map with millions of entries,
the stall problem from the last lesson.

## Gotchas you need to know

The implementation changed, but the rules you program against didn't.

```go
package main

import (
	"fmt"
	"maps"
	"slices"
)

func main() {
	// Reads from a nil map are fine. Writes would panic.
	var banned map[string]bool
	fmt.Println(banned["mira"], len(banned))

	levels := map[string]int{"mira": 12, "kai": 40, "bo": 3, "zed": 27}

	// Missing keys give the zero value. Use comma-ok to tell "0" from "absent".
	lvl, ok := levels["ash"]
	fmt.Println(lvl, ok)

	// Iteration order is random, so sort the keys when order matters.
	for _, name := range slices.Sorted(maps.Keys(levels)) {
		fmt.Print(name, "=", levels[name], " ")
	}
	fmt.Println()

	// Deleting during a range loop is allowed.
	for name, lvl := range levels {
		if lvl < 10 {
			delete(levels, name)
		}
	}
	fmt.Println(len(levels))

	clear(levels) // remove everything
	fmt.Println(len(levels))
}
```

Output:

```
false 0
0 false
bo=3 kai=40 mira=12 zed=27 
3
0
```

- **Nil maps**: a `var m map[K]V` is nil. Reading works, writing panics. Create maps
  with `make` or a literal.
- **Random iteration order**: Go deliberately randomises it, so code can't come to
  depend on an order that might change. Printing a map with `fmt` sorts the keys for
  you, which can hide this in quick experiments.
- **Deleting while ranging is safe**. Entries added during a loop may or may not
  be visited.
- **You can't take the address of a map element**: `&levels["mira"]` doesn't
  compile, because the map may move entries around when it grows. For the same reason,
  with a `map[string]Player` you can't write `m["mira"].Level++`. Either store
  pointers (`map[string]*Player`) or copy out, modify and store back.
- **Keys must be comparable**: strings, numbers, booleans, pointers, channels,
  arrays and structs of comparable types. Slices, maps and functions can't be keys.

## Complexity

Lookups, inserts and deletes are O(1) on average. The worst case is still O(n) in
theory if every key collides, but Go seeds its hash function randomly for each map,
so nobody can predict which keys will collide. More on that in the next lesson.

## Further reading

- [Faster Go maps with Swiss Tables](https://go.dev/blog/swisstable), the Go blog
  post that introduced the new design.
- [Go by Example: Maps](https://gobyexample.com/maps)
