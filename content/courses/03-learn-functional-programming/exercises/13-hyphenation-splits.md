---
title: Hyphenation Splits
difficulty: hard
after: closures
hints:
  - '`Memo` returns a closure that owns a `cache := map[K]V{}`. On a call, return the cached value if there is one; otherwise compute `v := f(self, k)`, store it and return it. The trick is the `self` you pass to `f`: it must be the **memoized** function itself, so declare `var self func(K) V` first and then assign the closure to it.'
  - 'For `Splits`, the only thing that changes between recursive calls is **where in the word you are**. Let the key be the start index `i`: the number of ways to split `word[i:]`. If `i == len(word)` there is exactly 1 way (split nothing). Otherwise, for every non-empty dictionary word that `word[i:]` starts with, add the number of ways to split from `i + len(w)`.'
  - 'Create the memoized function **inside** `Splits` (so each call gets a fresh cache for its own word) and return `ways(0)`. Without the cache, a word of 40 `a`s with the dictionary `a, aa, aaa` makes billions of calls; with it, each of the 41 positions is solved once.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    func Memo[K comparable, V any](f func(self func(K) V, k K) V) func(K) V {
    	var self func(K) V
    	self = func(k K) V {
    		return f(self, k) // no cache yet!
    	}
    	return self
    }

    func Splits(word string, dict []string) int {
    	_ = strings.HasPrefix
    	return 0
    }

    func main() {
    	fmt.Println(Splits("notebook", []string{"no", "note", "te", "book", "notebook"})) // want: 3

    	fib := Memo(func(fib func(int) int, n int) int {
    		if n < 2 {
    			return n
    		}
    		return fib(n-1) + fib(n-2)
    	})
    	fmt.Println(fib(30)) // want: 832040 (instantly, even for fib(90))
    }
  solution: |
    package main

    import (
    	"fmt"
    	"strings"
    )

    func Memo[K comparable, V any](f func(self func(K) V, k K) V) func(K) V {
    	cache := map[K]V{}
    	var self func(K) V
    	self = func(k K) V {
    		if v, ok := cache[k]; ok {
    			return v
    		}
    		v := f(self, k)
    		cache[k] = v
    		return v
    	}
    	return self
    }

    func Splits(word string, dict []string) int {
    	ways := Memo(func(self func(int) int, i int) int {
    		if i == len(word) {
    			return 1
    		}
    		total := 0
    		for _, w := range dict {
    			if w != "" && strings.HasPrefix(word[i:], w) {
    				total += self(i + len(w))
    			}
    		}
    		return total
    	})
    	return ways(0)
    }

    func main() {
    	fmt.Println(Splits("notebook", []string{"no", "note", "te", "book", "notebook"}))
    	fib := Memo(func(fib func(int) int, n int) int {
    		if n < 2 {
    			return n
    		}
    		return fib(n-1) + fib(n-2)
    	})
    	fmt.Println(fib(30))
    }
  tests: |
    package main

    import (
    	"strings"
    	"testing"
    	"time"
    )

    func TestMemoFib(t *testing.T) {
    	calls := 0
    	fib := Memo(func(fib func(int) int, n int) int {
    		calls++
    		if n < 2 {
    			return n
    		}
    		return fib(n-1) + fib(n-2)
    	})
    	if got := fib(10); got != 55 {
    		t.Errorf("memoized fib(10) = %d, want 55", got)
    	}
    	if calls != 11 {
    		t.Fatalf("fib(10) called f %d times, want 11 (once for each n from 0 to 10): cache results and pass the memoized function as self", calls)
    	}
    	if got := fib(90); got != 2880067194370816120 {
    		t.Errorf("memoized fib(90) = %d, want 2880067194370816120", got)
    	}
    	if calls != 91 {
    		t.Errorf("after fib(10) and fib(90), f was called %d times, want 91: reuse cached results across calls", calls)
    	}
    	fib(50)
    	if calls != 91 {
    		t.Errorf("fib(50) after fib(90) called f again (%d calls in total, want 91): it's already cached", calls)
    	}
    }

    func TestMemoSeparateCaches(t *testing.T) {
    	calls := 0
    	square := func(self func(int) int, n int) int {
    		calls++
    		return n * n
    	}
    	a, b := Memo(square), Memo(square)
    	a(7)
    	a(7)
    	b(7)
    	if calls != 2 {
    		t.Errorf("two separate Memo(square) functions each called with 7 (a twice) called square %d times, want 2: each Memo needs its own cache", calls)
    	}
    }

    func TestMemoStringKeys(t *testing.T) {
    	calls := 0
    	shout := Memo(func(self func(string) string, s string) string {
    		calls++
    		return strings.ToUpper(s) + "!"
    	})
    	for _, s := range []string{"hi", "yo", "hi", "hi", "yo"} {
    		if got, want := shout(s), strings.ToUpper(s)+"!"; got != want {
    			t.Errorf("shout(%q) = %q, want %q", s, got, want)
    		}
    	}
    	if calls != 2 {
    		t.Errorf("shout called with 2 distinct keys ran f %d times, want 2", calls)
    	}
    }

    func TestSplits(t *testing.T) {
    	tests := []struct {
    		word string
    		dict []string
    		want int
    	}{
    		{"notebook", []string{"no", "note", "te", "book", "notebook"}, 3},
    		{"notebook", []string{"note", "book"}, 1},
    		{"notebook", []string{"note", "books"}, 0},
    		{"", []string{"a"}, 1},
    		{"abc", nil, 0},
    		{"abab", []string{"ab", "a", "b"}, 4},
    		{"aaaa", []string{"a", "aa"}, 5},
    		{"aaa", []string{"a", "a"}, 8},
    		{"xyz", []string{"", "x", "yz"}, 1},
    		{"catsanddog", []string{"cat", "cats", "and", "sand", "dog"}, 2},
    	}
    	for _, tt := range tests {
    		if got := Splits(tt.word, tt.dict); got != tt.want {
    			t.Errorf("Splits(%q, %q) = %d, want %d", tt.word, tt.dict, got, tt.want)
    		}
    	}
    }

    func TestSplitsFast(t *testing.T) {
    	dict := []string{"a", "aa", "aaa"}
    	tests := []struct {
    		word string
    		want int
    	}{
    		{strings.Repeat("a", 60), 4680045560037375},
    		{strings.Repeat("a", 40) + "b", 0},
    	}
    	for _, tt := range tests {
    		done := make(chan int, 1)
    		go func() { done <- Splits(tt.word, dict) }()
    		select {
    		case got := <-done:
    			if got != tt.want {
    				t.Errorf("Splits(%d-letter word, [a aa aaa]) = %d, want %d", len(tt.word), got, tt.want)
    			}
    		case <-time.After(time.Second):
    			t.Fatalf("Splits(%d-letter word, [a aa aaa]) took over a second: memoize the recursion so each position is solved once", len(tt.word))
    		}
    	}
    }
