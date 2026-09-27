---
title: Cryptographic Hash Functions
quiz:
  - question: Which property would you need to break to make two *different* Keybox bundles that share one SHA-256 fingerprint, when you get to choose both bundles?
    options:
      - text: Preimage resistance
      - text: Second-preimage resistance
      - text: Collision resistance
        correct: true
      - text: Determinism
    explanation: |
      A collision is *any* pair of different inputs with the same hash, both chosen by
      the attacker. That's the easiest of the three to attack: thanks to the birthday
      paradox it costs about 2^(n/2) work, 2^128 for SHA-256. Second preimage is
      harder: the first input is fixed and you must match it.
  - question: A hash map uses FNV-1a. Why not use SHA-256 there instead, to be safe?
    options:
      - text: SHA-256 isn't deterministic
      - text: SHA-256 outputs can't be used as array indexes
      - text: A hash map needs speed and good spread, not resistance to deliberate attacks on the hash; SHA-256 would be much slower for no benefit (and Go's maps already use a randomly seeded hash to resist flooding)
        correct: true
      - text: FNV-1a has stronger collision resistance
    explanation: |
      Non-cryptographic hashes are built for speed. Cryptographic hashes are built so
      that nobody can find collisions or preimages even when trying. Use each for its
      own job.
---

You met hash functions in
[Hash Functions](/courses/learn-data-structures/hashmaps/hash-functions), where FNV-1a
turned keys into bucket numbers. Those hashes only need to be fast and spread keys
evenly. A **cryptographic** hash function must also hold up against someone
*deliberately* trying to break it.

## What it does

A cryptographic hash takes any amount of data and produces a short, fixed-size
**digest**. SHA-256 always gives 32 bytes, whether you feed it one byte or a 40 GB
disk image:

```go
package main

import (
	"crypto/sha256"
	"fmt"
)

func main() {
	for _, s := range []string{"keybox", "keyboy", "Keybox"} {
		fmt.Printf("%-7s %x\n", s, sha256.Sum256([]byte(s)))
	}
}
```

```
keybox  bf3d3b21c23b934be88eb74d9a756545963ec7e8f803fac4ec949ee5ec21140f
keyboy  5e4859abedcf9a8ba51c60abc85e31728c8752b547ccc39391716fbb38e8ad95
Keybox  16adbe36a83d71f73ef2dce49c0f8d36205ce8a9bdc2310fcc81bd599f19ba95
```

One changed letter changes about half of the output bits, with no visible pattern.
That's the **avalanche effect**, and it's what you'd expect if the output were random.

## The three security properties

Think of a secure hash as a function that behaves like a giant random lookup table
that everyone can consult but nobody can predict. Concretely, for an n-bit hash:

1. **Preimage resistance.** Given a digest `h`, you can't find *any* input with
   `hash(x) == h`, short of about 2^n guesses. The hash is one-way.
2. **Second-preimage resistance.** Given an input `x`, you can't find a *different*
   `y` with the same hash, short of about 2^n guesses.
3. **Collision resistance.** You can't find *any* two different inputs with the same
   hash, short of about **2^(n/2)** guesses.

Collisions are cheaper because of the **birthday paradox**: in a room of just 23
people there's a 50% chance two share a birthday, because you're comparing every pair.
Hash 2^128 random inputs and you probably have two with the same 256-bit digest. That's
why a 256-bit hash gives "128-bit security" against collisions, and why 128-bit hashes
like MD5 were too small even before they were broken outright.

Of course, collisions *exist*: there are infinitely many inputs and only 2^256 outputs.
The claim is only that nobody can *find* one.

## What hashes are good for

- **Fingerprints.** A short, unforgeable name for a big piece of data: a Git commit ID,
  a Docker image digest, a public key fingerprint that Alice and Bob compare out loud.
- **Integrity against accidents**, and against attackers *if* the digest itself comes
  from a trusted place. A download page's checksum only helps if Mallory can't change
  the page too. (Signatures fix that in chapter 8.)
- **Building blocks** for MACs, key derivation, signatures and random generators.
- **Commitments.** Publish `hash(prediction || random)` now, reveal later.

## What they're not

- **Not encryption.** There's no key, and nothing to decrypt. "Hashing" a secret with
  a small input space, like a phone number or a password, is barely protection at all:
  the attacker just hashes every candidate.
- **Not a MAC.** Anyone can compute a hash, so a hash alone can't prove who made
  something. You'll see in chapter 4 why even `sha256(key || message)` isn't a safe MAC.
- **Not a password hash.** They're designed to be fast, which is exactly wrong for
  passwords (chapter 5).
