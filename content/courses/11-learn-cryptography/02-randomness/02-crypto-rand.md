---
title: crypto/rand
quiz:
  - question: |
      Since Go 1.24, what does this code do if the operating system's random source
      fails?

      ```go
      key := make([]byte, 32)
      rand.Read(key)
      ```
    options:
      - text: Returns an error, which this code silently ignores, leaving `key` all zeros
      - text: Falls back to `math/rand`
      - text: Crashes the program irrecoverably, so a zero key can never be used by accident
        correct: true
      - text: Blocks forever
    explanation: |
      `crypto/rand.Read` is documented to never return an error and always fill the
      buffer. If the OS source somehow fails, it crashes the program rather than hand
      back predictable bytes. Before Go 1.24, ignoring the error was a genuine bug.
  - question: How many bits of randomness does `rand.Text()` promise, at least?
    options:
      - text: '64'
      - text: '128'
        correct: true
      - text: '256'
      - text: It depends on the length you pass in
    explanation: |
      `rand.Text` returns a base32 string with at least 128 bits of randomness, enough
      that guessing it or two of them ever colliding is out of the question. It takes no
      arguments, and a future version may return longer strings if needed.
---

`crypto/rand` is Go's cryptographically secure random number generator. It's small,
and since Go 1.24 it's very hard to misuse.

## Where the randomness comes from

`crypto/rand` doesn't generate randomness itself. It asks the operating system, which
mixes unpredictable events (interrupt timings, hardware RNG instructions, device noise)
into a kernel CSPRNG. On Linux, Go uses the `getrandom(2)` system call; on macOS
`arc4random_buf`; on Windows `ProcessPrng`. In FIPS 140-3 mode the bytes also pass
through an approved DRBG.

You don't need to know any of that to use it, and you should never try to "improve"
it by mixing in your own entropy or seeding something else from it "for speed".

## The four functions

```go
package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func main() {
	// 1. Random bytes: keys, nonces, salts.
	key := make([]byte, 32)
	rand.Read(key)
	fmt.Printf("key:   %x\n", key)

	// 2. A random string for humans or URLs, with at least 128 bits of randomness.
	fmt.Println("text: ", rand.Text())

	// 3. A uniform random integer in [0, max): a six-digit code.
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		panic(err)
	}
	fmt.Printf("code:  %06d\n", n)

	// 4. rand.Reader, an io.Reader for APIs that take one.
	var buf [4]byte
	rand.Reader.Read(buf[:])
	fmt.Printf("reader: %x\n", buf)
}
```

Your output will differ on every run, which is the point. It looks something like:

```
key:   5d83888bca992d2cab9d7a63f8decfc43087b696db7cf1176a2283522fe8a889
text:  7CMZKHFSSHA7AP23ZICEATWJK4
code:  525419
reader: 377486cc
```

- **`rand.Read(b)`** fills `b` completely. It "returns" `(n, err)` for compatibility with
  `io.Reader`, but since Go 1.24 it's documented to **never return an error**. If the
  OS source ever failed, it crashes the program instead, because continuing with
  predictable "random" bytes would be far worse than stopping.
- **`rand.Text()`** (Go 1.24) returns a string of the 26-character RFC 4648 base32
  alphabet (`A`-`Z`, `2`-`7`) carrying at least 128 bits of randomness. It's the easiest
  way to get a password, an API key, or a recovery code that humans might type.
- **`rand.Int(rand.Reader, max)`** returns a uniformly random `*big.Int` in `[0, max)`.
  It's correctly unbiased, which matters (see the modulo bias lesson).
- **`rand.Reader`** is the shared generator as an `io.Reader`, for APIs that want one.

## How much randomness is enough?

Measure in **bits**. Each random bit doubles the attacker's work.

| Use | Bits | Bytes |
| --- | --- | --- |
| Session tokens, reset links, API keys | 128 or more | 16+ |
| Symmetric keys (AES-256, HMAC-SHA256) | 256 | 32 |
| AES-GCM random nonce | 96 | 12 |
| Salts for password hashing | 128 | 16 |

128 bits is past the point of brute force: at a trillion guesses per second, finding
one specific 128-bit token takes about 10^19 years. For keys, 256 bits also leaves a
comfortable margin against future attacks, including Grover's algorithm on a quantum
computer, which in theory halves the number of bits of security a symmetric key gives
you (256 becomes about 128).

## Mistakes to avoid

- **Using `math/rand` for anything secret.** Tokens, keys, nonces, salts, IDs that must
  be unguessable: all `crypto/rand`.
- **Hashing a timestamp or a counter** to "make it random". `sha256(time.Now())` is
  exactly as guessable as `time.Now()`.
- **Shrinking randomness with formatting.** Generating 32 random bytes, then taking
  `%x` and truncating to 8 characters leaves you 32 bits.
- **Rolling your own conversion to a range**, which is next-but-one lesson's topic.
