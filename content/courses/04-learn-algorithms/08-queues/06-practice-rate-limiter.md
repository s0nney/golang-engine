---
title: 'Practice: Post Rate Limiter'
exercise:
  starter: |
    package main

    import "fmt"

    // Limiter lets each account post at most Limit times in any Window seconds.
    type Limiter struct {
    	Limit  int
    	Window int
    	times  []int // when each accepted post happened, oldest first: a queue
    }

    // expire dequeues every time that has fallen out of the window ending at
    // now. A post at time t counts until time t+Window, when it expires.
    func (l *Limiter) expire(now int) {
    	// ?
    }

    // Allow reports whether a post at time now (in seconds; calls never go
    // back in time) is accepted. Accepted posts join the queue. Rejected
    // posts don't count against the limit.
    func (l *Limiter) Allow(now int) bool {
    	// ?
    	return true
    }

    // Recent returns how many accepted posts are still inside the window
    // ending at now.
    func (l *Limiter) Recent(now int) int {
    	// ?
    	return 0
    }

    func main() {
    	l := Limiter{Limit: 3, Window: 10}
    	for _, now := range []int{0, 1, 2, 3, 9, 10, 11, 12, 25} {
    		fmt.Print(now, ":", l.Allow(now), " ")
    	}
    	fmt.Println()
    	// want: 0:true 1:true 2:true 3:false 9:false 10:true 11:true 12:true 25:true
    	fmt.Println("recent at 26:", l.Recent(26)) // want: recent at 26: 1
    }
  solution: |
    package main

    import "fmt"

    // Limiter lets each account post at most Limit times in any Window seconds.
    type Limiter struct {
    	Limit  int
    	Window int
    	times  []int // when each accepted post happened, oldest first: a queue
    }

    func (l *Limiter) expire(now int) {
    	for len(l.times) > 0 && l.times[0] <= now-l.Window {
    		l.times = l.times[1:]
    	}
    }

    func (l *Limiter) Allow(now int) bool {
    	l.expire(now)
    	if len(l.times) >= l.Limit {
    		return false
    	}
    	l.times = append(l.times, now)
    	return true
    }

    func (l *Limiter) Recent(now int) int {
    	l.expire(now)
    	return len(l.times)
    }

    func main() {
    	l := Limiter{Limit: 3, Window: 10}
    	for _, now := range []int{0, 1, 2, 3, 9, 10, 11, 12, 25} {
    		fmt.Print(now, ":", l.Allow(now), " ")
    	}
    	fmt.Println()
    	// want: 0:true 1:true 2:true 3:false 9:false 10:true 11:true 12:true 25:true
    	fmt.Println("recent at 26:", l.Recent(26)) // want: recent at 26: 1
    }
  tests: |
    package main

    import (
    	"fmt"
    	"testing"
    )

    func run(l *Limiter, times []int) string {
    	s := ""
    	for _, now := range times {
    		if l.Allow(now) {
    			s += fmt.Sprint(now, "✓ ")
    		} else {
    			s += fmt.Sprint(now, "✗ ")
    		}
    	}
    	return s
    }

    func TestAllow(t *testing.T) {
    	tests := []struct {
    		limit, window int
    		times         []int
    		want          string
    	}{
    		{3, 10, []int{0, 1, 2, 3, 9, 10, 11, 12, 25}, "0✓ 1✓ 2✓ 3✗ 9✗ 10✓ 11✓ 12✓ 25✓ "},
    		{1, 5, []int{0, 4, 5, 5, 9, 10}, "0✓ 4✗ 5✓ 5✗ 9✗ 10✓ "},
    		{2, 60, []int{0, 0, 0, 30, 59, 60, 61}, "0✓ 0✓ 0✗ 30✗ 59✗ 60✓ 61✓ "},
    	}
    	for _, tt := range tests {
    		l := &Limiter{Limit: tt.limit, Window: tt.window}
    		if got := run(l, tt.times); got != tt.want {
    			t.Errorf("Limit %d, Window %d, posts at %v:\n got  %s\n want %s", tt.limit, tt.window, tt.times, got, tt.want)
    		}
    	}
    }

    func TestRejectedPostsDontCount(t *testing.T) {
    	l := &Limiter{Limit: 2, Window: 10}
    	// 0 and 1 are accepted; 2..9 are rejected and must not join the queue.
    	for now := range 10 {
    		l.Allow(now)
    	}
    	if !l.Allow(10) {
    		t.Error("Allow(10) = false, want true: the post at 0 has expired, and rejected posts shouldn't count")
    	}
    	if got := l.Recent(10); got != 2 {
    		t.Errorf("Recent(10) = %d, want 2 (the posts at 1 and 10)", got)
    	}
    }

    func TestRecent(t *testing.T) {
    	l := &Limiter{Limit: 5, Window: 10}
    	for _, now := range []int{0, 2, 4, 6} {
    		l.Allow(now)
    	}
    	for _, tt := range []struct{ now, want int }{{6, 4}, {10, 3}, {13, 2}, {16, 0}, {100, 0}} {
    		if got := l.Recent(tt.now); got != tt.want {
    			t.Errorf("posts at 0,2,4,6 (Window 10): Recent(%d) = %d, want %d", tt.now, got, tt.want)
    		}
    	}
    }

    func TestManyPosts(t *testing.T) {
    	l := &Limiter{Limit: 100, Window: 50}
    	accepted := 0
    	for now := range 300_000 {
    		if l.Allow(now) {
    			accepted++
    		}
    	}
    	if accepted != 300_000 {
    		t.Errorf("one post per second with Limit 100, Window 50: accepted %d of 300000, want all", accepted)
    	}
    	if got := l.Recent(299_999); got != 50 {
    		t.Errorf("Recent(299999) = %d, want 50", got)
    	}
    }
