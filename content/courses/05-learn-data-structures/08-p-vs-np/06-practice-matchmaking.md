---
title: 'Practice: Matchmaking'
exercise:
  starter: |
    package main

    import "fmt"

    // ---- Heap from the heaps chapter (finished) ----

    type Heap[T any] struct {
    	items  []T
    	before func(a, b T) bool // true if a should come out before b
    }

    func NewHeap[T any](before func(a, b T) bool) *Heap[T] {
    	return &Heap[T]{before: before}
    }

    func (h *Heap[T]) Len() int { return len(h.items) }

    // Peek returns the item that would come out next, without removing it.
    func (h *Heap[T]) Peek() T { return h.items[0] }

    func (h *Heap[T]) Push(v T) {
    	h.items = append(h.items, v)
    	i := len(h.items) - 1
    	for i > 0 {
    		p := (i - 1) / 2
    		if !h.before(h.items[i], h.items[p]) {
    			break
    		}
    		h.items[i], h.items[p] = h.items[p], h.items[i]
    		i = p
    	}
    }

    func (h *Heap[T]) Pop() (T, bool) {
    	var zero T
    	if len(h.items) == 0 {
    		return zero, false
    	}
    	top := h.items[0]
    	last := len(h.items) - 1
    	h.items[0] = h.items[last]
    	h.items[last] = zero
    	h.items = h.items[:last]
    	for i := 0; ; {
    		l, r, best := 2*i+1, 2*i+2, i
    		if l < len(h.items) && h.before(h.items[l], h.items[best]) {
    			best = l
    		}
    		if r < len(h.items) && h.before(h.items[r], h.items[best]) {
    			best = r
    		}
    		if best == i {
    			break
    		}
    		h.items[i], h.items[best] = h.items[best], h.items[i]
    		i = best
    	}
    	return top, true
    }

    // ---- The lobby ----

    // Lobby holds every online player: a hashmap of ratings and an undirected
    // friendship graph stored as adjacency lists.
    type Lobby struct {
    	rating  map[string]int
    	friends map[string][]string
    }

    func NewLobby() *Lobby {
    	return &Lobby{rating: map[string]int{}, friends: map[string][]string{}}
    }

    // Join adds a player (or updates their rating).
    func (l *Lobby) Join(name string, rating int) { l.rating[name] = rating }

    // Befriend records a two-way friendship.
    func (l *Lobby) Befriend(a, b string) {
    	l.friends[a] = append(l.friends[a], b)
    	l.friends[b] = append(l.friends[b], a)
    }

    func gap(a, b int) int { return max(a-b, b-a) }

    // Hops returns the fewest friendship links between from and to: 0 for the
    // same player, 1 for friends, 2 for friends of friends, and so on. It
    // returns -1 if there's no chain of friends between them, or if either
    // player isn't in the lobby.
    func (l *Lobby) Hops(from, to string) int {
    	// ? breadth-first search
    	return -1
    }

    // Rivals returns up to k opponents for name: the players whose ratings are
    // closest to name's, closest first (equal gaps: alphabetical order).
    // To keep matches fresh, it skips name itself and anyone within 2 hops
    // (friends and friends of friends). It returns nil if name isn't in the
    // lobby.
    func (l *Lobby) Rivals(name string, k int) []string {
    	// ? find who's within 2 hops, then keep the k best candidates in a Heap
    	return nil
    }

    func main() {
    	l := NewLobby()
    	for name, r := range map[string]int{"mira": 1500, "kai": 1480, "bo": 1530, "ada": 1400, "zed": 1510, "lu": 1900} {
    		l.Join(name, r)
    	}
    	l.Befriend("mira", "kai")
    	l.Befriend("kai", "ada")
    	l.Befriend("ada", "lu")

    	fmt.Println(l.Hops("mira", "lu"), l.Hops("mira", "bo")) // want: 3 -1
    	fmt.Println(l.Rivals("mira", 3))                        // want: [zed bo lu]
    }
  solution: |
    package main

    import "fmt"

    // ---- Heap from the heaps chapter (finished) ----

    type Heap[T any] struct {
    	items  []T
    	before func(a, b T) bool // true if a should come out before b
    }

    func NewHeap[T any](before func(a, b T) bool) *Heap[T] {
    	return &Heap[T]{before: before}
    }

    func (h *Heap[T]) Len() int { return len(h.items) }

    // Peek returns the item that would come out next, without removing it.
    func (h *Heap[T]) Peek() T { return h.items[0] }

    func (h *Heap[T]) Push(v T) {
    	h.items = append(h.items, v)
    	i := len(h.items) - 1
    	for i > 0 {
    		p := (i - 1) / 2
    		if !h.before(h.items[i], h.items[p]) {
    			break
    		}
    		h.items[i], h.items[p] = h.items[p], h.items[i]
    		i = p
    	}
    }

    func (h *Heap[T]) Pop() (T, bool) {
    	var zero T
    	if len(h.items) == 0 {
    		return zero, false
    	}
    	top := h.items[0]
    	last := len(h.items) - 1
    	h.items[0] = h.items[last]
    	h.items[last] = zero
    	h.items = h.items[:last]
    	for i := 0; ; {
    		l, r, best := 2*i+1, 2*i+2, i
    		if l < len(h.items) && h.before(h.items[l], h.items[best]) {
    			best = l
    		}
    		if r < len(h.items) && h.before(h.items[r], h.items[best]) {
    			best = r
    		}
    		if best == i {
    			break
    		}
    		h.items[i], h.items[best] = h.items[best], h.items[i]
    		i = best
    	}
    	return top, true
    }

    // ---- The lobby ----

    // Lobby holds every online player: a hashmap of ratings and an undirected
    // friendship graph stored as adjacency lists.
    type Lobby struct {
    	rating  map[string]int
    	friends map[string][]string
    }

    func NewLobby() *Lobby {
    	return &Lobby{rating: map[string]int{}, friends: map[string][]string{}}
    }

    // Join adds a player (or updates their rating).
    func (l *Lobby) Join(name string, rating int) { l.rating[name] = rating }

    // Befriend records a two-way friendship.
    func (l *Lobby) Befriend(a, b string) {
    	l.friends[a] = append(l.friends[a], b)
    	l.friends[b] = append(l.friends[b], a)
    }

    func gap(a, b int) int { return max(a-b, b-a) }

    // within returns every player at most maxHops links from start, with their
    // distance: a breadth-first search that stops expanding at maxHops.
    func (l *Lobby) within(start string, maxHops int) map[string]int {
    	dist := map[string]int{start: 0}
    	queue := []string{start}
    	for len(queue) > 0 {
    		cur := queue[0]
    		queue = queue[1:]
    		if dist[cur] == maxHops {
    			continue
    		}
    		for _, f := range l.friends[cur] {
    			if _, seen := dist[f]; !seen {
    				dist[f] = dist[cur] + 1
    				queue = append(queue, f)
    			}
    		}
    	}
    	return dist
    }

    func (l *Lobby) Hops(from, to string) int {
    	_, okFrom := l.rating[from]
    	_, okTo := l.rating[to]
    	if !okFrom || !okTo {
    		return -1
    	}
    	if d, ok := l.within(from, len(l.rating))[to]; ok {
    		return d
    	}
    	return -1
    }

    type candidate struct {
    	name string
    	gap  int
    }

    // worse reports whether a is a worse match than b.
    func worse(a, b candidate) bool {
    	if a.gap != b.gap {
    		return a.gap > b.gap
    	}
    	return a.name > b.name
    }

    func (l *Lobby) Rivals(name string, k int) []string {
    	mine, ok := l.rating[name]
    	if !ok {
    		return nil
    	}
    	near := l.within(name, 2)
    	// A heap with the WORST of the current best k on top, so it's cheap to evict.
    	h := NewHeap(worse)
    	for other, r := range l.rating {
    		if _, tooClose := near[other]; tooClose {
    			continue
    		}
    		c := candidate{other, gap(mine, r)}
    		switch {
    		case h.Len() < k:
    			h.Push(c)
    		case k > 0 && worse(h.Peek(), c):
    			h.Pop()
    			h.Push(c)
    		}
    	}
    	out := make([]string, h.Len())
    	for i := len(out) - 1; i >= 0; i-- {
    		c, _ := h.Pop()
    		out[i] = c.name
    	}
    	return out
    }

    func main() {
    	l := NewLobby()
    	for name, r := range map[string]int{"mira": 1500, "kai": 1480, "bo": 1530, "ada": 1400, "zed": 1510, "lu": 1900} {
    		l.Join(name, r)
    	}
    	l.Befriend("mira", "kai")
    	l.Befriend("kai", "ada")
    	l.Befriend("ada", "lu")

    	fmt.Println(l.Hops("mira", "lu"), l.Hops("mira", "bo")) // want: 3 -1
    	fmt.Println(l.Rivals("mira", 3))                        // want: [zed bo lu]
    }
  tests: |
    package main

    import (
    	"fmt"
    	"maps"
    	"slices"
    	"testing"
    )

    func testLobby() *Lobby {
    	l := NewLobby()
    	for name, r := range map[string]int{
    		"mira": 1500, "kai": 1480, "bo": 1530, "ada": 1400, "zed": 1510,
    		"lu": 1900, "rex": 1490, "ivy": 1520, "tom": 1000,
    	} {
    		l.Join(name, r)
    	}
    	// mira - kai - ada - lu,  mira - rex,  bo - ivy
    	l.Befriend("mira", "kai")
    	l.Befriend("kai", "ada")
    	l.Befriend("ada", "lu")
    	l.Befriend("mira", "rex")
    	l.Befriend("bo", "ivy")
    	return l
    }

    func snapshot(l *Lobby) (map[string]int, map[string][]string) {
    	f := map[string][]string{}
    	for k, v := range l.friends {
    		f[k] = slices.Clone(v)
    	}
    	return maps.Clone(l.rating), f
    }

    func TestHops(t *testing.T) {
    	l := testLobby()
    	for _, tt := range []struct {
    		from, to string
    		want     int
    	}{
    		{"mira", "mira", 0}, {"mira", "kai", 1}, {"kai", "mira", 1}, {"mira", "ada", 2},
    		{"rex", "lu", 4}, {"lu", "mira", 3}, {"mira", "bo", -1}, {"tom", "mira", -1},
    		{"bo", "ivy", 1}, {"mira", "ghost", -1}, {"ghost", "ghost", -1},
    	} {
    		if got := l.Hops(tt.from, tt.to); got != tt.want {
    			t.Errorf("Hops(%q, %q) = %d, want %d", tt.from, tt.to, got, tt.want)
    		}
    	}
    	// Shortest, not first-found: add a shortcut.
    	l.Befriend("rex", "lu")
    	if got := l.Hops("mira", "lu"); got != 2 {
    		t.Errorf("after rex and lu become friends, Hops(mira, lu) = %d, want 2 (mira-rex-lu)", got)
    	}
    }

    func TestRivals(t *testing.T) {
    	l := testLobby()
    	rating, friends := snapshot(l)
    	for _, tt := range []struct {
    		name string
    		k    int
    		want []string
    	}{
    		// mira 1500: kai, rex (friends) and ada (2 hops) are skipped.
    		{"mira", 3, []string{"zed", "ivy", "bo"}},
    		{"mira", 10, []string{"zed", "ivy", "bo", "lu", "tom"}},
    		{"mira", 1, []string{"zed"}},
    		// bo 1530: only ivy is a friend. Gaps: ivy 10 (skip), zed 20, mira 30, rex 40, kai 50.
    		{"bo", 4, []string{"zed", "mira", "rex", "kai"}},
    		// lu has lots of room at the top: ada (friend) and kai (2 hops) are skipped.
    		{"lu", 2, []string{"bo", "ivy"}},
    		{"tom", 2, []string{"ada", "kai"}},
    		{"mira", 0, []string{}},
    	} {
    		got := l.Rivals(tt.name, tt.k)
    		if !slices.Equal(got, tt.want) {
    			t.Errorf("Rivals(%q, %d) = %q, want %q", tt.name, tt.k, got, tt.want)
    		}
    	}
    	if got := l.Rivals("ghost", 3); got != nil {
    		t.Errorf("Rivals of a player who isn't in the lobby = %q, want nil", got)
    	}
    	r2, f2 := snapshot(l)
    	if !maps.Equal(rating, r2) || !maps.EqualFunc(friends, f2, slices.Equal) {
    		t.Error("Hops/Rivals changed the lobby's ratings or friendships")
    	}
    }

    func TestRivalsTies(t *testing.T) {
    	l := NewLobby()
    	l.Join("me", 1000)
    	for _, n := range []string{"eve", "dan", "cat", "bob", "amy"} {
    		l.Join(n, 1100)
    	}
    	l.Join("far", 1300)
    	l.Join("low", 900)
    	if got, want := l.Rivals("me", 4), []string{"amy", "bob", "cat", "dan"}; !slices.Equal(got, want) {
    		t.Errorf("Rivals with five equal gaps of 100 (and low at 100 too) = %q, want %q", got, want)
    	}
    	if got, want := l.Rivals("me", 6), []string{"amy", "bob", "cat", "dan", "eve", "low"}; !slices.Equal(got, want) {
    		t.Errorf("Rivals(me, 6) = %q, want %q", got, want)
    	}
    }

    func TestBigLobby(t *testing.T) {
    	l := NewLobby()
    	for i := range 50_000 {
    		l.Join(fmt.Sprintf("p%05d", i), i*2)
    		if i > 0 {
    			l.Befriend(fmt.Sprintf("p%05d", i-1), fmt.Sprintf("p%05d", i)) // one long chain
    		}
    	}
    	if got := l.Hops("p00000", "p49999"); got != 49_999 {
    		t.Errorf("in a 50000-player chain, Hops(first, last) = %d, want 49999", got)
    	}
    	want := []string{"p24997", "p25003", "p24996", "p25004"}
    	if got := l.Rivals("p25000", 4); !slices.Equal(got, want) {
    		t.Errorf("Rivals(p25000, 4) in a chain = %q, want %q (neighbours within 2 hops are skipped)", got, want)
    	}
    }
