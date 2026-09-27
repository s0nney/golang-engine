---
title: Carousel Orders
difficulty: easy
after: math
hints:
  - 'The first slot can hold any of the `n` creators, the second any of the `n-1` that are left, and so on for `k` slots.'
  - 'So the answer is `n × (n-1) × … × (n-k+1)`: exactly `k` factors. Multiply them in a loop that runs `k` times.'
  - 'Don''t compute `n!` and divide by `(n-k)!`. `21!` already overflows an `int`, even when the final answer is small.'
exercise:
  starter: |
    package main

    import "fmt"

    // carouselOrders returns how many different ways k of n creators can be
    // placed, in order, into the k slots of the homepage carousel.
    // If k > n there aren't enough creators, so return 0.
    // If k == 0 there is exactly one (empty) carousel, so return 1.
    func carouselOrders(n, k int) int {
    	// Multiply the choices for slot 1, slot 2, ... slot k.
    	return 0
    }

    func main() {
    	fmt.Println(carouselOrders(4, 2))  // want 12
    	fmt.Println(carouselOrders(5, 5))  // want 120 (that's 5!)
    	fmt.Println(carouselOrders(30, 3)) // want 24360
    }
  solution: |
    package main

    import "fmt"

    func carouselOrders(n, k int) int {
    	if k > n {
    		return 0
    	}
    	total := 1
    	for i := range k {
    		total *= n - i
    	}
    	return total
    }

    func main() {
    	fmt.Println(carouselOrders(4, 2))
    	fmt.Println(carouselOrders(5, 5))
    	fmt.Println(carouselOrders(30, 3))
    }
  tests: |
    package main

    import "testing"

    func TestCarouselOrders(t *testing.T) {
    	tests := []struct{ n, k, want int }{
    		{4, 2, 12},
    		{5, 5, 120},
    		{30, 3, 24360},
    		{7, 1, 7},
    		{7, 0, 1},
    		{0, 0, 1},
    		{3, 4, 0},
    		{0, 1, 0},
    		{20, 20, 2432902008176640000},
    		{100, 2, 9900},
    		{50, 10, 37276043023296000},
    		{1_000_000, 3, 999_997_000_002_000_000},
    	}
    	for _, tt := range tests {
    		if got := carouselOrders(tt.n, tt.k); got != tt.want {
    			t.Errorf("carouselOrders(%d, %d) = %d, want %d", tt.n, tt.k, got, tt.want)
    		}
    	}
    }
---

Clout's homepage has a **carousel** of `k` featured creator slots. Marketing has a
shortlist of `n` creators and wants to know how many different carousels are
possible. Order matters: Ava in slot 1 and Bo in slot 2 is a different carousel
from Bo then Ava, and a creator can't appear twice.

Complete `carouselOrders(n, k)`.

## Examples

With 4 creators and 2 slots there are 4 choices for the first slot and 3 for the
second:

```
carouselOrders(4, 2)  // 12
carouselOrders(5, 5)  // 120, which is 5!
carouselOrders(7, 0)  // 1: the empty carousel
carouselOrders(3, 4)  // 0: not enough creators
```

## Constraints

- `0 ≤ n ≤ 1,000,000` and `0 ≤ k`.
- Every test's answer fits in an `int`, but the factorials along the way may not:
  `carouselOrders(1_000_000, 3)` is 999,997,000,002,000,000, while `1,000,000!`
  is astronomically bigger than any `int`.
