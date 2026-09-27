---
title: Like Counter
difficulty: easy
after: storage
hints:
  - 'A set of users per squeak works well: `map[string]map[string]struct{}` from squeak ID to the IDs of users who liked it. Remember that the inner maps start out `nil` and have to be made before you write to them.'
  - 'Every request runs on its own goroutine, so every method (reads included) must hold a `sync.Mutex` while it touches the maps. `mu.Lock()` then `defer mu.Unlock()` at the top of each method is the simplest safe pattern.'
exercise:
  starter: |
    package main

    import (
    	"fmt"
    )

    // LikeStore remembers which users liked which squeaks.
    // It must be safe to use from many goroutines at once.
    type LikeStore struct {
    	// your fields here (a mutex and a map or two)
    }

    func NewLikeStore() *LikeStore {
    	return &LikeStore{}
    }

    // Like records that userID liked squeakID. It returns true if this is a
    // new like, and false if userID had already liked it.
    func (s *LikeStore) Like(squeakID, userID string) bool {
    	return false
    }

    // Unlike removes userID's like from squeakID. It returns true if there
    // was a like to remove.
    func (s *LikeStore) Unlike(squeakID, userID string) bool {
    	return false
    }

    // Count returns how many users currently like squeakID.
    func (s *LikeStore) Count(squeakID string) int {
    	return 0
    }

    func main() {
    	s := NewLikeStore()
    	fmt.Println(s.Like("sq1", "pip"), s.Like("sq1", "pip"), s.Like("sq1", "bree")) // want: true false true
    	fmt.Println(s.Count("sq1"), s.Count("sq2"))                                    // want: 2 0
    	fmt.Println(s.Unlike("sq1", "pip"), s.Unlike("sq1", "pip"), s.Count("sq1"))    // want: true false 1
    }
  solution: |
    package main

    import (
    	"fmt"
    	"sync"
    )

    // LikeStore remembers which users liked which squeaks.
    // It must be safe to use from many goroutines at once.
    type LikeStore struct {
    	mu    sync.Mutex
    	likes map[string]map[string]struct{} // squeak ID -> user IDs
    }

    func NewLikeStore() *LikeStore {
    	return &LikeStore{likes: make(map[string]map[string]struct{})}
    }

    func (s *LikeStore) Like(squeakID, userID string) bool {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	users := s.likes[squeakID]
    	if users == nil {
    		users = make(map[string]struct{})
    		s.likes[squeakID] = users
    	}
    	if _, ok := users[userID]; ok {
    		return false
    	}
    	users[userID] = struct{}{}
    	return true
    }

    func (s *LikeStore) Unlike(squeakID, userID string) bool {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	users := s.likes[squeakID]
    	if _, ok := users[userID]; !ok {
    		return false
    	}
    	delete(users, userID)
    	return true
    }

    func (s *LikeStore) Count(squeakID string) int {
    	s.mu.Lock()
    	defer s.mu.Unlock()
    	return len(s.likes[squeakID])
    }

    func main() {
    	s := NewLikeStore()
    	fmt.Println(s.Like("sq1", "pip"), s.Like("sq1", "pip"), s.Like("sq1", "bree"))
    	fmt.Println(s.Count("sq1"), s.Count("sq2"))
    	fmt.Println(s.Unlike("sq1", "pip"), s.Unlike("sq1", "pip"), s.Count("sq1"))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"sync"
    	"sync/atomic"
    	"testing"
    )

    func TestLikeUnlike(t *testing.T) {
    	s := NewLikeStore()
    	steps := []struct {
    		op, squeak, user string
    		want             bool
    		count            int
    	}{
    		{"like", "sq1", "pip", true, 1},
    		{"like", "sq1", "pip", false, 1},
    		{"like", "sq1", "bree", true, 2},
    		{"unlike", "sq1", "pip", true, 1},
    		{"unlike", "sq1", "pip", false, 1},
    		{"unlike", "sq1", "nobody", false, 1},
    		{"like", "sq1", "pip", true, 2},
    		{"unlike", "sq9", "pip", false, 0},
    	}
    	for i, st := range steps {
    		var got bool
    		if st.op == "like" {
    			got = s.Like(st.squeak, st.user)
    		} else {
    			got = s.Unlike(st.squeak, st.user)
    		}
    		if got != st.want {
    			t.Errorf("step %d: %s(%q, %q) = %v, want %v", i+1, st.op, st.squeak, st.user, got, st.want)
    		}
    		if c := s.Count(st.squeak); c != st.count {
    			t.Errorf("step %d: after %s(%q, %q), Count(%q) = %d, want %d", i+1, st.op, st.squeak, st.user, st.squeak, c, st.count)
    		}
    	}
    }

    func TestSqueaksAreSeparate(t *testing.T) {
    	s := NewLikeStore()
    	s.Like("sq1", "pip")
    	if !s.Like("sq2", "pip") {
    		t.Errorf("pip liking sq2 after liking sq1 returned false: likes are per squeak")
    	}
    	if c := s.Count("never-liked"); c != 0 {
    		t.Errorf("Count of a squeak nobody liked = %d, want 0", c)
    	}
    	other := NewLikeStore()
    	if c := other.Count("sq1"); c != 0 {
    		t.Errorf("a second LikeStore already has %d likes on sq1: stores must not share state", c)
    	}
    }

    func TestConcurrentLikes(t *testing.T) {
    	s := NewLikeStore()
    	var wins atomic.Int64
    	var wg sync.WaitGroup
    	for i := range 200 {
    		user := fmt.Sprintf("user%d", i%50) // every user tries 4 times
    		wg.Go(func() {
    			if s.Like("hot", user) {
    				wins.Add(1)
    			}
    			s.Count("hot")
    			s.Like(fmt.Sprintf("sq%d", i%7), user)
    		})
    	}
    	wg.Wait()
    	if got := wins.Load(); got != 50 {
    		t.Errorf("200 concurrent likes from 50 users: Like returned true %d times, want 50", got)
    	}
    	if got := s.Count("hot"); got != 50 {
    		t.Errorf("Count after 50 users liked concurrently = %d, want 50", got)
    	}

    	for i := range 50 {
    		user := fmt.Sprintf("user%d", i)
    		wg.Go(func() { s.Unlike("hot", user) })
    		wg.Go(func() { s.Unlike("hot", user) })
    	}
    	wg.Wait()
    	if got := s.Count("hot"); got != 0 {
    		t.Errorf("Count after everyone unliked concurrently = %d, want 0", got)
    	}
    }
---

Every squeak has a little heart. Tapping it sends `PUT /api/squeaks/{id}/like`,
tapping again sends `DELETE`. Both are idempotent: liking twice still counts
once. Hundreds of mice tap hearts at the same moment, so the store behind these
routes has to be safe for concurrent use.

Implement `LikeStore`:

- `NewLikeStore()` returns an empty store.
- `Like(squeakID, userID)` records the like. It returns `true` if it's new and
  `false` if that user already liked that squeak.
- `Unlike(squeakID, userID)` removes it and returns `true`, or returns `false`
  if there was nothing to remove.
- `Count(squeakID)` returns how many users like the squeak right now (`0` for a
  squeak nobody has liked).

## Example

```go
s := NewLikeStore()
s.Like("sq1", "pip")    // true
s.Like("sq1", "pip")    // false: already liked
s.Like("sq1", "bree")   // true
s.Count("sq1")          // 2
s.Unlike("sq1", "pip")  // true
s.Count("sq1")          // 1
```

## Constraints

- All three methods may be called from many goroutines at once. The tests fire
  200 simultaneous likes from 50 users and expect `Like` to return `true`
  exactly 50 times.
- Two stores must not share state.
