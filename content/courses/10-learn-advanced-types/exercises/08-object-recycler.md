---
title: Object Recycler
difficulty: medium
after: designing-generic-apis
hints:
  - '`PT interface{ *T; Reset() }` says two things: `PT` *is* `*T`, and it has a `Reset` method. So `PT(new(T))` turns a fresh `*T` into a `PT` you can call `Reset` on, and callers only write `NewRecycler[Buffer](4)`: Go infers `PT = *Buffer` from the constraint.'
  - 'Keep the free objects in a `[]PT` used as a stack: `Put` appends, `Get` pops the last one (most recently returned first). Count `created` and `reused` as you go.'
  - '`PT` is a pointer type, so its zero value is `nil`, and `p == nil` compiles because every pointer type is comparable. Check it first in `Put` so you never call `Reset` on a nil pointer.'
exercise:
  starter: |
    package main

    import "fmt"

    // Recycler hands out *T values and takes them back for reuse, calling
    // Reset on every value that is returned. It keeps at most max free values.
    type Recycler[T any, PT interface {
    	*T
    	Reset()
    }] struct {
    	// your fields here
    }

    // NewRecycler returns an empty recycler that keeps at most max free values.
    func NewRecycler[T any, PT interface {
    	*T
    	Reset()
    }](max int) *Recycler[T, PT] {
    	return &Recycler[T, PT]{}
    }

    // Get returns the most recently returned free value, or a new zero T if
    // there are none.
    func (r *Recycler[T, PT]) Get() PT {
    	return nil
    }

    // Put resets p and keeps it for reuse, unless max values are already free
    // (then p is dropped). A nil p is ignored.
    func (r *Recycler[T, PT]) Put(p PT) {
    }

    // Stats reports how many values Get has created and how many it has reused.
    func (r *Recycler[T, PT]) Stats() (created, reused int) {
    	return 0, 0
    }

    // Buffer collects encoded bytes for one Stash write.
    type Buffer struct {
    	data []byte
    }

    func (b *Buffer) Write(s string) { b.data = append(b.data, s...) }
    func (b *Buffer) Reset()         { b.data = b.data[:0] }

    func main() {
    	r := NewRecycler[Buffer](2) // PT is inferred as *Buffer
    	b := r.Get()
    	fmt.Println("got a buffer:", b != nil) // want true
    	if b != nil {
    		b.Write("user:1=ada")
    		r.Put(b)
    	}
    	again := r.Get()
    	fmt.Println("same buffer:", again == b && again != nil) // want true
    	if again != nil {
    		fmt.Printf("reset: %q\n", again.data) // want ""
    	}
    	fmt.Println(r.Stats()) // want 1 1
    }
  solution: |
    package main

    import "fmt"

    // Recycler hands out *T values and takes them back for reuse, calling
    // Reset on every value that is returned. It keeps at most max free values.
    type Recycler[T any, PT interface {
    	*T
    	Reset()
    }] struct {
    	free            []PT
    	max             int
    	created, reused int
    }

    // NewRecycler returns an empty recycler that keeps at most max free values.
    func NewRecycler[T any, PT interface {
    	*T
    	Reset()
    }](max int) *Recycler[T, PT] {
    	return &Recycler[T, PT]{max: max}
    }

    // Get returns the most recently returned free value, or a new zero T if
    // there are none.
    func (r *Recycler[T, PT]) Get() PT {
    	if n := len(r.free); n > 0 {
    		p := r.free[n-1]
    		r.free[n-1] = nil
    		r.free = r.free[:n-1]
    		r.reused++
    		return p
    	}
    	r.created++
    	return PT(new(T))
    }

    // Put resets p and keeps it for reuse, unless max values are already free
    // (then p is dropped). A nil p is ignored.
    func (r *Recycler[T, PT]) Put(p PT) {
    	if p == nil {
    		return
    	}
    	p.Reset()
    	if len(r.free) < r.max {
    		r.free = append(r.free, p)
    	}
    }

    // Stats reports how many values Get has created and how many it has reused.
    func (r *Recycler[T, PT]) Stats() (created, reused int) {
    	return r.created, r.reused
    }

    // Buffer collects encoded bytes for one Stash write.
    type Buffer struct {
    	data []byte
    }

    func (b *Buffer) Write(s string) { b.data = append(b.data, s...) }
    func (b *Buffer) Reset()         { b.data = b.data[:0] }

    func main() {
    	r := NewRecycler[Buffer](2)
    	b := r.Get()
    	fmt.Println("got a buffer:", b != nil)
    	b.Write("user:1=ada")
    	r.Put(b)
    	again := r.Get()
    	fmt.Println("same buffer:", again == b)
    	fmt.Printf("reset: %q\n", again.data)
    	fmt.Println(r.Stats())
    }
  tests: |
    package main

    import "testing"

    // Counter is a named non-struct type with a pointer Reset.
    type Counter int

    func (c *Counter) Reset() { *c = 0 }

    // Session counts its resets so the tests can see when Reset ran.
    type Session struct {
    	User   string
    	Resets int
    }

    func (s *Session) Reset() {
    	s.User = ""
    	s.Resets++
    }

    // Label has a value-receiver Reset; *Label still has it in its method set.
    type Label struct{ Name string }

    func (Label) Reset() {}

    func TestGetCreatesZeroValues(t *testing.T) {
    	r := NewRecycler[Session](3)
    	a, b := r.Get(), r.Get()
    	if a == nil || b == nil {
    		t.Fatalf("Get() on an empty recycler returned %v and %v, want two new *Session values", a, b)
    	}
    	if a == b {
    		t.Errorf("two Get() calls on an empty recycler returned the same pointer, want two new values")
    	}
    	if *a != (Session{}) {
    		t.Errorf("Get() on an empty recycler = %+v, want a zero Session", *a)
    	}
    	if c, u := r.Stats(); c != 2 || u != 0 {
    		t.Errorf("Stats() after two Gets = %d, %d, want 2, 0", c, u)
    	}
    }

    func TestPutResetsAndReuses(t *testing.T) {
    	r := NewRecycler[Session](3)
    	s := r.Get()
    	if s == nil {
    		t.Fatal("Get() returned nil")
    	}
    	s.User = "ada"
    	r.Put(s)
    	if s.User != "" || s.Resets != 1 {
    		t.Errorf("after Put, the session is %+v, want User \"\" and Resets 1 (Put must call Reset once)", *s)
    	}
    	if got := r.Get(); got != s {
    		t.Errorf("Get() after Put(s) returned %p, want the recycled %p", got, s)
    	}
    	if c, u := r.Stats(); c != 1 || u != 1 {
    		t.Errorf("Stats() = %d, %d, want 1, 1", c, u)
    	}
    }

    func TestReuseIsLastInFirstOut(t *testing.T) {
    	r := NewRecycler[Counter](5)
    	a, b, c := r.Get(), r.Get(), r.Get()
    	if a == nil || b == nil || c == nil {
    		t.Fatal("Get() returned nil")
    	}
    	*a, *b, *c = 1, 2, 3
    	r.Put(a)
    	r.Put(b)
    	r.Put(c)
    	if *a != 0 || *b != 0 || *c != 0 {
    		t.Errorf("after Put, counters are %d %d %d, want 0 0 0", *a, *b, *c)
    	}
    	got := []*Counter{r.Get(), r.Get(), r.Get()}
    	if got[0] != c || got[1] != b || got[2] != a {
    		t.Errorf("Get order after Put(a), Put(b), Put(c) = %p %p %p, want c=%p b=%p a=%p", got[0], got[1], got[2], c, b, a)
    	}
    	if d := r.Get(); d == nil || d == a || d == b || d == c {
    		t.Errorf("Get() with nothing free returned %p, want a brand new *Counter", d)
    	}
    	if cr, u := r.Stats(); cr != 4 || u != 3 {
    		t.Errorf("Stats() = %d, %d, want 4, 3", cr, u)
    	}
    }

    func TestMaxFree(t *testing.T) {
    	r := NewRecycler[Session](2)
    	ss := []*Session{r.Get(), r.Get(), r.Get()}
    	for _, s := range ss {
    		if s == nil {
    			t.Fatal("Get() returned nil")
    		}
    		s.User = "x"
    		r.Put(s)
    	}
    	if ss[2].Resets != 1 || ss[2].User != "" {
    		t.Errorf("a dropped value should still be reset: got %+v", *ss[2])
    	}
    	g1, g2, g3 := r.Get(), r.Get(), r.Get()
    	if g1 != ss[1] || g2 != ss[0] {
    		t.Errorf("with max 2, Get after 3 Puts returned %p, %p, want the first two returned values %p, %p", g1, g2, ss[1], ss[0])
    	}
    	if g3 == ss[2] {
    		t.Errorf("with max 2, the third returned value should have been dropped, but Get handed it out again")
    	}

    	zero := NewRecycler[Session](0)
    	s := zero.Get()
    	zero.Put(s)
    	if zero.Get() == s {
    		t.Errorf("a recycler with max 0 kept a value; it should drop everything")
    	}
    }

    func TestPutNil(t *testing.T) {
    	r := NewRecycler[Counter](2)
    	r.Put(nil)
    	if c := r.Get(); c == nil {
    		t.Errorf("after Put(nil), Get() returned nil: Put must ignore nil")
    	}
    	if c, u := r.Stats(); c != 1 || u != 0 {
    		t.Errorf("Stats() after Put(nil) and Get() = %d, %d, want 1, 0", c, u)
    	}
    }

    func TestValueReceiverReset(t *testing.T) {
    	r := NewRecycler[Label](1)
    	l := r.Get()
    	if l == nil {
    		t.Fatal("Get() returned nil")
    	}
    	l.Name = "keep"
    	r.Put(l)
    	if got := r.Get(); got != l || got.Name != "keep" {
    		t.Errorf("Recycler[Label] Get after Put = %p %+v, want %p with Name \"keep\" (Label.Reset does nothing)", got, got, l)
    	}
    }

    func TestRecyclersAreIndependent(t *testing.T) {
    	a, b := NewRecycler[Counter](2), NewRecycler[Counter](2)
    	c := a.Get()
    	a.Put(c)
    	if got := b.Get(); got == c {
    		t.Errorf("a value Put into one recycler came out of another")
    	}
    }
