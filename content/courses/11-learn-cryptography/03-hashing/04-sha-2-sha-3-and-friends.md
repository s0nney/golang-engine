---
title: SHA-2, SHA-3 and Friends
quiz:
  - question: A legacy system identifies files by their SHA-1 hash. Why is that a problem today?
    options:
      - text: SHA-1 digests are too long to store
      - text: Practical SHA-1 collisions have been demonstrated since 2017, so an attacker can craft two different files with the same ID
        correct: true
      - text: SHA-1 is not deterministic
      - text: SHA-1 can be reversed to recover the file
    explanation: |
      The SHAttered attack (2017) produced two different PDFs with the same SHA-1, and
      later attacks made chosen-prefix collisions affordable. Anything that relies on
      SHA-1 collision resistance, such as signatures or content addressing, is
      broken. Preimages are still out of reach, but that's cold comfort.
  - question: Which of these is *not* a reason to choose SHA-512/256 or SHA3-256 over plain SHA-256?
    options:
      - text: They aren't vulnerable to length-extension attacks
      - text: SHA-512/256 is often faster than SHA-256 on 64-bit CPUs without SHA extensions
      - text: SHA-256 has known practical collision attacks
        correct: true
    explanation: |
      SHA-256 is unbroken and fine for fingerprints, checksums and signatures. The
      other two are genuine reasons to pick a different member of the family, and
      you'll see length extension in the next chapter.
---

SHA-256 isn't the only hash in Go's standard library. Here's the family tree, and
which branches are safe to sit on.

## The standard library's hashes

```go
package main

import (
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"fmt"
)

func main() {
	msg := []byte("keybox")
	fmt.Printf("SHA-256     %x\n", sha256.Sum256(msg))
	fmt.Printf("SHA-512/256 %x\n", sha512.Sum512_256(msg))
	fmt.Printf("SHA3-256    %x\n", sha3.Sum256(msg))
	fmt.Printf("SHAKE256    %x\n", sha3.SumSHAKE256(msg, 32))
}
```

```
SHA-256     bf3d3b21c23b934be88eb74d9a756545963ec7e8f803fac4ec949ee5ec21140f
SHA-512/256 1470987fde7337b1e5d21394579259c4d71efdc237f69c4669fd063bff1cae9f
SHA3-256    c4f3843e785d31c929159c18b8ce7a81e3038c52745628bb3625a782815544ee
SHAKE256    a4a5ad58fcbeb5ba66a9952ce15133cd84076d8d60d18373fa15bb737bc28f97
```

Four 32-byte digests of the same input, completely unrelated to one another. Each is a
different function, so a digest is only meaningful alongside the name of the algorithm
that made it. That's why formats like `sha256:bf3d3b...` label their hashes.

## SHA-2

`crypto/sha256` and `crypto/sha512` implement the **SHA-2** family (2001), the
workhorse of today's internet. TLS certificates, Git's newer object format, Bitcoin
and most signatures use it.

- **SHA-256**: 32-byte digest, 64-byte blocks, 32-bit arithmetic. Most modern CPUs
  have dedicated instructions for it.
- **SHA-512**: 64-byte digest, 128-byte blocks, 64-bit arithmetic.
- **SHA-384** and **SHA-512/256** (`sha512.Sum384`, `sha512.Sum512_256`) are SHA-512
  with different starting constants and a **truncated** output. Because part of the
  internal state is never revealed, they resist the length-extension attack that
  affects SHA-256 and SHA-512 (next chapter).
- **SHA-224** is the same idea for SHA-256, rarely used.

SHA-2 is unbroken. There's no reason to migrate away from SHA-256 for fingerprints or
signatures.

## SHA-3

`crypto/sha3` (standard library since Go 1.24) implements **SHA-3**, standardized in
2015 as a backup in case SHA-2 ever fell. It's built on a completely different design,
the Keccak **sponge**, so an attack on one family is unlikely to carry over to the other.

- `sha3.Sum256`, `sha3.Sum512` and friends are drop-in fixed-size hashes. They're
  immune to length extension by design.
- **SHAKE128** and **SHAKE256** are *extendable-output functions*: you choose the output
  length. `sha3.NewSHAKE256()` gives you a hash you `Write` into and then `Read` as many
  bytes out of as you like. ML-KEM and ML-DSA, the post-quantum algorithms you'll meet
  later, use SHAKE internally.

SHA-3 is slower than SHA-256 in software on most CPUs, which is the main reason it
hasn't replaced it.

## Outside the standard library

- **BLAKE2b / BLAKE2s** (`golang.org/x/crypto/blake2b`): very fast in software, with
  built-in keyed mode. Argon2 uses BLAKE2b internally.
- **BLAKE3**: faster still and parallelisable; third-party Go packages only.

## Retired: MD5 and SHA-1

`crypto/md5` and `crypto/sha1` stay in the standard library so Go can talk to old
formats and protocols, but their **collision resistance is broken**:

- MD5 collisions take seconds on a laptop. In 2012 the Flame malware forged a
  Microsoft code-signing certificate using an MD5 collision.
- SHA-1 fell in public in 2017, when Google and CWI Amsterdam produced two different
  PDFs with the same SHA-1 ("SHAttered"). Chosen-prefix collisions followed in 2020.

For anything new, use SHA-256 (or SHA-512/256 or SHA3-256 when you need resistance to
length extension). Seeing MD5 or SHA-1 in a security context during code review is
always worth a question.
