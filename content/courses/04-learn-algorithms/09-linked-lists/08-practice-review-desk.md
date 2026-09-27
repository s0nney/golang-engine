---
title: 'Practice: The Review Desk'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    	"slices"
    )

    // ---- Pieces you built earlier in the course (already done) ----

    type Node[T any] struct {
    	Value T
    	Next  *Node[T]
    }

    // LinkedList is a singly linked list with O(1) PushFront, PushBack and PopFront.
    type LinkedList[T any] struct {
    	head, tail *Node[T]
    	size       int
    }

    func (l *LinkedList[T]) Len() int { return l.size }

    func (l *LinkedList[T]) PushFront(v T) {
    	l.head = &Node[T]{Value: v, Next: l.head}
    	if l.tail == nil {
    		l.tail = l.head
    	}
    	l.size++
    }

    func (l *LinkedList[T]) PushBack(v T) {
    	n := &Node[T]{Value: v}
    	if l.tail == nil {
    		l.head, l.tail = n, n
    	} else {
    		l.tail.Next = n
    		l.tail = n
    	}
    	l.size++
    }

    func (l *LinkedList[T]) PopFront() (T, bool) {
    	if l.head == nil {
    		var zero T
    		return zero, false
    	}
    	n := l.head
    	l.head = n.Next
    	if l.head == nil {
    		l.tail = nil
    	}
    	l.size--
    	return n.Value, true
    }

    // Stack is a LIFO stack.
    type Stack[T any] struct{ items []T }

    func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }
    func (s *Stack[T]) Pop() (T, bool) {
    	var zero T
    	if len(s.items) == 0 {
    		return zero, false
    	}
    	v := s.items[len(s.items)-1]
    	s.items[len(s.items)-1] = zero
    	s.items = s.items[:len(s.items)-1]
    	return v, true
    }

    // ---- The review desk ----

    type Post struct {
    	ID     int
    	Author string
    	Likes  int
    }

    // Decision records what a moderator did with a post.
    type Decision struct {
    	Post     Post
    	Approved bool
    }

    // Desk is Clout's moderation desk. The zero value is ready to use.
    type Desk struct {
    	pending LinkedList[Post] // posts waiting for review, oldest first
    	feed    []Post           // approved posts, in the order they were approved
    	history Stack[Decision]  // every decision, most recent on top
    }

    // Pending returns how many posts are waiting for review.
    func (d *Desk) Pending() int { return d.pending.Len() }

    // Feed returns a copy of the approved posts, in approval order.
    func (d *Desk) Feed() []Post { return slices.Clone(d.feed) }

    // Submit adds p to the back of the review queue.
    func (d *Desk) Submit(p Post) {
    	// ?
    }

    // Approve takes the oldest pending post, adds it to the feed, records the
    // decision and returns the post. It returns false if nothing is pending.
    func (d *Desk) Approve() (Post, bool) {
    	// ?
    	return Post{}, false
    }

    // Reject takes the oldest pending post, drops it, records the decision and
    // returns the post. It returns false if nothing is pending.
    func (d *Desk) Reject() (Post, bool) {
    	// ?
    	return Post{}, false
    }

    // Undo reverses the most recent decision: the post leaves the feed (if it
    // was approved) and goes back to the FRONT of the queue, so it's next up
    // for review again. It returns false if there's nothing to undo.
    func (d *Desk) Undo() bool {
    	// ?
    	return false
    }

    // Top returns up to n approved posts, most liked first. Posts with the same
    // likes stay in approval order. It must not reorder the feed.
    func (d *Desk) Top(n int) []Post {
    	// ?
    	return nil
    }

    func main() {
    	var d Desk
    	d.Submit(Post{1, "ava", 120})
    	d.Submit(Post{2, "bo", 900})
    	d.Submit(Post{3, "cy", 450})
    	d.Approve()                        // ava
    	d.Reject()                         // bo
    	d.Undo()                           // bo is back at the front
    	d.Approve()                        // bo
    	d.Approve()                        // cy
    	fmt.Println(d.Feed(), d.Pending()) // want: [{1 ava 120} {2 bo 900} {3 cy 450}] 0
    	fmt.Println(d.Top(2))              // want: [{2 bo 900} {3 cy 450}]
    }
  solution: |
    package main

    import (
    	"cmp"
    	"fmt"
    	"slices"
    )

    // ---- Pieces you built earlier in the course (already done) ----

    type Node[T any] struct {
    	Value T
    	Next  *Node[T]
    }

    // LinkedList is a singly linked list with O(1) PushFront, PushBack and PopFront.
    type LinkedList[T any] struct {
    	head, tail *Node[T]
    	size       int
    }

    func (l *LinkedList[T]) Len() int { return l.size }

    func (l *LinkedList[T]) PushFront(v T) {
    	l.head = &Node[T]{Value: v, Next: l.head}
    	if l.tail == nil {
    		l.tail = l.head
    	}
    	l.size++
    }

    func (l *LinkedList[T]) PushBack(v T) {
    	n := &Node[T]{Value: v}
    	if l.tail == nil {
    		l.head, l.tail = n, n
    	} else {
    		l.tail.Next = n
    		l.tail = n
    	}
    	l.size++
    }

    func (l *LinkedList[T]) PopFront() (T, bool) {
    	if l.head == nil {
    		var zero T
    		return zero, false
    	}
    	n := l.head
    	l.head = n.Next
    	if l.head == nil {
    		l.tail = nil
    	}
    	l.size--
    	return n.Value, true
    }

    // Stack is a LIFO stack.
    type Stack[T any] struct{ items []T }

    func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }
    func (s *Stack[T]) Pop() (T, bool) {
    	var zero T
    	if len(s.items) == 0 {
    		return zero, false
    	}
    	v := s.items[len(s.items)-1]
    	s.items[len(s.items)-1] = zero
    	s.items = s.items[:len(s.items)-1]
    	return v, true
    }

    // ---- The review desk ----

    type Post struct {
    	ID     int
    	Author string
    	Likes  int
    }

    // Decision records what a moderator did with a post.
    type Decision struct {
    	Post     Post
    	Approved bool
    }

    // Desk is Clout's moderation desk. The zero value is ready to use.
    type Desk struct {
    	pending LinkedList[Post] // posts waiting for review, oldest first
    	feed    []Post           // approved posts, in the order they were approved
    	history Stack[Decision]  // every decision, most recent on top
    }

    // Pending returns how many posts are waiting for review.
    func (d *Desk) Pending() int { return d.pending.Len() }

    // Feed returns a copy of the approved posts, in approval order.
    func (d *Desk) Feed() []Post { return slices.Clone(d.feed) }

    func (d *Desk) Submit(p Post) {
    	d.pending.PushBack(p)
    }

    func (d *Desk) Approve() (Post, bool) {
    	p, ok := d.pending.PopFront()
    	if !ok {
    		return Post{}, false
    	}
    	d.feed = append(d.feed, p)
    	d.history.Push(Decision{Post: p, Approved: true})
    	return p, true
    }

    func (d *Desk) Reject() (Post, bool) {
    	p, ok := d.pending.PopFront()
    	if !ok {
    		return Post{}, false
    	}
    	d.history.Push(Decision{Post: p, Approved: false})
    	return p, true
    }

    func (d *Desk) Undo() bool {
    	last, ok := d.history.Pop()
    	if !ok {
    		return false
    	}
    	if last.Approved {
    		// Later approvals were undone first, so this post is the last one in the feed.
    		d.feed[len(d.feed)-1] = Post{}
    		d.feed = d.feed[:len(d.feed)-1]
    	}
    	d.pending.PushFront(last.Post)
    	return true
    }

    func (d *Desk) Top(n int) []Post {
    	ranked := slices.Clone(d.feed)
    	slices.SortStableFunc(ranked, func(a, b Post) int {
    		return cmp.Compare(b.Likes, a.Likes)
    	})
    	return ranked[:min(n, len(ranked))]
    }

    func main() {
    	var d Desk
    	d.Submit(Post{1, "ava", 120})
    	d.Submit(Post{2, "bo", 900})
    	d.Submit(Post{3, "cy", 450})
    	d.Approve()                        // ava
    	d.Reject()                         // bo
    	d.Undo()                           // bo is back at the front
    	d.Approve()                        // bo
    	d.Approve()                        // cy
    	fmt.Println(d.Feed(), d.Pending()) // want: [{1 ava 120} {2 bo 900} {3 cy 450}] 0
    	fmt.Println(d.Top(2))              // want: [{2 bo 900} {3 cy 450}]
    }
  tests: |
    package main

    import (
    	"slices"
    	"testing"
    )

    var (
    	ava = Post{1, "ava", 120}
    	bo  = Post{2, "bo", 900}
    	cy  = Post{3, "cy", 450}
    	dee = Post{4, "dee", 900}
    	eve = Post{5, "eve", 10}
    )

    func TestApproveAndRejectAreFIFO(t *testing.T) {
    	var d Desk
    	if _, ok := d.Approve(); ok {
    		t.Fatal("Approve() on an empty desk returned ok = true")
    	}
    	if _, ok := d.Reject(); ok {
    		t.Fatal("Reject() on an empty desk returned ok = true")
    	}
    	for _, p := range []Post{ava, bo, cy} {
    		d.Submit(p)
    	}
    	if d.Pending() != 3 {
    		t.Fatalf("after 3 Submits, Pending() = %d, want 3", d.Pending())
    	}
    	if p, ok := d.Approve(); !ok || p != ava {
    		t.Fatalf("first Approve() = %v, %v; want %v, true (oldest first)", p, ok, ava)
    	}
    	if p, ok := d.Reject(); !ok || p != bo {
    		t.Fatalf("Reject() = %v, %v; want %v, true", p, ok, bo)
    	}
    	d.Approve()
    	if got, want := d.Feed(), []Post{ava, cy}; !slices.Equal(got, want) {
    		t.Errorf("Feed() = %v, want %v (rejected posts stay out)", got, want)
    	}
    	if d.Pending() != 0 {
    		t.Errorf("Pending() = %d, want 0", d.Pending())
    	}
    }

    func TestUndo(t *testing.T) {
    	var d Desk
    	if d.Undo() {
    		t.Fatal("Undo() with no decisions returned true")
    	}
    	for _, p := range []Post{ava, bo, cy} {
    		d.Submit(p)
    	}
    	d.Approve() // ava
    	d.Reject()  // bo
    	d.Approve() // cy
    	if !d.Undo() {
    		t.Fatal("Undo() = false, want true")
    	}
    	if got, want := d.Feed(), []Post{ava}; !slices.Equal(got, want) {
    		t.Fatalf("after undoing cy's approval, Feed() = %v, want %v", got, want)
    	}
    	d.Undo() // bo's rejection
    	if d.Pending() != 2 {
    		t.Fatalf("after two Undos, Pending() = %d, want 2", d.Pending())
    	}
    	if p, _ := d.Approve(); p != bo {
    		t.Fatalf("after undoing, the next Approve() gave %v, want %v: undone posts go back to the FRONT", p, bo)
    	}
    	if p, _ := d.Approve(); p != cy {
    		t.Fatalf("then Approve() gave %v, want %v", p, cy)
    	}
    	d.Undo()
    	d.Undo()
    	d.Undo()
    	if got := d.Feed(); len(got) != 0 {
    		t.Errorf("after undoing every approval, Feed() = %v, want empty", got)
    	}
    	if d.Pending() != 3 {
    		t.Errorf("after undoing everything, Pending() = %d, want 3", d.Pending())
    	}
    	if p, _ := d.Reject(); p != ava {
    		t.Errorf("after undoing everything, the front of the queue is %v, want %v", p, ava)
    	}
    }

    func TestTop(t *testing.T) {
    	var d Desk
    	if got := d.Top(3); len(got) != 0 {
    		t.Errorf("Top(3) on an empty feed = %v, want empty", got)
    	}
    	for _, p := range []Post{ava, bo, cy, dee, eve} {
    		d.Submit(p)
    		d.Approve()
    	}
    	if got, want := d.Top(3), []Post{bo, dee, cy}; !slices.Equal(got, want) {
    		t.Errorf("Top(3) = %v, want %v (ties keep approval order)", got, want)
    	}
    	if got, want := d.Top(10), []Post{bo, dee, cy, ava, eve}; !slices.Equal(got, want) {
    		t.Errorf("Top(10) with 5 posts = %v, want %v", got, want)
    	}
    	if got := d.Top(0); len(got) != 0 {
    		t.Errorf("Top(0) = %v, want empty", got)
    	}
    }

    func TestTopDoesNotTouchFeed(t *testing.T) {
    	var d Desk
    	for _, p := range []Post{ava, bo, cy} {
    		d.Submit(p)
    		d.Approve()
    	}
    	top := d.Top(3)
    	if got, want := d.Feed(), []Post{ava, bo, cy}; !slices.Equal(got, want) {
    		t.Fatalf("after Top(3), Feed() = %v, want %v: Top must sort a copy", got, want)
    	}
    	if len(top) > 0 {
    		top[0].Likes = -1
    	}
    	if got := d.Feed(); slices.ContainsFunc(got, func(p Post) bool { return p.Likes == -1 }) {
    		t.Errorf("changing Top's result changed the feed: %v", got)
    	}
    }
