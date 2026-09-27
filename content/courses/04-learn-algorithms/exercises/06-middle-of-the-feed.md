---
title: Middle of the Feed
difficulty: easy
after: linked-lists
hints:
  - 'You could count the nodes in one pass and walk to `count/2` in a second. That''s fine! Try the one-pass trick too.'
  - 'Start two pointers at `head`. Move `slow` one node and `fast` two nodes per step. When `fast` can''t take two more steps, `slow` is in the middle.'
exercise:
  starter: |
    package main

    import "fmt"

    type Post struct {
    	ID   int
    	Next *Post
    }

    // middle returns the middle post of the feed that starts at head.
    // With an even number of posts there are two middles; return the second.
    // An empty feed (head == nil) returns nil.
    func middle(head *Post) *Post {
    	// Walk the list with pointers. You can't index a linked list!
    	return nil
    }

    // feed builds a linked list of posts with the given IDs.
    func feed(ids ...int) *Post {
    	var head *Post
    	for i := len(ids) - 1; i >= 0; i-- {
    		head = &Post{ID: ids[i], Next: head}
    	}
    	return head
    }

    func main() {
    	if m := middle(feed(10, 20, 30, 40, 50)); m != nil {
    		fmt.Println(m.ID) // want 30
    	} else {
    		fmt.Println("middle returned nil")
    	}
    }
  solution: |
    package main

    import "fmt"

    type Post struct {
    	ID   int
    	Next *Post
    }

    func middle(head *Post) *Post {
    	slow, fast := head, head
    	for fast != nil && fast.Next != nil {
    		slow = slow.Next
    		fast = fast.Next.Next
    	}
    	return slow
    }

    func feed(ids ...int) *Post {
    	var head *Post
    	for i := len(ids) - 1; i >= 0; i-- {
    		head = &Post{ID: ids[i], Next: head}
    	}
    	return head
    }

    func main() {
    	if m := middle(feed(10, 20, 30, 40, 50)); m != nil {
    		fmt.Println(m.ID)
    	} else {
    		fmt.Println("middle returned nil")
    	}
    }
  tests: |
    package main

    import "testing"

    func TestMiddle(t *testing.T) {
    	tests := []struct {
    		ids  []int
    		want int
    	}{
    		{[]int{10, 20, 30, 40, 50}, 30},
    		{[]int{10, 20, 30, 40}, 30},
    		{[]int{7}, 7},
    		{[]int{7, 8}, 8},
    		{[]int{1, 2, 3}, 2},
    		{[]int{5, 5, 9, 5, 5, 5}, 5},
    	}
    	for _, tt := range tests {
    		head := feed(tt.ids...)
    		got := middle(head)
    		if got == nil {
    			t.Errorf("middle(%v) = nil, want the post with ID %d", tt.ids, tt.want)
    			continue
    		}
    		// It must be a node of the original list, not a copy.
    		pos, want := 0, len(tt.ids)/2
    		n := head
    		for range want {
    			n = n.Next
    		}
    		if got != n {
    			for p := head; p != nil && p != got; p = p.Next {
    				pos++
    			}
    			t.Errorf("middle(%v) returned the post at position %d (ID %d), want position %d (ID %d)", tt.ids, pos, got.ID, want, tt.want)
    		}
    	}
    	if got := middle(nil); got != nil {
    		t.Errorf("middle(nil) = %v, want nil", got)
    	}
    }
---

A Clout feed is a singly linked list of posts. To preload images smoothly, the
app wants to jump to the **middle** of the feed.

Complete `middle(head)`. It returns the middle `*Post` of the list (the node
itself, not a copy). If the feed has an even number of posts, return the
**second** of the two middle posts. An empty feed returns `nil`.

## Examples

```
10 → 20 → 30 → 40 → 50   middle is 30
10 → 20 → 30 → 40        middle is 30 (the second of 20 and 30)
7                        middle is 7
```

## Constraints

- Up to 100,000 posts. O(n) time and O(1) extra memory: don't copy the posts into
  a slice.
- Bonus: find it in a **single** pass over the list.
