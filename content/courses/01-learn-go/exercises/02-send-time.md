---
title: Send Time
difficulty: easy
after: functions
hints:
  - 'First find the total number of seconds. `queued / perSecond` rounds **down**, but a half-full last second still takes a whole second.'
  - 'A classic trick to round a division up: `(queued + perSecond - 1) / perSecond`. Try it with 10 and 3: `12 / 3 = 4`.'
  - 'Once you have the total seconds, `total / 60` gives the minutes and `total % 60` gives the seconds left over. Return both.'
exercise:
  starter: |
    package main

    import "fmt"

    // sendTime reports how long it takes to send queued messages when Textio
    // sends perSecond messages every second. A partly used last second still
    // counts as a whole second. The result is split into minutes and seconds.
    func sendTime(queued, perSecond int) (int, int) {
    	// 1. Work out the total seconds, rounding up.
    	// 2. Split it into minutes and leftover seconds.
    	return 0, 0
    }

    func main() {
    	minutes, seconds := sendTime(1000, 7)
    	fmt.Println(minutes, "min", seconds, "s") // want: 2 min 23 s
    }
  solution: |
    package main

    import "fmt"

    func sendTime(queued, perSecond int) (int, int) {
    	total := (queued + perSecond - 1) / perSecond
    	return total / 60, total % 60
    }

    func main() {
    	minutes, seconds := sendTime(1000, 7)
    	fmt.Println(minutes, "min", seconds, "s")
    }
  tests: |
    package main

    import "testing"

    func TestSendTime(t *testing.T) {
    	tests := []struct {
    		queued, perSecond int
    		wantMin, wantSec  int
    	}{
    		{1000, 7, 2, 23},
    		{0, 5, 0, 0},
    		{1, 100, 0, 1},
    		{60, 1, 1, 0},
    		{59, 1, 0, 59},
    		{61, 1, 1, 1},
    		{10, 3, 0, 4},
    		{9, 3, 0, 3},
    		{3600, 1, 60, 0},
    		{120, 2, 1, 0},
    		{121, 2, 1, 1},
    	}
    	for _, tt := range tests {
    		gotMin, gotSec := sendTime(tt.queued, tt.perSecond)
    		if gotMin != tt.wantMin || gotSec != tt.wantSec {
    			t.Errorf("sendTime(%d, %d) = %d, %d, want %d, %d", tt.queued, tt.perSecond, gotMin, gotSec, tt.wantMin, tt.wantSec)
    		}
    	}
    }
---

Before a big campaign goes out, Textio tells the customer how long it will
take. The sending queue pushes out `perSecond` messages every second.

Complete `sendTime(queued, perSecond)`. It returns two values: the number of
whole **minutes** and the leftover **seconds** needed to send `queued`
messages. If the last second is only partly used, it still counts as a full
second.

## Examples

```
sendTime(1000, 7)  // 2, 23   (1000 / 7 = 142.9, so 143 seconds)
sendTime(60, 1)    // 1, 0
sendTime(10, 3)    // 0, 4    (3 + 3 + 3 + 1)
sendTime(0, 5)     // 0, 0
```

## Constraints

- `queued` is 0 or more, and `perSecond` is at least 1.
- You don't need `if` for this one: integer division and `%` are enough.
