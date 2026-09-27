---
title: Memory-Hard KDFs
quiz:
  - question: Why do memory-hard KDFs like Argon2id resist GPU cracking better than PBKDF2?
    options:
      - text: They use a secret algorithm GPUs can't run
      - text: Each guess needs a large block of memory, and GPUs have thousands of cores but little fast memory per core, so they can run far fewer guesses in parallel
        correct: true
      - text: They produce longer outputs
      - text: They can't be computed on GPUs at all
    explanation: |
      PBKDF2 needs only a few hundred bytes per guess, so a GPU runs thousands at once.
      If each guess needs 64 MiB, the memory, not the compute, becomes the bottleneck,
      and custom hardware stops being cheap.
  - question: Keybox lets users choose passphrases up to 200 bytes long. Why would bcrypt be a poor fit?
    options:
      - text: bcrypt only accepts digits
      - text: bcrypt only uses the first 72 bytes of the password, so long passphrases lose their extra strength (or are rejected, depending on the library)
        correct: true
      - text: bcrypt produces a key that's too long for AES
      - text: bcrypt isn't a password hash
    explanation: |
      bcrypt's design limits input to 72 bytes. Some libraries silently truncate, which
      means two passphrases that share their first 72 bytes are the same password;
      Go's `x/crypto/bcrypt` returns an error instead. bcrypt also outputs a 23-byte
      hash, not a key, which makes it awkward for key derivation.
---

PBKDF2 is standardized, FIPS-approved, and in Go's standard library, which is why Keybox
uses it. But it has a weakness: it's cheap in **memory**. Password crackers exploit that
with GPUs and custom chips. The modern answer is a **memory-hard** function.

## The problem with PBKDF2

PBKDF2 is `c` sequential HMACs, needing only a few hundred bytes of state. A GPU has
thousands of cores, each able to run its own guess. Custom ASICs are even more efficient,
as Bitcoin mining showed for SHA-256. Raising the iteration count slows *you* down
exactly as much as it slows the attacker, but the attacker has thousands of times more
parallel hardware.

## Memory hardness

A memory-hard function forces every evaluation to fill and repeatedly read a large
block of memory, in an order that depends on the data, so it can't be cheaply shortcut.
Memory is expensive to put next to thousands of cores, so the attacker's parallelism
collapses. Your one login or vault unlock uses, say, 64 MiB for a fraction of a
second, which a phone handles fine.

## The options

All three live in `golang.org/x/crypto`, not the standard library. They're explained
here but can't run in this course's exercises.

**Argon2id** (RFC 9106) won the Password Hashing Competition in 2015 and is the
default recommendation for new systems. It has three parameters: time (passes over
memory), memory (in KiB) and parallelism (threads).

```go
import "golang.org/x/crypto/argon2"

// 3 passes, 64 MiB, 4 threads, 32-byte key.
key := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
```

OWASP's current minimum is 19 MiB with 2 passes and 1 thread. For a vault key that's
derived only at unlock time, you can afford much more; RFC 9106 suggests 64 MiB with 3
passes as a memory-constrained default.

**scrypt** (RFC 7914, 2009) was the first widely used memory-hard KDF, and it still
works well. Parameters: `N` (CPU/memory cost, a power of 2), `r` (block size) and `p`
(parallelism). OWASP suggests `N=2^17, r=8, p=1`, which uses 128 MiB.

**bcrypt** (1999) is not memory-hard in the modern sense, but uses 4 KiB of rapidly
changing memory that GPUs handle poorly. It remains a respectable choice for **login
password hashes**, with a cost of 10 or more. Two limits make it a bad fit for Keybox's
master key: it only reads the first **72 bytes** of the password, and it produces a
password *hash* string, not a key of the length you ask for.

## Which should you use?

| Situation | Choice |
| --- | --- |
| New system, no constraints | Argon2id |
| FIPS 140-3 compliance required | PBKDF2-HMAC-SHA256 with a high iteration count |
| Existing bcrypt hashes | Keep bcrypt, and upgrade on login if you migrate |
| Standard library only | PBKDF2 (`crypto/pbkdf2`) |

Whatever you choose, the surrounding rules are the same:

- **Store the algorithm and parameters** with every hash or vault header, like the
  self-describing `$argon2id$v=19$m=65536,t=3,p=4$...` strings, so you can upgrade.
- **Tune for your hardware.** Aim for 100 ms to 1 s per derivation on your slowest
  supported device, depending on how often you run it.
- **Protect your own servers.** A memory-hard KDF on a login endpoint is also a
  denial-of-service lever. Rate-limit and cap concurrent hash computations.

For Keybox, a real product would likely use Argon2id for the master key. Swapping it
in later is exactly what the `KDF` field in the vault header is for.
