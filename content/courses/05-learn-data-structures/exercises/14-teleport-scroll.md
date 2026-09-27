---
title: Teleport Scroll
difficulty: hard
after: bfs-and-dfs
hints:
  - '`TravelTimes` is Dijkstra''s algorithm from the last chapter, on zone numbers instead of names: build an adjacency list (`[][]edge`, both directions for each road), keep a `dist` slice, and use a min-heap of `(zone, dist)` items with lazy deletion. Picking the next zone by scanning every zone is O(V²): far too slow for 100,000 zones.'
  - 'Trying every road as the free one and re-running Dijkstra each time is O(E · (V + E) log V): hours. Instead, think about **where** the scroll is used: a route that uses it on road `a → b` costs (fastest time from `from` to `a`) + 0 + (fastest time from `b` to `to`). Both parts are ordinary shortest paths.'
  - 'Run `TravelTimes` twice: once from `from` and once from `to` (roads are two-way, so the time from `b` to `to` equals the time from `to` to `b`). Then loop over every road, in **both** directions, and take the minimum, skipping unreachable ends (-1). Don''t forget the route that doesn''t use the scroll at all.'
exercise:
  starter: |
    package main

    import "fmt"

    // Road is a two-way road between zones A and B that takes Minutes to walk.
    type Road struct {
    	A, B, Minutes int
    }

    // TravelTimes returns the fastest time from zone `from` to every zone
    // 0..n-1, or -1 for zones that can't be reached.
    func TravelTimes(n int, roads []Road, from int) []int {
    	return nil
    }

    // ScrollTime returns the fastest time from `from` to `to` if the party
    // may cross at most one road instantly, or -1 if `to` can't be reached.
    func ScrollTime(n int, roads []Road, from, to int) int {
    	return 0
    }

    func main() {
    	// 0 Village, 1 Forest, 2 Lake, 3 Caves, 4 Castle, 5 Mountain, 6 Isle
    	roads := []Road{{0, 1, 5}, {0, 2, 3}, {2, 1, 4}, {1, 3, 15}, {3, 4, 6}, {2, 5, 8}, {5, 4, 12}}
    	fmt.Println(TravelTimes(7, roads, 0))   // want [0 5 3 20 23 11 -1]
    	fmt.Println(ScrollTime(7, roads, 0, 4)) // want 11
    	fmt.Println(ScrollTime(7, roads, 3, 2)) // want 4
    }
  solution: |
    package main

    import (
    	"container/heap"
    	"fmt"
    )

    type Road struct {
    	A, B, Minutes int
    }

    type edge struct{ to, minutes int }

    type item struct{ zone, dist int }

    type minHeap []item

    func (h minHeap) Len() int           { return len(h) }
    func (h minHeap) Less(i, j int) bool { return h[i].dist < h[j].dist }
    func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
    func (h *minHeap) Push(x any)        { *h = append(*h, x.(item)) }
    func (h *minHeap) Pop() any {
    	old := *h
    	it := old[len(old)-1]
    	*h = old[:len(old)-1]
    	return it
    }

    // TravelTimes returns the fastest time from `from` to every zone, or -1
    // for zones that can't be reached. It's Dijkstra's algorithm with a heap.
    func TravelTimes(n int, roads []Road, from int) []int {
    	adj := make([][]edge, n)
    	for _, r := range roads {
    		adj[r.A] = append(adj[r.A], edge{r.B, r.Minutes})
    		adj[r.B] = append(adj[r.B], edge{r.A, r.Minutes})
    	}
    	dist := make([]int, n)
    	for i := range dist {
    		dist[i] = -1
    	}
    	dist[from] = 0
    	h := &minHeap{{from, 0}}
    	for h.Len() > 0 {
    		it := heap.Pop(h).(item)
    		if it.dist > dist[it.zone] {
    			continue // stale entry
    		}
    		for _, e := range adj[it.zone] {
    			d := it.dist + e.minutes
    			if dist[e.to] == -1 || d < dist[e.to] {
    				dist[e.to] = d
    				heap.Push(h, item{e.to, d})
    			}
    		}
    	}
    	return dist
    }

    // ScrollTime returns the fastest time from `from` to `to` when at most one
    // road can be crossed instantly, or -1 if `to` can't be reached.
    func ScrollTime(n int, roads []Road, from, to int) int {
    	fromStart := TravelTimes(n, roads, from)
    	toGoal := TravelTimes(n, roads, to) // roads are two-way, so this is "time to reach to"
    	best := fromStart[to]
    	for _, r := range roads {
    		for _, ends := range [2][2]int{{r.A, r.B}, {r.B, r.A}} {
    			a, b := ends[0], ends[1]
    			if fromStart[a] >= 0 && toGoal[b] >= 0 {
    				if t := fromStart[a] + toGoal[b]; best == -1 || t < best {
    					best = t
    				}
    			}
    		}
    	}
    	return best
    }

    func main() {
    	// 0 Village, 1 Forest, 2 Lake, 3 Caves, 4 Castle, 5 Mountain, 6 Isle
    	roads := []Road{{0, 1, 5}, {0, 2, 3}, {2, 1, 4}, {1, 3, 15}, {3, 4, 6}, {2, 5, 8}, {5, 4, 12}}
    	fmt.Println(TravelTimes(7, roads, 0))
    	fmt.Println(ScrollTime(7, roads, 0, 4))
    	fmt.Println(ScrollTime(7, roads, 3, 2))
    }
  tests: |
    package main

    import (
    	"fmt"
    	"slices"
    	"testing"
    	"time"
    )

    func TestTravelTimes(t *testing.T) {
    	world := []Road{{0, 1, 5}, {0, 2, 3}, {2, 1, 4}, {1, 3, 15}, {3, 4, 6}, {2, 5, 8}, {5, 4, 12}}
    	tests := []struct {
    		name  string
    		n     int
    		roads []Road
    		from  int
    		want  []int
    	}{
    		{"example", 7, world, 0, []int{0, 5, 3, 20, 23, 11, -1}},
    		{"from the Caves", 7, world, 3, []int{20, 15, 19, 0, 6, 18, -1}},
    		{"from the island", 7, world, 6, []int{-1, -1, -1, -1, -1, -1, 0}},
    		{"one zone, no roads", 1, nil, 0, []int{0}},
    		{"parallel roads and a loop", 2, []Road{{0, 1, 10}, {1, 0, 4}, {1, 1, 3}}, 0, []int{0, 4}},
    		{"zero-minute portal", 3, []Road{{0, 1, 0}, {1, 2, 7}}, 2, []int{7, 7, 0}},
    		{"slow direct road, fast detour", 3, []Road{{0, 2, 100}, {0, 1, 1}, {1, 2, 1}}, 0, []int{0, 1, 2}},
    	}
    	for _, tt := range tests {
    		if got := TravelTimes(tt.n, tt.roads, tt.from); !slices.Equal(got, tt.want) {
    			t.Errorf("%s: TravelTimes(%d, %v, %d) = %v, want %v", tt.name, tt.n, tt.roads, tt.from, got, tt.want)
    		}
    	}
    }

    func TestScrollTime(t *testing.T) {
    	world := []Road{{0, 1, 5}, {0, 2, 3}, {2, 1, 4}, {1, 3, 15}, {3, 4, 6}, {2, 5, 8}, {5, 4, 12}}
    	tests := []struct {
    		name     string
    		n        int
    		roads    []Road
    		from, to int
    		want     int
    	}{
    		{"Village to Castle", 7, world, 0, 4, 11},
    		{"Caves to Lake", 7, world, 3, 2, 4},
    		{"already there", 7, world, 5, 5, 0},
    		{"island", 7, world, 0, 6, -1},
    		{"one road", 2, []Road{{0, 1, 7}}, 0, 1, 0},
    		{"no roads", 2, nil, 0, 1, -1},
    		{"scroll on the long middle road", 4, []Road{{0, 1, 1}, {1, 2, 50}, {2, 3, 1}}, 0, 3, 2},
    		{"scroll on a shortcut", 4, []Road{{0, 1, 10}, {1, 2, 10}, {2, 3, 10}, {0, 3, 100}}, 0, 3, 0},
    	}
    	for _, tt := range tests {
    		if got := ScrollTime(tt.n, tt.roads, tt.from, tt.to); got != tt.want {
    			t.Errorf("%s: ScrollTime(%d, %v, %d, %d) = %d, want %d", tt.name, tt.n, tt.roads, tt.from, tt.to, got, tt.want)
    		}
    	}
    }

    // bellmanFord is a slow but simple reference: relax every road n times.
    func bellmanFord(n int, roads []Road, from int) []int {
    	d := make([]int, n)
    	for i := range d {
    		d[i] = -1
    	}
    	d[from] = 0
    	for range n {
    		for _, r := range roads {
    			for _, e := range [2][2]int{{r.A, r.B}, {r.B, r.A}} {
    				if d[e[0]] >= 0 && (d[e[1]] < 0 || d[e[0]]+r.Minutes < d[e[1]]) {
    					d[e[1]] = d[e[0]] + r.Minutes
    				}
    			}
    		}
    	}
    	return d
    }

    func TestAgainstBruteForce(t *testing.T) {
    	var x uint64 = 1
    	next := func(k int) int {
    		x = x*6364136223846793005 + 1442695040888963407
    		return int(x>>33) % k
    	}
    	for range 300 {
    		n := 1 + next(10)
    		var roads []Road
    		for range next(16) {
    			roads = append(roads, Road{next(n), next(n), next(30)})
    		}
    		from, to := next(n), next(n)
    		if got, want := TravelTimes(n, roads, from), bellmanFord(n, roads, from); !slices.Equal(got, want) {
    			t.Fatalf("TravelTimes(%d, %v, %d) = %v, want %v", n, roads, from, got, want)
    		}
    		want := bellmanFord(n, roads, from)[to]
    		for i := range roads { // try every road as the free one
    			free := slices.Clone(roads)
    			free[i].Minutes = 0
    			if d := bellmanFord(n, free, from)[to]; d >= 0 && (want < 0 || d < want) {
    				want = d
    			}
    		}
    		if got := ScrollTime(n, roads, from, to); got != want {
    			t.Fatalf("ScrollTime(%d, %v, %d, %d) = %d, want %d", n, roads, from, to, got, want)
    		}
    	}
    }

    func TestBigWorld(t *testing.T) {
    	const n = 100_000
    	var x uint64 = 314
    	next := func(k int) int {
    		x = x*6364136223846793005 + 1442695040888963407
    		return int(x>>33) % k
    	}
    	var roads []Road
    	// A winding main road through zones 0..n-11, plus 100,000 random
    	// roads. The last 10 zones are unreachable islands.
    	for i := range n - 11 {
    		roads = append(roads, Road{i, i + 1, 1 + next(100)})
    	}
    	for range 100_000 {
    		roads = append(roads, Road{next(n - 10), next(n - 10), 50 + next(5000)})
    	}
    	type result struct {
    		sum, unreachable, far, scroll int
    	}
    	done := make(chan result, 1)
    	go func() {
    		var res result
    		d := TravelTimes(n, roads, 0)
    		for _, v := range d {
    			if v < 0 {
    				res.unreachable++
    			} else {
    				res.sum += v
    			}
    		}
    		res.far = d[n-11]
    		res.scroll = ScrollTime(n, roads, 0, n-11)
    		done <- res
    	}()
    	select {
    	case got := <-done:
    		want := result{308203008, 10, 4372, 2375}
    		if got != want {
    			t.Errorf("100,000 zones and %d roads: got (sum of times, unreachable, time to zone %d, scroll time) = %v, want %v", len(roads), n-11, fmt.Sprint(got), fmt.Sprint(want))
    		}
    	case <-time.After(time.Second):
    		t.Fatalf("TravelTimes + ScrollTime on 100,000 zones and %d roads took over a second: use a heap in Dijkstra, and run it a constant number of times", len(roads))
    	}
    }
