---
title: Profile Cache
difficulty: hard
after: linked-lists
hints:
  - 'You need two things to be fast: finding a handle (a **map**) and keeping handles in recently-used order while moving any one of them to the front (a **doubly linked list**). A slice gives you the order, but moving or removing from its middle is O(n).'
  - 'Store `*list.Element` values in a `map[string]*list.Element` (from `container/list`), with each element''s `Value` holding the handle and follower count. `MoveToFront`, `PushFront`, `Back` and `Remove` are all O(1).'
  - 'On `Put` of a new handle when the cache is full, evict `l.Back()`: remove it from the list **and** delete its handle from the map. That''s why the element needs to remember its handle, not just the count.'
exercise:
  starter: |
    package main

    import "fmt"

    // ProfileCache keeps the follower counts of the most recently used
    // profiles, evicting the least recently used one when it's full.
    type ProfileCache struct {
    	// your fields here
    }

    // NewProfileCache returns an empty cache holding at most capacity
    // profiles (capacity >= 1).
    func NewProfileCache(capacity int) *ProfileCache {
    	return &ProfileCache{}
    }

    // Get returns the cached follower count for handle and marks it as
    // the most recently used. ok is false if handle isn't cached.
    func (c *ProfileCache) Get(handle string) (followers int, ok bool) {
    	return 0, false
    }

    // Put stores followers for handle and marks it as the most recently
    // used, evicting the least recently used profile if the cache is over
    // capacity.
    func (c *ProfileCache) Put(handle string, followers int) {
    }

    // Len returns the number of cached profiles.
    func (c *ProfileCache) Len() int {
    	return 0
    }

    // Handles returns the cached handles, most recently used first.
    func (c *ProfileCache) Handles() []string {
    	return nil
    }

    func main() {
    	c := NewProfileCache(2)
    	c.Put("ava", 900)
    	c.Put("bo", 40)
    	c.Get("ava")
    	c.Put("cy", 7)                    // evicts bo, the least recently used
    	fmt.Println(c.Handles(), c.Len()) // want [cy ava] 2
    	fmt.Println(c.Get("bo"))          // want 0 false
    }
  solution: |
    package main

    import (
    	"container/list"
    	"fmt"
    )

    type profile struct {
    	handle    string
    	followers int
    }

    // ProfileCache keeps the follower counts of the most recently used
    // profiles, evicting the least recently used one when it's full.
    type ProfileCache struct {
    	capacity int
    	order    *list.List               // front = most recently used; values are profile
    	byHandle map[string]*list.Element // handle -> its element in order
    }

    // NewProfileCache returns an empty cache holding at most capacity
    // profiles (capacity >= 1).
    func NewProfileCache(capacity int) *ProfileCache {
    	return &ProfileCache{
    		capacity: capacity,
    		order:    list.New(),
    		byHandle: map[string]*list.Element{},
    	}
    }

    // Get returns the cached follower count for handle and marks it as
    // the most recently used. ok is false if handle isn't cached.
    func (c *ProfileCache) Get(handle string) (followers int, ok bool) {
    	e, ok := c.byHandle[handle]
    	if !ok {
    		return 0, false
    	}
    	c.order.MoveToFront(e)
    	return e.Value.(profile).followers, true
    }

    // Put stores followers for handle and marks it as the most recently
    // used, evicting the least recently used profile if the cache is over
    // capacity.
    func (c *ProfileCache) Put(handle string, followers int) {
    	if e, ok := c.byHandle[handle]; ok {
    		e.Value = profile{handle, followers}
    		c.order.MoveToFront(e)
    		return
    	}
    	c.byHandle[handle] = c.order.PushFront(profile{handle, followers})
    	if c.order.Len() > c.capacity {
    		oldest := c.order.Back()
    		c.order.Remove(oldest)
    		delete(c.byHandle, oldest.Value.(profile).handle)
    	}
    }

    // Len returns the number of cached profiles.
    func (c *ProfileCache) Len() int {
    	return c.order.Len()
    }

    // Handles returns the cached handles, most recently used first.
    func (c *ProfileCache) Handles() []string {
    	handles := make([]string, 0, c.order.Len())
    	for e := c.order.Front(); e != nil; e = e.Next() {
    		handles = append(handles, e.Value.(profile).handle)
    	}
    	return handles
    }

    func main() {
    	c := NewProfileCache(2)
    	c.Put("ava", 900)
    	c.Put("bo", 40)
    	c.Get("ava")
    	c.Put("cy", 7)
    	fmt.Println(c.Handles(), c.Len())
    	fmt.Println(c.Get("bo"))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"slices"
    	"testing"
    	"time"
    )

    func expect(t *testing.T, step string, c *ProfileCache, want []string) {
    	t.Helper()
    	if got := c.Handles(); !slices.Equal(got, want) {
    		t.Errorf("after %s: Handles() = %q, want %q", step, got, want)
    	}
    	if got := c.Len(); got != len(want) {
    		t.Errorf("after %s: Len() = %d, want %d", step, got, len(want))
    	}
    }

    func TestProfileCache(t *testing.T) {
    	c := NewProfileCache(3)
    	expect(t, "NewProfileCache(3)", c, []string{})
    	if f, ok := c.Get("ava"); f != 0 || ok {
    		t.Errorf("Get(%q) on an empty cache = %d, %v, want 0, false", "ava", f, ok)
    	}
    	c.Put("ava", 900)
    	c.Put("bo", 40)
    	c.Put("cy", 7)
    	expect(t, `Put ava, bo, cy`, c, []string{"cy", "bo", "ava"})
    	if f, ok := c.Get("ava"); f != 900 || !ok {
    		t.Errorf("Get(%q) = %d, %v, want 900, true", "ava", f, ok)
    	}
    	expect(t, `Get("ava")`, c, []string{"ava", "cy", "bo"})
    	c.Put("dee", 12)
    	expect(t, `Put("dee") into a full cache`, c, []string{"dee", "ava", "cy"})
    	if f, ok := c.Get("bo"); f != 0 || ok {
    		t.Errorf("Get(%q) after it was evicted = %d, %v, want 0, false", "bo", f, ok)
    	}
    	expect(t, `Get("bo") (a miss)`, c, []string{"dee", "ava", "cy"})
    	c.Put("cy", 8)
    	expect(t, `Put("cy", 8) (an update)`, c, []string{"cy", "dee", "ava"})
    	if f, ok := c.Get("cy"); f != 8 || !ok {
    		t.Errorf("Get(%q) after updating it to 8 = %d, %v, want 8, true", "cy", f, ok)
    	}
    	c.Put("eve", -1)
    	expect(t, `Put("eve")`, c, []string{"eve", "cy", "dee"})
    	if f, ok := c.Get("eve"); f != -1 || !ok {
    		t.Errorf("Get(%q) = %d, %v, want -1, true", "eve", f, ok)
    	}
    }

    func TestProfileCacheCapacityOne(t *testing.T) {
    	c := NewProfileCache(1)
    	c.Put("ava", 1)
    	c.Put("ava", 2)
    	expect(t, `Put ava twice`, c, []string{"ava"})
    	c.Put("bo", 3)
    	expect(t, `Put("bo")`, c, []string{"bo"})
    	if f, ok := c.Get("ava"); ok {
    		t.Errorf("Get(%q) = %d, true, want it evicted", "ava", f)
    	}
    }

    func TestTwoCaches(t *testing.T) {
    	a, b := NewProfileCache(2), NewProfileCache(2)
    	a.Put("ava", 1)
    	if _, ok := b.Get("ava"); ok {
    		t.Errorf("two caches share state: b.Get(%q) found a's profile", "ava")
    	}
    }

    func TestProfileCacheLarge(t *testing.T) {
    	capacity, n := 100_000, 300_000
    	handles := make([]string, n)
    	for i := range handles {
    		handles[i] = fmt.Sprint("user", i)
    	}
    	type result struct {
    		hits, len int
    		bad       string // the first wrong Get, if any
    	}
    	done := make(chan result, 1)
    	go func() {
    		c := NewProfileCache(capacity)
    		var res result
    		var x uint64 = 42
    		for i, h := range handles {
    			c.Put(h, i)
    			// Look up a random earlier profile: some are still cached, some evicted.
    			x = x*6364136223846793005 + 1442695040888963407
    			j := int(x>>33) % (i + 1)
    			if f, ok := c.Get(handles[j]); ok {
    				if f != j && res.bad == "" {
    					res.bad = fmt.Sprintf("Get(%q) = %d, want %d", handles[j], f, j)
    				}
    				res.hits++
    			}
    		}
    		res.len = c.Len()
    		done <- res
    	}()
    	select {
    	case got := <-done:
    		if got.bad != "" {
    			t.Errorf("%d Put and Get calls with capacity %d: %s", n, capacity, got.bad)
    		}
    		if got.hits != 209_581 || got.len != capacity {
    			t.Errorf("%d Put and Get calls with capacity %d: %d hits and Len() = %d, want 209581 hits and Len() = %d", n, capacity, got.hits, got.len, capacity)
    		}
    	case <-time.After(time.Second):
    		t.Fatalf("%d Put and Get calls with capacity %d took over a second: every operation must be O(1)", 2*n, capacity)
    	}
    }
