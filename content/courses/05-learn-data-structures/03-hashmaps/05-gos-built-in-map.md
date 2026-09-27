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
exercise:
  starter: |
    package main

    import "fmt"

    // lootReport counts how many times each item dropped and returns lines like
    // "sword x3", most common first. Items with the same count are in
    // alphabetical order. drops must not be modified.
    func lootReport(drops []string) []string {
    	var lines []string
    	// ? count with a map, then sort the items: ranging over a map gives
    	// a different order every time
    	return lines
    }

    // byGuild inverts a player → guild map into guild → players, with each
    // guild's players sorted alphabetically.
    func byGuild(guildOf map[string]string) map[string][]string {
    	var guilds map[string][]string
    	// ?
    	return guilds
    }

    func main() {
    	drops := []string{"potion", "sword", "potion", "gem", "sword", "potion", "bow"}
    	fmt.Println(lootReport(drops))
    	// want: [potion x3 sword x2 bow x1 gem x1]

    	g := byGuild(map[string]string{"mira": "owls", "kai": "foxes", "bo": "owls", "ada": "owls"})
    	fmt.Println(g["owls"], g["foxes"], len(g))
    	// want: [ada bo mira] [kai] 2
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"maps"
    	"slices"
    )

    // lootReport counts how many times each item dropped and returns lines like
    // "sword x3", most common first. Items with the same count are in
    // alphabetical order. drops must not be modified.
    func lootReport(drops []string) []string {
    	counts := map[string]int{}
    	for _, d := range drops {
    		counts[d]++
    	}
    	items := slices.Collect(maps.Keys(counts))
    	slices.SortFunc(items, func(a, b string) int {
    		return cmp.Or(cmp.Compare(counts[b], counts[a]), cmp.Compare(a, b))
    	})
    	lines := make([]string, 0, len(items))
    	for _, it := range items {
    		lines = append(lines, fmt.Sprintf("%s x%d", it, counts[it]))
    	}
    	return lines
    }

    // byGuild inverts a player → guild map into guild → players, with each
    // guild's players sorted alphabetically.
    func byGuild(guildOf map[string]string) map[string][]string {
    	guilds := make(map[string][]string)
    	for player, guild := range guildOf {
    		guilds[guild] = append(guilds[guild], player)
    	}
    	for _, players := range guilds {
    		slices.Sort(players)
    	}
    	return guilds
    }

    func main() {
    	drops := []string{"potion", "sword", "potion", "gem", "sword", "potion", "bow"}
    	fmt.Println(lootReport(drops))
    	// want: [potion x3 sword x2 bow x1 gem x1]

    	g := byGuild(map[string]string{"mira": "owls", "kai": "foxes", "bo": "owls", "ada": "owls"})
    	fmt.Println(g["owls"], g["foxes"], len(g))
    	// want: [ada bo mira] [kai] 2
    }
  tests: |
    package main

    import (
    	"maps"
    	"slices"
    	"testing"
    )

    func TestLootReport(t *testing.T) {
    	tests := []struct {
    		drops []string
    		want  []string
    	}{
    		{[]string{"potion", "sword", "potion", "gem", "sword", "potion", "bow"},
    			[]string{"potion x3", "sword x2", "bow x1", "gem x1"}},
    		{[]string{"zap", "axe", "mug", "axe", "zap", "mug"},
    			[]string{"axe x2", "mug x2", "zap x2"}},
    		{[]string{"gem"}, []string{"gem x1"}},
    		{nil, []string{}},
    	}
    	for _, tt := range tests {
    		in := slices.Clone(tt.drops)
    		for range 5 { // map order is random: the answer must not be
    			got := lootReport(in)
    			if !slices.Equal(got, tt.want) {
    				t.Fatalf("lootReport(%q) = %q, want %q", tt.drops, got, tt.want)
    			}
    		}
    		if !slices.Equal(in, tt.drops) {
    			t.Errorf("lootReport changed its input to %q", in)
    		}
    	}
    }

    func TestByGuild(t *testing.T) {
    	guildOf := map[string]string{
    		"mira": "owls", "kai": "foxes", "bo": "owls", "ada": "owls", "zed": "foxes", "lu": "bats",
    	}
    	orig := maps.Clone(guildOf)
    	want := map[string][]string{
    		"owls":  {"ada", "bo", "mira"},
    		"foxes": {"kai", "zed"},
    		"bats":  {"lu"},
    	}
    	for range 5 {
    		got := byGuild(guildOf)
    		if !maps.EqualFunc(got, want, slices.Equal) {
    			t.Fatalf("byGuild(%v) = %v, want %v", guildOf, got, want)
    		}
    	}
    	if !maps.Equal(guildOf, orig) {
    		t.Errorf("byGuild changed its input to %v", guildOf)
    	}

    	empty := byGuild(map[string]string{})
    	if empty == nil {
    		t.Fatal("byGuild of an empty map returned a nil map; return an empty one so callers can add to it")
    	}
    	empty["newbies"] = []string{"you"} // must not panic
    }
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

## Your turn: loot reports

The studio's analytics bot posts a loot summary after every raid, and players
keep complaining that the order changes each time. Guess why.

Complete two functions:

- **`lootReport(drops)`** counts each item in `drops` with a map and returns
  lines like `"potion x3"`, most common first, with ties in alphabetical order.
  Collect the keys with `slices.Collect(maps.Keys(counts))` and sort them with
  `slices.SortFunc`. `cmp.Or(cmp.Compare(...), cmp.Compare(...))` is a neat way
  to say "compare counts, and if they're equal, compare names". Don't modify
  `drops`.
- **`byGuild(guildOf)`** inverts a player → guild map into guild → sorted list of
  players. Remember the nil-map gotcha: the result must be a real (non-nil) map,
  even for empty input. Appending to a missing key's `nil` slice is fine, though.

The tests call each function several times, so an answer that happens to come
out in the right order by luck won't pass.

## Complexity

Lookups, inserts and deletes are O(1) on average. The worst case is still O(n) in
theory if every key collides, but Go seeds its hash function randomly for each map,
so nobody can predict which keys will collide. More on that in the next lesson.

## Further reading

- [Faster Go maps with Swiss Tables](https://go.dev/blog/swisstable), the Go blog
  post that introduced the new design.
- [Go by Example: Maps](https://gobyexample.com/maps)
