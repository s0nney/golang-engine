---
title: Peak Viewers
difficulty: hard
after: queues
hints:
  - 'Keeping the last `window` readings in a queue and scanning it on every `Peak` is correct, but O(window) per call. With a window of 100,000 that''s 20 billion steps for the big test.'
  - 'Notice that once a reading arrives, any **older, smaller-or-equal** reading can never be the peak again: the new one is at least as big and will stay in the window longer. So throw those away.'
  - 'Keep a queue of `(second, viewers)` pairs whose viewer counts **decrease** from front to back. `Add` pops smaller-or-equal readings off the **back**, then pushes; it also drops the front if it has slid out of the window. `Peak` is simply the front. Every reading is pushed and popped at most once, so it''s O(1) amortized.'
exercise:
  starter: |
    package main

    import "fmt"

    // PeakTracker reports the peak viewer count over the most recent readings.
    type PeakTracker struct {
    	// your fields here
    }

    // NewPeakTracker returns a tracker for the last window readings (window >= 1).
    func NewPeakTracker(window int) *PeakTracker {
    	return &PeakTracker{}
    }

    // Add records the next reading.
    func (p *PeakTracker) Add(viewers int) {
    }

    // Peak returns the largest of the last window readings, or 0 and false
    // if nothing has been added yet.
    func (p *PeakTracker) Peak() (int, bool) {
    	return 0, false
    }

    func main() {
    	p := NewPeakTracker(3)
    	for _, v := range []int{4, 9, 2, 1, 7, 3} {
    		p.Add(v)
    		peak, _ := p.Peak()
    		fmt.Print(peak, " ")
    	}
    	fmt.Println() // want: 4 9 9 9 7 7
    }
  solution: |
    package main

    import "fmt"

    type reading struct {
    	second, viewers int
    }

    // PeakTracker reports the peak viewer count over the most recent readings.
    type PeakTracker struct {
    	window int
    	added  int       // readings added so far
    	queue  []reading // viewers strictly decrease from front to back
    }

    // NewPeakTracker returns a tracker for the last window readings (window >= 1).
    func NewPeakTracker(window int) *PeakTracker {
    	return &PeakTracker{window: window}
    }

    // Add records the next reading.
    func (p *PeakTracker) Add(viewers int) {
    	for len(p.queue) > 0 && p.queue[len(p.queue)-1].viewers <= viewers {
    		p.queue = p.queue[:len(p.queue)-1]
    	}
    	p.queue = append(p.queue, reading{p.added, viewers})
    	p.added++
    	if p.queue[0].second <= p.added-1-p.window {
    		p.queue = p.queue[1:]
    	}
    }

    // Peak returns the largest of the last window readings, or 0 and false
    // if nothing has been added yet.
    func (p *PeakTracker) Peak() (int, bool) {
    	if len(p.queue) == 0 {
    		return 0, false
    	}
    	return p.queue[0].viewers, true
    }

    func main() {
    	p := NewPeakTracker(3)
    	for _, v := range []int{4, 9, 2, 1, 7, 3} {
    		p.Add(v)
    		peak, _ := p.Peak()
    		fmt.Print(peak, " ")
    	}
    	fmt.Println()
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    	"time"
    )

    func peaks(window int, readings []int) []int {
    	p := NewPeakTracker(window)
    	var got []int
    	for _, v := range readings {
    		p.Add(v)
    		peak, ok := p.Peak()
    		if !ok {
    			peak = -999 // Peak must report ok after an Add
    		}
    		got = append(got, peak)
    	}
    	return got
    }

    func TestPeakEmpty(t *testing.T) {
    	p := NewPeakTracker(5)
    	if v, ok := p.Peak(); v != 0 || ok {
    		t.Errorf("Peak() on a new tracker = %d, %v, want 0, false", v, ok)
    	}
    }

    func TestPeaks(t *testing.T) {
    	tests := []struct {
    		name     string
    		window   int
    		readings []int
    		want     []int
    	}{
    		{"example", 3, []int{4, 9, 2, 1, 7, 3}, []int{4, 9, 9, 9, 7, 7}},
    		{"window 1", 1, []int{4, 9, 2, 1}, []int{4, 9, 2, 1}},
    		{"window bigger than stream", 10, []int{3, 1, 5, 2}, []int{3, 3, 5, 5}},
    		{"rising", 2, []int{1, 2, 3, 4, 5}, []int{1, 2, 3, 4, 5}},
    		{"falling", 2, []int{5, 4, 3, 2, 1}, []int{5, 5, 4, 3, 2}},
    		{"duplicates", 2, []int{7, 7, 7, 1, 1}, []int{7, 7, 7, 7, 1}},
    		{"peak expires on time", 3, []int{9, 1, 1, 1, 1}, []int{9, 9, 9, 1, 1}},
    		{"negative", 2, []int{-5, -3, -8, -9}, []int{-5, -3, -3, -8}},
    		{"zeros", 2, []int{0, 0, 0}, []int{0, 0, 0}},
    	}
    	for _, tt := range tests {
    		if got := peaks(tt.window, tt.readings); !slices.Equal(got, tt.want) {
    			t.Errorf("%s: window %d, readings %v: peaks after each Add = %v, want %v (-999 means Peak returned false)", tt.name, tt.window, tt.readings, got, tt.want)
    		}
    	}
    }

    func TestTwoTrackers(t *testing.T) {
    	a, b := NewPeakTracker(2), NewPeakTracker(2)
    	a.Add(100)
    	b.Add(1)
    	if v, _ := b.Peak(); v != 1 {
    		t.Errorf("two trackers share state: second tracker's Peak() = %d, want 1", v)
    	}
    }

    func TestPeakLarge(t *testing.T) {
    	n, window := 200_000, 100_000
    	readings := make([]int, n)
    	for i := range readings {
    		readings[i] = n - i // falling viewers: the window stays full of candidates
    	}
    	done := make(chan []int, 1)
    	go func() { done <- peaks(window, readings) }()
    	select {
    	case got := <-done:
    		for _, want := range []int{0, window - 1, window, n - 1} {
    			exp := readings[max(0, want-window+1)]
    			if got[want] != exp {
    				t.Errorf("after reading %d of %d falling readings (window %d): Peak() = %d, want %d", want, n, window, got[want], exp)
    			}
    		}
    	case <-time.After(time.Second):
    		t.Fatalf("%d Add+Peak calls with window %d took over a second: Peak must not scan the whole window", n, window)
    	}
    }
---

Clout's livestream page shows a **peak viewers** badge: the highest viewer count
from the last few readings. A reading arrives every second, and at 100,000
readings per window, it has to be fast.

Implement `PeakTracker`:

- `NewPeakTracker(window)` creates a tracker that looks at the most recent
  `window` readings (`window ≥ 1`).
- `Add(viewers)` records the next reading.
- `Peak()` returns the largest of the last `window` readings and `true`, or `0`
  and `false` if nothing has been added yet.

## Example

```go
p := NewPeakTracker(3)
// Add:   4  9  2  1  7  3
// Peak:  4  9  9  9  7  7
```

After the 5th reading the window holds `2 1 7`: the 9 has slid out.

## Constraints

- Up to 200,000 readings with a window of up to 100,000. Values are any `int`
  (the tests also use negatives and duplicates).
- `Add` and `Peak` must be **O(1) amortized**. The performance test calls `Add`
  then `Peak` 200,000 times with a window of 100,000 under a one-second limit;
  scanning the window every time would take tens of seconds.
- Separate trackers must not share state.