---

Loading a profile's follower count from the database is slow, so the Clout app
keeps a **cache** of recently viewed profiles in memory. It can only hold
`capacity` profiles; when it's full, it evicts the profile that was **used
least recently**. (This is known as an *LRU cache*.)

Implement `ProfileCache`:

- `NewProfileCache(capacity)` returns an empty cache (`capacity ≥ 1`).
- `Get(handle)` returns the follower count and `true` if `handle` is cached, and
  marks it as the most recently used. A miss returns `0, false` and changes
  nothing.
- `Put(handle, followers)` stores the count (replacing any old value) and marks
  `handle` as the most recently used. If that takes the cache over capacity,
  evict the least recently used profile.
- `Len()` returns the number of cached profiles.
- `Handles()` returns the cached handles, **most recently used first**. It's
  used for debugging and by the tests, so it may be O(n).

## Example

```go
c := NewProfileCache(2)
c.Put("ava", 900)  // [ava]
c.Put("bo", 40)    // [bo ava]
c.Get("ava")       // 900, true    -> [ava bo]
c.Put("cy", 7)     // full: evict bo -> [cy ava]
c.Get("bo")        // 0, false
```

## Constraints

- Capacity up to 100,000 and up to 600,000 operations.
- `Get`, `Put` and `Len` must be **O(1)**. The performance test runs 600,000
  of them under a one-second limit, so keeping the order in a slice (which
  shifts up to 100,000 handles on every use) won't finish in time.
- `container/list` from the standard library is allowed, or you can build your
  own doubly linked list.
