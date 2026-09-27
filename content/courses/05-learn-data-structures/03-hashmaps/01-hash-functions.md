---
title: Hash Functions
quiz:
  - question: Which property must **every** hash function used by a hashmap have?
    options:
      - text: It never produces the same output for two different inputs
      - text: The same input always produces the same output
        correct: true
      - text: The output can't be reversed to find the input
      - text: The output is always a prime number
    explanation: |
      Determinism is non-negotiable: if `hash("mira")` changed between `Set` and `Get`,
      you'd look in the wrong bucket. Collisions can't be avoided (there are more
      possible keys than outputs), and being irreversible is a *cryptographic* property
      that hashmaps don't need.
  - question: |
      With the `sumHash` from this lesson and 8 buckets, which of these usernames lands in the same bucket as `"stone"`?
    options:
      - text: '`"stones"`'
      - text: '`"store"`'
      - text: '`"notes"`'
        correct: true
      - text: '`"stun"`'
    explanation: |
      `sumHash` adds up the bytes, and addition doesn't care about order. `"notes"`
      has exactly the same letters as `"stone"`, so it has the same sum and the same
      bucket. That's why good hash functions mix in each byte's position.
  - question: Why does the FNV-1a loop multiply by a prime after each byte?
    options:
      - text: To make the hash reversible
      - text: So each byte affects all the bits of the result and the order of the bytes matters
        correct: true
      - text: To keep the result below the number of buckets
      - text: Multiplication is faster than addition
    explanation: |
      XOR alone would behave much like the sum hash. Multiplying after each step
      smears each byte across the whole 64-bit state, so `"stone"` and `"onset"` end up
      far apart.
---

The leaderboard tree gives us O(log n) lookups. For the studio's most common question,
"give me the profile for username `mira`", we can do even better: O(1) on average,
with a **hashmap**. You've used Go's built-in `map` plenty. This chapter builds one
from scratch so you know what's going on inside.

## The idea

A hashmap stores entries in a slice of **buckets**. To decide which bucket a key goes in:

1. Run the key through a **hash function**, which turns it into a big number.
2. Take that number modulo the bucket count to get an index.

```
"mira" --hash--> 15734435404047900358 --% 8--> bucket 6
```

To look the key up later, you hash it again, land on the same bucket, and look
there. No searching through everything.

## What makes a good hash function?

- **Deterministic**: the same key always gives the same hash. Absolutely required.
- **Uniform**: keys spread evenly across buckets, so no bucket gets crowded.
- **Fast**: it runs on every single lookup.

## A bad hash function

Here's the simplest thing that could work: add up the bytes.

```go
func sumHash(s string) uint64 {
	var h uint64
	for i := range len(s) {
		h += uint64(s[i])
	}
	return h
}
```

It's deterministic and fast, but not uniform. Anagrams always collide, and short
usernames all produce small sums that clump together.

## A good one: FNV-1a

FNV-1a is a classic non-cryptographic hash. For each byte, XOR it into the state and
then multiply by a special prime. The multiply spreads every byte's influence across
all 64 bits, and it makes the order of bytes matter. Go ships it in `hash/fnv`.

```go
package main

import (
	"fmt"
	"hash/fnv"
)

func sumHash(s string) uint64 {
	var h uint64
	for i := range len(s) {
		h += uint64(s[i])
	}
	return h
}

func fnvHash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func main() {
	const buckets = 8
	for _, name := range []string{"stone", "onset", "tones"} {
		fmt.Println(name, sumHash(name)%buckets, fnvHash(name)%buckets)
	}
}
```

Output:

```
stone 1 2
onset 1 0
tones 1 6
```

With `sumHash` all three anagrams pile into bucket 1. FNV-1a scatters them.
`h.Write` on an FNV hasher never returns an error, which is why we don't check it.

The core of FNV-1a is short enough to write by hand, and it's worth seeing:

```go
func fnv1a(s string) uint64 {
	h := uint64(14695981039346656037) // "offset basis"
	for i := range len(s) {
		h ^= uint64(s[i])
		h *= 1099511628211 // "FNV prime"
	}
	return h
}
```

`uint64` arithmetic wraps around on overflow instead of failing, which is exactly
what a hash function wants.

## Hashing for hashmaps vs hashing for security

Don't confuse these with *cryptographic* hashes like SHA-256 (`crypto/sha256`).
Those are designed so nobody can reverse them or deliberately find collisions, and
they're much slower. Hashmaps need speed and a good spread, not secrecy. There's
one security wrinkle, though: if an attacker knows your hash function, they can
send keys that all collide. We'll see how Go defends against that at the end of
this chapter.