---

Stash encodes thousands of writes a second, and allocating a fresh buffer for
each one keeps the garbage collector busy. A **recycler** hands out objects and
takes them back for reuse.

Complete the generic `Recycler[T, PT]`. `T` is the object type and `PT` is `*T`
with a `Reset()` method, so callers only name `T`:

```go
r := NewRecycler[Buffer](4) // PT is inferred as *Buffer
```

- `Get()` returns the **most recently returned** free object. If none is free,
  it returns a pointer to a brand-new zero `T`.
- `Put(p)` calls `p.Reset()`, then keeps `p` for reuse, unless `max` objects
  are already free, in which case `p` is dropped (still reset). `Put(nil)` does
  nothing.
- `Stats()` reports how many objects `Get` has **created** and how many it has
  **reused**.

## Example

```go
r := NewRecycler[Buffer](2)
b := r.Get()          // new Buffer
b.Write("user:1=ada")
r.Put(b)              // b.Reset() runs, b is kept
r.Get() == b          // true, and its data is empty again
r.Stats()             // 1, 1
```

## Constraints

- The hidden tests use their own types: a struct with a pointer `Reset`, a
  named integer type `type Counter int` with a pointer `Reset`, and a struct
  whose `Reset` has a value receiver.
- Separate recyclers never share objects.
- Don't change the type parameters or signatures: the tests rely on
  `NewRecycler[T](max)` inferring `PT`.