---

Time to put the course together. Every post on Clout from a brand-new account
goes to a human moderator before it reaches anyone's feed. You're building the
**review desk** they use, and it needs three of the structures you've built:

- a **queue** of posts waiting for review, oldest first,
- a **stack** of decisions, so a moderator can **undo** a misclick,
- and a **sorted view** of the approved feed for the "top posts" panel.

The `LinkedList` and `Stack` from earlier chapters are already in the editor,
finished. Your job is the `Desk` that wires them together.

## Why a linked list for the queue?

A ring buffer would handle submit and approve just fine. But look at what undo
needs: when a moderator undoes a decision, the post goes back to the **front**
of the queue, so it's the next one they see. That's a `PushFront`, which a
singly linked list does in O(1). So `pending` is a `LinkedList[Post]` used as
a queue (`PushBack` to join, `PopFront` to leave) that also allows cutting in
at the front.

## Why a stack for undo?

Undo always reverses the **most recent** decision first: last in, first out.
Each `Decision` remembers the post and whether it was approved, which is all
you need to reverse it.

That LIFO order gives you a nice guarantee. If the decision you're undoing was
an approval, every approval made after it has already been undone. So its post
is always the **last** one in `feed`: you can remove it by shortening the
slice, with no searching. (Zero the slot before you truncate, like `Pop` does,
so the old post isn't kept alive.)

## Your task

Complete these `Desk` methods:

- **`Submit(p)`** adds `p` to the back of the queue.
- **`Approve()`** pops the oldest pending post, appends it to `feed`, pushes a
  `Decision{Post: p, Approved: true}` and returns the post. If nothing is
  pending, return `Post{}, false`.
- **`Reject()`** is the same, but the post doesn't join the feed.
- **`Undo()`** pops the latest decision. If it was an approval, remove the post
  from the end of `feed`. Either way, put the post back at the **front** of the
  queue. Return `false` if there's nothing to undo.
- **`Top(n)`** returns up to `n` approved posts, most likes first. Posts with
  equal likes must stay in **approval order**, so you need a **stable** sort:
  `slices.SortStableFunc` with `cmp.Compare(b.Likes, a.Likes)` (note the
  swapped arguments for descending order). Add `"cmp"` to the imports.

`Top` must **not** reorder `feed`. Sorting `d.feed` in place would scramble the
approval order that `Undo` depends on! Sort a copy from `slices.Clone`
instead, and return at most `min(n, len(copy))` posts.

**Run** approves, rejects and undoes a few posts. **Submit** checks FIFO order,
undo after every mix of decisions, stable ranking, and that `Top` leaves the
feed alone.

## Course complete

That's Data Structures and Algorithms 1 done. You've measured algorithms with
Big O, sorted five different ways, met exponential time, and built stacks,
queues and linked lists from scratch. Now you've made them work together.

Next up is [Data Structures and Algorithms 2](/courses/learn-data-structures),
which builds on these foundations with trees, hashmaps, tries, heaps and
graphs.