---

Doc2Doc's PDF exporter hyphenates long words at line ends, and to pick good
break points it asks: **in how many ways can this word be split into pieces from
a dictionary?** The natural recursive answer is beautifully short, and
hopelessly slow, because it solves the same sub-problems over and over. The fix
is **memoization**: remember every answer you've already computed.

Write two functions.

**`Memo(f)`** turns a recursive function into a memoized one. `f` receives the
key to compute and a function `self` to use for its recursive calls:

```go
fib := Memo(func(fib func(int) int, n int) int {
	if n < 2 {
		return n
	}
	return fib(n-1) + fib(n-2)
})
fib(90) // 2880067194370816120, instantly
```

- `f` runs **at most once per key** for each memoized function, including for
  the recursive calls made through `self`, and across separate top-level calls.
- Every call to `Memo` gets its own, separate cache.

**`Splits(word, dict)`** returns the number of ways to write `word` as a
sequence of dictionary words (each may be used any number of times). The empty
word has exactly 1 split (use no words). Empty strings in `dict` are ignored,
and a word listed twice in `dict` counts as two different choices.

## Example

```go
Splits("notebook", []string{"no", "note", "te", "book", "notebook"}) // 3
// no+te+book, note+book, notebook

Splits("aaaa", []string{"a", "aa"})       // 5
Splits("notebook", []string{"note"})      // 0
Splits(strings.Repeat("a", 60), []string{"a", "aa", "aaa"}) // 4680045560037375
```

## Constraints

- Words up to 60 letters; answers fit in an `int`.
- The performance test must finish in under a second. Plain recursion on a
  40-letter word of `a`s followed by `b` (answer: 0) takes minutes.
- Solve `Splits` with your `Memo`.
