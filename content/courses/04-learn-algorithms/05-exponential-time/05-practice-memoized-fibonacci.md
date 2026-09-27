---
title: 'Practice: Memoized Fibonacci'
exercise:
  starter: |
    package main

    import "fmt"

    // fib returns the nth Fibonacci number: fib(0) = 0, fib(1) = 1,
    // fib(n) = fib(n-1) + fib(n-2). It's correct, but exponentially slow.
    // Make it fast with memoization: fib(90) must return instantly.
    func fib(n int) int {
    	if n < 2 {
    		return n
    	}
    	return fib(n-1) + fib(n-2)
    }

    func main() {
    	for _, n := range []int{10, 20, 30} {
    		fmt.Printf("fib(%d) = %d\n", n, fib(n))
    	}
    	// Uncomment once fib is memoized. As written, this would run for years:
    	// fmt.Println(fib(90)) // want: 2880067194370816120
    }
  solution: |
    package main

    import "fmt"

    var memo = map[int]int{}

    func fib(n int) int {
    	if n < 2 {
    		return n
    	}
    	if v, ok := memo[n]; ok {
    		return v
    	}
    	v := fib(n-1) + fib(n-2)
    	memo[n] = v
    	return v
    }

    func main() {
    	for _, n := range []int{10, 20, 30} {
    		fmt.Printf("fib(%d) = %d\n", n, fib(n))
    	}
    	fmt.Println(fib(90))
    }
  tests: |
    package main

    import (
    	"testing"
    	"time"
    )

    func TestFibSmall(t *testing.T) {
    	want := []int{0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55}
    	for n, w := range want {
    		if got := fib(n); got != w {
    			t.Errorf("fib(%d) = %d, want %d", n, got, w)
    		}
    	}
    }

    func TestFibFast(t *testing.T) {
    	cases := []struct{ n, want int }{
    		{40, 102334155},
    		{70, 190392490709135},
    		{90, 2880067194370816120},
    	}
    	for _, c := range cases {
    		done := make(chan int, 1)
    		go func() { done <- fib(c.n) }()
    		select {
    		case got := <-done:
    			if got != c.want {
    				t.Errorf("fib(%d) = %d, want %d", c.n, got, c.want)
    			}
    		case <-time.After(time.Second):
    			t.Fatalf("fib(%d) took over a second: is it memoized?", c.n)
    		}
    	}
    }
---

Clout's growth model uses Fibonacci numbers, and the data team wants to
forecast 90 months ahead. The starter's `fib` is the naive recursive version
from the lesson. It's correct, and `fib(30)` is fine, but `fib(90)` would make
around 10¹⁹ calls. That's centuries of CPU time.

## Your task

Make `fib` fast **with memoization**, without changing its signature:

- Keep a cache of results you've already computed, for example a package-level
  `map[int]int`.
- Before computing `fib(n)`, check the cache. After computing it, store it.

The tests check small values, then `fib(40)`, `fib(70)` and `fib(90)`, each of
which must come back within a second. With a memo, each value from 2 to 90 is
computed exactly once: O(n).

Remember the nil-map gotcha. This compiles, but panics on the first write:

```go
var memo map[int]int // nil!
```

Initialise it with `map[int]int{}` or `make(map[int]int)`.

(Yes, a simple bottom-up loop with two variables is even better, and it would
pass the tests too. But practise the memo: it's the technique that
generalises to problems where a loop isn't obvious.)