---

Launch week. The studio's new ranked mode needs a **matchmaker**, and it's
going to use three structures from this course at once:

- a **hashmap** from player name to rating, for O(1) lookups,
- a **graph** of friendships, searched **breadth-first**,
- and a **heap** to pick the best few opponents out of everyone online.

The `Heap[T]` from the heaps chapter and the `Lobby` type are already in the
editor. `Lobby.rating` is the hashmap and `Lobby.friends` is an undirected
graph stored as adjacency lists. `Join` and `Befriend` are done. Your job is the
two queries.

## Hops: how connected are two players?

`Hops(from, to)` returns the fewest friendship links between two players:
0 for the same player, 1 for friends, 2 for friends of friends. Return -1 if
no chain of friends connects them, or if either name isn't in `rating`.

"Fewest links" in an unweighted graph is exactly what **BFS** finds, as you saw
in the shortest-path lesson: keep a queue of players to visit and a map of
each discovered player's distance. The first time you discover a player, you've
found their shortest distance. Every player and friendship is handled at most
once, so it's O(V + E).

## Rivals: who should you play?

Good matches are close in rating. But players get bored facing the same
friends every night, so the matchmaker skips anyone **within 2 hops**:
friends and friends of friends.

`Rivals(name, k)` returns up to `k` opponents, closest rating first (use the
`gap` helper), with equal gaps in alphabetical order. It skips `name` itself
and everyone within 2 hops, and returns `nil` if `name` isn't in the lobby.

