---
title: What Are Data Structures?
quiz:
  - question: Clout needs to look up an influencer by handle, millions of times a day. Which structure fits best?
    options:
      - text: A slice, scanned from the start each time
      - text: A map keyed by handle
        correct: true
      - text: A slice sorted by follower count
    explanation: |
      A map finds a key in O(1) on average. Scanning a slice is O(n), and sorting
      by follower count doesn't help you find a *handle*. The right structure
      depends on the operations you'll do most.
  - question: What's the relationship between a data structure and an algorithm?
    options:
      - text: They're unrelated topics
      - text: A data structure organises data so that the algorithms you run on it are efficient
        correct: true
      - text: A data structure is just an algorithm that stores data in a file
    explanation: |
      The structure decides which operations are cheap. Binary search needs a
      sorted slice; O(1) lookups need a hash map. Choosing the structure is often
      the biggest algorithmic decision you make.
---

The first half of this course was about **algorithms**: the steps. The rest is
about **data structures**: the way data is laid out in memory so those steps
can be efficient.

## A way of organising data

A **data structure** is a particular way of storing and organising data, along
with the operations it supports and what each one costs. You've been using
them since your first Go program:

| structure | good at | bad at |
|-----------|---------|--------|
| array / slice | O(1) access by index, fast iteration | O(n) insert or delete at the front |
| map | O(1) average lookup by key | no ordering, more memory per item |
| struct | grouping related fields | not a collection at all |

None of these is "the best". Each makes some operations cheap by making others
expensive. That's the recurring theme: **every data structure is a trade-off**.

## Choosing is algorithm design

Think back to the duplicate-follower-count check. The algorithm barely
changed between the O(n²) and O(n) versions. What changed was the *structure*:
adding a map turned "have I seen this before?" from an O(n) scan into an O(1)
lookup.

Here's another Clout example. The dashboard needs both "top 10 by followers"
and "look up by handle". One structure can't do both well, so we keep two:

```go
package main

import (
	"cmp"
	"fmt"
	"slices"
)

type Influencer struct {
	Handle    string
	Followers int
}

type Directory struct {
	byHandle map[string]*Influencer
	ranked   []*Influencer // sorted by followers, biggest first
}

func NewDirectory(infs []Influencer) *Directory {
	d := &Directory{byHandle: make(map[string]*Influencer, len(infs))}
	for i := range infs {
		p := &infs[i]
		d.byHandle[p.Handle] = p
		d.ranked = append(d.ranked, p)
	}
	slices.SortFunc(d.ranked, func(a, b *Influencer) int {
		return cmp.Compare(b.Followers, a.Followers)
	})
	return d
}

func main() {
	d := NewDirectory([]Influencer{{"ava", 90_000}, {"bo", 1_200}, {"cy", 450_000}})
	fmt.Println(d.byHandle["bo"].Followers)
	fmt.Println(d.ranked[0].Handle)
}
```

```
1200
cy
```

Both structures point at the same `Influencer` values, so there's one copy of
the data and two ways to reach it. Building the directory costs O(n log n)
once; after that, lookups are O(1) and the top account is O(1).

Note the `p := &infs[i]`. Taking the address of the *slice element* matters.
Writing `for _, inf := range infs { d.byHandle[inf.Handle] = &inf }` would
store pointers to the loop variable, which is a *copy* of each element, not
the element itself. Since Go 1.22 each iteration gets a fresh copy, so the
pointers don't all collide any more, but they still point at copies: updating
`d.byHandle["bo"].Followers` wouldn't change `infs`.

## What's coming

In the rest of the course we'll look under the hood of Go's arrays and slices,
then build three classic structures ourselves:

- **Stacks**: last in, first out. Undo history, bracket matching.
- **Queues**: first in, first out. Job processing, rate limiting.
- **Linked lists**: chains of nodes connected by pointers.

Each one is simple. The skill is knowing which operation costs what, and
picking the structure whose cheap operations match your workload.
