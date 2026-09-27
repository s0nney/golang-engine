---
title: O(1) and O(n)
quiz:
  - question: |
      What's the Big O of this function, where `n` is `len(counts)`?

      ```go
      func firstAndLast(counts []int) (int, int) {
      	return counts[0], counts[len(counts)-1]
      }
      ```
    options:
      - text: O(1)
        correct: true
      - text: O(n)
      - text: O(2)
    explanation: |
      Indexing a slice and calling `len` are both constant time, no matter how big
      the slice is. Two constant-time steps is still O(1). There is no "O(2)":
      constants are dropped.
  - question: Why is looking up a key in a Go map considered O(1) on average?
    options:
      - text: Maps are stored in sorted order, so lookups are instant
      - text: The key is hashed straight to the bucket where its value lives, so the work doesn't grow with the map's size
        correct: true
      - text: Go caches the last key you looked up
    explanation: |
      A hash map computes where a key lives from the key itself, then checks a
      small bucket. The size of the map doesn't change that, so lookups are
      constant time on average. (Unlucky collisions make individual lookups
      slower, which is why it's "on average".)
---

Let's put the two simplest and most common classes under the microscope.

## O(1): constant time

An **O(1)** operation takes the same amount of time no matter how big the
input is. Ten influencers or ten million, it doesn't care.

Examples in Go:

- Indexing a slice or array: `counts[42]`
- `len(s)` and `cap(s)`: the length is stored in the slice header
- Reading or writing a struct field
- Arithmetic and comparisons on numbers
- Map lookup, insert and delete (on average)
- Appending to a slice (*amortized*, which we'll unpack in chapter 6)

```go
func topInfluencer(sorted []Influencer) Influencer {
	return sorted[len(sorted)-1]
}
```

If `sorted` is already in ascending order, grabbing the biggest account is one
index operation. O(1).

O(1) doesn't mean "fast", it means "doesn't grow". A function that always
sleeps for 10 seconds is O(1). It's just a big constant.

## O(n): linear time

An **O(n)** algorithm does a fixed amount of work per input item. Double the
input and you double the time. Any single loop over the input, doing O(1)
work per iteration, is O(n):

```go
package main

import "fmt"

type Influencer struct {
	Handle    string
	Followers int
}

func countAbove(infs []Influencer, threshold int) int {
	n := 0
	for _, inf := range infs {
		if inf.Followers > threshold {
			n++
		}
	}
	return n
}

func main() {
	infs := []Influencer{
		{"ava", 120_000}, {"bo", 800}, {"cy", 56_000}, {"dee", 2_300_000},
	}
	fmt.Println(countAbove(infs, 50_000))
}
```

This prints `3`. Each influencer is looked at once, so the work is
proportional to `len(infs)`.

## Hidden loops

Watch out for function calls that *hide* a loop. These all look like one line
but are O(n):

- `slices.Contains(s, x)` and `slices.Index(s, x)` scan the slice.
- `strings.Contains(text, word)` scans the text.
- `copy(dst, src)` and `slices.Clone(s)` touch every element.
- `s = append([]T{x}, s...)`, prepending, copies the whole slice.

So this innocent-looking loop is really O(n × m), where `n` is `len(handles)`
and `m` is `len(banned)`, because it hides a scan inside a loop. If both lists
are about the same size, that's O(n²):

```go
for _, h := range handles {
	if slices.Contains(banned, h) { // O(len(banned)) each time
		fmt.Println("blocked:", h)
	}
}
```

Build a set from `banned` first (a `map[string]struct{}`), and each check
becomes O(1), making the whole thing O(n) again. Knowing the cost of the
functions you call is half of Big O analysis.
