---
title: Shuffle the Recovery Words
difficulty: easy
after: randomness
hints:
  - 'Fisher-Yates walks `i` from the last index down to 1 and swaps `words[i]` with a `words[j]` where `j` is uniform in **[0, i]**, including `i` itself. Picking `j` from the whole slice every time looks random but makes some orders more likely than others.'
  - '`rand.Int(r, big.NewInt(int64(i+1)))` returns a uniform `*big.Int` in `[0, i+1)` without modulo bias; `.Int64()` turns it into an index. Reducing a random byte with `% (i+1)` is biased whenever 256 isn''t a multiple of `i+1`.'
exercise:
  starter: |
    package main

    import (
    	"crypto/rand"
    	"fmt"
    	"io"
    )

    // shuffle puts words into a uniformly random order, in place, drawing all
    // randomness from r. Every one of the len(words)! orders must be equally
    // likely. If reading from r fails, return the error.
    func shuffle(r io.Reader, words []string) error {
    	// Fisher-Yates: for i from len(words)-1 down to 1,
    	//   pick j uniformly in [0, i] (inclusive!) with rand.Int(r, big.NewInt(...)),
    	//   then swap words[i] and words[j].
    	return nil
    }

    func main() {
    	words := []string{"otter", "maple", "quartz", "violet", "harbor", "ember"}
    	if err := shuffle(rand.Reader, words); err != nil {
    		fmt.Println("error:", err)
    		return
    	}
    	fmt.Println(words) // a different order on (almost) every run
    }
  solution: |
    package main

    import (
    	"crypto/rand"
    	"fmt"
    	"io"
    	"math/big"
    )

    // shuffle puts words into a uniformly random order, in place, drawing all
    // randomness from r. If reading from r fails, it returns the error.
    func shuffle(r io.Reader, words []string) error {
    	for i := len(words) - 1; i > 0; i-- {
    		j, err := rand.Int(r, big.NewInt(int64(i+1)))
    		if err != nil {
    			return err
    		}
    		k := j.Int64()
    		words[i], words[k] = words[k], words[i]
    	}
    	return nil
    }

    func main() {
    	words := []string{"otter", "maple", "quartz", "violet", "harbor", "ember"}
    	if err := shuffle(rand.Reader, words); err != nil {
    		fmt.Println("error:", err)
    		return
    	}
    	fmt.Println(words)
    }
  tests: |
    package main

    import (
    	"errors"
    	"fmt"
    	"math/rand/v2"
    	"slices"
    	"strings"
    	"testing"
    )

    // seeded returns a deterministic stand-in for crypto/rand.Reader.
    func seeded(n byte) *rand.ChaCha8 {
    	var seed [32]byte
    	seed[0] = n
    	return rand.NewChaCha8(seed)
    }

    func TestShuffleIsPermutation(t *testing.T) {
    	r := seeded(1)
    	for n := range 12 {
    		words := make([]string, n)
    		for i := range words {
    			words[i] = fmt.Sprintf("w%d", i)
    		}
    		orig := slices.Clone(words)
    		if err := shuffle(r, words); err != nil {
    			t.Fatalf("shuffle(%d words) returned error %v", n, err)
    		}
    		got := slices.Clone(words)
    		slices.Sort(got)
    		slices.Sort(orig)
    		if !slices.Equal(got, orig) {
    			t.Fatalf("shuffle(%d words) lost or duplicated words: sorted result %v, want %v", n, got, orig)
    		}
    	}
    }

    func TestShuffleUniformOverThree(t *testing.T) {
    	// 60,000 shuffles of three words: each of the 6 orders should come up
    	// about 10,000 times.
    	r := seeded(2)
    	const trials = 60_000
    	counts := map[string]int{}
    	for range trials {
    		w := []string{"a", "b", "c"}
    		if err := shuffle(r, w); err != nil {
    			t.Fatalf("shuffle returned error %v", err)
    		}
    		counts[strings.Join(w, "")]++
    	}
    	if len(counts) != 6 {
    		t.Fatalf("shuffle of [a b c] produced %d distinct orders in %d tries, want all 6: %v", len(counts), trials, counts)
    	}
    	chi2 := 0.0
    	for _, c := range counts {
    		d := float64(c) - trials/6.0
    		chi2 += d * d / (trials / 6.0)
    	}
    	// With 5 degrees of freedom, a fair shuffle exceeds 30 about once in 60,000 runs.
    	if chi2 > 30 {
    		t.Errorf("shuffle of [a b c] is biased: counts per order %v (chi-squared %.0f, want under 30). Is j drawn from [0, i], not [0, n)?", counts, chi2)
    	}
    }

    func TestShuffleNoModuloBias(t *testing.T) {
    	// With 200 words, the last slot is filled by a word picked from all 200.
    	// Reducing one random byte mod 200 favours words 0-55 two to one.
    	r := seeded(3)
    	const trials = 3000
    	words := make([]string, 200)
    	for i := range words {
    		words[i] = fmt.Sprint(i)
    	}
    	low := 0
    	for range trials {
    		w := slices.Clone(words)
    		if err := shuffle(r, w); err != nil {
    			t.Fatalf("shuffle returned error %v", err)
    		}
    		var n int
    		fmt.Sscan(w[199], &n)
    		if n < 56 {
    			low++
    		}
    	}
    	// Fair: about 28% (840). Byte-mod-200: about 44% (1310).
    	if low < 700 || low > 980 {
    		t.Errorf("in %d shuffles of 200 words, one of words 0-55 ended up last %d times, want about 840 (28%%): the pick is biased. Use rand.Int, not a byte %% n", trials, low)
    	}
    }

    type failingReader struct{}

    func (failingReader) Read([]byte) (int, error) { return 0, errors.New("entropy source unavailable") }

    func TestShuffleReaderError(t *testing.T) {
    	w := []string{"a", "b", "c"}
    	if err := shuffle(failingReader{}, w); err == nil {
    		t.Errorf("shuffle with a failing reader returned nil, want the read error")
    	}
    	if err := shuffle(failingReader{}, []string{"solo"}); err != nil {
    		t.Errorf("shuffle of 1 word needs no randomness, but returned %v", err)
    	}
    }
---

When you create a Keybox account, it shows you a list of **recovery words** and then
asks you to tap them back in a scrambled order, to prove you wrote them down. The
scramble must be fair: if some orders were more likely than others, the check would
be easier to fake, and the same code gets reused for shuffling generated passphrases.

Complete `shuffle(r, words)`. It rearranges `words` **in place** so that every one of
the `len(words)!` orders is equally likely, reading all its randomness from `r`. In
production `r` is `crypto/rand.Reader`; the tests pass a seeded reader so they're
repeatable. If reading from `r` fails, return the error.

## Example

```go
words := []string{"otter", "maple", "quartz"}
err := shuffle(rand.Reader, words)
// words is now one of the 6 orders, e.g. [quartz otter maple], each with chance 1/6
// err == nil
```

## Constraints

- `words` holds 0 to 1,000 words. A slice of 0 or 1 words needs no randomness at all.
- The tests shuffle three words 60,000 times and check all six orders come up about
  equally often, then check a 200-word shuffle for **modulo bias**.
- Use `crypto/rand` helpers, not `math/rand`: don't seed anything yourself.
