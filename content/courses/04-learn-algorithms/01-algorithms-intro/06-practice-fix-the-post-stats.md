---
title: 'Practice: Fix the Post Stats'
exercise:
  starter: |
    package main

    import "fmt"

    // averageLikes returns the mean likes per post, rounded down, and true.
    // For no posts at all it returns 0 and false.
    func averageLikes(likes []int) (int, bool) {
    	sum := 0
    	for _, n := range likes {
    		sum += n
    	}
    	return sum / len(likes), true
    }

    // topPost returns the index of the post with the most likes.
    // If several tie, it returns the earliest one. For no posts it returns -1.
    func topPost(likes []int) int {
    	best := 0
    	bestLikes := 0
    	for i := 1; i < len(likes)-1; i++ {
    		if likes[i] >= bestLikes {
    			best = i
    			bestLikes = likes[i]
    		}
    	}
    	return best
    }

    func main() {
    	week := []int{120, 45, 980, 300, 980, 12}
    	avg, ok := averageLikes(week)
    	fmt.Println("average:", avg, ok)                              // want: average: 406 true
    	fmt.Println("top post:", topPost(week))                       // want: top post: 2
    	fmt.Println("top of [700 50 3]:", topPost([]int{700, 50, 3})) // want: top of [700 50 3]: 0
    }
  solution: |
    package main

    import "fmt"

    func averageLikes(likes []int) (int, bool) {
    	if len(likes) == 0 {
    		return 0, false
    	}
    	sum := 0
    	for _, n := range likes {
    		sum += n
    	}
    	return sum / len(likes), true
    }

    func topPost(likes []int) int {
    	if len(likes) == 0 {
    		return -1
    	}
    	best := 0
    	for i := 1; i < len(likes); i++ {
    		if likes[i] > likes[best] {
    			best = i
    		}
    	}
    	return best
    }

    func main() {
    	week := []int{120, 45, 980, 300, 980, 12}
    	avg, ok := averageLikes(week)
    	fmt.Println("average:", avg, ok)
    	fmt.Println("top post:", topPost(week))
    	fmt.Println("top of [700 50 3]:", topPost([]int{700, 50, 3}))
    }
  tests: |
    package main

    import "testing"

    func TestAverageLikes(t *testing.T) {
    	tests := []struct {
    		name   string
    		in     []int
    		want   int
    		wantOK bool
    	}{
    		{"nil", nil, 0, false},
    		{"empty", []int{}, 0, false},
    		{"one post", []int{77}, 77, true},
    		{"rounds down", []int{1, 2}, 1, true},
    		{"a week", []int{120, 45, 980, 300, 980, 12}, 406, true},
    		{"all zero", []int{0, 0, 0}, 0, true},
    	}
    	for _, tt := range tests {
    		var got int
    		var ok bool
    		func() {
    			defer func() {
    				if r := recover(); r != nil {
    					t.Fatalf("%s: averageLikes(%v) panicked: %v", tt.name, tt.in, r)
    				}
    			}()
    			got, ok = averageLikes(tt.in)
    		}()
    		if got != tt.want || ok != tt.wantOK {
    			t.Errorf("%s: averageLikes(%v) = %d, %v; want %d, %v", tt.name, tt.in, got, ok, tt.want, tt.wantOK)
    		}
    	}
    }

    func TestTopPost(t *testing.T) {
    	tests := []struct {
    		name string
    		in   []int
    		want int
    	}{
    		{"nil", nil, -1},
    		{"empty", []int{}, -1},
    		{"one post", []int{5}, 0},
    		{"best is first", []int{700, 50, 3}, 0},
    		{"two posts", []int{700, 5}, 0},
    		{"best is last", []int{3, 9, 40}, 2},
    		{"best in middle", []int{3, 90, 40}, 1},
    		{"tie keeps earliest", []int{120, 45, 980, 300, 980, 12}, 2},
    		{"all zero", []int{0, 0, 0}, 0},
    	}
    	for _, tt := range tests {
    		if got := topPost(tt.in); got != tt.want {
    			t.Errorf("%s: topPost(%v) = %d, want %d", tt.name, tt.in, got, tt.want)
    		}
    	}
    }
---

An intern wrote two helpers for Clout's weekly creator report. They pass the
intern's one demo, so they shipped. Then the support tickets started:

> *My report page crashes when I didn't post this week.*
>
> *It says my second post was my best one, but my first post got 700 likes
> and the second got 50!*

Time to put the correctness checklist to work. The code compiles and looks
reasonable, but it's wrong for some inputs, and your job is to find out which.

## Your task

Fix `averageLikes` and `topPost` so they match their comments:

- `averageLikes(likes)` returns the mean, rounded down (plain integer
  division), and `true`. With no posts it returns `0, false` instead of
  panicking.
- `topPost(likes)` returns the **index** of the post with the most likes. On a
  tie it returns the **earliest** index. With no posts it returns `-1`.

There are several bugs hiding in these few lines. Before you change anything,
run through the edge-case list from the correctness lesson against each
function:

- The empty slice.
- One element.
- The best value in the **first** position, and in the **last** position.
- Duplicates (ties).

For each one, trace the loop by hand. Does it even look at that element? Does
the comparison do the right thing on a tie?

## Hints

- Integer division by zero panics, so some input needs a guard clause.
- Look closely at where the loop in `topPost` **starts** and where it
  **stops**. Which indexes does it never visit?
- A loop invariant for `topPost` would be: *after looking at the first `k`
  posts, `best` is the index of the earliest post with the most likes among
  them.* Is that true before the loop starts, with `bestLikes` set to `0`?
- `>=` versus `>` decides which of two tied posts wins.

**Run** prints three lines with the expected values in comments. **Submit**
runs a table of tests covering empty input, a single post, the best post first,
last and in the middle, ties and all-zero weeks.
