---
title: Retry Backoff
difficulty: easy
after: loops
hints:
  - 'Keep two variables: the `wait` for the next retry (starting at `1`) and the running `total` (starting at `0`).'
  - 'Loop `retries` times (`for range retries` does nothing when `retries` is 0 or negative). Each time, add `wait` to `total`, then double `wait`.'
  - 'Apply the cap when you double: if the doubled wait is over 60, set it to 60. `min(wait*2, 60)` does it in one step.'
exercise:
  starter: |
    package main

    import "fmt"

    // totalBackoff returns how many seconds Textio waits in total before
    // making `retries` retries. The first wait is 1 second and each wait
    // doubles the one before, but no single wait is longer than 60 seconds.
    func totalBackoff(retries int) int {
    	// Loop once per retry, adding up the waits.
    	return 0
    }

    func main() {
    	fmt.Println(totalBackoff(3)) // want: 7 (1 + 2 + 4)
    	fmt.Println(totalBackoff(8)) // want: 183
    }
  solution: |
    package main

    import "fmt"

    func totalBackoff(retries int) int {
    	total, wait := 0, 1
    	for range retries {
    		total += wait
    		wait = min(wait*2, 60)
    	}
    	return total
    }

    func main() {
    	fmt.Println(totalBackoff(3))
    	fmt.Println(totalBackoff(8))
    }
  tests: |
    package main

    import "testing"

    func TestTotalBackoff(t *testing.T) {
    	tests := []struct {
    		retries int
    		want    int
    	}{
    		{0, 0},
    		{-2, 0},
    		{1, 1},
    		{2, 3},
    		{3, 7},
    		{6, 63},
    		{7, 123},
    		{8, 183},
    		{10, 303},
    		{100, 63 + 94*60},
    	}
    	for _, tt := range tests {
    		if got := totalBackoff(tt.retries); got != tt.want {
    			t.Errorf("totalBackoff(%d) = %d, want %d", tt.retries, got, tt.want)
    		}
    	}
    }
---

When a carrier is down, Textio doesn't hammer it with retries. It waits
before each retry, and doubles the wait every time: 1 second, then 2, 4, 8,
and so on. To stay responsive, no single wait is ever longer than
**60 seconds**.

Complete `totalBackoff(retries)`. It returns the **total** number of seconds
spent waiting before `retries` retries. If `retries` is 0 or negative, nothing
is retried, so it returns `0`.

## Examples

```
totalBackoff(3)  // 7    (1 + 2 + 4)
totalBackoff(7)  // 123  (1 + 2 + 4 + 8 + 16 + 32 + 60)
totalBackoff(8)  // 183  (... + 60 + 60)
totalBackoff(0)  // 0
```

The 7th wait would be 64 seconds, so it's capped at 60, and so is every wait
after it.

## Constraints

- `retries` is any `int` up to 1,000,000.
