---
title: Merge Two Feeds
difficulty: medium
after: linked-lists
hints:
  - 'This is the `merge` step of merge sort, on linked lists: repeatedly take whichever front post is newer, and move that list forward by one.'
  - 'A **dummy** head node (`var dummy Post; tail := &dummy`) saves you from special-casing the first post. Attach nodes with `tail.Next = p; tail = p`, and return `dummy.Next`.'
  - 'When one list runs out, attach the rest of the other in one step: it''s already in order. For ties, take from `a` first (use `>=`).'
exercise:
  starter: |
    package main

    import "fmt"

    type Post struct {
    	Author string
    	Time   int // seconds since launch; bigger is newer
    	Next   *Post
    }

    func mergeFeeds(a, b *Post) *Post {
    	return nil
    }

    func feed(posts ...Post) *Post {
    	var head *Post
    	for i := len(posts) - 1; i >= 0; i-- {
    		p := posts[i]
    		p.Next = head
    		head = &p
    	}
    	return head
    }

    func show(p *Post) string {
    	s := "["
    	for ; p != nil; p = p.Next {
    		s += fmt.Sprintf(" %s@%d", p.Author, p.Time)
    	}
    	return s + " ]"
    }

    func main() {
    	following := feed(Post{Author: "ava", Time: 90}, Post{Author: "bo", Time: 40}, Post{Author: "ava", Time: 10})
    	sponsored := feed(Post{Author: "ad", Time: 70}, Post{Author: "ad", Time: 40})
    	fmt.Println(show(mergeFeeds(following, sponsored)))
    	// want [ ava@90 ad@70 bo@40 ad@40 ava@10 ]
    }
  solution: |
    package main

    import "fmt"

    type Post struct {
    	Author string
    	Time   int // seconds since launch; bigger is newer
    	Next   *Post
    }

    func mergeFeeds(a, b *Post) *Post {
    	var dummy Post
    	tail := &dummy
    	for a != nil && b != nil {
    		if a.Time >= b.Time {
    			tail.Next, a = a, a.Next
    		} else {
    			tail.Next, b = b, b.Next
    		}
    		tail = tail.Next
    	}
    	if a != nil {
    		tail.Next = a
    	} else {
    		tail.Next = b
    	}
    	return dummy.Next
    }

    func feed(posts ...Post) *Post {
    	var head *Post
    	for i := len(posts) - 1; i >= 0; i-- {
    		p := posts[i]
    		p.Next = head
    		head = &p
    	}
    	return head
    }

    func show(p *Post) string {
    	s := "["
    	for ; p != nil; p = p.Next {
    		s += fmt.Sprintf(" %s@%d", p.Author, p.Time)
    	}
    	return s + " ]"
    }

    func main() {
    	following := feed(Post{Author: "ava", Time: 90}, Post{Author: "bo", Time: 40}, Post{Author: "ava", Time: 10})
    	sponsored := feed(Post{Author: "ad", Time: 70}, Post{Author: "ad", Time: 40})
    	fmt.Println(show(mergeFeeds(following, sponsored)))
    }
  tests: |
    package main

    import (
    	"testing"
    	"time"
    )

    func nodes(p *Post) map[*Post]bool {
    	m := map[*Post]bool{}
    	for ; p != nil && len(m) < 1_000_000; p = p.Next {
    		m[p] = true
    	}
    	return m
    }

    func TestMergeFeeds(t *testing.T) {
    	p := func(author string, time int) Post { return Post{Author: author, Time: time} }
    	tests := []struct {
    		name string
    		a, b []Post
    		want string
    	}{
    		{"interleaved with a tie", []Post{p("ava", 90), p("bo", 40), p("ava", 10)}, []Post{p("ad", 70), p("ad", 40)}, "[ ava@90 ad@70 bo@40 ad@40 ava@10 ]"},
    		{"both empty", nil, nil, "[ ]"},
    		{"a empty", nil, []Post{p("ad", 5), p("ad", 1)}, "[ ad@5 ad@1 ]"},
    		{"b empty", []Post{p("ava", 5)}, nil, "[ ava@5 ]"},
    		{"all of b first", []Post{p("ava", 3), p("ava", 2)}, []Post{p("ad", 9), p("ad", 8)}, "[ ad@9 ad@8 ava@3 ava@2 ]"},
    		{"all ties", []Post{p("a1", 7), p("a2", 7)}, []Post{p("b1", 7), p("b2", 7)}, "[ a1@7 a2@7 b1@7 b2@7 ]"},
    		{"negative times", []Post{p("ava", -1), p("ava", -20)}, []Post{p("ad", 0), p("ad", -5)}, "[ ad@0 ava@-1 ad@-5 ava@-20 ]"},
    	}
    	for _, tt := range tests {
    		a, b := feed(tt.a...), feed(tt.b...)
    		want := nodes(a)
    		for n := range nodes(b) {
    			want[n] = true
    		}
    		got := mergeFeeds(a, b)
    		if s := show(got); s != tt.want {
    			t.Errorf("%s: mergeFeeds(%s, %s) = %s, want %s", tt.name, show(feed(tt.a...)), show(feed(tt.b...)), s, tt.want)
    			continue
    		}
    		for n := range nodes(got) {
    			if !want[n] {
    				t.Errorf("%s: the merged feed contains a new node %s@%d; relink the existing posts instead of copying them", tt.name, n.Author, n.Time)
    				break
    			}
    		}
    	}
    }

    func TestMergeFeedsLarge(t *testing.T) {
    	n := 100_000
    	var as, bs []Post
    	for i := range n {
    		as = append(as, Post{Author: "a", Time: 2 * (n - i)})
    		bs = append(bs, Post{Author: "b", Time: 2*(n-i) - 1})
    	}
    	start := time.Now()
    	got := mergeFeeds(feed(as...), feed(bs...))
    	d := time.Since(start)
    	count, prev := 0, 1<<62
    	for p := got; p != nil; p = p.Next {
    		if p.Time > prev {
    			t.Fatalf("mergeFeeds of two %d-post feeds: post at %d comes after %d, want newest first", n, p.Time, prev)
    		}
    		prev = p.Time
    		count++
    	}
    	if count != 2*n {
    		t.Errorf("mergeFeeds of two %d-post feeds has %d posts, want %d", n, count, 2*n)
    	}
    	if d > time.Second {
    		t.Errorf("mergeFeeds of two %d-post feeds took %v, want O(n)", n, d)
    	}
    }
---

Your Clout home screen mixes two feeds: posts from people you **follow**, and
**sponsored** posts. Each feed is a singly linked list, already sorted **newest
first** (biggest `Time` first). The app needs one combined feed, still newest
first.

Write `mergeFeeds(a, b)`, which merges the two lists and returns the head of the
merged list.

- **Relink** the existing `*Post` nodes by changing their `Next` pointers. Don't
  allocate new posts or copy them into a slice.
- When two posts have the same `Time`, the one from `a` comes first. Posts within
  each feed keep their order.
- Either feed (or both) may be empty (`nil`).

## Example

```
a: ava@90 → bo@40 → ava@10
b: ad@70 → ad@40

mergeFeeds(a, b): ava@90 → ad@70 → bo@40 → ad@40 → ava@10
```

`bo@40` and `ad@40` tie, so `bo` (from `a`) goes first.

## Constraints

- Up to 100,000 posts per feed; `Time` may be any `int`, including negative.
- O(n + m) time and O(1) extra memory.

`feed` and `show` are helpers for building and printing lists; the tests use
them too, so leave them as they are.
