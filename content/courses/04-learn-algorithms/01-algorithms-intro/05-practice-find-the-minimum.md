---
title: 'Practice: Smallest Account'
exercise:
  starter: |
    package main

    import "fmt"

    type Influencer struct {
    	Handle    string
    	Followers int
    }

    // smallestAccount returns the influencer with the fewest followers.
    // If several tie, return the first one in the slice.
    // If infs is empty, return the zero Influencer and false.
    func smallestAccount(infs []Influencer) (Influencer, bool) {
    	// ?
    	return Influencer{}, false
    }

    func main() {
    	campaign := []Influencer{
    		{"ava", 8200}, {"bo", 312}, {"cy", 99000}, {"dee", 45}, {"eve", 7100},
    	}
    	inf, ok := smallestAccount(campaign)
    	fmt.Println(inf.Handle, inf.Followers, ok) // want: dee 45 true
    }
  solution: |
    package main

    import "fmt"

    type Influencer struct {
    	Handle    string
    	Followers int
    }

    func smallestAccount(infs []Influencer) (Influencer, bool) {
    	if len(infs) == 0 {
    		return Influencer{}, false
    	}
    	lowest := infs[0]
    	for _, inf := range infs[1:] {
    		if inf.Followers < lowest.Followers {
    			lowest = inf
    		}
    	}
    	return lowest, true
    }

    func main() {
    	campaign := []Influencer{
    		{"ava", 8200}, {"bo", 312}, {"cy", 99000}, {"dee", 45}, {"eve", 7100},
    	}
    	inf, ok := smallestAccount(campaign)
    	fmt.Println(inf.Handle, inf.Followers, ok)
    }
  tests: |
    package main

    import "testing"

    func TestSmallestAccount(t *testing.T) {
    	tests := []struct {
    		name   string
    		in     []Influencer
    		want   Influencer
    		wantOK bool
    	}{
    		{"empty", nil, Influencer{}, false},
    		{"one", []Influencer{{"ava", 10}}, Influencer{"ava", 10}, true},
    		{"middle", []Influencer{{"ava", 500}, {"bo", 20}, {"cy", 90}}, Influencer{"bo", 20}, true},
    		{"first", []Influencer{{"ava", 1}, {"bo", 20}, {"cy", 90}}, Influencer{"ava", 1}, true},
    		{"last", []Influencer{{"ava", 500}, {"bo", 20}, {"cy", 3}}, Influencer{"cy", 3}, true},
    		{"all positive, big", []Influencer{{"ava", 1_000_000}, {"bo", 2_000_000}}, Influencer{"ava", 1_000_000}, true},
    		{"tie keeps first", []Influencer{{"ava", 50}, {"bo", 7}, {"cy", 7}}, Influencer{"bo", 7}, true},
    	}
    	for _, tt := range tests {
    		got, ok := smallestAccount(tt.in)
    		if got != tt.want || ok != tt.wantOK {
    			t.Errorf("%s: smallestAccount(%v) = %v, %v; want %v, %v", tt.name, tt.in, got, ok, tt.want, tt.wantOK)
    		}
    	}
    }
---

Time to write your first algorithm yourself. The campaigns team has a new ticket:

> *Brands want to see the smallest account in each campaign, with its handle,
> so they know who to cut.*

Last time we found the smallest **number**. This time the slice holds whole
`Influencer` structs, and you need to return the **influencer**, not just the
count.

## Your task

Complete `smallestAccount` so that it:

- returns the `Influencer` with the fewest `Followers`, and `true`;
- returns the **first** one if several tie for smallest;
- returns the zero `Influencer{}` and `false` for an empty (or nil) slice,
  without panicking.

Remember the traps from the lessons:

- Don't start from `0` (or from any made-up value). Start from the first
  element, after checking there *is* one.
- Compare with `<`, not `<=`. With `<=`, a later tie would replace the earlier
  influencer.

Structs whose fields are all comparable can be compared with `==`, which is how
the tests check your answer:

```go
Influencer{"dee", 45} == Influencer{"dee", 45} // true
```

Use **Run** to try `main`, then **Submit** to run the tests, which cover empty
input, a single influencer, the minimum in first, middle and last position,
and ties.