---

Spam bots love Clout. The trust and safety team wants a simple rule: an
account may post at most **`Limit` times in any `Window` seconds**. Post a
fourth time inside ten seconds with a limit of three, and the post bounces.

This is a **sliding window** rate limiter, and a queue is exactly the right
structure for it.

## Why a queue?

Keep the timestamps of the account's accepted posts, oldest first. When a new
post arrives at time `now`:

1. **Expire** old posts. Any timestamp `t` with `t <= now - Window` has slid
   out of the window. Because the queue is in time order, all the expired ones
   are at the **front**, so keep dequeuing until the front is recent enough.
2. **Check** the length. If the queue already holds `Limit` posts, reject.
3. Otherwise **enqueue** `now` and accept.

New timestamps go on the back, old ones leave from the front: first in, first
out. Every timestamp is enqueued once and dequeued at most once, so the work
is **amortized O(1) per post**, however many posts an account makes.

Here's a trace with `Limit: 3, Window: 10`:

| now | queue before    | expired | result   | queue after     |
|----:|-----------------|---------|----------|-----------------|
| 0   | `[]`            |         | accept   | `[0]`           |
| 1   | `[0]`           |         | accept   | `[0 1]`         |
| 2   | `[0 1]`         |         | accept   | `[0 1 2]`       |
| 3   | `[0 1 2]`       |         | reject   | `[0 1 2]`       |
| 9   | `[0 1 2]`       |         | reject   | `[0 1 2]`       |
| 10  | `[0 1 2]`       | 0       | accept   | `[1 2 10]`      |
| 11  | `[1 2 10]`      | 1       | accept   | `[2 10 11]`     |
| 12  | `[2 10 11]`     | 2       | accept   | `[10 11 12]`    |
| 25  | `[10 11 12]`    | all     | accept   | `[25]`          |

A post at time 0 counts until time 10, when it expires. Rejected posts are
**not** enqueued: a bounced post shouldn't count against you.

## Your task

Complete three methods on `Limiter`:

- **`expire(now)`** dequeues every timestamp that has left the window. Use the
  re-slicing trick from the naive queue lesson (`l.times = l.times[1:]`), which
  is O(1) per dequeue. For short-lived per-account queues like this, it's
  fine.
- **`Allow(now)`** runs the three steps above and reports whether the post was
  accepted.
- **`Recent(now)`** expires old posts and returns how many accepted posts are
  still in the window.

You can assume `now` never goes backwards between calls.

**Run** prints the table above as `now:result` pairs. **Submit** also checks
boundary times, rejected posts, and 300,000 posts in a row.
