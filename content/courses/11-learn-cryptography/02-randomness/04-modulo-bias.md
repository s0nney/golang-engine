---
title: Modulo Bias
quiz:
  - question: You pick a random letter with `alphabet[b % 26]`, where `b` is one random byte. Which letters are more likely than the rest?
    options:
      - text: None; every letter is equally likely
      - text: 'The last 4 letters, `w` to `z`'
      - text: 'The first 22 letters appear 10 times in 256; the last 4 only 9 times, so `a`-`v` are about 11% more likely'
        correct: true
      - text: Only `z`
    explanation: |
      256 = 9 × 26 + 22. The 22 leftover byte values (234 to 255) wrap around onto
      `a` to `v`, giving those letters 10 chances each while `w` to `z` get 9. The fix
      is to reject bytes 234 and above and draw again.
exercise:
  starter: |
    package main

    import (
    	"crypto/rand"
    	"errors"
    	"fmt"
    	"io"
    )

    // randomString returns length characters chosen uniformly from alphabet,
    // reading random bytes from r one at a time. It uses rejection sampling:
    // bytes that would bias the result are skipped. It returns an error if
    // alphabet is empty or longer than 256 bytes, or if reading from r fails.
    func randomString(r io.Reader, alphabet string, length int) (string, error) {
    	// ?
    	return "", errors.New("not implemented")
    }

    func main() {
    	pin, err := randomString(rand.Reader, "0123456789", 6)
    	fmt.Println("PIN:", pin, err)
    	word, err := randomString(rand.Reader, "abcdefghijklmnopqrstuvwxyz", 12)
    	fmt.Println("word:", word, err)
    }
  solution: |
    package main

    import (
    	"crypto/rand"
    	"errors"
    	"fmt"
    	"io"
    )

    func randomString(r io.Reader, alphabet string, length int) (string, error) {
    	n := len(alphabet)
    	if n == 0 || n > 256 {
    		return "", errors.New("alphabet must have 1 to 256 characters")
    	}
    	limit := 256 - 256%n // largest multiple of n that fits in a byte's range
    	out := make([]byte, 0, length)
    	var b [1]byte
    	for len(out) < length {
    		if _, err := io.ReadFull(r, b[:]); err != nil {
    			return "", err
    		}
    		if int(b[0]) >= limit {
    			continue // would be biased; draw again
    		}
    		out = append(out, alphabet[int(b[0])%n])
    	}
    	return string(out), nil
    }

    func main() {
    	pin, err := randomString(rand.Reader, "0123456789", 6)
    	fmt.Println("PIN:", pin, err)
    	word, err := randomString(rand.Reader, "abcdefghijklmnopqrstuvwxyz", 12)
    	fmt.Println("word:", word, err)
    }
  tests: |
    package main

    import (
    	"bytes"
    	"crypto/rand"
    	"strings"
    	"testing"
    )

    func TestRejectionSampling(t *testing.T) {
    	for _, tt := range []struct {
    		name     string
    		alphabet string
    		length   int
    		input    []byte
    		want     string
    	}{
    		{"digits: 250-255 are rejected", "0123456789", 4, []byte{255, 250, 7, 249, 0, 123}, "7903"},
    		{"digits: all in range", "0123456789", 3, []byte{0, 9, 10}, "090"},
    		{"26 letters: 234+ rejected", "abcdefghijklmnopqrstuvwxyz", 3, []byte{234, 233, 255, 25, 26}, "zza"},
    		{"32 symbols: nothing rejected", "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567", 2, []byte{255, 32}, "7A"},
    		{"single symbol", "x", 3, []byte{255, 0, 128}, "xxx"},
    		{"zero length reads nothing", "0123456789", 0, nil, ""},
    	} {
    		got, err := randomString(bytes.NewReader(tt.input), tt.alphabet, tt.length)
    		if err != nil {
    			t.Errorf("%s: randomString(bytes %v, %q, %d) error = %v", tt.name, tt.input, tt.alphabet, tt.length, err)
    			continue
    		}
    		if got != tt.want {
    			t.Errorf("%s: randomString(bytes %v, %q, %d) = %q, want %q", tt.name, tt.input, tt.alphabet, tt.length, got, tt.want)
    		}
    	}
    }

    func TestErrors(t *testing.T) {
    	if _, err := randomString(bytes.NewReader([]byte{250, 251, 252}), "0123456789", 1); err == nil {
    		t.Errorf("randomString with only rejected bytes left before EOF: error = nil, want the reader's error")
    	}
    	if _, err := randomString(rand.Reader, "", 5); err == nil {
    		t.Errorf("randomString with an empty alphabet: error = nil, want an error")
    	}
    	if _, err := randomString(rand.Reader, strings.Repeat("a", 257), 5); err == nil {
    		t.Errorf("randomString with a 257-character alphabet: error = nil, want an error")
    	}
    }

    func TestRealRandomness(t *testing.T) {
    	seen := map[rune]bool{}
    	s, err := randomString(rand.Reader, "0123456789", 2000)
    	if err != nil || len(s) != 2000 {
    		t.Fatalf("randomString(rand.Reader, digits, 2000) = %d chars, %v", len(s), err)
    	}
    	for _, r := range s {
    		seen[r] = true
    	}
    	if len(seen) != 10 {
    		t.Errorf("2000 random digits used only %d distinct digits", len(seen))
    	}
    }
