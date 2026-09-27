---
title: Livestream Guests
difficulty: medium
after: queues
hints:
  - 'Put all guests in a queue. Each turn: dequeue the front guest, give them `min(slot, minutes they still need)`, and either record them as finished or enqueue them again at the back.'
  - 'Don''t change the caller''s `Guest` values. Keep your own queue of `(name, minutes left)` pairs.'
  - 'Removing the front of a slice with `append(q[:0], q[1:]...)` copies the whole queue every turn: O(n) per dequeue. Use `q = q[1:]` (or the ring buffer from the Queues chapter) so each dequeue is O(1).'
exercise:
  starter: |
    package main

    import "fmt"

    type Guest struct {
    	Name    string
    	Minutes int // total airtime this guest needs
    }

    func finishOrder(guests []Guest, slot int) []string {
    	return nil
    }

    func main() {
    	guests := []Guest{{"ava", 5}, {"bo", 2}, {"cy", 7}}
    	fmt.Println(finishOrder(guests, 3)) // want [bo ava cy]
    }
  solution: |
    package main

    import "fmt"

    type Guest struct {
    	Name    string
    	Minutes int // total airtime this guest needs
    }

    func finishOrder(guests []Guest, slot int) []string {
    	queue := make([]Guest, len(guests))
    	copy(queue, guests)
    	order := make([]string, 0, len(guests))
    	for len(queue) > 0 {
    		g := queue[0]
    		queue = queue[1:]
    		g.Minutes -= min(slot, g.Minutes)
    		if g.Minutes == 0 {
    			order = append(order, g.Name)
    		} else {
    			queue = append(queue, g)
    		}
    	}
    	return order
    }

    func main() {
    	guests := []Guest{{"ava", 5}, {"bo", 2}, {"cy", 7}}
    	fmt.Println(finishOrder(guests, 3))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"slices"
    	"testing"
    	"time"
    )

    func TestFinishOrder(t *testing.T) {
    	tests := []struct {
    		guests []Guest
    		slot   int
    		want   []string
    	}{
    		{[]Guest{{"ava", 5}, {"bo", 2}, {"cy", 7}}, 3, []string{"bo", "ava", "cy"}},
    		{[]Guest{{"ava", 5}, {"bo", 2}, {"cy", 7}}, 10, []string{"ava", "bo", "cy"}},
    		{[]Guest{{"ava", 5}, {"bo", 2}, {"cy", 7}}, 1, []string{"bo", "ava", "cy"}},
    		{[]Guest{{"ava", 4}, {"bo", 4}, {"cy", 4}}, 2, []string{"ava", "bo", "cy"}},
    		{[]Guest{{"ava", 6}, {"bo", 1}, {"cy", 3}, {"dee", 2}}, 2, []string{"bo", "dee", "cy", "ava"}},
    		{[]Guest{{"lurker", 0}, {"ava", 1}}, 5, []string{"lurker", "ava"}},
    		{[]Guest{{"solo", 9}}, 2, []string{"solo"}},
    		{nil, 3, []string{}},
    	}
    	for _, tt := range tests {
    		in := slices.Clone(tt.guests)
    		got := finishOrder(in, tt.slot)
    		if !slices.Equal(got, tt.want) {
    			t.Errorf("finishOrder(%v, %d) = %q, want %q", tt.guests, tt.slot, got, tt.want)
    		}
    		if !slices.Equal(in, tt.guests) {
    			t.Errorf("finishOrder changed its input from %v to %v", tt.guests, in)
    		}
    	}
    }

    func TestFinishOrderLarge(t *testing.T) {
    	n := 100_000
    	guests := make([]Guest, n)
    	for i := range guests {
    		guests[i] = Guest{fmt.Sprint("g", i), 1 + i%5}
    	}
    	done := make(chan []string, 1)
    	go func() { done <- finishOrder(guests, 1) }()
    	select {
    	case got := <-done:
    		if len(got) != n || got[0] != "g0" || got[n-1] != "g99999" || got[n/5] != "g1" {
    			t.Errorf("finishOrder(%d guests, 1) gave the wrong order", n)
    		}
    	case <-time.After(time.Second):
    		t.Fatalf("finishOrder(%d guests, 1) took over a second: is dequeue O(n)?", n)
    	}
    }
---

Clout creators host **livestreams** with a line of guests. To keep things fair,
the host gives airtime **round robin**: the guest at the front of the line talks
for up to `slot` minutes. If they still need more time, they go to the back of
the line; otherwise they're done and leave.

Write `finishOrder(guests, slot)`. `guests` is the line in order and each guest
needs `Minutes` of airtime in total. Return the guests' names in the order they
**finish**.

## Example

```
guests: ava needs 5, bo needs 2, cy needs 7; slot = 3

ava talks 3 (2 left) → back of the line
bo  talks 2          → done         finished: bo
cy  talks 3 (4 left) → back of the line
ava talks 2          → done         finished: bo ava
cy  talks 3 (1 left) → back of the line
cy  talks 1          → done         finished: bo ava cy
```

## Details

- A guest who needs 0 minutes finishes as soon as they reach the front of the line.
- Don't modify the caller's `guests` slice.
- No guests gives an empty result.

## Constraints

- Up to 100,000 guests, `slot ≥ 1`, and at most 1,000,000 turns in total.
- Each dequeue must be O(1). The large test has 300,000 turns and runs under a
  one-second limit.
