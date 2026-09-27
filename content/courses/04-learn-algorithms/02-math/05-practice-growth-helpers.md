---
title: 'Practice: Halvings and Factorials'
exercise:
  starter: |
    package main

    import "fmt"

    // halvings returns how many times n can be halved (with integer
    // division) before it reaches 1. That's the floor of log2(n).
    // For n <= 1 it returns 0.
    func halvings(n int) int {
    	// ?
    	return 0
    }

    // factorial returns n! and true. If n is negative, or n! is too
    // big to fit in an int (n > 20), it returns 0 and false.
    func factorial(n int) (int, bool) {
    	// ?
    	return 0, false
    }

    func main() {
    	fmt.Println(halvings(1_000_000)) // want: 19
    	fmt.Println(factorial(5))        // want: 120 true
    	fmt.Println(factorial(21))       // want: 0 false
    }
  solution: |
    package main

    import "fmt"

    func halvings(n int) int {
    	steps := 0
    	for n > 1 {
    		n /= 2
    		steps++
    	}
    	return steps
    }

    func factorial(n int) (int, bool) {
    	if n < 0 || n > 20 {
    		return 0, false
    	}
    	result := 1
    	for i := 2; i <= n; i++ {
    		result *= i
    	}
    	return result, true
    }

    func main() {
    	fmt.Println(halvings(1_000_000))
    	fmt.Println(factorial(5))
    	fmt.Println(factorial(21))
    }
  tests: |
    package main

    import "testing"

    func TestHalvings(t *testing.T) {
    	tests := []struct{ n, want int }{
    		{-5, 0}, {0, 0}, {1, 0}, {2, 1}, {3, 1}, {8, 3}, {1000, 9}, {1024, 10},
    		{1_000_000, 19}, {8_000_000_000, 32},
    	}
    	for _, tt := range tests {
    		if got := halvings(tt.n); got != tt.want {
    			t.Errorf("halvings(%d) = %d, want %d", tt.n, got, tt.want)
    		}
    	}
    }

    func TestFactorial(t *testing.T) {
    	tests := []struct {
    		n      int
    		want   int
    		wantOK bool
    	}{
    		{0, 1, true}, {1, 1, true}, {3, 6, true}, {5, 120, true}, {10, 3628800, true},
    		{20, 2432902008176640000, true}, {21, 0, false}, {100, 0, false}, {-1, 0, false},
    	}
    	for _, tt := range tests {
    		got, ok := factorial(tt.n)
    		if got != tt.want || ok != tt.wantOK {
    			t.Errorf("factorial(%d) = %d, %v; want %d, %v", tt.n, got, ok, tt.want, tt.wantOK)
    		}
    	}
    }
---

Two tiny helpers for Clout's growth dashboard, one for each end of the growth
spectrum you met in this chapter.

## halvings: a logarithm without floats

`math.Log2` works on `float64`s, and floats can round in surprising ways. For
integer work it's often clearer to count **halvings** directly: how many times
can you divide `n` by 2 (with integer division) before reaching 1?

```
1000 → 500 → 250 → 125 → 62 → 31 → 15 → 7 → 3 → 1   (9 halvings)
```

That count is ⌊log₂ n⌋. Complete `halvings` so it returns it, and returns `0`
for any `n <= 1` (including zero and negative numbers).

## factorial: know when you've overflowed

Factorials outgrow `int` fast. 20! = 2,432,902,008,176,640,000 is the largest
that fits in 64 bits, and 21! silently wraps around to a negative number.
Complete `factorial` so that:

- `factorial(n)` returns `n!` and `true` for `0 <= n <= 20` (and `0! = 1`);
- it returns `0, false` for negative `n` or `n > 20`, instead of a garbage
  value.

Returning an `ok` flag instead of a wrong number is the same comma-ok idiom
you used for `findMin`: the caller can't accidentally trust an overflowed
result.