---

You have uniformly random *bytes*. You want a uniformly random *digit*, or a letter,
or an index into a word list. The obvious conversion, `b % n`, is subtly wrong.

## Counting the damage

A byte has 256 values. Map them onto 10 digits with `% 10` and count:

```go
package main

import "fmt"

func main() {
	var counts [10]int
	for b := range 256 {
		counts[b%10]++
	}
	fmt.Println(counts)
}
```

```
[26 26 26 26 26 26 25 25 25 25]
```

256 isn't a multiple of 10. The six leftover values, 250 to 255, wrap around onto
digits 0 to 5, so those digits turn up 26 times in 256 while 6 to 9 turn up only 25
times. That's **modulo bias**. Here it's a 4% skew, which is enough for statistical
attacks on things like ECDSA nonces, where tiny biases have leaked whole private keys.
For passwords and PINs it quietly shaves off entropy.

The bias gets worse as `n` approaches the size of the random range. `b % 200` makes
values 0 to 55 twice as likely as 56 to 199.

## Rejection sampling

The fix is simple: **throw away the values that cause the wraparound** and draw again.
Keep only bytes below the largest multiple of `n` that fits:

```go
limit := 256 - 256%n // for n = 10: 250
if b >= limit {
	// reject and draw another byte
}
index := b % n // now perfectly uniform
```

Bytes 0 to 249 split evenly into 25 of each digit. Bytes 250 to 255 are discarded, so
you occasionally need an extra byte. For `n = 10` that's 6 in 256, about 2% of draws.
When `n` divides 256 exactly (2, 4, 8, 16, 32, 64, 128, 256), `limit` is 256 and
nothing is ever rejected. That's one reason base32 and base64 alphabets are popular.

## Use the library when you can

`crypto/rand.Int(rand.Reader, max)` already does this correctly for any size, returning
a `*big.Int`. And `rand.Text()` avoids the problem entirely with its 32-character
alphabet. Writing it yourself makes sense when you need a custom alphabet, such as
digits for a PIN or an unambiguous character set for codes read over the phone.

Floating point is no better, by the way. `int(f * n)` for a random float `f` has its
own small non-uniformity, and it's a common way to smuggle `math/rand` back in.

## Your task

Complete `randomString(r, alphabet, length)`:

1. Return an error if `alphabet` is empty or longer than 256 bytes (treat it as bytes,
   not runes).
2. Compute `limit := 256 - 256%n`, where `n` is the alphabet length.
3. Until you have `length` characters: read **one byte** from `r` (use `io.ReadFull`
   with a one-byte buffer, and return any error), skip it if it's `>= limit`, and
   otherwise append `alphabet[b % n]`.

Taking an `io.Reader` rather than calling `rand.Read` directly makes the function
testable: the tests feed it exact byte sequences, including values that must be
rejected. In production you pass `rand.Reader`.
