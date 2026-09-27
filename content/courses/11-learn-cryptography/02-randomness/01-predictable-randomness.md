---
title: Predictable Randomness
quiz:
  - question: A server seeds `math/rand/v2` with `time.Now().UnixNano()` instead of `Unix()`. Does that make its reset tokens safe?
    options:
      - text: Yes, nanoseconds give about a billion times more possible seeds
      - text: No; the attacker can often narrow the time to milliseconds or less, and even without that, the generator isn't designed to hide its state from someone who sees its output
        correct: true
      - text: Yes, as long as the server restarts often
      - text: No, because `UnixNano` returns the same value on every call
    explanation: |
      A timestamp is a guessable seed however fine-grained it is, since the attacker
      knows roughly when the server started. Worse, non-cryptographic generators like
      PCG make no promise that their output hides their internal state. Secrets come
      from `crypto/rand`, full stop.
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"math/rand/v2"
    )

    // tokenGen is Keybox's (broken!) first attempt at password-reset tokens.
    type tokenGen struct {
    	r *rand.Rand
    }

    // newTokenGen seeds the generator with the server's start time in Unix seconds.
    func newTokenGen(startUnix int64) *tokenGen {
    	return &tokenGen{r: rand.New(rand.NewPCG(uint64(startUnix), 0x6b6579626f78))}
    }

    const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

    // next returns a 16-character token.
    func (g *tokenGen) next() string {
    	b := make([]byte, 16)
    	for i := range b {
    		b[i] = alphabet[g.r.IntN(len(alphabet))]
    	}
    	return string(b)
    }

    // predictNext is Mallory's attack. She requested a reset for her own account
    // and got token observed, the first token issued since the server started
    // at some unknown second in [from, to]. Return the token the server will
    // issue next (Alice's!) and true, or "", false if no start time matches.
    func predictNext(observed string, from, to int64) (string, bool) {
    	// ?
    	return "", false
    }

    func main() {
    	start := int64(1790000000) + 1234 // unknown to Mallory
    	server := newTokenGen(start)
    	mallorys := server.next()
    	alices := server.next()

    	guess, ok := predictNext(mallorys, 1790000000, 1790003600)
    	fmt.Println("Mallory's token:", mallorys)
    	fmt.Println("Alice's token:  ", alices)
    	fmt.Println("prediction:     ", guess, ok)
    }
  solution: |
    package main

    import (
    	"fmt"
    	"math/rand/v2"
    )

    type tokenGen struct {
    	r *rand.Rand
    }

    func newTokenGen(startUnix int64) *tokenGen {
    	return &tokenGen{r: rand.New(rand.NewPCG(uint64(startUnix), 0x6b6579626f78))}
    }

    const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

    func (g *tokenGen) next() string {
    	b := make([]byte, 16)
    	for i := range b {
    		b[i] = alphabet[g.r.IntN(len(alphabet))]
    	}
    	return string(b)
    }

    func predictNext(observed string, from, to int64) (string, bool) {
    	for seed := from; seed <= to; seed++ {
    		g := newTokenGen(seed)
    		if g.next() == observed {
    			return g.next(), true
    		}
    	}
    	return "", false
    }

    func main() {
    	start := int64(1790000000) + 1234
    	server := newTokenGen(start)
    	mallorys := server.next()
    	alices := server.next()

    	guess, ok := predictNext(mallorys, 1790000000, 1790003600)
    	fmt.Println("Mallory's token:", mallorys)
    	fmt.Println("Alice's token:  ", alices)
    	fmt.Println("prediction:     ", guess, ok)
    }
  tests: |
    package main

    import "testing"

    func TestPredictNext(t *testing.T) {
    	for _, start := range []int64{1790000000, 1790000001, 1790001800, 1790003599, 1790003600} {
    		g := newTokenGen(start)
    		observed := g.next()
    		want := g.next()
    		got, ok := predictNext(observed, 1790000000, 1790003600)
    		if !ok || got != want {
    			t.Errorf("server started at %d: predictNext(%q, window) = %q, %v; want %q, true",
    				start, observed, got, ok, want)
    		}
    	}
    }

    func TestPredictNextOutsideWindow(t *testing.T) {
    	observed := newTokenGen(1700000000).next()
    	if got, ok := predictNext(observed, 1790000000, 1790003600); ok {
    		t.Errorf("predictNext for a server started outside the window = %q, true; want \"\", false", got)
    	}
    }
---

Cryptography runs on randomness. Keys, nonces, salts, session tokens and reset links
are only as unguessable as the random numbers behind them. And "random" means
something much stronger here than it does in a game.

## Two kinds of random

`math/rand/v2` gives you **pseudo-random** numbers: a deterministic algorithm (PCG or
ChaCha8) that turns a **seed** into a long, statistically even stream. Same seed, same
stream, every time. That's a feature for simulations, shuffling a playlist, or jitter
in retry delays.

```go
package main

import (
	"fmt"
	"math/rand/v2"
)

func main() {
	for range 2 {
		r := rand.New(rand.NewPCG(1790000000, 0))
		fmt.Println(r.IntN(1000), r.IntN(1000), r.IntN(1000))
	}
}
```

```
597 311 126
597 311 126
```

For security, "statistically even" isn't enough. You need **unpredictability**: even an
attacker who has seen lots of output, and knows the algorithm, can't guess the next
value. That's a **cryptographically secure** random number generator (CSPRNG), and in
Go it's `crypto/rand`, covered next lesson.

(The top-level functions like `rand.IntN` are randomly seeded at startup, and the
package's ChaCha8 source is designed to resist prediction. Even so, the package's own
documentation says it "should not be used for security-sensitive work", and `rand.New`
with a seed *you* chose is predictable by construction. The rule is simple: secrets
come from `crypto/rand`.)

## Seeds are keys

When you seed a generator, the seed *is* the secret. Everything the generator produces
can be recomputed by anyone who knows it. The classic mistake is seeding with the clock:

```go
r := rand.New(rand.NewPCG(uint64(time.Now().Unix()), 0))
```

An attacker who knows the server started "sometime this morning" has only a few
thousand seconds to try. For each candidate seed, generate a token and compare it with
one they legitimately received. When it matches, they own the generator: every token
it will ever issue is now predictable.

This isn't hypothetical. Early Netscape SSL seeded its generator from the time and
process IDs, and researchers recovered session keys in seconds. Debian shipped an
OpenSSL build in 2006 to 2008 whose only entropy was the process ID, which left just
at most 32,767 possible keys of each type and size for every SSH and TLS key generated
on those machines.

## Your task

Keybox's first reset-token generator seeds PCG with the server's start time. Break it.

Complete `predictNext(observed, from, to)`. For every candidate start time `seed` from
`from` to `to` **inclusive**, build `newTokenGen(seed)` and generate its first token.
If it equals `observed`, the next token from that same generator is Alice's; return it
and `true`. If no seed matches, return `"", false`.

An hour-wide window is 3,601 guesses. Your laptop will finish before you let go of the
Run button.