Two tips:

1. **Reuse your BFS.** A helper that runs BFS from a start player and **stops
   expanding** at a maximum distance gives you the "within 2 hops" set in one
   call. `Hops` can use the same helper with no real limit.
2. **Top-k with a heap.** This is the top-k pattern from the heaps chapter:
   keep a heap of at most `k` candidates with the **worst** one on top
   (largest gap, then latest name). For each other player, push if the heap
   isn't full, or replace the top if they're a better match. At the end, pop
   everything: the pops come out worst first, so fill the result from the back.
   `NewHeap` takes a `before(a, b)` function; for a worst-on-top heap that's
   "a is a worse match than b".

Ranging over `rating` visits players in random order. That's fine here: the
heap and the tie-break rule make the answer the same every time.

**Run** prints two `Hops` answers and a `Rivals` list. **Submit** also checks
shortcuts in the friend graph, tie-breaks, unknown players, that nothing in the
lobby changes, and a 50,000-player friendship chain, where anything slower
than O(V + E) will time out.

## Course complete

That's Data Structures and Algorithms 2 finished. You've built trees that stay
balanced, hashmaps that resize, tries that autocomplete, heaps that rank,
graphs you can search, and you know when a problem has no fast exact answer.
Now you've made them work together.

Next up is [Learn Concurrency in Go](/courses/learn-concurrency), where you'll
put all those cores to work.
