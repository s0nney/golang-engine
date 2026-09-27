---
title: Grab What's Ready
difficulty: easy
after: select
hints:
  - 'A `select` with a `default` case never blocks: if no other case can go right now, `default` runs instead.'
  - 'Loop until you have `max` meals. Each time round, `select` between receiving from `ready` and `default` (nothing waiting: return what you have).'
  - 'Use the two-value receive, `meal, ok := <-ready`. When `ok` is false the channel is closed and will never deliver another meal, so stop. Otherwise a closed channel hands you `""` forever.'
exercise:
  starter: |
    package main

    import "fmt"

    // grabReady takes up to max meals that are already waiting on ready and
    // returns them in order. It never waits for a meal that isn't there yet.
    func grabReady(ready <-chan string, max int) []string {
    	var meals []string
    	// Loop while len(meals) < max:
    	//   select {
    	//   case receive from ready: stop if it's closed, else keep the meal
    	//   default: nothing is waiting right now, so return
    	//   }
    	return meals
    }

    func main() {
    	ready := make(chan string, 10)
    	ready <- "burger"
    	ready <- "pad thai"
    	ready <- "tacos"

    	fmt.Println(grabReady(ready, 2)) // want: [burger pad thai]
    	fmt.Println(grabReady(ready, 2)) // want: [tacos]
    	fmt.Println(grabReady(ready, 2)) // want: []
    }
  solution: |
    package main

    import "fmt"

    // grabReady takes up to max meals that are already waiting on ready and
    // returns them in order. It never waits for a meal that isn't there yet.
    func grabReady(ready <-chan string, max int) []string {
    	var meals []string
    	for len(meals) < max {
    		select {
    		case meal, ok := <-ready:
    			if !ok {
    				return meals
    			}
    			meals = append(meals, meal)
    		default:
    			return meals
    		}
    	}
    	return meals
    }

    func main() {
    	ready := make(chan string, 10)
    	ready <- "burger"
    	ready <- "pad thai"
    	ready <- "tacos"

    	fmt.Println(grabReady(ready, 2))
    	fmt.Println(grabReady(ready, 2))
    	fmt.Println(grabReady(ready, 2))
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    	"testing/synctest"
    	"time"
    )

    // grab runs grabReady in a goroutine and fails if it blocks.
    func grab(t *testing.T, ready <-chan string, max int) []string {
    	t.Helper()
    	done := make(chan []string, 1)
    	go func() { done <- grabReady(ready, max) }()
    	select {
    	case got := <-done:
    		return got
    	case <-time.After(time.Hour):
    		t.Fatalf("grabReady(ready, %d) was still waiting after an hour: it must return straight away when no meal is waiting", max)
    		return nil
    	}
    }

    func filled(closed bool, meals ...string) chan string {
    	ch := make(chan string, 10)
    	for _, m := range meals {
    		ch <- m
    	}
    	if closed {
    		close(ch)
    	}
    	return ch
    }

    func TestGrabReady(t *testing.T) {
    	tests := []struct {
    		name   string
    		meals  []string
    		closed bool
    		max    int
    		want   []string
    	}{
    		{"more waiting than max", []string{"burger", "pad thai", "tacos"}, false, 2, []string{"burger", "pad thai"}},
    		{"fewer waiting than max", []string{"burger"}, false, 3, []string{"burger"}},
    		{"exactly max", []string{"a", "b", "c"}, false, 3, []string{"a", "b", "c"}},
    		{"nothing waiting", nil, false, 5, nil},
    		{"max 0", []string{"burger"}, false, 0, nil},
    		{"closed and empty", nil, true, 5, nil},
    		{"closed with leftovers", []string{"soup", "salad"}, true, 5, []string{"soup", "salad"}},
    	}
    	for _, tt := range tests {
    		synctest.Test(t, func(t *testing.T) {
    			ready := filled(tt.closed, tt.meals...)
    			if got := grab(t, ready, tt.max); !slices.Equal(got, tt.want) {
    				t.Errorf("%s: grabReady(%q, %d) = %q, want %q", tt.name, tt.meals, tt.max, got, tt.want)
    			}
    		})
    	}
    }

    func TestGrabReadyLeavesTheRest(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		ready := filled(false, "a", "b", "c", "d", "e")
    		first := grab(t, ready, 2)
    		second := grab(t, ready, 2)
    		third := grab(t, ready, 2)
    		fourth := grab(t, ready, 2)
    		got := [][]string{first, second, third, fourth}
    		want := [][]string{{"a", "b"}, {"c", "d"}, {"e"}, nil}
    		if !slices.EqualFunc(got, want, slices.Equal) {
    			t.Errorf("four grabReady(ready, 2) calls on [a b c d e] returned %q, want %q", got, want)
    		}
    	})
    }

    func TestGrabReadyDoesNotWait(t *testing.T) {
    	synctest.Test(t, func(t *testing.T) {
    		ready := make(chan string)
    		go func() {
    			time.Sleep(time.Second)
    			ready <- "late order"
    		}()
    		if got := grab(t, ready, 1); len(got) != 0 {
    			t.Errorf("grabReady returned %q, want nothing: that meal wasn't ready yet", got)
    		}
    		select { // let the late order arrive
    		case <-ready:
    		case <-time.After(time.Minute):
    		}
    	})
    }
---

Couriers waiting at a Dispatchly kitchen don't hang around. When a courier
arrives, they grab whatever meals are **already** waiting on the pass, up to
what fits in their bag, and leave straight away.

Complete `grabReady(ready, max)`. It receives meals from `ready`, in order,
until it has `max` of them, but it must **never block**:

- If no meal is waiting right now, return what you have (possibly nothing).
- If `ready` is closed, return what you have. Don't count the zero values a
  closed channel keeps handing out as meals.

## Example

```go
ready := make(chan string, 10)
ready <- "burger"
ready <- "pad thai"
ready <- "tacos"

grabReady(ready, 2) // [burger pad thai]
grabReady(ready, 2) // [tacos]
grabReady(ready, 2) // [] (nothing waiting, returns immediately)
```

## Constraints

- `max` ≥ 0. Return `nil` (or an empty slice) when you take nothing.
- The tests run in a `synctest` bubble. If `grabReady` waits for a meal, the
  fake clock jumps ahead an hour and the test reports it at once.