---

The shop sells a one-use **teleport scroll**: read it at one end of a road and
you're instantly at the other end, however long the road is. Players want to know
how much time a scroll would save them on a trip.

Zones are numbered `0` to `n-1`, and each `Road{A, B, Minutes}` is a **two-way**
road that takes `Minutes` (≥ 0) to walk. Implement:

- **`TravelTimes(n, roads, from)`** returns a slice of length `n` with the fastest
  walking time from `from` to every zone, or `-1` for zones that can't be reached.
- **`ScrollTime(n, roads, from, to)`** returns the fastest time from `from` to `to`
  when the party may cross **at most one** road in 0 minutes, or `-1` if `to`
  can't be reached at all.

There may be several roads between the same two zones, and a road may loop from a
zone back to itself.

## Example

```
            5            15          6
  Village ───── Forest ───── Caves ───── Castle
     │         ╱                           │
   3 │      4 ╱                         12 │
     │       ╱                             │
    Lake ───────────── 8 ────────────── Mountain

  Isle (no roads)
```

Zones: 0 Village, 1 Forest, 2 Lake, 3 Caves, 4 Castle, 5 Mountain, 6 Isle.

```go
TravelTimes(7, roads, 0) // [0 5 3 20 23 11 -1]
ScrollTime(7, roads, 0, 4) // 11: walk Village → Lake → Mountain (11), scroll to the Castle
ScrollTime(7, roads, 3, 2) // 4: scroll Caves → Forest, walk to the Lake (4)
```

## Constraints

- Up to 100,000 zones and 200,000 roads; times fit easily in an `int`.
- The performance test runs `TravelTimes` once and `ScrollTime` once on that size,
  with a one-second limit. Dijkstra with a binary heap is O((V + E) log V).
  Choosing the next zone by scanning all of them is O(V²), and re-running
  Dijkstra once per road is far worse.
